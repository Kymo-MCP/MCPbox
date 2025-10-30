package app

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"qm-mcp-server/pkg/common"
	"qm-mcp-server/pkg/database/model"
	"qm-mcp-server/pkg/database/repository/mysql"
	"strings"
)

//go:embed template.json
var EmbeddedTemplatesFS embed.FS

// GetEmbeddedTemplateJSON returns embedded template JSON data
func GetEmbeddedTemplateJSON() []byte {
	data, err := fs.ReadFile(EmbeddedTemplatesFS, "template.json")
	if err != nil {
		log.Printf("Failed to read embedded template.json: %v", err)
		return nil
	}
	return data
}

// getDefaultEnvironmentID gets the default environment ID by name
func getDefaultEnvironmentID(ctx context.Context) (uint, error) {
	defaultEnv, err := mysql.McpEnvironmentRepo.FindByName(ctx, common.EnvironmentDefaultName)
	if err != nil {
		return 0, fmt.Errorf("failed to find default environment '%s': %w", common.EnvironmentDefaultName, err)
	}
	if defaultEnv == nil {
		return 0, fmt.Errorf("default environment '%s' not found", common.EnvironmentDefaultName)
	}
	log.Printf("Found default environment '%s' with ID: %d", common.EnvironmentDefaultName, defaultEnv.ID)
	return defaultEnv.ID, nil
}

// initMcpTemplateData initializes data using embedded JSON templates
func (a *App) initMcpTemplateData(ctx context.Context, env *model.McpEnvironment) error {
	log.Printf("Starting MCP template data initialization...")

	embeddedTemplateJSON := GetEmbeddedTemplateJSON()
	if len(embeddedTemplateJSON) == 0 {
		return fmt.Errorf("embedded template json is empty")
	}

	// Define structure corresponding to embedded JSON
	type embeddedTemplateItem struct {
		Environment    bool   `json:"environment,omitempty"` // Environment field to determine if default environment ID should be used
		Name           string `json:"name"`
		Port           int32  `json:"port"`
		InitScript     string `json:"init_script"`
		Command        string `json:"command"`
		StartupTimeout int32  `json:"startup_timeout"`
		RunningTimeout int32  `json:"running_timeout"`
		PackageID      string `json:"package_id"`
		PackageName    string `json:"package_name"` // Package name field for querying existing packages
		AccessType     string `json:"access_type"`
		McpProtocol    string `json:"mcp_protocol"`
		McpServers     string `json:"mcp_servers,omitempty"`
		ImgAddress     string `json:"img_address"`
		Notes          string `json:"notes,omitempty"`
		ServicePath    string `json:"service_path,omitempty"`
		IconPath       string `json:"icon_path,omitempty"`
	}

	var items []embeddedTemplateItem
	if err := json.Unmarshal(embeddedTemplateJSON, &items); err != nil {
		return fmt.Errorf("failed to unmarshal embedded template json: %w", err)
	}
	if len(items) == 0 {
		log.Printf("No template items found in embedded JSON")
		return nil
	}

	log.Printf("Found %d template items in embedded JSON", len(items))

	// Get default environment ID once for templates that need it
	var defaultEnvironmentID uint
	var defaultEnvErr error
	needsDefaultEnv := false

	// Check if any template needs default environment
	for _, it := range items {
		if it.Environment {
			needsDefaultEnv = true
			break
		}
	}

	if needsDefaultEnv {
		log.Printf("Some templates require default environment, fetching default environment ID...")
		defaultEnvironmentID, defaultEnvErr = getDefaultEnvironmentID(ctx)
		if defaultEnvErr != nil {
			return fmt.Errorf("failed to get default environment ID: %w", defaultEnvErr)
		}
	}

	createdCount := 0
	skippedCount := 0

	for _, it := range items {
		// Check for duplicates by name
		existing, err := mysql.McpTemplateRepo.FindByName(ctx, it.Name)
		if err == nil && existing != nil {
			log.Printf("Template '%s' already exists, skipping", it.Name)
			skippedCount++
			continue
		}

		// Determine environment ID based on environment field
		var environmentID int32
		var envSource string
		if it.Environment {
			// Use default environment ID when environment is true
			environmentID = int32(defaultEnvironmentID)
			envSource = fmt.Sprintf("default environment (ID: %d)", defaultEnvironmentID)
		} else {
			// Use provided environment ID when environment is false or not set
			environmentID = int32(env.ID)
			envSource = fmt.Sprintf("provided environment (ID: %d)", env.ID)
		}

		log.Printf("Creating template '%s' with %s", it.Name, envSource)

		// Check if package_name exists and set PackageID accordingly
		packageID := it.PackageID
		if it.PackageName != "" {
			// Query code package by original name
			if codePackage, err := mysql.McpCodePackageRepo.FindByOriginalName(ctx, it.PackageName); err == nil {
				// Package found, use its ID
				packageID = codePackage.PackageID
				log.Printf("Found existing package '%s' with ID: %s", it.PackageName, packageID)
			} else {
				// Package not found, keep original PackageID (could be empty)
				log.Printf("Package '%s' not found, using original PackageID: %s", it.PackageName, packageID)
			}
		}

		tpl := &model.McpTemplate{
			Name:           it.Name,
			Port:           it.Port,
			InitScript:     it.InitScript,
			Command:        it.Command,
			StartupTimeout: it.StartupTimeout,
			RunningTimeout: it.RunningTimeout,
			EnvironmentID:  environmentID,
			PackageID:      packageID,
			AccessType:     convertAccessType(it.AccessType),
			McpProtocol:    convertMcpProtocol(it.McpProtocol),
			ImgAddress:     it.ImgAddress,
			Notes:          it.Notes,
			ServicePath:    it.ServicePath,
			IconPath:       it.IconPath,
		}
		if len(it.McpServers) > 0 {
			// Parse JSON string and validate format
			mcpServersBytes := []byte(it.McpServers)
			if err := validateMcpServersConfig(json.RawMessage(mcpServersBytes)); err != nil {
				log.Printf("Invalid McpServers config for template '%s': %v", it.Name, err)
				return fmt.Errorf("invalid McpServers config for template '%s': %w", it.Name, err)
			}
			tpl.McpServers = json.RawMessage(mcpServersBytes)
		}

		if err := mysql.McpTemplateRepo.Create(ctx, tpl); err != nil {
			log.Printf("Failed to create template '%s': %v", it.Name, err)
			return fmt.Errorf("failed to create template '%s': %w", it.Name, err)
		}

		log.Printf("Successfully created template '%s' with ID: %d", it.Name, tpl.ID)
		createdCount++
	}

	log.Printf("MCP template data initialization completed. Created: %d, Skipped: %d", createdCount, skippedCount)
	return nil
}

// validateMcpServersConfig validates the JSON format of McpServers configuration
func validateMcpServersConfig(rawConfig json.RawMessage) error {
	if len(rawConfig) == 0 {
		return nil // Empty config is valid
	}

	// Try to unmarshal to McpServersConfig structure
	var config struct {
		McpServers map[string]json.RawMessage `json:"mcpServers"`
	}

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return fmt.Errorf("failed to unmarshal McpServers config: %w", err)
	}

	if len(config.McpServers) == 0 {
		return fmt.Errorf("no mcpServers found in config")
	}

	// Validate each server configuration
	for serverName, serverConfig := range config.McpServers {
		var mcpConfig struct {
			Args      []string `json:"args,omitempty"`
			Command   string   `json:"command,omitempty"`
			Type      string   `json:"type,omitempty"`
			Transport string   `json:"transport,omitempty"`
			URL       string   `json:"url,omitempty"`
		}

		if err := json.Unmarshal(serverConfig, &mcpConfig); err != nil {
			return fmt.Errorf("failed to unmarshal server config for '%s': %w", serverName, err)
		}

		// Determine protocol type based on configuration
		protocolType := determineProtocolType(mcpConfig)
		if protocolType == "" {
			return fmt.Errorf("cannot determine protocol type for server '%s'", serverName)
		}

		// Validate required fields based on protocol type
		if err := validateProtocolFields(protocolType, mcpConfig); err != nil {
			return fmt.Errorf("validation failed for server '%s': %w", serverName, err)
		}
	}

	return nil
}

// determineProtocolType determines the protocol type based on configuration
func determineProtocolType(config struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}) string {
	// 1. Check type and transport fields first
	if config.Type != "" {
		return config.Type
	}
	if config.Transport != "" {
		return config.Transport
	}

	// 2. Check URL field
	if config.URL != "" {
		if strings.Contains(strings.ToLower(config.URL), "sse") {
			return "sse"
		}
		return "steamableHttp"
	}

	// 3. Check command field
	if config.Command != "" {
		return "stdio"
	}

	// Default return empty string indicating unable to determine
	return ""
}

// validateProtocolFields validates protocol fields
func validateProtocolFields(protocolType string, config struct {
	Args      []string `json:"args,omitempty"`
	Command   string   `json:"command,omitempty"`
	Type      string   `json:"type,omitempty"`
	Transport string   `json:"transport,omitempty"`
	URL       string   `json:"url,omitempty"`
}) error {
	switch protocolType {
	case "sse", "steamableHttp":
		if config.URL == "" {
			return fmt.Errorf("%s protocol requires a valid url field", protocolType)
		}
	case "stdio":
		if config.Command == "" {
			return fmt.Errorf("%s protocol requires a valid command field", protocolType)
		}
	default:
		return fmt.Errorf("unknown protocol type: %s", protocolType)
	}
	return nil
}

func convertAccessType(s string) model.AccessType {
	switch s {
	case string(model.AccessTypeDirect):
		return model.AccessTypeDirect
	case string(model.AccessTypeProxy):
		return model.AccessTypeProxy
	case string(model.AccessTypeHosting):
		return model.AccessTypeHosting
	default:
		return model.AccessTypeProxy
	}
}

func convertMcpProtocol(s string) model.McpProtocol {
	switch s {
	case string(model.McpProtocolSSE):
		return model.McpProtocolSSE
	case string(model.McpProtocolStreamableHttp):
		return model.McpProtocolStreamableHttp
	case string(model.McpProtocolStdio):
		return model.McpProtocolStdio
	default:
		return model.McpProtocolSSE
	}
}
