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
    <div class="h-14 flex items-center justify-between px-4 sm:px-6">
      <div class="flex items-center gap-3">
        <!-- Mobile hamburger -->
        <button
          class="md:hidden p-1.5 rounded-md text-gray-400 hover:text-white hover:bg-gray-800"
          @click="$emit('toggle-drawer')"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none"
            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="3" y1="6" x2="21" y2="6"/>
            <line x1="3" y1="12" x2="21" y2="12"/>
            <line x1="3" y1="18" x2="21" y2="18"/>
          </svg>
        </button>
        <h1 class="text-sm font-semibold text-gray-300">{{ title }}</h1>
      </div>
      <div class="flex items-center gap-4">
        <button
          v-if="showGuardianButton"
          type="button"
          class="flex items-center gap-1.5 rounded-lg border px-2 py-1 text-xs font-semibold transition-all duration-200"
          :class="guardianPanel.open
            ? 'border-green-500/70 bg-green-950/70 text-green-300 shadow-sm shadow-green-900/30'
            : 'border-green-800/50 bg-green-950/40 text-green-400 hover:border-green-600/60 hover:bg-green-900/50 hover:text-green-300 guardian-topbar-btn'"
          title="Farm Guardian"
          data-test="topbar-guardian-toggle"
          @click="guardianPanel.toggle()"
        >
          <span class="text-base leading-none guardian-topbar-icon" aria-hidden="true">✨</span>
          <span class="hidden sm:inline">Guardian</span>
          <span class="sr-only sm:hidden">Farm Guardian</span>
          <span
            v-if="guardianProposals.pendingCount > 0"
            class="min-w-[1.125rem] h-[1.125rem] px-1 rounded-full bg-amber-500 text-[10px] font-bold text-amber-950 flex items-center justify-center"
            data-test="topbar-guardian-pending-badge"
          >
            {{ guardianProposals.pendingCount > 9 ? '9+' : guardianProposals.pendingCount }}
          </span>
        </button>
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
        <span :class="apiOk ? 'text-gr33n-400' : 'text-danger'" class="text-xs font-mono hidden sm:inline">
          {{ apiOk ? '● API online' : '● API offline' }}
        </span>
        <span class="text-xs text-gray-500 hidden sm:inline">{{ now }}</span>
        <span v-if="auth.username" class="text-xs text-gray-500 hidden sm:inline">{{ auth.username }}</span>
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
import { useCapabilitiesStore } from '../stores/capabilities'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import api from '../api'

defineEmits(['toggle-drawer'])

const route = useRoute()
const auth  = useAuthStore()
const farmStore = useFarmStore()
const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()
const guardianPanel = useGuardianPanelStore()
const guardianProposals = useGuardianProposalsStore()

const showGuardianButton = computed(() => capabilities.loaded && !capabilities.isLite)
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
  '/plants': 'Plants',
  '/catalog': 'Catalog',
  '/costs': 'Costs',
  '/settings': 'Settings',
  '/chat': 'Farm Guardian',
  '/guardian/requests': 'Guardian requests',
}
const title = computed(() => {
  if (route.path.startsWith('/zones/')) return 'Zone Details'
  return labels[route.path] ?? 'gr33n'
})

let tick
onMounted(async () => {
  auth.fetchAuthMode()
  if (!capabilities.loaded) await capabilities.fetch()
  tick = setInterval(async () => {
    now.value = new Date().toLocaleTimeString()
    try { await api.get('/health'); apiOk.value = true }
    catch { apiOk.value = false }
    if (farmContext.farmId) {
      try { await farmStore.countUnreadAlerts(farmContext.farmId) } catch {}
      if (showGuardianButton.value) {
        try { await guardianProposals.refreshPendingCount(farmContext.farmId) } catch {}
      }
    }
  }, 5000)
  now.value = new Date().toLocaleTimeString()
  if (farmContext.farmId) {
    try { await farmStore.countUnreadAlerts(farmContext.farmId) } catch {}
    if (showGuardianButton.value) {
      try { await guardianProposals.refreshPendingCount(farmContext.farmId) } catch {}
    }
  }
})
onUnmounted(() => clearInterval(tick))
</script>

<style scoped>
.guardian-topbar-btn:hover .guardian-topbar-icon {
  animation: guardian-wiggle 0.45s ease-in-out;
}

@keyframes guardian-wiggle {
  0%, 100% { transform: rotate(0deg) scale(1); }
  20% { transform: rotate(-8deg) scale(1.08); }
  40% { transform: rotate(8deg) scale(1.08); }
  60% { transform: rotate(-4deg) scale(1.04); }
  80% { transform: rotate(4deg) scale(1.04); }
}
</style>
