import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

export const CodeAPI = {
  // 代码包列表
  list(data: TableData | null) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/code/packages`,
      method: 'GET',
      data,
    })
  },
  // 下载代码包
  download(code: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/code/download/${code.id}`,
      method: 'GET',
      responseType: 'blob',
    })
  },
  // 删除代码包
  delete(id: string) {
    return request<any, List>({
      url: `${baseConfig.baseUrlVersion}/market/code/packages/${id}`,
      method: 'DELETE',
    })
  },
  // 获取代码包文件树结构
  fileTree(params: any) {
    return request<any, FileTree>({
      url: `${baseConfig.baseUrlVersion}/market/code/tree`,
      method: 'GET',
      params,
    })
  },
  // 获取文件内容
  fileContent(params: any) {
    return request<any, any>({
      url: `${baseConfig.baseUrlVersion}/market/code/get`,
      method: 'GET',
      params,
    })
  },
}

// 列表请求
export interface TableData {
  /** 页码 */
  page: number
  /** 每页显示数量 */
  pageSize: number
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

// 列表返回
export interface FileTree {
  fileStructure: FileTree[]
  isDir: boolean
  name: string
  path: string
}
