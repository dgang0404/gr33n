<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-2">Ready to Deploy 🚀</h2>
      <p class="text-zinc-400">Final checklist before your Pi goes live.</p>
    </section>

    <!-- Pre-deployment checklist -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-3">
      <div class="text-sm font-semibold text-zinc-300 mb-4">Before you hit Finish:</div>

      <label class="flex items-start gap-3 cursor-pointer">
        <input
          v-model="checklist.configLoaded"
          type="checkbox"
          class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
        />
        <div>
          <div class="text-sm text-zinc-200">Configuration file loaded on Pi</div>
          <div class="text-xs text-zinc-500">config.yaml is at ~/.gr33n/config.yaml</div>
        </div>
      </label>

      <label class="flex items-start gap-3 cursor-pointer">
        <input
          v-model="checklist.hardwareSeated"
          type="checkbox"
          class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
        />
        <div>
          <div class="text-sm text-zinc-200">Relay HAT is stacked and powered</div>
          <div class="text-xs text-zinc-500">All relays should click when Pi boots</div>
        </div>
      </label>

      <label class="flex items-start gap-3 cursor-pointer">
        <input
          v-model="checklist.actuatorsWired"
          type="checkbox"
          class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
        />
        <div>
          <div class="text-sm text-zinc-200">Actuators wired to correct relays</div>
          <div class="text-xs text-zinc-500">Pumps, fans, lights all connected</div>
        </div>
      </label>

      <label class="flex items-start gap-3 cursor-pointer">
        <input
          v-model="checklist.networkOnline"
          type="checkbox"
          class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
        />
        <div>
          <div class="text-sm text-zinc-200">Pi is online and visible on network</div>
          <div class="text-xs text-zinc-500">Ping it to confirm: ping {{ piHostname }}</div>
        </div>
      </label>
    </div>

    <!-- API configuration summary -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
      <div class="text-sm font-semibold text-green-400 mb-3">Configuration Summary</div>
      <dl class="text-xs text-zinc-400 space-y-2">
        <div class="flex justify-between">
          <dt>Device Name:</dt>
          <dd class="text-zinc-200">{{ wizard.formData.device.name }}</dd>
        </div>
        <div class="flex justify-between">
          <dt>Device UID:</dt>
          <dd class="text-zinc-200">{{ wizard.formData.device.uid }}</dd>
        </div>
        <div class="flex justify-between">
          <dt>API Server:</dt>
          <dd class="text-zinc-200">{{ wizard.formData.network.apiBaseUrl }}</dd>
        </div>
        <div class="flex justify-between">
          <dt>Relays Configured:</dt>
          <dd class="text-zinc-200">{{ channelCount }}/8</dd>
        </div>
      </dl>
    </div>

    <!-- Optional: Test a relay -->
    <div class="bg-green-950/20 border border-green-900/50 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-green-400 mb-3">Optional: Test a Relay</h3>
      <p class="text-xs text-zinc-400 mb-3">
        Send a test pulse to one relay to verify wiring is correct:
      </p>
      <div class="flex gap-2">
        <select
          v-model="selectedRelay"
          class="flex-1 text-xs rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-zinc-300 focus:outline-none focus:border-green-600"
        >
          <option v-for="i in 8" :key="i - 1">Test Relay {{ i }}</option>
        </select>
        <button
          type="button"
          @click="testRelay"
          class="px-3 py-2 rounded-lg bg-green-700 text-white hover:bg-green-600 text-xs font-medium"
        >
          Pulse (1s)
        </button>
      </div>
    </div>

    <!-- Final validation -->
    <div v-if="allChecked" class="bg-green-950/30 border border-green-700/50 rounded-lg p-3">
      <div class="text-xs text-green-300">✓ All pre-checks complete. Ready to finish!</div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, ref } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'

const wizard = usePiWizardStore()

const checklist = reactive({
  configLoaded: false,
  hardwareSeated: false,
  actuatorsWired: false,
  networkOnline: false,
})

const selectedRelay = ref('Test Relay 1')

const piHostname = computed(() => {
  return wizard.formData.device.uid || 'pi.local'
})

const channelCount = computed(() => {
  return Object.keys(wizard.formData.channelAssignments).length
})

const allChecked = computed(() => {
  return Object.values(checklist).every(v => v === true)
})

async function testRelay() {
  // TODO: Call API to test relay
  console.log('Pulse relay:', selectedRelay.value)
}
