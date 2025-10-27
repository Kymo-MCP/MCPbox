package app

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"qm-mcp-server/pkg/codepackage"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	"qm-mcp-server/pkg/logger"
)

// initCodePackage 初始化代码包
func (a *App) initCodePackage(ctx context.Context) error {
	logger.Info("Starting code package initialization")

	// 检查源目录是否存在
	sourceDir := "./init-data/code-package"
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		logger.Info("Code package source directory does not exist, skipping initialization",
			zap.String("sourceDir", sourceDir))
		return nil
	}

	// 遍历源目录下的所有文件
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	if len(entries) == 0 {
		logger.Info("No code packages found in source directory", zap.String("sourceDir", sourceDir))
		return nil
	}

	// 创建代码包管理器 - 使用默认配置
	codeConfig := &common.CodeConfig{
		Upload: common.UploadConfig{
			MaxFileSize:       100, // 100MB
			AllowedExtensions: []string{".zip", ".tar", ".tar.gz", ".dxt", ".mcpb"},
		},
	}
	packageManager := codepackage.NewCodePackageManager(codeConfig, a.config.Storage.CodePath)

	processedCount := 0
	skippedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		filePath := filepath.Join(sourceDir, fileName)

		// 检查文件扩展名，只处理支持的压缩包格式
		ext := strings.ToLower(filepath.Ext(fileName))
		extList := []string{model.PackageTypeTar.String(), model.PackageTypeZip.String(), model.PackageTypeTarGz.String(), model.PackageTypeDxt.String(), model.PackageTypeMcpb.String()}
		isSupported := false
		for _, e := range extList {
			supportExt := fmt.Sprintf(".%s", e)
			if supportExt == ext {
				isSupported = true
				break
			}
		}
		if !isSupported {
			logger.Debug("Skipping unsupported file type",
				zap.String("fileName", fileName),
				zap.String("ext", ext))
			skippedCount++
			continue
		}

		// 检查数据库中是否已存在同名的代码包
		existingPackages, err := mysql.McpCodePackageRepo.FindAll(ctx)
		if err != nil {
			logger.Error("Failed to query existing packages", zap.Error(err))
			skippedCount++
			continue
		}

		// 检查是否已存在同名包
		exists := false
		for _, pkg := range existingPackages {
			if pkg.OriginalName == fileName {
				exists = true
				logger.Debug("Code package already exists, skipping", zap.String("fileName", fileName))
				skippedCount++
				a.codePackageList = append(a.codePackageList, pkg)
				break
			}
		}

		if exists {
			continue
		}

		// 处理代码包文件
		if err := a.processCodePackageFile(ctx, filePath, fileName, packageManager); err != nil {
			logger.Error("Failed to process code package file",
				zap.String("fileName", fileName),
				zap.Error(err))
			continue
		}

		processedCount++
		logger.Info("Successfully processed code package", zap.String("fileName", fileName))
	}

	logger.Info("Code package initialization completed",
		zap.Int("processed", processedCount),
		zap.Int("skipped", skippedCount))

	return nil
}

// processCodePackageFile 处理单个代码包文件
func (a *App) processCodePackageFile(ctx context.Context, filePath, fileName string,
	packageManager *codepackage.CodePackageManager) error {

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// 创建模拟的 multipart.FileHeader
	header := &multipart.FileHeader{
		Filename: fileName,
		Size:     fileInfo.Size(),
		Header:   make(map[string][]string),
	}

	// 创建文件读取器包装
	fileReader := &fileReaderWrapper{file: file}

	// 使用代码包管理器处理上传和解压
	packageInfo, err := packageManager.UploadAndExtractPackage(fileReader, header)
	if err != nil {
		return fmt.Errorf("failed to upload and extract package: %w", err)
	}

	// 保存到数据库
	codePackage := &model.McpCodePackage{
		PackageID:     packageInfo.PackageID,
		PackageType:   packageInfo.PackageType,
		PackagePath:   packageInfo.PackagePath,
		OriginalPath:  packageInfo.OriginalPath,
		ExtractedPath: packageInfo.ExtractedPath,
		OriginalName:  packageInfo.OriginalName,
		FileSize:      packageInfo.FileSize,
	}

	if err := mysql.McpCodePackageRepo.Create(ctx, codePackage); err != nil {
		// 清理创建的目录
		if absPath, convErr := packageManager.ToAbsolutePath(packageInfo.PackagePath); convErr == nil {
			os.RemoveAll(absPath)
		}
		return fmt.Errorf("failed to save package to database: %w", err)
	}
	a.codePackageList = append(a.codePackageList, codePackage)
	return nil
}

// fileReaderWrapper 包装 os.File 以实现 multipart.File 接口
type fileReaderWrapper struct {
	file *os.File
}

func (f *fileReaderWrapper) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

func (f *fileReaderWrapper) ReadAt(p []byte, off int64) (n int, err error) {
	return f.file.ReadAt(p, off)
}

func (f *fileReaderWrapper) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

func (f *fileReaderWrapper) Close() error {
	return f.file.Close()
}
