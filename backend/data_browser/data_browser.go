package data_browser

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bronze-backend/storage"
	_ "github.com/microsoft/go-mssqldb" // Import for MDB support
	"github.com/tealeg/xlsx/v3"
)

type DataBrowserHandler struct {
	minioClient *storage.MinIOClient
}

func NewDataBrowserHandler(minioClient *storage.MinIOClient) *DataBrowserHandler {
	return &DataBrowserHandler{
		minioClient: minioClient,
	}
}

type BrowseRequest struct {
	FileName          string `json:"file_name"`
	SheetName         string `json:"sheet_name,omitempty"`
	MaxRows           int    `json:"max_rows,omitempty"`
	Offset            int    `json:"offset,omitempty"`
	HasHeaders        bool   `json:"has_headers,omitempty"`
	TreatAsCSV        bool   `json:"treat_as_csv,omitempty"`
	AutoDetectHeaders bool   `json:"auto_detect_headers,omitempty"`
	StreamMode        bool   `json:"stream_mode,omitempty"`
	ChunkSize         int    `json:"chunk_size,omitempty"`
}

type BrowseResponse struct {
	Success    bool       `json:"success"`
	Message    string     `json:"message"`
	DataType   string     `json:"data_type"`
	FileName   string     `json:"file_name"`
	SheetName  string     `json:"sheet_name,omitempty"`
	Columns    []string   `json:"columns"`
	Rows       [][]string `json:"rows"`
	TotalRows  int64      `json:"total_rows"`
	RowCount   int        `json:"row_count"`
	Offset     int        `json:"offset"`
	HasHeaders bool       `json:"has_headers"`
	Sheets     []string   `json:"sheets,omitempty"`
}

type FileInfoListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Files   []DataFileInfo `json:"files"`
	Count   int            `json:"count"`
}

type DataFileInfo struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	DataType     string    `json:"data_type"`
	Sheets       []string  `json:"sheets,omitempty"`
	Columns      []string  `json:"columns,omitempty"`
	RowCount     int64     `json:"row_count,omitempty"`
}

func (h *DataBrowserHandler) BrowseData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request BrowseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Failed to decode request", http.StatusBadRequest, err)
		return
	}

	response, err := h.BrowseDataRequest(r.Context(), request)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DataBrowserHandler) BrowseDataRequest(ctx context.Context, request BrowseRequest) (BrowseResponse, error) {
	if request.FileName == "" {
		return BrowseResponse{}, fmt.Errorf("file name is required")
	}

	// Set defaults
	if request.MaxRows <= 0 {
		request.MaxRows = 100
	}
	if request.MaxRows > 10000 {
		request.MaxRows = 10000 // Cap at 10k rows
	}
	if request.ChunkSize <= 0 {
		request.ChunkSize = 1000 // Default chunk size for streaming
	}

	// Get file from S3
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second) // Longer timeout for streaming
	defer cancel()

	reader, err := h.minioClient.DownloadFile(ctx, request.FileName)
	if err != nil {
		return BrowseResponse{}, fmt.Errorf("failed to download file: %w", err)
	}
	defer reader.Close()

	// Handle streaming mode (not supported in request mode)
	if request.StreamMode {
		return BrowseResponse{}, fmt.Errorf("streaming mode not supported in request mode")
	}

	// Read file into memory for non-streaming mode
	data, err := io.ReadAll(reader)
	if err != nil {
		return BrowseResponse{}, fmt.Errorf("failed to read file data: %w", err)
	}

	// Determine file type and process
	ext := strings.ToLower(filepath.Ext(request.FileName))
	var response BrowseResponse

	// If treat_as_csv is true, process as CSV regardless of extension
	if request.TreatAsCSV {
		response, err = h.processCSVFile(data, request)
	} else {
		switch ext {
		case ".xlsx", ".xls", ".xlsm":
			response, err = h.processExcelFile(data, request)
		case ".csv":
			response, err = h.processCSVFile(data, request)
		case ".mdb":
			response, err = h.processMDBFile(data, request)
		default:
			return BrowseResponse{}, fmt.Errorf("unsupported file type: %s", ext)
		}
	}

	if err != nil {
		return BrowseResponse{}, fmt.Errorf("processing failed: %w", err)
	}

	return response, nil
}

func (h *DataBrowserHandler) ListDataFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// List all files
	files, err := h.minioClient.ListFiles(ctx, "", 0)
	if err != nil {
		h.writeError(w, "Failed to list files", http.StatusInternalServerError, err)
		return
	}

	var dataFiles []DataFileInfo
	supportedExtensions := map[string]bool{
		".xlsx":  true,
		".xls":   true,
		".xlsm":  true,
		".csv":   true,
		".mdb":   true,
		".accdb": true, // Add ACCDB support
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Key))

		dataFile := DataFileInfo{
			Name:         file.Key,
			Size:         file.Size,
			LastModified: file.LastModified,
			DataType:     h.getDataType(ext),
		}

		// For Excel files (including XLSM), try to get sheet names without reading all data
		if ext == ".xlsx" || ext == ".xls" || ext == ".xlsm" {
			if sheets, columns, rowCount, err := h.getExcelInfo(ctx, file.Key); err == nil {
				dataFile.Sheets = sheets
				dataFile.Columns = columns
				dataFile.RowCount = rowCount
			}
		} else if ext == ".csv" || !supportedExtensions[ext] {
			// For CSV files and other files that can be treated as CSV, get basic info
			if columns, rowCount, err := h.getCSVInfo(ctx, file.Key); err == nil {
				dataFile.Columns = columns
				dataFile.RowCount = rowCount
				if !supportedExtensions[ext] {
					dataFile.DataType = "treatable_as_csv"
				}
			}
		} else if ext == ".mdb" || ext == ".accdb" {
			// For MDB files, get table and column info
			if tables, columns, rowCount, err := h.getMDBInfo(ctx, file.Key); err == nil {
				dataFile.Sheets = tables
				dataFile.Columns = columns
				dataFile.RowCount = rowCount
			}
		}

		// Include all supported files plus mention that others can be treated as CSV
		if supportedExtensions[ext] || !supportedExtensions[ext] {
			dataFiles = append(dataFiles, dataFile)
		}
	}

	response := FileInfoListResponse{
		Success: true,
		Message: "Data files listed successfully (all files can be treated as CSV with treat_as_csv=true)",
		Files:   dataFiles,
		Count:   len(dataFiles),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *DataBrowserHandler) processExcelFile(data []byte, request BrowseRequest) (BrowseResponse, error) {
	response := BrowseResponse{
		Success:    true,
		Message:    "Excel file processed successfully",
		DataType:   "excel",
		FileName:   request.FileName,
		HasHeaders: request.HasHeaders,
		Offset:     request.Offset,
	}

	// Open Excel file
	wb, err := xlsx.OpenBinary(data)
	if err != nil {
		return response, fmt.Errorf("failed to open Excel file: %w", err)
	}

	// Get all sheet names
	var sheetNames []string
	for _, sheet := range wb.Sheets {
		sheetNames = append(sheetNames, sheet.Name)
	}
	response.Sheets = sheetNames

	// Select sheet
	targetSheet := request.SheetName
	if targetSheet == "" {
		if len(wb.Sheets) > 0 {
			targetSheet = wb.Sheets[0].Name
		} else {
			return response, fmt.Errorf("no sheets found in workbook")
		}
	}

	sheet, ok := wb.Sheet[targetSheet]
	if !ok {
		return response, fmt.Errorf("sheet '%s' not found", targetSheet)
	}

	response.SheetName = targetSheet

	// Get all rows to calculate total and extract data
	var allRows []*xlsx.Row
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		allRows = append(allRows, row)
		return nil
	})
	if err != nil {
		return response, fmt.Errorf("failed to read sheet rows: %w", err)
	}

	response.TotalRows = int64(len(allRows))

	// Determine start and end rows
	startRow := request.Offset
	if startRow >= len(allRows) {
		response.Rows = [][]string{}
		response.RowCount = 0
		return response, nil
	}

	endRow := startRow + request.MaxRows
	if endRow > len(allRows) {
		endRow = len(allRows)
	}

	if len(allRows) == 0 {
		return response, nil
	}

	// Get columns from first row
	firstRow := allRows[0]
	var cols []string
	firstRow.ForEachCell(func(cell *xlsx.Cell) error {
		cellValue, _ := cell.FormattedValue()
		cols = append(cols, cellValue)
		return nil
	})
	response.Columns = cols

	// Process data rows
	dataStart := 0
	if request.HasHeaders {
		dataStart = 1
	}

	var rows [][]string
	for i := startRow + dataStart; i < endRow; i++ {
		if i >= len(allRows) {
			break
		}

		row := allRows[i]
		var rowData []string
		row.ForEachCell(func(cell *xlsx.Cell) error {
			cellValue, _ := cell.FormattedValue()
			rowData = append(rowData, cellValue)
			return nil
		})

		// Ensure row has same number of columns as header
		for len(rowData) < len(response.Columns) {
			rowData = append(rowData, "")
		}
		if len(rowData) > len(response.Columns) {
			rowData = rowData[:len(response.Columns)]
		}
		rows = append(rows, rowData)
	}

	response.Rows = rows
	response.RowCount = len(rows)

	return response, nil
}

func (h *DataBrowserHandler) processCSVFile(data []byte, request BrowseRequest) (BrowseResponse, error) {
	response := BrowseResponse{
		Success:    true,
		Message:    "CSV file processed successfully",
		DataType:   "csv",
		FileName:   request.FileName,
		HasHeaders: request.HasHeaders,
		Offset:     request.Offset,
	}

	// Handle empty data
	if len(data) == 0 {
		response.Message = "File is empty"
		return response, nil
	}

	// Auto-detect delimiter
	detectedDelim := h.detectDelimiter(data)
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = detectedDelim
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read all records to get total count
	allRecords, err := reader.ReadAll()
	if err != nil {
		return response, fmt.Errorf("failed to read CSV data: %w", err)
	}

	// Update message with detected delimiter info
	delimName := "comma"
	switch detectedDelim {
	case ';':
		delimName = "semicolon"
	case '\t':
		delimName = "tab"
	case '|':
		delimName = "pipe"
	}

	if request.TreatAsCSV {
		response.Message = fmt.Sprintf("File processed as CSV (detected delimiter: %s)", delimName)
	} else {
		response.Message = fmt.Sprintf("CSV file processed successfully (delimiter: %s)", delimName)
	}

	response.TotalRows = int64(len(allRecords))

	if len(allRecords) == 0 {
		return response, nil
	}

	// Auto-detect headers if requested
	hasHeaders := request.HasHeaders
	if request.AutoDetectHeaders && !hasHeaders {
		hasHeaders = h.detectHeaders(allRecords)
		response.HasHeaders = hasHeaders
		if hasHeaders {
			response.Message += " (headers auto-detected)"
		}
	}

	// Get columns from first row
	response.Columns = allRecords[0]

	// Determine data start
	dataStart := 0
	if hasHeaders {
		dataStart = 1
	}

	// Calculate range
	startRow := request.Offset + dataStart
	if startRow >= len(allRecords) {
		response.Rows = [][]string{}
		response.RowCount = 0
		return response, nil
	}

	endRow := startRow + request.MaxRows
	if endRow > len(allRecords) {
		endRow = len(allRecords)
	}

	// Extract rows
	var rows [][]string
	for i := startRow; i < endRow; i++ {
		// Ensure row has same number of columns as header
		rowData := make([]string, len(response.Columns))
		for j := 0; j < len(allRecords[i]) && j < len(response.Columns); j++ {
			rowData[j] = allRecords[i][j]
		}
		rows = append(rows, rowData)
	}

	response.Rows = rows
	response.RowCount = len(rows)

	return response, nil
}

func (h *DataBrowserHandler) processMDBFile(data []byte, request BrowseRequest) (BrowseResponse, error) {
	response := BrowseResponse{
		Success:    true,
		Message:    "MDB file processed successfully",
		DataType:   "mdb",
		FileName:   request.FileName,
		HasHeaders: request.HasHeaders,
		Offset:     request.Offset,
	}

	// Create temporary file for MDB database
	tempFile, err := os.CreateTemp("", "tempdb_*.mdb")
	if err != nil {
		return response, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write MDB data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return response, fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open MDB database using connection string
	connStr := fmt.Sprintf("Provider=Microsoft.Jet.OLEDB.4.0;Data Source=%s;", tempFile.Name())

	// Alternative connection string for newer Access versions
	connStrAlt := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};Dbq=%s;", tempFile.Name())

	var db *sql.DB
	var errOpen error

	// Try different connection strings
	db, errOpen = sql.Open("mssql", connStr)
	if errOpen != nil {
		// Try with alternative driver
		db, errOpen = sql.Open("access", connStrAlt)
		if errOpen != nil {
			// Try ODBC approach
			db, errOpen = sql.Open("odbc", connStrAlt)
			if errOpen != nil {
				// If all fail, provide helpful error
				return response, fmt.Errorf("failed to connect to MDB database: %w. Please ensure ODBC/Jet drivers are installed", errOpen)
			}
		}
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return response, fmt.Errorf("failed to connect to MDB database: %w", err)
	}

	// Get list of tables
	tables, err := h.getMDBTables(db)
	if err != nil {
		return response, fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		return response, fmt.Errorf("no tables found in MDB database")
	}

	// Use first table if not specified
	tableName := request.SheetName // Reuse SheetName field as table selector
	if tableName == "" {
		tableName = tables[0]
	}

	// Check if table exists
	tableExists := false
	for _, t := range tables {
		if t == tableName {
			tableExists = true
			break
		}
	}
	if !tableExists {
		return response, fmt.Errorf("table '%s' not found in MDB database. Available tables: %v", tableName, tables)
	}

	// Get column information and data
	columns, rows, totalRows, err := h.getMDBTableData(db, tableName, request)
	if err != nil {
		return response, fmt.Errorf("failed to read table data: %w", err)
	}

	response.Columns = columns
	response.Rows = rows
	response.RowCount = len(rows)
	response.TotalRows = totalRows
	response.Sheets = tables

	return response, nil
}

// getMDBTables retrieves all table names from MDB database
func (h *DataBrowserHandler) getMDBTables(db *sql.DB) ([]string, error) {
	// Query for table names
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := db.Query(query)
	if err != nil {
		// Try alternative query for older Access versions
		queryAlt := `
			SELECT Name 
			FROM MSysObjects 
			WHERE Type=1 AND Flags=0
			ORDER BY Name
		`
		rows, err = db.Query(queryAlt)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// getMDBTableData retrieves data from specified table
func (h *DataBrowserHandler) getMDBTableData(db *sql.DB, tableName string, request BrowseRequest) ([]string, [][]string, int64, error) {
	// First, get column information
	columnsQuery := fmt.Sprintf(`
		SELECT COLUMN_NAME, DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_NAME = '%s'
		ORDER BY ORDINAL_POSITION
	`, tableName)

	columnRows, err := db.Query(columnsQuery)
	if err != nil {
		// Try alternative approach
		columnRows, err = db.Query(fmt.Sprintf("SELECT TOP 1 * FROM [%s]", tableName))
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to get column info: %w", err)
		}
	}
	defer columnRows.Close()

	var columns []string
	for columnRows.Next() {
		var columnName, dataType string
		if err := columnRows.Scan(&columnName, &dataType); err != nil {
			// If the above fails, try to get columns from the actual query
			continue
		}
		columns = append(columns, columnName)
	}

	// If we couldn't get columns from schema, try to get them from a sample query
	if len(columns) == 0 {
		sampleQuery := fmt.Sprintf("SELECT TOP 1 * FROM [%s]", tableName)
		sampleRows, err := db.Query(sampleQuery)
		if err != nil {
			return nil, nil, 0, err
		}
		columns, _ = sampleRows.Columns()
		sampleRows.Close()
	}

	if len(columns) == 0 {
		return nil, nil, 0, fmt.Errorf("unable to determine table structure")
	}

	// Get total row count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM [%s]", tableName)
	var totalRows int64
	err = db.QueryRow(countQuery).Scan(&totalRows)
	if err != nil {
		totalRows = 0
	}

	// Get paginated data
	dataQuery := fmt.Sprintf("SELECT * FROM [%s] ORDER BY 1 OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
		tableName, request.Offset, request.MaxRows)

	// For older Access versions that don't support OFFSET/FETCH
	dataQueryAlt := fmt.Sprintf("SELECT * FROM [%s] ORDER BY 1", tableName)

	dataRows, err := db.Query(dataQuery)
	if err != nil {
		// Try without pagination for older versions
		dataRows, err = db.Query(dataQueryAlt)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("failed to query table data: %w", err)
		}
	}
	defer dataRows.Close()

	var rows [][]string
	processedRows := 0

	// Skip offset rows
	for request.Offset > 0 && dataRows.Next() && processedRows < request.Offset {
		processedRows++
	}

	// Process actual data rows
	for dataRows.Next() && len(rows) < request.MaxRows {
		// Create slice for row data
		rowData := make([]string, len(columns))

		// Scan into interface{} slice to handle different data types
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := dataRows.Scan(valuePtrs...); err != nil {
			// On scan error, fill with error indicators
			for i := range rowData {
				rowData[i] = "SCAN_ERROR"
			}
		} else {
			// Convert interface{} values to strings
			for i, value := range values {
				rowData[i] = h.convertInterfaceToString(value)
			}
		}

		rows = append(rows, rowData)
	}

	return columns, rows, totalRows, nil
}

// convertInterfaceToString converts various data types to strings for consistent output
func (h *DataBrowserHandler) convertInterfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (h *DataBrowserHandler) getExcelInfo(ctx context.Context, fileName string) ([]string, []string, int64, error) {
	reader, err := h.minioClient.DownloadFile(ctx, fileName)
	if err != nil {
		return nil, nil, 0, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, 0, err
	}

	wb, err := xlsx.OpenBinary(data)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get sheet names
	var sheetNames []string
	for _, sheet := range wb.Sheets {
		sheetNames = append(sheetNames, sheet.Name)
	}

	// Get info from first sheet
	var columns []string
	var rowCount int64
	if len(wb.Sheets) > 0 {
		sheet := wb.Sheets[0]
		err := sheet.ForEachRow(func(row *xlsx.Row) error {
			rowCount++
			if rowCount == 1 {
				// Get columns from first row
				var cols []string
				row.ForEachCell(func(cell *xlsx.Cell) error {
					cellValue, _ := cell.FormattedValue()
					cols = append(cols, cellValue)
					return nil
				})
				columns = cols
			}
			return nil
		})
		if err != nil {
			return nil, nil, 0, err
		}
	}

	return sheetNames, columns, rowCount, nil
}

func (h *DataBrowserHandler) getDataType(ext string) string {
	switch ext {
	case ".xlsx", ".xls", ".xlsm":
		return "excel"
	case ".csv":
		return "csv"
	case ".mdb", ".accdb":
		return "mdb"
	default:
		return "unknown"
	}
}

// getCSVInfo gets basic info about CSV files without processing all data
func (h *DataBrowserHandler) getCSVInfo(ctx context.Context, fileName string) ([]string, int64, error) {
	reader, err := h.minioClient.DownloadFile(ctx, fileName)
	if err != nil {
		return nil, 0, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}

	if len(data) == 0 {
		return []string{}, 0, nil
	}

	// Auto-detect delimiter
	detectedDelim := h.detectDelimiter(data)
	csvReader := csv.NewReader(bytes.NewReader(data))
	csvReader.Comma = detectedDelim
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	// Read just enough to get column count and row estimate
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, 0, err
	}

	if len(records) == 0 {
		return []string{}, 0, nil
	}

	return records[0], int64(len(records)), nil
}

// getMDBInfo gets basic info about MDB files without processing all data
func (h *DataBrowserHandler) getMDBInfo(ctx context.Context, fileName string) ([]string, []string, int64, error) {
	reader, err := h.minioClient.DownloadFile(ctx, fileName)
	if err != nil {
		return nil, nil, 0, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, 0, err
	}

	if len(data) == 0 {
		return []string{}, []string{}, 0, nil
	}

	// Create temporary file for MDB database
	tempFile, err := os.CreateTemp("", "tempdb_info_*.mdb")
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write MDB data to temp file
	if _, err := tempFile.Write(data); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open MDB database
	connStr := fmt.Sprintf("Provider=Microsoft.Jet.OLEDB.4.0;Data Source=%s;", tempFile.Name())
	connStrAlt := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};Dbq=%s;", tempFile.Name())

	var db *sql.DB
	db, err = sql.Open("mssql", connStr)
	if err != nil {
		db, err = sql.Open("access", connStrAlt)
		if err != nil {
			db, err = sql.Open("odbc", connStrAlt)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("failed to connect to MDB database: %w", err)
			}
		}
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to connect to MDB database: %w", err)
	}

	// Get list of tables
	tables, err := h.getMDBTables(db)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get tables: %w", err)
	}

	if len(tables) == 0 {
		return tables, []string{}, 0, nil
	}

	// Get column info from first table
	columns, _, totalRows, err := h.getMDBTableData(db, tables[0], BrowseRequest{MaxRows: 1})
	if err != nil {
		return tables, []string{}, 0, err
	}

	return tables, columns, totalRows, nil
}

func (h *DataBrowserHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *DataBrowserHandler) writeError(w http.ResponseWriter, message string, statusCode int, err error) {
	response := map[string]any{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
		log.Printf("Data Browser Error: %v", err)
	}

	h.writeJSON(w, statusCode, response)
}

// detectDelimiter tries to detect the most likely delimiter in CSV data
func (h *DataBrowserHandler) detectDelimiter(data []byte) rune {
	dataStr := string(data)
	delimiters := []struct {
		char  rune
		name  string
		count int
	}{
		{',', "comma", 0},
		{';', "semicolon", 0},
		{'\t', "tab", 0},
		{'|', "pipe", 0},
	}

	// Count occurrences of each delimiter in the first few lines
	lines := strings.Split(dataStr, "\n")
	sampleLines := 5
	if len(lines) < sampleLines {
		sampleLines = len(lines)
	}

	for i := 0; i < sampleLines && i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		for _, delim := range delimiters {
			delimiters[0].count += strings.Count(line, string(delim.char))
			delimiters[1].count += strings.Count(line, string(delimiters[1].char))
			delimiters[2].count += strings.Count(line, string(delimiters[2].char))
			delimiters[3].count += strings.Count(line, string(delimiters[3].char))
		}
	}

	// Find delimiter with highest count (excluding periods which are common in text)
	maxCount := 0
	bestDelim := ','
	for _, delim := range delimiters {
		if delim.count > maxCount {
			maxCount = delim.count
			bestDelim = delim.char
		}
	}

	return bestDelim
}

// detectHeaders tries to determine if the first row contains headers
func (h *DataBrowserHandler) detectHeaders(records [][]string) bool {
	if len(records) < 2 {
		return false
	}

	firstRow := records[0]
	secondRow := records[1]

	// Check if first row contains non-numeric values while second row has more numeric values
	firstRowNumeric := 0
	secondRowNumeric := 0

	for i := 0; i < len(firstRow) && i < len(secondRow); i++ {
		firstCol := strings.TrimSpace(firstRow[i])
		secondCol := strings.TrimSpace(secondRow[i])

		// Check if first value is numeric
		if h.isNumeric(firstCol) {
			firstRowNumeric++
		}

		// Check if second value is numeric
		if h.isNumeric(secondCol) {
			secondRowNumeric++
		}
	}

	// If first row has fewer numeric values than second row, it's likely headers
	return firstRowNumeric < secondRowNumeric || (len(firstRow) > 0 && !h.isNumeric(firstRow[0]))
}

// isNumeric checks if a string represents a number
func (h *DataBrowserHandler) isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	// Remove common numeric formatting characters
	original := s
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "%", "")

	// Try to convert to float
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	if err != nil {
		return false
	}

	// Ensure the cleaned string actually represents the full number
	// Check if removing non-numeric characters changed the string significantly
	nonNumericChars := 0
	for _, char := range original {
		if !((char >= '0' && char <= '9') || char == '.' || char == '-' || char == '+' ||
			char == ',' || char == '$' || char == '%') {
			nonNumericChars++
		}
	}

	// If there are too many non-numeric characters, it's probably not a pure number
	return nonNumericChars <= 2 // Allow for currency symbols and decimals
}

// streamCSVData streams CSV data in chunks for large files
func (h *DataBrowserHandler) streamCSVData(w http.ResponseWriter, r *http.Request, reader io.Reader, request BrowseRequest) {
	// Set headers for streaming response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx

	// Create streaming encoder
	encoder := json.NewEncoder(w)

	// Send initial response metadata
	firstChunk := map[string]interface{}{
		"success":     true,
		"message":     "Streaming CSV data",
		"data_type":   "csv",
		"file_name":   request.FileName,
		"streaming":   true,
		"has_headers": request.HasHeaders,
		"offset":      request.Offset,
		"chunk_size":  request.ChunkSize,
	}

	if err := encoder.Encode(firstChunk); err != nil {
		log.Printf("Failed to send initial chunk: %v", err)
		return
	}

	// Flush response to client
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Create CSV reader with auto-detected delimiter
	bufReader := bufio.NewReader(reader)
	peekBytes, err := bufReader.Peek(1024) // Read first KB for delimiter detection
	if err != nil && err != io.EOF {
		h.writeError(w, "Failed to peek file for delimiter detection", http.StatusInternalServerError, err)
		return
	}

	detectedDelim := h.detectDelimiter(peekBytes)

	// Reset reader and create CSV parser
	csvReader := csv.NewReader(bufReader)
	csvReader.Comma = detectedDelim
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	currentRow := int64(0)
	processedRows := 0
	var columns []string
	hasSentHeaders := false

	// Read and process rows in chunks
	chunk := make([][]string, 0, request.ChunkSize)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Send error chunk and continue
			errorChunk := map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("CSV parsing error at row %d: %v", currentRow+1, err),
				"row":     currentRow + 1,
			}
			encoder.Encode(errorChunk)
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			continue
		}

		currentRow++

		// Skip rows until offset is reached
		if currentRow <= int64(request.Offset) {
			if currentRow == 1 && !request.HasHeaders && request.AutoDetectHeaders && len(record) > 0 {
				// Store first row for header detection
				columns = record
			}
			continue
		}

		// Handle headers if this is the first data row
		if !hasSentHeaders && len(record) > 0 {
			if request.AutoDetectHeaders && len(columns) > 0 {
				// Use the stored first row as potential headers
				if h.detectHeaders([][]string{columns, record}) {
					request.HasHeaders = true

					// Send header information
					headerChunk := map[string]interface{}{
						"success":     true,
						"columns":     columns,
						"has_headers": true,
						"message":     "Headers auto-detected",
					}
					encoder.Encode(headerChunk)
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				} else {
					// Use current record as columns
					columns = make([]string, len(record))
					copy(columns, record)
				}
			} else if request.HasHeaders && !hasSentHeaders {
				// Use current record as headers
				columns = make([]string, len(record))
				copy(columns, record)

				// Send header information
				headerChunk := map[string]interface{}{
					"success":     true,
					"columns":     columns,
					"has_headers": true,
				}
				encoder.Encode(headerChunk)
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				hasSentHeaders = true
				continue // Skip this row as it's headers
			} else if len(columns) == 0 {
				// Use current record as columns
				columns = make([]string, len(record))
				copy(columns, record)
			}

			hasSentHeaders = true
		}

		// Skip if we've reached max rows
		if request.MaxRows > 0 && processedRows >= request.MaxRows {
			break
		}

		// Add to chunk
		chunk = append(chunk, record)
		processedRows++

		// Send chunk when it reaches the desired size
		if len(chunk) >= request.ChunkSize {
			dataChunk := map[string]interface{}{
				"success":   true,
				"data":      chunk,
				"row_count": len(chunk),
				"progress": map[string]interface{}{
					"processed":   processedRows,
					"current_row": currentRow,
				},
			}

			if err := encoder.Encode(dataChunk); err != nil {
				log.Printf("Failed to send data chunk: %v", err)
				return
			}

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			// Reset chunk
			chunk = chunk[:0]
		}
	}

	// Send final chunk with any remaining data
	if len(chunk) > 0 {
		finalChunk := map[string]interface{}{
			"success":   true,
			"data":      chunk,
			"row_count": len(chunk),
			"progress": map[string]interface{}{
				"processed":   processedRows,
				"current_row": currentRow,
			},
			"complete": true,
		}
		encoder.Encode(finalChunk)
	} else {
		// Send completion marker
		completionChunk := map[string]interface{}{
			"success":    true,
			"row_count":  processedRows,
			"total_rows": currentRow,
			"complete":   true,
			"message":    "Streaming completed",
		}
		encoder.Encode(completionChunk)
	}

	// Final flush
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
