// serverce status
export enum InstanceStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
}
// Type of access
export enum AccessType {
  UNKONWN = 0, // Unkonwn
  DIRECT = 1, // direct
  PROXY = 2, // proxy
  HOSTING = 3, // hosting
}

// MCP Protocol
export enum McpProtocol {
  UNKONWN = 0, // Unkonwn
  SSE = 1, // SSE
  STEAMABLE_HTTP = 2, // STEAMABLE_HTTP
  STDIO = 3, // STDIO
}

// List of Container States
export enum ContainerOptions {
  PENDING = 'pending',
  RUNNING = 'running',
  INIT_TIMEOUT_STOP = 'init-timeout-stop',
  RUN_TIMEOUT_STOP = 'run-timeout-stop',
  EXCEPTION_FORCE_STOP = 'exception-force-stop',
  MANUAL_STOP = 'manual-stop',
  CREATE_FAILED = 'create-failed',
  RUNNING_UNREADY = 'running-unready',
}

// Source of Instance
export enum SourceType {
  UNKONWN = 0, // Unkonwn
  MARKET = 1, // market
  TEMPLATE = 2, // template
  CUSTOM = 3, // custom
}

// the default of instance
export enum InstanceData {
  // IMGADDRESS = 'ccr.ccs.tencentyun.com/itqm-private/mcp-hosting:v2', // imgAddress
  IMGADDRESS = 'ccr.ccs.tencentyun.com/itqm-private/mcp-hosting:v2.1', // imgAddress
  PORT = 8080, // port
  INITSCRIPT = `#!/bin/bash
# 初始化脚本
echo 'Initialization completed.'
`, // InitScript
  TIP_IMGADDRESS = `
  1. 基于 python:3.12-alpine 镜像构建，轻量且兼容主流容器环境<br />
  2. 预装组件及版本明确锁定：<br />
    &nbsp &nbsp 系统基础命令：tar、wget、zip、unzip<br />
    &nbsp &nbsp Python 环境：Python 3.12.11，配套工具 uv 0.7.12、uvx 0.7.12<br />
    &nbsp &nbsp Node.js 环境：Node.js v18.20.1，配套工具 npm 9.6.6、npx 9.6.6<br />
  3. 默认以 mcp-hosting 为启动命令，启动时自动完成 MCP STDIO 协议到 steamable-http 协议的转换，并支持 SSE 协议模式运行
  `,
  TIP_MCP_SERVER = `
    {
      "servers": [
        {
          "name": "example",
          "command": "node",
          "args": ["server.js"]
        }
      ]
    }
  `, // placeholderServer
  PACKAGE_PATH = '/app/codepkg/', // the default of package path
  COMMAND_TIP = 'mcp-hosting --port=%d --mcp-servers-config /app/mcp-servers.json', // default start command
}

export enum NodeVisible {
  RWO = 'ReadWriteOnce',
  ROM = 'ReadOnlyMany',
  RWM = 'ReadWriteMany',
}

export interface VolumeMountsItme {
  type: string
  nodeName: string
  hostPath: string
  mountPath: string
  pvcName: string
  readOnly: boolean
}

// base-form
export interface InstanceForm {
  instanceName: string
  accessType: AccessType
  mcpProtocol: McpProtocol
  notes: string
  mcpConfig: string
  environmentId: string
  iconPath: string
  servicePath: string
}

// creat-instance-form
export interface InstanceCreate extends InstanceForm {
  sourceType: SourceType
  name: string
  notes: string
  mcpServers: string
  packageId: string
  environmentId: string
  port: number
  environmentVariables: { key: string; value: string }[]
  volumeMounts: { [key: string]: any }[]
  initScript: InstanceData.INITSCRIPT
  command: string
}

// list-result-form
export interface InstanceResult extends InstanceForm {
  instanceId: string
  containerName: string
  containerStatus: ContainerOptions
  status: InstanceStatus
  publicProxyConfig: string
  createdAt: string
  environmentName: string
  containerIsReady: boolean
}

// template-form-by-instance
export interface TemplateForm extends InstanceForm {
  name: string
}

// template-result-by-instance
export interface TemplateResult extends InstanceForm {
  templateId: string
  name: string
  environmentName: string
  createdAt: string
}
