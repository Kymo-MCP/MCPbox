<template>
  <div class="flex h-full">
    <div class="justify-center align-center flex pl-2 link-hover" v-if="isShowBack">
      <el-link link @click="handleBack" class="link-hover" underline="never">
        <el-icon class="mr-2">
          <i class="icon iconfont MCP-fanhui"></i>
        </el-icon>
        {{ t('common.back') }}
      </el-link>
    </div>
    <div class="flex align-center flex-1 justify-end">
      <el-icon
        class="cursor-pointer mr-6 link-hover"
        @click="handleJumpToPage(github, { isOpen: true })"
      >
        <i class="icon iconfont MCP-GitHub"></i>
      </el-icon>
      <el-icon
        v-if="locale === 'zh-cn'"
        class="cursor-pointer mr-6 link-hover"
        @click="handleLanguageChange('en')"
      >
        <i class="icon iconfont MCP-ying"></i>
      </el-icon>
      <el-icon
        v-if="locale === 'en'"
        class="cursor-pointer mr-6 link-hover"
        @click="handleLanguageChange('zh-cn')"
      >
        <i class="icon iconfont MCP-zhong"></i>
      </el-icon>
      <el-icon
        v-if="themeType === 'dark'"
        class="cursor-pointer link-hover"
        @click="handleThemeChange"
      >
        <Sunny />
      </el-icon>
      <el-icon v-else class="cursor-pointer link-hover" @click="handleThemeChange">
        <Moon />
      </el-icon>
      <div class="cursor-pointer ml-6 mr-8">
        <el-dropdown trigger="click" @command="handleJumpToPage" :show-arrow="false">
          <GlareHover
            width="100%"
            height="auto"
            background="transparent"
            border-color="#222222"
            border-radius="20px"
            glare-color="#ffffff"
            :glare-opacity="0.3"
            :glare-size="300"
            :transition-duration="800"
            :play-once="false"
            class="user-avatar"
          >
            <div class="center">
              <el-avatar v-if="userInfo.avatar" :src="userInfo.avatar" fit="cover" :size="28" />
              <el-avatar v-else :icon="UserFilled" :size="28" />
              <span class="ml-2"> {{ userInfo.nickname || 'XXX' }}</span>
            </div>
          </GlareHover>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="/user-center">
                <div class="flex align-center">
                  <div class="user-avatar">
                    <el-avatar
                      v-if="userInfo.avatar"
                      :src="userInfo.avatar"
                      fit="cover"
                      :size="28"
                    />
                    <el-avatar v-else :icon="UserFilled" :size="28" />
                  </div>
                  <div class="user-info">
                    <div class="name">{{ userInfo.nickname || 'XXX' }}</div>
                    <div class="role">超管</div>
                  </div>
                </div>
              </el-dropdown-item>
              <el-dropdown-item divided :command="issues">
                <el-icon>
                  <i class="icon iconfont MCP-fankui"></i>
                </el-icon>
                {{ t('login.issues') }}
              </el-dropdown-item>
              <!-- <el-dropdown-item >
                <el-icon>
                  <i class="icon iconfont MCP-a-aixin_shixin"></i>
                </el-icon>
                {{ t('login.support') }}
              </el-dropdown-item>
              <el-dropdown-item>
                <el-icon>
                  <i class="icon iconfont MCP-weixin"></i>
                </el-icon>
                {{ t('login.weichat') }}
              </el-dropdown-item>
              <el-dropdown-item>
                <el-icon>
                  <i class="icon iconfont MCP-guanyu"></i>
                </el-icon>
                {{ t('login.about') }}
              </el-dropdown-item> -->
              <el-dropdown-item command="/user-center">
                <el-icon>
                  <i class="icon iconfont MCP-shezhi"></i>
                </el-icon>
                {{ t('login.setting') }}
              </el-dropdown-item>
              <el-dropdown-item @click="handleLoginOut" divided>
                <el-icon>
                  <i class="icon iconfont MCP-tuichudenglu"></i>
                </el-icon>
                {{ t('login.logout') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { useSystemStoreHook } from '@/stores/modules/system-store'
import { Sunny, Moon, UserFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores'
import { useRouterHooks } from '@/utils/url'
import GlareHover from '@/components/Animation/GlareHover.vue'

const { t, locale } = useI18n()
const systemStore = useSystemStoreHook()
const { themeType } = toRefs(systemStore)
const { userInfo } = useUserStore()
const { jumpBack, jumpToPage } = useRouterHooks()
const github = 'https://github.com/kymo-mcp/mcpcan'
const issues = 'https://github.com/kymo-mcp/mcpcan/issues'
const route = useRoute()
const router = useRouter()
// condition of show back button
const isShowBack = computed(() => {
  return !route.meta.isMenu
})

// back last class page
const handleBack = () => {
  // If the previous level is not a menu page; Return to Home
  const preRouteMeta = router
    .getRoutes()
    .find((item) => item.path === router.options.history.state.back)?.meta
  if (preRouteMeta?.isMenu) {
    jumpBack()
  } else {
    router.push('/')
  }
}

// jump to the sub page
const handleJumpToPage = (url: string, options?: any) => {
  if (url === issues) {
    Object.assign(options, { isOpen: true })
  }
  jumpToPage({
    url: url,
    data: {},
    ...options,
  })
}

/**
 * Handle change language
 *
 * @param lang  语言（zh-cn、en）
 */
function handleLanguageChange(lang: string) {
  systemStore.changeLanguage(lang)
}

/**
 * Handle change theme
 *
 */
const handleThemeChange = () => {
  systemStore.changeTheme(systemStore.themeType === 'dark' ? 'light' : 'dark')
}

/**
 * LoginOut
 */
const handleLoginOut = () => {
  useUserStore().logout()
}
</script>

<style lang="scss" scoped>
.align-center {
  align-items: center;
}
.border-none {
  border: none;
}
.user-info {
  margin-left: 12px;
  .name {
    font-size: 16px;
  }
  .role {
    font-size: 12px;
    color: #ccc;
  }
}

.el-menu-demo {
  &.ep-menu--horizontal > .ep-menu-item:nth-child(1) {
    margin-right: auto;
  }
}
:deep(.el-dropdown-menu__item):not(.is-disabled):focus,
:deep(.el-dropdown-menu__item):not(.is-disabled):hover {
  color: var(--ep-purple-color) !important;
  background-color: var(--ep-bg-purple-color) !important;
}
.user-avatar {
  transition: 0.5s;
  &:hover {
    scale: 1.1;
  }
  :deep(.el-icon) {
    margin-right: 0;
  }
}
</style>
