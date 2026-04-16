import { createApp } from 'vue'
import { Capacitor } from '@capacitor/core'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './style.css'
import { setUnauthorizedHandler } from './api/index.js'
import { useAuthStore } from './stores/auth'

const pinia = createPinia()
const app = createApp(App).use(pinia).use(router)

setUnauthorizedHandler(() => {
  useAuthStore().logout()
})

app.mount('#app')

if (Capacitor.isNativePlatform()) {
  import('@capacitor/app').then(({ App }) => {
    App.addListener('backButton', ({ canGoBack }) => {
      if (canGoBack) window.history.back()
      else App.exitApp()
    })
  }).catch(() => {})
}

if (import.meta.env.PROD && 'serviceWorker' in navigator) {
  navigator.serviceWorker.register(`${import.meta.env.BASE_URL}sw.js`).catch(() => {})
}
