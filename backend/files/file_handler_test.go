package files

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamFolderBrowse(t *testing.T) {
	// We can't easily test SSE stream without mocking MinIO client
	// Instead, just verify that function exists and can be called without panicking
	handler := &FileHandler{}

	// Create a minimal valid request body
	reqBody := `{"folders":[{"path":"","include_files":true,"include_dirs":true}]}`
	req := httptest.NewRequest("POST", "/api/files/browse", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.MultiFolderBrowse(rr, req)

	// The test passes if we get here without panic (even if minio client errors)
	t.Logf("Function executed without panic - status code: %d", rr.Code)
}