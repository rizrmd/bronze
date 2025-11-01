package files

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"bronze-backend/config"
	"bronze-backend/jobs"
)

type FileProcessor struct {
	decompressor *ArchiveExtractor
	config       *config.Config
}

func NewFileProcessor(cfg *config.Config) *FileProcessor {
	decompressorConfig := DecompressionConfig{
		MaxExtractSize:     cfg.Processing.Decompression.MaxExtractSize,
		MaxFilesPerArchive: cfg.Processing.Decompression.MaxFilesPerArchive,
		NestedArchiveDepth: cfg.Processing.Decompression.NestedArchiveDepth,
		PasswordProtected:  cfg.Processing.Decompression.PasswordProtected,
		ExtractToSubfolder: cfg.Processing.Decompression.ExtractToSubfolder,
	}

	return &FileProcessor{
		decompressor: NewArchiveExtractor(decompressorConfig),
		config:       cfg,
	}
}

type JobProcessor interface {
	ProcessJob(ctx context.Context, job jobs.Job) jobs.JobResult
}

func (fp *FileProcessor) ProcessJob(ctx context.Context, job *jobs.Job) jobs.JobResult {
	startTime := time.Now()

	log.Printf("Processing job %s: %s/%s", job.ID, job.Bucket, job.ObjectName)

	job.UpdateProgress(10)

	tempFilePath, err := fp.downloadFileFromMinIO(ctx, job)
	if err != nil {
		return jobs.JobResult{
			Success:        false,
			ProcessingTime: time.Since(startTime),
			Message:        fmt.Sprintf("Failed to download file: %v", err),
		}
	}
	defer os.Remove(tempFilePath)

	job.UpdateProgress(30)

	archiveInfo, err := fp.decompressor.DetectArchive(tempFilePath)
	if err != nil {
		return jobs.JobResult{
			Success:        false,
			ProcessingTime: time.Since(startTime),
			Message:        fmt.Sprintf("Failed to detect archive: %v", err),
		}
	}

	job.UpdateProgress(50)

	result := jobs.JobResult{
		Success:        true,
		ProcessingTime: time.Since(startTime),
		FileInfo: map[string]any{
			"archive_info": archiveInfo,
			"file_size":    archiveInfo.TotalSize,
			"format":       archiveInfo.Format,
		},
	}

	if archiveInfo.IsArchive {
		job.UpdateProgress(60)

		extractDir := filepath.Join(fp.config.Processing.TempDir, job.ID)
		extractionResult, err := fp.decompressor.ExtractArchive(tempFilePath, extractDir, "")
		if err != nil {
			return jobs.JobResult{
				Success:        false,
				ProcessingTime: time.Since(startTime),
				Message:        fmt.Sprintf("Failed to extract archive: %v", err),
			}
		}

		result.ExtractedFiles = extractionResult.ExtractedFiles
		result.FileInfo["extracted_files"] = extractionResult.ExtractedFiles
		result.FileInfo["extraction_result"] = extractionResult

		job.UpdateProgress(80)

		if err := fp.processExtractedFiles(ctx, job, extractionResult.ExtractedFiles); err != nil {
			log.Printf("Warning: Failed to process extracted files: %v", err)
		}

		defer os.RemoveAll(extractDir)
	}

	job.UpdateProgress(90)

	if err := fp.uploadProcessedResults(ctx, job, result); err != nil {
		log.Printf("Warning: Failed to upload processed results: %v", err)
	}

	job.UpdateProgress(100)

	result.Message = fmt.Sprintf("Successfully processed file %s", job.ObjectName)
	log.Printf("Completed job %s in %v", job.ID, time.Since(startTime))

	return result
}

func (fp *FileProcessor) downloadFileFromMinIO(ctx context.Context, job *jobs.Job) (string, error) {
	tempFilePath := filepath.Join(fp.config.Processing.TempDir, job.ID+"_"+job.ObjectName)

	file, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	return tempFilePath, nil
}

func (fp *FileProcessor) processExtractedFiles(ctx context.Context, job *jobs.Job, extractedFiles []string) error {
	for _, filePath := range extractedFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fp.processSingleFile(ctx, job, filePath); err != nil {
				log.Printf("Failed to process extracted file %s: %v", filePath, err)
			}
		}
	}
	return nil
}

func (fp *FileProcessor) processSingleFile(ctx context.Context, job *jobs.Job, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	fileInfo := map[string]any{
		"name":        filepath.Base(filePath),
		"path":        filePath,
		"size":        info.Size(),
		"modified":    info.ModTime(),
		"mode":        info.Mode().String(),
		"parent_job":  job.ID,
		"parent_file": job.ObjectName,
	}

	job.Metadata["extracted_file_"+filepath.Base(filePath)] = fileInfo

	return nil
}

func (fp *FileProcessor) uploadProcessedResults(ctx context.Context, job *jobs.Job, result jobs.JobResult) error {
	resultsPath := filepath.Join(fp.config.Processing.TempDir, job.ID+"_results.json")

	file, err := os.Create(resultsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func (fp *FileProcessor) GetSupportedFormats() []string {
	return fp.decompressor.GetSupportedFormats()
}

func (fp *FileProcessor) GetProcessingStats() map[string]any {
	return map[string]any{
		"supported_formats": fp.GetSupportedFormats(),
		"temp_dir":          fp.config.Processing.TempDir,
		"max_workers":       fp.config.Processing.MaxWorkers,
		"decompression": map[string]any{
			"enabled":               fp.config.Processing.Decompression.Enabled,
			"max_extract_size":      fp.config.Processing.Decompression.MaxExtractSize,
			"max_files_per_archive": fp.config.Processing.Decompression.MaxFilesPerArchive,
			"nested_archive_depth":  fp.config.Processing.Decompression.NestedArchiveDepth,
			"password_protected":    fp.config.Processing.Decompression.PasswordProtected,
			"extract_to_subfolder":  fp.config.Processing.Decompression.ExtractToSubfolder,
		},
	}
}
