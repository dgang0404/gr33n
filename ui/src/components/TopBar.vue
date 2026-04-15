<template>
  <header class="bg-gray-900 border-b border-gray-800">
    <div v-if="auth.isDevMode"
      class="bg-amber-900/50 border-b border-amber-700/50 px-6 py-1 text-center">
      <span class="text-amber-300 text-xs font-medium">DEV MODE — auth is disabled, all routes open</span>
    </div>
    <div class="h-14 flex items-center justify-between px-6">
      <h1 class="text-sm font-semibold text-gray-300">{{ title }}</h1>
      <div class="flex items-center gap-3">
        <span :class="apiOk ? 'text-gr33n-400' : 'text-danger'" class="text-xs font-mono">
          {{ apiOk ? '● API online' : '● API offline' }}
        </span>
        <span class="text-xs text-gray-500">{{ now }}</span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../api'

const route = useRoute()
const auth  = useAuthStore()
const apiOk = ref(true)
const now   = ref('')
const labels = {
  '/': 'Dashboard',
  '/zones': 'Zones',
  '/sensors': 'Sensors',
  '/actuators': 'Controls',
  '/schedules': 'Schedules',
  '/tasks': 'Tasks',
  '/fertigation': 'Fertigation',
  '/inventory': 'Inventory',
}
const title = computed(() => {
  if (route.path.startsWith('/zones/')) return 'Zone Details'
  return labels[route.path] ?? 'gr33n'
})

let tick
onMounted(async () => {
  auth.fetchAuthMode()
  tick = setInterval(async () => {
    now.value = new Date().toLocaleTimeString()
    try { await api.get('/health'); apiOk.value = true }
    catch { apiOk.value = false }
  }, 5000)
  now.value = new Date().toLocaleTimeString()
})
onUnmounted(() => clearInterval(tick))
</script>
