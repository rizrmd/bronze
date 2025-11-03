// True streaming SSE implementation for file browsing
package files

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
)

func (h *FileHandler) streamFolderBrowseRealtime(w http.ResponseWriter, r *http.Request) {
	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Mutex for thread-safe flushing
	var flushMutex sync.Mutex
	safeFlush := func() {
		flushMutex.Lock()
		defer flushMutex.Unlock()
		flusher.Flush()
	}

	var req MultiFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSSEError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	// Track goroutines for proper cleanup
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Send connected event
	h.writeSSEEvent(w, "connected", `{"status":"connected"}`)
	safeFlush()

	// Process each folder with true streaming
	for _, folderReq := range req.Folders {
		wg.Add(1)
		go func(folderReq FolderRequest) {
			defer wg.Done()
			h.streamFolderContents(ctx, w, folderReq, safeFlush)
		}(folderReq)
	}

	// Wait for all goroutines to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(done)
	}()

	// Keep connection alive
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-keepalive.C:
			h.writeSSEEvent(w, "keepalive", `{"status":"alive"}`)
			safeFlush()
		case <-done:
			// All folders processed, send completion and return
			h.writeSSEEvent(w, "complete", `{"status":"all_folders_completed"}`)
			safeFlush()
			return
		case <-ctx.Done():
			h.writeSSEEvent(w, "closed", `{"status":"connection_closed"}`)
			return
		}
	}
}

// Stream folder contents in real-time as they're discovered
func (h *FileHandler) streamFolderContents(ctx context.Context, w http.ResponseWriter, folderReq FolderRequest, safeFlush func()) {
	// Add panic recovery to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			h.writeSSEError(w, "Panic in folder processing", http.StatusInternalServerError, fmt.Errorf("%v", r))
			safeFlush()
		}
	}()

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return
	default:
	}

	path := strings.TrimPrefix(folderReq.Path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Send folder start event
	h.writeSSEEvent(w, "folder_start", fmt.Sprintf(`{"path":"%s","status":"processing"}`, folderReq.Path))
	safeFlush()

	// Use MinIO's ListFiles method for streaming with smaller limit for responsiveness
	objects, err := h.minioClient.ListFiles(ctx, path, 500) // Reduced from 1000
	if err != nil {
		h.writeSSEError(w, fmt.Sprintf("Error listing %s", path), http.StatusInternalServerError, err)
		return
	}
	
	fileCount := 0
	dirCount := 0
	totalSize := int64(0)
	
	for _, obj := range objects {
		// Check for context cancellation before processing each item
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Skip the folder marker itself (the requested path)
		if obj.Key == path {
			continue
		}

		// Send each file/directory as individual SSE event
		eventData := map[string]interface{}{
			"path":         obj.Key,
			"size":         obj.Size,
			"lastModified": obj.LastModified.Format(time.RFC3339),
			"etag":         obj.ETag,
		}

		isDirectory := strings.HasSuffix(obj.Key, "/") && obj.Size == 0
		
		if isDirectory {
			// Count items in this directory
			itemCount := h.countItemsInFolder(ctx, obj.Key)
			
			dirCount++
			eventData["type"] = "directory"
			// Properly decode URL-encoded folder names
			decodedKey, _ := url.PathUnescape(obj.Key)
			eventData["name"] = filepath.Base(strings.TrimSuffix(decodedKey, "/"))
			eventData["size"] = int64(itemCount) // Show item count as size
		} else {
			fileCount++
			totalSize += obj.Size
			eventData["type"] = "file"
			eventData["name"] = filepath.Base(obj.Key)
			eventData["contentType"] = h.getContentType(obj.Key)
		}

		jsonData, _ := json.Marshal(eventData)
		h.writeSSEEvent(w, "item", string(jsonData))
		
		// Flush immediately for each item
		safeFlush()

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	// Send folder completion summary
	completionData := map[string]interface{}{
		"path":        folderReq.Path,
		"status":      "completed",
		"fileCount":   fileCount,
		"dirCount":    dirCount,
		"totalSize":   totalSize,
		"totalItems":  fileCount + dirCount,
	}
	
	jsonData, _ := json.Marshal(completionData)
	h.writeSSEEvent(w, "folder_complete", string(jsonData))
	
	safeFlush()
}

// Count total items (files + subdirectories) in a folder
func (h *FileHandler) countItemsInFolder(ctx context.Context, folderPath string) int {
	count := 0
	
	// Decode URL-encoded folder path
	decodedPath, _ := url.PathUnescape(folderPath)
	
	// Use MinIO client to list items in this folder (non-recursive)
	objectsCh := h.minioClient.GetClient().ListObjects(ctx, h.minioClient.GetBucketName(), minio.ListObjectsOptions{
		Prefix:    decodedPath,
		Recursive: false, // Only direct children
	})
	
	for obj := range objectsCh {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return count // Return current count if cancelled
		default:
		}
		
		if obj.Err != nil {
			continue
		}
		
		// Skip the folder marker itself (ending with / and same path)
		if obj.Key == folderPath {
			continue
		}
		
		count++
	}
	
	return count
}

// Helper to get content type from file extension
func (h *FileHandler) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt", ".md":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".xlsx", ".xls":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
