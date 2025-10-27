package codepackage

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CodePackageManager 代码包管理器
type CodePackageManager struct {
	config     *common.CodeConfig
	pathPrefix string
}

// NewCodePackageManager 创建代码包管理器实例
func NewCodePackageManager(codeConfig *common.CodeConfig, pathPrefix string) *CodePackageManager {
	return &CodePackageManager{
		config:     codeConfig,
		pathPrefix: pathPrefix,
	}
}

// PackageInfo 包信息结构
type PackageInfo struct {
	PackageID     string
	PackagePath   string
	OriginalPath  string
	ExtractedPath string
	OriginalName  string
	FileSize      int64
	PackageType   model.PackageType
}

// UploadAndExtractPackage 上传并解压代码包
func (m *CodePackageManager) UploadAndExtractPackage(file multipart.File, header *multipart.FileHeader) (*PackageInfo, error) {
	// 记录上传开始信息
	logger.Info("Starting code package upload",
		zap.String("filename", header.Filename),
		zap.Int64("fileSize", header.Size),
		zap.Int("configMaxSizeMB", m.config.Upload.MaxFileSize))

	// 验证文件类型
	packageType, err := m.validateFileType(header.Filename)
	if err != nil {
		logger.Error("File type validation failed",
			zap.String("filename", header.Filename),
			zap.Error(err))
		return nil, err
	}

	// 验证文件大小
	if err := m.validateFileSize(header.Size); err != nil {
		logger.Error("File size validation failed",
			zap.String("filename", header.Filename),
			zap.Int64("fileSize", header.Size),
			zap.Int("maxSizeMB", m.config.Upload.MaxFileSize),
			zap.Error(err))
		return nil, err
	}

	// 生成包ID
	packageID := uuid.New().String()

	// 创建包目录结构
	packageDir, err := m.createPackageDirectory(packageID)
	if err != nil {
		return nil, fmt.Errorf("failed to create package directory: %v", err)
	}

	// 保存原始压缩包
	originalPath, err := m.saveOriginalPackage(file, packageDir, header.Filename)
	if err != nil {
		// 清理目录
		os.RemoveAll(packageDir)
		return nil, fmt.Errorf("failed to save original package: %v", err)
	}

	// 解压包到同级目录
	extractedPath, err := m.extractPackage(originalPath, packageDir, packageType)
	if err != nil {
		// 清理目录
		os.RemoveAll(packageDir)
		return nil, fmt.Errorf("failed to extract package: %v", err)
	}

	// 转换为相对于配置根路径的相对路径
	relPackagePath, _ := m.ToRelativePath(packageDir)
	relOriginalPath, _ := m.ToRelativePath(originalPath)
	relExtractedPath, _ := m.ToRelativePath(extractedPath)

	packageInfo := &PackageInfo{
		PackageID:     packageID,
		PackagePath:   relPackagePath,
		OriginalPath:  relOriginalPath,
		ExtractedPath: relExtractedPath,
		OriginalName:  header.Filename,
		FileSize:      header.Size,
		PackageType:   packageType,
	}

	logger.Info("Package uploaded and extracted successfully",
		zap.String("packageId", packageID),
		zap.String("originalPath", relOriginalPath),
		zap.String("extractedPath", relExtractedPath))

	return packageInfo, nil
}

// validateFileType 验证文件类型
func (m *CodePackageManager) validateFileType(filename string) (model.PackageType, error) {
	allowedExtensions := m.config.Upload.AllowedExtensions
	filename = strings.ToLower(filename)

	for _, ext := range allowedExtensions {
		if strings.HasSuffix(filename, ext) {
			if ext == ".zip" {
				return model.PackageTypeZip, nil
			} else if ext == ".tar" || ext == ".tar.gz" {
				return model.PackageTypeTar, nil
			}
		}
	}

	return "", fmt.Errorf("unsupported file type, allowed extensions: %v", allowedExtensions)
}

// validateFileSize 验证文件大小
func (m *CodePackageManager) validateFileSize(size int64) error {
	maxSize := int64(m.config.Upload.MaxFileSize) * 1024 * 1024 // 转换为字节
	if size > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d MB", size, m.config.Upload.MaxFileSize)
	}
	return nil
}

// createPackageDirectory 创建包目录
func (m *CodePackageManager) createPackageDirectory(packageID string) (string, error) {
	// 根据配置创建目录结构：root_path/package-{id}
	packageDirName := fmt.Sprintf("package-%s", packageID)
	packageDir := filepath.Join(m.pathPrefix, packageDirName)

	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", packageDir, err)
	}

	return packageDir, nil
}

// saveOriginalPackage 保存原始压缩包
func (m *CodePackageManager) saveOriginalPackage(file multipart.File, packageDir, filename string) (string, error) {
	// 重置文件指针到开始位置
	file.Seek(0, 0)

	originalPath := filepath.Join(packageDir, filename)
	outFile, err := os.Create(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", originalPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}

	return originalPath, nil
}

// extractPackage 解压包到同级目录
func (m *CodePackageManager) extractPackage(originalPath, packageDir string, packageType model.PackageType) (string, error) {
	// 记录解压开始时间
	startTime := time.Now()
	logger.Info("Starting package extraction",
		zap.String("originalPath", originalPath),
		zap.String("packageType", string(packageType)))

	// 获取压缩包文件名（不含扩展名）作为解压目录名
	originalFileName := filepath.Base(originalPath)
	var extractDirName string

	// 根据不同的压缩包类型去除相应的扩展名
	switch packageType {
	case model.PackageTypeTar:
		extractDirName = strings.TrimSuffix(originalFileName, ".tar")
	case model.PackageTypeZip:
		extractDirName = strings.TrimSuffix(originalFileName, ".zip")
	case model.PackageTypeTarGz:
		// 处理 .tar.gz 和 .gz 扩展名
		if strings.HasSuffix(originalFileName, ".tar.gz") {
			extractDirName = strings.TrimSuffix(originalFileName, ".tar.gz")
		} else {
			extractDirName = strings.TrimSuffix(originalFileName, ".gz")
		}
	case model.PackageTypeDxt:
		extractDirName = strings.TrimSuffix(originalFileName, ".dxt")
	case model.PackageTypeMcpb:
		extractDirName = strings.TrimSuffix(originalFileName, ".mcpb")
	default:
		// 默认情况下，去除最后一个点后的扩展名
		if dotIndex := strings.LastIndex(originalFileName, "."); dotIndex != -1 {
			extractDirName = originalFileName[:dotIndex]
		} else {
			extractDirName = originalFileName
		}
	}

	// 创建以压缩包文件名命名的解压目录
	extractedDir := filepath.Join(packageDir, extractDirName)
	if err := os.MkdirAll(extractedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create extracted directory: %v", err)
	}

	switch packageType {
	case model.PackageTypeTar:
		if err := m.extractTarFile(originalPath, extractedDir); err != nil {
			return "", err
		}
	case model.PackageTypeZip:
		if err := m.extractZipFile(originalPath, extractedDir); err != nil {
			return "", err
		}
	case model.PackageTypeTarGz:
		if err := m.extractTarGzFile(originalPath, extractedDir); err != nil {
			return "", err
		}
	case model.PackageTypeDxt:
		if err := m.extractDxtFile(originalPath, extractedDir); err != nil {
			return "", err
		}
	case model.PackageTypeMcpb:
		if err := m.extractMcpbFile(originalPath, extractedDir); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported package type: %s", packageType)
	}

	// 记录解压结束时间和耗时
	elapsed := time.Since(startTime)
	logger.Info("Package extraction completed",
		zap.String("extractedDir", extractedDir),
		zap.Duration("elapsed", elapsed))

	return extractedDir, nil
}

// extractTarFile 解压tar文件
func (m *CodePackageManager) extractTarFile(src, destPath string) error {
	logger.Info("Starting tar file extraction", zap.String("src", src), zap.String("destPath", destPath))

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %v", err)
	}
	defer file.Close()

	// 尝试作为gzip压缩的tar文件处理
	var tarReader *tar.Reader
	if strings.HasSuffix(strings.ToLower(src), ".gz") {
		logger.Info("Detected gzip compressed tar file")
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		logger.Info("Processing standard tar file")
		tarReader = tar.NewReader(file)
	}

	fileCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		target := filepath.Join(destPath, header.Name)

		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(target, destPath) {
			logger.Warn("Skipping file due to path traversal check", zap.String("path", header.Name))
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", target, err)
			}
			fileCount++
			if fileCount%100 == 0 {
				logger.Info("Extraction progress", zap.Int("files_processed", fileCount))
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", target, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			outFile.Close()
			fileCount++
			if fileCount%100 == 0 {
				logger.Info("Extraction progress", zap.Int("files_processed", fileCount))
			}
		}
	}

	logger.Info("Tar extraction completed", zap.Int("total_files", fileCount))
	return nil
}

// extractTarGzFile 解压tar.gz文件
func (m *CodePackageManager) extractTarGzFile(src, destPath string) error {
	logger.Info("Starting tar.gz file extraction", zap.String("src", src), zap.String("destPath", destPath))

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	fileCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		target := filepath.Join(destPath, header.Name)

		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(target, destPath) {
			logger.Warn("Skipping file due to path traversal check", zap.String("path", header.Name))
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", target, err)
			}
			fileCount++
			if fileCount%100 == 0 {
				logger.Info("Extraction progress", zap.Int("files_processed", fileCount))
			}
		case tar.TypeReg:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", target, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to copy file content: %v", err)
			}
			outFile.Close()
			fileCount++
			if fileCount%100 == 0 {
				logger.Info("Extraction progress", zap.Int("files_processed", fileCount))
			}
		}
	}

	logger.Info("Tar.gz extraction completed", zap.Int("total_files", fileCount))
	return nil
}

// extractDxtFile 解压dxt文件（使用zip格式）
func (m *CodePackageManager) extractDxtFile(src, destPath string) error {
	logger.Info("Starting dxt file extraction (using zip format)", zap.String("src", src), zap.String("destPath", destPath))

	// dxt 文件使用 zip 格式，直接调用 extractZipFile
	return m.extractZipFile(src, destPath)
}

// extractMcpbFile 解压mcpb文件（使用zip格式）
func (m *CodePackageManager) extractMcpbFile(src, destPath string) error {
	logger.Info("Starting mcpb file extraction (using zip format)", zap.String("src", src), zap.String("destPath", destPath))

	// mcpb 文件使用 zip 格式，直接调用 extractZipFile
	return m.extractZipFile(src, destPath)
}

// extractZipFile 解压zip文件
func (m *CodePackageManager) extractZipFile(src, destPath string) error {
	logger.Info("Starting zip file extraction", zap.String("src", src), zap.String("destPath", destPath))

	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer reader.Close()

	totalFiles := len(reader.File)
	logger.Info("Zip file opened", zap.Int("total_files", totalFiles))

	fileCount := 0
	for _, file := range reader.File {
		target := filepath.Join(destPath, file.Name)

		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(target, destPath) {
			logger.Warn("Skipping file due to path traversal check", zap.String("path", file.Name))
			continue
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", target, err)
			}
			fileCount++
			if fileCount%100 == 0 || fileCount == totalFiles {
				logger.Info("Extraction progress",
					zap.Int("files_processed", fileCount),
					zap.Int("total_files", totalFiles),
					zap.Float64("percent_complete", float64(fileCount)/float64(totalFiles)*100))
			}
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %v", err)
		}

		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %v", err)
		}

		outFile, err := os.Create(target)
		if err != nil {
			fileReader.Close()
			return fmt.Errorf("failed to create file %s: %v", target, err)
		}

		if _, err := io.Copy(outFile, fileReader); err != nil {
			fileReader.Close()
			outFile.Close()
			return fmt.Errorf("failed to copy file content: %v", err)
		}

		fileReader.Close()
		outFile.Close()

		fileCount++
		if fileCount%100 == 0 || fileCount == totalFiles {
			logger.Info("Extraction progress",
				zap.Int("files_processed", fileCount),
				zap.Int("total_files", totalFiles),
				zap.Float64("percent_complete", float64(fileCount)/float64(totalFiles)*100))
		}
	}

	logger.Info("Zip extraction completed", zap.Int("total_files", fileCount))
	return nil
}

// toRelativePath 将绝对路径转换为相对于配置根路径的相对路径
func (m *CodePackageManager) ToRelativePath(absolutePath string) (string, error) {
	// 获取配置根路径的绝对路径
	absRootPath, err := filepath.Abs(m.pathPrefix)
	if err != nil {
		return absolutePath, err
	}

	// 获取目标路径的绝对路径
	absTargetPath, err := filepath.Abs(absolutePath)
	if err != nil {
		return absolutePath, err
	}

	// 计算相对路径
	relPath, err := filepath.Rel(absRootPath, absTargetPath)
	if err != nil {
		return absolutePath, err
	}

	return relPath, nil
}

// toAbsolutePath 将相对路径转换为绝对路径
func (m *CodePackageManager) ToAbsolutePath(relativePath string) (string, error) {
	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(relativePath) {
		return relativePath, nil
	}

	// 获取配置根路径的绝对路径
	absRootPath, err := filepath.Abs(m.pathPrefix)
	if err != nil {
		return "", err
	}

	// 拼接得到绝对路径
	absolutePath := filepath.Join(absRootPath, relativePath)
	return absolutePath, nil
}

// DeletePackage removes a code package directory and all its contents
func (m *CodePackageManager) DeletePackage(packagePath string) error {
	// Convert relative path to absolute path
	absPackagePath, err := m.ToAbsolutePath(packagePath)
	if err != nil {
		return fmt.Errorf("failed to convert to absolute path: %v", err)
	}

	// Check if package directory exists
	if _, err := os.Stat(absPackagePath); os.IsNotExist(err) {
		logger.Warn("Package directory does not exist", zap.String("path", absPackagePath))
		return nil // Consider it as already deleted
	}

	// Remove the entire package directory
	if err := os.RemoveAll(absPackagePath); err != nil {
		logger.Error("Failed to remove package directory", 
			zap.String("path", absPackagePath), 
			zap.Error(err))
		return fmt.Errorf("failed to remove package directory: %v", err)
	}

	logger.Info("Package deleted successfully", zap.String("path", absPackagePath))
	return nil
}
