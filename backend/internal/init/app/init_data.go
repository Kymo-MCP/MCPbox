package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/kymo-mcp/mcpcan/pkg/logger"

	"go.uber.org/zap"
)

// copyInitData 拷贝初始化数据目录到目标目录
func (a *App) copyInitData(srcDir, dstDir string) error {
	// 检查源目录是否存在
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Source directory does not exist, skipping copy", zap.String("srcDir", srcDir))
			return nil // 源目录不存在时不报错，直接返回
		}
		return fmt.Errorf("failed to stat source directory %s: %w", srcDir, err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path %s is not a directory", srcDir)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dstDir, err)
	}

	// 递归拷贝目录内容
	return a.copyDirRecursive(srcDir, dstDir)
}

// copyDirRecursive 递归拷贝目录内容
func (a *App) copyDirRecursive(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// 创建目标子目录
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			// 递归拷贝子目录
			if err := a.copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 拷贝文件
			if err := a.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 拷贝单个文件
func (a *App) copyFile(srcPath, dstPath string) error {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", srcPath, err)
	}

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// 拷贝文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content from %s to %s: %w", srcPath, dstPath, err)
	}

	// 设置文件权限
	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions for %s: %w", dstPath, err)
	}

	logger.Debug("File copied successfully", zap.String("src", srcPath), zap.String("dst", dstPath))
	return nil
}
