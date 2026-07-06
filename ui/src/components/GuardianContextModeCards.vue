<template>
  <div class="space-y-2" data-test="guardian-context-mode-cards">
    <p class="text-[10px] uppercase tracking-widest text-zinc-500">How should Guardian answer?</p>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
      <button
        type="button"
        class="text-left rounded-lg border px-3 py-2.5 transition-colors"
        :class="quickSelected ? 'border-green-700 bg-green-950/40 ring-1 ring-green-800/60' : 'border-zinc-700 bg-zinc-950 hover:border-zinc-500'"
        data-test="guardian-mode-quick"
        @click="selectQuick"
      >
        <span class="block text-sm font-medium text-zinc-100">Quick chat</span>
        <span class="block text-[11px] text-zinc-400 mt-0.5">Fast — general horticulture, no farm data</span>
        <span class="inline-flex mt-2 text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Chat only</span>
      </button>
      <button
        type="button"
        class="text-left rounded-lg border px-3 py-2.5 transition-colors"
        :class="farmSelected ? 'border-green-700 bg-green-950/40 ring-1 ring-green-800/60' : 'border-zinc-700 bg-zinc-950 hover:border-zinc-500'"
        :disabled="farmDisabled"
        data-test="guardian-mode-farm-counsel"
        @click="selectFarmCounsel"
      >
        <span class="block text-sm font-medium text-zinc-100">Farm counsel</span>
        <span class="block text-[11px] text-zinc-400 mt-0.5">Reads your farm — may take many minutes on CPU</span>
        <span class="flex flex-wrap gap-1 mt-2">
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Snapshot</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Read tools</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">RAG</span>
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300">Chat</span>
        </span>
        <span v-if="!farmContext.farmId" class="block text-amber-300/80 text-[11px] mt-1">Select a farm in the sidebar first.</span>
        <span v-else-if="ragWarning" class="block text-amber-300/80 text-[11px] mt-1">Field memories not ingested — run bootstrap.</span>
      </button>
    </div>
    <p class="text-[10px] text-zinc-500">
      Cold models and CPU timing — see docs/connectivity-requirements.md in the repo.
    </p>
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'

const useFarmContext = defineModel('useFarmContext', { type: Boolean, default: false })

const farmContext = useFarmContextStore()
const readiness = useGuardianReadinessStore()
const { awakening } = storeToRefs(readiness)

const quickSelected = computed(() => !useFarmContext.value)
const farmSelected = computed(() => useFarmContext.value)
const farmDisabled = computed(() => !farmContext.farmId)
const ragWarning = computed(() =>
  useFarmContext.value && awakening.value && awakening.value.rag_corpus_ok === false,
)

function selectQuick() {
  useFarmContext.value = false
  void readiness.ensureAwake(null, 'quick')
}

function selectFarmCounsel() {
  if (!farmContext.farmId) return
  useFarmContext.value = true
  void readiness.ensureAwake(farmContext.farmId, 'farm_counsel')
}

watch(
  () => useFarmContext.value,
  (on) => {
    if (on && farmContext.farmId) {
      void readiness.ensureAwake(farmContext.farmId, 'farm_counsel')
    } else if (!on) {
      void readiness.ensureAwake(null, 'quick')
    }
  },
)
</script>
