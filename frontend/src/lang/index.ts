import { createI18n } from 'vue-i18n'
import { Storage } from '@/utils/storage'

// global
import enGlobal from './global/en.json'
import zhCnGlobal from './global/zh-cn.json'

// home
import enHome from './home/en.json'
import zhCnHome from './home/zh-cn.json'
// mcp instance and template
import enMcpManage from './mcp/en.json'
import zhMcpManage from './mcp/zh-cn.json'
// env manage
import enEnvManage from './env/en.json'
import zhEnvManage from './env/zh-cn.json'
// code page
import enCodePackage from './code-package/en.json'
import zhCodePackage from './code-package/zh-cn.json'

const messages = {
  'zh-cn': { ...zhCnGlobal, ...zhCnHome, ...zhMcpManage, ...zhEnvManage, ...zhCodePackage },
  en: { ...enGlobal, ...enHome, ...enMcpManage, ...enEnvManage, ...enCodePackage },
}

const i18n = createI18n({
  legacy: false,
  locale: Storage.get('language'),
  messages,
  globalInjection: true,
})

export default i18n
