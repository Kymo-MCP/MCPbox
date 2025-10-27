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

// BuildFileTreeRecursive 递归构建文件树结构
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

		// 计算子节点的相对路径
		var childRelativePath string
		if relativePath == "" {
			childRelativePath = entry.Name()
		} else {
			childRelativePath = filepath.Join(relativePath, entry.Name())
		}

		childNode, err := BuildFileTreeRecursive(childPath, rootPath, childRelativePath)
		if err != nil {
			continue // 跳过无法访问的文件
		}

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

// CreatePackageZip 创建压缩包
func CreatePackageZip(extractedPath, zipFilePath string) error {
	// 创建压缩包
	return CreateZipFile(extractedPath, zipFilePath)
}

// CreateZipFile 创建压缩包文件，存在则覆盖
func CreateZipFile(sourcePath, zipFilePath string) error {
	// 检查目标文件是否存在，如果存在则删除（实现覆盖）
	if _, err := os.Stat(zipFilePath); err == nil {
		if err := os.Remove(zipFilePath); err != nil {
			return fmt.Errorf("failed to remove existing zip file: %w", err)
		}
	}

	// 创建压缩包文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// 创建 zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 获取源路径信息
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	if sourceInfo.IsDir() {
		// 如果是目录，递归添加所有文件
		return AddDirToZip(zipWriter, sourcePath, "")
	} else {
		// 如果是单个文件，直接添加
		return AddFileToZip(zipWriter, sourcePath, sourceInfo.Name())
	}
}

// AddDirToZip 递归添加目录到压缩包
func AddDirToZip(zipWriter *zip.Writer, dirPath, baseInZip string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算在压缩包中的相对路径
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseInZip, relPath)
		// 统一使用正斜杠作为路径分隔符
		zipPath = filepath.ToSlash(zipPath)

		if info.IsDir() {
			// 创建目录条目
			if zipPath != "" && zipPath != "." {
				_, err := zipWriter.Create(zipPath + "/")
				return err
			}
			return nil
		}

		// 添加文件
		return AddFileToZip(zipWriter, path, zipPath)
	})
}

// AddFileToZip 添加单个文件到压缩包
func AddFileToZip(zipWriter *zip.Writer, filePath, nameInZip string) error {
	// 打开源文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// 在压缩包中创建文件
	writer, err := zipWriter.Create(nameInZip)
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %w", err)
	}

	// 复制文件内容
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// CheckExtractedPathNotEmpty 检查解压路径是否不为空
func CheckExtractedPathNotEmpty(path string) error {
	// 检查路径是否存在
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	// 如果是文件，则认为不为空
	if !info.IsDir() {
		return nil
	}

	// 如果是目录，检查是否为空
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
		fmt.Printf("创建文件夹: %v \n", path)
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
	// 构建存储目录路径
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, "storage")

	// 确保目录存在
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("创建存储目录失败", zap.Error(err))
		return ""
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logger.Error("写入图片文件失败", zap.Error(err))
		return ""
	}

	logger.Info("图片文件保存成功", zap.String("path", filePath))
	return filePath
}
