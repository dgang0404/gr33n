<template>
  <div class="flex h-screen overflow-hidden">
    <SideNav class="hidden md:flex" />

    <div class="flex-1 flex flex-col overflow-hidden">
      <TopBar @toggle-drawer="drawerOpen = !drawerOpen" />
      <main class="flex-1 overflow-y-auto p-3 sm:p-6 pb-20 md:pb-6">
        <RouterView />
      </main>
    </div>

    <!-- Mobile bottom nav -->
    <nav class="md:hidden fixed bottom-0 inset-x-0 bg-gray-900 border-t border-gray-800 flex justify-around py-2 z-40"
         style="padding-bottom: max(0.5rem, env(safe-area-inset-bottom))">
      <RouterLink v-for="item in mobileNav" :key="item.to" :to="item.to"
        class="flex flex-col items-center gap-0.5 text-gray-500 text-[10px]"
        active-class="text-green-400">
        <span class="text-base">{{ item.icon }}</span>
        {{ item.label }}
      </RouterLink>
    </nav>

    <!-- Mobile drawer overlay -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="drawerOpen" class="md:hidden fixed inset-0 z-50 flex">
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="-translate-x-full"
          enter-to-class="translate-x-0"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="translate-x-0"
          leave-to-class="-translate-x-full"
          appear
        >
          <aside class="w-64 bg-gray-900 border-r border-gray-800 flex flex-col overflow-y-auto">
            <div class="px-4 py-4 flex items-center justify-between border-b border-gray-800">
              <span class="text-gr33n-400 text-2xl font-bold tracking-tight">gr33n</span>
              <button @click="drawerOpen = false" class="p-1 text-gray-400 hover:text-white">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
            <nav class="flex-1 px-3 py-4 space-y-4">
              <div v-for="group in drawerNavGroups" :key="group.label">
                <p class="px-3 mb-1 text-[10px] uppercase tracking-widest text-gray-600 font-semibold">{{ group.label }}</p>
                <div class="space-y-0.5">
                  <RouterLink
                    v-for="item in group.items" :key="item.to" :to="item.to"
                    @click="drawerOpen = false"
                    class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
                    active-class="bg-gr33n-900 text-gr33n-400 font-semibold"
                  >
                    <span class="text-lg">{{ item.icon }}</span>
                    {{ item.label }}
                  </RouterLink>
                </div>
              </div>
            </nav>
          </aside>
        </Transition>
        <div class="flex-1 bg-black/60" @click="drawerOpen = false" />
      </div>
    </Transition>
  </div>
</template>

<script setup>
import SideNav from './components/SideNav.vue'
import TopBar  from './components/TopBar.vue'
import { useFarmStore } from './stores/farm'
import { useFarmContextStore } from './stores/farmContext'
import { useAuthStore } from './stores/auth'
import { usePush } from './composables/usePush'
import { onMounted, onUnmounted, ref, watch } from 'vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const auth = useAuthStore()
const push = usePush()
let evtSource = null

const drawerOpen = ref(false)

const mobileNav = [
  { to: '/',           icon: '🌿', label: 'Home' },
  { to: '/tasks',      icon: '✅', label: 'Tasks' },
  { to: '/zones',      icon: '🗂️', label: 'Zones' },
  { to: '/alerts',     icon: '🔔', label: 'Alerts' },
  { to: '/settings',   icon: '⚙️', label: 'More' },
]

const drawerNavGroups = [
  {
    label: 'Operate',
    items: [
      { to: '/',          icon: '🌿', label: 'Dashboard' },
      { to: '/tasks',     icon: '✅', label: 'Tasks' },
      { to: '/schedules', icon: '📅', label: 'Schedules' },
      { to: '/actuators', icon: '⚡', label: 'Controls' },
      { to: '/sensors',   icon: '📡', label: 'Sensors' },
    ],
  },
  {
    label: 'Grow',
    items: [
      { to: '/zones',       icon: '🗂️', label: 'Zones' },
      { to: '/plants',      icon: '🌱', label: 'Plants' },
      { to: '/fertigation', icon: '💧', label: 'Fertigation' },
      { to: '/inventory',   icon: '🧪', label: 'Inventory' },
    ],
  },
  {
    label: 'Monitor',
    items: [
      { to: '/alerts', icon: '🔔', label: 'Alerts' },
      { to: '/costs',  icon: '💰', label: 'Costs' },
    ],
  },
  {
    label: 'System',
    items: [
      { to: '/catalog',  icon: '📚', label: 'Catalog' },
      { to: '/settings', icon: '⚙️', label: 'Settings' },
    ],
  },
]

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

async function bootstrapFarmData() {
  const token = localStorage.getItem('gr33n_token')
  if (!token) return

  await farmContext.fetchFarms()
  if (!farmContext.farmId && farmContext.farms.length) {
    await farmContext.selectFarm(farmContext.farms[0].id)
  } else if (farmContext.farmId) {
    await store.loadAll(farmContext.farmId)
  }
  await store.refreshReadings()
  connectSSE(farmContext.farmId)
}

onMounted(() => {
  bootstrapFarmData()
  if (localStorage.getItem('gr33n_token')) push.init()
})

watch(
  () => auth.token,
  (t) => {
    if (t) bootstrapFarmData()
  }
)

onUnmounted(() => {
  if (evtSource) evtSource.close()
})
</script>
