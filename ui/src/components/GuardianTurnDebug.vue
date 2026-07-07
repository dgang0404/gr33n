<template>
  <details
    v-if="debug"
    class="rounded border border-violet-900/50 bg-violet-950/20 px-3 py-2 text-[10px] text-violet-100/90 font-mono"
    data-test="guardian-turn-debug"
  >
    <summary class="cursor-pointer text-violet-200/90 select-none">Last turn debug</summary>
    <dl class="mt-2 space-y-1">
      <div v-if="debug.request_id">
        <dt class="text-violet-400/80 inline">request_id:</dt>
        <dd class="inline ml-1">{{ debug.request_id }}</dd>
      </div>
      <div v-if="debug.tools_planned?.length">
        <dt class="text-violet-400/80">tools:</dt>
        <dd>{{ debug.tools_planned.join(', ') }}</dd>
      </div>
      <div v-if="ragLine">
        <dt class="text-violet-400/80 inline">rag_chunks:</dt>
        <dd class="inline ml-1">{{ ragLine }}</dd>
      </div>
      <div v-if="trimLine">
        <dt class="text-violet-400/80 inline">trim:</dt>
        <dd class="inline ml-1">{{ trimLine }}</dd>
      </div>
      <div v-if="debug.rag_filter_applied">
        <dt class="text-violet-400/80 inline">rag_filter:</dt>
        <dd class="inline ml-1">{{ debug.rag_filter_applied }}</dd>
      </div>
      <div v-if="debug.leak_trimmed">
        <dt class="text-violet-400/80 inline">leak_trim:</dt>
        <dd class="inline ml-1">removed {{ debug.leak_chars_removed }} chars</dd>
      </div>
      <div v-if="debug.meta_correction_trimmed">
        <dt class="text-violet-400/80 inline">meta_correction:</dt>
        <dd class="inline ml-1">removed {{ debug.meta_correction_chars_removed }} chars</dd>
      </div>
      <div v-if="debug.low_relevance">
        <dt class="text-violet-400/80 inline">relevance:</dt>
        <dd class="inline ml-1">
          low (q↔a {{ formatRel(debug.question_answer_relevance) }}, open↔tail {{ formatRel(debug.opening_tail_relevance) }}, min {{ formatRel(debug.relevance_min_threshold) }})
        </dd>
      </div>
      <div v-else-if="debug.question_answer_relevance != null">
        <dt class="text-violet-400/80 inline">relevance:</dt>
        <dd class="inline ml-1">
          q↔a {{ formatRel(debug.question_answer_relevance) }}, open↔tail {{ formatRel(debug.opening_tail_relevance) }}
        </dd>
      </div>
      <div v-if="debug.citation_urls_sanitized">
        <dt class="text-violet-400/80 inline">citation_urls:</dt>
        <dd class="inline ml-1">rewrote {{ debug.citation_links_rewritten }} fake links</dd>
      </div>
      <div v-if="modelLine">
        <dt class="text-violet-400/80 inline">model:</dt>
        <dd class="inline ml-1">{{ modelLine }}</dd>
      </div>
    </dl>
  </details>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  debug: { type: Object, default: null },
})

const ragLine = computed(() => {
  const d = props.debug
  if (!d?.rag_chunks && !d?.rag_chunk_total) return ''
  const parts = []
  if (d.rag_chunks) {
    for (const [k, v] of Object.entries(d.rag_chunks)) {
      parts.push(`${k}×${v}`)
    }
  }
  const total = d.rag_chunk_total ?? parts.reduce((s, p) => s + Number(p.split('×')[1] || 0), 0)
  return parts.length ? `${total} (${parts.join(', ')})` : String(total)
})

const trimLine = computed(() => {
  const t = props.debug?.trim_summary
  if (!t) return ''
  const bits = []
  if (t.history_turns) bits.push(`history ${t.history_turns}`)
  if (t.rag_top_k) bits.push(`rag ${t.rag_top_k}`)
  if (t.snapshot_reduced) bits.push('snapshot reduced')
  return bits.join(' · ')
})

const modelLine = computed(() => {
  const d = props.debug
  if (!d?.model) return ''
  const eff = d.effective_context_window
  return eff ? `${d.model} · ${eff} effective ctx` : d.model
})

function formatRel(v) {
  if (v == null || Number.isNaN(Number(v))) return '—'
  return Number(v).toFixed(2)
}
</script>
