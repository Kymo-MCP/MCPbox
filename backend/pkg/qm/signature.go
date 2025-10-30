package qm

import (
	"crypto/md5"
	"fmt"
)

// GenerateSignature generates a signature
//
// @param customerUuid Customer UUID
// @param timestamp    Timestamp
// @param secretKey    Secret key
// @return Signature
func GenerateSignature(customerUuid, timestamp, secretKey string) string {
	raw := customerUuid + "|" + timestamp + "|" + secretKey
	// Using simple MD5 signature here, more secure algorithms can be used in actual projects
	hash := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", hash)
}
