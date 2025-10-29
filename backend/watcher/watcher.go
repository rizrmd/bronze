package watcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// EventType represents the type of file event
type EventType string

const (
	EventCreated  EventType = "s3:ObjectCreated:*"
	EventRemoved  EventType = "s3:ObjectRemoved:*"
	EventMetadata EventType = "s3:ObjectMetadata:*"
)

// FileEvent represents a file change event
type FileEvent struct {
	ID          string            `json:"id"`
	Bucket      string            `json:"bucket"`
	Key         string            `json:"key"`
	Size        int64             `json:"size"`
	ETag        string            `json:"etag"`
	EventType   EventType         `json:"event_type"`
	EventTime   time.Time         `json:"event_time"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Processed   bool              `json:"processed"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
}

// EventStorage interface for storing file events
type EventStorage interface {
	Store(event *FileEvent) error
	GetUnprocessed(limit int) ([]*FileEvent, error)
	MarkProcessed(eventID string) error
	GetHistory(limit int) ([]*FileEvent, error)
}

// MemoryEventStorage implements in-memory event storage
type MemoryEventStorage struct {
	events map[string]*FileEvent
	mu     sync.RWMutex
}

func NewMemoryEventStorage() *MemoryEventStorage {
	return &MemoryEventStorage{
		events: make(map[string]*FileEvent),
	}
}

func (m *MemoryEventStorage) Store(event *FileEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.ID] = event
	return nil
}

func (m *MemoryEventStorage) GetUnprocessed(limit int) ([]*FileEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var unprocessed []*FileEvent
	count := 0
	for _, event := range m.events {
		if !event.Processed && count < limit {
			unprocessed = append(unprocessed, event)
			count++
		}
	}
	return unprocessed, nil
}

func (m *MemoryEventStorage) MarkProcessed(eventID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event, exists := m.events[eventID]; exists {
		event.Processed = true
		now := time.Now()
		event.ProcessedAt = &now
	}
	return nil
}

func (m *MemoryEventStorage) GetHistory(limit int) ([]*FileEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get all events and sort by time (most recent first)
	var allEvents []*FileEvent
	for _, event := range m.events {
		allEvents = append(allEvents, event)
	}

	// Simple sort by time (in production, use more efficient sorting)
	for i := 0; i < len(allEvents)-1; i++ {
		for j := i + 1; j < len(allEvents); j++ {
			if allEvents[i].EventTime.Before(allEvents[j].EventTime) {
				allEvents[i], allEvents[j] = allEvents[j], allEvents[i]
			}
		}
	}

	if limit > 0 && len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}

	return allEvents, nil
}

// FileWatcher watches for file changes in MinIO buckets
type FileWatcher struct {
	client     *minio.Client
	storage    EventStorage
	bucketName string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// Event handlers
	onEvent func(*FileEvent)

	// Configuration
	pollInterval time.Duration
}

// Config holds configuration for the file watcher
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	BucketName      string
	PollInterval    time.Duration
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(config Config, storage EventStorage) (*FileWatcher, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}

	return &FileWatcher{
		client:       client,
		storage:      storage,
		bucketName:   config.BucketName,
		ctx:          ctx,
		cancel:       cancel,
		pollInterval: config.PollInterval,
	}, nil
}

// SetEventHandler sets the event handler function
func (fw *FileWatcher) SetEventHandler(handler func(*FileEvent)) {
	fw.onEvent = handler
}

// Start starts the file watcher
func (fw *FileWatcher) Start() error {
	// Check if bucket exists
	exists, err := fw.client.BucketExists(fw.ctx, fw.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", fw.bucketName)
	}

	fw.wg.Add(1)
	go fw.watchLoop()

	log.Printf("File watcher started for bucket: %s", fw.bucketName)
	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	fw.cancel()
	fw.wg.Wait()
	log.Println("File watcher stopped")
}

// watchLoop runs the main watching loop
func (fw *FileWatcher) watchLoop() {
	defer fw.wg.Done()

	ticker := time.NewTicker(fw.pollInterval)
	defer ticker.Stop()

	// Get initial state
	lastKnownObjects := make(map[string]string)
	err := fw.updateObjectState(lastKnownObjects)
	if err != nil {
		log.Printf("Error getting initial object state: %v", err)
	}

	for {
		select {
		case <-fw.ctx.Done():
			return
		case <-ticker.C:
			currentObjects := make(map[string]string)
			err := fw.updateObjectState(currentObjects)
			if err != nil {
				log.Printf("Error updating object state: %v", err)
				continue
			}

			// Detect changes
			fw.detectChanges(lastKnownObjects, currentObjects)

			// Update last known state
			lastKnownObjects = currentObjects
		}
	}
}

// updateObjectState gets the current state of all objects in the bucket
func (fw *FileWatcher) updateObjectState(state map[string]string) error {
	ctx, cancel := context.WithTimeout(fw.ctx, 30*time.Second)
	defer cancel()

	objectsCh := fw.client.ListObjects(ctx, fw.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return object.Err
		}
		state[object.Key] = object.ETag
	}

	return nil
}

// detectChanges compares two states and creates events for changes
func (fw *FileWatcher) detectChanges(oldState, newState map[string]string) {
	// Check for new and modified objects
	for key, newETag := range newState {
		oldETag, exists := oldState[key]
		if !exists {
			// New object
			fw.createObjectEvent(key, EventCreated)
		} else if oldETag != newETag {
			// Modified object
			fw.createObjectEvent(key, EventMetadata)
		}
	}

	// Check for deleted objects
	for key := range oldState {
		if _, exists := newState[key]; !exists {
			// Deleted object
			fw.createObjectEvent(key, EventRemoved)
		}
	}
}

// createObjectEvent creates and processes a file event
func (fw *FileWatcher) createObjectEvent(key string, eventType EventType) {
	ctx, cancel := context.WithTimeout(fw.ctx, 10*time.Second)
	defer cancel()

	// Get object info
	objInfo, err := fw.client.StatObject(ctx, fw.bucketName, key, minio.StatObjectOptions{})
	if err != nil && eventType != EventRemoved {
		log.Printf("Error getting object info for %s: %v", key, err)
		return
	}

	event := &FileEvent{
		ID:        fmt.Sprintf("%s-%d", key, time.Now().UnixNano()),
		Bucket:    fw.bucketName,
		Key:       key,
		EventType: eventType,
		EventTime: time.Now(),
		Processed: false,
	}

	if eventType != EventRemoved {
		event.Size = objInfo.Size
		event.ETag = objInfo.ETag
		// Convert http.Header to map[string]string
		event.Metadata = make(map[string]string)
		for k, v := range objInfo.Metadata {
			if len(v) > 0 {
				event.Metadata[k] = v[0]
			}
		}
	}

	// Store event
	err = fw.storage.Store(event)
	if err != nil {
		log.Printf("Error storing event: %v", err)
		return
	}

	// Call event handler if set
	if fw.onEvent != nil {
		fw.onEvent(event)
	}

	log.Printf("File event created: %s - %s", eventType, key)
}

// GetUnprocessedEvents returns unprocessed events
func (fw *FileWatcher) GetUnprocessedEvents(limit int) ([]*FileEvent, error) {
	return fw.storage.GetUnprocessed(limit)
}

// MarkEventProcessed marks an event as processed
func (fw *FileWatcher) MarkEventProcessed(eventID string) error {
	return fw.storage.MarkProcessed(eventID)
}

// GetEventHistory returns event history
func (fw *FileWatcher) GetEventHistory(limit int) ([]*FileEvent, error) {
	return fw.storage.GetHistory(limit)
}
