<template>
  <GlareHover
    width="auto"
    height="auto"
    background="transparent"
    border-color="#222222"
    border-radius="4px"
    glare-color="#ffffff"
    :glare-opacity="0.3"
    :glare-size="300"
    :transition-duration="800"
    :play-once="false"
  >
    <el-button
      type="success"
      :icon="icon"
      class="base-btn"
      :class="class"
      :loading="loading"
      @click="handleClick"
      ><slot></slot
    ></el-button>
  </GlareHover>
</template>

<script setup lang="ts">
import GlareHover from '../Animation/GlareHover.vue'
import { type Component } from 'vue'

defineProps({
  icon: {
    type: Object as () => Component,
    required: false,
  },
  class: {
    type: String,
    required: false,
  },
  loading: {
    type: Boolean,
    required: false,
  },
})
const emit = defineEmits<{
  (e: 'click'): void
}>()
const handleClick = () => {
  emit('click')
}
</script>
<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
