<template>
  <div class="flex h-screen overflow-hidden">
    <SideNav />
    <div class="flex-1 flex flex-col overflow-hidden">
      <TopBar />
      <main class="flex-1 overflow-y-auto p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import SideNav from './components/SideNav.vue'
import TopBar  from './components/TopBar.vue'
import { useFarmStore } from './stores/farm'
import { useFarmContextStore } from './stores/farmContext'
import { onMounted, onUnmounted, watch } from 'vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()
let evtSource = null

function connectSSE(farmId) {
  if (evtSource) evtSource.close()
  if (!farmId) return
  const base = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  const token = localStorage.getItem('gr33n_token')
  const url = `${base}/farms/${farmId}/sensors/stream${token ? '?token=' + token : ''}`
  evtSource = new EventSource(url)
  evtSource.addEventListener('readings', (e) => {
    try {
      const data = JSON.parse(e.data)
      for (const [id, reading] of Object.entries(data)) {
        store.readings[Number(id)] = reading
      }
    } catch { /* ignore parse errors */ }
  })
  evtSource.onerror = () => {
    evtSource.close()
    setTimeout(() => connectSSE(farmContext.farmId), 5000)
  }
}

watch(() => farmContext.farmId, (id) => {
  if (id) connectSSE(id)
})

onMounted(async () => {
  await farmContext.fetchFarms()
  if (!farmContext.farmId && farmContext.farms.length) {
    await farmContext.selectFarm(farmContext.farms[0].id)
  } else if (farmContext.farmId) {
    await store.loadAll(farmContext.farmId)
  }
  await store.refreshReadings()
  connectSSE(farmContext.farmId)
})
onUnmounted(() => {
  if (evtSource) evtSource.close()
})
</script>
