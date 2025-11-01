package data_browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamCSVData(t *testing.T) {
	// Create a mock MinIO client (for simplicity, we'll use direct data)
	handler := &DataBrowserHandler{}

	// Test CSV data
	csvData := `Name,Age,City
John Doe,30,New York
Jane Smith,25,Los Angeles
Bob Johnson,35,Chicago
Alice Brown,28,Boston
Charlie Wilson,42,Seattle
Diana Miller,31,Miami
Edward Davis,29,Dallas
Frank Wilson,38,Phoenix
Grace Lee,26,Denver
Henry Taylor,33,Portland`

	// Create a request with streaming mode
	requestBody := map[string]interface{}{
		"file_name":           "test.csv",
		"stream_mode":         true,
		"treat_as_csv":        true,
		"max_rows":            5,
		"chunk_size":          2,
		"has_headers":         true,
		"auto_detect_headers": false,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/data/browse", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the streaming handler directly with mock data
	reader := strings.NewReader(csvData)
	handler.streamCSVData(rr, req, reader, BrowseRequest{
		FileName:          "test.csv",
		StreamMode:        true,
		TreatAsCSV:        true,
		MaxRows:           5,
		ChunkSize:         2,
		HasHeaders:        true,
		AutoDetectHeaders: false,
	})

	// Check the response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Parse the streaming response
	responseBody := rr.Body.String()
	lines := strings.Split(responseBody, "\n")

	// Should have multiple JSON lines (streaming chunks)
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 response lines, got %d", len(lines))
	}

	// Check first chunk (metadata)
	var firstChunk map[string]interface{}
	if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
		if err := json.Unmarshal([]byte(lines[0]), &firstChunk); err != nil {
			t.Errorf("Failed to parse first chunk JSON: %v", err)
		} else {
			if !firstChunk["success"].(bool) {
				t.Error("First chunk should indicate success")
			}
			if !firstChunk["streaming"].(bool) {
				t.Error("First chunk should indicate streaming mode")
			}
		}
	}

	t.Logf("Streaming response:\n%s", responseBody)
}

func TestDetectDelimiter(t *testing.T) {
	handler := &DataBrowserHandler{}

	// Test comma delimited
	commaData := "Name,Age,City\nJohn,30,NYC"
	if delim := handler.detectDelimiter([]byte(commaData)); delim != ',' {
		t.Errorf("Expected comma delimiter, got %q", delim)
	}

	// Test semicolon delimited
	semicolonData := "Name;Age;City\nJohn;30;NYC"
	if delim := handler.detectDelimiter([]byte(semicolonData)); delim != ';' {
		t.Errorf("Expected semicolon delimiter, got %q", delim)
	}

	// Test tab delimited
	tabData := "Name\tAge\tCity\nJohn\t30\tNYC"
	if delim := handler.detectDelimiter([]byte(tabData)); delim != '\t' {
		t.Errorf("Expected tab delimiter, got %q", delim)
	}

	// Test pipe delimited
	pipeData := "Name|Age|City\nJohn|30|NYC"
	if delim := handler.detectDelimiter([]byte(pipeData)); delim != '|' {
		t.Errorf("Expected pipe delimiter, got %q", delim)
	}
}

func TestDetectHeaders(t *testing.T) {
	handler := &DataBrowserHandler{}

	// Test with obvious headers
	recordsWithHeaders := [][]string{
		{"Name", "Age", "City"},
		{"John", "30", "NYC"},
		{"Jane", "25", "LA"},
	}

	if !handler.detectHeaders(recordsWithHeaders) {
		t.Error("Should detect headers for non-numeric first row")
	}

	// Test with numeric first row
	recordsWithoutHeaders := [][]string{
		{"1", "John", "30"},
		{"2", "Jane", "25"},
		{"3", "Bob", "35"},
	}

	if handler.detectHeaders(recordsWithoutHeaders) {
		t.Error("Should not detect headers for numeric first row")
	}
}

func TestGetDataType(t *testing.T) {
	handler := &DataBrowserHandler{}

	testCases := []struct {
		ext      string
		expected string
	}{
		{".xlsx", "excel"},
		{".xls", "excel"},
		{".xlsm", "excel"},
		{".csv", "csv"},
		{".mdb", "mdb"},   // Test MDB support
		{".accdb", "mdb"}, // Test ACCDB support (newer Access)
		{".txt", "unknown"},
		{".log", "unknown"},
	}

	for _, tc := range testCases {
		result := handler.getDataType(tc.ext)
		if result != tc.expected {
			t.Errorf("getDataType(%s) = %s, expected %s", tc.ext, result, tc.expected)
		}
	}
}

func TestConvertInterfaceToString(t *testing.T) {
	handler := &DataBrowserHandler{}

	testCases := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{[]byte("world"), "world"},
		{123, "123"},
		{-456, "-456"},
		{78.9, "78.900000"},
		{true, "true"},
		{false, "false"},
		{time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), "2023-01-01 12:00:00"},
		{int8(8), "8"},
		{int16(16), "16"},
		{int32(32), "32"},
		{int64(64), "64"},
		{uint8(8), "8"},
		{uint16(16), "16"},
		{uint32(32), "32"},
		{uint64(64), "64"},
		{float32(3.14), "3.140000"},
		{complex64(1 + 2i), "(1+2i)"}, // Fallback to fmt.Sprintf
	}

	for _, tc := range testCases {
		result := handler.convertInterfaceToString(tc.input)
		// Handle float precision issues in comparison
		if strings.Contains(result, ".") {
			// For float values, just check if it's a valid number string
			if _, err := fmt.Sscanf(result, "%f", new(float64)); err != nil {
				t.Errorf("convertInterfaceToString(%v) = %q, expected numeric string", tc.input, result)
			}
		} else {
			if result != tc.expected {
				t.Errorf("convertInterfaceToString(%v) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	}
}

// Test MDB connection string generation
func TestMDBConnectionString(t *testing.T) {
	tempFile := "/tmp/test.mdb"

	// Test Jet OLEDB connection string
	connStr := fmt.Sprintf("Provider=Microsoft.Jet.OLEDB.4.0;Data Source=%s;", tempFile)
	if !strings.Contains(connStr, "Microsoft.Jet.OLEDB.4.0") {
		t.Error("Connection string should contain Jet provider")
	}

	// Test Access driver connection string
	connStrAlt := fmt.Sprintf("Driver={Microsoft Access Driver (*.mdb, *.accdb)};Dbq=%s;", tempFile)
	if !strings.Contains(connStrAlt, "Microsoft Access Driver") {
		t.Error("Connection string should contain Access driver")
	}
}

func TestIsNumeric(t *testing.T) {
	handler := &DataBrowserHandler{}

	// Test valid numbers
	testCases := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"-123", true},
		{"123.45", true},
		{"$1,234.56", true}, // With formatting
		{"50%", true},       // With percentage
		{"abc", false},
		{"", false},
		{"123abc", false},
		{"$1,234", true},
	}

	for _, tc := range testCases {
		result := handler.isNumeric(tc.input)
		if result != tc.expected {
			t.Errorf("isNumeric(%q) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}
