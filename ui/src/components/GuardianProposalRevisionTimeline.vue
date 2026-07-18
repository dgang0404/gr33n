<template>
  <div
    class="text-[11px] text-violet-200/90 bg-violet-950/15 border border-violet-900/40 rounded-md overflow-hidden"
    data-test="guardian-proposal-revision-timeline"
  >
    <button
      type="button"
      class="w-full flex items-center gap-2 px-2.5 py-1.5 text-left hover:bg-violet-950/30"
      :aria-expanded="expanded"
      data-test="guardian-proposal-revision-timeline-toggle"
      @click="toggleExpanded"
    >
      <span class="text-violet-300/90 shrink-0">{{ expanded ? '▾' : '▸' }}</span>
      <span class="font-medium text-violet-200/95">{{ headerLabel }}</span>
      <span v-if="loading" class="text-[10px] text-violet-400/70 ml-auto">Loading…</span>
    </button>

    <div
      v-if="expanded"
      class="border-t border-violet-900/40 px-2.5 py-1.5 space-y-1.5"
      data-test="guardian-proposal-revision-timeline-body"
    >
      <p v-if="error" class="text-red-400/90 text-[10px]" data-test="guardian-proposal-revision-timeline-error">
        {{ error }}
      </p>
      <ol v-else-if="entries.length" class="space-y-1.5">
        <li
          v-for="entry in entries"
          :key="entry.index"
          class="space-y-0.5"
          data-test="guardian-proposal-revision-timeline-entry"
        >
          <p class="text-zinc-200">
            <span class="text-violet-400/80 shrink-0">{{ entry.index }}.</span>
            <span class="text-zinc-400"> You: </span>
            <span>{{ entry.userMessage }}</span>
          </p>
          <p v-if="entry.cue" class="text-[10px] text-violet-300/80 pl-4 font-mono">→ {{ entry.cue }}</p>
        </li>
      </ol>
      <p v-else-if="!loading" class="text-zinc-500 italic text-[10px]">No chat turns found for this session.</p>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import {
  buildRevisionTimeline,
  revisionTimelineLabel,
} from '../lib/guardianRevisionTimeline.js'

const props = defineProps({
  sessionId: { type: String, required: true },
  revision: { type: Number, default: 1 },
  tool: { type: String, default: '' },
})

const expanded = ref(false)
const loading = ref(false)
const error = ref('')
const entries = ref([])
const loadedSessionId = ref('')

const headerLabel = computed(() => revisionTimelineLabel(props.revision, entries.value.length))

async function loadTurns() {
  if (!props.sessionId || loading.value) return
  if (loadedSessionId.value === props.sessionId && entries.value.length) return

  loading.value = true
  error.value = ''
  try {
    const r = await api.get(`/v1/chat/sessions/${props.sessionId}`)
    const turns = Array.isArray(r.data?.turns) ? r.data.turns : []
    entries.value = buildRevisionTimeline(turns, { tool: props.tool })
    loadedSessionId.value = props.sessionId
  } catch (e) {
    entries.value = []
    error.value = e?.response?.data?.error || e.message || 'Failed to load revision history'
  } finally {
    loading.value = false
  }
}

async function toggleExpanded() {
  expanded.value = !expanded.value
  if (expanded.value) await loadTurns()
}

watch(
  () => props.sessionId,
  () => {
    loadedSessionId.value = ''
    entries.value = []
    error.value = ''
    if (expanded.value) void loadTurns()
  },
)
</script>
