package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thomas/mavt/internal/version"
)

// Metadata represents backup file metadata
type Metadata struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	AppCount  int       `json:"app_count"`
}

// Export creates a ZIP backup of the data directory
func Export(dataDir, outputPath string) error {
	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// Count apps for metadata
	appsDir := filepath.Join(dataDir, "apps")
	appCount := 0
	if entries, err := os.ReadDir(appsDir); err == nil {
		appCount = len(entries)
	}

	// Create and write metadata file
	metadata := Metadata{
		Version:   version.Version,
		CreatedAt: time.Now(),
		AppCount:  appCount,
	}
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	metadataWriter, err := zipWriter.Create("metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create metadata file in ZIP: %w", err)
	}
	if _, err := metadataWriter.Write(metadataJSON); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Walk the data directory and add all files to ZIP
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the data directory itself
		if path == dataDir {
			return nil
		}

		// Get relative path from data directory
		relPath, err := filepath.Rel(dataDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Create ZIP entry with "data/" prefix to maintain structure
		zipPath := filepath.Join("data", relPath)

		if info.IsDir() {
			// Add directory entry
			_, err := zipWriter.Create(zipPath + "/")
			return err
		}

		// Add file
		fileWriter, err := zipWriter.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to create ZIP entry for %s: %w", zipPath, err)
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		if _, err := io.Copy(fileWriter, file); err != nil {
			return fmt.Errorf("failed to write file %s to ZIP: %w", zipPath, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk data directory: %w", err)
	}

	return nil
}

// Import extracts a ZIP backup to the data directory with security validation
func Import(zipPath, dataDir string) (*Metadata, error) {
	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	var metadata Metadata
	const (
		maxFileSize       = 5 * 1024 * 1024  // 5MB per file (generous for JSON)
		maxTotalSize      = 50 * 1024 * 1024 // 50MB total extraction size
		maxFiles          = 2000             // Maximum number of files (~1000 apps)
	)

	var totalExtractedSize int64
	fileCount := 0

	// Get absolute path for data directory to prevent traversal
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute data directory path: %w", err)
	}

	// First pass: Read and validate metadata
	metadataFound := false
	for _, file := range reader.File {
		if file.Name == "metadata.json" {
			if file.UncompressedSize64 > 1024*1024 { // 1MB limit for metadata
				return nil, fmt.Errorf("metadata file too large")
			}

			metadataReader, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open metadata file: %w", err)
			}
			defer metadataReader.Close()

			if err := json.NewDecoder(metadataReader).Decode(&metadata); err != nil {
				return nil, fmt.Errorf("failed to decode metadata: %w", err)
			}
			metadataFound = true
			break
		}
	}

	if !metadataFound {
		return nil, fmt.Errorf("backup metadata not found - this may not be a valid MAVT backup")
	}

	// Second pass: Validate and extract files
	for _, file := range reader.File {
		fileCount++
		if fileCount > maxFiles {
			return nil, fmt.Errorf("backup contains too many files (max %d)", maxFiles)
		}

		// Skip metadata (already processed)
		if file.Name == "metadata.json" {
			continue
		}

		// Validate file size to prevent zip bombs
		if file.UncompressedSize64 > maxFileSize {
			return nil, fmt.Errorf("file %s exceeds maximum size of %d bytes", file.Name, maxFileSize)
		}

		totalExtractedSize += int64(file.UncompressedSize64)
		if totalExtractedSize > maxTotalSize {
			return nil, fmt.Errorf("total extracted size exceeds maximum of %d bytes", maxTotalSize)
		}

		// Skip non-data files
		if filepath.Dir(file.Name) == "." {
			continue
		}

		// Validate path structure: should be data/apps/*.json or data/updates/*.json
		parts := filepath.SplitList(filepath.ToSlash(file.Name))
		if len(parts) < 3 || parts[0] != "data" {
			// On Windows, SplitList uses ; as separator, use custom split
			pathParts := strings.Split(filepath.ToSlash(file.Name), "/")
			if len(pathParts) < 3 || pathParts[0] != "data" {
				return nil, fmt.Errorf("invalid file path structure: %s", file.Name)
			}
			parts = pathParts
		}

		// Validate subdirectory is "apps" or "updates"
		subDir := parts[1]
		if subDir != "apps" && subDir != "updates" {
			return nil, fmt.Errorf("invalid subdirectory: %s (expected 'apps' or 'updates')", subDir)
		}

		// Handle directory entries first (before filename validation)
		// Directory paths may have empty last part (e.g., "data/apps/" splits to ["data", "apps", ""])
		if file.FileInfo().IsDir() {
			// For directories, construct path from the full file.Name
			targetPath := filepath.Join(absDataDir, strings.TrimPrefix(filepath.ToSlash(file.Name), "data/"))
			absTargetPath, err := filepath.Abs(targetPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve target path: %w", err)
			}

			// Security check: ensure target path is within data directory
			if !strings.HasPrefix(absTargetPath, absDataDir) {
				return nil, fmt.Errorf("path traversal attempt detected: %s", file.Name)
			}

			if err := os.MkdirAll(absTargetPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", absTargetPath, err)
			}
			continue
		}

		fileName := parts[len(parts)-1]

		// Sanitize and validate filename to prevent directory traversal
		cleanFileName := filepath.Base(fileName)
		if cleanFileName != fileName || strings.Contains(fileName, "..") {
			return nil, fmt.Errorf("suspicious filename detected: %s", file.Name)
		}

		// Construct target path and ensure it's within data directory
		targetPath := filepath.Join(absDataDir, subDir, cleanFileName)
		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve target path: %w", err)
		}

		// Security check: ensure target path is within data directory
		if !strings.HasPrefix(absTargetPath, absDataDir) {
			return nil, fmt.Errorf("path traversal attempt detected: %s", file.Name)
		}

		// Validate filename extension (only for files, not directories)
		if !strings.HasSuffix(fileName, ".json") {
			return nil, fmt.Errorf("invalid file extension: %s (expected .json)", fileName)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(absTargetPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s: %w", absTargetPath, err)
		}

		// Extract and validate file content
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in ZIP: %w", err)
		}

		// Read content with size limit
		limitedReader := io.LimitReader(fileReader, maxFileSize+1)
		content, err := io.ReadAll(limitedReader)
		fileReader.Close()

		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		if int64(len(content)) > maxFileSize {
			return nil, fmt.Errorf("file %s exceeds size limit after decompression", file.Name)
		}

		// Validate JSON format
		var jsonCheck interface{}
		if err := json.Unmarshal(content, &jsonCheck); err != nil {
			return nil, fmt.Errorf("invalid JSON in file %s: %w", file.Name, err)
		}

		// Write validated content to file
		if err := os.WriteFile(absTargetPath, content, 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", absTargetPath, err)
		}
	}

	return &metadata, nil
}
