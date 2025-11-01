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
	"bronze-backend/data_browser"
	"bronze-backend/files"
	"bronze-backend/jobs"
	"bronze-backend/monitoring"
	"bronze-backend/routes"
	"bronze-backend/storage"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting Bronze Backend...")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	} else {
		if wd, err := os.Getwd(); err == nil {
			log.Printf("Loaded .env file from: %s/.env", wd)
		} else {
			log.Println("Loaded .env file successfully")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Server: %s", cfg.GetServerAddr())
	log.Printf("MinIO: %s (bucket: %s)", cfg.MinIO.Endpoint, cfg.MinIO.Bucket)
	log.Printf("Workers: %d", cfg.Processing.MaxWorkers)

	storageClient, err := storage.NewMinIOClient(&cfg.MinIO)
	if err != nil {
		log.Printf("Warning: Failed to create MinIO client: %v", err)
		log.Println("MinIO features will be disabled until connection is restored")
		storageClient = nil
	} else {
		log.Println("MinIO client created successfully")
	}

	nessieClient, err := storage.NewNessieClient(&cfg.Nessie)
	if err != nil {
		log.Printf("Warning: Failed to create Nessie client: %v", err)
		log.Println("Nessie export features will be disabled")
		nessieClient = nil
	} else {
		log.Println("Nessie client created successfully")

		fileProcessor := files.NewFileProcessor(cfg)
		log.Println("File processor created successfully")

		jobQueue := jobs.NewJobQueue(cfg.Processing.MaxWorkers, cfg.Processing.QueueSize)
		log.Println("Job queue created successfully")

		workerPool := jobs.NewWorkerPool(cfg.Processing.MaxWorkers, jobQueue, fileProcessor)
		workerPool.Start()
		log.Printf("Worker pool started with %d workers", cfg.Processing.MaxWorkers)

		// Create file watcher (disabled for now to avoid startup issues)
		var fileWatcher *monitoring.FileWatcher
		log.Println("File watcher disabled")

		fileHandler := files.NewFileHandlerWithQueue(storageClient, fileProcessor, jobQueue)
		jobHandler := jobs.NewJobHandler(jobQueue, workerPool)
		watcherHandler := monitoring.NewWatcherHandler(fileWatcher)
		dataBrowserHandler := data_browser.NewDataBrowserHandler(storageClient)
		exportHandler := data_browser.NewExportHandler(storageClient, nessieClient, cfg, dataBrowserHandler)

		router := routes.NewRouter(fileHandler, jobHandler, watcherHandler, dataBrowserHandler, exportHandler)
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
}
