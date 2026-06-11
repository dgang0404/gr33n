<template>
  <div
    v-if="recent?.prompt"
    class="rounded-lg border border-sky-800/50 bg-sky-950/30 px-3 py-2.5 flex flex-col gap-2"
    data-test="guardian-recent-topic-chip"
  >
    <p class="text-xs text-sky-100/90 leading-snug">{{ recent.prompt }}</p>
    <button
      type="button"
      class="self-start text-xs font-medium px-2.5 py-1 rounded bg-sky-900/60 border border-sky-700/80 text-sky-100 hover:bg-sky-900/80"
      data-test="guardian-recent-topic-continue"
      @click="onContinue"
    >
      Pick up where I left off
    </button>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import api from '../api'
import { buildContinueTopicPayload } from '../lib/guardianSessionMemory.js'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'

const props = defineProps({
  routePath: { type: String, default: '' },
})

const emit = defineEmits(['continue'])

const farmContext = useFarmContextStore()
const guardianPanel = useGuardianPanelStore()
const recent = ref(null)

async function loadRecent() {
  const farmId = farmContext.farmId
  const route = props.routePath || guardianPanel.routeRef?.path || ''
  if (!farmId || !route) {
    recent.value = null
    return
  }
  try {
    const r = await api.get(`/farms/${farmId}/guardian-memory/recent`, {
      params: { route },
      validateStatus: (s) => s === 200 || s === 204 || s === 404,
    })
    recent.value = r.status === 200 ? r.data : null
  } catch {
    recent.value = null
  }
}

function onContinue() {
  const payload = buildContinueTopicPayload(recent.value)
  if (!payload) return
  emit('continue', payload)
}

watch(
  () => [farmContext.farmId, props.routePath, guardianPanel.routeRef?.path],
  () => { void loadRecent() },
  { immediate: true },
)
</script>
