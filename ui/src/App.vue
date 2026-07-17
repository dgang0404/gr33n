<template>
  <div class="flex h-screen overflow-hidden">
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:fixed focus:top-3 focus:left-3 focus:z-[100] focus:px-4 focus:py-2 focus:rounded-lg focus:bg-green-700 focus:text-white focus:text-sm focus:font-medium"
    >
      Skip to main content
    </a>
    <SideNav class="hidden md:flex" />

    <div class="flex-1 flex flex-col overflow-hidden">
      <TopBar @toggle-drawer="drawerOpen = !drawerOpen" />
      <main
        id="main-content"
        tabindex="-1"
        :class="mainClass"
      >
        <div :class="routeShellClass">
          <RouterView />
        </div>
      </main>
    </div>

    <!-- Mobile bottom nav -->
    <nav
      class="md:hidden fixed bottom-0 inset-x-0 bg-gray-900 border-t border-gray-800 flex justify-around py-2 z-40"
      aria-label="Main navigation"
      style="padding-bottom: max(0.5rem, env(safe-area-inset-bottom))"
    >
      <RouterLink
        v-for="item in mobileNav"
        :key="item.to"
        :to="item.to"
        class="flex flex-col items-center justify-center gap-0.5 text-gray-500 text-[10px] min-h-[44px] min-w-[44px] px-2"
        active-class="text-green-400"
        :aria-current="isMobileNavCurrent(item.to) ? 'page' : undefined"
      >
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
          <aside
            ref="mobileDrawerRef"
            role="dialog"
            aria-modal="true"
            aria-label="Navigation menu"
            class="w-64 bg-gray-900 border-r border-gray-800 flex flex-col overflow-y-auto"
          >
            <div class="px-4 py-4 flex items-center justify-between border-b border-gray-800">
              <span class="text-gr33n-400 text-2xl font-bold tracking-tight">gr33n</span>
              <button
                type="button"
                class="p-1 text-gray-400 hover:text-white"
                aria-label="Close navigation menu"
                @click="drawerOpen = false"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            </div>
            <div class="px-3 pt-3 shrink-0">
              <GuardianNavLaunch :collapsed="false" />
            </div>
            <nav class="flex-1 px-3 py-4 space-y-4 overflow-y-auto">
              <div v-for="group in drawerNavGroups" :key="group.label">
                <p class="px-3 mb-1 text-[10px] uppercase tracking-widest text-gray-600 font-semibold">{{ group.label }}</p>
                <div class="space-y-0.5">
                  <RouterLink
                    v-for="item in group.items" :key="item.to" :to="item.to"
                    @click="drawerOpen = false"
                    class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
                    active-class="bg-gr33n-900 text-gr33n-400 font-semibold"
                    :aria-current="isMobileNavCurrent(item.to) ? 'page' : undefined"
                  >
                    <span class="text-lg">{{ item.icon }}</span>
                    {{ item.label }}
                  </RouterLink>
                </div>
              </div>
            </nav>
          </aside>
        </Transition>
        <div class="flex-1 bg-black/60" aria-hidden="true" @click="drawerOpen = false" />
      </div>
    </Transition>

    <!-- Farm Guardian drawer (opened from TopBar, sidebar, or mobile nav) -->
    <GuardianDrawer v-if="auth.token" />
  </div>
</template>

<script setup>
import SideNav from './components/SideNav.vue'
import TopBar  from './components/TopBar.vue'
import GuardianDrawer from './components/GuardianDrawer.vue'
import GuardianNavLaunch from './components/GuardianNavLaunch.vue'
import { useFarmStore } from './stores/farm'
import { useFarmContextStore } from './stores/farmContext'
import { useGuardianPanelStore } from './stores/guardianPanel'
import { useAuthStore } from './stores/auth'
import { usePush } from './composables/usePush'
import { useDialogFocusTrap } from './composables/useDialogFocusTrap'
import { onMounted, onUnmounted, ref, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { buildNavGroups, mobileBottomNav } from './lib/navGroups.js'
import { moduleMapFromRows } from './lib/farmModules.js'
import { workspaceByRoute } from './lib/workspaces.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const guardianPanel = useGuardianPanelStore()
const auth = useAuthStore()
const push = usePush()
let evtSource = null

const drawerOpen = ref(false)
const mobileDrawerRef = ref(null)
const route = useRoute()

/** Workspace shells scroll internally — chrome stays pinned below TopBar. */
const isWorkspaceRoute = computed(() => !!workspaceByRoute(route.path))
const mainClass = computed(() =>
  isWorkspaceRoute.value
    ? 'flex-1 min-h-0 overflow-hidden flex flex-col'
    : 'flex-1 min-h-0 overflow-y-auto p-3 sm:p-6 pb-20 md:pb-6',
)
const routeShellClass = computed(() =>
  isWorkspaceRoute.value ? 'flex-1 min-h-0 flex flex-col overflow-hidden' : '',
)

useDialogFocusTrap(drawerOpen, mobileDrawerRef, {
  onEscape: () => { drawerOpen.value = false },
})

function isMobileNavCurrent(to) {
  if (to === '/') return route.path === '/' || route.path === '/today'
  return route.path === to || route.path.startsWith(`${to}/`)
}

const mobileNav = mobileBottomNav

const drawerNavGroups = computed(() => buildNavGroups({ modules: moduleMapFromRows(store.farmModules) }))

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
    if (evtSource.readyState === EventSource.CONNECTING) return
    evtSource.close()
    evtSource = null
    setTimeout(() => connectSSE(farmContext.farmId), 5000)
  }
}

watch(() => farmContext.farmId, (id) => {
  if (id) connectSSE(id)
  if (id && auth.token) void guardianPanel.fetchNudge(id)
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
  await store.refreshReadings(farmContext.farmId)
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
