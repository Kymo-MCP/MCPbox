package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"qm-mcp-server/api/market/storage"
	"qm-mcp-server/internal/market/config"
	"qm-mcp-server/pkg/common"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/utils"
)

// StorageService provides storage-related operations
type StorageService struct {
	storage.UnimplementedStorageServiceServer
	ctx context.Context
}

// NewStorageService creates a new storage service instance
func NewStorageService(ctx context.Context) *StorageService {
	return &StorageService{
		ctx: ctx,
	}
}

// UploadImageHandler handles HTTP requests for image upload
func (s *StorageService) UploadImageHandler(c *gin.Context) {
	// 获取上传的文件
	imageFile, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "No image file provided")
		return
	}

	// 打开文件
	file, err := imageFile.Open()
	if err != nil {
		logger.Error("Failed to open image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to open image file")
		return
	}
	defer file.Close()

	// 读取文件内容
	imageData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to read image file")
		return
	}

	// 验证图片类型
	if !utils.IsValidImageType(imageData) {
		logger.Error("Invalid image type")
		common.GinError(c, i18nresp.CodeInternalError, "Unsupported image type")
		return
	}

	// 验证文件大小 (5MB限制)
	maxSize := int64(5 * 1024 * 1024)
	if int64(len(imageData)) > maxSize {
		logger.Error("Image file too large", zap.Int("size", len(imageData)))
		common.GinError(c, i18nresp.CodeInternalError, "Image file too large")
		return
	}

	// 生成文件名
	ext := utils.GetImageFileExtension(imageData)
	if ext == "" {
		ext = "jpg"
	}
	fileName := fmt.Sprintf("%d.%s", time.Now().UnixNano(), ext)

	// 构建存储路径
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, strings.Trim(common.ImagesPath, "/"))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create storage directory")
		return
	}

	// 保存文件
	filePath := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		logger.Error("Failed to save image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to save image file")
		return
	}

	// 生成访问URL
	imagePath := filepath.Join(common.StaticPrefix, common.ImagesPath, fileName)

	resp := &storage.UploadImageResponse{
		Path: imagePath,
		Size: int64(len(imageData)),
		Mime: fmt.Sprintf("image/%s", ext),
	}
	common.GinSuccess(c, resp)
}
