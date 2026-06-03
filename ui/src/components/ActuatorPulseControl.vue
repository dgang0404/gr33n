<template>
  <div v-if="supportsPulse" class="flex flex-wrap items-center gap-2 mt-2 pt-2 border-t border-zinc-800">
    <span class="text-zinc-500 text-xs">Timed run</span>
    <input
      v-model.number="seconds"
      type="number"
      min="1"
      max="3600"
      class="w-16 bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
      title="Seconds to stay on, then auto-off"
    />
    <span class="text-zinc-600 text-xs">sec</span>
    <button
      type="button"
      class="px-2 py-1 rounded bg-green-800 hover:bg-green-700 text-white text-xs disabled:opacity-40"
      :disabled="busy || !seconds || seconds < 1"
      @click="runPulse"
    >
      {{ busy ? 'Running…' : 'Run pulse' }}
    </button>
    <p v-if="error" class="text-red-400 text-xs w-full">{{ error }}</p>
    <p v-else-if="ok" class="text-green-400 text-xs w-full">{{ ok }}</p>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { supportsPulseCommand } from '../lib/plantNeeds.js'
import { useFarmStore } from '../stores/farm.js'

const props = defineProps({
  actuator: { type: Object, required: true },
  defaultSeconds: { type: Number, default: 2 },
})

const store = useFarmStore()
const seconds = ref(props.defaultSeconds)
const busy = ref(false)
const error = ref('')
const ok = ref('')

const supportsPulse = computed(() => supportsPulseCommand(props.actuator?.actuator_type))

async function runPulse() {
  busy.value = true
  error.value = ''
  ok.value = ''
  try {
    await store.enqueueActuatorCommand(props.actuator.id, 'on', '', Math.round(seconds.value))
    ok.value = `Queued ${seconds.value}s pulse on ${props.actuator.name}`
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Pulse failed'
  } finally {
    busy.value = false
  }
}
</script>
