package processor

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Decompressor struct {
	config DecompressionConfig
}

type DecompressionConfig struct {
	MaxExtractSize     string
	MaxFilesPerArchive int
	NestedArchiveDepth int
	PasswordProtected  bool
	ExtractToSubfolder bool
}

func NewDecompressor(config DecompressionConfig) *Decompressor {
	return &Decompressor{
		config: config,
	}
}

type ArchiveInfo struct {
	Format      string         `json:"format"`
	IsArchive   bool           `json:"is_archive"`
	FileCount   int            `json:"file_count,omitempty"`
	TotalSize   int64          `json:"total_size,omitempty"`
	HasPassword bool           `json:"has_password,omitempty"`
	Files       []string       `json:"files,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ExtractionResult struct {
	Success        bool        `json:"success"`
	ExtractedFiles []string    `json:"extracted_files"`
	FileCount      int         `json:"file_count"`
	Message        string      `json:"message"`
	ArchiveInfo    ArchiveInfo `json:"archive_info"`
}

func (d *Decompressor) DetectArchive(filePath string) (ArchiveInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return ArchiveInfo{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ArchiveInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	format, isArchive := d.getArchiveFormat(ext, baseName)

	info := ArchiveInfo{
		Format:    format,
		IsArchive: isArchive,
		TotalSize: stat.Size(),
		Metadata:  make(map[string]any),
	}

	if !isArchive {
		return info, nil
	}

	if format == "" {
		return info, fmt.Errorf("unsupported archive format")
	}

	info.Metadata["extension"] = ext
	info.Metadata["base_name"] = baseName

	return info, nil
}

func (d *Decompressor) ExtractArchive(filePath, outputDir string, password string) (ExtractionResult, error) {
	result := ExtractionResult{}

	info, err := d.DetectArchive(filePath)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to detect archive: %v", err)
		return result, err
	}

	result.ArchiveInfo = info

	if !info.IsArchive {
		result.Success = false
		result.Message = "File is not an archive"
		return result, fmt.Errorf("file is not an archive")
	}

	extractDir := outputDir
	if d.config.ExtractToSubfolder {
		baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		extractDir = filepath.Join(outputDir, baseName)
	}

	if err := os.MkdirAll(extractDir, 0755); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to create extract directory: %v", err)
		return result, err
	}

	extractedFiles, err := d.extractFiles(filePath, extractDir, password)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("Failed to extract archive: %v", err)
		return result, err
	}

	result.Success = true
	result.ExtractedFiles = extractedFiles
	result.FileCount = len(extractedFiles)
	result.Message = fmt.Sprintf("Successfully extracted %d files", len(extractedFiles))

	return result, nil
}

func (d *Decompressor) getArchiveFormat(ext, baseName string) (string, bool) {
	archiveFormats := map[string]string{
		".zip":     "zip",
		".tar":     "tar",
		".gz":      "gzip",
		".tgz":     "tar.gz",
		".tar.gz":  "tar.gz",
		".bz2":     "bzip2",
		".tbz2":    "tar.bz2",
		".tar.bz2": "tar.bz2",
		".xz":      "xz",
		".txz":     "tar.xz",
		".tar.xz":  "tar.xz",
	}

	if format, exists := archiveFormats[ext]; exists {
		return format, true
	}

	for ext := range archiveFormats {
		if strings.HasSuffix(baseName, ext) {
			return archiveFormats[ext], true
		}
	}

	return "", false
}

func (d *Decompressor) extractFiles(filePath, outputDir, password string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	var extractedFiles []string
	var err error

	switch {
	case ext == ".zip":
		extractedFiles, err = d.extractZip(filePath, outputDir, password)
	case ext == ".tar":
		extractedFiles, err = d.extractTar(filePath, outputDir)
	case ext == ".gz" || strings.HasSuffix(baseName, ".tar.gz"):
		extractedFiles, err = d.extractTarGz(filePath, outputDir)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", ext)
	}

	return extractedFiles, err
}

func (d *Decompressor) extractZip(filePath, outputDir, password string) ([]string, error) {
	var extractedFiles []string

	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		outputPath := filepath.Join(outputDir, file.Name)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, err
		}

		outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			return nil, err
		}

		fileReader, err := file.Open()
		if err != nil {
			outputFile.Close()
			return nil, err
		}

		_, err = io.Copy(outputFile, fileReader)
		fileReader.Close()
		outputFile.Close()

		if err != nil {
			return nil, err
		}

		extractedFiles = append(extractedFiles, outputPath)
	}

	return extractedFiles, nil
}

func (d *Decompressor) extractTar(filePath, outputDir string) ([]string, error) {
	var extractedFiles []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := tar.NewReader(file)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		outputPath := filepath.Join(outputDir, header.Name)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, err
		}

		outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(outputFile, reader)
		outputFile.Close()

		if err != nil {
			return nil, err
		}

		extractedFiles = append(extractedFiles, outputPath)
	}

	return extractedFiles, nil
}

func (d *Decompressor) extractTarGz(filePath, outputDir string) ([]string, error) {
	var extractedFiles []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	reader := tar.NewReader(gzReader)

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		outputPath := filepath.Join(outputDir, header.Name)

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return nil, err
		}

		outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(outputFile, reader)
		outputFile.Close()

		if err != nil {
			return nil, err
		}

		extractedFiles = append(extractedFiles, outputPath)
	}

	return extractedFiles, nil
}

func (d *Decompressor) GetSupportedFormats() []string {
	return []string{
		"zip", "tar", "tar.gz",
	}
}
