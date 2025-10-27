<script setup lang="ts">
import { onMounted, onUnmounted, watch, useTemplateRef } from 'vue'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

gsap.registerPlugin(ScrollTrigger)

interface AnimatedContentProps {
  distance?: number
  direction?: 'vertical' | 'horizontal'
  reverse?: boolean
  duration?: number
  ease?: string | ((progress: number) => number)
  initialOpacity?: number
  animateOpacity?: boolean
  scale?: number
  threshold?: number
  delay?: number
  className?: string
}

const props = withDefaults(defineProps<AnimatedContentProps>(), {
  distance: 100,
  direction: 'vertical',
  reverse: false,
  duration: 0.8,
  ease: 'power3.out',
  initialOpacity: 0,
  animateOpacity: true,
  scale: 1,
  threshold: 0.1,
  delay: 0,
  className: '',
})

const emit = defineEmits<{
  complete: []
}>()

const containerRef = useTemplateRef<HTMLDivElement>('containerRef')

onMounted(() => {
  const el = containerRef.value
  if (!el) return

  const axis = props.direction === 'horizontal' ? 'x' : 'y'
  const offset = props.reverse ? -props.distance : props.distance
  const startPct = (1 - props.threshold) * 100

  gsap.set(el, {
    [axis]: offset,
    scale: props.scale,
    opacity: props.animateOpacity ? props.initialOpacity : 1,
  })

  gsap.to(el, {
    [axis]: 0,
    scale: 1,
    opacity: 1,
    duration: props.duration,
    ease: props.ease,
    delay: props.delay,
    onComplete: () => emit('complete'),
    scrollTrigger: {
      trigger: el,
      start: `top ${startPct}%`,
      toggleActions: 'play none none none',
      once: true,
    },
  })
})

watch(
  () => [
    props.distance,
    props.direction,
    props.reverse,
    props.duration,
    props.ease,
    props.initialOpacity,
    props.animateOpacity,
    props.scale,
    props.threshold,
    props.delay,
  ],
  () => {
    const el = containerRef.value
    if (!el) return

    ScrollTrigger.getAll().forEach((t) => t.kill())
    gsap.killTweensOf(el)

    const axis = props.direction === 'horizontal' ? 'x' : 'y'
    const offset = props.reverse ? -props.distance : props.distance
    const startPct = (1 - props.threshold) * 100

    gsap.set(el, {
      [axis]: offset,
      scale: props.scale,
      opacity: props.animateOpacity ? props.initialOpacity : 1,
    })

    gsap.to(el, {
      [axis]: 0,
      scale: 1,
      opacity: 1,
      duration: props.duration,
      ease: props.ease,
      delay: props.delay,
      onComplete: () => emit('complete'),
      scrollTrigger: {
        trigger: el,
        start: `top ${startPct}%`,
        toggleActions: 'play none none none',
        once: true,
      },
    })
  },
  { deep: true },
)

onUnmounted(() => {
  const el = containerRef.value
  if (el) {
    ScrollTrigger.getAll().forEach((t) => t.kill())
    gsap.killTweensOf(el)
  }
})
</script>

<template>
  <div ref="containerRef" :class="`animated-content ${props.className}`">
    <slot />
  </div>
</template>

<style scoped>
/* GSAP will handle all transforms and opacity */
</style>
