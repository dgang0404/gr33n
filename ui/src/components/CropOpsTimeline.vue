<template>
  <section
    id="crop-ops-timeline"
    data-test="crop-ops-timeline"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5"
  >
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3 mb-4">
      <div>
        <h2 class="text-white text-sm font-semibold flex items-center gap-2">
          📋 Ops log
          <HelpTip>
            Feed, mix, light, and stage events for this grow — including the recipe formula pinned at mix and program run time.
          </HelpTip>
        </h2>
        <p class="text-zinc-500 text-xs mt-1">
          What was this room getting? Filter by date or scroll the full grow window.
        </p>
      </div>
      <div class="flex flex-wrap items-end gap-2 shrink-0">
        <label class="text-[11px] text-zinc-500">
          From
          <input
            v-model="fromInput"
            type="date"
            class="block mt-0.5 text-xs bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-200"
            data-test="crop-ops-from"
          />
        </label>
        <label class="text-[11px] text-zinc-500">
          To
          <input
            v-model="toInput"
            type="date"
            class="block mt-0.5 text-xs bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-zinc-200"
            data-test="crop-ops-to"
          />
        </label>
        <button
          type="button"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-900 text-zinc-300 border border-zinc-700 hover:bg-zinc-800"
          data-test="crop-ops-refresh"
          @click="refresh"
        >
          Refresh
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading ops log…</div>
    <p v-else-if="error" class="text-red-400 text-sm" data-test="crop-ops-error">{{ error }}</p>
    <p v-else-if="!events.length" class="text-zinc-500 text-xs" data-test="crop-ops-empty">
      No ops events in this date range yet.
    </p>
    <ol v-else class="border-l border-zinc-700 ml-2 space-y-3" data-test="crop-ops-list">
      <li
        v-for="ev in events"
        :key="`${ev.kind}-${ev.id}`"
        class="pl-4 relative"
        data-test="crop-ops-row"
        :data-kind="ev.kind"
      >
        <span
          class="absolute -left-1.5 top-1.5 w-3 h-3 rounded-full border"
          :class="cropOpsKindClass(ev.kind)"
        />
        <div class="flex flex-wrap items-center gap-2">
          <span
            class="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded border"
            :class="cropOpsKindClass(ev.kind)"
          >
            {{ cropOpsKindLabel(ev.kind) }}
          </span>
          <span class="text-zinc-500 text-[11px]">{{ formatCropOpsWhen(ev.occurred_at) }}</span>
        </div>
        <p class="text-white text-sm mt-0.5">{{ ev.summary || cropOpsKindLabel(ev.kind) }}</p>
        <p v-if="cropOpsEventSubtitle(ev)" class="text-zinc-500 text-xs">{{ cropOpsEventSubtitle(ev) }}</p>
        <div
          v-if="cropOpsEventHasFormula(ev.details)"
          class="mt-1.5 text-[11px] bg-zinc-900/80 border border-zinc-700 rounded-md px-2.5 py-1.5 space-y-0.5"
          data-test="crop-ops-formula"
        >
          <p class="text-zinc-500 uppercase tracking-wide text-[10px]">Formula at time</p>
          <p v-for="(line, idx) in formulaSnapshotLines(ev.details)" :key="idx" class="text-zinc-300 font-mono">
            {{ line }}
          </p>
        </div>
      </li>
    </ol>
  </section>
</template>

<script setup>
import { ref, watch } from 'vue'
import HelpTip from './HelpTip.vue'
import { useFarmStore } from '../stores/farm'
import {
  cropOpsEventHasFormula,
  cropOpsEventSubtitle,
  cropOpsKindClass,
  cropOpsKindLabel,
  formatCropOpsDateInput,
  formatCropOpsWhen,
  formulaSnapshotLines,
} from '../lib/cropOpsTimeline.js'

const props = defineProps({
  farmId: { type: Number, required: true },
  cycleId: { type: Number, required: true },
  defaultFrom: { type: String, default: '' },
  defaultTo: { type: String, default: '' },
})

const store = useFarmStore()

const loading = ref(false)
const error = ref('')
const events = ref([])
const fromInput = ref('')
const toInput = ref('')

function seedRangeInputs() {
  if (!fromInput.value && props.defaultFrom) fromInput.value = formatCropOpsDateInput(props.defaultFrom)
  if (!toInput.value && props.defaultTo) toInput.value = formatCropOpsDateInput(props.defaultTo)
}

async function refresh() {
  if (!props.farmId || !props.cycleId) return
  loading.value = true
  error.value = ''
  try {
    const data = await store.loadCropCycleOpsTimeline(props.farmId, props.cycleId, {
      from: fromInput.value || undefined,
      to: toInput.value || undefined,
    })
    events.value = Array.isArray(data?.events) ? data.events : []
    if (data?.from) fromInput.value = formatCropOpsDateInput(data.from)
    if (data?.to) toInput.value = formatCropOpsDateInput(data.to)
  } catch (err) {
    events.value = []
    error.value = err?.response?.data?.error || err?.message || 'Failed to load ops log'
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.farmId, props.cycleId, props.defaultFrom, props.defaultTo],
  () => {
    seedRangeInputs()
    void refresh()
  },
  { immediate: true },
)
</script>
