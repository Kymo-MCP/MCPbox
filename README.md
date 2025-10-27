# MCPBox

<div align="center">
  <img src="https://img.shields.io/badge/Next.js-15.5.4-black?style=for-the-badge&logo=next.js" alt="Next.js"/>
  <img src="https://img.shields.io/badge/TypeScript-5.0-blue?style=for-the-badge&logo=typescript" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/PostgreSQL-14-blue?style=for-the-badge&logo=postgresql" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Kubernetes-1.28-326ce5?style=for-the-badge&logo=kubernetes" alt="Kubernetes"/>
  <img src="https://img.shields.io/badge/License-GPL--3.0-blue?style=for-the-badge" alt="GPL-3.0"/>
</div>

## What is MCPBox?

MCPBox is an open-source platform focused on efficient management of MCP (Model Context Protocol) services, providing DevOps and development teams with comprehensive MCP service lifecycle management capabilities through a modern web interface. <mcreference link="https://github.com/jumpserver/jumpserver" index="0">0</mcreference>

MCPBox supports multi-protocol compatibility and conversion, enabling seamless integration between different MCP service architectures while providing visual monitoring, security authentication, and one-stop deployment capabilities.

<div align="center">
  <img width="1879" height="896" alt="MCPBox Dashboard" src="https://github.com/user-attachments/assets/ee804f92-7e69-419b-8cfc-d5676783fe3d" />
</div>

## ✨ Key Features

- **🛡️ Multi-protocol Compatibility**: Supports automatic conversion between MCP's stdio and SSE protocols
- **🔗 Multi-mode Connection Management**: Provides direct connection, proxy, and managed modes
- **📊 Visual Service Monitoring**: Real-time monitoring of MCP service status, traffic, and logs
- **🧩 Modular Management System**: Complete service lifecycle management with template, instance, environment, and code package management
- **🔒 Security & Authentication**: Token-based authentication with multi-level permission control
- **🚀 One-stop Deployment**: Quick release, configuration, and distribution of MCP services

## Quickstart

Prepare a clean Linux Server (64 bit, >= 4c8g)

```bash
# Clone the repository
git clone https://github.com/Kymo-MCP/MCPbox.git
cd MCPbox

# Quick deployment using Kubernetes
kubectl apply -f deploy/

# Or use Helm for advanced configuration
helm install mcpbox ./helm --namespace mcpbox --create-namespace
```

Access MCPBox in your browser at `http://your-mcpbox-ip/`

**Default credentials:**
- Username: `admin`
- Password: `ChangeMe`

For detailed deployment instructions, please refer to our [Deployment Guide](https://github.com/Kymo-MCP/mcp-box-deploy).

## Screenshots

| Dashboard | Service Management | Monitoring |
|-----------|-------------------|------------|
| ![Dashboard](docs/images/dashboard.png) | ![Services](docs/images/services.png) | ![Monitoring](docs/images/monitoring.png) |

## Components

MCPBox consists of multiple key components, which collectively form the functional framework of MCPBox, providing users with comprehensive MCP service management capabilities.

| Project | Status | Description |
|---------|--------|-------------|
| [MCPBox-Web](web/) | ![Status](https://img.shields.io/badge/status-active-green) | MCPBox Web UI (Next.js Frontend) |
| [MCPBox-Backend](backend/) | ![Status](https://img.shields.io/badge/status-active-green) | MCPBox Backend Services (Go Microservices) |
| [MCPBox-Gateway](backend/cmd/gateway/) | ![Status](https://img.shields.io/badge/status-active-green) | API Gateway and Load Balancer |
| [MCPBox-Market](backend/cmd/market/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP Service Marketplace |
| [MCPBox-Authz](backend/cmd/authz/) | ![Status](https://img.shields.io/badge/status-active-green) | Authentication and Authorization Service |
| [MCPBox-Proxy](mcp-proxy/) | ![Status](https://img.shields.io/badge/status-active-green) | MCP Protocol Proxy and Converter |

## Technology Stack

### Frontend
- **Framework**: Next.js 15.5.4 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4
- **UI Components**: Shadcn/UI
- **State Management**: React Hooks

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin, gRPC
- **Database**: PostgreSQL, Redis
- **Message Queue**: NATS
- **Container**: Docker, Kubernetes

### Infrastructure
- **Container Orchestration**: Kubernetes
- **Database**: PostgreSQL (via KubeBlocks)
- **Monitoring**: Prometheus, Grafana
- **Logging**: ELK Stack

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   API Gateway   │    │  Backend Services│
│    (Next.js)    │◄──►│     (Go)        │◄──►│      (Go)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  MCP Services   │    │   Database      │
                       │   (Containers)  │    │  (MySQL)        │
                       └─────────────────┘    └─────────────────┘
```

## Third-party Projects

- [mcpbox-grafana-dashboard](https://github.com/your-org/mcpbox-grafana-dashboard) - MCPBox with Grafana dashboard
- [mcpbox-helm-charts](https://github.com/Kymo-MCP/mcp-box-deploy) - Official Helm charts for MCPBox

## Contributing

Welcome to submit PR to contribute. Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Before contributing, please:
1. Read our [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing issues and pull requests
3. Follow our coding standards and commit message conventions

## Security

If you discover a security vulnerability, please refer to our [Security Policy](SECURITY.md) for responsible disclosure guidelines.

## License

Copyright (c) 2024-2025 MCPBox Team, All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3) (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

https://www.gnu.org/licenses/gpl-3.0.html

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Community & Support

- 📖 [Documentation](https://mcpbox.dev/docs)
- 💬 [Discord Community](https://discord.com/channels/1428637640856571995/1428637896532820038)
- 🐛 [Issue Tracker](https://github.com/Kymo-MCP/MCPbox/issues)
- 📧 [Mailing List](mailto:opensource@kymo.cn)

## Acknowledgements

- Thanks to the [MCP Protocol](https://modelcontextprotocol.io/) community
- Thanks to all contributors and supporters
- Special thanks to the open-source projects that make MCPBox possible
