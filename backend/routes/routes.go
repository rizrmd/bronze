package routes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"bronze-backend/data_browser"
	"bronze-backend/files"
	"bronze-backend/jobs"
	"bronze-backend/monitoring"
	"github.com/gorilla/mux"
)

type Router struct {
	router *mux.Router
}

func NewRouter(
	fileHandler *files.FileHandler,
	jobHandler *jobs.JobHandler,
	watcherHandler *monitoring.WatcherHandler,
	dataBrowserHandler *data_browser.DataBrowserHandler,
	exportHandler *data_browser.ExportHandler,
) *Router {
	router := mux.NewRouter()

	r := &Router{
		router: router,
	}

	r.setupRoutes(fileHandler, jobHandler, watcherHandler, dataBrowserHandler, exportHandler)

	return r
}

func (r *Router) setupRoutes(
	fileHandler *files.FileHandler,
	jobHandler *jobs.JobHandler,
	watcherHandler *monitoring.WatcherHandler,
	dataBrowserHandler *data_browser.DataBrowserHandler,
	exportHandler *data_browser.ExportHandler,
) {
	// Add CORS middleware
	r.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check
	r.router.HandleFunc("/api/health", r.healthCheck).Methods("GET")
	r.router.HandleFunc("/api", r.healthCheck).Methods("GET")

	// File routes - comprehensive endpoints
	fileRouter := r.router.PathPrefix("/api/files").Subrouter()
	
	// New multi-folder endpoint
	fileRouter.HandleFunc("/browse", fileHandler.MultiFolderBrowse).Methods("POST")
	
	// Specific operation endpoints
	fileRouter.HandleFunc("/upload", fileHandler.UploadFile).Methods("POST")
	fileRouter.HandleFunc("/download/{filename:.+}", fileHandler.DownloadFile).Methods("GET")
	fileRouter.HandleFunc("/info/{filename:.+}", fileHandler.GetFileInfo).Methods("GET")
	fileRouter.HandleFunc("/presigned/{filename:.+}", fileHandler.GetPresignedURL).Methods("GET")
	fileRouter.HandleFunc("/delete", fileHandler.DeleteFile).Methods("POST")
	fileRouter.HandleFunc("/copy", fileHandler.CopyFile).Methods("POST")
	fileRouter.HandleFunc("/extract", fileHandler.ExtractArchive).Methods("POST")
	
	// Legacy root-level endpoints for compatibility
	fileRouter.HandleFunc("", fileHandler.ListFiles).Methods("GET")
	fileRouter.HandleFunc("", fileHandler.BatchListFiles).Methods("POST")
	fileRouter.HandleFunc("", fileHandler.DeleteFilesByPrefix).Methods("DELETE")
	fileRouter.HandleFunc("/{filename:.+}", fileHandler.DownloadFile).Methods("GET")
	fileRouter.HandleFunc("/{filename:.+}/info", fileHandler.GetFileInfo).Methods("GET")
	fileRouter.HandleFunc("/{filename:.+}/presigned", fileHandler.GetPresignedURL).Methods("GET")
	fileRouter.HandleFunc("/{filename:.+}", fileHandler.DeleteFile).Methods("DELETE")

	// Bucket management routes
	bucketRouter := r.router.PathPrefix("/api/buckets").Subrouter()
	bucketRouter.HandleFunc("", fileHandler.ListBuckets).Methods("GET")
	bucketRouter.HandleFunc("/current", fileHandler.GetCurrentBucket).Methods("GET")
	bucketRouter.HandleFunc("/status", fileHandler.GetBucketStatus).Methods("GET")
	bucketRouter.HandleFunc("/set", fileHandler.SetBucket).Methods("POST")

	// Job routes
	jobRouter := r.router.PathPrefix("/api/jobs").Subrouter()
	jobRouter.HandleFunc("", jobHandler.CreateJob).Methods("POST")
	jobRouter.HandleFunc("", jobHandler.GetJobs).Methods("GET")
	jobRouter.HandleFunc("/stats", jobHandler.GetStats).Methods("GET")
	jobRouter.HandleFunc("/workers", jobHandler.UpdateWorkerCount).Methods("PUT")
	jobRouter.HandleFunc("/workers/calculate-max", jobHandler.CalculateMaxWorkers).Methods("GET")
	jobRouter.HandleFunc("/workers/active", jobHandler.GetActiveJobs).Methods("GET")
	jobRouter.HandleFunc("/{id}", jobHandler.GetJob).Methods("GET")
	jobRouter.HandleFunc("/{id}", jobHandler.CancelJob).Methods("DELETE")
	jobRouter.HandleFunc("/{id}/priority", jobHandler.UpdateJobPriority).Methods("PUT")

	// Watcher routes
	watcherRouter := r.router.PathPrefix("/api/watcher").Subrouter()
	watcherRouter.HandleFunc("/events/unprocessed", watcherHandler.GetUnprocessedEvents).Methods("GET")
	watcherRouter.HandleFunc("/events/history", watcherHandler.GetEventHistory).Methods("GET")
	watcherRouter.HandleFunc("/events/mark-processed", watcherHandler.MarkEventProcessed).Methods("POST")

	// Data browser routes
	dataRouter := r.router.PathPrefix("/api/data").Subrouter()
	dataRouter.HandleFunc("/browse", dataBrowserHandler.BrowseData).Methods("POST")
	dataRouter.HandleFunc("/files", dataBrowserHandler.ListDataFiles).Methods("GET")

	// Export routes
	dataRouter.HandleFunc("/export-single", exportHandler.ExportSingleFile).Methods("POST")
	dataRouter.HandleFunc("/export-multiple", exportHandler.ExportMultipleFiles).Methods("POST")
	dataRouter.HandleFunc("/export-job", exportHandler.CreateExportJob).Methods("POST")

	// Configuration routes
	r.router.HandleFunc("/api/config", r.getConfig).Methods("GET")
	r.router.HandleFunc("/api/config", r.updateConfig).Methods("PUT")

	// API documentation routes
	r.router.HandleFunc("/api", r.apiInfo).Methods("GET")
	r.router.HandleFunc("/api/openapi.json", r.openAPISpec).Methods("GET")
}

func (r *Router) GetRouter() *mux.Router {
	return r.router
}

func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":  "ok",
		"service": "bronze-backend",
	}
	json.NewEncoder(w).Encode(response)
}

func (r *Router) apiInfo(w http.ResponseWriter, req *http.Request) {
	apiInfo := map[string]any{
		"name":        "Bronze Backend API",
		"version":     "1.0.0",
		"description": "A Go backend with MinIO integration, file processing, and job management",
		"openapi":     "/api/openapi.json",
		"endpoints": map[string]any{
			"files": map[string]any{
				"browse": map[string]any{
					"method": "POST",
					"path":   "/api/files/browse",
					"description": "Browse multiple folders with rich metadata and recursive options",
					"body": map[string]any{
						"folders": "[]FolderRequest - Array of folder requests with options",
						"limit":   "int (optional) - Maximum items per folder",
					},
				},
				"upload": map[string]any{
					"method":      "POST",
					"path":        "/api/files/upload",
					"description": "Upload a file to MinIO",
					"body":        "multipart/form-data with file field",
				},
				"download": map[string]any{
					"method":      "GET",
					"path":        "/api/files/download/{filename}",
					"description": "Download a specific file",
				},
				"info": map[string]any{
					"method":      "GET",
					"path":        "/api/files/info/{filename}",
					"description": "Get file information",
				},
				"delete": map[string]any{
					"method":      "POST",
					"path":        "/api/files/delete",
					"description": "Delete a specific file",
					"body": map[string]any{
						"filename": "string - File to delete",
					},
				},
				"presigned": map[string]any{
					"method":       "GET",
					"path":         "/api/files/presigned/{filename}",
					"description":  "Generate presigned URL for file access",
					"query_params": []string{"expiry"},
				},
				"copy": map[string]any{
					"method":      "POST",
					"path":        "/api/files/copy",
					"description": "Copy a file to a new location",
					"body": map[string]any{
						"source":      "string - Source file path",
						"destination": "string - Destination file path",
					},
				},
				"extract": map[string]any{
					"method":      "POST",
					"path":        "/api/files/extract",
					"description": "Extract archive files (ZIP, TAR, TAR.GZ)",
					"body": map[string]any{
						"filename":           "string - Archive file to extract",
						"destination_folder":  "string (optional) - Extract to specific folder",
						"delete_after":       "bool (optional) - Delete archive after extraction",
					},
				},
			},
			"buckets": map[string]any{
				"list": map[string]any{
					"method":      "GET",
					"path":        "/api/buckets",
					"description": "List all available buckets",
				},
				"current": map[string]any{
					"method":      "GET",
					"path":        "/api/buckets/current",
					"description": "Get currently active bucket",
				},
				"status": map[string]any{
					"method":      "GET",
					"path":        "/api/buckets/status",
					"description": "Get bucket status and availability",
				},
				"set": map[string]any{
					"method":      "POST",
					"path":        "/api/buckets/set",
					"description": "Set active bucket",
					"body": map[string]any{
						"bucket_name": "string",
					},
				},
			},
			"jobs": map[string]any{
				"create": map[string]any{
					"method":      "POST",
					"path":        "/api/jobs",
					"description": "Create a new processing job",
				},
				"list": map[string]any{
					"method":       "GET",
					"path":         "/api/jobs",
					"description":  "List all jobs",
					"query_params": []string{"status"},
				},
				"get": map[string]any{
					"method":      "GET",
					"path":        "/api/jobs/{id}",
					"description": "Get specific job details",
				},
				"cancel": map[string]any{
					"method":      "DELETE",
					"path":        "/api/jobs/{id}",
					"description": "Cancel a specific job",
				},
				"update_priority": map[string]any{
					"method":      "PUT",
					"path":        "/api/jobs/{id}/priority",
					"description": "Update job priority",
				},
				"stats": map[string]any{
					"method":      "GET",
					"path":        "/api/jobs/stats",
					"description": "Get job queue and worker statistics",
				},
				"update_workers": map[string]any{
					"method":      "PUT",
					"path":        "/api/jobs/workers",
					"description": "Update worker pool size",
				},
				"calculate_max_workers": map[string]any{
					"method":      "GET",
					"path":        "/api/jobs/workers/calculate-max",
					"description": "Calculate optimal number of workers based on CPU cores",
				},
				"active_jobs": map[string]any{
					"method":      "GET",
					"path":        "/api/jobs/workers/active",
					"description": "Get currently active jobs",
				},
			},
			"data": map[string]any{
				"browse": map[string]any{
					"method":      "POST",
					"path":        "/api/data/browse",
					"description": "Browse data from Excel (XLSX, XLS, XLSM), CSV, or MDB files in S3",
					"body": map[string]any{
						"file_name":           "string (required)",
						"sheet_name":          "string (optional, for Excel files)",
						"max_rows":            "int (optional, default 100, max 10000)",
						"offset":              "int (optional, default 0)",
						"has_headers":         "bool (optional, default false)",
						"treat_as_csv":        "bool (optional, default false)",
						"auto_detect_headers": "bool (optional, default false)",
						"stream_mode":         "bool (optional, default false)",
						"chunk_size":          "int (optional, default 1000, streaming only)",
					},
				},
				"files": map[string]any{
					"method":      "GET",
					"path":        "/api/data/files",
					"description": "List all supported data files (Excel XLSX/XLS/XLSM, CSV, MDB)",
				},
			},
			"watcher": map[string]any{
				"unprocessed_events": map[string]any{
					"method":       "GET",
					"path":         "/api/watcher/events/unprocessed",
					"description":  "Get unprocessed file change events",
					"query_params": []string{"limit"},
				},
				"event_history": map[string]any{
					"method":       "GET",
					"path":         "/api/watcher/events/history",
					"description":  "Get file change event history",
					"query_params": []string{"limit"},
				},
				"mark_processed": map[string]any{
					"method":      "POST",
					"path":        "/api/watcher/events/mark-processed",
					"description": "Mark a file event as processed",
				},
			},
		},
		"features": []string{
			"MinIO object storage integration",
			"File upload/download/management",
			"Bucket management and selection",
			"Priority-based job queue",
			"Configurable worker pool",
			"Archive decompression (ZIP, TAR, TAR.GZ)",
			"File processing pipeline",
			"Real-time job tracking",
			"File watching and change tracking",
			"Automatic job creation for new files",
			"Event history and processing status",
			"Unified data browser for Excel (XLSX/XLS/XLSM), CSV, MDB files",
			"Streaming support for large CSV files",
			"Auto-detection of delimiters and headers",
			"Universal CSV processing (any file extension)",
			"RESTful API",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiInfo)
}

func (r *Router) getConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Read current .env file
	envData := make(map[string]string)

	// Default values
	envData["SERVER_HOST"] = "localhost"
	envData["SERVER_PORT"] = "8060"
	envData["MINIO_ENDPOINT"] = "localhost:9000"
	envData["MINIO_ACCESS_KEY"] = "minioadmin"
	envData["MINIO_SECRET_KEY"] = "minioadmin"
	envData["MINIO_USE_SSL"] = "false"
	envData["MINIO_BUCKET"] = "files"
	envData["MINIO_REGION"] = "us-east-1"
	envData["MAX_WORKERS"] = "3"
	envData["QUEUE_SIZE"] = "100"
	envData["WATCH_INTERVAL"] = "5s"
	envData["TEMP_DIR"] = "/tmp/bronze"
	envData["DECOMPRESSION_ENABLED"] = "true"
	envData["MAX_EXTRACT_SIZE"] = "1GB"
	envData["MAX_FILES_PER_ARCHIVE"] = "1000"
	envData["NESTED_ARCHIVE_DEPTH"] = "3"
	envData["PASSWORD_PROTECTED"] = "true"
	envData["EXTRACT_TO_SUBFOLDER"] = "true"

	// Try to read actual .env file
	if envFile, err := os.Open(".env"); err == nil {
		defer envFile.Close()
		scanner := bufio.NewScanner(envFile)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					envData[key] = value
				}
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    envData,
	})
}

func (r *Router) updateConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var updates map[string]string
	if err := json.NewDecoder(req.Body).Decode(&updates); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid JSON",
		})
		return
	}

	// Read existing .env file
	envFile, err := os.Open(".env")
	var envLines []string
	if err == nil {
		defer envFile.Close()
		scanner := bufio.NewScanner(envFile)
		for scanner.Scan() {
			envLines = append(envLines, scanner.Text())
		}
	}

	// Update values in memory
	for key, value := range updates {
		// Find and replace existing line or add new one
		found := false
		for i, line := range envLines {
			if strings.HasPrefix(strings.TrimSpace(line), key+"=") {
				envLines[i] = fmt.Sprintf("%s=%s", key, value)
				found = true
				break
			}
		}
		if !found {
			envLines = append(envLines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Write back to .env file
	if err := os.WriteFile(".env", []byte(strings.Join(envLines, "\n")), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to write .env file: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
		"data":    updates,
	})
}

func (r *Router) openAPISpec(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Read and serve OpenAPI JSON file
	http.ServeFile(w, req, "openapi.json")
}
