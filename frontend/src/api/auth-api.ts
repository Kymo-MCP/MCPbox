import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

const AuthAPI = {
  login(formData: LoginFormData) {
    // const formData = new FormData()
    // formData.append('username', data.username)
    // formData.append('encryptedPassword', data.password)
    // formData.append('captchaKey', data.captchaKey)
    // formData.append('captchaCode', data.captchaCode)
    return request<any, LoginResult<any>>({
      url: `${baseConfig.baseUrlVersion}/authz/login`,
      method: 'POST',
      data: formData,
    })
  },

  getEncryptionKey() {
    return request<unknown, EncryptionInfo>({
      url: `${baseConfig.baseUrlVersion}/authz/encryption-key`,
      method: 'POST',
      params: {},
    })
  },

  refreshToken(refreshToken: string) {
    return request<unknown, LoginResult<any>>({
      url: `${baseConfig.baseUrlVersion}/authz/refresh`,
      method: 'POST',
      data: {
        refreshToken,
      },
    })
  },

  logout(params: any) {
    return request({
      url: `${baseConfig.baseUrlVersion}/authz/logout`,
      method: 'POST',
      params,
    })
  },

  getCaptcha() {
    return {}
  },

  /**
   * change password API
   */
  changePassword(params: any) {
    return request({
      url: `${baseConfig.baseUrlVersion}/authz/users/update-password`,
      method: 'PUT',
      params,
    })
  },
}

export default AuthAPI
/** 登录表单数据 */
export interface LoginFormData {
  /** 用户名 */
  username: string
  /** 密码 */
  password?: string
  encryptedPassword: string
  keyId?: string
  timestamp?: string
  /** 验证码缓存key */
  // captchaKey: string

  /** 验证码 */
  // captchaCode: string
  /** 记住我 */
  // rememberMe: boolean
}
/** 登录响应 */
export interface LoginResult<T> {
  /** 访问令牌 */
  token: string
  /** 刷新令牌 */
  refreshToken: string
  /** 用户信息 */
  userInfo: T
  /** 过期时间(秒) */
  expiresIn: number
}
export interface EncryptionInfo {
  algorithm: string
  expiresAt: string
  issuedAt: string
  keyId: string
  publicKey: string
}
