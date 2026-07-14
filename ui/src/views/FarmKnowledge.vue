<template>
  <div :class="embedded ? 'p-4 sm:p-6 max-w-4xl mx-auto space-y-8' : 'p-4 sm:p-6 max-w-4xl mx-auto space-y-8'">
    <header v-if="!embedded">
      <h1 class="text-2xl font-bold text-green-400 mb-2 flex items-center gap-2">
        Farm knowledge
        <HelpTip position="bottom">
          <strong>Semantic search</strong> over text the farm has already indexed (tasks, automation runs, etc.).
          Results come from your database — not the open web, not API access logs. Static how-to lives under <strong>System → Guide</strong>.
          <strong>Ask (LLM)</strong> only appears when the server has AI features enabled.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500">
        Search indexed chunks for this farm — plain language works; results match by meaning.
      </p>
    </header>

    <div v-if="!farmContext.farmId" class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200">
      Select a farm in the sidebar to use knowledge search.
    </div>

    <template v-else>
      <p
        v-if="citedDoc"
        class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200"
        data-test="farm-knowledge-cited-doc"
        role="status"
      >
        Guardian cited indexed doc: <code class="text-amber-100/90 text-xs">{{ citedDoc }}</code>
      </p>
      <!-- Search form -->
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4" data-test="farm-knowledge-search">
        <div class="space-y-1">
          <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Search</h2>
          <p class="text-sm text-zinc-400 leading-relaxed" data-test="farm-knowledge-semantic-hint">
            Ask in plain language — search is by <strong class="text-zinc-300 font-medium">meaning</strong>, not exact words.
          </p>
        </div>
        <div class="flex flex-col gap-2">
          <label class="sr-only" for="farm-knowledge-query">Question or keywords</label>
          <textarea
            id="farm-knowledge-query"
            v-model="query"
            rows="4"
            placeholder="e.g. wilting in the flower room, when did feed volume change, what failed on the irrigation rule"
            class="bg-zinc-950 border border-zinc-700 rounded-lg px-4 py-3 text-base text-white placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-gr33n-600"
            data-test="farm-knowledge-query"
            @keydown.enter.exact.prevent="runSearch"
          />
        </div>
        <p v-if="!query.trim() && !results.length && answerText === null" class="text-xs text-zinc-600" data-test="farm-knowledge-examples">
          Try: “wilting in flower room”, “when did feed volume change”, “unread humidity alerts”
        </p>
        <div class="flex flex-wrap gap-3">
          <button
            type="button"
            data-test="farm-knowledge-search-button"
            :disabled="searchLoading || !query.trim()"
            @click="runSearch"
            class="px-5 py-2.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40 text-sm font-medium"
          >
            {{ searchLoading ? 'Searching…' : 'Search' }}
          </button>
          <button
            type="button"
            data-test="ask-llm-button"
            :disabled="answerLoading || !query.trim() || capabilities.isLite"
            :title="capabilities.isLite ? 'AI is disabled on this installation (Lite mode) — set AI_ENABLED=true and restart the API.' : ''"
            @click="runAnswer"
            class="px-5 py-2.5 rounded-lg bg-zinc-800 text-gr33n-400 border border-zinc-600 hover:bg-zinc-700 disabled:opacity-40 text-sm font-medium"
          >
            {{ answerLoading ? 'Asking…' : (capabilities.isLite ? 'Ask (LLM) — Lite mode' : 'Ask (LLM)') }}
          </button>
          <button
            type="button"
            class="text-xs text-zinc-500 hover:text-zinc-300 px-2 py-2"
            data-test="farm-knowledge-advanced-toggle"
            :aria-expanded="showAdvanced ? 'true' : 'false'"
            @click="showAdvanced = !showAdvanced"
          >
            {{ showAdvanced ? 'Hide advanced filters' : 'Advanced filters' }}
          </button>
        </div>
        <div
          v-if="showAdvanced"
          class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1 border-t border-zinc-800"
          data-test="farm-knowledge-advanced"
        >
          <div class="flex flex-col gap-1 sm:col-span-2">
            <label class="text-[11px] text-zinc-500 uppercase tracking-wide">Module filter</label>
            <input
              v-model="moduleFilter"
              placeholder="core, automation, field_guide…"
              class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-[11px] text-zinc-500 uppercase tracking-wide">Limit</label>
            <input
              v-model.number="limitN"
              type="number"
              min="1"
              max="50"
              class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white"
            />
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-[11px] text-zinc-500 uppercase tracking-wide">Since (RFC3339)</label>
            <input
              v-model="sinceIso"
              type="text"
              placeholder="optional"
              class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white font-mono"
            />
          </div>
          <div class="flex flex-col gap-1 sm:col-span-2">
            <label class="text-[11px] text-zinc-500 uppercase tracking-wide">Until (RFC3339)</label>
            <input
              v-model="untilIso"
              type="text"
              placeholder="optional"
              class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white font-mono"
            />
          </div>
        </div>
        <p
          v-if="capabilities.isLite"
          data-test="ask-llm-lite-note"
          class="text-amber-200/90 text-xs bg-amber-950/40 border border-amber-900/70 rounded-lg px-3 py-2"
        >
          Farm Guardian and answer synthesis are off on this installation.
          Search still works; the API runs in <strong>Lite mode</strong>.
          Ask your farm admin to enable AI on the server if you need synthesized answers.
        </p>
        <div
          v-if="searchFeedback"
          class="text-sm rounded-lg px-3 py-2 border"
          :class="searchFeedback.degraded
            ? 'text-amber-200 bg-amber-950/40 border-amber-900/70'
            : 'text-red-400 bg-red-950/50 border-red-900'"
        >
          {{ searchFeedback.message }}
        </div>
        <div
          v-if="answerFeedback"
          class="text-sm rounded-lg px-3 py-2 border"
          :class="answerFeedback.degraded
            ? 'text-amber-200 bg-amber-950/40 border-amber-900/70'
            : 'text-red-400 bg-red-950/50 border-red-900'"
        >
          {{ answerFeedback.message }}
        </div>
      </section>

      <FieldGuideBrowse @search-guide="onSearchFieldGuide" />

      <!-- Vector results -->
      <section v-if="results.length" class="space-y-3">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">
          Chunks <span v-if="embedModel" class="text-zinc-600 font-normal normal-case">· embedding {{ embedModel }}</span>
        </h2>
        <div
          v-for="(r, idx) in results"
          :key="r.id ?? idx"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-2"
        >
          <div class="flex flex-wrap gap-2 text-[11px] text-zinc-500 font-mono">
            <span class="text-gr33n-500">#{{ r.chunk_index }}</span>
            <span>{{ r.source_type }}</span>
            <span>source_id={{ r.source_id }}</span>
            <span v-if="r.distance != null">distance={{ formatDistance(r.distance) }}</span>
          </div>
          <pre class="text-sm text-zinc-200 whitespace-pre-wrap font-sans">{{ r.content_text }}</pre>
        </div>
      </section>

      <section
        v-else-if="searchRanEmpty"
        class="rounded-xl border border-zinc-800 bg-zinc-900/80 px-4 py-4 text-sm text-zinc-400"
      >
        <p class="text-zinc-300 font-medium mb-1">No chunks matched</p>
        <p class="leading-relaxed">
          Try different keywords, widen the date filters, or ask your farm admin to index field memories for this farm in Settings.
        </p>
      </section>

      <!-- LLM answer -->
      <section v-if="answerText !== null" class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">
          Answer
          <span v-if="llmModel" class="text-zinc-600 font-normal normal-case">· {{ llmModel }}</span>
        </h2>
        <div class="text-zinc-100 whitespace-pre-wrap text-sm leading-relaxed">{{ answerText }}</div>
        <div v-if="citations.length" class="border-t border-zinc-800 pt-4 space-y-2">
          <p class="text-[11px] uppercase tracking-widest text-zinc-500">Citations</p>
          <ul class="space-y-2">
            <li
              v-for="c in citations"
              :key="c.ref + '-' + c.chunk_id"
              class="text-xs bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-zinc-300"
            >
              <span class="text-gr33n-500 font-mono">[{{ c.ref }}]</span>
              {{ c.source_type }} #{{ c.source_id }} · chunk {{ c.chunk_id }}
              <p class="mt-1 text-zinc-500">{{ c.excerpt }}</p>
            </li>
          </ul>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import HelpTip from '../components/HelpTip.vue'
import FieldGuideBrowse from '../components/FieldGuideBrowse.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'

defineProps({
  embedded: { type: Boolean, default: false },
})

const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()
const route = useRoute()
const citedDoc = ref('')

function applyCitationQuery() {
  const raw = route.query.cited_doc
  citedDoc.value = typeof raw === 'string' ? raw : ''
  if (citedDoc.value) {
    moduleFilter.value = String(route.query.cited_type || 'field_guide')
    showAdvanced.value = true
    const base = citedDoc.value.split('/').pop() || citedDoc.value
    query.value = base.replace(/\.md$/i, '').replace(/[-_]/g, ' ')
  }
}

function onSearchFieldGuide({ citedDoc: doc }) {
  if (!doc) return
  citedDoc.value = doc
  moduleFilter.value = 'field_guide'
  showAdvanced.value = true
  const base = doc.split('/').pop() || doc
  query.value = base.replace(/\.md$/i, '').replace(/[-_]/g, ' ')
  void runSearch()
}

onMounted(() => {
  if (!capabilities.loaded) capabilities.fetch()
  applyCitationQuery()
})

watch(() => route.query.cited_doc, applyCitationQuery)

const query = ref('')
const showAdvanced = ref(false)
const moduleFilter = ref('')
const sinceIso = ref('')
const untilIso = ref('')
const limitN = ref(10)

const searchLoading = ref(false)
const searchFeedback = ref(null)
const results = ref([])
const embedModel = ref('')
const searchRanEmpty = ref(false)

const answerLoading = ref(false)
const answerFeedback = ref(null)
const answerText = ref(null)
const citations = ref([])
const llmModel = ref('')

function axiosRagFeedback(err, fallback) {
  const status = err.response?.status
  const serverMsg = err.response?.data?.error
  const msg =
    typeof serverMsg === 'string' && serverMsg.trim()
      ? serverMsg
      : fallback || err.message || 'Request failed'
  const degraded =
    status === 503 ||
    status === 429 ||
    status === 502 ||
    status === 504
  return { degraded, message: msg }
}

function payloadFilters() {
  const body = {}
  const q = query.value.trim()
  if (!q) return null
  body.query = q
  body.limit = Math.min(50, Math.max(1, Number(limitN.value) || 10))
  const m = moduleFilter.value.trim()
  if (m) body.module = m
  const s = sinceIso.value.trim()
  if (s) body.created_since = s
  const u = untilIso.value.trim()
  if (u) body.created_until = u
  return body
}

async function runSearch() {
  searchFeedback.value = null
  answerFeedback.value = null
  searchRanEmpty.value = false
  answerText.value = null
  citations.value = []
  const body = payloadFilters()
  if (!body || !farmContext.farmId) return
  searchLoading.value = true
  results.value = []
  embedModel.value = ''
  try {
    const { data } = await api.post(`/farms/${farmContext.farmId}/rag/search`, body)
    results.value = Array.isArray(data.results) ? data.results : []
    embedModel.value = data.model_id || ''
    searchRanEmpty.value = results.value.length === 0
  } catch (e) {
    searchFeedback.value = axiosRagFeedback(e, 'Search failed')
  } finally {
    searchLoading.value = false
  }
}

async function runAnswer() {
  answerFeedback.value = null
  answerText.value = null
  citations.value = []
  llmModel.value = ''
  const body = payloadFilters()
  if (!body || !farmContext.farmId) return
  answerLoading.value = true
  try {
    const { data } = await api.post(`/farms/${farmContext.farmId}/rag/answer`, {
      ...body,
      max_context_chunks: Math.min(15, Math.max(1, Number(limitN.value) || 8)),
    }, { timeout: 125000 })
    answerText.value = data.answer ?? ''
    citations.value = Array.isArray(data.citations) ? data.citations : []
    llmModel.value = data.llm_model || ''
    embedModel.value = data.embedding_model_id || embedModel.value
  } catch (e) {
    answerFeedback.value = axiosRagFeedback(e, 'Answer failed')
  } finally {
    answerLoading.value = false
  }
}

function formatDistance(d) {
  const n = Number(d)
  return Number.isFinite(n) ? n.toFixed(4) : String(d)
}
</script>
