<template>
  <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3" data-test="device-ip-history-panel">
    <div>
      <h3 class="text-sm font-semibold text-white">IP address history</h3>
      <p class="text-xs text-zinc-500 mt-0.5">
        Read-only — auto-logged on heartbeat when the device's observed IP changes.
      </p>
    </div>

    <p v-if="loading" class="text-xs text-zinc-500">Loading history…</p>
    <p v-else-if="loadError" class="text-xs text-red-400">{{ loadError }}</p>

    <ul v-else-if="events.length" class="space-y-2 text-xs">
      <li
        v-for="(e, i) in events"
        :key="`${e.device_id}-${e.observed_at}-${i}`"
        class="flex items-center justify-between gap-2 bg-zinc-950/60 border border-zinc-800 rounded-lg px-3 py-2"
        data-test="device-ip-event-row"
      >
        <div class="min-w-0 font-mono">
          <span v-if="e.old_ip" class="text-zinc-500">{{ e.old_ip }} → </span>
          <span class="text-zinc-200">{{ e.new_ip }}</span>
        </div>
        <span class="text-zinc-600 text-[11px] shrink-0">{{ formatDate(e.observed_at) }}</span>
      </li>
    </ul>

    <p v-else class="text-xs text-zinc-500">No IP changes recorded yet.</p>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import api from '../api'

const props = defineProps({
  deviceId: { type: Number, required: true },
})

const events = ref([])
const loading = ref(false)
const loadError = ref('')

async function refresh() {
  if (!props.deviceId) return
  loading.value = true
  loadError.value = ''
  try {
    const r = await api.get(`/devices/${props.deviceId}/ip-history`)
    events.value = Array.isArray(r.data) ? r.data : []
  } catch (e) {
    loadError.value = e.response?.data?.error || e.message || 'Failed to load IP history'
  } finally {
    loading.value = false
  }
}

function formatDate(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

onMounted(refresh)
watch(() => props.deviceId, refresh)
</script>
