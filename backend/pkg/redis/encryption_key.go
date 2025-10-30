package redis

import (
	"fmt"
	"time"
)

const (
	// EncryptionPrivateKeyPrefix 加密私钥前缀
	EncryptionPrivateKeyPrefix = "encryption_private_key"
	// EncryptionPublicKeyPrefix 加密公钥前缀
	EncryptionPublicKeyPrefix = "encryption_public_key"
	// DefaultEncryptionKeyTTL 默认加密密钥过期时间（todo: 5分钟）
	DefaultEncryptionKeyTTL = 24 * 60 * time.Minute
)

// SetEncryptionPrivateKey 保存加密私钥到Redis
func SetEncryptionPrivateKey(keyID, privateKey string) error {
	key := fmt.Sprintf("%s:%s", EncryptionPrivateKeyPrefix, keyID)
	return Set(key, privateKey, DefaultEncryptionKeyTTL)
}

// GetEncryptionPrivateKey 从Redis获取加密私钥
func GetEncryptionPrivateKey(keyID string) (string, error) {
	key := fmt.Sprintf("%s:%s", EncryptionPrivateKeyPrefix, keyID)
	value, err := Get(key)
	if err != nil {
		return "", fmt.Errorf("获取加密私钥失败: %v", err)
	}

	privateKey, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("私钥类型转换失败")
	}

	return privateKey, nil
}
