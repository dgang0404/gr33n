<template>
  <router-link
    v-if="sensor?.id"
    v-nav-hint="'/sensors'"
    :to="{ name: 'sensor-detail', params: { id: sensor.id } }"
    class="block rounded-xl ring-0 hover:ring-1 ring-zinc-600 focus:outline-none focus:ring-2 focus:ring-green-600 transition-shadow min-w-0"
  >
    <div class="card flex flex-col gap-2 min-w-0">
      <div class="flex items-center justify-between">
        <span class="text-xs text-gray-500 uppercase tracking-wide truncate">{{ label }}</span>
        <span :class="badgeClass">{{ statusLabel }}</span>
      </div>
      <div class="flex items-end gap-1">
        <span class="text-2xl font-bold font-mono text-white">{{ displayValue }}</span>
        <span class="text-sm text-gray-500 mb-0.5">{{ unit }}</span>
      </div>
      <div class="text-xs text-gray-600">{{ ago }}</div>
    </div>
  </router-link>
  <div v-else class="card flex flex-col gap-2 min-w-0 opacity-50">
    <div class="text-xs text-gray-500">Invalid sensor</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { classifySensorQuality, resolveSensitivity, idealRangeFromCropStage, QUALITY_LABEL } from '../lib/sensorQuality.js'

const props = defineProps({
  sensor:  Object,
  reading: Object,
  // Phase 217 WS6 — matched crop_library/effective-profile stage row for this
  // sensor's zone (if it has an active crop cycle). Null when there isn't one
  // or the caller hasn't resolved it — falls back to hardcoded SENSOR_META.
  cropStageRow: { type: Object, default: null },
})

const store = useFarmStore()

const SENSOR_META = {
  temperature:   { icon: '🌡️', unit: '°C',       low: 15, high: 30 },
  humidity:      { icon: '💧', unit: '%',         low: 40, high: 80 },
  co2:           { icon: '💨', unit: 'ppm',       low: 400, high: 1500 },
  ph:            { icon: '⚗️', unit: 'pH',        low: 5.5, high: 7.0 },
  ec:            { icon: '⚡', unit: 'mS/cm',     low: 1.0, high: 3.5 },
  par:           { icon: '☀️', unit: 'µmol/m²/s', low: 100, high: 900 },
  soil_moisture: { icon: '🌱', unit: '%',         low: 30, high: 80  },
}

const meta  = computed(() => SENSOR_META[props.sensor?.sensor_type] ?? { unit: '', low: 0, high: 100 })
const label = computed(() => `${meta.value.icon ?? ''} ${props.sensor?.name ?? props.sensor?.sensor_type}`)
const unit  = computed(() => meta.value.unit)
const val   = computed(() => props.reading?.value_raw ?? null)

const displayValue = computed(() => val.value == null ? '—' : Number(val.value).toFixed(1))

// Precedence: per-sensor DB override > crop-library stage row (if this
// sensor's zone has an active crop cycle) > hardcoded SENSOR_META default.
const effectiveRange = computed(() => {
  const dbLow = props.sensor?.alert_threshold_low
  const dbHigh = props.sensor?.alert_threshold_high
  if (dbLow != null && dbHigh != null) {
    return { low: Number(dbLow), high: Number(dbHigh) }
  }
  const cropRange = idealRangeFromCropStage(props.sensor?.sensor_type, props.cropStageRow)
  if (cropRange) return cropRange
  return { low: meta.value.low, high: meta.value.high }
})

const sensitivity = computed(() => resolveSensitivity(props.sensor, store.farm))

const quality = computed(() => classifySensorQuality(val.value, effectiveRange.value, sensitivity.value))

const badgeClass = computed(() => ({
  critical: 'badge-critical',
  bad:      'badge-bad',
  mid:      'badge-mid',
  good:     'badge-good',
  awesome:  'badge-awesome',
  unknown:  'badge-off',
}[quality.value]))

const statusLabel = computed(() => quality.value === 'unknown' ? 'NO DATA' : QUALITY_LABEL[quality.value])

const ago = computed(() => {
  if (!props.reading?.reading_time) return 'No reading yet'
  const d = new Date(props.reading.reading_time)
  const s = Math.floor((Date.now() - d) / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s/60)}m ago`
  return `${Math.floor(s/3600)}h ago`
})
</script>
