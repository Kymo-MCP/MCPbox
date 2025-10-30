package service

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qm-mcp-server/api/market/code"
	"qm-mcp-server/internal/market/config"
	"qm-mcp-server/pkg/codepackage"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	i18nresp "qm-mcp-server/pkg/i18n"
	"qm-mcp-server/pkg/logger"
	"qm-mcp-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CodeService provides code package management services
type CodeService struct {
	codePackageRepo *mysql.McpCodePackageRepository
	instanceRepo    *mysql.McpInstanceRepository
	templateRepo    *mysql.McpTemplateRepository
	packageManager  *codepackage.CodePackageManager
}

// NewCodeService creates a new CodeService instance
func NewCodeService() *CodeService {
	return &CodeService{
		codePackageRepo: mysql.McpCodePackageRepo,
		instanceRepo:    mysql.McpInstanceRepo,
		templateRepo:    mysql.McpTemplateRepo,
		packageManager:  codepackage.NewCodePackageManager(&config.GlobalConfig.Code, config.GlobalConfig.Storage.CodePath),
	}
}

// UploadPackage uploads a code package
func (s *CodeService) UploadPackage(c *gin.Context) {
	// 记录上传开始时间
	startTime := time.Now()
	logger.Info("Starting code package upload request",
		zap.String("client_ip", c.ClientIP()),
		zap.String("request_id", c.GetString("RequestID")),
		zap.String("content_type", c.ContentType()))

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("Failed to get uploaded file",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", c.GetString("RequestID")))
		common.GinError(c, i18nresp.CodeInternalError, "failed to get uploaded file")
		return
	}
	defer file.Close()

	// 记录上传文件的详细信息
	logger.Info("File received for upload",
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
		zap.Int("configured_max_size_mb", config.GlobalConfig.Code.Upload.MaxFileSize),
		zap.Float64("size_mb", float64(header.Size)/(1024*1024)),
		zap.String("content_type", header.Header.Get("Content-Type")))

	// 使用代码包管理器处理上传和解压
	packageInfo, err := s.packageManager.UploadAndExtractPackage(file, header)
	if err != nil {
		logger.Error("Failed to upload and extract package", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	ctx := context.Background()

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

	if err := s.codePackageRepo.Create(ctx, codePackage); err != nil {
		logger.Error("Failed to save package to database", zap.Error(err))
		// 清理创建的目录
		os.RemoveAll(packageInfo.PackagePath)
		common.GinError(c, i18nresp.CodeInternalError, "failed to save package information")
		return
	}

	// 计算总体耗时
	totalElapsed := time.Since(startTime)
	logger.Info("Package uploaded successfully",
		zap.String("packageId", packageInfo.PackageID),
		zap.String("filename", packageInfo.OriginalName),
		zap.String("packagePath", packageInfo.PackagePath),
		zap.String("extractedPath", packageInfo.ExtractedPath),
		zap.Duration("total_elapsed", totalElapsed),
		zap.Float64("total_elapsed_seconds", totalElapsed.Seconds()))

	common.GinSuccess(c, &code.UploadPackageResponse{
		PackageId:   packageInfo.PackageID,
		PackagePath: packageInfo.ExtractedPath, // 返回相对路径
	})
}

// GetCodeTree retrieves the code tree structure
func (s *CodeService) GetCodeTree(c *gin.Context) {
	var req code.GetCodeTreeRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	ctx := context.Background()

	// 查找代码包
	codePackage, err := s.codePackageRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to find code package", zap.String("packageId", req.PackageId), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "code package not found")
		return
	}

	// 使用解压后的路径构建文件结构
	extractedPath := codePackage.ExtractedPath
	if extractedPath == "" {
		// 兼容旧数据，如果没有解压路径，使用包路径
		extractedPath = codePackage.PackagePath
	}

	// 将相对路径转换为绝对路径
	absExtractedPath, err := s.packageManager.ToAbsolutePath(extractedPath)
	if err != nil {
		logger.Error("Failed to convert to absolute path", zap.String("relativePath", extractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid package path")
		return
	}

	fileStructure, err := s.buildFileTree(absExtractedPath)
	if err != nil {
		logger.Error("Failed to build file tree", zap.String("path", absExtractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to build file structure")
		return
	}

	common.GinSuccess(c, &code.GetCodeTreeResponse{
		FileStructure: fileStructure,
	})
}

// GetCodeFile retrieves a specific code file
func (s *CodeService) GetCodeFile(c *gin.Context) {
	var req code.GetCodeFileRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	ctx := context.Background()

	// 查找代码包
	codePackage, err := s.codePackageRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to find code package", zap.String("instanceId", req.PackageId), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "code package not found")
		return
	}

	// 使用解压后的路径
	extractedPath := codePackage.ExtractedPath
	if extractedPath == "" {
		// 兼容旧数据，如果没有解压路径，使用包路径
		extractedPath = codePackage.PackagePath
	}

	// 将相对路径转换为绝对路径
	absExtractedPath, err := s.packageManager.ToAbsolutePath(extractedPath)
	if err != nil {
		logger.Error("Failed to convert to absolute path", zap.String("relativePath", extractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid package path")
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(absExtractedPath, req.FilePath)

	// 安全检查：确保文件路径在包目录内
	absPackagePath, err := filepath.Abs(absExtractedPath)
	if err != nil {
		logger.Error("Failed to get absolute package path", zap.String("path", absExtractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid package path")
		return
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		logger.Error("Failed to get absolute file path", zap.String("path", fullPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid file path")
		return
	}

	if !strings.HasPrefix(absFilePath, absPackagePath) {
		logger.Warn("Attempted to access file outside package directory",
			zap.String("filePath", absFilePath),
			zap.String("packagePath", absPackagePath))
		common.GinError(c, i18nresp.CodeInternalError, "file path not allowed")
		return
	}

	// 检查文件是否存在
	if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
		common.GinError(c, i18nresp.CodeInternalError, "file not found")
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Error("Failed to read file", zap.String("path", fullPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to read file")
		return
	}

	common.GinSuccess(c, &code.GetCodeFileResponse{
		Content: string(content),
	})
}

// EditCodeFile edits a code file
func (s *CodeService) EditCodeFile(c *gin.Context) {
	var req code.EditCodeFileRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	ctx := context.Background()

	// 查找代码包
	codePackage, err := s.codePackageRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to find code package", zap.String("instanceId", req.PackageId), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "code package not found")
		return
	}

	// 使用解压后的路径
	extractedPath := codePackage.ExtractedPath
	if extractedPath == "" {
		// 兼容旧数据，如果没有解压路径，使用包路径
		extractedPath = codePackage.PackagePath
	}

	// 将相对路径转换为绝对路径
	absExtractedPath, err := s.packageManager.ToAbsolutePath(extractedPath)
	if err != nil {
		logger.Error("Failed to convert to absolute path", zap.String("relativePath", extractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid package path")
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(absExtractedPath, req.FilePath)

	// 安全检查：确保文件路径在包目录内
	absPackagePath, err := filepath.Abs(absExtractedPath)
	if err != nil {
		logger.Error("Failed to get absolute package path", zap.String("path", absExtractedPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid package path")
		return
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		logger.Error("Failed to get absolute file path", zap.String("path", fullPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid file path")
		return
	}

	if !strings.HasPrefix(absFilePath, absPackagePath) {
		logger.Warn("Attempted to access file outside package directory",
			zap.String("filePath", absFilePath),
			zap.String("packagePath", absPackagePath))
		common.GinError(c, i18nresp.CodeInternalError, "file path not allowed")
		return
	}

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Failed to create directory", zap.String("dir", dir), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to create directory")
		return
	}

	// 写入文件内容
	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		logger.Error("Failed to write file", zap.String("path", fullPath), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to write file")
		return
	}
	common.GinSuccess(c, &code.EditCodeFileResponse{
		Success: true,
		Message: "file edited successfully",
	})
}

// buildFileTree 构建文件树结构
func (s *CodeService) buildFileTree(rootPath string) (*code.FileNode, error) {
	return utils.BuildFileTreeRecursive(rootPath, rootPath, "")
}

// DownloadPackage handles package download requests
func (s *CodeService) DownloadPackage(c *gin.Context) {
	req := &code.DownloadPackageRequest{}
	if err := common.BindAndValidateQuery(c, req); err != nil {
		return
	}
	packageID := req.PackageId

	// Parameter validation
	if packageID == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "package ID and file name are required",
		})
		return
	}

	// Find code package
	codePackage, err := s.codePackageRepo.FindByPackageID(c, packageID)
	if err != nil {
		logger.Error("Failed to find code package", zap.String("packageId", packageID), zap.Error(err))
		c.JSON(404, gin.H{
			"code":    404,
			"message": "code package not found",
		})
		return
	}

	// Build file path
	absFilePath := filepath.Join(config.GlobalConfig.Storage.CodePath, codePackage.PackagePath, codePackage.OriginalName)

	// Check if file exists
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		logger.Error("File not found", zap.String("filePath", absFilePath), zap.Error(err))
		c.JSON(404, gin.H{
			"code":    404,
			"message": "file not found",
		})
		return
	}

	// Determine the actual filename to use for download
	downloadFileName := codePackage.OriginalName

	// Set proper Content-Type based on file extension
	ext := filepath.Ext(downloadFileName)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		// Default to application/octet-stream for unknown file types
		contentType = "application/octet-stream"
	}

	// Set response headers for proper file download
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadFileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Serve the file
	c.File(absFilePath)
}

// GenerateDownloadZip 处理下载请求的通用逻辑
// GetCodePackageList 获取代码包列表
func (s *CodeService) GetCodePackageList(c *gin.Context) {
	var req code.CodePackageListRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid request parameters")
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filters := map[string]interface{}{}
	if req.Keyword != "" {
		filters["keyword"] = req.Keyword
	}
	if len(req.PackageType) > 0 {
		// 转换为模型类型
		var packageTypes []model.PackageType
		for _, t := range req.PackageType {
			modelType, _ := common.ConvertToModelPackageType(t)
			packageTypes = append(packageTypes, modelType)
		}
		if len(packageTypes) > 0 {
			filters["packageType"] = packageTypes
		}
	}
	// 查询代码包列表
	packages, total, err := s.codePackageRepo.FindWithPagination(c.Request.Context(), req.Page, req.PageSize, filters)
	if err != nil {
		logger.Error("Failed to query code packages", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to query code packages")
		return
	}

	// 转换为响应格式
	var packageList []*code.CodePackageInfo
	for _, pkg := range packages {
		packageInfo := &code.CodePackageInfo{
			Id:        pkg.PackageID,
			Name:      pkg.OriginalName,
			Path:      pkg.PackagePath,
			Size:      pkg.FileSize,
			Type:      convertPackageType(pkg.PackageType),
			CreatedAt: pkg.CreatedAt.String(),
			UpdatedAt: pkg.UpdatedAt.String(),
		}
		packageList = append(packageList, packageInfo)
	}

	response := &code.CodePackageListResponse{
		List:     packageList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	common.GinSuccess(c, response)
}

// convertPackageType 转换包类型
func convertPackageType(modelType model.PackageType) code.PackageType {
	switch modelType {
	case model.PackageTypeTar:
		return code.PackageType_PackageTypeTar
	case model.PackageTypeZip:
		return code.PackageType_PackageTypeZip
	default:
		return code.PackageType_PackageTypeUnspecified
	}
}

// DeleteCodePackage deletes a code package and its associated files
func (s *CodeService) DeleteCodePackage(c *gin.Context) {
	var req code.DeleteCodePackageRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "invalid request parameters")
		return
	}

	// Validate package ID
	if req.PackageId == "" {
		logger.Warn("Empty package ID provided for deletion")
		common.GinError(c, i18nresp.CodeBadRequest, "package ID is required")
		return
	}

	ctx := context.Background()

	// Find the code package
	codePackage, err := s.codePackageRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to find code package", zap.String("packageId", req.PackageId), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "code package not found")
		return
	}

	if codePackage == nil {
		logger.Warn("Code package not found", zap.String("packageId", req.PackageId))
		common.GinError(c, i18nresp.CodeInternalError, "code package not found")
		return
	}

	// Check if package is being used by any instances
	instances, err := s.instanceRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to check package usage", zap.String("packageId", req.PackageId), zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to check package usage")
		return
	}

	// Check if any templates are using this package
	templates, err := s.templateRepo.FindByPackageID(ctx, req.PackageId)
	if err != nil {
		logger.Error("Failed to check template usage", zap.String("packageId", req.PackageId), zap.Error(err))

		common.GinError(c, i18nresp.CodeInternalError, "failed to check template usage")
		return
	}
	if len(templates) > 0 {
		logger.Warn("Cannot delete package: templates are using it",
			zap.String("packageId", req.PackageId),
			zap.Int("templateCount", len(templates)))
		names := []string{}
		for _, template := range templates {
			names = append(names, template.Name)
		}
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("cannot delete package that is being used by templates %v", strings.Join(names, ", ")))
		return
	}

	if len(instances) > 0 {
		logger.Warn("Cannot delete package in use",
			zap.String("packageId", req.PackageId),
			zap.Int("instanceCount", len(instances)))
		names := []string{}
		for _, instance := range instances {
			names = append(names, instance.InstanceName)
		}
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("cannot delete package that is being used by instances %v", strings.Join(names, ", ")))
		return
	}

	// Delete physical files using package manager
	if codePackage.PackagePath != "" {
		if err := s.packageManager.DeletePackage(codePackage.PackagePath); err != nil {
			logger.Error("Failed to delete package files",
				zap.String("packageId", req.PackageId),
				zap.String("packagePath", codePackage.PackagePath),
				zap.Error(err))
			// Continue with database deletion even if file deletion fails
			logger.Warn("Continuing with database deletion despite file deletion failure")
		}
	}

	// Delete database record
	if err := s.codePackageRepo.DeleteByPackageID(ctx, req.PackageId); err != nil {
		logger.Error("Failed to delete code package from database",
			zap.String("packageId", req.PackageId),
			zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to delete package record")
		return
	}

	logger.Info("Code package deleted successfully", zap.String("packageId", req.PackageId))

	// Return success response
	response := &code.DeleteCodePackageResponse{
		Success: true,
		Message: "Code package deleted successfully",
	}

	common.GinSuccess(c, response)
}
