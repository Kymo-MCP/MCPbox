import request from '@/utils/request'
import baseConfig from '@/config/base_config.ts'

const UserAPI = {
  /**
   * 获取当前登录用户信息
   *
   * @returns 登录用户昵称、头像信息，包括角色和权限
   */
  getInfo() {
    return request<any, UserInfo>({
      url: `${baseConfig.baseUrlVersion}/authz/users/get-current-user`,
      method: 'get',
    })
  },
}

export default UserAPI

/** 登录用户信息 */
export interface UserInfo {
  /** 用户ID */
  userId?: string

  /** 用户名 */
  username?: string

  /** 昵称 */
  nickname?: string

  /** 头像URL */
  avatar?: string

  /** 角色 */
  roles: string[]

  /** 权限 */
  perms: string[]
}
