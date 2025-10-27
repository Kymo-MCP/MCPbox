package qm

import (
	"crypto/md5"
	"fmt"
)

// GenerateSignature 生成签名
//
// @param customerUuid 客户UUID
// @param timestamp    时间戳
// @param secretKey    密钥
// @return 签名
func GenerateSignature(customerUuid, timestamp, secretKey string) string {
	raw := customerUuid + "|" + timestamp + "|" + secretKey
	// 这里使用简单的MD5签名，实际项目中可以使用更安全的算法
	hash := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", hash)
}
