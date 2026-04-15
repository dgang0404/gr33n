<template>
  <aside class="w-56 bg-gray-900 border-r border-gray-800 flex flex-col">
    <div class="px-5 py-5 flex items-center gap-2 border-b border-gray-800">
      <span class="text-gr33n-400 text-2xl font-bold tracking-tight">gr33n</span>
      <span class="text-gray-500 text-xs mt-1">v0.1</span>
    </div>
    <nav class="flex-1 px-3 py-4 space-y-1">
      <RouterLink v-for="item in nav" :key="item.to" :to="item.to"
        class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
        active-class="bg-gr33n-900 text-gr33n-400 font-semibold">
        <span class="text-lg">{{ item.icon }}</span>
        {{ item.label }}
      </RouterLink>
    </nav>
    <div class="px-3 py-3 border-t border-gray-800">
      <label class="block text-[10px] uppercase tracking-wide text-gray-500 mb-1">Farm</label>
      <select
        :value="farmContext.farmId ?? ''"
        :disabled="!farmContext.farms.length"
        @change="onFarmSelect($event)"
        class="w-full bg-gray-800 border border-gray-700 text-gray-300 text-xs rounded-lg px-2 py-1.5 focus:outline-none focus:ring-1 focus:ring-gr33n-600 disabled:opacity-60"
      >
        <option v-if="!farmContext.farms.length" value="" disabled>
          {{ emptyFarmHint }}
        </option>
        <option v-for="f in farmContext.farms" :key="f.id" :value="f.id">
          {{ f.name }}
        </option>
      </select>
    </div>
  </aside>
</template>

<script setup>
import { computed } from 'vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useAuthStore } from '../stores/auth'

const farmContext = useFarmContextStore()
const auth = useAuthStore()

const emptyFarmHint = computed(() => {
  if (!auth.token) return 'Sign in to load farms'
  return 'No farms — check API DB / seed (see README)'
})

function onFarmSelect(ev) {
  const raw = ev.target.value
  if (raw === '' || raw == null) return
  const id = Number(raw)
  if (!Number.isFinite(id)) return
  farmContext.selectFarm(id)
}

const nav = [
  { to: '/',          icon: '🌿', label: 'Dashboard'  },
  { to: '/zones',     icon: '🗂️', label: 'Zones'       },
  { to: '/sensors',   icon: '📡', label: 'Sensors'     },
  { to: '/actuators', icon: '⚡', label: 'Controls'    },
  { to: '/schedules', icon: '📅', label: 'Schedules'   },
  { to: '/tasks',        icon: '✅', label: 'Tasks'        },
  { to: '/fertigation', icon: '💧', label: 'Fertigation' },
  { to: '/inventory',   icon: '🧪', label: 'Inventory'   },
  { to: '/costs',       icon: '💰', label: 'Costs'       },
  { to: '/alerts',      icon: '🔔', label: 'Alerts'      },
  { to: '/settings',    icon: '⚙️', label: 'Settings'    },
]
</script>
