package utils_test

import (
	"qm-mcp-server/pkg/utils"
	"testing"
)

func TestGetHostIPs(t *testing.T) {
	tests := []struct {
		name string // description of this test case
	}{
		// TODO: Add test cases.
		{
			name: "TestGetHostIPs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := utils.GetHostIPs()
			if gotErr != nil {
				t.Errorf("GetHostIPs() failed: %v", gotErr)
				return
			}
			t.Logf("GetHostIPs() = %v", got)
		})
	}
}
