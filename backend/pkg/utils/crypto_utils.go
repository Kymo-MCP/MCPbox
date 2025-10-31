package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"time"

	"github.com/google/uuid"
)

const AlgorithmRSA2048 = "RSA-2048"

const KeySize = 2048

// RSAKeyPair RSA key pair
type RSAKeyPair struct {
	KeyID      string
	PublicKey  string // PEM format
	PrivateKey string // PEM format
	KeySize    int
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// GenerateRSAKeyPair generate RSA key pair
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA private key: %v", err)
	}

	// Encode private key to PEM format
	privateKeyPEM, err := encodePrivateKeyToPEM(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key: %v", err)
	}

	// Encode public key to PEM format
	publicKeyPEM, err := encodePublicKeyToPEM(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode public key: %v", err)
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

// RSAEncrypt encrypt data using RSA public key
func RSAEncrypt(data []byte, publicKeyPEM string) (string, error) {
	// Parse public key
	publicKey, err := parsePublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	// Encrypt using OAEP padding
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
	if err != nil {
		return "", fmt.Errorf("RSA encryption failed: %v", err)
	}

	// Return Base64 encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// RSADecrypt decrypt data using RSA private key
func RSADecrypt(ciphertext string, privateKeyPEM string) ([]byte, error) {
	// Parse private key
	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	// Decrypt using OAEP padding
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, []byte(ciphertext), nil)
	if err != nil {
		return nil, fmt.Errorf("RSA decryption failed: %v", err)
	}

	return plaintext, nil
}

// EncryptPassword encrypt password (with timestamp validation)
func EncryptPassword(password string, timestamp int64, publicKeyPEM string) (string, error) {
	// Construct data to be encrypted: password|timestamp
	data := fmt.Sprintf("%s|%d", password, timestamp)
	return RSAEncrypt([]byte(data), publicKeyPEM)
}

// DecryptPassword decrypt password (with timestamp validation)
func DecryptPassword(encryptedPassword string, privateKeyPEM string, maxAge time.Duration) (string, int64, error) {
	// Decrypt data
	plaintext, err := RSADecrypt(encryptedPassword, privateKeyPEM)
	if err != nil {
		return "", 0, err
	}

	// Parse password and timestamp
	data := string(plaintext)
	var password string
	var timestamp int64
	n, err := fmt.Sscanf(data, "%s|%d", &password, &timestamp)
	if err != nil || n != 2 {
		return "", 0, fmt.Errorf("failed to parse password data: %v", err)
	}

	// Validate timestamp
	if maxAge > 0 {
		now := time.Now().Unix()
		if now-timestamp > int64(maxAge.Seconds()) {
			return "", 0, fmt.Errorf("password expired, timestamp: %d, current: %d", timestamp, now)
		}
	}

	return password, timestamp, nil
}

// generateKeyID generate key ID
func generateKeyID() string {
	return fmt.Sprintf("key_%s_%d", uuid.New().String()[:8], time.Now().Unix())
}

// encodePrivateKeyToPEM encode private key to PEM format
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

// encodePublicKeyToPEM encode public key to PEM format
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

// parsePrivateKeyFromPEM parse private key from PEM format
func parsePrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM format private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaPrivateKey, nil
}

// parsePublicKeyFromPEM parse public key from PEM format
func parsePublicKeyFromPEM(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM format public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// GetPublicKeyBase64 get Base64 encoded public key (for transmission)
func GetPublicKeyBase64(publicKeyPEM string) string {
	return base64.StdEncoding.EncodeToString([]byte(publicKeyPEM))
}

// ParsePublicKeyFromBase64 parse public key from Base64 encoding
func ParsePublicKeyFromBase64(publicKeyBase64 string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return "", fmt.Errorf("Base64 decoding failed: %v", err)
	}
	return string(publicKeyBytes), nil
}

// GenerateRandomSalt generate random salt value
func GenerateRandomSalt(length int) (string, error) {
	if length <= 0 {
		length = 32 // default 32 bytes
	}

	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate random salt: %v", err)
	}

	// Return Base64 encoded salt
	return base64.StdEncoding.EncodeToString(salt), nil
}

// HashPasswordWithSalt hash password with salt
func HashPasswordWithSalt(password, salt string) (string, error) {
	// Combine password and salt
	saltedPassword := password + salt

	// Use SHA256 hash
	hash := sha256.Sum256([]byte(saltedPassword))

	// Return Base64 encoded hash
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

// VerifyPasswordWithSalt verify password with salt
func VerifyPasswordWithSalt(password, salt, hashedPassword string) bool {
	// Hash the input password in the same way
	computedHash, err := HashPasswordWithSalt(password, salt)
	if err != nil {
		return false
	}

	// Compare hashes
	return computedHash == hashedPassword
}

// GeneratePublicKeyFromPrivateKey generate public key from private key
func GeneratePublicKeyFromPrivateKey(privateKeyPEM string) (string, error) {
	// Parse private key
	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Extract public key from private key
	publicKey := &privateKey.PublicKey

	// Encode public key to PEM format
	publicKeyPEM, err := encodePublicKeyToPEM(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to encode public key: %v", err)
	}

	return publicKeyPEM, nil
}
