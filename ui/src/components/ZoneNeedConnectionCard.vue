<template>
  <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 space-y-2">
    <div class="flex items-start justify-between gap-2">
      <div>
        <p class="text-white text-sm font-medium">{{ title }}</p>
        <p v-if="subtitle" class="text-zinc-500 text-xs mt-0.5">{{ subtitle }}</p>
      </div>
      <router-link v-if="manageTo" :to="manageTo" class="text-xs text-green-600 hover:text-green-400 shrink-0">
        Details →
      </router-link>
    </div>

    <div v-if="readingLabel" class="flex items-center justify-between text-xs">
      <span class="text-zinc-500">Reading</span>
      <span class="text-zinc-200">{{ readingLabel }}</span>
    </div>
    <div v-if="targetLabel" class="flex items-center justify-between text-xs">
      <span class="text-zinc-500">Target</span>
      <span class="text-zinc-300">{{ targetLabel }}</span>
    </div>
    <div v-if="automationLabel" class="flex items-center justify-between text-xs">
      <span class="text-zinc-500">Automation</span>
      <span class="text-zinc-300">{{ automationLabel }}</span>
    </div>
    <div v-if="controlLabel" class="flex items-center justify-between text-xs">
      <span class="text-zinc-500">Control</span>
      <span :class="controlOnline ? 'text-green-400' : 'text-zinc-500'">{{ controlLabel }}</span>
    </div>
    <p v-if="lastEventLabel" class="text-zinc-600 text-[10px] border-t border-zinc-800 pt-2">
      Last: {{ lastEventLabel }}
    </p>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  manageTo: { type: [String, Object], default: '' },
  readingLabel: { type: String, default: '' },
  targetLabel: { type: String, default: '' },
  automationLabel: { type: String, default: '' },
  controlLabel: { type: String, default: '' },
  controlOnline: { type: Boolean, default: false },
  lastEventLabel: { type: String, default: '' },
})
</script>
