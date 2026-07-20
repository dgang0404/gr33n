<template>
  <details class="text-sm text-zinc-400 border border-zinc-800 rounded-lg px-3 py-2 bg-zinc-950/50">
    <summary class="cursor-pointer text-zinc-300 hover:text-white select-none" data-test="nf-learn-how">
      Learn how
    </summary>
    <p class="mt-2 text-xs text-zinc-500">
      Opens the matching field guide in Help search — same content Guardian cites when farm context is on.
    </p>
    <router-link
      v-if="guideFile"
      :to="learnRoute"
      class="inline-block mt-2 text-xs text-green-400 hover:text-green-300"
      :data-test="`nf-learn-how-link-${slug}`"
    >
      Read {{ slug }} →
    </router-link>
  </details>
</template>

<script setup>
import { computed } from 'vue'
import { fieldGuideLearnRoute } from '../../lib/naturalFarmingSwitchover.js'

const props = defineProps({
  guideFile: { type: String, default: '' },
})

const slug = computed(() => String(props.guideFile || '').replace(/\.md$/, ''))
const learnRoute = computed(() => fieldGuideLearnRoute(props.guideFile))
</script>
