<template>
  <section
    v-if="capabilities.aiEnabled && farmId"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-guardian-corpus"
  >
    <h2 class="text-white font-semibold mb-2 flex items-center gap-2">
      <span>📚</span> Field memories (RAG corpus)
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Indexed guides and farm rows power Farm counsel citations. Re-ingest when guides change or
      operational data is stale. First-time setup:
      <code class="text-zinc-400">make guardian-bootstrap-farm FARM_ID={{ farmId }}</code>
    </p>

    <div v-if="!readiness.loaded && readiness.loading" class="text-zinc-500 text-sm">Loading corpus…</div>

    <div v-else-if="corpus" class="space-y-3">
      <p
        v-if="corpusWarning"
        class="text-xs text-amber-300/90 rounded border border-amber-900/50 bg-amber-950/30 px-3 py-2"
        data-test="settings-guardian-corpus-warn"
      >
        {{ corpusWarning }}
      </p>

      <div class="overflow-x-auto rounded-lg border border-zinc-700">
        <table class="w-full text-xs" data-test="settings-guardian-corpus-table">
          <thead class="bg-zinc-900/80 text-zinc-500 text-left">
            <tr>
              <th class="px-3 py-2 font-medium">Corpus</th>
              <th class="px-3 py-2 font-medium">Chunks</th>
              <th class="px-3 py-2 font-medium">Last ingested</th>
              <th v-if="isFarmAdmin" class="px-3 py-2 font-medium">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-700/80">
            <tr v-for="row in rows" :key="row.scope" :class="row.rowClass">
              <td class="px-3 py-2 text-zinc-200">{{ row.label }}</td>
              <td class="px-3 py-2 text-zinc-300 font-mono" :data-test="`settings-corpus-chunks-${row.scope}`">
                {{ row.chunks }}
              </td>
              <td class="px-3 py-2" :class="row.ageClass" :data-test="`settings-corpus-age-${row.scope}`">
                {{ row.ageLabel }}
              </td>
              <td v-if="isFarmAdmin" class="px-3 py-2">
                <button
                  type="button"
                  class="text-xs px-2 py-1 rounded border border-zinc-600 text-zinc-200 hover:bg-zinc-800 disabled:opacity-40"
                  :data-test="`settings-corpus-reingest-${row.scope}`"
                  :disabled="reingestBusy"
                  @click="startReingest(row.scope)"
                >
                  Re-ingest
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div
        v-if="job?.status === 'running'"
        class="text-xs text-amber-200/90 flex items-center gap-2"
        data-test="settings-guardian-corpus-progress"
      >
        <span class="inline-block w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
        Re-ingesting {{ job.scope }}…
      </div>
      <p v-if="reingestError" class="text-xs text-red-300/90" data-test="settings-guardian-corpus-error">
        {{ reingestError }}
      </p>
      <p v-if="job?.status === 'done'" class="text-xs text-green-400/90" data-test="settings-guardian-corpus-done">
        Last re-ingest finished — refresh health to see updated counts.
      </p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'
import api from '../api'

const props = defineProps({
  isFarmAdmin: { type: Boolean, default: false },
})

const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const readiness = useGuardianReadinessStore()

const farmId = computed(() => farmContext.farmId)
const corpus = computed(() => readiness.awakening?.corpus ?? null)
const job = ref(null)
const reingestError = ref('')
let pollTimer = null

const reingestBusy = computed(() => job.value?.status === 'running')

const corpusWarning = computed(() => {
  const c = corpus.value
  if (!c) return ''
  if (c.staleness === 'field_guide_empty') {
    return 'Field memories not loaded — bootstrap or re-ingest field guides.'
  }
  if (c.staleness === 'operational_stale') {
    return 'Operational memories are stale (>7 days) — re-ingest for fresher farm context.'
  }
  if (c.staleness === 'operational_aging') {
    return 'Operational memories are aging — consider re-ingest soon.'
  }
  return ''
})

function formatAge(iso, freshness) {
  if (freshness === 'empty') return 'Never'
  if (!iso) return 'Unknown'
  try {
    const d = new Date(iso)
    const diffMs = Date.now() - d.getTime()
    const days = Math.floor(diffMs / 86400000)
    if (days < 1) return 'Today'
    if (days === 1) return '1d ago'
    return `${days}d ago`
  } catch {
    return '—'
  }
}

function ageClass(freshness) {
  if (freshness === 'stale' || freshness === 'empty') return 'text-amber-300/90'
  if (freshness === 'aging') return 'text-amber-200/70'
  return 'text-zinc-400'
}

const rows = computed(() => {
  const c = corpus.value
  if (!c) return []
  return [
    {
      scope: 'field_guides',
      label: 'Field guides',
      chunks: c.field_guide_chunks ?? 0,
      ageLabel: formatAge(c.field_guide_last_ingested_at, c.field_guide_freshness),
      ageClass: ageClass(c.field_guide_freshness),
      rowClass: c.field_guide_freshness === 'empty' ? 'bg-amber-950/20' : '',
    },
    {
      scope: 'platform_docs',
      label: 'Platform docs',
      chunks: c.platform_doc_chunks ?? 0,
      ageLabel: formatAge(c.platform_last_ingested_at, c.platform_freshness),
      ageClass: ageClass(c.platform_freshness),
      rowClass: '',
    },
    {
      scope: 'operational',
      label: 'Operational',
      chunks: c.operational_chunks ?? 0,
      ageLabel: formatAge(c.operational_last_ingested_at, c.operational_freshness),
      ageClass: ageClass(c.operational_freshness),
      rowClass: c.operational_freshness === 'stale' ? 'bg-amber-950/20' : '',
    },
  ]
})

async function refreshHealth() {
  if (!farmId.value) return
  await readiness.fetchHealth(farmId.value, 'farm_counsel')
}

async function pollStatus() {
  if (!farmId.value) return
  try {
    const { data } = await api.get(`/farms/${farmId.value}/guardian/reingest/status`)
    job.value = data?.status === 'idle' ? null : data
    if (job.value?.status === 'running') {
      pollTimer = setTimeout(pollStatus, 2000)
    } else if (job.value?.status === 'done') {
      await refreshHealth()
    } else if (job.value?.status === 'failed') {
      reingestError.value = job.value.error || 'Re-ingest failed'
    }
  } catch (e) {
    reingestError.value = e.response?.data?.error || e.message || 'Status check failed'
  }
}

async function startReingest(scope) {
  if (!farmId.value || !props.isFarmAdmin) return
  reingestError.value = ''
  try {
    const { data } = await api.post(`/farms/${farmId.value}/guardian/reingest`, { scope })
    job.value = data
    if (data?.status === 'running') {
      if (pollTimer) clearTimeout(pollTimer)
      pollTimer = setTimeout(pollStatus, 1500)
    }
  } catch (e) {
    reingestError.value = e.response?.data?.error || e.message || 'Re-ingest failed'
  }
}

onMounted(async () => {
  await refreshHealth()
  await pollStatus()
})

watch(farmId, async () => {
  job.value = null
  reingestError.value = ''
  await refreshHealth()
  await pollStatus()
})

onUnmounted(() => {
  if (pollTimer) clearTimeout(pollTimer)
})
</script>
