import { fileURLToPath, URL } from 'node:url'
import path from 'node:path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import Unocss from 'unocss/vite'
import AutoImport from 'unplugin-auto-import/vite' // 自动根据需求导入vue的相关API如；ref、reactive等

// https://vite.dev/config/
export default defineConfig({
  server: {
    open: true,
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api/authz': {
        target: 'http://localhost:8082',
        changeOrigin: true,
        rewrite: (path: string) => path.replace(/^\/api/, ''),
      },
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        rewrite: (path: string) => path.replace(/^\/api/, ''),
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "~/styles/element/index.scss" as *;`,
        // api: 'modern-compiler',
      },
    },
  },
  plugins: [
    vue(),
    Unocss(),
    AutoImport({
      // 1. 自动导入 Vue 的 Composition API
      imports: ['vue', '@vueuse/core', 'pinia', 'vue-router', 'vue-i18n'],
      // 禁用默认导入（若有冲突）
      defaultExportByFilename: false,
      // 会自动导入 vue 中的 ref、reactive、onMounted 等所有 API
      // 2. 生成声明文件（可选，解决 TypeScript 类型提示问题）
      dts: 'src/auto-imports.d.ts', // 生成的声明文件路径，需手动创建空文件
      // 3. 全局注册的组件（可选，若需自动导入组件）
      // components: ['vue'],
    }),
  ],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '~/': `${path.resolve(__dirname, 'src')}/`,
    },
  },
})
