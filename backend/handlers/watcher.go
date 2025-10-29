package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bronze-backend/watcher"
)

// WatcherHandler handles file watcher related requests
type WatcherHandler struct {
	watcher *watcher.FileWatcher
}

// NewWatcherHandler creates a new watcher handler
func NewWatcherHandler(fileWatcher *watcher.FileWatcher) *WatcherHandler {
	return &WatcherHandler{
		watcher: fileWatcher,
	}
}

// GetUnprocessedEvents returns unprocessed file events
func (h *WatcherHandler) GetUnprocessedEvents(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events, err := h.watcher.GetUnprocessedEvents(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"events": events,
		"count":  len(events),
	})
}

// GetEventHistory returns file event history
func (h *WatcherHandler) GetEventHistory(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events, err := h.watcher.GetEventHistory(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"events": events,
		"count":  len(events),
	})
}

// MarkEventProcessed marks an event as processed
func (h *WatcherHandler) MarkEventProcessed(w http.ResponseWriter, r *http.Request) {
	var request struct {
		EventID string `json:"event_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.EventID == "" {
		http.Error(w, "event_id is required", http.StatusBadRequest)
		return
	}

	err := h.watcher.MarkEventProcessed(request.EventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Event marked as processed",
	})
}
