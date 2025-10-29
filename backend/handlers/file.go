package handlers

import (
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

	"bronze-backend/minio"
	"bronze-backend/processor"
	"github.com/gorilla/mux"
)

type FileHandler struct {
	minioClient *minio.MinIOClient
	processor   *processor.FileProcessor
}

func NewFileHandler(minioClient *minio.MinIOClient, fileProcessor *processor.FileProcessor) *FileHandler {
	return &FileHandler{
		minioClient: minioClient,
		processor:   fileProcessor,
	}
}

type UploadResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ObjectName string `json:"object_name"`
	Size       int64  `json:"size"`
	ETag       string `json:"etag,omitempty"`
}

type FileListResponse struct {
	Success bool                     `json:"success"`
	Message string                   `json:"message"`
	Files   []minio.FileInfoResponse `json:"files"`
	Count   int                      `json:"count"`
}

type FileInfoResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	File    minio.FileInfoDetail `json:"file"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Deleted string `json:"deleted,omitempty"`
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

	prefix := r.URL.Query().Get("prefix")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	files, err := h.minioClient.ListFiles(ctx, prefix)
	if err != nil {
		h.writeError(w, "Failed to list files", http.StatusInternalServerError, err)
		return
	}

	fileResponses := make([]minio.FileInfoResponse, len(files))
	for i, file := range files {
		fileResponses[i] = minio.FileInfoResponse{
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
		File: minio.FileInfoDetail{
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
		Deleted: objectName,
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
