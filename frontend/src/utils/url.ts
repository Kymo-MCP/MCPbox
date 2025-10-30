import { useRouter } from 'vue-router'

/**
 * 对象转url字符串参数
 */
export const obj2UrlString = (obj: any) => {
  if (!obj) {
    return ''
  }
  // 使用 Object.entries 获取对象的键值对数组，然后使用 filter 方法过滤掉值为空字符串的项
  const filteredEntries = Object.entries(obj).filter(
    ([, value]) => value !== '' && value !== undefined,
  )
  // 使用 map 方法将每个键值对转换成 'key=value' 的形式，然后使用 join 方法将它们连接成一个字符串
  const queryString = filteredEntries.map(([key, value]) => `${key.trim()}=${value}`).join('&')

  // 如果 queryString 不为空，则在其前面添加 '?'，否则返回空字符串
  return queryString ? `?${queryString}` : ''
}

interface jumpType {
  url: string
  data?: Record<string, any>
  isOpen?: boolean
  isCurrentProject?: boolean
  jumpWithTenantId?: boolean
}
/**
 *
 * @returns
 */
export const useRouterHooks = () => {
  const router = useRouter()
  const jumpToPage = (params: jumpType) => {
    return new Promise((resolve: any) => {
      const { url, data = {}, isOpen = false } = params

      if (isOpen) {
        window.open(url + obj2UrlString(data))
      } else {
        router
          .push({
            path: url,
            query: {
              ...data,
            },
          })
          .then(() => {
            resolve()
          })
      }
    })
  }

  const jumpBack = () => {
    router.go(-1)
  }

  const reload = () => {
    window.location.reload()
  }

  return {
    jumpToPage,
    jumpBack,
    reload,
  }
}
