<template>
  <div
    v-if="open && cycle"
    class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
    data-test="harvest-weigh-in"
    @click.self="close"
  >
    <div class="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4">
      <div>
        <h2 class="text-white font-semibold">Harvest weigh-in</h2>
        <p class="text-zinc-500 text-xs mt-0.5">
          Record yield for <span class="text-zinc-300">{{ cycle.name }}</span> and close this grow.
        </p>
      </div>

      <GuardianStarterChips
        v-if="harvestStarters.length"
        :starters="harvestStarters"
        data-test="harvest-flow-starters"
      />

      <div>
        <label class="block text-xs text-zinc-500 mb-1">Yield (grams)</label>
        <input
          v-model.number="yieldGrams"
          type="number"
          min="0"
          step="0.1"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white font-mono"
          placeholder="e.g. 420"
          data-test="harvest-yield-grams"
        />
      </div>

      <div>
        <label class="block text-xs text-zinc-500 mb-1">Harvest date</label>
        <input
          v-model="harvestedAt"
          type="date"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          data-test="harvest-date"
        />
      </div>

      <div>
        <label class="block text-xs text-zinc-500 mb-1">Notes (optional)</label>
        <textarea
          v-model="yieldNotes"
          rows="2"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          placeholder="Wet weight, trim notes, etc."
          data-test="harvest-notes"
        />
      </div>

      <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>

      <div class="flex justify-end gap-3 pt-1">
        <button
          type="button"
          class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          @click="close"
        >
          Cancel
        </button>
        <button
          type="button"
          class="px-4 py-1.5 text-xs rounded-lg bg-amber-700 hover:bg-amber-600 text-white font-medium disabled:opacity-40"
          :disabled="submitting"
          data-test="harvest-submit"
          @click="submit"
        >
          {{ submitting ? 'Saving…' : 'Finish harvest' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { buildHarvestPayload } from '../lib/growHub.js'
import { buildHarvestFlowStarters } from '../lib/guardianStarters.js'
import GuardianStarterChips from './GuardianStarterChips.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  cycle: { type: Object, default: null },
  zone: { type: Object, default: null },
  priorHarvestedCycle: { type: Object, default: null },
})

const harvestStarters = computed(() => buildHarvestFlowStarters({
  zone: props.zone,
  activeCycle: props.cycle,
  priorHarvestedCycle: props.priorHarvestedCycle,
}))

const emit = defineEmits(['close', 'harvested'])

const store = useFarmStore()
const submitting = ref(false)
const formError = ref('')
const yieldGrams = ref(null)
const yieldNotes = ref('')
const harvestedAt = ref(new Date().toISOString().slice(0, 10))

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    formError.value = ''
    yieldGrams.value = props.cycle?.yield_grams != null ? Number(props.cycle.yield_grams) : null
    yieldNotes.value = props.cycle?.yield_notes || ''
    harvestedAt.value = new Date().toISOString().slice(0, 10)
  },
)

function close() {
  emit('close')
}

async function submit() {
  if (!props.cycle) return
  formError.value = ''
  submitting.value = true
  try {
    const payload = buildHarvestPayload(props.cycle, {
      yieldGrams: yieldGrams.value,
      yieldNotes: yieldNotes.value,
      harvestedAt: harvestedAt.value,
    })
    const updated = await store.updateCropCycle(props.cycle.id, payload)
    emit('harvested', updated)
    close()
  } catch (e) {
    formError.value = e?.response?.data?.error || e?.message || 'Failed to save harvest'
  } finally {
    submitting.value = false
  }
}
</script>
