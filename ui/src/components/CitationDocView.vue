<template>
  <article
    class="rounded-xl border border-green-900/50 bg-zinc-900 p-5 space-y-4"
    data-test="citation-doc-view"
  >
    <header class="space-y-2">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div class="space-y-1 min-w-0">
          <p class="text-[11px] uppercase tracking-widest text-green-500/80">
            {{ citationTypeLabel(docType) }}
          </p>
          <h2 class="text-lg font-semibold text-green-300 leading-snug" data-test="citation-doc-title">
            {{ title }}
          </h2>
          <p class="text-xs text-zinc-500 font-mono truncate" data-test="citation-doc-path">{{ docPath }}</p>
        </div>
        <button
          v-if="dismissible"
          type="button"
          class="text-xs text-zinc-500 hover:text-zinc-300 shrink-0"
          data-test="citation-doc-dismiss"
          @click="emit('dismiss')"
        >
          Close doc view
        </button>
      </div>
      <p v-if="highlightChunkId" class="text-xs text-amber-200/90">
        Highlighted section is what Guardian cited in your answer.
      </p>
    </header>

    <div v-if="loading" class="text-sm text-zinc-500">Loading document…</div>
    <div v-else-if="error" class="text-sm text-red-400">{{ error }}</div>
    <div
      v-else-if="chunks.length"
      class="space-y-3 max-h-[min(60vh,28rem)] overflow-y-auto pr-1"
      data-test="citation-doc-chunks"
    >
      <section
        v-for="chunk in chunks"
        :key="chunk.id"
        :ref="(el) => setChunkRef(chunk.id, el)"
        class="rounded-lg border px-4 py-3 transition-colors"
        :class="chunk.id === highlightChunkId
          ? 'border-amber-700/80 bg-amber-950/30 ring-1 ring-amber-800/60'
          : 'border-zinc-800 bg-zinc-950/50'"
        :data-test="chunk.id === highlightChunkId ? 'citation-doc-chunk-highlight' : `citation-doc-chunk-${chunk.chunk_index}`"
      >
        <p class="text-[10px] uppercase tracking-widest text-zinc-600 mb-2">
          Section {{ chunk.chunk_index + 1 }}
          <span v-if="chunk.id === highlightChunkId" class="text-amber-400/90"> · cited</span>
        </p>
        <pre class="text-sm text-zinc-200 whitespace-pre-wrap font-sans leading-relaxed">{{ chunkDisplayText(chunk.content_text) }}</pre>
      </section>
    </div>
    <p v-else class="text-sm text-zinc-500">
      No indexed chunks found for this document on the selected farm. Try re-ingesting field guides in Settings.
    </p>

    <div class="flex flex-wrap gap-2 pt-1 border-t border-zinc-800">
      <button
        type="button"
        class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        data-test="citation-doc-ask-guardian"
        @click="askGuardian"
      >
        Ask Guardian about this
      </button>
      <router-link
        v-if="browseLink"
        :to="browseLink"
        class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-400 hover:text-zinc-200"
        data-test="citation-doc-browse-link"
      >
        Browse all field guides
      </router-link>
    </div>
  </article>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import {
  chunkDisplayText,
  citationTypeLabel,
  fieldGuideSlugFromDocPath,
  guardianDocPrefill,
  humanDocTitle,
} from '../lib/citationDoc.js'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'

const props = defineProps({
  docPath: { type: String, required: true },
  docType: { type: String, default: 'field_guide' },
  highlightChunkId: { type: Number, default: 0 },
  dismissible: { type: Boolean, default: true },
})

const emit = defineEmits(['dismiss'])

const router = useRouter()
const farmContext = useFarmContextStore()
const guardianPanel = useGuardianPanelStore()

const loading = ref(false)
const error = ref('')
const chunks = ref([])
const title = ref('')
const chunkRefs = ref({})

const browseLink = computed(() => {
  if (props.docType !== 'field_guide') return null
  return { path: '/operator-guide', query: { tab: 'knowledge' } }
})

function setChunkRef(id, el) {
  if (el) chunkRefs.value[id] = el
}

async function loadFieldGuideTitle() {
  if (props.docType !== 'field_guide') return
  const slug = fieldGuideSlugFromDocPath(props.docPath)
  if (!slug) return
  try {
    const { data } = await api.get(`/commons/agronomy-field-guides/${encodeURIComponent(slug)}`)
    if (data?.title) title.value = data.title
  } catch {
    // fall back to path-derived title
  }
}

async function loadChunks() {
  if (!props.docPath || !farmContext.farmId) return
  loading.value = true
  error.value = ''
  chunks.value = []
  try {
    const { data } = await api.get(`/farms/${farmContext.farmId}/rag/docs`, {
      params: { doc_path: props.docPath },
    })
    chunks.value = Array.isArray(data?.chunks) ? data.chunks : []
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load document'
  } finally {
    loading.value = false
  }
}

async function scrollToHighlight() {
  if (!props.highlightChunkId) return
  await nextTick()
  const el = chunkRefs.value[props.highlightChunkId]
  el?.scrollIntoView?.({ behavior: 'smooth', block: 'center' })
}

async function loadDoc() {
  title.value = humanDocTitle(props.docPath)
  await Promise.all([loadFieldGuideTitle(), loadChunks()])
  await scrollToHighlight()
}

function askGuardian() {
  const query = {
    cited_doc: props.docPath,
    cited_type: props.docType,
  }
  if (props.highlightChunkId > 0) {
    query.cited_chunk = String(props.highlightChunkId)
  }
  guardianPanel.prefilledMessage = guardianDocPrefill(title.value)
  router.push({ path: '/chat', query })
}

onMounted(loadDoc)

watch(
  () => [props.docPath, props.highlightChunkId, farmContext.farmId],
  () => { void loadDoc() },
)
</script>
