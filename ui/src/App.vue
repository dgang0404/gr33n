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
let interval

onMounted(async () => {
  await store.loadAll(1)
  await store.refreshReadings()
  interval = setInterval(() => store.refreshReadings(), 30_000)
})
onUnmounted(() => clearInterval(interval))
</script>
