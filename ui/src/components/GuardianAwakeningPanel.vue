<template>
  <div
    v-if="visible"
    class="rounded-lg border px-3 py-2.5 text-xs space-y-2"
    :class="panelClass"
    role="status"
    data-test="guardian-awakening-panel"
  >
    <p class="font-medium" data-test="guardian-awakening-headline">{{ headline }}</p>
    <ul v-if="checklist.length" class="space-y-1 text-zinc-300" data-test="guardian-awakening-checklist">
      <li v-for="(item, i) in checklist" :key="i">{{ item }}</li>
    </ul>
    <p v-for="(msg, i) in messages" :key="'m'+i" class="text-zinc-400">{{ msg }}</p>
    <p v-if="readiness.error" class="text-red-300/90" data-test="guardian-awakening-error">{{ readiness.error }}</p>
    <div v-if="showQuickFallback" class="flex flex-wrap gap-2">
      <button
        type="button"
        class="px-2 py-1 rounded border border-zinc-600 bg-zinc-900 text-zinc-200 hover:bg-zinc-800"
        data-test="guardian-awakening-quick-fallback"
        @click="$emit('switch-quick')"
      >
        Try Quick chat
      </button>
      <button
        v-if="canRetry"
        type="button"
        class="px-2 py-1 rounded border border-green-800/80 bg-green-950/50 text-green-200 hover:bg-green-900/40"
        data-test="guardian-awakening-retry"
        @click="retry"
      >
        Retry awakening
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'

const props = defineProps({
  farmId: { type: [Number, String], default: null },
  mode: { type: String, default: 'farm_counsel' },
  autoWarm: { type: Boolean, default: true },
})

defineEmits(['switch-quick'])

const readiness = useGuardianReadinessStore()
const { awakening, isStirring } = storeToRefs(readiness)

const visible = computed(() => {
  const s = awakening.value?.state
  if (!s || s === 'ready') return false
  return s === 'sleeping' || s === 'stirring' || s === 'unavailable'
})

const headline = computed(() => {
  const s = awakening.value?.state
  if (s === 'stirring' || isStirring.value) return 'The Guardian is stirring…'
  if (s === 'sleeping') return 'The Guardian rests. Awakening…'
  if (s === 'unavailable') return 'Guardian unavailable'
  return 'Checking Guardian…'
})

const checklist = computed(() => {
  if (awakening.value?.state !== 'stirring' && !isStirring.value) return []
  const items = []
  if (props.mode === 'farm_counsel') {
    items.push(awakening.value?.rag_corpus_ok ? '☑ Field memories' : '☐ Field memories')
    items.push('☐ Live farm snapshot')
  }
  items.push(awakening.value?.chat_model_loaded ? '☑ Voice ready' : '☐ Voice model')
  return items
})

const messages = computed(() => {
  const msgs = [...(awakening.value?.messages || [])]
  if (awakening.value?.last_warmup_error) msgs.push(awakening.value.last_warmup_error)
  return msgs
})

const panelClass = computed(() => {
  const s = awakening.value?.state
  if (s === 'unavailable') return 'border-red-900/50 bg-red-950/20 text-red-100/90'
  if (s === 'stirring') return 'border-amber-900/50 bg-amber-950/25 text-amber-100/90'
  return 'border-zinc-700 bg-zinc-900/60 text-zinc-200'
})

const showQuickFallback = computed(() =>
  awakening.value?.state === 'unavailable' || !!awakening.value?.last_warmup_error,
)

const canRetry = computed(() => awakening.value?.state !== 'stirring')

async function boot() {
  await readiness.fetchHealth(props.farmId, props.mode)
  if (props.autoWarm && awakening.value?.state === 'sleeping') {
    await readiness.warmup(props.farmId, props.mode)
  } else if (isStirring.value) {
    readiness.startPolling(props.farmId, props.mode)
  }
}

function retry() {
  readiness.warmupStarted = false
  void readiness.warmup(props.farmId, props.mode)
}

watch(() => [props.farmId, props.mode], () => {
  void boot()
})

onMounted(() => {
  void boot()
})

onUnmounted(() => {
  readiness.stopPolling()
})
</script>
