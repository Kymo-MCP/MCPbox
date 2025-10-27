import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const ResourceAPI = {
  // 节点列表
  nodeList(data: TableData) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/resources/nodes`,
      method: 'POST',
      data,
    })
  },
}

// 列表请求
export interface TableData {
  /** 页码 */
  page: string
  /** 每页显示数量 */
  pageSize: string
  /** 允许传入其他任意类型的参数 */
  [key: string]: any
}
// 列表返回
export interface List<T = any> {
  list: T[]
  page: number
  pageSize: number
  total: number
}
