package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"qm-mcp-server/pkg/redis"
	"time"

	"github.com/google/uuid"
)

const AlgorithmRSA2048 = "RSA-2048"

const KeySize = 2048

// RSAKeyPair RSA密钥对
type RSAKeyPair struct {
	KeyID      string
	PublicKey  string // PEM格式
	PrivateKey string // PEM格式
	KeySize    int
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// GenerateRSAKeyPair 生成RSA密钥对
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("生成RSA私钥失败: %v", err)
	}

	// 编码私钥为PEM格式
	privateKeyPEM, err := encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("编码私钥失败: %v", err)
	}

	// 编码公钥为PEM格式
	publicKeyPEM, err := encodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("编码公钥失败: %v", err)
	}

	now := time.Now()
	keyPair := &RSAKeyPair{
		KeyID:      generateKeyID(),
		PublicKey:  publicKeyPEM,
		PrivateKey: privateKeyPEM,
		KeySize:    KeySize,
		IssuedAt:   now,
		ExpiresAt:  now.Add(redis.DefaultEncryptionKeyTTL),
	}

	return keyPair, nil
}

// RSAEncrypt 使用RSA公钥加密数据
func RSAEncrypt(data []byte, publicKeyPEM string) (string, error) {
	// 解析公钥
	publicKey, err := parsePublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", fmt.Errorf("解析公钥失败: %v", err)
	}

	// 使用OAEP填充进行加密
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA加密失败: %v", err)
	}

	// 返回Base64编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt 使用RSA私钥解密数据
func RSADecrypt(ciphertext string, privateKeyPEM string) ([]byte, error) {
	// 解析私钥
	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 使用OAEP填充进行解密
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(ciphertext), nil)
	if err != nil {
		return nil, fmt.Errorf("RSA解密失败: %v", err)
	}

	return plaintext, nil
}

// EncryptPassword 加密密码（包含时间戳验证）
func EncryptPassword(password string, timestamp int64, publicKeyPEM string) (string, error) {
	// 构造待加密的数据：密码|时间戳
	data := fmt.Sprintf("%s|%d", password, timestamp)
	return RSAEncrypt([]byte(data), publicKeyPEM)
}

// DecryptPassword 解密密码（包含时间戳验证）
func DecryptPassword(encryptedPassword string, privateKeyPEM string, maxAge time.Duration) (string, int64, error) {
	// 解密数据
	plaintext, err := RSADecrypt(encryptedPassword, privateKeyPEM)
	if err != nil {
		return "", 0, err
	}

	// 解析密码和时间戳
	data := string(plaintext)
	var password string
	var timestamp int64
	n, err := fmt.Sscanf(data, "%s|%d", &password, &timestamp)
	if err != nil || n != 2 {
		return "", 0, fmt.Errorf("解析密码数据失败: %v", err)
	}

	// 验证时间戳
	if maxAge > 0 {
		now := time.Now().Unix()
		if now-timestamp > int64(maxAge.Seconds()) {
			return "", 0, fmt.Errorf("密码已过期，时间戳: %d, 当前: %d", timestamp, now)
		}
	}

	return password, timestamp, nil
}

// generateKeyID 生成密钥ID
func generateKeyID() string {
	return fmt.Sprintf("key_%s_%d", uuid.New().String()[:8], time.Now().Unix())
}

// encodePrivateKeyToPEM 将私钥编码为PEM格式
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// encodePublicKeyToPEM 将公钥编码为PEM格式
func encodePublicKeyToPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// parsePrivateKeyFromPEM 从PEM格式解析私钥
func parsePrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式私钥")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA私钥")
	}

	return rsaPrivateKey, nil
}

// parsePublicKeyFromPEM 从PEM格式解析公钥
func parsePublicKeyFromPEM(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式公钥")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}

	return rsaPublicKey, nil
}

// GetPublicKeyBase64 获取公钥的Base64编码（用于传输）
func GetPublicKeyBase64(publicKeyPEM string) string {
	return base64.StdEncoding.EncodeToString([]byte(publicKeyPEM))
}

// ParsePublicKeyFromBase64 从Base64编码解析公钥
func ParsePublicKeyFromBase64(publicKeyBase64 string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %v", err)
	}
	return string(publicKeyBytes), nil
}

// GenerateRandomSalt 生成随机盐值
func GenerateRandomSalt(length int) (string, error) {
	if length <= 0 {
		length = 32 // 默认32字节
	}

	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("生成随机盐值失败: %v", err)
	}

	// 返回Base64编码的盐值
	return base64.StdEncoding.EncodeToString(salt), nil
}

// HashPasswordWithSalt 使用盐值哈希密码
func HashPasswordWithSalt(password, salt string) (string, error) {
	// 将密码和盐值组合
	saltedPassword := password + salt

	// 使用SHA256哈希
	hash := sha256.Sum256([]byte(saltedPassword))

	// 返回Base64编码的哈希值
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// VerifyPasswordWithSalt 验证带盐值的密码
func VerifyPasswordWithSalt(password, salt, hashedPassword string) bool {
	// 使用相同的方式哈希输入的密码
	computedHash, err := HashPasswordWithSalt(password, salt)
	if err != nil {
		return false
	}

	// 比较哈希值
	return computedHash == hashedPassword
}

// GeneratePublicKeyFromPrivateKey 从私钥生成公钥
func GeneratePublicKeyFromPrivateKey(privateKeyPEM string) (string, error) {
	// 解析私钥
	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %v", err)
	}

	// 从私钥提取公钥
	publicKey := &privateKey.PublicKey

	// 编码公钥为PEM格式
	publicKeyPEM, err := encodePublicKeyToPEM(publicKey)
	if err != nil {
		return "", fmt.Errorf("编码公钥失败: %v", err)
	}

	return publicKeyPEM, nil
}
