package i18n_test

import (
	"qm-mcp-server/pkg/i18n"
	"testing"
)

func TestNewMessageLoader(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		basePath string
		want     *i18n.MessageLoader
	}{
		// TODO: Add test cases.
		{
			name:     "ValidBasePath",
			basePath: "./locales",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalMessageLoader := i18n.GetGlobalMessageLoader()
			err := globalMessageLoader.LoadMessages()
			if err != nil {
				t.Errorf("LoadMessages() error = %v", err)
			}
		})
	}
}
