<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-alerts-panel">
    <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
      <div class="flex items-center gap-2">
        <h2 class="text-sm font-semibold text-white">Alerts for this zone</h2>
        <span
          v-if="unreadCount"
          class="text-[10px] font-semibold px-2 py-0.5 rounded-full bg-amber-900/60 text-amber-200"
          data-test="zone-alerts-unread-badge"
        >
          {{ unreadCount }} open
        </span>
      </div>
      <router-link
        :to="{ path: '/alerts', query: { zone_id: String(zoneId) } }"
        class="text-xs text-green-600 hover:text-green-400"
      >
        All farm alerts →
      </router-link>
    </div>

    <p v-if="!displayAlerts.length" class="text-zinc-500 text-sm">No alerts for this zone right now.</p>

    <div v-else class="space-y-2">
      <div
        v-for="a in displayAlerts"
        :key="a.id"
        class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 flex flex-col sm:flex-row gap-3 sm:items-start"
        :class="{ 'opacity-70': a.is_acknowledged }"
        :data-test="`zone-alert-row-${a.id}`"
      >
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2 mb-1">
            <span :class="severityClass(a.severity)" class="text-[10px] font-bold px-1.5 py-0.5 rounded uppercase">
              {{ severityLabel(a.severity) }}
            </span>
            <span v-if="!a.is_read" class="text-[10px] text-amber-400">Unread</span>
          </div>
          <p class="text-sm text-zinc-200 font-medium truncate">{{ a.subject_rendered || 'Alert' }}</p>
          <p class="text-xs text-zinc-500 mt-0.5 line-clamp-2">{{ a.message_text_rendered }}</p>
          <p class="text-[10px] text-zinc-600 mt-1">{{ formatTime(a.created_at) }}</p>
        </div>
        <div class="flex flex-wrap gap-2 shrink-0">
          <AskGuardianButton
            v-if="!a.is_read"
            :prefilled-message="`Explain alert #${a.id} (${a.subject_rendered || 'alert'}) for ${zoneName} and what I should do in the next 10 minutes.`"
            :context-ref="{ type: 'alert', id: a.id, zone_id: zoneId, name: zoneName }"
          />
          <button
            v-if="!a.is_read"
            type="button"
            class="text-xs text-zinc-400 hover:text-white border border-zinc-700 rounded px-2 py-1"
            :disabled="busyId === a.id"
            @click="markRead(a.id)"
          >
            Mark read
          </button>
          <span v-if="a.is_acknowledged" class="text-xs text-green-600 font-medium self-center">Acknowledged</span>
          <button
            v-else
            type="button"
            class="text-xs text-green-500 hover:text-green-300 border border-green-800 rounded px-2 py-1"
            :disabled="busyId === a.id"
            data-test="zone-alert-ack"
            @click="acknowledge(a.id)"
          >
            Acknowledge
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { filterZoneAlertsForRoom } from '../lib/zoneGrowSummary.js'
import AskGuardianButton from './AskGuardianButton.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  zoneName: { type: String, default: '' },
  sensors: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  limit: { type: Number, default: 8 },
})

const emit = defineEmits(['refresh'])

const store = useFarmStore()
const busyId = ref(null)

const roomAlerts = computed(() =>
  filterZoneAlertsForRoom(props.alerts, props.sensors, props.zoneName),
)

const unreadCount = computed(() =>
  roomAlerts.value.filter((a) => !a.is_read && !a.is_acknowledged).length,
)

const displayAlerts = computed(() => {
  const sorted = [...roomAlerts.value].sort(
    (a, b) => new Date(b.created_at) - new Date(a.created_at),
  )
  const unread = sorted.filter((a) => !a.is_read || !a.is_acknowledged)
  const rest = sorted.filter((a) => a.is_read && a.is_acknowledged)
  return [...unread, ...rest].slice(0, props.limit)
})

function severityLabel(sev) {
  const s = sev?.gr33ncore_notification_priority_enum || sev
  return String(s || 'medium')
}

function severityClass(sev) {
  const s = severityLabel(sev).toLowerCase()
  if (s === 'critical') return 'bg-red-900/70 text-red-300'
  if (s === 'high') return 'bg-orange-900/60 text-orange-300'
  if (s === 'low') return 'bg-zinc-800 text-zinc-400'
  return 'bg-amber-900/50 text-amber-300'
}

function formatTime(iso) {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}

async function markRead(id) {
  busyId.value = id
  try {
    await store.markAlertRead(id)
    emit('refresh')
  } finally {
    busyId.value = null
  }
}

async function acknowledge(id) {
  busyId.value = id
  try {
    await store.markAlertAcknowledged(id)
    emit('refresh')
  } finally {
    busyId.value = null
  }
}
</script>
