<template>
  <router-link
    v-if="sensor?.id"
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

const props = defineProps({
  sensor:  Object,
  reading: Object,
})

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

const thresholdLow = computed(() => {
  const db = props.sensor?.alert_threshold_low
  return db != null ? Number(db) : meta.value.low
})
const thresholdHigh = computed(() => {
  const db = props.sensor?.alert_threshold_high
  return db != null ? Number(db) : meta.value.high
})

const status = computed(() => {
  if (val.value == null) return 'unknown'
  if (val.value < thresholdLow.value || val.value > thresholdHigh.value) return 'danger'
  const lo = thresholdLow.value, hi = thresholdHigh.value, range = hi - lo
  if (range > 0 && (val.value < lo + range * 0.15 || val.value > hi - range * 0.15)) return 'warn'
  return 'ok'
})

const badgeClass = computed(() => ({
  ok:      'badge-ok',
  warn:    'badge-warn',
  danger:  'badge-danger',
  unknown: 'badge-off',
}[status.value]))

const statusLabel = computed(() => ({ ok: 'OK', warn: 'WARN', danger: 'ALERT', unknown: 'NO DATA' }[status.value]))

const ago = computed(() => {
  if (!props.reading?.reading_time) return 'No reading yet'
  const d = new Date(props.reading.reading_time)
  const s = Math.floor((Date.now() - d) / 1000)
  if (s < 60) return `${s}s ago`
  if (s < 3600) return `${Math.floor(s/60)}m ago`
  return `${Math.floor(s/3600)}h ago`
})
</script>
