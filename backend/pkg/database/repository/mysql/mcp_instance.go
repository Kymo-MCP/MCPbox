package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"

	"gorm.io/gorm"
)

var McpInstanceRepo *McpInstanceRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewMcpInstanceRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_instance table: %v", err))
		}
	})
}

// McpInstanceRepository 封装 mcp_instance 表的增删改查操作
type McpInstanceRepository struct{}

// NewMcpInstanceRepository 创建 McpInstanceRepository 实例
func NewMcpInstanceRepository() *McpInstanceRepository {
	McpInstanceRepo = &McpInstanceRepository{}
	return McpInstanceRepo
}

// getDB 获取数据库连接
func (r *McpInstanceRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpInstance{})
}

// FindByInstanceID 通过 instanceId 查询数据
func (r *McpInstanceRepository) FindByInstanceIDAndAccessType(ctx context.Context, instanceID string, accessType model.AccessType) (*model.McpInstance, error) {
	var instance model.McpInstance
	if err := r.getDB().WithContext(ctx).Where("instance_id = ? AND access_type = ?", instanceID, accessType).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("instance not found: %s", instanceID)
		}
		return nil, fmt.Errorf("failed to find instance: %v", err)
	}
	return &instance, nil
}

// Create 创建实例
func (r *McpInstanceRepository) Create(ctx context.Context, instance *model.McpInstance) error {
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()
	return r.getDB().WithContext(ctx).Create(instance).Error
}

// Update 更新实例
func (r *McpInstanceRepository) Update(ctx context.Context, instance *model.McpInstance) error {
	instance.UpdatedAt = time.Now()
	return r.getDB().WithContext(ctx).Where("instance_id = ?", instance.InstanceID).Save(instance).Error
}

// Delete 删除实例
func (r *McpInstanceRepository) Delete(ctx context.Context, instanceId string) error {
	return r.getDB().WithContext(ctx).Where("instance_id = ?", instanceId).Delete(&model.McpInstance{}).Error
}

// FindByID 根据ID查找实例
func (r *McpInstanceRepository) FindByID(ctx context.Context, id uint) (*model.McpInstance, error) {
	var instance model.McpInstance
	err := r.getDB().WithContext(ctx).First(&instance, id).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// FindAll 查找所有实例
func (r *McpInstanceRepository) FindAll(ctx context.Context) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	err := r.getDB().WithContext(ctx).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// FindByStatus 根据状态查找实例
func (r *McpInstanceRepository) FindByStatus(ctx context.Context, status model.InstanceStatus) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	err := r.getDB().WithContext(ctx).Where("status = ?", status).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// FindByInstanceID 根据实例ID查找例（不限制访问类型）
func (r *McpInstanceRepository) FindByInstanceID(ctx context.Context, instanceID string) (*model.McpInstance, error) {
	var instance model.McpInstance
	if err := r.getDB().WithContext(ctx).Where("instance_id = ?", instanceID).First(&instance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("instance not found: %s", instanceID)
		}
		return nil, fmt.Errorf("failed to find instance: %v", err)
	}
	return &instance, nil
}

// InitTable 初始化表结构
func (r *McpInstanceRepository) InitTable() error {
	// 创建表
	mod := &model.McpInstance{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	// 检查索引是否存在
	var count int64
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_instance_instance_id'", (&model.McpInstance{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建索引
		sql2 := fmt.Sprintf("CREATE INDEX idx_mcp_instance_instance_id ON %v(instance_id)", (&model.McpInstance{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	// name 唯一索引是否存在，不存在则创建
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_mcp_instance_name'", (&model.McpInstance{}).TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// 创建索引
		sql2 := fmt.Sprintf("CREATE UNIQUE INDEX idx_mcp_instance_name ON %v(instance_name)", (&model.McpInstance{}).TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create unique index: %v", err)
		}
	}
	return nil
}

// FindByContainerStatus 根据容器状态查找实例
func (r *McpInstanceRepository) FindByContainerStatus(ctx context.Context, statuses []model.ContainerStatus) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	err := r.getDB().WithContext(ctx).Where("container_status IN ?", statuses).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// FindHostingInstances 查询服务中托管实例：条件 => 查询所有 status : inactive, container_status : 创建或运行中， access_type = hosting 部署模式的实例
func (r *McpInstanceRepository) FindHostingInstances(ctx context.Context) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	instanceStatus := model.InstanceStatusActive
	containerStatus := []model.ContainerStatus{model.ContainerStatusPending, model.ContainerStatusRunning, model.ContainerStatusRunningUnready}
	err := r.getDB().WithContext(ctx).Model(&model.McpInstance{}).Where("status = ? AND container_status IN ?", instanceStatus, containerStatus).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// FindWithPagination 分页查询实例
func (r *McpInstanceRepository) FindWithPagination(ctx context.Context, page, pageSize int32, filters map[string]interface{}, sortBy, sortOrder string) ([]*model.McpInstance, int64, error) {
	var instances []*model.McpInstance
	var total int64

	// 构建查询条件
	query := r.getDB().WithContext(ctx).Model(&model.McpInstance{})

	// 应用筛选条件
	for key, value := range filters {
		switch key {
		case "instanceName":
			if instanceName, ok := value.(string); ok && instanceName != "" {
				query = query.Where("instance_name LIKE ? OR instance_id LIKE ?", "%"+instanceName+"%", "%"+instanceName+"%")
			}
		case "environmentId":
			if envId, ok := value.(int32); ok && envId > 0 {
				query = query.Where("environment_id = ?", envId)
			}
		case "deployMode":
			if deployMode, ok := value.(uint32); ok {
				query = query.Where("deploy_mode = ?", deployMode)
			}
		case "status":
			if status, ok := value.(string); ok && status != "" {
				query = query.Where("status = ?", status)
			}
		case "accessType":
			if accessType, ok := value.(model.AccessType); ok && len(accessType.String()) > 0 {
				query = query.Where("access_type = ?", accessType)
			}
		case "containerStatus":
			if containerStatus, ok := value.(string); ok && containerStatus != "" {
				query = query.Where("container_status = ?", containerStatus)
			}
		case "mcpProtocol":
			if mcpProtocol, ok := value.(model.McpProtocol); ok {
				query = query.Where("mcp_protocol = ?", mcpProtocol)
			}
		}
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	if sortBy != "" {
		order := "ASC"
		if sortOrder == "desc" {
			order = "DESC"
		}
		switch sortBy {
		case "createdAt":
			query = query.Order(fmt.Sprintf("created_at %s", order))
		case "updatedAt":
			query = query.Order(fmt.Sprintf("updated_at %s", order))
		case "instanceName":
			query = query.Order(fmt.Sprintf("instance_name %s", order))
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// 应用分页
	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&instances).Error; err != nil {
		return nil, 0, err
	}

	return instances, total, nil
}

// FindByPackageID finds instances by package ID
func (r *McpInstanceRepository) FindByPackageID(ctx context.Context, packageID string) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	err := r.getDB().WithContext(ctx).Where("package_id = ?", packageID).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

// FindByName 根据实例名称查询实例
func (r *McpInstanceRepository) FindByName(ctx context.Context, name string) (*model.McpInstance, error) {
	var instance model.McpInstance
	err := r.getDB().WithContext(ctx).Where("instance_name = ?", name).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// FindByEnvironmentID finds instances by environment ID
func (r *McpInstanceRepository) FindByEnvironmentID(ctx context.Context, environmentID uint) ([]*model.McpInstance, error) {
	var instances []*model.McpInstance
	err := r.getDB().WithContext(ctx).Where("environment_id = ?", environmentID).Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}
