package handlers

import (
	"encoding/json"
	"net/http"

	"bronze-backend/processor"
	"github.com/gorilla/mux"
)

type JobHandler struct {
	jobQueue   *processor.JobQueue
	workerPool *processor.WorkerPool
}

func NewJobHandler(jobQueue *processor.JobQueue, workerPool *processor.WorkerPool) *JobHandler {
	return &JobHandler{
		jobQueue:   jobQueue,
		workerPool: workerPool,
	}
}

type CreateJobRequest struct {
	Type       string `json:"type"`
	FilePath   string `json:"file_path"`
	Bucket     string `json:"bucket"`
	ObjectName string `json:"object_name"`
	Priority   string `json:"priority"`
}

type JobResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Job     *processor.Job `json:"job,omitempty"`
}

type JobsListResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Jobs    []*processor.Job `json:"jobs"`
	Count   int              `json:"count"`
}

type JobStatsResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Queue   processor.QueueStats      `json:"queue"`
	Workers processor.WorkerPoolStats `json:"workers"`
}

type UpdatePriorityRequest struct {
	Priority string `json:"priority"`
}

type UpdateWorkersRequest struct {
	Count int `json:"count"`
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	if req.Type == "" {
		h.writeError(w, "Job type is required", http.StatusBadRequest, nil)
		return
	}

	if req.FilePath == "" {
		h.writeError(w, "File path is required", http.StatusBadRequest, nil)
		return
	}

	if req.Bucket == "" {
		h.writeError(w, "Bucket is required", http.StatusBadRequest, nil)
		return
	}

	if req.ObjectName == "" {
		h.writeError(w, "Object name is required", http.StatusBadRequest, nil)
		return
	}

	priority := processor.ParsePriority(req.Priority)
	if priority == processor.PriorityMedium && req.Priority != "" && req.Priority != "medium" {
		h.writeError(w, "Invalid priority. Use: high, medium, low", http.StatusBadRequest, nil)
		return
	}

	job := processor.NewJob(req.Type, req.FilePath, req.Bucket, req.ObjectName, priority)

	err := h.jobQueue.Enqueue(job)
	if err != nil {
		h.writeError(w, "Failed to enqueue job", http.StatusInternalServerError, err)
		return
	}

	response := JobResponse{
		Success: true,
		Message: "Job created successfully",
		Job:     job,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := r.URL.Query().Get("status")

	var jobs []*processor.Job
	if status != "" {
		jobStatus := processor.JobStatus(status)
		jobs = h.jobQueue.ListJobsByStatus(jobStatus)
	} else {
		jobs = h.jobQueue.ListJobs()
	}

	response := JobsListResponse{
		Success: true,
		Message: "Jobs retrieved successfully",
		Jobs:    jobs,
		Count:   len(jobs),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["id"]

	if jobID == "" {
		h.writeError(w, "Job ID is required", http.StatusBadRequest, nil)
		return
	}

	job, exists := h.jobQueue.GetJob(jobID)
	if !exists {
		h.writeError(w, "Job not found", http.StatusNotFound, nil)
		return
	}

	response := JobResponse{
		Success: true,
		Message: "Job retrieved successfully",
		Job:     job,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["id"]

	if jobID == "" {
		h.writeError(w, "Job ID is required", http.StatusBadRequest, nil)
		return
	}

	success := h.jobQueue.CancelJob(jobID)
	if !success {
		h.writeError(w, "Job not found or cannot be cancelled", http.StatusNotFound, nil)
		return
	}

	response := map[string]any{
		"success": true,
		"message": "Job cancelled successfully",
		"job_id":  jobID,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) UpdateJobPriority(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	jobID := vars["id"]

	if jobID == "" {
		h.writeError(w, "Job ID is required", http.StatusBadRequest, nil)
		return
	}

	var req UpdatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	priority := processor.ParsePriority(req.Priority)
	if priority == processor.PriorityMedium && req.Priority != "" && req.Priority != "medium" {
		h.writeError(w, "Invalid priority. Use: high, medium, low", http.StatusBadRequest, nil)
		return
	}

	job, exists := h.jobQueue.GetJob(jobID)
	if !exists {
		h.writeError(w, "Job not found", http.StatusNotFound, nil)
		return
	}

	if job.Status != processor.JobStatusPending {
		h.writeError(w, "Cannot update priority of job that is not pending", http.StatusBadRequest, nil)
		return
	}

	job.Priority = priority

	response := map[string]any{
		"success":  true,
		"message":  "Job priority updated successfully",
		"job_id":   jobID,
		"priority": priority.String(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueStats := h.jobQueue.GetStats()
	workerStats := h.workerPool.GetStats()

	response := JobStatsResponse{
		Success: true,
		Message: "Stats retrieved successfully",
		Queue:   queueStats,
		Workers: workerStats,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) UpdateWorkerCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateWorkersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	if req.Count <= 0 || req.Count > 100 {
		h.writeError(w, "Worker count must be between 1 and 100", http.StatusBadRequest, nil)
		return
	}

	h.workerPool.UpdateWorkerCount(req.Count)

	response := map[string]any{
		"success": true,
		"message": "Worker count updated successfully",
		"count":   req.Count,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) GetActiveJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activeJobs := h.workerPool.GetActiveJobs()

	response := JobsListResponse{
		Success: true,
		Message: "Active jobs retrieved successfully",
		Jobs:    activeJobs,
		Count:   len(activeJobs),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *JobHandler) writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *JobHandler) writeError(w http.ResponseWriter, message string, statusCode int, err error) {
	response := map[string]any{
		"success": false,
		"message": message,
	}
	if err != nil {
		response["error"] = err.Error()
	}

	h.writeJSON(w, statusCode, response)
}
