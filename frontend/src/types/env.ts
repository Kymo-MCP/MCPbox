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

export interface EnvResult {
  config?: string

  createdAt?: string

  environment?: string

  id?: number

  name?: string

  namespace?: string

  updatedAt?: string
  [property: string]: any
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

export interface PvcResult {
  name: string
  storageClass: string
  creationTime: string
  accessModes: NodeVisible
  capacity: number
  namespace: string
  pods: string[]
  status: string
  volumeName: string
}

export interface StorageClass {
  allowVolumeExpansion?: boolean

  mountOptions?: string[]

  name?: string

  parameters?: { [key: string]: string }

  provisioner?: string

  reclaimPolicy?: string

  volumeBindingMode?: string
  [property: string]: any
}

export interface nodeResult {
  allocatableCpu?: string

  allocatableMemory?: string

  allocatablePods?: string

  annotations?: { [key: string]: string }

  architecture?: string

  containerRuntime?: string

  creationTime?: string

  externalIp?: string

  internalIp?: string

  kernelVersion?: string

  labels?: { [key: string]: string }

  name?: string

  operatingSystem?: string

  roles?: string[]

  status?: string

  version?: string
}
