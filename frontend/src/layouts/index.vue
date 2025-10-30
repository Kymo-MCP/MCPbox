<template>
  <div class="common-layout">
    <el-container :key="locale">
      <BaseSide v-model="isCollapse" />
      <el-container>
        <el-header class="p-0">
          <BaseHeader />
        </el-header>
        <!-- <el-header class="p-0 tag-view">
            <TagsView />
          </el-header> -->
        <el-main :class="route.path === '/home' ? 'p-0' : ''" class="hide-scrollbar">
          <!-- when change language then refresh page -->
          <el-config-provider :locale="elLocale">
            <RouterView></RouterView>
          </el-config-provider>
        </el-main>
      </el-container>
    </el-container>
  </div>
  <SplashCursor
    :SIM_RESOLUTION="128"
    :DYE_RESOLUTION="1440"
    :CAPTURE_RESOLUTION="512"
    :DENSITY_DISSIPATION="10.5"
    :VELOCITY_DISSIPATION="2"
    :PRESSURE="0.1"
    :PRESSURE_ITERATIONS="20"
    :CURL="3"
    :SPLAT_RADIUS="0.2"
    :SPLAT_FORCE="6000"
    :SHADING="true"
    :COLOR_UPDATE_SPEED="10"
    :BACK_COLOR="{ r: 0.5, g: 0, b: 0 }"
    :TRANSPARENT="true"
  />
</template>
<script setup lang="ts">
import BaseHeader from './BaseHeader/index.vue'
import BaseSide from './BaseSide/index.vue'
// import TagsView from './TagsView/index.vue'
import { useSystemStoreHook } from '@/stores'
import SplashCursor from '@/components/Animation/SplashCursor.vue'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

const { locale } = useI18n()
const route = useRoute()
const { isCollapse, language } = toRefs(useSystemStoreHook())
const elLocale = computed(() => (language.value === 'zh-cn' ? zhCn : en))

onMounted(() => {})
</script>
<style lang="scss" scoped>
.common-layout {
  width: 100vm;
  height: 100vh;
  .el-header {
    border-bottom: 1px solid var(--el-menu-border-color);
    &.tag-view {
      padding: 0;
      border: 0;
      height: auto;
    }
  }
  .el-container {
    height: 100%;
  }
  .el-main {
    height: calc(100vh - 94px);
    background: var(--ep-bg-color);
    &.p-0 {
      padding: 0 !important;
    }
  }
}
</style>
