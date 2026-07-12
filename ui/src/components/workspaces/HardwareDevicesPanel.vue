<template>
  <div class="p-4 sm:p-6 space-y-4 max-w-2xl">
    <p class="text-sm text-zinc-400">
      Register Pis, rotate API keys, and open the device wizard from here.
    </p>
    <div class="flex flex-wrap gap-3">
      <RouterLink
        v-if="deviceWizardTo"
        v-nav-hint="'/hardware'"
        :to="deviceWizardTo"
        class="inline-flex items-center px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-white text-sm font-medium"
      >
        Add Pi device
      </RouterLink>
      <RouterLink
        v-nav-hint="'/hardware'"
        to="/pi-setup-wizard"
        class="inline-flex items-center px-4 py-2 rounded-lg bg-blue-700 hover:bg-blue-600 text-white text-sm font-medium"
      >
        🧙 Pi Setup Wizard
      </RouterLink>
      <RouterLink
        v-nav-hint="'/settings'"
        to="/settings"
        class="inline-flex items-center px-4 py-2 rounded-lg border border-zinc-700 text-zinc-300 hover:text-white text-sm"
      >
        Device settings
      </RouterLink>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useFarmContextStore } from '../../stores/farmContext'

const farmContext = useFarmContextStore()

const deviceWizardTo = computed(() => {
  const id = farmContext.farmId
  return id ? `/farms/${id}/devices/new` : null
})
</script>
