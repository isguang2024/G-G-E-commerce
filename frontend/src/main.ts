import App from './App.vue'
import { createApp } from 'vue'
import { initStore } from './store'                 // Store
import { initRouter } from './router'               // Router
import language, { loadLocaleMessages } from './locales' // 国际化
import '@styles/core/tailwind.css'                  // tailwind
import '@styles/index.scss'                         // 样式
import '@utils/sys/console.ts'                      // 控制台输出内容
import { setupGlobDirectives } from './directives'
import { setupErrorHandle } from './utils/sys/error-handle'

document.addEventListener(
  'touchstart',
  function () {},
  { passive: false }
)

const app = createApp(App)
initStore(app)
initRouter(app)
setupGlobDirectives(app)
setupErrorHandle(app)

app.use(language)

// 若用户偏好非默认语言，先按需加载语言包再 mount，避免首屏闪烁兜底文案
const localeConfig = language.global.locale
const initialLocale = typeof localeConfig === 'string' ? localeConfig : localeConfig.value
if (initialLocale && initialLocale !== 'zh') {
  loadLocaleMessages(initialLocale).finally(() => app.mount('#app'))
} else {
  app.mount('#app')
}
