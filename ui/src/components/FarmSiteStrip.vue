<template>
  <section
    class="bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3 space-y-3"
    data-test="farm-site-strip"
  >
    <div class="flex flex-wrap items-start justify-between gap-3">
      <!-- Sun dial -->
      <div class="flex items-center gap-3 min-w-[200px]" data-test="farm-site-sun">
        <div class="relative w-16 h-8 shrink-0" aria-hidden="true">
          <svg viewBox="0 0 64 32" class="w-full h-full text-zinc-600">
            <path d="M4 28 A 28 28 0 0 1 60 28" fill="none" stroke="currentColor" stroke-width="2" />
            <circle
              v-if="sunMarker"
              :cx="sunMarker.cx"
              :cy="sunMarker.cy"
              r="3"
              class="fill-amber-400"
            />
          </svg>
        </div>
        <div v-if="sunTimes" class="text-[11px] leading-snug">
          <p class="text-zinc-300">
            <span class="text-zinc-500">↑</span> {{ sunTimes.sunrise }}
            <span class="text-zinc-500 ml-2">↓</span> {{ sunTimes.sunset }}
          </p>
          <p v-if="sunTimes.daylength" class="text-zinc-500">{{ sunTimes.daylength }} daylight</p>
        </div>
        <p v-else class="text-[11px] text-zinc-500">Set farm coordinates for sun times</p>
      </div>

      <!-- Outdoor rollup -->
      <div class="text-[11px] min-w-[140px]" data-test="farm-site-outdoor">
        <p class="text-zinc-500 uppercase tracking-wide text-[10px]">Outdoor</p>
        <p class="text-zinc-300 mt-0.5">{{ outdoorSummary }}</p>
      </div>

      <!-- Water source -->
      <div class="text-[11px] min-w-[160px]" data-test="farm-site-water">
        <p class="text-zinc-500 uppercase tracking-wide text-[10px]">Water</p>
        <router-link
          v-nav-hint="'/feed-water'"
          :to="feedWaterLink"
          class="text-zinc-300 mt-0.5 hover:text-green-400 block"
        >
          {{ waterSourceLine }}
        </router-link>
      </div>

      <!-- Phase 176 — farm pulse cells -->
      <div
        v-for="cell in pulseCells"
        :key="cell.id"
        class="text-[11px] min-w-[140px]"
        :data-test="`farm-site-pulse-${cell.id}`"
      >
        <p class="text-zinc-500 uppercase tracking-wide text-[10px]">{{ cell.label }}</p>
        <router-link
          v-nav-hint="cell.link"
          :to="cell.link"
          class="text-zinc-300 mt-0.5 hover:text-green-400 block"
        >
          {{ cell.value }}
        </router-link>
      </div>

      <!-- Lat/long chip -->
      <div v-if="!coordsSet" class="flex items-center gap-2" data-test="farm-site-coords-prompt">
        <button
          type="button"
          class="text-[11px] px-2.5 py-1 rounded-full bg-zinc-800 text-amber-300/90 border border-amber-900/50 hover:border-amber-700"
          @click="showCoords = !showCoords"
        >
          Set farm location for sun &amp; weather
        </button>
      </div>
    </div>

    <form
      v-if="showCoords && !coordsSet"
      class="grid grid-cols-1 sm:grid-cols-3 gap-2 pt-1 border-t border-zinc-800"
      data-test="farm-site-coords-form"
      @submit.prevent="saveSite"
    >
      <input
        v-model.number="siteForm.latitude"
        type="number"
        step="any"
        min="-90"
        max="90"
        placeholder="Latitude"
        class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
      />
      <input
        v-model.number="siteForm.longitude"
        type="number"
        step="any"
        min="-180"
        max="180"
        placeholder="Longitude"
        class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
      />
      <button
        type="submit"
        class="text-xs px-3 py-2 rounded-lg bg-green-900/50 text-green-400 border border-green-800 disabled:opacity-40"
        :disabled="siteSaving || !coordsValid"
      >
        {{ siteSaving ? 'Saving…' : 'Save location' }}
      </button>
      <p v-if="siteError" class="text-xs text-red-400 sm:col-span-3">{{ siteError }}</p>
    </form>
  </section>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { useFarmContextStore } from '../stores/farmContext'
import { parseFarmCoordinates } from '../lib/siteWeather.js'
import { formatSunTimes, sunDialProgress } from '../lib/farmCanvasLayout.js'
import { classifySensorHardwareState } from '../lib/farmVisualStatus.js'
import { feedWaterRoute } from '../lib/dashboardWorkspaceLinks.js'
import { buildFarmTodayPulse } from '../lib/farmTodayPulse.js'

const props = defineProps({
  siteWeather: { type: Object, default: null },
  zones: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
  reservoirs: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  cropCycles: { type: Array, default: () => [] },
  devices: { type: Array, default: () => [] },
  queueDepth: { type: Number, default: 0 },
  tasks: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
})

const farmContext = useFarmContextStore()
const showCoords = ref(false)
const siteForm = reactive({ latitude: null, longitude: null })
const siteSaving = ref(false)
const siteError = ref('')

const sunTimes = computed(() => formatSunTimes(props.siteWeather?.solar))
const sunProgress = computed(() => sunDialProgress(props.siteWeather?.solar))
const sunMarker = computed(() => {
  if (sunProgress.value == null) return null
  const p = sunProgress.value
  return { cx: 4 + p * 56, cy: 28 - Math.sin(p * Math.PI) * 24 }
})

const coordsSet = computed(() => {
  const { latitude, longitude } = parseFarmCoordinates(farmContext.selectedFarm)
  return Number.isFinite(latitude) && Number.isFinite(longitude)
})

const coordsValid = computed(() =>
  Number.isFinite(siteForm.latitude) && Number.isFinite(siteForm.longitude),
)

const outdoorZones = computed(() =>
  props.zones.filter((z) => String(z.zone_type || '').toLowerCase().includes('outdoor')),
)

const outdoorSummary = computed(() => {
  const outdoor = outdoorZones.value
  if (!outdoor.length) return 'No outdoor zones'
  const outdoorSensors = props.sensors.filter(
    (s) => outdoor.some((z) => Number(z.id) === Number(s.zone_id)),
  )
  if (!outdoorSensors.length) return 'No outdoor sensors yet'
  const states = outdoorSensors.map((s) => classifySensorHardwareState(s, props.readings[s.id]))
  const healthy = states.filter((x) => x === 'healthy').length
  if (healthy === outdoorSensors.length) return `${healthy} outdoor sensor${healthy === 1 ? '' : 's'} healthy`
  const unset = states.filter((x) => x === 'not_set_up').length
  if (unset === outdoorSensors.length) return 'Outdoor sensors not set up yet'
  return `${healthy} healthy · ${outdoor.length} outdoor zone${outdoor.length === 1 ? '' : 's'}`
})

const feedWaterLink = computed(() => feedWaterRoute(props.zones))

const pulseCells = computed(() => buildFarmTodayPulse({
  zones: props.zones,
  programs: props.programs,
  schedules: props.schedules,
  cropCycles: props.cropCycles,
  devices: props.devices,
  queueDepth: props.queueDepth,
  actuators: props.actuators,
  sensors: props.sensors,
  readings: props.readings,
  tasks: props.tasks,
  alerts: props.alerts,
  fertigationEvents: props.fertigationEvents,
}).cells)

const waterSourceLine = computed(() => {
  const res = props.reservoirs?.[0]
  const progCount = (props.programs || []).filter((p) => p.is_active !== false).length
  const zoneCount = new Set(
    (props.programs || []).filter((p) => p.is_active !== false).map((p) => p.target_zone_id),
  ).size
  const hasGravity = (props.actuators || []).some(
    (a) => String(a.actuator_type || '').toLowerCase() === 'drip'
      || String(a.name || '').toLowerCase().includes('gravity'),
  )
  const delivery = hasGravity ? 'gravity' : 'pump'
  if (res && zoneCount) {
    return `${res.name} → ${delivery} → ${zoneCount} zone${zoneCount === 1 ? '' : 's'}`
  }
  if (progCount && zoneCount) {
    return `${delivery} → ${zoneCount} zone${zoneCount === 1 ? '' : 's'}`
  }
  return 'Set up feeding plans →'
})

function syncFromFarm() {
  const { latitude, longitude } = parseFarmCoordinates(farmContext.selectedFarm)
  siteForm.latitude = latitude
  siteForm.longitude = longitude
}

async function saveSite() {
  const farmId = farmContext.farmId
  if (!farmId || !coordsValid.value) return
  siteSaving.value = true
  siteError.value = ''
  try {
    await farmContext.patchSite(farmId, {
      latitude: siteForm.latitude,
      longitude: siteForm.longitude,
      elevation_m: null,
    })
    showCoords.value = false
  } catch (e) {
    siteError.value = e?.response?.data?.error || e.message || 'Save failed'
  } finally {
    siteSaving.value = false
  }
}

watch(() => farmContext.selectedFarm, syncFromFarm, { immediate: true })
</script>
