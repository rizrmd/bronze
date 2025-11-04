package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bronze-backend/jobs"
	"bronze-backend/storage"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
)

type FileHandler struct {
	minioClient *storage.MinIOClient
	processor   interface {
		ProcessJob(ctx context.Context, job *jobs.Job) jobs.JobResult
	}
	jobQueue *jobs.JobQueue
}

func NewFileHandler(minioClient *storage.MinIOClient, fileProcessor interface {
	ProcessJob(ctx context.Context, job *jobs.Job) jobs.JobResult
}) *FileHandler {
	return &FileHandler{
		minioClient: minioClient,
		processor:   fileProcessor,
	}
}

func NewFileHandlerWithQueue(minioClient *storage.MinIOClient, fileProcessor interface {
	ProcessJob(ctx context.Context, job *jobs.Job) jobs.JobResult
}, jobQueue *jobs.JobQueue) *FileHandler {
	return &FileHandler{
		minioClient: minioClient,
		processor:   fileProcessor,
		jobQueue:    jobQueue,
	}
}

// Multi-folder request for browsing multiple directories at once
type MultiFolderRequest struct {
	Folders []FolderRequest `json:"folders"`
	Limit   int             `json:"limit,omitempty"`
}

// Individual folder request with options
type FolderRequest struct {
	Path         string `json:"path"`                   // Folder path to browse
	IncludeFiles bool   `json:"include_files"`          // Include files in response
	IncludeDirs  bool   `json:"include_dirs"`           // Include directories in response
	Recursive    bool   `json:"recursive"`              // Include subdirectories
	MaxDepth     int    `json:"max_depth,omitempty"`    // Max recursion depth (if recursive)
	IncludeMetadata bool `json:"include_metadata,omitempty"` // Include file counts and sizes for directories
}

// Multi-folder response with rich metadata
type MultiFolderResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Folders map[string]FolderResult `json:"folders"` // path -> result mapping
}

// Individual folder result with comprehensive information
type FolderResult struct {
	Path         string                  `json:"path"`
	Directories  []DirectoryInfo         `json:"directories,omitempty"`
	Files        []FileInfo              `json:"files,omitempty"`
	TotalCount   int                     `json:"total_count"`
	FileCount    int                     `json:"file_count"`
	DirCount     int                     `json:"dir_count"`
	Size         int64                   `json:"total_size_bytes"`
	LastModified string                  `json:"last_modified"`
	Subfolders   map[string]*FolderResult `json:"subfolders,omitempty"` // recursive results
}

// Enhanced directory information
type DirectoryInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	LastModified string    `json:"last_modified"`
	FileCount    int       `json:"file_count,omitempty"`     // optional metadata
	Size         int64     `json:"size,omitempty"`           // total size of files inside
}

// Enhanced file information
type FileInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	ContentType  string `json:"content_type,omitempty"`
	ETag         string `json:"etag,omitempty"`
}

// Legacy types for backward compatibility
type BatchListRequest struct {
	Prefixes []string `json:"prefixes"`
	Limit    int      `json:"limit,omitempty"`
}

type BatchListResponse struct {
	Success bool                          `json:"success"`
	Files   map[string][]minio.ObjectInfo `json:"files"`
	Message string                        `json:"message,omitempty"`
}

func (h *FileHandler) BatchListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check bucket status first
	log.Printf("BatchListFiles handler: checking bucket status")
	bucketOk, bucketMsg := h.checkBucketStatus()
	log.Printf("BatchListFiles handler: bucketOk=%v, bucketMsg=%s", bucketOk, bucketMsg)
	if !bucketOk {
		h.writeError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	var req BatchListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	// Set default limit
	limit := 1000
	if req.Limit > 0 {
		limit = req.Limit
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Fetch files for each prefix in parallel
	results := make(map[string][]minio.ObjectInfo)
	resultChan := make(chan struct {
		prefix string
		files  []minio.ObjectInfo
		err    error
	}, len(req.Prefixes))

	// Limit concurrent goroutines to prevent resource exhaustion
	maxConcurrency := 10
	if len(req.Prefixes) < maxConcurrency {
		maxConcurrency = len(req.Prefixes)
	}
	
	semaphore := make(chan struct{}, maxConcurrency)
	
	// Start goroutines for each prefix with concurrency control
	for i, prefix := range req.Prefixes {
		go func(idx int, p string) {
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release
			
			files, err := h.minioClient.ListFiles(ctx, p, limit)
			resultChan <- struct {
				prefix string
				files  []minio.ObjectInfo
				err    error
			}{prefix: p, files: files, err: err}
		}(i, prefix)
	}

	// Collect results
	for i := 0; i < len(req.Prefixes); i++ {
		result := <-resultChan
		if result.err != nil {
			log.Printf("Error fetching files for prefix '%s': %v", result.prefix, result.err)
			results[result.prefix] = []minio.ObjectInfo{}
		} else {
			results[result.prefix] = result.files
		}
	}

	response := BatchListResponse{
		Success: true,
		Files:   results,
		Message: fmt.Sprintf("Successfully fetched files for %d prefixes", len(req.Prefixes)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type FileListResponse struct {
	Success bool                       `json:"success"`
	Message string                     `json:"message"`
	Files   []storage.FileInfoResponse `json:"files"`
	Count   int                        `json:"count"`
}

type FileInfoResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	File    storage.FileInfoDetail `json:"file"`
}

type DeleteResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Deleted []string `json:"deleted,omitempty"`
	Count   int      `json:"count,omitempty"`
}

type BucketListResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Buckets []BucketInfoResponse `json:"buckets"`
	Count   int                  `json:"count"`
}

type BucketInfoResponse struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
}

type SetBucketResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Bucket  string `json:"bucket"`
}

type UploadResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ObjectName string `json:"object_name"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag"`
}

func (h *FileHandler) MultiFolderBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// All requests to /api/files/browse return SSE streams
	h.streamFolderBrowseRealtime(w, r)
}

// SSE streaming for folder browsing
func (h *FileHandler) streamFolderBrowse(w http.ResponseWriter, r *http.Request) {
	// Check bucket status first
	bucketOk, bucketMsg := h.checkBucketStatus()
	if !bucketOk {
		h.writeSSEError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	// Parse request body
	var req MultiFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSSEError(w, "Invalid JSON", http.StatusBadRequest, err)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	// Send initial connection event
	h.writeSSEEvent(w, "connected", `{"status":"connected"}`)
	
	// Create a flusher for real-time updates
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Process folders with streaming updates
	resultChan := make(chan struct {
		path   string
		result FolderResult
		err    error
	}, len(req.Folders))

	// Process folders in parallel with controlled concurrency
	maxConcurrency := 10
	if len(req.Folders) < maxConcurrency {
		maxConcurrency = len(req.Folders)
	}
	
	semaphore := make(chan struct{}, maxConcurrency)
	completed := make(chan string, len(req.Folders))
	results := make(map[string]FolderResult)

	// Start goroutines for each folder
	for i, folderReq := range req.Folders {
		go func(idx int, folderReq FolderRequest) {
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			result, err := h.processFolder(ctx, folderReq, 1000)
			resultChan <- struct {
				path   string
				result FolderResult
				err    error
			}{path: folderReq.Path, result: result, err: error(err)}
		}(i, folderReq)
	}

	// Stream results as they complete
	go func() {
		for i := 0; i < len(req.Folders); i++ {
			select {
			case result := <-resultChan:
				if result.err != nil {
					h.writeSSEError(w, fmt.Sprintf("Error processing %s", result.path), http.StatusInternalServerError, result.err)
				} else {
					// Create directory listing from existing folder result
					var items []map[string]interface{}
					
					// Add directories
					for _, dir := range result.result.Directories {
						// Extract just the directory name from path
						dirName := dir.Name
						if dirName == "" {
							parts := strings.Split(strings.TrimSuffix(dir.Path, "/"), "/")
							if len(parts) > 0 {
								dirName = parts[len(parts)-1]
							}
						}
						if dirName != "" {
							items = append(items, map[string]interface{}{
								"name": dirName,
								"type": "dir",
							})
						}
					}
					
					// Add files
					for _, file := range result.result.Files {
						// Use the Name field directly
						fileName := file.Name
						if fileName != "" {
							items = append(items, map[string]interface{}{
								"name": fileName,
								"type": "file",
							})
						}
					}
					
					// Create folder_start event with directory listing
					folderStartData := map[string]interface{}{
						"path":   result.path,
						"status": "processing",
						"items":  items,
					}
					folderStartJSON, _ := json.Marshal(folderStartData)
					h.writeSSEEvent(w, "folder_start", string(folderStartJSON))
					
					// Stream folder metadata
					folderJSON, _ := json.Marshal(result.result)
					h.writeSSEEvent(w, "folder_data", string(folderJSON))
					
					// Send folder complete event
					fileCount := result.result.FileCount + result.result.DirCount
					h.writeSSEEvent(w, "folder_complete", fmt.Sprintf(`{"path":"%s","status":"completed","items":%d}`, result.path, fileCount))
					
					results[result.path] = result.result
					completed <- result.path
					
					// Flush immediately
					if flusher != nil {
						flusher.Flush()
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for completion or timeout
	for i := 0; i < len(req.Folders); i++ {
		select {
		case <-completed:
			// Continue waiting for more completions
		case <-ctx.Done():
			h.writeSSEEvent(w, "timeout", `{"status":"timeout"}`)
			return
		}
	}

	// Send final completion event
	finalResponse := MultiFolderResponse{
		Success: true,
		Folders: results,
		Message: fmt.Sprintf("Successfully processed %d folders", len(req.Folders)),
	}
	finalJSON, _ := json.Marshal(finalResponse)
	h.writeSSEEvent(w, "complete", string(finalJSON))
	
	// Send keepalive events periodically
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	
	for {
		select {
		case <-keepalive.C:
			h.writeSSEEvent(w, "keepalive", `{"status":"alive"}`)
			if flusher != nil {
				flusher.Flush()
			}
		case <-ctx.Done():
			h.writeSSEEvent(w, "closed", `{"status":"connection_closed"}`)
			return
		}
	}
}

// Helper functions for SSE
func (h *FileHandler) writeSSEEvent(w http.ResponseWriter, event string, data string) {
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func (h *FileHandler) writeSSEError(w http.ResponseWriter, message string, code int, err error) {
	errorData := map[string]interface{}{
		"error":   message,
		"code":    code,
		"details": err.Error(),
	}
	errorJSON, _ := json.Marshal(errorData)
	fmt.Fprintf(w, "event: error\n")
	fmt.Fprintf(w, "data: %s\n\n", string(errorJSON))
}

// Helper function to process a single folder with all its options
func (h *FileHandler) processFolder(ctx context.Context, folderReq FolderRequest, limit int) (FolderResult, error) {
	// Normalize path
	path := strings.TrimPrefix(folderReq.Path, "/")
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// Get all objects for this path
	objects, err := h.minioClient.ListFiles(ctx, path, limit)
	if err != nil {
		return FolderResult{}, err
	}

	result := FolderResult{
		Path:       path,
		Directories: []DirectoryInfo{},
		Files:       []FileInfo{},
		TotalCount:  0,
		FileCount:   0,
		DirCount:    0,
		Size:        0,
		LastModified: "",
		Subfolders:  make(map[string]*FolderResult),
	}

	// Track directories for recursive processing
	dirMap := make(map[string]DirectoryInfo)
	fileMap := make(map[string]FileInfo)

	for _, obj := range objects {
		result.LastModified = obj.LastModified.Format(time.RFC3339)
		
		// Determine if this is a directory or file
		isDirectory := strings.HasSuffix(obj.Key, "/") && obj.Size == 0
		relativePath := strings.TrimPrefix(strings.TrimPrefix(obj.Key, path), "/")

		if isDirectory {
			// Handle directory
			dirName := strings.TrimSuffix(relativePath, "/")
			if dirName != "" && folderReq.IncludeDirs {
				// Skip current directory from being added to its own listing
				if relativePath == "" {
					continue // Skip self (current directory)
				}
				
				dirInfo := DirectoryInfo{
					Name:         dirName,
					Path:         obj.Key,
					LastModified: obj.LastModified.Format(time.RFC3339),
				}
				
				// Count items in this directory if metadata is requested
				if folderReq.IncludeMetadata {
					subFiles, err := h.minioClient.ListFiles(ctx, obj.Key, 0)
					if err == nil {
						fileCount, dirCount, totalSize := 0, 0, int64(0)
						for _, subObj := range subFiles {
							relativeSubPath := strings.TrimPrefix(subObj.Key, obj.Key)
							relativeSubPath = strings.TrimPrefix(relativeSubPath, "/")
							
							if relativeSubPath == "" {
								continue // Skip self
							}
							
							if strings.HasSuffix(subObj.Key, "/") && subObj.Size == 0 {
								dirCount++
							} else {
								fileCount++
								totalSize += subObj.Size
							}
						}
						dirInfo.FileCount = fileCount
						dirInfo.Size = totalSize
					}
				}
				
				dirMap[dirName] = dirInfo
				result.DirCount++
			}
		} else {
			// Handle file
			if folderReq.IncludeFiles {
				fileInfo := FileInfo{
					Name:         filepath.Base(obj.Key),
					Path:         obj.Key,
					Size:         obj.Size,
					LastModified: obj.LastModified.Format(time.RFC3339),
					ContentType:  obj.ContentType,
					ETag:         obj.ETag,
				}
				fileMap[filepath.Base(obj.Key)] = fileInfo
				result.FileCount++
				result.Size += obj.Size
			}
		}
	}

	// Convert maps to slices
	for _, dir := range dirMap {
		result.Directories = append(result.Directories, dir)
	}
	for _, file := range fileMap {
		result.Files = append(result.Files, file)
	}

	result.TotalCount = result.FileCount + result.DirCount

	// Process subdirectories recursively if requested
	if folderReq.Recursive && folderReq.MaxDepth > 0 {
		for dirName, dirInfo := range dirMap {
			subFolderReq := FolderRequest{
				Path:         dirInfo.Path,
				IncludeFiles: folderReq.IncludeFiles,
				IncludeDirs:  folderReq.IncludeDirs,
				Recursive:    false, // Only go one level deep per recursion call
				MaxDepth:     folderReq.MaxDepth - 1,
			}
			
			subResult, err := h.processFolder(ctx, subFolderReq, limit)
			if err == nil {
				result.Subfolders[dirName] = &subResult
			}
		}
		
		// Populate file_count and dir_count for directories from subfolder results
		for i, dir := range result.Directories {
			if subResult, exists := result.Subfolders[dir.Name]; exists {
				result.Directories[i].FileCount = subResult.FileCount
				result.Directories[i].Size = subResult.Size
			}
		}
	}

	return result, nil
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		h.writeError(w, "Failed to parse multipart form", http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, "Failed to get file from form", http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectName := header.Filename
	if customName := r.FormValue("object_name"); customName != "" {
		objectName = customName
	}

	objectName = filepath.Clean(objectName)
	if strings.HasPrefix(objectName, "/") || strings.Contains(objectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	// Check bucket status first
	bucketOk, bucketMsg := h.checkBucketStatus()
	if !bucketOk {
		h.writeError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	uploadInfo, err := h.minioClient.UploadFile(ctx, objectName, file, header.Size, contentType)
	if err != nil {
		h.writeError(w, "Failed to upload file", http.StatusInternalServerError, err)
		return
	}

	response := UploadResponse{
		Success:    true,
		Message:    "File uploaded successfully",
		ObjectName: objectName,
		Size:       uploadInfo.Size,
		ETag:       uploadInfo.ETag,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	objectName := vars["filename"]

	if objectName == "" {
		h.writeError(w, "Filename is required", http.StatusBadRequest, nil)
		return
	}

	objectName = filepath.Clean(objectName)
	if strings.HasPrefix(objectName, "/") || strings.Contains(objectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		h.writeError(w, "MinIO storage is not available", http.StatusServiceUnavailable, fmt.Errorf("MinIO client not initialized"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	exists, err := h.minioClient.FileExists(ctx, objectName)
	if err != nil {
		h.writeError(w, "Failed to check file existence", http.StatusInternalServerError, err)
		return
	}

	if !exists {
		h.writeError(w, "File not found", http.StatusNotFound, nil)
		return
	}

	fileInfo, err := h.minioClient.GetFileInfo(ctx, objectName)
	if err != nil {
		h.writeError(w, "Failed to get file info", http.StatusInternalServerError, err)
		return
	}

	reader, err := h.minioClient.DownloadFile(ctx, objectName)
	if err != nil {
		h.writeError(w, "Failed to download file", http.StatusInternalServerError, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", fileInfo.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(objectName)))
	w.Header().Set("ETag", fileInfo.ETag)
	w.Header().Set("Last-Modified", fileInfo.LastModified.Format(http.TimeFormat))

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Printf("Failed to copy file to response: %v", err)
	}
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check bucket status first
	log.Printf("ListFiles handler: checking bucket status")
	bucketOk, bucketMsg := h.checkBucketStatus()
	log.Printf("ListFiles handler: bucketOk=%v, bucketMsg=%s", bucketOk, bucketMsg)
	if !bucketOk {
		h.writeError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	prefix := r.URL.Query().Get("prefix")
	limitStr := r.URL.Query().Get("limit")

	// Set default limit to 1000 for better performance
	limit := 1000
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 10000 {
			limit = parsedLimit
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	files, err := h.minioClient.ListFiles(ctx, prefix, limit)
	if err != nil {
		h.writeError(w, "Failed to list files", http.StatusInternalServerError, err)
		return
	}

	fileResponses := make([]storage.FileInfoResponse, len(files))
	for i, file := range files {
		fileResponses[i] = storage.FileInfoResponse{
			Key:          file.Key,
			Size:         file.Size,
			LastModified: file.LastModified,
			ETag:         file.ETag,
			ContentType:  file.ContentType,
		}
	}

	response := FileListResponse{
		Success: true,
		Message: "Files listed successfully",
		Files:   fileResponses,
		Count:   len(files),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	objectName := vars["filename"]

	if objectName == "" {
		h.writeError(w, "Filename is required", http.StatusBadRequest, nil)
		return
	}

	objectName = filepath.Clean(objectName)
	if strings.HasPrefix(objectName, "/") || strings.Contains(objectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	fileInfo, err := h.minioClient.GetFileInfo(ctx, objectName)
	if err != nil {
		h.writeError(w, "Failed to get file info", http.StatusInternalServerError, err)
		return
	}

	response := FileInfoResponse{
		Success: true,
		Message: "File info retrieved successfully",
		File: storage.FileInfoDetail{
			Key:          fileInfo.Key,
			Size:         fileInfo.Size,
			LastModified: fileInfo.LastModified,
			ETag:         fileInfo.ETag,
			ContentType:  fileInfo.ContentType,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	objectName := vars["filename"]

	if objectName == "" {
		h.writeError(w, "Filename is required", http.StatusBadRequest, nil)
		return
	}

	objectName = filepath.Clean(objectName)
	if strings.HasPrefix(objectName, "/") || strings.Contains(objectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err := h.minioClient.DeleteFile(ctx, objectName)
	if err != nil {
		h.writeError(w, "Failed to delete file", http.StatusInternalServerError, err)
		return
	}

	response := DeleteResponse{
		Success: true,
		Message: "File deleted successfully",
		Deleted: []string{objectName},
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) DeleteFilesByPrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check bucket status first
	bucketOk, bucketMsg := h.checkBucketStatus()
	if !bucketOk {
		h.writeError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		h.writeError(w, "Prefix parameter is required", http.StatusBadRequest, nil)
		return
	}

	prefix = filepath.Clean(prefix)
	if strings.HasPrefix(prefix, "/") || strings.Contains(prefix, "..") {
		h.writeError(w, "Invalid prefix", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// First, list all files with the prefix
	files, err := h.minioClient.ListFiles(ctx, prefix, 0)
	if err != nil {
		h.writeError(w, "Failed to list files for deletion", http.StatusInternalServerError, err)
		return
	}

	if len(files) == 0 {
		response := DeleteResponse{
			Success: true,
			Message: "No files found with the given prefix",
			Deleted: []string{},
			Count:   0,
		}
		h.writeJSON(w, http.StatusOK, response)
		return
	}

	// Extract object names
	objectNames := make([]string, len(files))
	for i, file := range files {
		objectNames[i] = file.Key
	}

	// Delete all files
	err = h.minioClient.DeleteFiles(ctx, objectNames)
	if err != nil {
		h.writeError(w, "Failed to delete files", http.StatusInternalServerError, err)
		return
	}

	response := DeleteResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully deleted %d files", len(objectNames)),
		Deleted: objectNames,
		Count:   len(objectNames),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) GetPresignedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	objectName := vars["filename"]

	if objectName == "" {
		h.writeError(w, "Filename is required", http.StatusBadRequest, nil)
		return
	}

	objectName = filepath.Clean(objectName)
	if strings.HasPrefix(objectName, "/") || strings.Contains(objectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		h.writeError(w, "MinIO storage is not available", http.StatusServiceUnavailable, fmt.Errorf("MinIO client not initialized"))
		return
	}

	expiryStr := r.URL.Query().Get("expiry")
	expiry := 24 * time.Hour // default 24 hours
	if expiryStr != "" {
		if parsedExpiry, err := time.ParseDuration(expiryStr); err == nil {
			expiry = parsedExpiry
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	presignedURL, err := h.minioClient.GetPresignedURL(ctx, objectName, expiry)
	if err != nil {
		h.writeError(w, "Failed to generate presigned URL", http.StatusInternalServerError, err)
		return
	}

	response := map[string]any{
		"success":     true,
		"message":     "Presigned URL generated successfully",
		"url":         presignedURL,
		"expiry":      expiry.String(),
		"object_name": objectName,
	}

	h.writeJSON(w, http.StatusOK, response)
}

type CopyFileRequest struct {
	SourceObjectName string `json:"source_object_name"`
	DestObjectName   string `json:"dest_object_name"`
}

type CopyFileResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ETag         string `json:"etag,omitempty"`
	Size         int64  `json:"size,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

func (h *FileHandler) CopyFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request CopyFileRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Failed to decode request body", http.StatusBadRequest, err)
		return
	}

	if request.SourceObjectName == "" || request.DestObjectName == "" {
		h.writeError(w, "Source and destination object names are required", http.StatusBadRequest, nil)
		return
	}

	// Validate object names
	sourceObjectName := filepath.Clean(request.SourceObjectName)
	destObjectName := filepath.Clean(request.DestObjectName)

	if strings.HasPrefix(sourceObjectName, "/") || strings.Contains(sourceObjectName, "..") ||
		strings.HasPrefix(destObjectName, "/") || strings.Contains(destObjectName, "..") {
		h.writeError(w, "Invalid object name", http.StatusBadRequest, nil)
		return
	}

	// Check bucket status first
	bucketOk, bucketMsg := h.checkBucketStatus()
	if !bucketOk {
		h.writeError(w, bucketMsg, http.StatusServiceUnavailable, fmt.Errorf("bucket not accessible"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Check if source file exists
	exists, err := h.minioClient.FileExists(ctx, sourceObjectName)
	if err != nil {
		h.writeError(w, "Failed to check source file existence", http.StatusInternalServerError, err)
		return
	}

	if !exists {
		h.writeError(w, "Source file does not exist", http.StatusNotFound, nil)
		return
	}

	// Copy the file
	copyInfo, err := h.minioClient.CopyFile(ctx, sourceObjectName, destObjectName)
	if err != nil {
		h.writeError(w, "Failed to copy file", http.StatusInternalServerError, err)
		return
	}

	response := CopyFileResponse{
		Success:      true,
		Message:      "File copied successfully",
		ETag:         copyInfo.ETag,
		Size:         copyInfo.Size,
		LastModified: copyInfo.LastModified.Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) ListBuckets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		h.writeError(w, "MinIO storage is not available", http.StatusServiceUnavailable, fmt.Errorf("MinIO client not initialized"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	buckets, err := h.minioClient.GetClient().ListBuckets(ctx)
	if err != nil {
		h.writeError(w, "Failed to list buckets", http.StatusInternalServerError, err)
		return
	}

	bucketResponses := make([]BucketInfoResponse, len(buckets))
	for i, bucket := range buckets {
		bucketResponses[i] = BucketInfoResponse{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
		}
	}

	response := BucketListResponse{
		Success: true,
		Message: "Buckets listed successfully",
		Buckets: bucketResponses,
		Count:   len(buckets),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) SetBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		BucketName string `json:"bucket_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Failed to decode request body", http.StatusBadRequest, err)
		return
	}

	if request.BucketName == "" {
		h.writeError(w, "Bucket name is required", http.StatusBadRequest, nil)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		h.writeError(w, "MinIO storage is not available", http.StatusServiceUnavailable, fmt.Errorf("MinIO client not initialized"))
		return
	}

	if err := h.minioClient.SetBucket(request.BucketName); err != nil {
		h.writeError(w, "Failed to set bucket", http.StatusBadRequest, err)
		return
	}

	response := SetBucketResponse{
		Success: true,
		Message: "Bucket set successfully",
		Bucket:  request.BucketName,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) GetCurrentBucket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		h.writeError(w, "MinIO storage is not available", http.StatusServiceUnavailable, fmt.Errorf("MinIO client not initialized"))
		return
	}

	currentBucket := h.minioClient.GetBucketName()

	response := map[string]any{
		"success":     true,
		"message":     "Current bucket retrieved successfully",
		"bucket_name": currentBucket,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) GetBucketStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if MinIO is available
	if h.minioClient == nil {
		response := map[string]any{
			"success": false,
			"message": "MinIO storage is not available",
			"bucket":  "",
			"exists":  false,
			"error":   "MinIO client not initialized",
		}
		h.writeJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	currentBucket := h.minioClient.GetBucketName()
	bucketExists, bucketError := h.minioClient.GetBucketStatus()

	response := map[string]any{
		"success": true,
		"message": "Bucket status retrieved successfully",
		"bucket":  currentBucket,
		"exists":  bucketExists,
		"error":   bucketError,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) checkBucketStatus() (bool, string) {
	log.Printf("checkBucketStatus: starting")
	if h.minioClient == nil {
		log.Printf("checkBucketStatus: minioClient is nil")
		return false, "MinIO client not initialized"
	}

	bucketExists, bucketError := h.minioClient.GetBucketStatus()
	log.Printf("checkBucketStatus: bucketExists=%v, bucketError=%s", bucketExists, bucketError)
	if !bucketExists {
		errorMsg := fmt.Sprintf("Bucket '%s' is not accessible", h.minioClient.GetBucketName())
		if bucketError != "" {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, bucketError)
		}
		log.Printf("checkBucketStatus: returning false with errorMsg=%s", errorMsg)
		return false, errorMsg
	}

	log.Printf("checkBucketStatus: returning true")
	return true, ""
}

// ExtractArchive extracts an archive file and returns information about the extraction
func (h *FileHandler) ExtractArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
		return
	}

	var request struct {
		FileName string `json:"file_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid JSON request", http.StatusBadRequest, err)
		return
	}

	if request.FileName == "" {
		h.writeError(w, "file_name is required", http.StatusBadRequest, nil)
		return
	}

	// Create a job request for archive extraction
	jobRequest := map[string]any{
		"type":        "extract",
		"file_path":   request.FileName,
		"bucket":      h.minioClient.GetBucketName(),
		"object_name": request.FileName,
		"priority":    "medium",
	}

	// Convert to JSON for job handler
	jobRequestJSON, _ := json.Marshal(jobRequest)

	// Create HTTP request to job handler
	req, _ := http.NewRequest("POST", "/api/jobs", bytes.NewBuffer(jobRequestJSON))
	req.Header.Set("Content-Type", "application/json")

	// Forward to job handler
	// Note: In a proper implementation, this would be handled by dependency injection
	// For now, we'll create the job directly through the job system

	// Create job through job system
	job := &jobs.Job{
		ID:         fmt.Sprintf("extract_%d", time.Now().UnixNano()),
		Type:       "extract",
		Bucket:     h.minioClient.GetBucketName(),
		ObjectName: request.FileName,
		Priority:   jobs.PriorityMedium,
		Status:     jobs.JobStatusPending,
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]any),
	}

	// Enqueue job for async processing
	if h.jobQueue != nil {
		err := h.jobQueue.Enqueue(job)
		if err != nil {
			h.writeError(w, "Failed to enqueue extraction job", http.StatusInternalServerError, err)
			return
		}
	} else {
		// Fallback: process synchronously if no queue available
		ctx := r.Context()
		result := h.processor.ProcessJob(ctx, job)
		response := map[string]any{
			"success": result.Success,
			"message": result.Message,
			"job": map[string]any{
				"id":          job.ID,
				"type":        job.Type,
				"status":      job.Status,
				"file_path":   job.FilePath,
				"bucket":      job.Bucket,
				"object_name": job.ObjectName,
				"created_at":  job.CreatedAt.Format(time.RFC3339),
				"progress":    job.Progress,
			},
		}

		if result.Success {
			response["extracted_files"] = result.ExtractedFiles
			response["file_count"] = len(result.ExtractedFiles)
			if result.FileInfo != nil {
				response["archive_info"] = result.FileInfo
			}
		}

		h.writeJSON(w, http.StatusOK, response)
		return
	}

	response := map[string]any{
		"success": true,
		"message": "Extraction job created successfully",
		"job": map[string]any{
			"id":          job.ID,
			"type":        job.Type,
			"status":      job.Status,
			"file_path":   job.FilePath,
			"bucket":      job.Bucket,
			"object_name": job.ObjectName,
			"created_at":  job.CreatedAt.Format(time.RFC3339),
			"progress":    job.Progress,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *FileHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *FileHandler) writeError(w http.ResponseWriter, message string, statusCode int, err error) {
	response := map[string]any{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
		log.Printf("Error: %v", err)
	}

	h.writeJSON(w, statusCode, response)
}
