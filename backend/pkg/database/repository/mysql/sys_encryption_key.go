package mysql

import (
	"context"
	"fmt"
	"time"

	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/logger"

	"gorm.io/gorm"
)

var SysEncryptionKeyRepo *SysEncryptionKeyRepository

func init() {
	RegisterInit(func(db *gorm.DB) {
		repo := NewSysEncryptionKeyRepository(db)
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize sys_encryption_key table: %v", err))
		}
	})
}

// SysEncryptionKeyRepository 加密密钥仓库
type SysEncryptionKeyRepository struct{}

// NewSysEncryptionKeyRepository 创建加密密钥仓库
func NewSysEncryptionKeyRepository(db *gorm.DB) *SysEncryptionKeyRepository {
	SysEncryptionKeyRepo = &SysEncryptionKeyRepository{}
	return SysEncryptionKeyRepo
}

// getDB 获取数据库连接
func (r *SysEncryptionKeyRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.SysEncryptionKey{})
}

// GetByKeyID 根据密钥ID获取密钥
func (r *SysEncryptionKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*model.SysEncryptionKey, error) {
	var key model.SysEncryptionKey
	if err := r.getDB().WithContext(ctx).Where("key_id = ?", keyID).First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("密钥不存在: %s", keyID)
		}
		return nil, fmt.Errorf("获取密钥失败: %v", err)
	}
	return &key, nil
}

// GetActiveKeyForClient 获取客户端的活跃密钥
func (r *SysEncryptionKeyRepository) GetActiveKeyForClient(ctx context.Context, clientID string) (*model.SysEncryptionKey, error) {
	var key model.SysEncryptionKey
	if err := r.getDB().WithContext(ctx).
		Where("client_id = ? AND status = ? AND expires_at > ?",
			clientID, string(model.KeyStatusActive), time.Now()).
		Order("issued_at DESC").
		First(&key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 没有找到活跃密钥，返回nil而不是错误
		}
		return nil, fmt.Errorf("获取活跃密钥失败: %v", err)
	}
	return &key, nil
}

// Create 创建新密钥
func (r *SysEncryptionKeyRepository) Create(ctx context.Context, key *model.SysEncryptionKey) error {
	if err := key.PrepareForCreate(); err != nil {
		return fmt.Errorf("准备创建密钥失败: %v", err)
	}

	if err := r.getDB().WithContext(ctx).Create(key).Error; err != nil {
		return fmt.Errorf("创建密钥失败: %v", err)
	}

	return nil
}

// UpdateStatus 更新密钥状态
func (r *SysEncryptionKeyRepository) UpdateStatus(ctx context.Context, keyID string, status model.KeyStatus) error {
	if err := r.getDB().WithContext(ctx).
		Where("key_id = ?", keyID).
		Updates(map[string]interface{}{
			"status":      string(status),
			"update_time": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("更新密钥状态失败: %v", err)
	}

	return nil
}

// DeleteExpiredKeys 删除过期密钥
func (r *SysEncryptionKeyRepository) DeleteExpiredKeys(ctx context.Context, beforeTime time.Time) (int64, error) {
	result := r.getDB().WithContext(ctx).
		Where("expires_at < ? AND status != ?", beforeTime, string(model.KeyStatusActive)).
		Delete(&model.SysEncryptionKey{})

	if result.Error != nil {
		return 0, fmt.Errorf("删除过期密钥失败: %v", result.Error)
	}

	return result.RowsAffected, nil
}

// ListByClientID 根据客户端ID列出密钥
func (r *SysEncryptionKeyRepository) ListByClientID(ctx context.Context, clientID string, limit int) ([]*model.SysEncryptionKey, error) {
	var keys []*model.SysEncryptionKey
	query := r.getDB().WithContext(ctx).Where("client_id = ?", clientID).Order("issued_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("获取客户端密钥列表失败: %v", err)
	}

	return keys, nil
}

// Count 统计密钥数量
func (r *SysEncryptionKeyRepository) Count(ctx context.Context, status string) (int64, error) {
	var count int64
	query := r.getDB().WithContext(ctx)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计密钥数量失败: %v", err)
	}

	return count, nil
}

// HealthCheck 健康检查
func (r *SysEncryptionKeyRepository) HealthCheck(ctx context.Context) error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	var count int64
	err := r.getDB().WithContext(ctx).Model(&model.SysEncryptionKey{}).Count(&count).Error
	return err
}

// InitTable 初始化表
func (r *SysEncryptionKeyRepository) InitTable() error {
	if r.getDB() == nil {
		return fmt.Errorf("database connection is nil")
	}

	db := r.getDB()
	if db == nil {
		return fmt.Errorf("failed to get database connection")
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&model.SysEncryptionKey{}); err != nil {
		return fmt.Errorf("failed to migrate sys_encryption_key table: %v", err)
	}

	// 检查并创建索引
	indexes := []struct {
		name   string
		column string
		unique bool
	}{
		{"idx_client_id", "client_id", false},
		{"idx_status", "status", false},
		{"idx_expires_at", "expires_at", false},
		{"idx_issued_at", "issued_at", false},
		{"uniq_key_id", "key_id", true},
	}

	for _, idx := range indexes {
		var count int64
		db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
			(&model.SysEncryptionKey{}).TableName(), idx.name).Scan(&count)

		if count == 0 {
			var sql string
			if idx.unique {
				sql = fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", idx.name, (&model.SysEncryptionKey{}).TableName(), idx.column)
			} else {
				sql = fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx.name, (&model.SysEncryptionKey{}).TableName(), idx.column)
			}

			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create index %s: %v", idx.name, err)
			}
		}
	}

	logger.Info("Successfully initialized sys_encryption_key table")
	return nil
}
