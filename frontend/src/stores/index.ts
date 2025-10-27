import type { App } from 'vue'
import { createPinia } from 'pinia'

const store = createPinia()

// 全局注册 store
export function setupStore(app: App<Element>) {
  app.use(store)
}

export * from './modules/system-store'
export * from './modules/user-store'
export * from './modules/tags-view-store'
export * from './modules/mcp-store'

export { store }
