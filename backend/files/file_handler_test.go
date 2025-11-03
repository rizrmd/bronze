package files

import (
	"net/http/httptest"
	"testing"
)

func TestStreamFolderBrowse(t *testing.T) {
	// We can't easily test the SSE stream without mocking the entire HTTP response
	// Instead, let's test that the SSE headers are set correctly
	handler := &FileHandler{}

	// Test SSE headers
	req := httptest.NewRequest("POST", "/api/files/browse?stream=sse", nil)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.streamFolderBrowse(rr, req)

	// Check SSE headers
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", contentType)
	}

	if cacheControl := rr.Header().Get("Cache-Control"); cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control no-cache, got %s", cacheControl)
	}

	t.Logf("SSE Headers Test Passed - Content-Type: %s, Cache-Control: %s", 
		rr.Header().Get("Content-Type"), 
		rr.Header().Get("Cache-Control"))
}