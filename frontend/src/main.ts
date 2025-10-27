import { createApp } from 'vue'
// import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import App from './App.vue'
import router from './router'
import i18n from './lang'
import 'uno.css'
import 'element-plus/dist/index.css'
import '~/styles/index.scss'
// import '~/styles/element/index.scss'
// If you want to use ElMessage, import it.
import 'element-plus/theme-chalk/dark/css-vars.css'
// import 'element-plus/theme-chalk/src/message.scss'
// import 'element-plus/theme-chalk/src/message-box.scss'
import '~/assets/icon/iconfont.css'
import countTo from './directives/numberScrollDirective.ts'
import { store } from './stores'

const app = createApp(App)
app.directive('countTo', countTo)

// app.use(createPinia())
app.use(i18n)
app.use(router)
app.use(ElementPlus)
app.use(store)

app.mount('#app')
