import { store } from '@/stores'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

/**
 * 页签对象
 */
export interface TagView {
  /** 页签名称 */
  name: string
  /** 页签标题 */
  title: string
  /** 页签路由路径 */
  path: string
  /** 页签路由完整路径 */
  fullPath: string
  /** 页签图标 */
  icon?: string
  /** 是否固定页签 */
  affix?: boolean
  /** 是否开启缓存 */
  keepAlive?: boolean
  /** 路由查询参数 */
  query?: any
}

export const useTagsViewStore = defineStore('tagsView', () => {
  const route = useRoute()
  const router = useRouter()
  // 访问过的页面列表
  const visitedViews = ref<TagView[]>([])

  /**
   * 添加已访问视图到已访问视图列表中
   */
  function addVisitedView(view: TagView) {
    // 如果已经存在于已访问的视图列表中或者是重定向地址，则不再添加
    if (view.path.startsWith('/redirect') || visitedViews.value.some((v) => v.path === view.path)) {
      return
    }
    // 如果视图是固定的（affix），则在已访问的视图列表的开头添加
    if (view.affix) {
      visitedViews.value.unshift(view)
    } else {
      // 如果视图不是固定的，则在已访问的视图列表的末尾添加
      visitedViews.value.push(view)
    }
  }

  /**
   * 判断当前路径是否为当前标签
   * @param tag - 标签
   * @returns {boolean} - 是否为当前标签
   */
  function isActive(tag: TagView) {
    return tag.path === (route && route.path)
  }

  function delView(tag: TagView) {
    visitedViews.value = visitedViews.value.filter((tagView) => {
      return tagView.path !== tag.path
    })
    router.push('/')
  }

  return {
    visitedViews,
    addVisitedView,
    isActive,
    delView,
  }
})

export function useViewStoreHook() {
  return useTagsViewStore(store)
}
