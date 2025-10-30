import axios, { type InternalAxiosRequestConfig, type AxiosResponse } from 'axios'
import qs from 'qs'
import { useUserStoreHook } from '@/stores/modules/user-store'
import { Storage } from './storage'
import { ElMessage, ElNotification } from 'element-plus'
import router from '@/router'
import baseConfig from '@/config/base_config.ts'

/**
 * 创建 HTTP 请求实例
 */
const request = axios.create({
  baseURL: baseConfig.SERVER_BASE_URL,
  timeout: 60000,
  headers: { 'Content-Type': 'application/json;' },
  // headers: { 'Content-Type': 'application/json;charset=utf-8' },
  paramsSerializer: (params: unknown) => qs.stringify(params),
})
/**
 * 响应数据
 */
interface ApiResponse<T = any> {
  code: number
  data: T
  message: string
}

/**
 * 请求拦截器 - 添加 Authorization 头
 */
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // const accessToken = Storage.get('access_token')
    const token = Storage.get('token')
    const lang = { en: 'en-US', 'zh-cn': 'zh-CN' }
    // 如果 Authorization 设置为 no-auth，则不携带 Token
    if (config.headers.Authorization !== 'no-auth' && token) {
      config.headers.Authorization = `Bearer ${token}`
    } else {
      delete config.headers.Authorization
    }
    // 添加国际化
    config.headers['accept-language'] = lang[Storage.get('language') as keyof typeof lang]

    return config
  },
  (error: any) => {
    console.error('Request interceptor error:', error)
    return Promise.reject(error)
  },
)

/**
 * 响应拦截器 - 统一处理响应和错误
 */
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    // 如果响应是二进制流，则直接返回（用于文件下载、Excel 导出等）
    if (response.config.responseType === 'blob') {
      return response
    }

    const { code, data, message } = response.data

    // 请求成功

    if (code === 0) {
      return data
    } else {
      handleErrorStatus(code, response.config)
    }
    // 业务错误
    ElMessage.error(message || '系统出错')
    return Promise.reject(new Error(message || 'Business Error'))
  },
  async (error: any) => {
    console.error('Response interceptor error:', error)

    const { config, response } = error

    // 网络错误或服务器无响应
    if (!response) {
      ElMessage.error('网络连接失败，请检查网络设置')
      return Promise.reject(error)
    }

    const { code, message } = response.data as ApiResponse

    switch (code) {
      case 403:
        // Access Token 过期，尝试刷新
        return refreshTokenAndRetry(config)

      case 401:
        // Refresh Token 过期，跳转登录页
        await redirectToLogin('登录已过期，请重新登录')
        return Promise.reject(new Error(message || 'Refresh Token Invalid'))

      default:
        ElMessage.error(message || '请求失败')
        return Promise.reject(new Error(message || 'Request Error'))
    }
  },
)
/**
 * 重试请求的回调函数类型
 */
type RetryCallback = () => void
// Token 刷新相关状态
let isRefreshingToken = false
const pendingRequests: RetryCallback[] = []
/**
 * 刷新 Token 并重试请求
 */
async function refreshTokenAndRetry(config: InternalAxiosRequestConfig): Promise<any> {
  return new Promise((resolve, reject) => {
    // 封装需要重试的请求
    const retryRequest = () => {
      const newToken = Storage.get('token')
      if (newToken && config.headers) {
        config.headers.Authorization = `Bearer ${newToken}`
      }
      request(config).then(resolve).catch(reject)
    }

    // 将请求加入等待队列
    pendingRequests.push(retryRequest)

    // 如果没有正在刷新，则开始刷新流程
    if (!isRefreshingToken) {
      isRefreshingToken = true

      useUserStoreHook()
        .refreshToken()
        .then(() => {
          // 刷新成功，重试所有等待的请求
          pendingRequests.forEach((callback) => {
            try {
              callback()
            } catch (error) {
              console.error('Retry request error:', error)
            }
          })
          // 清空队列
          pendingRequests.length = 0
        })
        .catch(async (error) => {
          console.error('Token refresh failed:', error)
          // 刷新失败，清空队列并跳转登录页
          pendingRequests.length = 0
          await redirectToLogin('登录状态已失效，请重新登录')
          // 拒绝所有等待的请求
          pendingRequests.forEach(() => {
            reject(new Error('Token refresh failed'))
          })
        })
        .finally(() => {
          isRefreshingToken = false
        })
    }
  })
}

/**
 * 重定向到登录页面
 */
async function redirectToLogin(message: string = '请重新登录'): Promise<void> {
  try {
    ElNotification({
      title: '提示',
      message,
      type: 'warning',
      duration: 3000,
    })

    // 跳转到登录页，保留当前路由用于登录后跳转
    const currentPath = router.currentRoute.value.fullPath
    await useUserStoreHook().resetUserState()
    await router.push(`/login?redirect=${encodeURIComponent(currentPath)}`)
  } catch (error) {
    console.error('Redirect to login error:', error)
  }
}

/**
 * 其他错误类型处理
 */
async function handleErrorStatus(code: number, config: InternalAxiosRequestConfig) {
  switch (code) {
    case 1001:
      // Access Token 过期，尝试刷新
      return refreshTokenAndRetry(config)

    case 403:
      // Refresh Token 过期，跳转登录页
      await redirectToLogin('登录已过期，请重新登录')
    default:
      ElMessage.error('请求失败')
  }
}

export default request
