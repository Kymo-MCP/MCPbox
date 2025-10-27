<template>
  <el-image
    :src="logo"
    :style="{
      width: width + 'px',
      height: height + 'px',
    }"
    @click="handleToHome"
  ></el-image>
</template>

<script setup lang="ts">
import logo from '@/assets/logo.png'
import { useRouterHooks } from '@/utils/url'

const { jumpToPage } = useRouterHooks()

defineProps({
  width: {
    type: [String, Number],
    default: 40,
  },
  height: {
    type: [String, Number],
    default: 40,
  },
})
const handleToHome = () => {
  jumpToPage({
    url: '/home',
    data: {},
  })
}
</script>
