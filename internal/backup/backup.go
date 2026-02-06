package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// Import extracts a ZIP backup to the data directory
func Import(zipPath, dataDir string) (*Metadata, error) {
	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	var metadata Metadata

	// Extract all files
	for _, file := range reader.File {
		// Read metadata file first
		if file.Name == "metadata.json" {
			metadataReader, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open metadata file: %w", err)
			}
			defer metadataReader.Close()

			if err := json.NewDecoder(metadataReader).Decode(&metadata); err != nil {
				return nil, fmt.Errorf("failed to decode metadata: %w", err)
			}
			continue
		}

		// Skip non-data files
		if filepath.Dir(file.Name) == "." {
			continue
		}

		// Remove "data/" prefix to get target path
		targetPath := filepath.Join(dataDir, filepath.Base(filepath.Dir(file.Name)), filepath.Base(file.Name))

		// Handle directory entries
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
		}

		// Extract file
		outFile, err := os.Create(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s: %w", targetPath, err)
		}

		fileReader, err := file.Open()
		if err != nil {
			outFile.Close()
			return nil, fmt.Errorf("failed to open file in ZIP: %w", err)
		}

		if _, err := io.Copy(outFile, fileReader); err != nil {
			outFile.Close()
			fileReader.Close()
			return nil, fmt.Errorf("failed to extract file %s: %w", targetPath, err)
		}

		outFile.Close()
		fileReader.Close()
	}

	return &metadata, nil
}
