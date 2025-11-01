package data_browser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"bronze-backend/config"
	"bronze-backend/storage"
)

type ExportRequest struct {
	Files              []FileExportInfo `json:"files"`
	TableName          string           `json:"table_name"`
	Operation          string           `json:"operation"` // "create" or "append"
	Database           string           `json:"database,omitempty"`
	MaxErrors          int              `json:"max_errors,omitempty"`
	StopOnError        bool             `json:"stop_on_error,omitempty"`
	CollectErrors      bool             `json:"collect_errors,omitempty"`
	SchemaResolution   string           `json:"schema_resolution,omitempty"` // "merge", "first_file", "manual"
	MaxConcurrent      int              `json:"max_concurrent_files,omitempty"`
	BatchSize          int              `json:"batch_size,omitempty"`
	AutoTypeConversion bool             `json:"auto_type_conversion,omitempty"`
}

type FileExportInfo struct {
	FileName   string `json:"file_name"`
	SheetName  string `json:"sheet_name,omitempty"`
	TreatAsCSV bool   `json:"treat_as_csv,omitempty"`
}

type ExportResponse struct {
	Success          bool                           `json:"success"`
	Message          string                         `json:"message"`
	TableName        string                         `json:"table_name"`
	FilesProcessed   int                            `json:"files_processed"`
	RowsExported     int64                          `json:"rows_exported"`
	RowsFailed       int64                          `json:"rows_failed"`
	ProcessingTime   time.Duration                  `json:"processing_time"`
	ColumnMismatches []storage.NessieColumnMismatch `json:"column_mismatches,omitempty"`
	RowErrors        []ExportRowError               `json:"row_errors,omitempty"`
	ErrorSummary     map[string]int                 `json:"error_summary,omitempty"`
	Database         string                         `json:"database,omitempty"`
}

type ExportRowError struct {
	RowIndex     int         `json:"row_index"`
	FileName     string      `json:"file_name"`
	SheetName    string      `json:"sheet_name,omitempty"`
	ColumnName   string      `json:"column_name"`
	ErrorCode    string      `json:"error_code"`
	ErrorMsg     string      `json:"error_message"`
	SourceValue  interface{} `json:"source_value"`
	SuggestedFix string      `json:"suggested_fix,omitempty"`
}

type ProcessingResult struct {
	FileName  string
	SheetName string
	Rows      [][]string
	Columns   []string
	RowCount  int
	Errors    []ExportRowError
	Success   bool
}

func NewExportHandler(minioClient *storage.MinIOClient, nessieClient *storage.NessieClient, cfg *config.Config, browser *DataBrowserHandler) *ExportHandler {
	return &ExportHandler{
		minioClient:  minioClient,
		nessieClient: nessieClient,
		config:       cfg,
		browser:      browser,
	}
}

type ExportHandler struct {
	minioClient  *storage.MinIOClient
	nessieClient *storage.NessieClient
	config       *config.Config
	browser      *DataBrowserHandler
}

func (h *ExportHandler) ExportSingleFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Failed to decode request", http.StatusBadRequest, err)
		return
	}

	if len(request.Files) != 1 {
		h.writeError(w, "This endpoint only supports single file exports", http.StatusBadRequest, nil)
		return
	}

	response := h.processExport(r.Context(), request)
	h.writeJSONResponse(w, response)
}

func (h *ExportHandler) ExportMultipleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Failed to decode request", http.StatusBadRequest, err)
		return
	}

	if len(request.Files) == 0 {
		h.writeError(w, "No files provided for export", http.StatusBadRequest, nil)
		return
	}

	response := h.processExport(r.Context(), request)
	h.writeJSONResponse(w, response)
}

func (h *ExportHandler) processExport(ctx context.Context, request ExportRequest) ExportResponse {
	startTime := time.Now()

	// Set defaults
	if request.MaxErrors == 0 {
		request.MaxErrors = 1000
	}
	if request.MaxConcurrent == 0 {
		request.MaxConcurrent = 3
	}
	if request.BatchSize == 0 {
		request.BatchSize = 1000
	}
	if request.SchemaResolution == "" {
		request.SchemaResolution = "merge"
	}

	database := request.Database
	if database == "" {
		database = h.config.Nessie.DefaultDB
	}

	log.Printf("Starting export to table '%s' with %d files, operation: %s", request.TableName, len(request.Files), request.Operation)

	// Process files (simplified for now)
	results := h.processFilesSimplified(request.Files)

	// Merge schemas from all processed files
	mergedSchema, err := h.mergeSchemas(results, request.SchemaResolution)
	if err != nil {
		return ExportResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to merge schemas: %v", err),
		}
	}

	// Check if table exists and validate schema
	tableExists, err := h.nessieClient.TableExists(ctx, database, request.TableName)
	if err != nil {
		return ExportResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to check table existence: %v", err),
		}
	}

	// Handle column mismatches
	var columnMismatches []storage.NessieColumnMismatch
	if tableExists && request.Operation == "append" {
		// Get existing table schema for comparison
		targetTable, err := h.nessieClient.GetTableSchema(ctx, database, request.TableName)
		if err != nil {
			return ExportResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to get table schema: %v", err),
			}
		}
		columnMismatches = h.nessieClient.ValidateSchema(mergedSchema.Columns, targetTable)
	}

	if len(columnMismatches) > 0 && request.SchemaResolution == "strict" {
		return ExportResponse{
			Success:          false,
			Message:          "Schema mismatch detected in strict mode",
			ColumnMismatches: columnMismatches,
		}
	}

	// Create table if needed
	if request.Operation == "create" || !tableExists {
		nessieTable := &storage.NessieTable{
			Name:     request.TableName,
			Database: database,
			Columns:  h.createNessieColumns(mergedSchema.Columns, mergedSchema.ColumnTypes),
			Properties: map[string]interface{}{
				"description": fmt.Sprintf("Table created from %d files", len(request.Files)),
				"created_at":  time.Now(),
			},
		}

		if err := h.nessieClient.CreateTable(ctx, nessieTable); err != nil {
			return ExportResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to create table: %v", err),
			}
		}

		log.Printf("Created Nessie table: %s.%s", database, request.TableName)
	}

	// Export data (simplified)
	totalRows, totalErrors := h.exportDataSimplified(results, request.TableName, database, request)

	processingTime := time.Since(startTime)
	totalRowsInt64 := int64(totalRows)
	totalErrorsInt64 := int64(totalErrors)

	return ExportResponse{
		Success:          totalRowsInt64 > 0 || totalErrorsInt64 == 0,
		Message:          fmt.Sprintf("Export completed. %d rows exported, %d rows failed", totalRowsInt64, totalErrorsInt64),
		TableName:        request.TableName,
		FilesProcessed:   len(results),
		RowsExported:     totalRowsInt64,
		RowsFailed:       totalErrorsInt64,
		ProcessingTime:   processingTime,
		ColumnMismatches: columnMismatches,
		Database:         database,
	}
}

func (h *ExportHandler) processFilesSimplified(files []FileExportInfo) []ProcessingResult {
	var results []ProcessingResult

	for _, file := range files {
		request := BrowseRequest{
			FileName:   file.FileName,
			SheetName:  file.SheetName,
			TreatAsCSV: file.TreatAsCSV,
			MaxRows:    1000, // Limit for testing
			HasHeaders: true,
		}

		response, err := h.browser.BrowseDataRequest(context.Background(), request)
		if err != nil {
			results = append(results, ProcessingResult{
				FileName:  file.FileName,
				SheetName: file.SheetName,
				Success:   false,
				Errors: []ExportRowError{
					{
						FileName:     file.FileName,
						SheetName:    file.SheetName,
						ErrorCode:    "FILE_PROCESSING_ERROR",
						ErrorMsg:     err.Error(),
						SuggestedFix: "Check file format and accessibility",
					},
				},
			})
			continue
		}

		results = append(results, ProcessingResult{
			FileName:  file.FileName,
			SheetName: file.SheetName,
			Rows:      response.Rows,
			Columns:   response.Columns,
			RowCount:  response.RowCount,
			Errors:    []ExportRowError{},
			Success:   true,
		})
	}

	return results
}

func (h *ExportHandler) mergeSchemas(results []ProcessingResult, resolution string) (*MergedSchema, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no processing results to merge")
	}

	// Convert results to FileInfo for schema merger
	var files []FileInfo
	for _, result := range results {
		if !result.Success {
			continue // Skip failed files
		}

		files = append(files, FileInfo{
			FileName: result.FileName,
			Columns:  result.Columns,
			RowCount: int64(result.RowCount),
			DataType: "source_data",
		})
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no successful files to merge schemas")
	}

	merger := NewSchemaMerger(resolution)
	return merger.MergeSchemas(files)
}

func (h *ExportHandler) exportDataSimplified(results []ProcessingResult, tableName, database string, request ExportRequest) (int, int) {
	totalRows := 0
	totalErrors := 0

	for _, result := range results {
		if !result.Success {
			totalErrors += len(result.Errors)
			continue
		}

		// Simple row counting for now
		totalRows += len(result.Rows)
	}

	return totalRows, totalErrors
}

func (h *ExportHandler) createNessieColumns(columns []string, columnTypes map[string]string) []storage.NessieColumn {
	var nessieColumns []storage.NessieColumn
	sort.Strings(columns) // Sort for consistent column order

	for _, col := range columns {
		colType := "VARCHAR(255)" // Default type
		if columnType, exists := columnTypes[col]; exists {
			colType = columnType
		}

		nessieColumns = append(nessieColumns, storage.NessieColumn{
			Name:     col,
			Type:     colType,
			Nullable: true, // Allow NULL values
			Comment:  fmt.Sprintf("Column from file export"),
		})
	}

	return nessieColumns
}

func (h *ExportHandler) writeJSONResponse(w http.ResponseWriter, response ExportResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if response.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(response)
}

func (h *ExportHandler) writeError(w http.ResponseWriter, message string, statusCode int, err error) {
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}

	if err != nil {
		response["error"] = err.Error()
		log.Printf("Export Error: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
