<template>
  <canvas id="cobe" class="canvas-bg"></canvas>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, toRefs, watch } from 'vue'
import createGlobe from 'cobe'
import { useSystemStoreHook } from '@/stores/modules/system-store'

const systemStore = useSystemStoreHook()
const { themeType } = toRefs(systemStore)

const globe = ref()
const phi = ref(0)

// create a instance
const createGlobeInstance = () => {
  if (globe.value) {
    globe.value.destroy() // destroy the instance
    globe.value = null
  }
  const canvas = document.getElementById('cobe') as HTMLCanvasElement
  // Dynamically set Canvas size to the actual screen size (considering device pixel ratio)
  const setCanvasSize = () => {
    const dpr = window.devicePixelRatio || 2
    let width = window.innerWidth * dpr * 0.9
    let height = window.innerHeight * dpr * 0.9
    const main = document.getElementById('el-main-home')
    if (main) {
      width = main?.clientWidth * dpr * 0.9
      height = main?.clientHeight * dpr * 0.9
    }

    canvas.width = width
    canvas.height = height
    // Synchronize CSS size (ensure display size matches screen size)
    canvas.style.width = `${main?.clientWidth}px`
    canvas.style.height = `${main?.clientHeight}px`
    return { width, height, dpr }
  }
  // init size
  const { width, height, dpr } = setCanvasSize()
  const dark = themeType.value === 'dark' ? 1 : 0
  const baseColor: [number, number, number] =
    themeType.value === 'dark' ? [0.3, 0.3, 0.3] : [0.96, 0.95, 0.95]
  const glowColor: [number, number, number] =
    themeType.value === 'dark' ? [1, 1, 1] : [0.6, 0.6, 0.6]
  globe.value = createGlobe(canvas, {
    devicePixelRatio: dpr,
    width,
    height,
    phi: 0,
    theta: 0,
    dark,
    diffuse: 1.2,
    mapSamples: 16000,
    mapBrightness: 6,
    baseColor,
    markerColor: [0.1, 0.8, 1],
    glowColor,
    offset: [width / 9, -100],
    markers: [
      // longitude latitude
      // { location: [37.7595, -122.4367], size: 0.03 },
      // { location: [40.7128, -74.006], size: 0.1 },
    ],
    onRender: (state) => {
      // Called on every animation frame.
      // `state` will be an empty object, return updated params.
      state.phi = phi.value
      phi.value += 0.01
    },
  })

  window.addEventListener('resize', () => {})
}

watch(
  () => themeType.value,
  () => {
    setTimeout(createGlobeInstance, 350)
  },
)

// 初始化
const init = () => {
  createGlobeInstance()
  window.addEventListener('resize', () => {
    createGlobeInstance()
  })
}
onBeforeUnmount(() => {
  globe.value?.destroy()
})
onMounted(init)
</script>

<style lang="scss" scoped>
.canvas-bg {
  position: absolute;
  top: 0;
  left: 0;
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
  // z-index: -1;
}
</style>
