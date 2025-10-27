<template>
  <div class="tags-container">
    <el-scrollbar
      ref="scrollbarRef"
      class="scroll-container"
      :view-style="{ height: '100%' }"
      @wheel="handleScroll"
    >
      <div class="h-full flex justify-start align-center" gap-8px>
        <el-tag
          v-for="tag in tagsViewStore.visitedViews"
          :key="tag.fullPath"
          h-26px
          cursor-pointer
          :closable="!tag.affix"
          :effect="tagsViewStore.isActive(tag) ? 'dark' : 'light'"
          :type="tagsViewStore.isActive(tag) ? 'primary' : 'info'"
          @close="closeSelectedTag(tag)"
          @click="
            router.push({
              path: tag.fullPath,
              query: tag.query,
            })
          "
        >
          {{ translateRouteTitle(tag.title) }}
        </el-tag>
      </div>
    </el-scrollbar>
  </div>
</template>
<script setup lang="ts">
import { useTagsViewStore, type TagView } from '@/stores'
import { type RouteRecordRaw } from 'vue-router'
import { resolve } from 'path-browserify'
import i18n from '@/lang/index'

const router = useRouter()
const route = useRoute()
const tagsViewStore = useTagsViewStore()

function translateRouteTitle(title: any) {
  const hasKey = i18n.global.te('sideMenu.' + title)
  if (hasKey) {
    const translatedTitle = i18n.global.t('sideMenu.' + title)
    return translatedTitle
  }
  return title
}

watch(
  route,
  () => {
    if (!route.meta?.title) return
    tagsViewStore.addVisitedView({
      name: route.name as string,
      title: route.meta.title as string,
      path: route.path,
      fullPath: route.fullPath,
      affix: (route.meta.affix as boolean) || false,
      keepAlive: (route.meta.keepAlive as boolean) || false,
      query: route.query,
    })
  },
  { immediate: true },
)

/**
 * 水平滚动事件
 */
const handleScroll = () => {}

/**
 * 关闭标签
 */
const closeSelectedTag = (tag: TagView | null) => {
  if (!tag) return
  tagsViewStore.delView(tag)
}

/**
 * 递归提取固定标签
 */
const extractAffixTags = (routes: RouteRecordRaw[], basePath = '/'): TagView[] => {
  const affixTags: TagView[] = []

  const traverse = (routeList: RouteRecordRaw[], currentBasePath: string) => {
    routeList.forEach((route) => {
      const fullPath = resolve(currentBasePath, route.path)
      // 如果是固定标签，添加到列表
      if (route.meta?.affix) {
        affixTags.push({
          path: fullPath,
          fullPath,
          name: String(route.name || ''),
          title: (route.meta.title as string) || 'no-name',
          affix: true,
          keepAlive: (route.meta.keepAlive as boolean) || false,
        })
      }
    })
  }

  traverse(routes, basePath)
  return affixTags
}

/**
 * 初始化固定标签
 */
const initAffixTags = () => {
  const affixTags = extractAffixTags(router.getRoutes())
  affixTags.forEach((tag) => {
    if (tag.name) {
      tagsViewStore.addVisitedView(tag)
    }
  })
}

// 初始化
onMounted(() => {
  initAffixTags()
})
</script>
<style lang="scss" scoped>
.tags-container {
  // width: 100%;
  height: 34px;
  padding: 0 15px;
  background-color: #292727;
  .justify-start {
    justify-content: start;
  }
  .align-center {
    align-items: center;
  }
  .scroll-container {
    white-space: nowrap;
  }
}
</style>
