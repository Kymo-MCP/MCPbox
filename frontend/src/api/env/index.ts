import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const EnvAPI = {
  // 环境列表
  list(data: TableData | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/environments`,
      method: 'GET',
      data,
    })
  },
  // 命名空间列表
  namespaceList(data: any | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/environments/namespaces`,
      method: 'POST',
      data,
    })
  },
  // 创建运行环境
  createEnv(data: any) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/environments`,
      method: 'POST',
      data,
    })
  },
  // 删除运行环境
  delete(id: string) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/environments/${id}`,
      method: 'DELETE',
    })
  },
  // 编辑运行环境
  editEnv(data: any) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/environments/${data.id}`,
      method: 'PUT',
      data,
    })
  },
  // 连通性测试
  testEnv(id: string) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/environments/${id}/test`,
      method: 'POST',
    })
  },
}

// 列表请求
export interface TableData {
  /** 页码 */
  page?: string
  /** 每页显示数量 */
  pageSize?: string
  /** 允许传入其他任意类型的参数 */
  [key: string]: any
}
// 列表返回
export interface List<T = any> {
  code: number
  list: T[]
  page: number
  pageSize: number
  total: number
}

// 创建pvc
export interface PvcParams<T = any> {
  code: number
  data: T
  message: string
}

export const PvcAPI = {
  // PVC列表
  list(params: TableData | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/resources/pvcs`,
      method: 'GET',
      params,
    })
  },
  // 存储类型列表
  storageList(params: any | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/resources/storage-classes`,
      method: 'GET',
      params,
    })
  },
  // 创建PVC
  createPvc(data: any) {
    return request<any, PvcParams>({
      url: `${baseConfig.baseUrlVersion}/market/resources/pvcs`,
      method: 'POST',
      data,
    })
  },
}

export const NodeAPI = {
  // Node节点列表
  list(params: any | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/resources/nodes`,
      method: 'GET',
      params,
    })
  },
}
