<template>
  <header class="bg-gray-900 border-b border-gray-800">
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
          aria-label="Open navigation menu"
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
          v-if="guardianAvailable"
          type="button"
          class="relative hidden sm:inline-flex items-center gap-1.5 text-xs font-semibold text-green-400 border border-green-800/80 rounded-lg px-2.5 py-1 hover:bg-green-950/50 hover:text-green-300 transition-colors"
          title="Open Farm Guardian"
          data-test="topbar-guardian-button"
          @click="openGuardianDrawer"
        >
          <span class="relative" aria-hidden="true">
            ✨
            <span
              v-if="proposalsStore.pendingCount > 0"
              class="absolute -top-1.5 -right-2 min-w-[1rem] h-4 px-0.5 rounded-full bg-amber-500 text-[9px] font-bold text-amber-950 flex items-center justify-center ring-2 ring-gray-900"
              data-test="topbar-guardian-pending-badge"
            >
              {{ proposalsStore.pendingCount > 9 ? '9+' : proposalsStore.pendingCount }}
            </span>
            <span
              v-else-if="guardianPanel.showNudgeDot"
              class="absolute -top-1 -right-1.5 h-2 w-2 rounded-full ring-2 ring-gray-900"
              :class="nudgeDotStirring ? 'bg-amber-500 animate-pulse' : 'bg-amber-400'"
              data-test="topbar-guardian-nudge-dot"
              :data-stirring="nudgeDotStirring ? 'true' : undefined"
              aria-hidden="true"
            />
          </span>
          Ask gr33n
        </button>
        <RouterLink
          v-nav-hint="'/alerts'"
          to="/alerts"
          class="relative text-gray-400 hover:text-white transition-colors"
          :title="alertBellTitle"
        >
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
        <time
          class="text-xs text-gray-500 hidden sm:inline tabular-nums"
          :datetime="nowIso"
          :title="clockTitle"
        >
          {{ nowLabel }}
        </time>
        <span v-if="auth.username" class="text-xs text-gray-500 hidden sm:inline">{{ auth.username }}</span>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'
import api, { isUnauthorizedError } from '../api'

defineEmits(['toggle-drawer'])

const route = useRoute()
const auth  = useAuthStore()
const farmStore = useFarmStore()
const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()
const guardianPanel = useGuardianPanelStore()
const proposalsStore = useGuardianProposalsStore()
const guardianReadiness = useGuardianReadinessStore()

const guardianAvailable = computed(() => capabilities.loaded && !capabilities.isLite)

const nudgeDotStirring = computed(() =>
  guardianPanel.criticalNudgePending && guardianReadiness.isStirring,
)

const alertBellTitle = computed(() => {
  const n = farmStore.unreadAlertCount
  if (n <= 0) return 'Alerts'
  return `${n} unread alert${n === 1 ? '' : 's'} on this farm (includes device offline and supply alerts)`
})

function openGuardianDrawer() {
  if (proposalsStore.pendingCount > 0) {
    guardianPanel.openPendingTab()
  } else {
    guardianPanel.openDrawer({ tab: 'chat' })
  }
}

const apiOk = ref(true)
const nowLabel = ref('')
const nowIso = ref('')

const farmTimezone = computed(() => {
  const tz = farmContext.selectedFarm?.timezone?.trim()
  return tz && tz !== 'UTC' ? tz : undefined
})

const clockTitle = computed(() =>
  farmTimezone.value ? `Farm time (${farmTimezone.value})` : 'Local time',
)

function tickClock() {
  const d = new Date()
  nowIso.value = d.toISOString()
  const tz = farmTimezone.value
  const datePart = d.toLocaleDateString(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    timeZone: tz,
  })
  const timePart = d.toLocaleTimeString(undefined, {
    hour: 'numeric',
    minute: '2-digit',
    timeZone: tz,
  })
  nowLabel.value = `${datePart} · ${timePart}`
}
const labels = {
  '/': 'Today',
  '/zones': 'Zones',
  '/sensors': 'Sensors',
  '/actuators': 'Controls',
  '/schedules': 'Schedules',
  '/tasks': 'Tasks',
  '/feeding': 'Feed & water',
  '/fertigation': 'Fertigation',
  '/inventory': 'Inventory',
  '/alerts': 'Alerts',
  '/plants': 'Plants',
  '/catalog': 'Catalog',
  '/costs': 'Costs',
  '/settings': 'Settings',
  '/chat': 'Farm Guardian',
}
const title = computed(() => {
  if (route.path.startsWith('/zones/')) return 'Zone Details'
  return labels[route.path] ?? 'gr33n'
})

let tick
onMounted(async () => {
  auth.fetchAuthMode()
  if (!capabilities.loaded) await capabilities.fetch()
  if (auth.token && farmContext.farmId) proposalsStore.refreshPendingCount(farmContext.farmId)
  tick = setInterval(async () => {
    tickClock()
    try { await api.get('/health'); apiOk.value = true }
    catch { apiOk.value = false }
    if (auth.token && farmContext.farmId) {
      try {
        await farmStore.countUnreadAlerts(farmContext.farmId)
      } catch (e) {
        if (isUnauthorizedError(e) && tick) {
          clearInterval(tick)
          tick = null
        }
      }
    }
  }, 5000)
  tickClock()
  if (auth.token && farmContext.farmId) {
    try { await farmStore.countUnreadAlerts(farmContext.farmId) } catch {}
  }
})
watch(() => farmContext.farmId, (id) => {
  if (id && auth.token) proposalsStore.refreshPendingCount(id)
})
onUnmounted(() => clearInterval(tick))
</script>
