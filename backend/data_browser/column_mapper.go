package data_browser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ColumnMismatch struct {
	ColumnName   string `json:"column_name"`
	MismatchType string `json:"mismatch_type"` // "missing", "extra", "type_mismatch", "case_diff"
	SourceType   string `json:"source_type,omitempty"`
	TargetType   string `json:"target_type,omitempty"`
	Severity     string `json:"severity"` // "error", "warning", "info"
}

type RowError struct {
	RowIndex     int         `json:"row_index"`
	ColumnName   string      `json:"column_name"`
	ErrorCode    string      `json:"error_code"`
	ErrorMsg     string      `json:"error_message"`
	SourceValue  interface{} `json:"source_value"`
	SuggestedFix string      `json:"suggested_fix,omitempty"`
}

type ConversionAttempt struct {
	Method string `json:"method"`
	Result string `json:"result"` // "success", "fail"
}

type ColumnTransform struct {
	SourceColumn  string `json:"source_column"`
	TargetColumn  string `json:"target_column"`
	TransformType string `json:"transform_type"` // "direct", "convert", "extract", "format"
	TransformRule string `json:"transform_rule,omitempty"`
	DefaultValue  string `json:"default_value,omitempty"`
}

type ColumnMapper struct {
	sourceColumns      []string
	targetColumns      []string
	columnMap          map[string]string // source -> target
	mismatches         []ColumnMismatch
	caseSensitive      bool
	autoTypeConversion bool
	transformRules     []ColumnTransform
}

func NewColumnMapper(sourceColumns, targetColumns []string, caseSensitive bool) *ColumnMapper {
	mapper := &ColumnMapper{
		sourceColumns:  sourceColumns,
		targetColumns:  targetColumns,
		columnMap:      make(map[string]string),
		caseSensitive:  caseSensitive,
		transformRules: make([]ColumnTransform, 0),
		mismatches:     make([]ColumnMismatch, 0),
	}

	mapper.generateMapping()
	return mapper
}

func (cm *ColumnMapper) generateMapping() {
	targetColMap := cm.createColumnMap(cm.targetColumns)

	for _, sourceCol := range cm.sourceColumns {
		sourceColKey := cm.normalizeColumnName(sourceCol)
		targetCol, exists := targetColMap[sourceColKey]

		if exists {
			// Found exact or case-insensitive match
			cm.columnMap[sourceCol] = targetCol
		} else {
			// Try fuzzy matching
			if match := cm.findFuzzyMatch(sourceCol, cm.targetColumns); match != "" {
				cm.columnMap[sourceCol] = match
				cm.mismatches = append(cm.mismatches, ColumnMismatch{
					ColumnName:   sourceCol,
					MismatchType: "case_diff",
					SourceType:   "VARCHAR",
					TargetType:   "VARCHAR",
					Severity:     "info",
				})
			} else {
				// Extra column in source
				cm.mismatches = append(cm.mismatches, ColumnMismatch{
					ColumnName:   sourceCol,
					MismatchType: "extra",
					SourceType:   "VARCHAR",
					TargetType:   "",
					Severity:     "warning",
				})
			}
		}
	}

	// Check for missing target columns
	sourceColMap := cm.createColumnMap(cm.sourceColumns)
	for _, targetCol := range cm.targetColumns {
		targetColKey := cm.normalizeColumnName(targetCol)
		if _, exists := sourceColMap[targetColKey]; !exists {
			cm.mismatches = append(cm.mismatches, ColumnMismatch{
				ColumnName:   targetCol,
				MismatchType: "missing",
				SourceType:   "",
				TargetType:   "VARCHAR",
				Severity:     "warning",
			})
		}
	}
}

func (cm *ColumnMapper) createColumnMap(columns []string) map[string]string {
	colMap := make(map[string]string)
	for _, col := range columns {
		colMap[cm.normalizeColumnName(col)] = col
	}
	return colMap
}

func (cm *ColumnMapper) normalizeColumnName(colName string) string {
	if cm.caseSensitive {
		return colName
	}
	return strings.ToLower(colName)
}

func (cm *ColumnMapper) findFuzzyMatch(sourceCol string, targetColumns []string) string {
	sourceLower := strings.ToLower(sourceCol)

	// Remove common prefixes/suffixes for matching
	cleanSource := cm.cleanColumnName(sourceLower)

	for _, targetCol := range targetColumns {
		targetLower := strings.ToLower(targetCol)
		cleanTarget := cm.cleanColumnName(targetLower)

		if cleanSource == cleanTarget {
			return targetCol
		}
	}

	// Try Levenshtein distance for close matches
	bestMatch := ""
	bestDistance := len(sourceCol) // Max possible distance

	for _, targetCol := range targetColumns {
		targetLower := strings.ToLower(targetCol)
		distance := cm.levenshteinDistance(cleanSource, cm.cleanColumnName(targetLower))

		if distance < bestDistance && distance <= 2 { // Allow up to 2 character differences
			bestDistance = distance
			bestMatch = targetCol
		}
	}

	return bestMatch
}

func (cm *ColumnMapper) cleanColumnName(colName string) string {
	// Remove common prefixes
	prefixes := []string{"col_", "column_", "field_", "f_"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(colName, prefix) {
			colName = strings.TrimPrefix(colName, prefix)
			break
		}
	}

	// Remove common suffixes
	suffixes := []string{"_col", "_column", "_field", "_f"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(colName, suffix) {
			colName = strings.TrimSuffix(colName, suffix)
			break
		}
	}

	// Remove numbers at end
	re := regexp.MustCompile(`\d+$`)
	colName = re.ReplaceAllString(colName, "")

	return strings.Trim(colName, "_")
}

func (cm *ColumnMapper) levenshteinDistance(s1, s2 string) int {
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	if len(s1) == 0 {
		return len(s2)
	}

	prevRow := make([]int, len(s2)+1)
	for i := 0; i <= len(s1); i++ {
		currentRow := make([]int, len(s2)+1)
		currentRow[0] = i

		for j := 1; j <= len(s2); j++ {
			insertCost := 1
			deleteCost := 1
			replaceCost := 2

			var cost int
			if s1[i-1] == s2[j-1] {
				cost = 0
			} else {
				cost = replaceCost
			}

			minCost := prevRow[j] + insertCost
			if currentRow[j-1] < minCost {
				minCost = currentRow[j-1] + deleteCost
			}
			if prevRow[j-1]+cost < minCost {
				minCost = prevRow[j-1] + cost
			}

			currentRow[j] = minCost
		}

		prevRow = currentRow
	}

	return prevRow[len(s2)]
}

func (cm *ColumnMapper) MapRow(row []string, targetColumns []string) (map[string]interface{}, []RowError) {
	result := make(map[string]interface{})
	var errors []RowError

	for sourceColIndex, value := range row {
		if sourceColIndex >= len(cm.sourceColumns) {
			continue // Skip extra values
		}

		sourceCol := cm.sourceColumns[sourceColIndex]
		targetCol, mapped := cm.columnMap[sourceCol]

		if !mapped {
			// Column not mapped, skip or set to null
			continue
		}

		// Apply data type conversion if enabled
		convertedValue, err := cm.convertValue(value, targetCol)
		if err != nil {
			errors = append(errors, RowError{
				RowIndex:     0, // Will be set by caller
				ColumnName:   targetCol,
				ErrorCode:    "CONVERSION_ERROR",
				ErrorMsg:     err.Error(),
				SourceValue:  value,
				SuggestedFix: "Check data format or set to NULL",
			})
			result[targetCol] = nil
		} else {
			result[targetCol] = convertedValue
		}
	}

	return result, errors
}

func (cm *ColumnMapper) convertValue(value interface{}, targetColumn string) (interface{}, error) {
	if value == nil || value == "" {
		return nil, nil
	}

	// Try to infer data type and convert
	strValue := fmt.Sprintf("%v", value)

	// Date conversion attempts
	if cm.isDateColumn(targetColumn) {
		formats := []string{
			time.RFC3339,
			"2006-01-02",
			"01/02/2006",
			"02/01/2006",
			"Jan 2, 2006",
		}

		for _, format := range formats {
			if parsed, err := time.Parse(format, strValue); err == nil {
				return parsed, nil
			}
		}
	}

	// Number conversion attempts
	if cm.isNumericColumn(targetColumn) {
		// Remove currency symbols, commas, etc.
		cleanNumber := strings.Map(func(r rune) rune {
			if strings.ContainsRune("0123456789.-", r) {
				return r
			}
			return -1
		}, strValue)

		if cleaned, err := strconv.ParseFloat(cleanNumber, 64); err == nil {
			return cleaned, nil
		}
	}

	// Boolean conversion
	if cm.isBooleanColumn(targetColumn) {
		lowerStr := strings.ToLower(strValue)
		if lowerStr == "true" || lowerStr == "1" || lowerStr == "yes" || lowerStr == "y" {
			return true, nil
		}
		if lowerStr == "false" || lowerStr == "0" || lowerStr == "no" || lowerStr == "n" {
			return false, nil
		}
	}

	// Default to string
	return strValue, nil
}

func (cm *ColumnMapper) isDateColumn(columnName string) bool {
	colLower := strings.ToLower(columnName)
	dateKeywords := []string{"date", "time", "created", "updated", "timestamp", "modified", "birth"}

	for _, keyword := range dateKeywords {
		if strings.Contains(colLower, keyword) {
			return true
		}
	}
	return false
}

func (cm *ColumnMapper) isNumericColumn(columnName string) bool {
	colLower := strings.ToLower(columnName)
	numericKeywords := []string{"amount", "price", "cost", "total", "count", "quantity", "number", "id", "age", "score", "rating"}

	for _, keyword := range numericKeywords {
		if strings.Contains(colLower, keyword) {
			return true
		}
	}
	return false
}

func (cm *ColumnMapper) isBooleanColumn(columnName string) bool {
	colLower := strings.ToLower(columnName)
	booleanKeywords := []string{"active", "enabled", "disabled", "flag", "bool", "is_", "has_", "verified"}

	for _, keyword := range booleanKeywords {
		if strings.Contains(colLower, keyword) {
			return true
		}
	}
	return false
}

func (cm *ColumnMapper) GetMismatches() []ColumnMismatch {
	return cm.mismatches
}

func (cm *ColumnMapper) GetColumnMap() map[string]string {
	return cm.columnMap
}

func (cm *ColumnMapper) AddTransformRule(transform ColumnTransform) {
	cm.transformRules = append(cm.transformRules, transform)
}
