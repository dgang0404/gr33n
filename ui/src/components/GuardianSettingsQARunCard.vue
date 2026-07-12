<template>
  <section
    v-if="capabilities.aiEnabled"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-guardian-qa"
  >
    <h2 class="text-white font-semibold mb-2 flex items-center gap-2">
      <span>🧪</span> Guardian QA — last run
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Latest archived smoke or regression run from your server’s quality checks.
    </p>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading…</div>
    <p v-else-if="error" class="text-xs text-zinc-500" data-test="settings-guardian-qa-empty">
      {{ error }}
    </p>

    <div v-else-if="summary" class="space-y-3 text-sm">
      <div class="flex flex-wrap items-center gap-2 text-xs">
        <span
          class="px-2 py-0.5 rounded-md font-semibold border"
          :class="summary.all_passed ? 'bg-green-950/50 text-green-300 border-green-800' : 'bg-amber-950/40 text-amber-200 border-amber-800'"
          data-test="settings-guardian-qa-pass-badge"
        >
          {{ summary.passed }}/{{ summary.total }} passed
        </span>
        <span class="text-zinc-400">{{ summary.suite }} · {{ summary.model }}</span>
        <span v-if="whenLabel" class="text-zinc-600">{{ whenLabel }}</span>
      </div>

      <p v-if="summary.report_path" class="text-[10px] text-zinc-600 font-mono break-all" data-test="settings-guardian-qa-path">
        {{ summary.report_path }}
      </p>

      <div v-if="scores.length" class="overflow-x-auto rounded-lg border border-zinc-700">
        <table class="w-full text-xs" data-test="settings-guardian-qa-steps">
          <thead class="bg-zinc-900/80 text-zinc-500 text-left">
            <tr>
              <th class="px-3 py-2 font-medium">Step</th>
              <th class="px-3 py-2 font-medium">Result</th>
              <th v-if="showRelevanceCol" class="px-3 py-2 font-medium">Relevance</th>
              <th v-if="showCritiqueCol" class="px-3 py-2 font-medium">Critique</th>
              <th class="px-3 py-2 font-medium">Notes</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-700/80">
            <tr v-for="row in scores" :key="row.id">
              <td class="px-3 py-2 text-zinc-300 font-mono">{{ row.id }}</td>
              <td class="px-3 py-2">
                <span :class="row.passed ? 'text-green-400' : 'text-amber-300'">
                  {{ row.passed ? 'pass' : 'fail' }}
                </span>
              </td>
              <td v-if="showRelevanceCol" class="px-3 py-2 text-zinc-500 font-mono whitespace-nowrap">
                <span v-if="row.question_answer_relevance != null">
                  {{ formatRel(row.question_answer_relevance) }}
                  <span v-if="row.low_relevance" class="text-amber-400"> low</span>
                </span>
                <span v-else>—</span>
              </td>
              <td v-if="showCritiqueCol" class="px-3 py-2 whitespace-nowrap">
                <span v-if="row.critique_pass === true" class="text-green-400">yes</span>
                <span v-else-if="row.critique_pass === false" class="text-amber-300">no</span>
                <span v-else class="text-zinc-600">—</span>
              </td>
              <td class="px-3 py-2 text-zinc-500 max-w-xs truncate" :title="row.notes">{{ row.notes || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-if="summary" class="text-[10px] text-zinc-500" data-test="settings-guardian-qa-quality-nudge">
        Automated pass is not a quality sign-off — spot-check archived answers, then review Settings → Guardian feedback.
      </p>

      <p class="text-[10px] text-zinc-600">
        Ask your farm admin to re-run quality checks after Guardian upgrades.
      </p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import api from '../api'
import { useCapabilitiesStore } from '../stores/capabilities'

const capabilities = useCapabilitiesStore()
const loading = ref(false)
const error = ref('')
const summary = ref(null)
const scores = ref([])

const whenLabel = computed(() => {
  const raw = summary.value?.updated_at
  if (!raw) return ''
  const d = new Date(raw)
  return Number.isNaN(d.getTime()) ? raw : d.toLocaleString()
})

const showRelevanceCol = computed(() =>
  scores.value.some((row) => row.question_answer_relevance != null),
)

const showCritiqueCol = computed(() =>
  scores.value.some((row) => row.critique_pass != null),
)

function formatRel(v) {
  if (v == null || Number.isNaN(Number(v))) return '—'
  return Number(v).toFixed(2)
}

async function loadLatest() {
  loading.value = true
  error.value = ''
  summary.value = null
  scores.value = []
  try {
    const { data } = await api.get('/v1/guardian/qa/latest')
    summary.value = data.summary
    scores.value = Array.isArray(data.scores) ? data.scores : []
  } catch (e) {
    if (e?.response?.status === 404) {
      error.value = 'No QA runs archived yet — ask your farm admin to run a smoke check.'
    } else {
      error.value = e?.message || 'Could not load QA summary.'
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (capabilities.aiEnabled) void loadLatest()
})
</script>
