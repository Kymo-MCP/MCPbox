# MCPCan

<div align="center">
  <img src="https://img.shields.io/badge/Next.js-15.5.4-black?style=for-the-badge&logo=next.js" alt="Next.js"/>
  <img src="https://img.shields.io/badge/TypeScript-5.0-blue?style=for-the-badge&logo=typescript" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/MySQL-8.0-blue?style=for-the-badge&logo=mysql" alt="MySQL"/>
  <img src="https://img.shields.io/badge/Kubernetes-1.28-326ce5?style=for-the-badge&logo=kubernetes" alt="Kubernetes"/>
  <img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge" alt="GPL-3.0"/>
</div>

## 什么是 MCPCan？

MCPCan 是一个专注于高效管理 MCP（模型上下文协议）服务的开源平台，通过现代化的 Web 界面为 DevOps 和开发团队提供全面的 MCP 服务生命周期管理能力。

MCPCan 支持多协议兼容和转换，能够实现不同 MCP 服务架构之间的无缝集成，同时提供可视化监控、安全认证和一站式部署能力。

<div align="center">
  <img width="1879" height="896" alt="MCPCan Dashboard" src="https://github.com/user-attachments/assets/ee804f92-7e69-419b-8cfc-d5676783fe3d" />
</div>

## 核心特性

- **🎯 统一管理**: 集中管理所有 MCP 服务实例和配置
- **🔄 协议转换**: 支持多种 MCP 协议间的无缝转换
- **📊 实时监控**: 提供详细的服务状态和性能监控
- **🔐 安全认证**: 内置身份验证和权限管理系统
- **🚀 一站式部署**: 快速发布、配置和分发 MCP 服务
- **📈 可扩展性**: 基于 Kubernetes 的云原生架构

## DEMO 站 (建设中)

MCPCan 提供了一个在线 Demo 站，您可以在其中体验 MCPCan 的功能和性能。

建设中...

## 快速开始

详细部署说明请参考我们的[部署指南](https://kymo-mcp.github.io/mcpcan-deploy/)。

```bash
# 安装 Helm Chart repository
helm repo add mcpcan https://kymo-mcp.github.io/mcpcan-deploy/

# 更新 Helm repository
helm repo update mcpcan

# 安装最新版本
helm install mcpcan mcpcan/mcpcan-deploy

# 使用公共 IP 部署
helm install mcpcan mcpcan/mcpcan-deploy \
  --set global.publicIP=192.168.1.100 \
  --set infrastructure.mysql.auth.rootPassword=secure-password \
  --set infrastructure.redis.auth.password=secure-password

# 使用域名部署
helm install mcpcan mcpcan/mcpcan-deploy \
  --set global.domain=mcp.example.com \
  --set infrastructure.mysql.auth.rootPassword=secure-password \
  --set infrastructure.redis.auth.password=secure-password
```

## 组件

MCPCan 由多个关键组件组成，它们共同构成了 MCPCan 的功能框架，为用户提供全面的 MCP 服务管理能力。

| 项目 | 状态 | 描述 |
|------|------|------|
| [MCPCan-Web](frontend/) | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan Web UI (Next.js 前端) |
| [MCPCan-Backend](backend/) | ![Status](https://img.shields.io/badge/status-active-green) | MCPCan 后端服务 (Go 微服务) |
| [MCPCan-Gateway](backend/cmd/gateway/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP 网关服务 |
| [MCPCan-Market](backend/cmd/market/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP 服务市场 |
| [MCPCan-Authz](backend/cmd/authz/) | ![Status](https://img.shields.io/badge/status-active-green) | 认证和授权服务 |

## 技术栈

### 前端
- **框架**: Vue.js 3.5+ (Composition API)
- **语言**: TypeScript
- **样式**: UnoCSS, SCSS
- **UI 组件**: Element Plus
- **状态管理**: Pinia
- **构建工具**: Vite

### 后端
- **语言**: Go 1.24.2+
- **框架**: Gin, gRPC
- **数据库**: MySQL, Redis
- **容器**: Docker, Kubernetes

## 第三方项目

- [mcpcan-deploy](https://github.com/Kymo-MCP/mcpcan-deploy) - MCPCan 官方 Helm Charts 源码仓库
- [MCPCan Helm Charts](https://kymo-mcp.github.io/mcpcan-deploy/) - MCPCan 官方 Helm Charts 仓库

## 贡献

欢迎提交 PR 贡献代码。请参考 [CONTRIBUTING.md](CONTRIBUTING.md) 了解贡献指南。

在贡献之前，请：
1. 阅读我们的[行为准则](CODE_OF_CONDUCT.md)
2. 检查现有的 issues 和 pull requests
3. 遵循我们的编码标准和提交信息约定

## 安全

如果您发现安全漏洞，请参考我们的[安全政策](SECURITY.md)进行负责任的披露。

## 许可证

版权所有 (c) 2024-2025 MCPCan 团队，保留所有权利。

根据 GNU 通用公共许可证第 3 版 (GPLv3) 许可（"许可证"）；除非遵守许可证，否则您不得使用此文件。您可以在以下位置获得许可证副本：

https://www.gnu.org/licenses/gpl-3.0.html

除非适用法律要求或书面同意，否则根据许可证分发的软件按"原样"分发，不提供任何明示或暗示的保证或条件。请参阅许可证以了解许可证下的特定语言管理权限和限制。

## 社区与支持

- 📖 [文档](https://kymo-mcp.github.io/mcpcan-deploy/)
- 💬 [Discord 社区](https://discord.com/channels/1428637640856571995/1428637896532820038)
- 🐛 [问题跟踪](https://github.com/Kymo-MCP/mcpcan/issues)
- 📧 [邮件列表](mailto:opensource@kymo.cn)

## 致谢

- 感谢 [MCP 协议](https://modelcontextprotocol.io/) 社区
- 感谢所有贡献者和支持者
- 特别感谢使 MCPCan 成为可能的开源项目
