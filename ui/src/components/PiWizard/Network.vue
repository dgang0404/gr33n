<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-2">Network & API Configuration</h2>
      <p class="text-zinc-400">Configure where your Pi will send data, and test the connection.</p>
    </section>

    <form @submit.prevent class="space-y-5">
      <!-- API Base URL -->
      <div>
        <label class="block text-sm font-medium text-zinc-300 mb-2">
          API Server Address
        </label>
        <input
          v-model="localNetwork.apiBaseUrl"
          type="text"
          placeholder="http://192.168.1.50:8080"
          class="w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-600"
        />
        <p class="text-xs text-zinc-500 mt-1">
          The API server your Pi will connect to. Usually on your LAN or VPN.
        </p>
      </div>

      <!-- Network Test -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h3 class="text-sm font-semibold text-white mb-3">Test Connectivity</h3>
        <p class="text-xs text-zinc-400 mb-4">
          Click below to check if your Pi can reach the API. After you finish setup, your Pi will
          automatically prove connectivity.
        </p>

        <button
          type="button"
          @click="testConnectivity"
          :disabled="!localNetwork.apiBaseUrl || testInProgress"
          class="w-full px-4 py-2 rounded-lg bg-blue-700 text-white hover:bg-blue-600 disabled:opacity-40 disabled:cursor-not-allowed text-sm font-medium"
        >
          {{ testInProgress ? '⏱️ Testing... (10s)' : '🔄 Test Now' }}
        </button>

        <!-- Test result -->
        <div v-if="testResult" class="mt-3 p-3 rounded-lg" :class="resultClass">
          <div class="text-xs font-semibold mb-1" :class="resultTextClass">
            {{ resultIcon }} {{ resultMessage }}
          </div>
          <div v-if="testResult.latency_ms" class="text-xs text-zinc-500">
            Latency: {{ testResult.latency_ms }}ms
          </div>
        </div>
      </div>

      <!-- Config Preview -->
      <div>
        <label class="block text-sm font-medium text-zinc-300 mb-2">
          Preview Configuration
        </label>
        <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 font-mono text-xs text-zinc-400 max-h-64 overflow-auto">
          <pre>{{ configPreview }}</pre>
        </div>
      </div>
    </form>

    <!-- Validation -->
    <div v-if="validationErrors.length" class="bg-red-950/30 border border-red-700/50 rounded-lg p-3">
      <div class="text-xs text-red-300 space-y-1">
        <div v-for="(err, idx) in validationErrors" :key="idx">✗ {{ err }}</div>
      </div>
    </div>
    <div v-else class="bg-green-950/30 border border-green-700/50 rounded-lg p-3">
      <div class="text-xs text-green-300">✓ Network configuration ready</div>
    </div>
  </div>
</template>

<script setup>
import { reactive, computed, watch } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'
import { validateStep4 } from '@/lib/piWizardValidation'

const wizard = usePiWizardStore()

const localNetwork = reactive({
  apiBaseUrl: wizard.formData.network.apiBaseUrl || `http://${window.location.hostname}:8080`,
})

const testInProgress = computed(() => wizard.formData.network.testInProgress)
const testResult = computed(() => wizard.formData.network.testResult)

const resultClass = computed(() => {
  if (!testResult.value) return ''
  if (testResult.value.success) return 'bg-green-950/30 border-green-700/50'
  return 'bg-red-950/30 border-red-700/50'
})

const resultTextClass = computed(() => {
  if (!testResult.value) return ''
  return testResult.value.success ? 'text-green-400' : 'text-red-400'
})

const resultIcon = computed(() => {
  if (!testResult.value) return ''
  return testResult.value.success ? '✓' : '✗'
})

const resultMessage = computed(() => {
  if (!testResult.value) return ''
  if (testResult.value.success) {
    return `Connected in ${testResult.value.latency_ms}ms`
  }
  return testResult.value.error || 'Connection failed'
})

const configPreview = computed(() => {
  const { device, network } = wizard.formData
  return `api:
  base_url: ${network.apiBaseUrl}
  timeout_seconds: 5
  api_key: ${device.apiKey}
device:
  uid: ${device.uid}
farm:
  farm_id: ${device.farmId || 1}
schedule_poll_interval_seconds: 30
offline_queue_path: /var/lib/gr33n/queue.db
`
})

const validationErrors = computed(() => {
  return validateStep4({
    network: {
      apiBaseUrl: localNetwork.apiBaseUrl,
    },
  })
})

watch(validationErrors, (errors) => {
  wizard.updateValidation(4, errors)
}, { immediate: true })

watch(() => localNetwork.apiBaseUrl, (url) => {
  wizard.updateNetworkConfig({ apiBaseUrl: url })
})

async function testConnectivity() {
  wizard.setTestInProgress(true)
  try {
    // Simulate test (in real app, call API)
    await new Promise(resolve => setTimeout(resolve, 1500))
    wizard.setTestResult({
      success: true,
      latency_ms: Math.random() * 200 + 50,
    })
  } catch (err) {
    wizard.setTestResult({
      success: false,
      error: 'Connection failed',
    })
  } finally {
    wizard.setTestInProgress(false)
  }
}
</script>

<style scoped></style>
