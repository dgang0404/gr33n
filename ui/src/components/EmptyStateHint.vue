<template>
  <p
    class="text-sm"
    :class="compact ? 'text-zinc-600' : 'text-zinc-500 bg-zinc-950/60 border border-zinc-800/80 rounded-lg px-3 py-2'"
    :data-test="`empty-hint-${reason}`"
  >
    {{ config.message }}
    <router-link
      v-if="config.actionLabel && config.actionTo"
      v-nav-hint="config.actionTo"
      :to="config.actionTo"
      class="text-gr33n-500 hover:text-gr33n-400 hover:underline ml-1"
    >
      {{ config.actionLabel }} →
    </router-link>
    <button
      v-else-if="config.actionLabel && !config.actionTo"
      type="button"
      class="text-gr33n-500 hover:text-gr33n-400 hover:underline ml-1"
      @click="$emit('action')"
    >
      {{ config.actionLabel }} →
    </button>
  </p>
</template>

<script setup>
import { computed } from 'vue'
import { emptyHintConfig } from '../lib/emptyStateHint.js'

const props = defineProps({
  reason: { type: String, required: true },
  message: { type: String, default: '' },
  actionLabel: { type: String, default: '' },
  actionTo: { type: [String, Object], default: undefined },
  compact: { type: Boolean, default: false },
})

defineEmits(['action'])

const config = computed(() =>
  emptyHintConfig(props.reason, {
    message: props.message || undefined,
    actionLabel: props.actionLabel || undefined,
    actionTo: props.actionTo !== undefined ? props.actionTo : undefined,
  }),
)
</script>
