<template>
  <figure
    v-if="registered"
    class="shrink-0"
    :class="[sizeClass, { hidden: !loaded || failed }]"
    data-test="guardian-state-art"
    :data-state="state"
  >
    <img
      v-if="artUrl"
      :key="artUrl"
      :src="artUrl"
      :alt="alt"
      class="h-full w-full object-contain"
      decoding="async"
      @load="onLoad"
      @error="onError"
    />
  </figure>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import {
  fetchGuardianStateArtManifest,
  guardianStateArtAlt,
  guardianStateArtUrl,
  guardianStateHasArt,
} from '../lib/guardianStateArt'

const props = defineProps({
  state: { type: String, default: '' },
  size: {
    type: String,
    default: 'sm',
    validator: (v) => ['sm', 'md', 'panel'].includes(v),
  },
})

const manifest = ref(null)
const loaded = ref(false)
const failed = ref(false)

const artUrl = computed(() => guardianStateArtUrl(props.state, manifest.value))
const alt = computed(() => guardianStateArtAlt(props.state))
const registered = computed(() => guardianStateHasArt(props.state, manifest.value))

const sizeClass = computed(() => {
  if (props.size === 'panel') return 'h-[4.5rem] w-[4.5rem]'
  if (props.size === 'md') return 'h-12 w-12'
  return 'h-10 w-10'
})

function resetLoadState() {
  loaded.value = false
  failed.value = false
}

function onLoad() {
  loaded.value = true
  failed.value = false
}

function onError() {
  failed.value = true
  loaded.value = false
}

watch(() => props.state, resetLoadState)

onMounted(async () => {
  manifest.value = await fetchGuardianStateArtManifest()
})
</script>
