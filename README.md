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


MCP CAN is an open-source, lightweight back-end platform focused on agile management of MCP services. It supports converting the stdio configuration protocol to the SSE protocol and combining it with a token verification mechanism to help users manage and organize MCP services efficiently and securely. It also supports rapid deployment and remote access of local MCP services based on containerization technology, and supports centralized configuration management of external MCP services. Its core functions are designed around "ease of deployment" and "basic management capabilities".
<img width="1847" height="900" alt="image" src="https://github.com/user-attachments/assets/efe4c922-cb2a-4f18-a9e4-5115ade21506" />




# ✨ Key Features
## 🚀1. Containerized Agile Deployment: Focusing on Rapid Code Package Deployment

Leveraging a standardized container environment (pre-installed with Node.js and Python runtime), it supports lightweight deployment of MCP services, including:

- **Code Package Upload**: Directly upload locally written MCP service code packages (e.g., .zip format), and the platform automatically unzips and adapts them to the container runtime environment;
- **Storage Volume Mounting**: By mounting external storage volumes, code files are synchronized to the container, suitable for scenarios requiring frequent code updates;
- **Toolkit Deployment**: Provides official Node.js and Python toolkits, allowing developers to deploy code with a single click via command-line tools.


## 🔗2. Protocol Conversion: Enabling Remote Access to Local Services

A built-in protocol conversion gateway supports automatically converting local MCP service interactions based on the "standard input/output (stdio) protocol" to the "streamable HTTP protocol," achieving the following without manual code modification:

- Quickly mapping local services to remotely accessible HTTP endpoints;
- External systems can call local MCP services via HTTP requests, solving cross-network access challenges.


## 🛡️3. Access Modes: Covering Basic Management Needs

The platform offers three access modes, focusing on core management scenarios with clearly defined functional boundaries:

### (1) Direct Connection Mode: External MCP Configuration Management
Used for centralized management of configurations for "externally accessible MCP services," supporting the input of connection parameters (such as address, port, protocol version) for external MCP services, which are then stored as configuration items. No other additional functions (such as status monitoring or call statistics) are provided.


### (2) Proxy Mode: Unified Access Entry Point
External MCP services are accessed externally through the platform's proxy. Core capabilities include:

- Hiding the service's true connection information (such as IP address, port, and token) to reduce exposure risks;
- Restricting access permissions for specified terminals based on basic access control;
- Providing access logs to trace request sources and call details.


### (3) Managed Mode: Platform Provides Runtime Environment + SSE Protocol Adaptation

The MCP service is fully deployed in the container provided by the platform. In addition to basic management capabilities, protocol conversion and adaptation are added, specifically supporting:

- **Automatic Protocol Conversion**: For managed MCP services, the interaction method based on the "Standard Input/Output (stdio) protocol" is automatically converted to the "SSE (Server-Sent Events) protocol," without the need for additional development of adaptation;
- **Unified SSE Access Address**: After successful deployment, an external access address based on the SSE protocol is automatically generated. External systems can directly interact with the MCP service using the SSE protocol through this address;
- Container lifecycle management (start/stop/restart);
- Real-time viewing of runtime logs (such as error logs and output logs).

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

## 🏗️ Architecture

### System Architecture
                                                   ┌──────────────────────────────┐
                                                   │         Web Console          │
                                                   │ (Admin Interface / Frontend) │
                                                   └──────────────────────────────┘
                                                                 │
                                                                 ▼
                                    ┌───────────────────────────────────────────────┐
                                    │             MCP Management Service            │
                                    │ (Instance / Template / Environment Mgmt)      │
                                    └───────────────────────────────────────────────┘
                                          │             │             │
                                          │             │             │
                                          ▼             ▼             ▼
        ┌────────────────────────┐   ┌────────────────────────┐   ┌────────────────────────┐
        │ Protocol Gateway       │   │ Container Orchestration│   │ PostgreSQL Database    │
        │ (stdio ↔ HTTP / SSE)   │   │ (Kubernetes Cluster)   │   │ (Configs, States, etc.)│
        └────────────────────────┘   └────────────────────────┘   └────────────────────────┘
                                          │
                                          │
                                          ▼
        ┌────────────────────────┐   ┌────────────────────────┐   ┌────────────────────────┐
        │ Code Package Repository│   │ Market Case Library    │   │ Available Case Library │
        │ (User-uploaded MCP     │   │ (BigModel / Modelscope)│   │ (AutoNavi / Calculator)│
        │  Service Packages)     │   │                        │   │                        │
        └────────────────────────┘   └────────────────────────┘   └────────────────────────┘

## 📌 Frequently Asked Questions  

### 1. Why choose MCPCAN instead of other management platforms?
MCPCAN is an open-source management platform designed specifically for the MCP service ecosystem, unlike general-purpose admin systems.
It supports MCP stdio ↔ SSE protocol conversion, three connection modes, and visualized traffic monitoring, enabling end-to-end management from configuration to distribution.
Compared to other platforms, MCPcAN focuses on unified management of large-model toolchains and intelligent services, making development and operations more efficient and secure.

### 2. What unique advantages does MCPCAN have compared to similar products?

- Stronger protocol compatibility: Natively supports MCP stdio/SSE conversion and multiple connection methods.

- More flexible operating modes: Switch freely among Direct, Proxy, and Hosted modes.

- Enhanced security mechanisms: Supports Token authentication and multi-level permission control.


### 3. In what scenarios is MCPCAN suitable?

MCPCAN is ideal for scenarios requiring centralized management of multiple MCP services, including but not limited to:

- Individual/Small Team MCP Service Development: Quickly deploy local code as a remote SSE protocol service for real-time data push testing or small application calls;

- External MCP Service Configuration Archiving: Centrally manage connection information for multiple external MCP services, avoiding scattered and lost configurations;

- Lightweight Real-Time API Deployment: Convert MCP services into SSE protocol APIs for use by small systems that need to obtain real-time data (such as front-end real-time dashboards, simple monitoring tools).

## 🤝 Contributing  
Contributions of any kind are welcome!

- New features and improvements
- Documentation improvements
- Bug reports & fixes
- Translations & suggestions

Welcome to Join our [Discord](https://discord.com/channels/1428637640856571995/1428637896532820038) community for discussion and support.


## 🙌 Acknowledgements  
- Thanks to the open-source libraries or frameworks used  
- Thanks to contributors and supporters  
