# MCPCAN：A lightweight MCP service management platform built on a containerized architecture.
<p align="Left">
   <strong>English</strong> | <a href="./README.cn.md">中文版</a> 
</p>

<img width="1879" height="896" alt="image" src="https://github.com/user-attachments/assets/ee804f92-7e69-419b-8cfc-d5676783fe3d" />

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/users/Calcium-Ion/packages/container/package/new-api">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="docker">
  </a>
  <a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://goreportcard.com/report/github.com/Calcium-Ion/new-api">
    <img src="https://goreportcard.com/badge/github.com/Calcium-Ion/new-api" alt="GoReportCard">
  </a>
</p>
</div>


## 🚀 Overview


MCP BOX is a lightweight platform focused on agile management of MCP services. It relies on containerization technology to achieve rapid deployment and remote access of local MCP services, while also supporting centralized configuration management of external MCP services. Its core functions are designed around "deployment convenience" and "basic management capabilities."
<img width="1847" height="900" alt="image" src="https://github.com/user-attachments/assets/efe4c922-cb2a-4f18-a9e4-5115ade21506" />




### ✨ Key Features

- **🛡️ Multi-protocol Compatibility and Conversion**: Supports automatic conversion of MCP's stdio configuration protocol to SSE configuration protocol, simplifying the development and integration process and enabling seamless docking and communication between systems of different architectures.
- **🔗 Multi-mode Connection Management**: Provides three protocol modes: direct connection mode, proxy mode, and managed mode, meeting multi-scenario requirements from local debugging to enterprise-level hosting.
- **📊 MCP Service Visual Monitoring**: Displays the running status, traffic data, and log records of each MCP instance in real time in the form of a list, helping administrators intuitively grasp the service health status.
- **🧩 Modular Service Management System**: Includes four modules: template management, instance management, environment management, and code package management, forming a complete service lifecycle management loop.
- **🚀Containerized Agile Deployment**: Focusing on Rapid Code Package Rollout
Leveraging a standardized container environment (pre-installed with Node.js and Python runtimes), we support lightweight deployment of MCP services. Deployment methods include:
● Code Package Upload: Directly upload locally written MCP service code packages (e.g., .zip format). The platform automatically decompresses and adapts them to the container runtime environment.
● Volume Mounting: Synchronize code files to the container by mounting external storage volumes. This is ideal for scenarios requiring frequent code updates.
● Toolkit Deployment: Provides official Node.js and Python toolkits, enabling developers to deploy code with one click using command-line tools. AA
- **🚀 Protocol Conversion**: Enable Remote Access to Local Services
A built-in protocol conversion gateway automatically converts local MCP service interactions based on the Standard Input/Output (STDIO) protocol to the Streamable HTTP protocol. This eliminates the need for manual code modifications and enables:
● Local services are quickly mapped to remotely accessible HTTP endpoints;
● External systems call local MCP services via HTTP requests, solving cross-network access challenges.
---

## 🛠️ Technology Stack

### Frontend
- **Framework**: Next.js 15.5.4 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4
- **UI Components**: Shadcn/UI
- **State Management**: React Hooks

### Backend
- **Runtime**: Node.js
- **API**: Next.js API Routes
- **Database ORM**: Prisma
- **Authentication**: NextAuth v5 with GitHub OAuth

### Infrastructure
- **Container Orchestration**: Kubernetes
- **Database**: PostgreSQL (via KubeBlocks)
- **Web Terminal**: ttyd
- **Container Image**: fullstack-web-runtime (Custom Docker image with development tools）

  

## ⚡ Deployment
For detailed deployment steps, please refer to [https://kymo-mcp.github.io/mcp-box-deploy/](https://github.com/Kymo-MCP/mcp-box-deploy/tree/v1.0.0-dev) for details.


## 📌 Frequently Asked Questions  

### 1. Why choose MCPBOX instead of other management platforms?
MCPBOX is an open-source management platform designed specifically for the MCP service ecosystem, unlike general-purpose admin systems.
It supports MCP stdio ↔ SSE protocol conversion, three connection modes, and visualized traffic monitoring, enabling end-to-end management from configuration to distribution.
Compared to other platforms, MCPBOX focuses on unified management of large-model toolchains and intelligent services, making development and operations more efficient and secure.

### 2. What unique advantages does MCPBOX have compared to similar products?

- Stronger protocol compatibility: Natively supports MCP stdio/SSE conversion and multiple connection methods.

- More flexible operating modes: Switch freely among Direct, Proxy, and Hosted modes.

- Enhanced security mechanisms: Supports Token authentication and multi-level permission control.

- More intuitive operations experience: Graphical interface displays real-time service status and traffic data.

### 3. In what scenarios is MCPBOX suitable?

MCPBOX is ideal for scenarios requiring centralized management of multiple MCP services, including but not limited to:

- AI Agent development platforms that need to manage multiple service instances;

- Enterprises that need secure and controlled distribution of MCP tools across teams;

- Developers who want to monitor MCP service performance and traffic data for intelligent operations.

- Whether for individual research or enterprise deployment, MCPBOX provides a flexible architecture and open capabilities to support real-world business applications. 

---

## 🤝 Contributing  
Contributions of any kind are welcome!

- New features and improvements
- Documentation improvements
- Bug reports & fixes
- Translations & suggestions

Welcome to Join our [Discord](https://discord.com/channels/1428637640856571995/1428637896532820038) community for discussion and support.


---

## 📄 License  
This project is licensed under the **MIT License**. For details, see [LICENSE](./LICENSE).  

---

## 🙌 Acknowledgements  
- Thanks to the open-source libraries or frameworks used  
- Thanks to contributors and supporters  
