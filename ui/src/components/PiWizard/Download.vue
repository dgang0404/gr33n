<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-2">Download Configuration</h2>
      <p class="text-zinc-400">Your Pi needs this configuration file to connect and deploy.</p>
    </section>

    <!-- YAML Preview -->
    <div>
      <label class="block text-sm font-medium text-zinc-300 mb-2">config.yaml</label>
      <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 max-h-96 overflow-auto">
        <pre class="font-mono text-xs text-zinc-300 whitespace-pre-wrap"><code>{{ generatedYaml }}</code></pre>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="flex flex-wrap gap-2">
      <button
        type="button"
        @click="copyYaml"
        class="flex-1 min-w-[120px] px-4 py-2.5 rounded-lg bg-green-700 text-white hover:bg-green-600 text-sm font-medium"
      >
        📋 Copy YAML
      </button>
      <button
        type="button"
        @click="downloadYaml"
        class="flex-1 min-w-[120px] px-4 py-2.5 rounded-lg bg-green-700 text-white hover:bg-green-600 text-sm font-medium"
      >
        ⬇️ Download File
      </button>
    </div>

    <!-- SCP instruction -->
    <div class="bg-blue-950/30 border border-blue-900/50 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-blue-400 mb-2">Transfer Config to Pi</h3>
      <p class="text-xs text-zinc-400 mb-3">
        From your computer, run this command to copy the config to your Pi:
      </p>
      <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-2 mb-2">
        <pre class="font-mono text-xs text-green-300"><code>{{ scpCommand }}</code></pre>
      </div>
      <button
        type="button"
        @click="copyScp"
        class="text-xs px-3 py-1.5 rounded-lg border border-blue-700 text-blue-300 hover:text-blue-200"
      >
        Copy SCP command
      </button>
    </div>

    <!-- Alternative: paste via SSH -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-white mb-2">Or paste via SSH</h3>
      <ol class="text-xs text-zinc-400 space-y-2 list-decimal list-inside">
        <li>SSH to your Pi: <code class="text-zinc-300">ssh pi@192.168.1.x</code></li>
        <li>Create directory: <code class="text-zinc-300">mkdir -p ~/.gr33n && cd ~/.gr33n</code></li>
        <li>Paste YAML content into: <code class="text-zinc-300">nano config.yaml</code></li>
        <li>Save: <code class="text-zinc-300">Ctrl+O, Enter, Ctrl+X</code></li>
      </ol>
    </div>

    <!-- Validation -->
    <div v-if="!generatedYaml" class="bg-red-950/30 border border-red-700/50 rounded-lg p-3">
      <div class="text-xs text-red-300">✗ Config not generated yet</div>
    </div>
    <div v-else class="bg-green-950/30 border border-green-700/50 rounded-lg p-3">
      <div class="text-xs text-green-300">✓ Config ready for download</div>
    </div>
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'
import { generateConfigYaml, downloadYaml as dlYaml, copyToClipboard } from '@/lib/piWizardConfigGenerator'
import { validateStep5 } from '@/lib/piWizardValidation'

const wizard = usePiWizardStore()

const generatedYaml = computed(() => {
  if (wizard.formData.configYaml) {
    return wizard.formData.configYaml
  }
  const yaml = generateConfigYaml(wizard.formData)
  wizard.setConfigYaml(yaml)
  return yaml
})

const scpCommand = computed(() => {
  const uid = wizard.formData.device.uid || 'pi'
  return `scp config.yaml pi@192.168.1.x:~/.gr33n/config.yaml`
})

watch(generatedYaml, () => {
  const errors = validateStep5(wizard.formData)
  wizard.updateValidation(5, errors)
}, { immediate: true })

async function copyYaml() {
  try {
    await copyToClipboard(generatedYaml.value)
    // TODO: Show success toast
  } catch (err) {
    console.error('Copy failed:', err)
  }
}

function downloadYaml() {
  dlYaml(generatedYaml.value, 'config.yaml')
}

async function copyScp() {
  try {
    await copyToClipboard(scpCommand.value)
    // TODO: Show success toast
  } catch (err) {
    console.error('Copy failed:', err)
  }
}
</script>

<style scoped>
code {
  @apply font-mono text-xs;
}
</style>
