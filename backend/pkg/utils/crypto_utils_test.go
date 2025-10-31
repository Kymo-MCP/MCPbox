package utils_test

import (
	"encoding/base64"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"testing"
)

func TestRSAEncrypt(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		data      []byte
		publicKey string
	}{
		{
			name:      "RSAEncrypt success",
			data:      []byte("admin123"),
			publicKey: "LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FROEFNSUlCQ2dLQ0FRRUF6UnN2MjVVNEJqOURqODc4b0JFTwpYREh6ei9LbDFUMTZ3b2hoVzBLejY0VkdOMXFBb09ZTW1Zc2VEd2RHM3R4VXl5Qk5vSVppcTMvOGFLSXhhdDc2Ck9hYlRucXViRlkrODNvRWh4dWUxbWNaRUpTSFlJbkh3UlRjWDhLd3ZXZFZCQWRpUmZYRUR6Zm5TSEV5TUJPM1YKYjNCd2I5dG4vT3BmN2FSbFlxRm1IT1JkYklsRkhmVlpEM0laeTYrUG9tbnFzRVYrUUtkSEdQcWppMVZ2RDhCQgpSb3RnalByQnVkTm1YVnZTMFNoSUZOU3d2aHVSRkM4Y3NZRk1YT0ZQYldVUDU1dlg4MThxM211djBGYU85Um0yCllmZC8ra3JwM3ZFWEVSQURXY2tnTXZrMWJ3OUJaOG5ZbmRHTW5JbldqSjljQlVEK3dZV0NlOGR5bmlDakxBM1YKc1FJREFRQUIKLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Base64 decode
			publicKey, err := base64.StdEncoding.DecodeString(tt.publicKey)
			if err != nil {
				t.Fatalf("Base64 decoding failed: %v", err)
			}
			// Encrypt
			pwd, err := utils.RSAEncrypt(tt.data, string(publicKey))
			if err != nil {
				t.Fatalf("RSAEncrypt() failed: %v", err)
			}
			if len(pwd) == 0 {
				t.Fatalf("RSAEncrypt() failed: %v", pwd)
			}
			t.Logf("pwd encrypted:  %v", pwd)

		})
	}
}
