import { NodeVisible } from './instance.ts'

export enum EnvType {
  K8S = 'kubernetes',
  DOCKER = 'docker',
}

export enum EnvFormData {
  TIP_CONFIG = `
    apiVersion: v1
    kind: ConfigMap
    metadata:
    name: example
    data:
    key: value
  `,
}

// code list
export interface Code {
  id: string
  name: string
  path: string
  size: number
  type: 1 | 2
  createdAt: string
  updatedAt: string
}

export interface PvcForm {
  name: string
  storageClass: string
  accessMode: NodeVisible
  storageSize: number
  nodeName: string
  hostPath: string
}
