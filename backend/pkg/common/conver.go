package common

import (
	"fmt"
	codepb "qm-mcp-server/api/market/code"
	instancepb "qm-mcp-server/api/market/instance"
	"qm-mcp-server/pkg/database/model"
)

// ConvertToInstanceInfo converts database model to proto message
func ConvertToInstanceInfo(instance *model.McpInstance) *instancepb.ListResp_InstanceInfo {
	accessType, _ := ConvertToProtoAccessType(model.AccessType(instance.AccessType))
	mcpProtocol, _ := ConvertToProtoMcpProtocol(model.McpProtocol(instance.McpProtocol))
	tokens := ConvertToProtoMcpToken(instance.Tokens)
	return &instancepb.ListResp_InstanceInfo{
		InstanceId:                 instance.InstanceID,
		InstanceName:               instance.InstanceName,
		AccessType:                 accessType,
		Status:                     string(instance.Status),
		EnvironmentId:              uint32(instance.EnvironmentID),
		ContainerStatus:            string(instance.ContainerStatus),
		ContainerName:              instance.ContainerName,
		ContainerServiceName:       instance.ContainerServiceName,
		ContainerIsReady:           instance.ContainerIsReady,
		ContainerCreateOptions:     string(instance.ContainerCreateOptions),
		ContainerLastMessage:       instance.ContainerLastMessage,
		ContainerInitTimeoutStopAt: instance.StartupTimeout,
		ContainerRunTimeoutStopAt:  instance.RunningTimeout,
		SourceConfig:               string(instance.SourceConfig),
		TargetConfig:               string(instance.TargetConfig),
		PublicProxyConfig:          string(instance.PublicProxyConfig),
		CreatedAt:                  instance.CreatedAt.String(),
		UpdatedAt:                  instance.UpdatedAt.String(),
		McpProtocol:                mcpProtocol,
		Tokens:                     tokens,
		IconPath:                   instance.IconPath,
		ServicePath:                instance.ServicePath,
	}
}

// ConvertToModelMcpProtocol converts string to McpProtocol enum value
func ConvertToModelMcpProtocol(mcpProtocol instancepb.McpProtocol) (model.McpProtocol, error) {
	switch mcpProtocol {
	case instancepb.McpProtocol_STDIO:
		return model.McpProtocolStdio, nil
	case instancepb.McpProtocol_STEAMABLE_HTTP:
		return model.McpProtocolStreamableHttp, nil
	case instancepb.McpProtocol_SSE:
		return model.McpProtocolSSE, nil
	default:
		return model.McpProtocolStdio, fmt.Errorf("unknown mcp protocol: %v", mcpProtocol)
	}
}

// ConvertToProtoMcpProtocol converts model.McpProtocol to proto message McpProtocol
func ConvertToProtoMcpProtocol(mcpProtocol model.McpProtocol) (instancepb.McpProtocol, error) {
	switch mcpProtocol {
	case model.McpProtocolStdio:
		return instancepb.McpProtocol_STDIO, nil
	case model.McpProtocolStreamableHttp:
		return instancepb.McpProtocol_STEAMABLE_HTTP, nil
	case model.McpProtocolSSE:
		return instancepb.McpProtocol_SSE, nil
	default:
		return instancepb.McpProtocol_McpProtocolUnknown, fmt.Errorf("unknown mcp protocol: %v", mcpProtocol)
	}
}

// ConvertToModelSourceType converts string to SourceType enum value
func ConvertToModelSourceType(sourceType instancepb.SourceType) (model.SourceType, error) {
	switch sourceType {
	case instancepb.SourceType_MARKET:
		return model.SourceTypeMarket, nil
	case instancepb.SourceType_TEMPLATE:
		return model.SourceTypeTemplate, nil
	case instancepb.SourceType_CUSTOM:
		return model.SourceTypeCustom, nil
	default:
		return model.SourceTypeCustom, fmt.Errorf("unknown source type: %v", sourceType)
	}
}

// ConvertToProtoSourceType converts model.SourceType to proto message SourceType
func ConvertToProtoSourceType(sourceType model.SourceType) (instancepb.SourceType, error) {
	switch sourceType {
	case model.SourceTypeMarket:
		return instancepb.SourceType_MARKET, nil
	case model.SourceTypeTemplate:
		return instancepb.SourceType_TEMPLATE, nil
	case model.SourceTypeCustom:
		return instancepb.SourceType_CUSTOM, nil
	default:
		return instancepb.SourceType_SourceTypeUnknown, fmt.Errorf("unknown source type: %v", sourceType)
	}
}

// ConvertToModelAccessType converts proto message AccessType to model.AccessType
func ConvertToModelAccessType(accessType instancepb.AccessType) (model.AccessType, error) {
	switch accessType {
	case instancepb.AccessType_DIRECT:
		return model.AccessTypeDirect, nil
	case instancepb.AccessType_PROXY:
		return model.AccessTypeProxy, nil
	case instancepb.AccessType_HOSTING:
		return model.AccessTypeHosting, nil
	default:
		return "", fmt.Errorf("unknown access type: %v", accessType)
	}
}

// ConvertToProtoAccessType converts model.AccessType to proto message AccessType
func ConvertToProtoAccessType(accessType model.AccessType) (instancepb.AccessType, error) {
	switch accessType {
	case model.AccessTypeDirect:
		return instancepb.AccessType_DIRECT, nil
	case model.AccessTypeProxy:
		return instancepb.AccessType_PROXY, nil
	case model.AccessTypeHosting:
		return instancepb.AccessType_HOSTING, nil
	default:
		return instancepb.AccessType_AccessTypeUnknown, fmt.Errorf("unknown access type: %v", accessType)
	}
}

func ConvertToProtoMcpToken(tokens []model.McpToken) []*instancepb.McpToken {
	protoTokens := make([]*instancepb.McpToken, 0, len(tokens))
	for _, token := range tokens {
		protoTokens = append(protoTokens, &instancepb.McpToken{
			Token:     token.Token,
			ExpireAt:  token.ExpireAt,
			PublishAt: token.PublishAt,
			Usages:    token.Usages,
		})
	}
	return protoTokens
}

// convertProtoTokensToModel converts tokens from proto structure to model structure
func ConvertProtoTokensToModel(tokens []*instancepb.McpToken) []model.McpToken {
	var modelTokens []model.McpToken
	for _, token := range tokens {
		modelTokens = append(modelTokens, model.McpToken{
			Token:     token.Token,
			ExpireAt:  token.ExpireAt,
			PublishAt: token.PublishAt,
			Usages:    token.Usages,
		})
	}
	return modelTokens
}

// ConvertToModelPackageType converts proto PackageType to model PackageType
func ConvertToModelPackageType(packageType codepb.PackageType) (model.PackageType, error) {
	switch packageType {
	case codepb.PackageType_PackageTypeTar:
		return model.PackageTypeTar, nil
	case codepb.PackageType_PackageTypeZip:
		return model.PackageTypeZip, nil
	case codepb.PackageType_PackageTypeTarGz:
		return model.PackageTypeTarGz, nil
	case codepb.PackageType_PackageTypeDxt:
		return model.PackageTypeDxt, nil
	case codepb.PackageType_PackageTypeMcpb:
		return model.PackageTypeMcpb, nil
	default:
		return model.PackageTypeUnknown, fmt.Errorf("unknown package type: %v", packageType)
	}
}

// ConvertToProtoPackageType converts model PackageType to proto PackageType
func ConvertToProtoPackageType(packageType model.PackageType) (codepb.PackageType, error) {
	switch packageType {
	case model.PackageTypeTar:
		return codepb.PackageType_PackageTypeTar, nil
	case model.PackageTypeZip:
		return codepb.PackageType_PackageTypeZip, nil
	case model.PackageTypeTarGz:
		return codepb.PackageType_PackageTypeTarGz, nil
	case model.PackageTypeDxt:
		return codepb.PackageType_PackageTypeDxt, nil
	case model.PackageTypeMcpb:
		return codepb.PackageType_PackageTypeMcpb, nil
	default:
		return codepb.PackageType_PackageTypeUnspecified, fmt.Errorf("unknown package type: %v", packageType)
	}
}
