import { store } from '@/stores'
import { ref } from 'vue'
import AuthAPI, { type LoginFormData, type EncryptionInfo } from '@/api/auth-api.ts'
import UserAPI, { type UserInfo } from '@/api/user/user-api'
import { defineStore } from 'pinia'
import { Storage } from '@/utils/storage'
import { useRouterHooks } from '@/utils/url'
import { useStorage } from '@vueuse/core'
import { ElLoading } from 'element-plus'
import { KEYUTIL, KJUR, RSAKey, hextob64 } from 'jsrsasign'

const { reload } = useRouterHooks()
export const useUserStore = defineStore('user', () => {
  const userInfo = useStorage('userInfo', {} as UserInfo)
  const encryptionInfo = ref<EncryptionInfo>({} as EncryptionInfo)
  const loginFormData = useStorage('loginFormData', {} as LoginFormData)

  // PEM interchange ArrayBuffer
  function pemToArrayBuffer(pem: string) {
    // remove PEM empty info
    const stripped = pem
      .replace(/-----BEGIN PUBLIC KEY-----/, '')
      .replace(/-----END PUBLIC KEY-----/, '')
      .replace(/\s+/g, '')
    // Base64 interchange Binary string
    const binary = atob(stripped)
    // Uint8Array
    const buffer = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i++) {
      buffer[i] = binary.charCodeAt(i)
    }
    return buffer
  }
  /**
   * RSA-OPEN
   * @param plaintext
   * @returns
   */
  const encryptPassword = async (data: string) => {
    try {
      let publicKeyPem
      // 解码publicKeyPem
      if (encryptionInfo.value.publicKey) {
        publicKeyPem = atob(encryptionInfo.value.publicKey?.replace(/-/g, '+').replace(/_/g, '/'))
      } else {
        throw new Error('秘钥获取失败')
      }

      // 1. 将PEM公钥转换为ArrayBuffer
      const publicKeyBuffer = pemToArrayBuffer(publicKeyPem)
      // https 协议下加密原生
      if (window.isSecureContext && window.crypto?.subtle) {
        // 2. 导入公钥（指定RSA-OAEP算法）
        const publicKey = await window.crypto.subtle.importKey(
          'spki', // 公钥使用SPKI格式
          publicKeyBuffer,
          {
            name: 'RSA-OAEP',
            hash: 'SHA-256', // 与Node版本保持一致的哈希算法
          },
          false, // 公钥不可提取
          ['encrypt'], // 仅用于加密
        )

        // 3. 将明文转换为ArrayBuffer
        const encoder = new TextEncoder()
        const dataBuffer = encoder.encode(data)
        // 4. 执行RSA-OAEP加密
        const encryptedBuffer = await window.crypto.subtle.encrypt(
          { name: 'RSA-OAEP' },
          publicKey,
          dataBuffer,
        )
        // 5. 将加密结果转换为Base64字符串
        return btoa(String.fromCharCode(...new Uint8Array(encryptedBuffer)))
      }

      // http协议环境下加密
      // 转换公钥为jsrsasign可用的格式
      const publicKey = KEYUTIL.getKey(publicKeyPem) as RSAKey
      const encryptedHex = KJUR.crypto.Cipher.encrypt(
        data,
        publicKey,
        'RSAOAEP256', // 指定SHA-256哈希
      )
      // 转换为Base64
      return hextob64(encryptedHex)
    } catch (error: any) {
      console.error('加密失败:', error)
      throw new Error(`加密错误: ${error?.message}`)
    }
  }
  /**
   * Login
   *
   * @param {LoginFormData}
   * @returns
   */
  async function login(LoginFormData: LoginFormData) {
    loginFormData.value = LoginFormData
    const encryptedPassword = await encryptPassword(LoginFormData.password || '')
    return new Promise((resolve, reject) => {
      AuthAPI.login({
        encryptedPassword,
        keyId: encryptionInfo.value.keyId,
        timestamp: String(new Date().getTime()),
        username: LoginFormData.username,
      })
        .then((data) => {
          Storage.set('token', data.token)
          Storage.set('refreshToken', data.refreshToken)
          userInfo.value = data.userInfo
          resolve(data)
        })
        .catch((err) => {
          reject(err)
        })
    })
  }

  /**
   * Handle get Encryption info
   * @returns
   */
  function handleGetEncryptionKey() {
    return new Promise<EncryptionInfo>((resolve, reject) => {
      AuthAPI.getEncryptionKey()
        .then((data) => {
          if (!data) {
            reject('无法获取秘钥信息.')
            return
          }
          Object.assign(encryptionInfo.value, { ...data })
          resolve(data)
        })
        .catch((error) => reject(error))
    })
  }

  /**
   * Handle get user info data
   *
   * @returns {UserInfo}
   */
  function getUserInfo() {
    return new Promise<UserInfo>((resolve, reject) => {
      UserAPI.getInfo()
        .then((data) => {
          if (!data) {
            reject('Verification failed, please Login again.')
            return
          }
          Object.assign(userInfo.value, { ...data })
          resolve(data)
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  /**
   * Logout
   */
  function logout() {
    return new Promise<void>((resolve, reject) => {
      const loadingInstance = ElLoading.service({ fullscreen: true, text: '退出登录中' })
      AuthAPI.logout({
        token: Storage.get('token'),
        userId: userInfo.value.userId,
      })
        .then(() => {
          resetUserState()
          reload()
          resolve()
        })
        .catch((error) => {
          reject(error)
        })
        .finally(() => {
          loadingInstance.close()
        })
    })
  }

  /**
   * reset user status
   *
   */
  function resetUserState() {
    // remove token info
    Storage.remove('token')
    Storage.remove('refresh_token')
    // reset user info
    userInfo.value = {} as UserInfo
  }

  /**
   * refresh token
   */
  function refreshToken() {
    const refreshToken = Storage.get('refreshToken')

    if (!refreshToken) {
      return Promise.reject(new Error('没有有效的刷新令牌'))
    }
    return new Promise<void>((resolve, reject) => {
      AuthAPI.refreshToken(refreshToken as string)
        .then((data) => {
          Storage.set('token', data.token)
          Storage.set('refreshToken', data.refreshToken)
          resolve()
        })
        .catch((error) => {
          reject(error)
        })
    })
  }

  return {
    userInfo,
    isLogin: () => !!Storage.get('token'),
    getUserInfo,
    login,
    logout,
    // resetAllState,
    resetUserState,
    refreshToken,
    handleGetEncryptionKey,
    loginFormData,
  }
})

/**
 * 在组件外部使用UserStore的钩子函数
 * @see https://pinia.vuejs.org/core-concepts/outside-component-usage.html
 */
export function useUserStoreHook() {
  return useUserStore(store)
}
