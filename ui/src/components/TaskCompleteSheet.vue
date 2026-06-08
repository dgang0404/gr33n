<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
    data-test="task-complete-sheet"
    @click.self="onCancel"
  >
    <form
      class="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4"
      @submit.prevent="onSubmit"
    >
      <h2 class="text-sm font-semibold text-white">Mark task done</h2>
      <p class="text-xs text-zinc-400 truncate" :title="task?.title">{{ task?.title }}</p>

      <label class="flex items-center gap-2 text-sm text-zinc-300">
        <input v-model="logConsumption" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" data-test="task-complete-log-consumption" />
        Log supply used (optional)
      </label>

      <div v-if="logConsumption" class="space-y-3 pl-1 border-l-2 border-zinc-800">
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Batch</label>
          <select
            v-model="batchId"
            required
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="task-complete-batch-select"
          >
            <option value="">— Select batch —</option>
            <option v-for="b in batches" :key="b.id" :value="String(b.id)">
              {{ batchLabel(b) }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Quantity used</label>
          <input
            v-model.number="quantity"
            type="number"
            min="0"
            step="any"
            required
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="task-complete-qty"
          />
          <p v-if="selectedBatch" class="text-[10px] text-zinc-600 mt-1">
            On hand: {{ selectedBatch.current_quantity_remaining ?? '—' }}
          </p>
        </div>
        <p v-if="qtyError" class="text-xs text-red-400" data-test="task-complete-qty-error">{{ qtyError }}</p>
      </div>

      <p v-if="error" class="text-xs text-red-400">{{ error }}</p>

      <div class="flex gap-2 justify-end">
        <button type="button" class="text-xs px-3 py-1.5 rounded border border-zinc-700 text-zinc-400" @click="onCancel">
          Cancel
        </button>
        <button
          type="submit"
          class="text-xs px-3 py-1.5 rounded bg-green-900/60 border border-green-800 text-green-200 disabled:opacity-40"
          :disabled="submitting"
          data-test="task-complete-submit"
        >
          {{ submitting ? 'Saving…' : 'Mark done' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { validateConsumptionQty } from '../lib/taskConsumption.js'

const props = defineProps({
  open: { type: Boolean, default: false },
  task: { type: Object, default: null },
  batches: { type: Array, default: () => [] },
  inputs: { type: Array, default: () => [] },
})

const emit = defineEmits(['cancel', 'complete'])

const logConsumption = ref(false)
const batchId = ref('')
const quantity = ref(null)
const error = ref('')
const submitting = ref(false)

const inputById = computed(() => new Map((props.inputs || []).map((i) => [i.id, i])))

const selectedBatch = computed(() => {
  const id = Number(batchId.value)
  if (!id) return null
  return (props.batches || []).find((b) => b.id === id) || null
})

const qtyError = computed(() => {
  if (!logConsumption.value || !selectedBatch.value) return ''
  return validateConsumptionQty(quantity.value, selectedBatch.value)
})

watch(() => props.open, (isOpen) => {
  if (!isOpen) return
  logConsumption.value = false
  batchId.value = ''
  quantity.value = null
  error.value = ''
  submitting.value = false
})

function batchLabel(b) {
  const input = inputById.value.get(b.input_definition_id)
  const name = input?.name || `Input #${b.input_definition_id}`
  const rem = b.current_quantity_remaining ?? '—'
  return `${name} (${rem} on hand)`
}

function onCancel() {
  emit('cancel')
}

async function onSubmit() {
  if (!props.task) return
  error.value = ''
  if (logConsumption.value) {
    const qerr = qtyError.value
    if (qerr) {
      error.value = qerr
      return
    }
    const batch = selectedBatch.value
    const unitId = batch?.quantity_unit_id
    if (!batch || !unitId) {
      error.value = 'This batch has no quantity unit — set one in the full editor first.'
      return
    }
    submitting.value = true
    emit('complete', {
      task: props.task,
      consumption: {
        input_batch_id: batch.id,
        quantity: Number(quantity.value),
        unit_id: Number(unitId),
      },
    })
    submitting.value = false
    return
  }
  submitting.value = true
  emit('complete', { task: props.task, consumption: null })
  submitting.value = false
}
</script>
