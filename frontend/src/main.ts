import { createApp } from 'vue'
import { createPinia } from 'pinia'
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/700.css'
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import App from './App.vue'
import router from './router'
import './assets/styles/main.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// Global error handler — catches unhandled component errors
app.config.errorHandler = (err, _instance, info) => {
  console.error(`[Heimdall] Unhandled error (${info}):`, err)
  import('./composables/useToast').then(({ useToast }) => {
    useToast().show('An unexpected error occurred', 'error')
  })
}

// Catch unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  console.error('[Heimdall] Unhandled rejection:', event.reason)
  import('./composables/useToast').then(({ useToast }) => {
    useToast().show('An unexpected error occurred', 'error')
  })
})

app.mount('#app')
