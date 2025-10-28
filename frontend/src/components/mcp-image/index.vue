<template>
  <el-image
    :src="currentImgUrl"
    :style="{
      width: props.width + 'px',
      height: props.height + 'px',
      'border-radius': props.borderRadius + 'px',
    }"
    :fit="props.fit"
    @error="handleError"
  ></el-image>
</template>

<script lang="ts" setup>
import baseConfig from '@/config/base_config.ts'
import defaultImg from '@/assets/logo.png'

const props = defineProps({
  src: {
    type: String,
  },
  width: {
    type: [String, Number],
    default: 134,
  },
  height: {
    type: [String, Number],
    default: 134,
  },
  borderRadius: {
    type: [String, Number],
    default: 8,
  },
  fit: {
    type: String,
    default: 'cover',
  },
})

const originImgUrl = computed(() => {
  return props.src ? baseConfig.SERVER_BASE_URL + props.src : ''
})

const currentImgUrl = ref(originImgUrl.value)

const handleError = () => {
  if (currentImgUrl.value === defaultImg) return
  currentImgUrl.value = defaultImg
}
</script>
