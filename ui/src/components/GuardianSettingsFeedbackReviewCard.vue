<template>
  <section
    v-if="capabilities.aiEnabled && farmId && isFarmAdmin"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-guardian-feedback"
  >
    <h2 class="text-white font-semibold mb-2 flex items-center gap-2">
      <span>👎</span> Guardian feedback — review queue
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Thumbs-down turns from Farm Guardian chat. Review after
      <code class="text-zinc-400">make guardian-qa-smoke</code> or weekly agronomy triage.
      See <code class="text-zinc-400">docs/guardian-feedback-review-runbook.md</code> for the smoke quality checklist and triage steps.
    </p>

    <div class="flex flex-wrap items-center gap-2 mb-3 text-xs">
      <label class="text-zinc-500">Since</label>
      <select
        v-model="since"
        class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-200"
        data-test="settings-guardian-feedback-since"
        @change="loadFeedback"
      >
        <option value="7d">7 days</option>
        <option value="30d">30 days</option>
      </select>
      <button
        type="button"
        class="px-2 py-1 rounded border border-zinc-600 text-zinc-200 hover:bg-zinc-800 disabled:opacity-40"
        data-test="settings-guardian-feedback-refresh"
        :disabled="loading"
        @click="loadFeedback"
      >
        Refresh
      </button>
      <button
        type="button"
        class="px-2 py-1 rounded border border-zinc-600 text-zinc-200 hover:bg-zinc-800 disabled:opacity-40"
        data-test="settings-guardian-feedback-csv"
        :disabled="csvBusy || loading"
        @click="downloadCsv"
      >
        {{ csvBusy ? 'Downloading…' : 'Download CSV' }}
      </button>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading…</div>
    <p v-else-if="loadError" class="text-xs text-red-300/90">{{ loadError }}</p>

    <div v-else class="space-y-3 text-sm">
      <p class="text-xs text-zinc-400" data-test="settings-guardian-feedback-counts">
        <span class="text-amber-200/90 font-medium">{{ downRows.length }}</span> down
        · {{ upRows.length }} up
        · {{ rows.length }} total rated
      </p>

      <div v-if="!downRows.length" class="text-xs text-zinc-500 italic">
        No thumbs-down in this window — good sign after smoke, or operators haven't rated yet.
      </div>

      <div v-else class="overflow-x-auto rounded-lg border border-zinc-700">
        <table class="w-full text-xs" data-test="settings-guardian-feedback-table">
          <thead class="bg-zinc-900/80 text-zinc-500 text-left">
            <tr>
              <th class="px-3 py-2 font-medium">When</th>
              <th class="px-3 py-2 font-medium">Question</th>
              <th class="px-3 py-2 font-medium">Reason</th>
              <th class="px-3 py-2 font-medium">Model</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-700/80">
            <tr v-for="row in downRows" :key="`${row.session_id}-${row.turn_index}`">
              <td class="px-3 py-2 text-zinc-500 whitespace-nowrap">{{ formatWhen(row.feedback_at || row.created_at) }}</td>
              <td class="px-3 py-2 text-zinc-300 max-w-xs truncate" :title="row.question">{{ row.question }}</td>
              <td class="px-3 py-2 text-amber-200/90">{{ row.reason || '—' }}</td>
              <td class="px-3 py-2 text-zinc-500 font-mono">{{ row.model || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import api from '../api'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'

defineProps({
  isFarmAdmin: { type: Boolean, default: false },
})

const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()

const farmId = computed(() => farmContext.farmId)
const since = ref('7d')
const loading = ref(false)
const csvBusy = ref(false)
const loadError = ref('')
const rows = ref([])

const downRows = computed(() => rows.value.filter((r) => r.rating === 'down'))
const upRows = computed(() => rows.value.filter((r) => r.rating === 'up'))

function formatWhen(iso) {
  if (!iso) return '—'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString()
}

async function loadFeedback() {
  if (!farmId.value) return
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/v1/chat/feedback/export', {
      params: { farm_id: farmId.value, since: since.value },
    })
    rows.value = Array.isArray(data.rows) ? data.rows : []
  } catch (e) {
    loadError.value = e?.response?.data?.error || e.message || 'Could not load feedback.'
    rows.value = []
  } finally {
    loading.value = false
  }
}

async function downloadCsv() {
  if (!farmId.value) return
  csvBusy.value = true
  try {
    const r = await api.get('/v1/chat/feedback/export', {
      params: { farm_id: farmId.value, since: since.value, format: 'csv' },
      responseType: 'blob',
    })
    const blob = new Blob([r.data], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `guardian-feedback-farm-${farmId.value}-${since.value}.csv`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    loadError.value = e?.response?.data?.error || e.message || 'CSV download failed.'
  } finally {
    csvBusy.value = false
  }
}

watch(farmId, (id) => {
  if (id && capabilities.aiEnabled) void loadFeedback()
})

onMounted(() => {
  if (farmId.value && capabilities.aiEnabled) void loadFeedback()
})
</script>
