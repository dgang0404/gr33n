import { createApp } from 'vue'
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
