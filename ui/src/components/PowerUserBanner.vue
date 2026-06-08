<template>
  <div
    class="rounded-lg border px-3 py-2 text-xs"
    :class="bannerClass"
    data-test="power-user-banner"
  >
    <strong class="font-semibold">Power-user mode.</strong>
    {{ message }}
    <router-link
      v-if="farmerLink"
      v-nav-hint="farmerLink"
      :to="farmerLink"
      class="ml-1 underline hover:text-green-300"
    >
      {{ farmerLinkLabel }} →
    </router-link>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  farmerLink: { type: [Object, String], default: null },
  farmerLinkLabel: { type: String, default: 'Farmer view' },
  message: {
    type: String,
    default: 'Cron, JSON predicates, and bulk editing live here. Use Grow → Targets for everyday comfort bands and toggles.',
  },
})

const bannerClass = computed(() =>
  'text-amber-100/90 bg-amber-950/30 border-amber-900/50',
)
</script>
