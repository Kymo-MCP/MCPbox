package model

import (
	"time"
)

// SysEncryptionKey 系统加密密钥模型
type SysEncryptionKey struct {
	KeyID      string     `gorm:"column:key_id;primaryKey;size:64;comment:密钥ID" json:"keyId"`
	PublicKey  string     `gorm:"column:public_key;type:text;comment:公钥(PEM格式)" json:"publicKey"`
	PrivateKey string     `gorm:"column:private_key;type:text;comment:私钥(PEM格式)" json:"privateKey"`
	Algorithm  string     `gorm:"column:algorithm;size:32;default:'RSA-2048';comment:加密算法" json:"algorithm"`
	KeySize    int        `gorm:"column:key_size;default:2048;comment:密钥长度" json:"keySize"`
	Status     string     `gorm:"column:status;size:16;default:'ACTIVE';comment:密钥状态:ACTIVE,EXPIRED,REVOKED" json:"status"`
	ClientID   *string    `gorm:"column:client_id;size:128;comment:客户端标识" json:"clientId"`
	IssuedAt   time.Time  `gorm:"column:issued_at;comment:签发时间" json:"issuedAt"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;comment:过期时间" json:"expiresAt"`
	CreateTime *time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime"`
	UpdateTime *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
}

// TableName 返回表名
func (SysEncryptionKey) TableName() string {
	return "sys_encryption_key"
}

// KeyStatus 密钥状态枚举
type KeyStatus string

const (
	KeyStatusActive  KeyStatus = "ACTIVE"  // 活跃
	KeyStatusExpired KeyStatus = "EXPIRED" // 过期
	KeyStatusRevoked KeyStatus = "REVOKED" // 撤销
)

// IsActive 检查密钥是否活跃
func (k *SysEncryptionKey) IsActive() bool {
	return k.Status == string(KeyStatusActive) && time.Now().Before(k.ExpiresAt)
}

// IsExpired 检查密钥是否过期
func (k *SysEncryptionKey) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}

// PrepareForCreate 创建前的准备工作
func (k *SysEncryptionKey) PrepareForCreate() error {
	now := time.Now()
	k.CreateTime = &now
	k.UpdateTime = &now
	return nil
}

// PrepareForUpdate 更新前的准备工作
func (k *SysEncryptionKey) PrepareForUpdate() error {
	now := time.Now()
	k.UpdateTime = &now
	return nil
}
