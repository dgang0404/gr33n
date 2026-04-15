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
    <div class="px-5 py-4 border-t border-gray-800 text-xs text-gray-600">
      Farm 1 · local
    </div>
  </aside>
</template>

<script setup>
const nav = [
  { to: '/',          icon: '🌿', label: 'Dashboard'  },
  { to: '/zones',     icon: '🗂️', label: 'Zones'       },
  { to: '/sensors',   icon: '📡', label: 'Sensors'     },
  { to: '/actuators', icon: '⚡', label: 'Controls'    },
  { to: '/schedules', icon: '📅', label: 'Schedules'   },
  { to: '/tasks',     icon: '✅', label: 'Tasks'       },
  { to: '/inventory', icon: '🧪', label: 'Inventory'   },
]
</script>
