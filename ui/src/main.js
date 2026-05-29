import { createApp } from 'vue'
import { Capacitor } from '@capacitor/core'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './style.css'
import { setUnauthorizedHandler } from './api/index.js'
import { useAuthStore } from './stores/auth'
import { useCapabilitiesStore } from './stores/capabilities'
import { useGuardianPanelStore } from './stores/guardianPanel'

const pinia = createPinia()
const app = createApp(App).use(pinia).use(router)

router.afterEach((to) => {
  useGuardianPanelStore().setRouteFromRouter(to)
})

setUnauthorizedHandler(() => {
  useAuthStore().logout()
})

app.mount('#app')

useCapabilitiesStore().fetch().catch(() => { /* surfaced via fetchError */ })

if (Capacitor.isNativePlatform()) {
  import('@capacitor/app').then(({ App }) => {
    App.addListener('backButton', ({ canGoBack }) => {
      if (canGoBack) window.history.back()
      else App.exitApp()
    })
  }).catch(() => {})
}

// Service worker is now managed by vite-plugin-pwa (auto-registered).
