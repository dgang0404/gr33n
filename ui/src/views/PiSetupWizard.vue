<template>
  <div class="min-h-screen bg-zinc-950 text-zinc-200 pb-24">
    <!-- Progress bar -->
    <div class="sticky top-0 z-10 bg-zinc-900 border-b border-zinc-800 px-4 py-4 sm:px-6">
      <div class="max-w-4xl mx-auto">
        <div class="flex items-center justify-between mb-3">
          <div class="text-sm font-semibold text-zinc-300">
            Step {{ currentStep }} of 6: {{ stepTitle }}
          </div>
          <button
            type="button"
            @click="cancel"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-400 hover:text-zinc-200 hover:border-zinc-600"
          >
            Cancel
          </button>
        </div>
        <!-- Progress indicator -->
        <div class="flex gap-1">
          <div
            v-for="step in 6"
            :key="step"
            class="flex-1 h-1 rounded-full transition-colors"
            :class="
              step < currentStep
                ? 'bg-green-600'
                : step === currentStep
                  ? 'bg-green-500'
                  : 'bg-zinc-800'
            "
          />
        </div>
      </div>
    </div>

    <!-- Main content -->
    <div class="max-w-4xl mx-auto px-4 py-8 sm:px-6">
      <!-- Step 1: Welcome -->
      <PiWizardWelcome v-if="currentStep === 1" />

      <!-- Step 2: Register Pi -->
      <PiWizardRegister v-else-if="currentStep === 2" />

      <!-- Step 3: Assign Channels -->
      <PiWizardChannels v-else-if="currentStep === 3" />

      <!-- Step 4: Network Test -->
      <PiWizardNetwork v-else-if="currentStep === 4" />

      <!-- Step 5: Download Config -->
      <PiWizardDownload v-else-if="currentStep === 5" />

      <!-- Step 6: Confirm -->
      <PiWizardConfirm v-else-if="currentStep === 6" />
    </div>

    <!-- Navigation buttons -->
    <div class="fixed bottom-0 left-0 right-0 bg-zinc-900 border-t border-zinc-800 px-4 py-4 sm:px-6">
      <div class="max-w-4xl mx-auto flex justify-between gap-3">
        <button
          type="button"
          @click="prevStep"
          :disabled="currentStep === 1"
          class="px-4 py-2 rounded-lg border border-zinc-700 text-zinc-300 hover:text-white hover:border-zinc-600 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          ← Back
        </button>
        <button
          v-if="currentStep < 6"
          type="button"
          @click="nextStep"
          :disabled="!canAdvance"
          class="px-4 py-2 rounded-lg bg-green-700 text-white hover:bg-green-600 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Next →
        </button>
        <button
          v-else
          type="button"
          @click="finish"
          class="px-4 py-2 rounded-lg bg-green-600 text-white hover:bg-green-500"
        >
          🎉 Finish
        </button>
      </div>
    </div>

    <!-- Glossary panel -->
    <PiGlossaryPanel />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePiWizardStore } from '@/stores/piWizardStore'
import { useRouter } from 'vue-router'

import PiWizardWelcome from '@/components/PiWizard/Welcome.vue'
import PiWizardRegister from '@/components/PiWizard/Register.vue'
import PiWizardChannels from '@/components/PiWizard/Channels.vue'
import PiWizardNetwork from '@/components/PiWizard/Network.vue'
import PiWizardDownload from '@/components/PiWizard/Download.vue'
import PiWizardConfirm from '@/components/PiWizard/Confirm.vue'
import PiGlossaryPanel from '@/components/PiGlossaryPanel.vue'

const wizard = usePiWizardStore()
const router = useRouter()

const currentStep = computed(() => wizard.currentStep)
const canAdvance = computed(() => wizard.canAdvance)

const stepTitle = computed(() => {
  const titles = {
    1: 'Welcome & Hardware Checklist',
    2: 'Register Pi on Farm',
    3: 'Assign Relay Channels',
    4: 'Network & API Configuration',
    5: 'Download Config',
    6: 'Verify & Complete',
  }
  return titles[currentStep.value] || ''
})

function nextStep() {
  wizard.nextStep()
}

function prevStep() {
  wizard.prevStep()
}

function cancel() {
  if (confirm('Cancel Pi setup wizard? Your progress will be lost.')) {
    wizard.reset()
    router.push('/hardware')
  }
}

function finish() {
  wizard.reset()
  router.push('/hardware')
}
</script>

<style scoped></style>
