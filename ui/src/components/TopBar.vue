<template>
  <header class="h-14 bg-gray-900 border-b border-gray-800 flex items-center justify-between px-6">
    <h1 class="text-sm font-semibold text-gray-300">{{ title }}</h1>
    <div class="flex items-center gap-3">
      <span :class="apiOk ? 'text-gr33n-400' : 'text-danger'" class="text-xs font-mono">
        {{ apiOk ? '● API online' : '● API offline' }}
      </span>
      <span class="text-xs text-gray-500">{{ now }}</span>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'

const route  = useRoute()
const apiOk  = ref(true)
const now    = ref('')
const labels = {
  '/': 'Dashboard',
  '/zones': 'Zones',
  '/sensors': 'Sensors',
  '/actuators': 'Controls',
  '/schedules': 'Schedules',
  '/tasks': 'Tasks',
  '/inventory': 'Inventory',
}
const title = computed(() => {
  if (route.path.startsWith('/zones/')) return 'Zone Details'
  return labels[route.path] ?? 'gr33n'
})

let tick
onMounted(() => {
  tick = setInterval(async () => {
    now.value = new Date().toLocaleTimeString()
    try { await api.get('/health'); apiOk.value = true }
    catch { apiOk.value = false }
  }, 5000)
  now.value = new Date().toLocaleTimeString()
})
onUnmounted(() => clearInterval(tick))
</script>
