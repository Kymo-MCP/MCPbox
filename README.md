# MCPBOX - An open source platform focused on efficient management of MCP services, and a mid- and back-end management tool that supports protocol conversion.

<div align="center">
  <img src="https://img.shields.io/badge/Next.js-15.5.4-black?style=for-the-badge&logo=next.js" alt="Next.js"/>
  <img src="https://img.shields.io/badge/TypeScript-5.0-blue?style=for-the-badge&logo=typescript" alt="TypeScript"/>
  <img src="https://img.shields.io/badge/PostgreSQL-14-blue?style=for-the-badge&logo=postgresql" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Kubernetes-1.28-326ce5?style=for-the-badge&logo=kubernetes" alt="Kubernetes"/>
  <img src="https://img.shields.io/badge/Claude_Code-AI-purple?style=for-the-badge" alt="Claude Code"/>
</div>

## 🚀 Overview

MCPBOX is an open-source MCP service management platform designed to help users efficiently manage and organize MCP services. As a mid- and back-end product, MCPBOX supports converting MCP's stdio configuration protocol to the SSE configuration protocol, making it easier for administrators to distribute these protocols to other users. It also incorporates a unique token verification mechanism to ensure secure and controllable use.
<img width="1920" height="911" alt="867d671a9265e155d8988908e1932aef" src="https://github.com/user-attachments/assets/faef6d8e-d0d7-4203-8f07-cfeb66e24fd7" />


### ✨ Key Features

- **🛡️ Multi-protocol Compatibility and Conversion**: Supports automatic conversion of MCP's stdio configuration protocol to SSE configuration protocol, simplifying the development and integration process and enabling seamless docking and communication between systems of different architectures.
- **🔗 Multi-mode Connection Management**: Provides three protocol modes: direct connection mode, proxy mode, and managed mode, meeting multi-scenario requirements from local debugging to enterprise-level hosting.
- **📊 MCP Service Visual Monitoring**: Real-time displays the running status, traffic data, and log records of each MCP instance in the form of charts and lists, helping administrators intuitively grasp the service health status.
- **🧩 Modular Service Management System**: Includes four modules: template management, instance management, environment management, and code package management, forming a complete service lifecycle management loop.
- **🔒 Security Authentication and Permission Control**: Supports Token-based security verification mechanism. Administrators can assign different levels of access permissions to achieve secure multi-user collaboration and resource isolation.
- **🚀 One-stop Distribution and Deployment Capability**: Provides quick release, configuration, and distribution functions for MCP services, supports batch operations and multi-environment synchronization, allowing teams to efficiently complete service deployment and sharing.
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

  
## 🛠️ Installation  
Provide detailed steps to install and configure the environment.  
```bash
# Clone the repository
git clone https://github.com/your-username/your-project.git

# Navigate to the project directory
cd your-project

# Install dependencies
pip install -r requirements.txt
```

---

## ⚡ Deployment
For detailed deployment steps, please refer to https://kymo-mcp.github.io/mcp-box-deploy/ for details.


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
