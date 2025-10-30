package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bronze-backend/config"
	"bronze-backend/handlers"
	"bronze-backend/minio"
	"bronze-backend/processor"
	"bronze-backend/routes"
	"bronze-backend/watcher"
)

func main() {
	log.Println("Starting Bronze Backend...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Server: %s", cfg.GetServerAddr())
	log.Printf("MinIO: %s (bucket: %s)", cfg.MinIO.Endpoint, cfg.MinIO.Bucket)
	log.Printf("Workers: %d", cfg.Processing.MaxWorkers)

	minioClient, err := minio.NewMinIOClient(&cfg.MinIO)
	if err != nil {
		log.Printf("Warning: Failed to create MinIO client: %v", err)
		log.Println("MinIO features will be disabled until connection is restored")
		// Continue without MinIO for now
		minioClient = nil
	} else {
		log.Println("MinIO client created successfully")
	}

	fileProcessor := processor.NewFileProcessor(cfg)
	log.Println("File processor created successfully")

	jobQueue := processor.NewJobQueue(cfg.Processing.MaxWorkers, cfg.Processing.QueueSize)
	log.Println("Job queue created successfully")

	workerPool := processor.NewWorkerPool(cfg.Processing.MaxWorkers, jobQueue, fileProcessor)
	workerPool.Start()
	log.Printf("Worker pool started with %d workers", cfg.Processing.MaxWorkers)

	// Create file watcher (only if MinIO is available)
	var fileWatcher *watcher.FileWatcher
	if minioClient != nil {
		eventStorage := watcher.NewMemoryEventStorage()
		watcherConfig := watcher.Config{
			Endpoint:        cfg.MinIO.Endpoint,
			AccessKeyID:     cfg.MinIO.AccessKey,
			SecretAccessKey: cfg.MinIO.SecretKey,
			UseSSL:          cfg.MinIO.UseSSL(),
			Region:          cfg.MinIO.Region,
			BucketName:      cfg.MinIO.Bucket,
			PollInterval:    30 * time.Second,
		}

		fileWatcher, err = watcher.NewFileWatcher(watcherConfig, eventStorage)
		if err != nil {
			log.Printf("Failed to create file watcher: %v", err)
			fileWatcher = nil
		} else {
			// Set up event handler for file changes
			fileWatcher.SetEventHandler(func(event *watcher.FileEvent) {
				log.Printf("File event detected: %s - %s", event.EventType, event.Key)

				// If this is a new file, create a processing job
				if event.EventType == watcher.EventCreated {
					job := processor.NewJob(
						"decompress",
						"", // local file path will be set by processor
						cfg.MinIO.Bucket,
						event.Key,
						processor.PriorityMedium,
					)

					if err := jobQueue.Enqueue(job); err != nil {
						log.Printf("Failed to create job for file %s: %v", event.Key, err)
					} else {
						log.Printf("Created processing job for file: %s", event.Key)
					}
				}
			})

			// Start file watcher
			if err := fileWatcher.Start(); err != nil {
				log.Printf("Failed to start file watcher: %v", err)
				fileWatcher = nil
			} else {
				log.Println("File watcher started successfully")
			}
		}
	} else {
		log.Println("File watcher disabled (MinIO not available)")
	}

	fileHandler := handlers.NewFileHandler(minioClient, fileProcessor)
	jobHandler := handlers.NewJobHandler(jobQueue, workerPool)
	watcherHandler := handlers.NewWatcherHandler(fileWatcher)

	router := routes.NewRouter(fileHandler, jobHandler, watcherHandler)
	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router.GetRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting HTTP server on %s", cfg.GetServerAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	workerPool.Stop()
	log.Println("Worker pool stopped")

	if fileWatcher != nil {
		fileWatcher.Stop()
		log.Println("File watcher stopped")
	}

	log.Println("Server exited")
}
