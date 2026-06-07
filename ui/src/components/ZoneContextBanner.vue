<template>
  <div
    class="text-xs rounded-lg px-3 py-2 mb-3 border"
    :class="bannerClass"
    data-test="zone-context-banner"
  >
    <nav class="text-zinc-400 mb-1" aria-label="Breadcrumb">
      <router-link v-nav-hint="'/zones'" to="/zones" class="hover:text-green-400">Zones</router-link>
      <span class="mx-1">›</span>
      <router-link :to="zoneRoute" class="hover:text-green-400">{{ zoneName }}</router-link>
      <span class="mx-1">›</span>
      <span class="text-zinc-200">{{ pageLabel }}</span>
    </nav>
    <p class="text-zinc-300">
      Viewing <strong>{{ zoneName }}</strong>
      <template v-if="backToZoneTab">
        —
        <router-link :to="backToZoneRoute" class="text-green-400 hover:text-green-300">
          Back to zone {{ backToZoneTabLabel }} →
        </router-link>
      </template>
      <router-link
        v-if="clearRoute"
        :to="clearRoute"
        class="text-green-400 hover:text-green-300 ml-1"
      >
        Show all zones →
      </router-link>
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  zoneName: { type: String, required: true },
  pageLabel: { type: String, required: true },
  clearRoute: { type: [Object, String], default: null },
  /** e.g. 'water' | 'light' — adds back link to zone tab */
  backToZoneTab: { type: String, default: '' },
  variant: { type: String, default: 'default' },
})

const backToZoneTabLabel = computed(() => {
  const map = { water: 'Water', light: 'Light', air: 'Climate', overview: 'Overview' }
  return map[props.backToZoneTab] || props.backToZoneTab
})

const zoneRoute = computed(() => ({
  path: `/zones/${props.zoneId}`,
  query: props.backToZoneTab ? { tab: props.backToZoneTab } : {},
}))

const backToZoneRoute = computed(() => ({
  path: `/zones/${props.zoneId}`,
  query: { tab: props.backToZoneTab },
}))

const bannerClass = computed(() => {
  if (props.variant === 'fertigation') {
    return 'text-amber-100/90 bg-amber-950/30 border-amber-900/50'
  }
  return 'text-blue-200/90 bg-blue-950/30 border-blue-900/50'
})
</script>
