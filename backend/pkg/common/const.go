package common

import "time"

// 全局配置
const (
	DefaultPageSize = 10
	MaxPageSize     = 100

	AccessTokenExpireTime  = 60 * 60 * 24     // 1天
	RefreshTokenExpireTime = 60 * 60 * 24 * 7 // 7天

	DefaultLanguage    = "zh-CN"
	DefaultTheme       = "light"
	AutoLogoutTime     = 30 * 60 // 30分钟
	EnableNotification = true

	// 重放攻击时间窗口
	ReplayWindow = 5 * time.Second
	// 是否启用防重放
	EnableReplay = false
	// 是否启用防篡改
	EnableSign = false

	// 密码强度验证配置
	PasswordMinLength      = 6     // 最小长度
	PasswordMaxLength      = 128   // 最大长度
	PasswordRequireLetter  = true  // 是否要求包含字母
	PasswordRequireDigit   = true  // 是否要求包含数字
	PasswordRequireSpecial = false // 是否要求包含特殊字符（推荐但不强制）
	PasswordMinASCII       = 32    // 可打印ASCII字符最小值
	PasswordMaxASCII       = 126   // 可打印ASCII字符最大值

	// 头像上传路径
	AvatarPath = "/avatar"
	// 图片上传路径
	ImagesPath = "/images"
	// 静态资源访问路径前缀
	StaticPrefix = "/static"

	// 默认托管镜像地址
	DefatuleHostingImg = "ccr.ccs.tencentyun.com/itqm-private/mcp-hosting"

	SourceServerName = "qm-mcp-server"

	McpProxyServiceName = "mcp-gateway-svc"

	MarketServerPrefix = "MCP_MARKET_SERVER_PREFIX"

	MarketRoutePrefix = "/market"

	AuthzServerPrefix = "MCP_AUTHZ_SERVER_PREFIX"

	AuthzRoutePrefix = "/authz"

	GatewayServerPrefix = "MCP_GATEWAY_SERVER_PREFIX"
	GatewayRoutePrefix  = "/mcp-gateway"

	EnvironmentDefaultName = "Default-Kubernetes-Env"
)

var SupportImageTypes = []string{"jpg", "jpeg", "png", "gif", "webp"}
