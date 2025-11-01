package data_browser

import (
	"fmt"
	"sort"
	"strings"
)

type SchemaMerger struct {
	files         []FileInfo
	mergedColumns []string
	columnTypes   map[string]string
	resolution    string // "merge", "first_file", "manual"
}

type FileInfo struct {
	FileName string   `json:"file_name"`
	Columns  []string `json:"columns"`
	RowCount int64    `json:"row_count"`
	DataType string   `json:"data_type"`
}

type MergedSchema struct {
	Columns     []string          `json:"columns"`
	ColumnTypes map[string]string `json:"column_types"`
	SourceFiles []FileInfo        `json:"source_files"`
	TotalRows   int64             `json:"total_rows"`
	Conflicts   []ColumnConflict  `json:"conflicts,omitempty"`
}

type ColumnConflict struct {
	ColumnName   string       `json:"column_name"`
	ConflictType string       `json:"conflict_type"` // "name_diff", "type_diff", "case_diff"
	Files        []FileColumn `json:"files"`
	Resolution   string       `json:"resolution,omitempty"`
}

type FileColumn struct {
	FileName   string `json:"file_name"`
	ColumnName string `json:"column_name"`
	DataType   string `json:"data_type"`
}

const (
	ResolutionMerge  = "merge"
	ResolutionFirst  = "first_file"
	ResolutionManual = "manual"
)

func NewSchemaMerger(resolution string) *SchemaMerger {
	return &SchemaMerger{
		columnTypes: make(map[string]string),
		resolution:  resolution,
	}
}

func (sm *SchemaMerger) MergeSchemas(files []FileInfo) (*MergedSchema, error) {
	sm.files = files

	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided for schema merging")
	}

	switch sm.resolution {
	case ResolutionFirst:
		return sm.mergeWithFirstFile(files)
	case ResolutionMerge:
		return sm.mergeWithUnion(files)
	case ResolutionManual:
		return sm.mergeWithManualResolution(files)
	default:
		return sm.mergeWithUnion(files) // Default to merge
	}
}

func (sm *SchemaMerger) mergeWithUnion(files []FileInfo) (*MergedSchema, error) {
	// Create union of all columns
	allColumns := make(map[string]bool)
	conflicts := make(map[string][]FileColumn)

	for _, file := range files {
		for _, col := range file.Columns {
			colKey := strings.ToLower(col)

			if existing := allColumns[colKey]; !existing {
				allColumns[colKey] = true
			} else {
				// Found potential conflict
				if _, exists := conflicts[colKey]; !exists {
					conflicts[colKey] = []FileColumn{}
				}

				conflicts[colKey] = append(conflicts[colKey], FileColumn{
					FileName:   file.FileName,
					ColumnName: col,
					DataType:   sm.inferType(file, col),
				})
			}
		}
	}

	// Detect conflicts and create suggestions
	var columnConflicts []ColumnConflict
	for colKey, fileCols := range conflicts {
		if len(fileCols) > 1 {
			conflict := ColumnConflict{
				ColumnName:   colKey,
				ConflictType: sm.detectConflictType(fileCols),
				Files:        fileCols,
				Resolution:   "use_first_occurrence",
			}
			columnConflicts = append(columnConflicts, conflict)
		}
	}

	// Create merged column list
	var mergedColumns []string
	for col := range allColumns {
		mergedColumns = append(mergedColumns, col)
	}

	// Sort columns alphabetically for consistency
	sort.Strings(mergedColumns)

	// Infer types for merged columns
	for _, col := range mergedColumns {
		sm.columnTypes[col] = sm.inferTypeFromFiles(files, col)
	}

	// Calculate total rows
	var totalRows int64
	for _, file := range files {
		totalRows += file.RowCount
	}

	return &MergedSchema{
		Columns:     mergedColumns,
		ColumnTypes: sm.columnTypes,
		SourceFiles: files,
		TotalRows:   totalRows,
		Conflicts:   columnConflicts,
	}, nil
}

func (sm *SchemaMerger) mergeWithFirstFile(files []FileInfo) (*MergedSchema, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	firstFile := files[0]
	mergedColumns := make([]string, len(firstFile.Columns))
	copy(mergedColumns, firstFile.Columns)

	// Check for columns in other files not in first file
	var conflicts []ColumnConflict
	for i := 1; i < len(files); i++ {
		file := files[i]
		for _, col := range file.Columns {
			if !sm.containsColumn(mergedColumns, col) {
				conflict := ColumnConflict{
					ColumnName:   col,
					ConflictType: "missing_in_first_file",
					Files: []FileColumn{
						{FileName: firstFile.FileName, ColumnName: col, DataType: "VARCHAR"},
						{FileName: file.FileName, ColumnName: col, DataType: "VARCHAR"},
					},
					Resolution: "exclude_column",
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	// Infer types from first file
	for _, col := range mergedColumns {
		sm.columnTypes[col] = sm.inferType(firstFile, col)
	}

	var totalRows int64
	for _, file := range files {
		totalRows += file.RowCount
	}

	return &MergedSchema{
		Columns:     mergedColumns,
		ColumnTypes: sm.columnTypes,
		SourceFiles: files,
		TotalRows:   totalRows,
		Conflicts:   conflicts,
	}, nil
}

func (sm *SchemaMerger) mergeWithManualResolution(files []FileInfo) (*MergedSchema, error) {
	// For manual resolution, detect all conflicts and return them for user resolution
	merged, err := sm.mergeWithUnion(files)
	if err != nil {
		return nil, err
	}

	// Mark all conflicts as requiring manual resolution
	for i := range merged.Conflicts {
		merged.Conflicts[i].Resolution = "manual_resolution_required"
	}

	return merged, nil
}

func (sm *SchemaMerger) containsColumn(columns []string, target string) bool {
	for _, col := range columns {
		if strings.EqualFold(col, target) {
			return true
		}
	}
	return false
}

func (sm *SchemaMerger) detectConflictType(fileCols []FileColumn) string {
	if len(fileCols) < 2 {
		return "unknown"
	}

	// Check if all columns have same name (case-insensitive)
	lowerNames := make(map[string]int)
	for _, col := range fileCols {
		lowerNames[strings.ToLower(col.ColumnName)]++
	}

	if len(lowerNames) == 1 {
		return "case_diff"
	}

	// Check if same name appears in different formats
	names := make(map[string]int)
	for _, col := range fileCols {
		names[col.ColumnName]++
	}

	if len(names) == 1 {
		return "format_diff"
	}

	return "name_diff"
}

func (sm *SchemaMerger) inferType(file FileInfo, columnName string) string {
	// Simple type inference - could be enhanced with actual data analysis
	colLower := strings.ToLower(columnName)

	// Numeric columns
	numericKeywords := []string{"id", "count", "amount", "price", "total", "age", "score", "rating", "quantity", "number"}
	for _, keyword := range numericKeywords {
		if strings.Contains(colLower, keyword) {
			return "BIGINT"
		}
	}

	// Date columns
	dateKeywords := []string{"date", "time", "created", "updated", "modified", "timestamp", "birth", "expires"}
	for _, keyword := range dateKeywords {
		if strings.Contains(colLower, keyword) {
			return "TIMESTAMP"
		}
	}

	// Boolean columns
	booleanKeywords := []string{"active", "enabled", "disabled", "flag", "bool", "is_", "has_", "verified"}
	for _, keyword := range booleanKeywords {
		if strings.Contains(colLower, keyword) {
			return "BOOLEAN"
		}
	}

	// Text columns with size limits
	textKeywords := []string{"name", "description", "comment", "notes", "email", "phone", "address"}
	for _, keyword := range textKeywords {
		if strings.Contains(colLower, keyword) {
			return "VARCHAR(255)"
		}
	}

	// Default text column
	return "VARCHAR(255)"
}

func (sm *SchemaMerger) inferTypeFromFiles(files []FileInfo, columnName string) string {
	// Find all files that contain this column and get consensus type
	var types []string
	for _, file := range files {
		if sm.containsColumn(file.Columns, columnName) {
			fileType := sm.inferType(file, columnName)
			types = append(types, fileType)
		}
	}

	if len(types) == 0 {
		return "VARCHAR(255)"
	}

	// Return most specific type (priority: BOOLEAN > BIGINT > TIMESTAMP > VARCHAR)
	for _, typeStr := range []string{"BOOLEAN", "BIGINT", "TIMESTAMP", "DECIMAL", "VARCHAR"} {
		for _, t := range types {
			if strings.HasPrefix(t, typeStr) {
				return t
			}
		}
	}

	return types[0] // Default to first occurrence
}

func (sm *SchemaMerger) GenerateColumnMapping(sourceColumns []string, targetSchema MergedSchema) (*ColumnMapper, error) {
	mapper := NewColumnMapper(sourceColumns, targetSchema.Columns, false)

	// Add inferred type mappings
	for _, targetCol := range targetSchema.Columns {
		if targetType, exists := targetSchema.ColumnTypes[targetCol]; exists {
			mapper.AddTransformRule(ColumnTransform{
				SourceColumn:  targetCol,
				TargetColumn:  targetCol,
				TransformType: "convert",
				TransformRule: targetType,
			})
		}
	}

	return mapper, nil
}
