<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-2">Register your Pi on the Farm</h2>
      <p class="text-zinc-400">Give your Pi a name and let gr33n generate a secure API key for it.</p>
    </section>

    <form @submit.prevent="handleSubmit" class="space-y-5">
      <!-- Device Name -->
      <div>
        <label class="block text-sm font-medium text-zinc-300 mb-2">
          Device Name (label only)
        </label>
        <input
          v-model="localForm.name"
          type="text"
          placeholder="e.g., Flower Room Pi"
          class="w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-600"
        />
        <p class="text-xs text-zinc-500 mt-1">Human-readable label for the dashboard</p>
      </div>

      <!-- Device UID -->
      <div>
        <label class="block text-sm font-medium text-zinc-300 mb-2">
          Device UID (unique identifier)
        </label>
        <input
          v-model="localForm.uid"
          type="text"
          placeholder="e.g., flower-room-01"
          class="w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-600"
        />
        <p class="text-xs text-zinc-500 mt-1">
          Unique identifier for this Pi. Can be hostname, MAC address, or custom name.
        </p>
      </div>

      <!-- API Key Section -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h3 class="text-sm font-semibold text-white mb-3">🔐 Generate API Key</h3>
        <p class="text-sm text-zinc-400 mb-4">
          ⚠️ This is a secret. Copy it NOW and save it somewhere safe. You won't see it again!
        </p>

        <div v-if="apiKey" class="space-y-3">
          <div class="bg-zinc-950 border border-green-700/50 rounded-lg p-3">
            <div class="font-mono text-xs text-green-300 break-all">{{ apiKey }}</div>
          </div>
          <div class="flex gap-2 flex-wrap">
            <button
              type="button"
              @click="copyApiKey"
              class="text-xs px-3 py-1.5 rounded-lg bg-green-700 text-white hover:bg-green-600"
            >
              📋 Copy
            </button>
            <button
              type="button"
              @click="downloadApiKey"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:text-white"
            >
              ⬇️ Download
            </button>
            <button
              type="button"
              @click="regenerateApiKey"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:text-white"
            >
              🔄 Regenerate
            </button>
          </div>
        </div>

        <button
          v-else
          type="button"
          @click="generateApiKey"
          :disabled="!localForm.uid"
          class="w-full px-4 py-2 rounded-lg bg-green-700 text-white hover:bg-green-600 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Generate API Key
        </button>
      </div>

      <!-- Farm assignment (optional) -->
      <div>
        <label class="block text-sm font-medium text-zinc-300 mb-2">
          Farm (optional — set to default)
        </label>
        <select
          v-model.number="localForm.farmId"
          class="w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        >
          <option :value="null">Auto-detect or use default</option>
          <option :value="1">Farm 1 (default)</option>
        </select>
      </div>

      <!-- Status -->
      <div v-if="validationErrors.length" class="bg-red-950/30 border border-red-700/50 rounded-lg p-3">
        <div class="text-xs text-red-300 space-y-1">
          <div v-for="(err, idx) in validationErrors" :key="idx">✗ {{ err }}</div>
        </div>
      </div>

      <div v-else-if="apiKey" class="bg-green-950/30 border border-green-700/50 rounded-lg p-3">
        <div class="text-xs text-green-300">✓ Pi registered and ready for next step</div>
      </div>
    </form>
  </div>
</template>

<script setup>
import { reactive, computed, watch } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'
import { validateStep2 } from '@/lib/piWizardValidation'

const wizard = usePiWizardStore()

const localForm = reactive({
  name: wizard.formData.device.name || '',
  uid: wizard.formData.device.uid || '',
  farmId: wizard.formData.device.farmId || null,
})

const apiKey = computed(() => wizard.formData.device.apiKey)

const validationErrors = computed(() => {
  return validateStep2({
    device: {
      name: localForm.name,
      uid: localForm.uid,
      apiKey: apiKey.value,
    },
  })
})

watch(() => validationErrors.value, (errors) => {
  wizard.updateValidation(2, errors)
}, { immediate: true })

function generateApiKey() {
  const newKey = `gdev_${localForm.uid}_${generateRandomSecret()}`
  wizard.setApiKey(newKey)
}

function regenerateApiKey() {
  generateApiKey()
}

function copyApiKey() {
  navigator.clipboard.writeText(apiKey.value)
  // TODO: show toast
}

function downloadApiKey() {
  const blob = new Blob([apiKey.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `api-key-${localForm.uid}.txt`
  a.click()
  URL.revokeObjectURL(url)
}

function handleSubmit() {
  wizard.updateDevice({
    name: localForm.name,
    uid: localForm.uid,
    farmId: localForm.farmId || 1,
  })
}

function generateRandomSecret() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  for (let i = 0; i < 24; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}
</script>

<style scoped></style>
