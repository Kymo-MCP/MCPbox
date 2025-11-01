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

	"github.com/kymo-mcp/mcpcan/api/market/storage"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
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
	// Get uploaded file
	imageFile, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "No image file provided")
		return
	}

	// Open file
	file, err := imageFile.Open()
	if err != nil {
		logger.Error("Failed to open image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to open image file")
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to read image file")
		return
	}

	// Validate image type
	if !utils.IsValidImageType(imageData) {
		logger.Error("Invalid image type")
		common.GinError(c, i18nresp.CodeInternalError, "Unsupported image type")
		return
	}

	// Validate file size (5MB limit)
	maxSize := int64(5 * 1024 * 1024)
	if int64(len(imageData)) > maxSize {
		logger.Error("Image file too large", zap.Int("size", len(imageData)))
		common.GinError(c, i18nresp.CodeInternalError, "Image file too large")
		return
	}

	// Generate file name
	ext := utils.GetImageFileExtension(imageData)
	if ext == "" {
		ext = "jpg"
	}
	fileName := fmt.Sprintf("%d.%s", time.Now().UnixNano(), ext)

	// Build storage path
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, strings.Trim(common.ImagesPath, "/"))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create storage directory")
		return
	}

	// Save file
	filePath := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		logger.Error("Failed to save image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to save image file")
		return
	}

	// Generate access URL
	imagePath := filepath.Join(common.StaticPrefix, common.ImagesPath, fileName)

	resp := &storage.UploadImageResponse{
		Path: imagePath,
		Size: int64(len(imageData)),
		Mime: fmt.Sprintf("image/%s", ext),
	}
	common.GinSuccess(c, resp)
}
