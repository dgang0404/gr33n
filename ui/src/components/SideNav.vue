<template>
  <aside
    class="bg-gray-900 border-r border-gray-800 flex flex-col transition-all duration-200 overflow-hidden"
    :class="collapsed ? 'w-14' : 'w-56'"
  >
    <!-- Logo + hamburger -->
    <div class="px-3 py-4 flex items-center border-b border-gray-800"
      :class="collapsed ? 'justify-center' : 'justify-between'">
      <span v-if="!collapsed" class="text-gr33n-400 text-2xl font-bold tracking-tight pl-2">gr33n</span>
      <button
        @click="toggle"
        class="p-1.5 rounded-md text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
        :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none"
          stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="6" x2="21" y2="6"/>
          <line x1="3" y1="12" x2="21" y2="12"/>
          <line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
      </button>
    </div>

    <!-- Guardian — top of sidebar so operators need not scroll (Phase 40 WS7) -->
    <div class="px-2 pt-2 shrink-0">
      <GuardianNavLaunch :collapsed="collapsed" />
    </div>

    <!-- Nav groups -->
    <nav class="flex-1 overflow-y-auto px-2 py-3 space-y-4">
      <div v-for="group in navGroups" :key="group.label">
        <p v-if="!collapsed" class="px-3 mb-1 text-[10px] uppercase tracking-widest text-gray-600 font-semibold">{{ group.label }}</p>
        <div class="space-y-0.5">
          <template v-for="item in group.items" :key="item.to">
            <RouterLink
              :to="item.to"
              class="flex items-center rounded-lg text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
              :class="[
                collapsed ? 'justify-center px-0 py-2' : 'gap-3 px-3 py-2',
                isHighlightedNav(item.to) ? 'nav-related' : '',
              ]"
              active-class="bg-gr33n-900 text-gr33n-400 font-semibold"
              :title="item.navTitle ?? (collapsed ? item.label : undefined)"
            >
              <span class="text-lg shrink-0">{{ item.icon }}</span>
              <span v-if="!collapsed" class="flex-1 min-w-0">{{ item.label }}</span>
            </RouterLink>
            <!-- Sub-items: indented, only visible when sidebar is expanded -->
            <template v-if="!collapsed && item.children?.length">
              <RouterLink
                v-for="child in item.children"
                :key="child.to"
                :to="child.to"
                class="flex items-center gap-2 rounded-lg text-xs text-gray-500 hover:text-white hover:bg-gray-800 transition-colors pl-8 pr-3 py-1.5"
                :class="isHighlightedNav(child.to) ? 'nav-related' : ''"
                active-class="text-gr33n-400 font-semibold"
                :title="child.navTitle"
              >
                <span class="text-sm shrink-0 opacity-70">{{ child.icon }}</span>
                <span class="flex-1 min-w-0 truncate">{{ child.label }}</span>
              </RouterLink>
            </template>
          </template>
        </div>
      </div>
    </nav>

    <!-- Farm selector -->
    <div class="px-2 py-3 border-t border-gray-800 shrink-0">
      <label v-if="!collapsed" class="block text-[10px] uppercase tracking-wide text-gray-500 mb-1 px-1">Farm</label>
      <select
        :value="farmContext.farmId ?? ''"
        :disabled="!farmContext.farms.length"
        @change="onFarmSelect($event)"
        class="w-full bg-gray-800 border border-gray-700 text-gray-300 rounded-lg focus:outline-none focus:ring-1 focus:ring-gr33n-600 disabled:opacity-60"
        :class="collapsed ? 'text-[10px] px-1 py-1' : 'text-xs px-2 py-1.5'"
      >
        <option v-if="!farmContext.farms.length" value="" disabled>
          {{ emptyFarmHint }}
        </option>
        <option v-for="f in farmContext.farms" :key="f.id" :value="f.id">
          {{ collapsed ? f.name.slice(0, 3) : f.name }}
        </option>
      </select>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useAuthStore } from '../stores/auth'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { buildNavGroups } from '../lib/navGroups.js'
import { useNavHighlightStore } from '../stores/navHighlight'
import GuardianNavLaunch from './GuardianNavLaunch.vue'

const farmContext = useFarmContextStore()
const auth = useAuthStore()
const proposalsStore = useGuardianProposalsStore()

const STORAGE_KEY = 'gr33n_sidebar_collapsed'
const collapsed = ref(localStorage.getItem(STORAGE_KEY) === '1')

function toggle() {
  collapsed.value = !collapsed.value
  localStorage.setItem(STORAGE_KEY, collapsed.value ? '1' : '0')
}

const emit = defineEmits(['collapse-change'])

const emptyFarmHint = computed(() => {
  if (!auth.token) return 'Sign in'
  return 'No farms'
})

function onFarmSelect(ev) {
  const raw = ev.target.value
  if (raw === '' || raw == null) return
  const id = Number(raw)
  if (!Number.isFinite(id)) return
  farmContext.selectFarm(id)
}

watch(
  () => farmContext.farmId,
  (id) => {
    if (id) proposalsStore.refreshPendingCount(id)
    else proposalsStore.pendingCount = 0
  },
  { immediate: true },
)

const navGroups = computed(() => buildNavGroups())

const navHighlight = useNavHighlightStore()

/** Wiggle only the single sidebar tab that matches v-nav-hint (no related-route fan-out). */
function isHighlightedNav(route) {
  return navHighlight.route != null && route === navHighlight.route
}
</script>

<style scoped>
.nav-related {
  color: rgb(74 222 128);
  box-shadow: inset 0 0 0 1px rgb(34 197 94 / 0.45);
}

@media (prefers-reduced-motion: no-preference) {
  .nav-related {
    animation: nav-related-wiggle 0.4s ease-in-out 2;
  }
}

@keyframes nav-related-wiggle {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-2px) rotate(-1deg); }
  75% { transform: translateX(2px) rotate(1deg); }
}
</style>
