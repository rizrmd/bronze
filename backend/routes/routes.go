package routes

import (
	"encoding/json"
	"net/http"

	"bronze-backend/handlers"
	"github.com/gorilla/mux"
)

type Router struct {
	router *mux.Router
}

func NewRouter(
	fileHandler *handlers.FileHandler,
	jobHandler *handlers.JobHandler,
	watcherHandler *handlers.WatcherHandler,
) *Router {
	router := mux.NewRouter()

	r := &Router{
		router: router,
	}

	r.setupRoutes(fileHandler, jobHandler, watcherHandler)

	return r
}

func (r *Router) setupRoutes(
	fileHandler *handlers.FileHandler,
	jobHandler *handlers.JobHandler,
	watcherHandler *handlers.WatcherHandler,
) {
	// Health check
	r.router.HandleFunc("/health", r.healthCheck).Methods("GET")
	r.router.HandleFunc("/", r.healthCheck).Methods("GET")

	// File routes
	fileRouter := r.router.PathPrefix("/files").Subrouter()
	fileRouter.HandleFunc("", fileHandler.UploadFile).Methods("POST")
	fileRouter.HandleFunc("", fileHandler.ListFiles).Methods("GET")
	fileRouter.HandleFunc("/{filename}", fileHandler.DownloadFile).Methods("GET")
	fileRouter.HandleFunc("/{filename}", fileHandler.GetFileInfo).Methods("GET")
	fileRouter.HandleFunc("/{filename}", fileHandler.DeleteFile).Methods("DELETE")
	fileRouter.HandleFunc("/{filename}/presigned", fileHandler.GetPresignedURL).Methods("GET")

	// Job routes
	jobRouter := r.router.PathPrefix("/jobs").Subrouter()
	jobRouter.HandleFunc("", jobHandler.CreateJob).Methods("POST")
	jobRouter.HandleFunc("", jobHandler.GetJobs).Methods("GET")
	jobRouter.HandleFunc("/stats", jobHandler.GetStats).Methods("GET")
	jobRouter.HandleFunc("/workers", jobHandler.UpdateWorkerCount).Methods("PUT")
	jobRouter.HandleFunc("/workers/active", jobHandler.GetActiveJobs).Methods("GET")
	jobRouter.HandleFunc("/{id}", jobHandler.GetJob).Methods("GET")
	jobRouter.HandleFunc("/{id}", jobHandler.CancelJob).Methods("DELETE")
	jobRouter.HandleFunc("/{id}/priority", jobHandler.UpdateJobPriority).Methods("PUT")

	// Watcher routes
	watcherRouter := r.router.PathPrefix("/watcher").Subrouter()
	watcherRouter.HandleFunc("/events/unprocessed", watcherHandler.GetUnprocessedEvents).Methods("GET")
	watcherRouter.HandleFunc("/events/history", watcherHandler.GetEventHistory).Methods("GET")
	watcherRouter.HandleFunc("/events/mark-processed", watcherHandler.MarkEventProcessed).Methods("POST")

	// API documentation routes
	r.router.HandleFunc("/api", r.apiInfo).Methods("GET")
	r.router.HandleFunc("/openapi.json", r.openAPISpec).Methods("GET")
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
		"openapi":     "/openapi.json",
		"endpoints": map[string]any{
			"files": map[string]any{
				"upload": map[string]any{
					"method":      "POST",
					"path":        "/files",
					"description": "Upload a file to MinIO",
				},
				"list": map[string]any{
					"method":       "GET",
					"path":         "/files",
					"description":  "List all files in MinIO bucket",
					"query_params": []string{"prefix"},
				},
				"download": map[string]any{
					"method":      "GET",
					"path":        "/files/{filename}",
					"description": "Download a specific file",
				},
				"info": map[string]any{
					"method":      "GET",
					"path":        "/files/{filename}",
					"description": "Get file information",
				},
				"delete": map[string]any{
					"method":      "DELETE",
					"path":        "/files/{filename}",
					"description": "Delete a specific file",
				},
				"presigned": map[string]any{
					"method":       "GET",
					"path":         "/files/{filename}/presigned",
					"description":  "Generate presigned URL for file access",
					"query_params": []string{"expiry"},
				},
			},
			"jobs": map[string]any{
				"create": map[string]any{
					"method":      "POST",
					"path":        "/jobs",
					"description": "Create a new processing job",
				},
				"list": map[string]any{
					"method":       "GET",
					"path":         "/jobs",
					"description":  "List all jobs",
					"query_params": []string{"status"},
				},
				"get": map[string]any{
					"method":      "GET",
					"path":        "/jobs/{id}",
					"description": "Get specific job details",
				},
				"cancel": map[string]any{
					"method":      "DELETE",
					"path":        "/jobs/{id}",
					"description": "Cancel a specific job",
				},
				"update_priority": map[string]any{
					"method":      "PUT",
					"path":        "/jobs/{id}/priority",
					"description": "Update job priority",
				},
				"stats": map[string]any{
					"method":      "GET",
					"path":        "/jobs/stats",
					"description": "Get job queue and worker statistics",
				},
				"update_workers": map[string]any{
					"method":      "PUT",
					"path":        "/jobs/workers",
					"description": "Update worker pool size",
				},
				"active_jobs": map[string]any{
					"method":      "GET",
					"path":        "/jobs/workers/active",
					"description": "Get currently active jobs",
				},
			},
			"watcher": map[string]any{
				"unprocessed_events": map[string]any{
					"method":       "GET",
					"path":         "/watcher/events/unprocessed",
					"description":  "Get unprocessed file change events",
					"query_params": []string{"limit"},
				},
				"event_history": map[string]any{
					"method":       "GET",
					"path":         "/watcher/events/history",
					"description":  "Get file change event history",
					"query_params": []string{"limit"},
				},
				"mark_processed": map[string]any{
					"method":      "POST",
					"path":        "/watcher/events/mark-processed",
					"description": "Mark a file event as processed",
				},
			},
		},
		"features": []string{
			"MinIO object storage integration",
			"File upload/download/management",
			"Priority-based job queue",
			"Configurable worker pool",
			"Archive decompression (ZIP, TAR, TAR.GZ)",
			"File processing pipeline",
			"Real-time job tracking",
			"File watching and change tracking",
			"Automatic job creation for new files",
			"Event history and processing status",
			"RESTful API",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiInfo)
}

func (r *Router) openAPISpec(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	// Read and serve the OpenAPI JSON file
	http.ServeFile(w, req, "openapi.json")
}
