<template>
  <div class="p-6 max-w-5xl">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-green-400">Alerts</h1>
      <div class="flex items-center gap-3">
        <select v-model="severityFilter"
          class="bg-zinc-800 border border-zinc-700 text-gray-300 text-xs rounded-lg px-3 py-1.5 focus:outline-none focus:ring-1 focus:ring-gr33n-600">
          <option value="">All severities</option>
          <option value="critical">Critical</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <button @click="refresh"
          class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded-lg px-3 py-1.5 transition-colors">
          Refresh
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading alerts...</div>
    <div v-else-if="filtered.length === 0" class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center">
      No alerts{{ severityFilter ? ` with severity "${severityFilter}"` : '' }}.
    </div>

    <div v-else class="space-y-2">
      <div v-for="a in filtered" :key="a.id"
        class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 flex items-start gap-4"
        :class="{ 'opacity-60': a.is_acknowledged }">
        <span :class="severityBadge(a.severity)" class="mt-0.5 text-xs font-bold px-2 py-0.5 rounded uppercase shrink-0">
          {{ a.severity?.gr33ncore_notification_priority_enum || a.severity || 'medium' }}
        </span>
        <div class="flex-1 min-w-0">
          <p class="text-white text-sm font-medium truncate">{{ a.subject_rendered || 'Alert' }}</p>
          <p class="text-zinc-400 text-xs mt-0.5">{{ a.message_text_rendered }}</p>
          <p class="text-zinc-600 text-xs mt-1">{{ formatTime(a.created_at) }}</p>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <span v-if="a.is_read" class="text-zinc-600 text-xs">Read</span>
          <button v-else @click="markRead(a.id)"
            class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded px-2 py-1 transition-colors">
            Mark read
          </button>
          <span v-if="a.is_acknowledged" class="text-green-600 text-xs font-medium">ACK</span>
          <button v-else @click="acknowledge(a.id)"
            class="text-xs text-green-500 hover:text-green-300 border border-green-800 rounded px-2 py-1 transition-colors">
            Acknowledge
          </button>
        </div>
      </div>
    </div>

    <div v-if="!loading && filtered.length >= 50" class="mt-4 text-center">
      <button @click="loadMore"
        class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded-lg px-4 py-2 transition-colors">
        Load more
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const farmStore = useFarmStore()
const farmContext = useFarmContextStore()
const loading = ref(false)
const severityFilter = ref('')
const offset = ref(0)

const filtered = computed(() => {
  if (!severityFilter.value) return farmStore.alerts
  return farmStore.alerts.filter(a => {
    const sev = a.severity?.gr33ncore_notification_priority_enum || a.severity || ''
    return sev === severityFilter.value
  })
})

function severityBadge(sev) {
  const s = sev?.gr33ncore_notification_priority_enum || sev || 'medium'
  return {
    critical: 'bg-red-900 text-red-300 border border-red-700',
    high:     'bg-orange-900 text-orange-300 border border-orange-700',
    medium:   'bg-yellow-900 text-yellow-300 border border-yellow-700',
    low:      'bg-zinc-700 text-zinc-300 border border-zinc-600',
  }[s] || 'bg-zinc-700 text-zinc-300'
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleString()
}

async function refresh() {
  if (!farmContext.farmId) return
  loading.value = true
  offset.value = 0
  try {
    await farmStore.loadAlerts(farmContext.farmId, { limit: 50, offset: 0 })
    await farmStore.countUnreadAlerts(farmContext.farmId)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  offset.value += 50
  const more = await farmStore.loadAlerts(farmContext.farmId, { limit: 50, offset: offset.value })
  if (more.length === 0) offset.value -= 50
}

async function markRead(id) {
  await farmStore.markAlertRead(id)
  await farmStore.countUnreadAlerts(farmContext.farmId)
}

async function acknowledge(id) {
  await farmStore.markAlertAcknowledged(id)
  await farmStore.countUnreadAlerts(farmContext.farmId)
}

onMounted(refresh)
watch(() => farmContext.farmId, refresh)
</script>
