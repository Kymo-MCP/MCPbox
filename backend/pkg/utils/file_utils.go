package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"qm-mcp-server/api/market/code"
	"qm-mcp-server/internal/authz/config"
	"qm-mcp-server/pkg/logger"

	"go.uber.org/zap"
)

// BuildFileTreeRecursive recursively builds a file tree structure
func BuildFileTreeRecursive(currentPath, rootPath, relativePath string) (*code.FileNode, error) {
	info, err := os.Stat(currentPath)
	if err != nil {
		return nil, err
	}

	node := &code.FileNode{
		Name:  filepath.Base(currentPath),
		Path:  relativePath,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		node.Size = info.Size()
		return node, nil
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(currentPath, entry.Name())

		// Calculate the relative path of the child node
		var childRelativePath string
		if relativePath == "" {
			childRelativePath = entry.Name()
		} else {
			childRelativePath = filepath.Join(relativePath, entry.Name())
		}

		childNode, err := BuildFileTreeRecursive(childPath, rootPath, childRelativePath)
		if err != nil {
			continue // Skip inaccessible files
		}

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// CreatePackageZip creates a zip package
func CreatePackageZip(extractedPath, zipFilePath string) error {
	// Create zip package
	return CreateZipFile(extractedPath, zipFilePath)
}

// CreateZipFile creates a zip file, overwriting if it exists
func CreateZipFile(sourcePath, zipFilePath string) error {
	// Check if the target file exists, delete if it does (to overwrite)
	if _, err := os.Stat(zipFilePath); err == nil {
		if err := os.Remove(zipFilePath); err != nil {
			return fmt.Errorf("failed to remove existing zip file: %w", err)
		}
	}

	// Create zip file
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Get source path information
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	if sourceInfo.IsDir() {
		// If it's a directory, recursively add all files
		return AddDirToZip(zipWriter, sourcePath, "")
	} else {
		// If it's a single file, add it directly
		return AddFileToZip(zipWriter, sourcePath, sourceInfo.Name())
	}
}

// AddDirToZip recursively adds a directory to the zip archive
func AddDirToZip(zipWriter *zip.Writer, dirPath, baseInZip string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path in the zip archive
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseInZip, relPath)
		// Unify path separators to forward slashes
		zipPath = filepath.ToSlash(zipPath)

		if info.IsDir() {
			// Create directory entry
			if zipPath != "" && zipPath != "." {
				_, err := zipWriter.Create(zipPath + "/")
				return err
			}
			return nil
		}

		// Add file
		return AddFileToZip(zipWriter, path, zipPath)
	})
}

// AddFileToZip adds a single file to the zip archive
func AddFileToZip(zipWriter *zip.Writer, filePath, nameInZip string) error {
	// Open source file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Create file in zip archive
	writer, err := zipWriter.Create(nameInZip)
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %w", err)
	}

	// Copy file content
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// CheckExtractedPathNotEmpty checks if the extracted path is not empty
func CheckExtractedPathNotEmpty(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	// If it's a file, it's considered not empty
	if !info.IsDir() {
		return nil
	}

	// If it's a directory, check if it's empty
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("extracted path is empty: %s", path)
	}
	return nil
}

func MkdirP(path string) error {
	if !FileExists(path) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
		// fmt.Printf("Created directory: %v \n", path)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsValidImageType validates if the given byte data represents a valid image type
// Supports JPEG, PNG, GIF, and WebP formats by checking magic numbers
func IsValidImageType(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check JPEG magic number (FF D8 FF)
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return true
	}

	// Check PNG magic number (89 50 4E 47)
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return true
	}

	// Check GIF magic number (47 49 46 38)
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return true
	}

	// Check WebP magic number (52 49 46 46 ... 57 45 42 50)
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return true
	}

	return false
}

// GetImageFileExtension returns the file extension based on image magic numbers
func GetImageFileExtension(data []byte) string {
	if len(data) < 4 {
		return ""
	}

	// Check JPEG magic number
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}

	// Check PNG magic number
	if len(data) >= 4 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// Check GIF magic number
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "gif"
	}

	// Check WebP magic number
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "webp"
	}

	return ""
}

func SaveImageFile(data []byte, filePath string) string {
	// Construct storage directory path
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, "storage")

	// Ensure directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		return ""
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logger.Error("Failed to write image file", zap.Error(err))
		return ""
	}

	logger.Info("Image file saved successfully", zap.String("path", filePath))
	return filePath
}
