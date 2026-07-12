<template>
  <div class="space-y-1" :class="spanClass" data-test="farm-maps-coords-paste">
    <label class="text-[11px] text-zinc-500 block">
      Paste from Google Maps
    </label>
    <div class="flex flex-wrap gap-2">
      <input
        v-model="pasteText"
        type="text"
        class="flex-1 min-w-[200px] bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder:text-zinc-600"
        placeholder="e.g. 40.8938° N, 81.4055° W"
        data-test="farm-maps-coords-paste-input"
        @paste="onPaste"
        @keydown.enter.prevent="applyPaste"
      />
      <button
        type="button"
        class="text-xs px-3 py-2 rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-700 hover:border-zinc-600 shrink-0"
        data-test="farm-maps-coords-paste-apply"
        @click="applyPaste"
      >
        Apply
      </button>
    </div>
    <p v-if="pasteOk" class="text-[11px] text-emerald-400/90" data-test="farm-maps-coords-paste-ok">
      Filled {{ formattedLat }}, {{ formattedLon }}
    </p>
    <p v-else-if="pasteError" class="text-[11px] text-red-400" data-test="farm-maps-coords-paste-error">
      {{ pasteError }}
    </p>
    <p v-else class="text-[11px] text-zinc-600">
      N/S and E/W set the sign automatically — west is negative only when the paste says W.
    </p>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { parseMapsCoordinates } from '../lib/siteWeather.js'

const props = defineProps({
  spanClass: { type: String, default: 'sm:col-span-3' },
})

const emit = defineEmits(['parsed'])

const pasteText = ref('')
const pasteError = ref('')
const pasteOk = ref(false)
const lastLat = ref(null)
const lastLon = ref(null)

const formattedLat = computed(() =>
  lastLat.value == null ? '' : Number(lastLat.value).toFixed(4),
)
const formattedLon = computed(() =>
  lastLon.value == null ? '' : Number(lastLon.value).toFixed(4),
)

function applyPaste() {
  pasteError.value = ''
  pasteOk.value = false
  const result = parseMapsCoordinates(pasteText.value)
  if (!result.ok) {
    pasteError.value = result.error
    return
  }
  lastLat.value = result.latitude
  lastLon.value = result.longitude
  pasteOk.value = true
  emit('parsed', { latitude: result.latitude, longitude: result.longitude })
}

function onPaste() {
  // Read value after the browser inserts pasted text.
  requestAnimationFrame(() => applyPaste())
}
</script>
