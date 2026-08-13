import { createApp } from 'vue'
import App from './App.vue'
import './style.css'
import './App.css'

function showFatal(error: unknown) {
  const message = error instanceof Error ? `${error.message}\n${error.stack ?? ''}` : String(error)
  document.body.innerHTML = `<pre style="white-space:pre-wrap;margin:32px;padding:20px;border:1px solid #ef4444;background:#fff1f2;color:#991b1b;font:14px/1.5 Consolas,monospace">Task Manager 前端启动错误：\n${message.replace(/</g, '&lt;')}</pre>`
}

window.addEventListener('error', event => showFatal(event.error ?? event.message))
window.addEventListener('unhandledrejection', event => showFatal(event.reason))

try {
  const app = createApp(App)
  app.config.errorHandler = showFatal
  app.mount('#app')
} catch (error) {
  showFatal(error)
}
