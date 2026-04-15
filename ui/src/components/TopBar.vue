<template>
  <header class="bg-gray-900 border-b border-gray-800">
    <div v-if="auth.isDevMode"
      class="bg-amber-900/50 border-b border-amber-700/50 px-6 py-1 text-center">
      <span class="text-amber-300 text-xs font-medium">DEV MODE — auth is disabled, all routes open</span>
    </div>
    <div v-else-if="auth.isAuthTestMode"
      class="bg-violet-950/60 border-b border-violet-600/40 px-6 py-1 text-center">
      <span class="text-violet-200 text-xs font-medium">AUTH TEST — JWT and API key enforced like production (local dev binary only)</span>
    </div>
    <div v-if="farmContext.farmSelectionNotice"
      class="bg-sky-950/60 border-b border-sky-700/40 px-6 py-1">
      <div class="flex items-center justify-between gap-3">
        <span class="text-sky-200 text-xs font-medium">{{ farmContext.farmSelectionNotice }}</span>
        <button
          class="text-sky-300 text-xs hover:text-sky-100 transition-colors"
          @click="farmContext.clearFarmSelectionNotice()"
        >
          Dismiss
        </button>
      </div>
    </div>
    <div class="h-14 flex items-center justify-between px-6">
      <h1 class="text-sm font-semibold text-gray-300">{{ title }}</h1>
      <div class="flex items-center gap-4">
        <RouterLink to="/alerts" class="relative text-gray-400 hover:text-white transition-colors" title="Alerts">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
            <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
          </svg>
          <span v-if="farmStore.unreadAlertCount > 0"
            class="absolute -top-1.5 -right-1.5 bg-red-500 text-white text-[10px] font-bold rounded-full w-4 h-4 flex items-center justify-center">
            {{ farmStore.unreadAlertCount > 9 ? '9+' : farmStore.unreadAlertCount }}
          </span>
        </RouterLink>
        <span :class="apiOk ? 'text-gr33n-400' : 'text-danger'" class="text-xs font-mono">
          {{ apiOk ? '● API online' : '● API offline' }}
        </span>
        <span class="text-xs text-gray-500">{{ now }}</span>
        <span v-if="auth.username" class="text-xs text-gray-500">{{ auth.username }}</span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import api from '../api'

const route = useRoute()
const auth  = useAuthStore()
const farmStore = useFarmStore()
const farmContext = useFarmContextStore()
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
  '/alerts': 'Alerts',
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
    if (farmContext.farmId) {
      try { await farmStore.countUnreadAlerts(farmContext.farmId) } catch {}
    }
  }, 5000)
  now.value = new Date().toLocaleTimeString()
  if (farmContext.farmId) {
    try { await farmStore.countUnreadAlerts(farmContext.farmId) } catch {}
  }
})
onUnmounted(() => clearInterval(tick))
</script>
