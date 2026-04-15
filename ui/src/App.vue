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
import { onMounted, onUnmounted } from 'vue'

const store = useFarmStore()
let evtSource = null

function connectSSE() {
  const base = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
  const token = localStorage.getItem('gr33n_token')
  const url = `${base}/farms/1/sensors/stream${token ? '?token=' + token : ''}`
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
    setTimeout(connectSSE, 5000)
  }
}

onMounted(async () => {
  await store.loadAll(1)
  await store.refreshReadings()
  connectSSE()
})
onUnmounted(() => {
  if (evtSource) evtSource.close()
})
</script>
