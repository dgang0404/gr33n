<template>
  <section
    v-if="farmContext.farmId"
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-4"
    data-test="farm-config-card"
  >
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div>
        <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Farm</h3>
        <p class="text-sm text-white mt-1">{{ farmContext.selectedFarm?.name ?? 'This farm' }}</p>
        <p v-if="farmContext.selectedFarm?.timezone" class="text-[11px] text-zinc-500 mt-0.5">
          Timezone: {{ farmContext.selectedFarm.timezone }}
        </p>
      </div>
      <router-link
        v-nav-hint="'/settings'"
        to="/settings"
        class="text-xs text-gr33n-500 hover:text-gr33n-400 shrink-0"
      >
        All settings →
      </router-link>
    </div>

    <form class="grid grid-cols-1 sm:grid-cols-3 gap-3" @submit.prevent="saveSite">
      <FarmMapsCoordsPaste span-class="sm:col-span-3" @parsed="onMapsPaste" />
      <div>
        <label class="text-zinc-500 text-[10px] uppercase tracking-wide">Latitude</label>
        <input
          v-model.number="siteForm.latitude"
          type="number"
          step="any"
          min="-90"
          max="90"
          class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          data-test="farm-config-latitude"
        />
      </div>
      <div>
        <label class="text-zinc-500 text-[10px] uppercase tracking-wide">Longitude</label>
        <input
          v-model.number="siteForm.longitude"
          type="number"
          step="any"
          min="-180"
          max="180"
          class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          data-test="farm-config-longitude"
        />
      </div>
      <div class="flex items-end">
        <button
          type="submit"
          class="text-xs px-3 py-2 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40"
          :disabled="siteSaving || !coordsValid"
          data-test="farm-config-save"
        >
          {{ siteSaving ? 'Saving…' : 'Save site' }}
        </button>
      </div>
    </form>
    <p class="text-[11px] text-zinc-600">
      Site coordinates power daylight hours on the morning strip — no internet required for solar math.
    </p>
    <p v-if="siteMessage" class="text-xs text-emerald-400">{{ siteMessage }}</p>
    <p v-if="siteError" class="text-xs text-red-400">{{ siteError }}</p>
  </section>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useFarmContextStore } from '../stores/farmContext'
import { parseFarmCoordinates } from '../lib/siteWeather.js'
import FarmMapsCoordsPaste from './FarmMapsCoordsPaste.vue'

const farmContext = useFarmContextStore()

const siteForm = reactive({ latitude: null, longitude: null })
const siteSaving = ref(false)
const siteError = ref('')
const siteMessage = ref('')

const coordsValid = computed(() =>
  Number.isFinite(siteForm.latitude) && Number.isFinite(siteForm.longitude),
)

function syncFromFarm() {
  const farm = farmContext.selectedFarm
  if (!farm) return
  const { latitude, longitude } = parseFarmCoordinates(farm)
  siteForm.latitude = latitude
  siteForm.longitude = longitude
}

function onMapsPaste({ latitude, longitude }) {
  siteForm.latitude = latitude
  siteForm.longitude = longitude
}

async function saveSite() {
  const farmId = farmContext.farmId
  if (!farmId || !coordsValid.value) return
  siteSaving.value = true
  siteError.value = ''
  siteMessage.value = ''
  try {
    await farmContext.patchSite(farmId, {
      latitude: siteForm.latitude,
      longitude: siteForm.longitude,
      elevation_m: null,
    })
    siteMessage.value = 'Site saved — daylight chip will use these coordinates.'
  } catch (e) {
    siteError.value = e?.response?.data?.error || e.message || 'Save failed'
  } finally {
    siteSaving.value = false
  }
}

watch(() => farmContext.selectedFarm, syncFromFarm, { immediate: true })
</script>
