<template>
  <div
    class="rounded-md border border-violet-900/50 bg-violet-950/20 px-2.5 py-2 space-y-1.5"
    data-test="setup-pack-proposal-card"
  >
    <p class="text-[10px] uppercase tracking-widest text-violet-300/90">
      Grow setup bundle
      <span v-if="bundle.profile" class="normal-case tracking-normal text-violet-200/70">
        · {{ profileLabel }}
      </span>
    </p>
    <ul class="text-[11px] text-zinc-200 space-y-1">
      <li
        v-for="(step, i) in bundle.steps"
        :key="i"
        class="flex gap-1.5"
        data-test="setup-pack-step"
      >
        <span class="text-violet-400 shrink-0">{{ i + 1 }}.</span>
        <span>{{ step }}</span>
      </li>
    </ul>
    <p v-if="!bundle.steps.length" class="text-[11px] text-zinc-500 italic">
      No setup steps in frozen args.
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { formatSetupPackBundle } from '../lib/guardianSetupPack.js'

const props = defineProps({
  args: { type: Object, default: () => ({}) },
})

const bundle = computed(() => formatSetupPackBundle(props.args))

const profileLabel = computed(() => {
  const p = bundle.value.profile
  if (p === 'house_plant') return 'house plant'
  if (p === 'commercial_zone') return 'commercial zone'
  return p.replace(/_/g, ' ')
})
</script>
