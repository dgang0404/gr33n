<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-2">Assign Relay Channels</h2>
      <p class="text-zinc-400">What pump or fan goes on each relay? Create actuators first if needed.</p>
    </section>

    <!-- Validation banner -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="text-xs font-semibold uppercase tracking-widest text-zinc-500 mb-3">
        Step Validation
      </div>
      <div class="space-y-1 text-sm">
        <div v-if="channelCount > 0" class="text-green-400">✓ {{ channelCount }} actuator(s) assigned</div>
        <div v-else class="text-red-400">✗ At least one actuator must be assigned</div>
      </div>
    </div>

    <!-- Channel grid -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
      <h3 class="text-sm font-semibold text-white mb-4">STACK 0 — Channels 0–7</h3>
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div
          v-for="channel in 8"
          :key="channel - 1"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 space-y-2"
        >
          <div class="text-xs text-zinc-500">Relay {{ channel }}</div>
          <div class="text-sm font-mono text-green-400">ch{{ channel - 1 }}</div>
          <select
            :value="assignments[channel - 1] || ''"
            @change="(e) => updateAssignment(channel - 1, e.target.value)"
            class="w-full text-xs rounded bg-zinc-900 border border-zinc-700 px-2 py-1 text-zinc-300 focus:outline-none focus:border-green-600"
          >
            <option value="">—</option>
            <option v-for="actuator in actuators" :key="actuator.id" :value="actuator.id">
              {{ actuator.name.substring(0, 20) }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Help -->
    <div class="bg-green-950/20 border border-green-900/50 rounded-xl p-4">
      <div class="text-xs font-semibold text-green-400 mb-2">💡 Channel Numbering</div>
      <p class="text-xs text-zinc-400">
        Channels 0–7 are on the first relay card (stack level 0). If you add more cards, they'll use channels 8–15, 16–23, etc.
        Each relay card must have its DIP switches set correctly before stacking.
      </p>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, watch } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'
import { validateStep3 } from '@/lib/piWizardValidation'

const wizard = usePiWizardStore()

// Mock actuators (in real app, fetch from API)
const actuators = reactive([
  { id: 1, name: 'Main Irrigation Pump' },
  { id: 2, name: 'Nutrient Pump A' },
  { id: 3, name: 'Nutrient Pump B' },
  { id: 4, name: 'Grow Lights' },
  { id: 5, name: 'Exhaust Fan' },
  { id: 6, name: 'Intake Fan' },
  { id: 7, name: 'Drain Pump' },
  { id: 8, name: 'Spare/Heater' },
])

const assignments = computed(() => wizard.formData.channelAssignments)
const channelCount = computed(() => Object.keys(assignments.value).length)

function updateAssignment(channel, actuatorId) {
  if (actuatorId === '') {
    wizard.updateChannelAssignment(channel, null)
  } else {
    wizard.updateChannelAssignment(channel, parseInt(actuatorId))
  }
}

watch(() => assignments.value, () => {
  const errors = validateStep3(wizard.formData)
  wizard.updateValidation(3, errors)
}, { immediate: true, deep: true })
</script>

<style scoped></style>
