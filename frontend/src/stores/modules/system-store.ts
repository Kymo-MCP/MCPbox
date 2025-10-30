import { defineStore } from 'pinia'
import { useStorage } from '@vueuse/core'
import { store } from '@/stores'
import { useI18n } from 'vue-i18n'
// import { useDark, useToggle } from '@vueuse/core'

export const useSystemStore = defineStore('system', () => {
  const { locale } = useI18n()
  // current language zh-cn/en
  const language = useStorage('language', 'zh-cn')
  /**
   * Handle change language
   *
   * @param val
   */
  function changeLanguage(val: string) {
    const root = document.documentElement
    animationPage(root)
    document
      .startViewTransition(() => {})
      .ready.then(() => {
        language.value = val
        locale.value = val
      })
  }

  // Init the theme
  const themeType = useStorage('theme', 'dark')
  /**
   * Handle change theme
   *
   * @param val
   */
  function changeTheme(val: string) {
    themeType.value = val
    const root = document.documentElement
    animation3D(root)
    // add a animation
    document.startViewTransition(() => {
      root.className = themeType.value
    })
  }

  // menu open status
  const isCollapse = useStorage('isCollapse', false)
  // change menu status
  function changeCollapse(status: boolean) {
    isCollapse.value = !status
  }

  // use 3D animation
  function animation3D(el: HTMLElement) {
    el.style.setProperty('--transition-animation-old', 'reveal-out')
    el.style.setProperty('--transition-animation-new', 'reveal-in')
    el.style.setProperty(
      '--page-ease',
      `linear(
        0,
        0.324 9.1%,
        0.584 18.6%,
        0.782 28.6%,
        0.858 33.8%,
        0.92 39.2%,
        0.997 49.5%,
        1.021 55.1%,
        1.033 61%,
        1.035 71.7%,
        1
      )`,
    )
  }
  // use page animation
  function animationPage(el: HTMLElement) {
    el.style.setProperty('--transition-animation-old', 'moveOut')
    el.style.setProperty('--transition-animation-new', 'moveIn')
    el.style.setProperty(
      '--page-ease',
      `linear(
        0,
        0.223 11.7%,
        0.392 18.4%,
        0.619 24.8%,
        0.999 33.3%,
        0.748 40%,
        0.691 42.7%,
        0.672 45.3%,
        0.69 47.8%,
        0.743 50.4%,
        0.999 57.7%,
        0.883 61.8%,
        0.856 63.6%,
        0.848 65.3%,
        0.855 67%,
        0.879 68.8%,
        0.999 74.5%,
        0.953 77.5%,
        0.94 80.2%,
        0.95 82.7%,
        1 88.2%,
        0.987 91.9%,
        1
      )`,
    )
  }

  return {
    language,
    changeLanguage,
    themeType,
    changeTheme,
    isCollapse,
    changeCollapse,
  }
})

export function useSystemStoreHook() {
  return useSystemStore(store)
}
