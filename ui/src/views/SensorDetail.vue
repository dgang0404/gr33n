<template>
  <div class="p-6 space-y-6">
    <div>
      <router-link to="/sensors" class="text-xs text-zinc-500 hover:text-zinc-300">&larr; Back to sensors</router-link>
      <div v-if="loadError" class="mt-4 rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-red-300 text-sm">
        {{ loadError }}
      </div>
      <div v-else-if="!sensor" class="mt-4 space-y-3 animate-pulse" aria-hidden="true">
        <div class="h-8 bg-zinc-800 rounded-lg w-2/3 max-w-md" />
        <div class="h-4 bg-zinc-800/80 rounded w-1/3" />
        <div class="h-24 bg-zinc-900 border border-zinc-800 rounded-xl" />
      </div>
      <template v-else>
        <div class="mt-3 flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
          <div>
            <h1 class="text-xl font-semibold text-white flex items-center gap-2">
              <span>{{ sensorIcon(sensor.sensor_type) }}</span>
              <span>{{ sensor.name }}</span>
            </h1>
            <p class="text-zinc-500 text-sm mt-1">
              <span class="capitalize">{{ sensor.sensor_type }}</span>
              <span v-if="zoneName" class="text-zinc-600"> · {{ zoneName }}</span>
            </p>
          </div>
          <div class="text-right">
            <p class="text-xs text-zinc-500 uppercase tracking-wide">Current</p>
            <p class="text-3xl font-mono font-semibold text-white tabular-nums">
              {{ currentDisplay }}
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2 mt-4">
          <button
            v-for="opt in rangeOptions"
            :key="opt.hours"
            type="button"
            @click="setRange(opt.hours)"
            class="text-xs font-medium px-3 py-1.5 rounded-lg border transition-colors"
            :class="rangeHours === opt.hours
              ? 'bg-green-900/40 border-green-700 text-green-300'
              : 'bg-zinc-900 border-zinc-700 text-zinc-400 hover:border-zinc-600'"
          >
            {{ opt.label }}
          </button>
          <button
            type="button"
            @click="exportReadingsCsv"
            :disabled="!readings.length || historyLoading"
            class="text-xs ml-auto px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-400 hover:border-zinc-500 hover:text-zinc-200 disabled:opacity-40"
          >
            Export range (CSV)
          </button>
        </div>

        <div v-if="historyLoading" class="text-zinc-500 text-sm mt-2">Loading stats and chart…</div>
        <div v-else class="grid grid-cols-2 md:grid-cols-4 gap-3 mt-2">
          <div v-for="card in statCards" :key="card.label" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
            <p class="text-zinc-500 text-xs mb-1">{{ card.label }}</p>
            <p class="text-white text-lg font-mono tabular-nums">{{ card.value }}</p>
          </div>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h2 class="text-sm font-semibold text-white mb-3">History</h2>
          <p v-if="historyLoading" class="text-zinc-500 text-sm">Loading chart…</p>
          <p v-else-if="!chartData.datasets[0]?.data?.length" class="text-zinc-500 text-sm">No readings in this range. Try a longer window or confirm the Pi is posting readings.</p>
          <div v-else class="h-80 w-full">
            <Line :data="chartData" :options="chartOptions" />
          </div>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-white">Alert thresholds</h2>
            <button
              v-if="!editingAlerts"
              type="button"
              @click="beginEditAlerts"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 hover:text-white"
            >Edit</button>
          </div>

          <div v-if="!editingAlerts" class="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
            <div>
              <p class="text-zinc-500 text-xs mb-1">Low threshold</p>
              <p class="text-white tabular-nums">{{ sensor.alert_threshold_low ?? '—' }}</p>
            </div>
            <div>
              <p class="text-zinc-500 text-xs mb-1">High threshold</p>
              <p class="text-white tabular-nums">{{ sensor.alert_threshold_high ?? '—' }}</p>
            </div>
            <div>
              <p class="text-zinc-500 text-xs mb-1">Alert duration</p>
              <p class="text-white tabular-nums">{{ formatSecondsAsMinutes(sensor.alert_duration_seconds) }}</p>
            </div>
            <div>
              <p class="text-zinc-500 text-xs mb-1">Alert cooldown</p>
              <p class="text-white tabular-nums">{{ formatSecondsAsMinutes(sensor.alert_cooldown_seconds) }}</p>
            </div>
          </div>

          <form v-else class="grid grid-cols-1 md:grid-cols-2 gap-4" @submit.prevent="saveAlerts">
            <label class="block">
              <span class="text-xs text-zinc-400">Low threshold</span>
              <input
                v-model="editForm.alert_threshold_low"
                type="number"
                step="any"
                placeholder="(none)"
                class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
              />
            </label>
            <label class="block">
              <span class="text-xs text-zinc-400">High threshold</span>
              <input
                v-model="editForm.alert_threshold_high"
                type="number"
                step="any"
                placeholder="(none)"
                class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
              />
            </label>
            <label class="block">
              <span class="inline-flex items-center text-xs text-zinc-400">
                Alert duration (minutes)
                <HelpTip position="top">
                  How long a reading must stay out of bounds before an alert fires.
                  Example: set to <strong>10</strong> to mean "only alert after 10 minutes
                  continuously below / above threshold". <strong>0</strong> alerts on the first breaching reading.
                </HelpTip>
              </span>
              <input
                v-model.number="editForm.alert_duration_minutes"
                type="number"
                min="0"
                max="1440"
                step="1"
                class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
              />
            </label>
            <label class="block">
              <span class="inline-flex items-center text-xs text-zinc-400">
                Alert cooldown (minutes)
                <HelpTip position="top">
                  Minimum quiet window before another alert can fire for this sensor.
                  Example: set to <strong>60</strong> to mean "after alerting, stay quiet for an hour
                  even if the value stays out of bounds". Defaults to 5 minutes.
                </HelpTip>
              </span>
              <input
                v-model.number="editForm.alert_cooldown_minutes"
                type="number"
                min="0"
                max="10080"
                step="1"
                class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
              />
            </label>

            <div v-if="editError" class="md:col-span-2 text-sm text-red-400">{{ editError }}</div>

            <div class="md:col-span-2 flex items-center gap-2 justify-end">
              <button
                type="button"
                :disabled="editSaving"
                @click="cancelEditAlerts"
                class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-400 hover:text-white disabled:opacity-50"
              >Cancel</button>
              <button
                type="submit"
                :disabled="editSaving"
                class="text-xs px-3 py-1.5 rounded-lg bg-green-700 border border-green-600 text-white hover:bg-green-600 disabled:opacity-50"
              >{{ editSaving ? 'Saving…' : 'Save' }}</button>
            </div>
          </form>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import {
  Chart as ChartJS,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import api from '../api'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'

ChartJS.register(LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const sensor = ref(null)
const stats = ref(null)
const readings = ref([])
const loadError = ref(null)
const historyLoading = ref(true)
const rangeHours = ref(24)

const editingAlerts = ref(false)
const editSaving = ref(false)
const editError = ref(null)
const editForm = ref({
  alert_threshold_low: '',
  alert_threshold_high: '',
  alert_duration_minutes: 0,
  alert_cooldown_minutes: 5,
})

function formatSecondsAsMinutes(secs) {
  if (secs == null || !Number.isFinite(Number(secs))) return '—'
  const n = Number(secs)
  if (n === 0) return '0 min (immediate)'
  if (n % 60 === 0) return `${n / 60} min`
  return `${(n / 60).toFixed(1)} min`
}

function beginEditAlerts() {
  if (!sensor.value) return
  editError.value = null
  editForm.value = {
    alert_threshold_low: sensor.value.alert_threshold_low ?? '',
    alert_threshold_high: sensor.value.alert_threshold_high ?? '',
    alert_duration_minutes: Math.round((sensor.value.alert_duration_seconds ?? 0) / 60),
    alert_cooldown_minutes: Math.round((sensor.value.alert_cooldown_seconds ?? 300) / 60),
  }
  editingAlerts.value = true
}

function cancelEditAlerts() {
  editingAlerts.value = false
  editError.value = null
}

function parseThreshold(v) {
  if (v === '' || v == null) return null
  const n = Number(v)
  if (!Number.isFinite(n)) return undefined
  return n
}

async function saveAlerts() {
  if (!sensor.value) return
  editError.value = null

  const low = parseThreshold(editForm.value.alert_threshold_low)
  const high = parseThreshold(editForm.value.alert_threshold_high)
  if (low === undefined || high === undefined) {
    editError.value = 'Thresholds must be numbers or blank.'
    return
  }

  const durMin = Number(editForm.value.alert_duration_minutes)
  const coolMin = Number(editForm.value.alert_cooldown_minutes)
  if (!Number.isFinite(durMin) || durMin < 0 || durMin > 1440) {
    editError.value = 'Duration must be between 0 and 1440 minutes.'
    return
  }
  if (!Number.isFinite(coolMin) || coolMin < 0 || coolMin > 10080) {
    editError.value = 'Cooldown must be between 0 and 10080 minutes.'
    return
  }

  const payload = {
    alert_threshold_low: low,
    alert_threshold_high: high,
    alert_duration_seconds: Math.round(durMin * 60),
    alert_cooldown_seconds: Math.round(coolMin * 60),
  }

  editSaving.value = true
  try {
    const r = await api.put(`/sensors/${sensor.value.id}`, payload)
    sensor.value = r.data
    editingAlerts.value = false
  } catch (e) {
    editError.value = e.response?.data?.error || e.message || 'Failed to save sensor settings'
  } finally {
    editSaving.value = false
  }
}

const rangeOptions = [
  { hours: 1, label: '1h' },
  { hours: 6, label: '6h' },
  { hours: 24, label: '24h' },
  { hours: 168, label: '7d' },
]

const ICONS = {
  temperature: '\u{1F321}\uFE0F',
  humidity: '\u{1F4A7}',
  co2: '\u{1FAE8}',
  ph: '\u{2697}\uFE0F',
  ec: '\u{26A1}',
  light: '\u2600\uFE0F',
  moisture: '\u{1F331}',
  pressure: '\u{1F535}',
  flow: '\u{1F30A}',
  default: '\u{1F4E1}',
}

function sensorIcon(type) {
  if (!type) return ICONS.default
  const k = type.toLowerCase()
  for (const [n, i] of Object.entries(ICONS)) {
    if (n === 'default') continue
    if (k.includes(n)) return i
  }
  return ICONS.default
}

function rangeToSinceUntil(hours) {
  const until = new Date()
  const since = new Date(until.getTime() - hours * 3600 * 1000)
  return { since: since.toISOString(), until: until.toISOString() }
}

function readingY(r) {
  const v = r.value_normalized ?? r.value_raw
  if (v == null) return null
  const n = typeof v === 'number' ? v : parseFloat(v)
  return Number.isFinite(n) ? n : null
}

const zoneName = computed(() => {
  if (!sensor.value?.zone_id) return ''
  return store.zones.find(z => z.id === sensor.value.zone_id)?.name ?? ''
})

const currentDisplay = computed(() => {
  if (!sensor.value) return '—'
  const r = store.readings[sensor.value.id]
  const y = r ? readingY(r) : null
  if (y == null) return '—'
  return y.toFixed(2)
})

const statCards = computed(() => {
  const s = stats.value
  if (!s) {
    return [
      { label: 'Avg', value: '—' },
      { label: 'Min', value: '—' },
      { label: 'Max', value: '—' },
      { label: 'Readings', value: '—' },
    ]
  }
  const fmt = (x) => (x == null || Number.isNaN(x) ? '—' : Number(x).toFixed(2))
  return [
    { label: 'Avg', value: fmt(s.avg) },
    { label: 'Min', value: fmt(s.min) },
    { label: 'Max', value: fmt(s.max) },
    { label: 'Readings', value: String(s.count ?? 0) },
  ]
})

const chartData = computed(() => {
  const pts = []
  for (const r of readings.value) {
    const y = readingY(r)
    if (y == null || !r.reading_time) continue
    const x = new Date(r.reading_time).getTime()
    pts.push({ x, y })
  }
  pts.sort((a, b) => a.x - b.x)

  const datasets = [
    {
      label: 'Reading',
      data: pts,
      borderColor: 'rgba(74, 222, 128, 0.9)',
      backgroundColor: 'rgba(74, 222, 128, 0.08)',
      fill: true,
      tension: 0.2,
      pointRadius: pts.length > 120 ? 0 : 2,
      pointHoverRadius: 4,
    },
  ]

  const s = sensor.value
  if (s && pts.length) {
    const xMin = pts[0].x
    const xMax = pts[pts.length - 1].x
    const low = s.alert_threshold_low != null ? Number(s.alert_threshold_low) : null
    const high = s.alert_threshold_high != null ? Number(s.alert_threshold_high) : null
    if (low != null && Number.isFinite(low)) {
      datasets.push({
        label: 'Low threshold',
        data: [{ x: xMin, y: low }, { x: xMax, y: low }],
        borderColor: 'rgba(248, 113, 113, 0.85)',
        borderWidth: 1.5,
        borderDash: [6, 4],
        pointRadius: 0,
        tension: 0,
        fill: false,
      })
    }
    if (high != null && Number.isFinite(high)) {
      datasets.push({
        label: 'High threshold',
        data: [{ x: xMin, y: high }, { x: xMax, y: high }],
        borderColor: 'rgba(251, 191, 36, 0.9)',
        borderWidth: 1.5,
        borderDash: [6, 4],
        pointRadius: 0,
        tension: 0,
        fill: false,
      })
    }
  }

  return { datasets }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index', intersect: false },
  parsing: false,
  scales: {
    x: {
      type: 'linear',
      title: { display: true, text: 'Time', color: '#71717a' },
      ticks: {
        color: '#a1a1aa',
        maxTicksLimit: 8,
        callback(v) {
          return new Date(v).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
        },
      },
      grid: { color: 'rgba(63, 63, 70, 0.35)' },
    },
    y: {
      title: { display: true, text: 'Value', color: '#71717a' },
      ticks: { color: '#a1a1aa' },
      grid: { color: 'rgba(63, 63, 70, 0.35)' },
    },
  },
  plugins: {
    legend: {
      labels: { color: '#a1a1aa' },
    },
    tooltip: {
      callbacks: {
        title(items) {
          const x = items[0]?.parsed?.x
          if (x == null) return ''
          return new Date(x).toLocaleString()
        },
      },
    },
  },
}

async function loadSensor(id) {
  loadError.value = null
  sensor.value = null
  try {
    const r = await api.get(`/sensors/${id}`)
    sensor.value = r.data
  } catch (e) {
    loadError.value = e.response?.data?.error || e.message || 'Failed to load sensor'
  }
}

function csvEscape(s) {
  if (s == null) return ''
  const t = String(s)
  if (/[",\n]/.test(t)) return `"${t.replace(/"/g, '""')}"`
  return t
}

function exportReadingsCsv() {
  const id = route.params.id
  if (!id || !readings.value.length) return
  const { since, until } = rangeToSinceUntil(rangeHours.value)
  const header = ['reading_time', 'value_raw', 'value_normalized', 'is_valid']
  const lines = [header.join(',')]
  for (const r of readings.value) {
    lines.push([
      csvEscape(r.reading_time),
      csvEscape(r.value_raw),
      csvEscape(r.value_normalized),
      csvEscape(r.is_valid),
    ].join(','))
  }
  const blob = new Blob([lines.join('\n')], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sensor-${id}-readings-${rangeHours.value}h.csv`
  a.click()
  URL.revokeObjectURL(url)
}

async function fetchHistory() {
  const id = route.params.id
  if (!id) return
  const { since, until } = rangeToSinceUntil(rangeHours.value)
  historyLoading.value = true
  try {
    loadError.value = null
    const [st, rs] = await Promise.all([
      store.loadSensorStats(id, { since, until }),
      store.loadSensorReadings(id, { since, until, limit: 2000 }),
    ])
    stats.value = st
    readings.value = rs
  } catch (e) {
    stats.value = null
    readings.value = []
    loadError.value = e.response?.data?.error || e.message || 'Failed to load history'
  } finally {
    historyLoading.value = false
  }
}

function setRange(hours) {
  rangeHours.value = hours
}

watch(() => route.params.id, async (id) => {
  if (!id) return
  await loadSensor(id)
  await fetchHistory()
  try {
    const r = await api.get(`/sensors/${id}/readings/latest`)
    store.readings[id] = r.data
  } catch { /* ignore */ }
})

watch(rangeHours, () => {
  fetchHistory()
})

onMounted(async () => {
  if (farmContext.farmId && !store.sensors.length) {
    try {
      await store.loadAll(farmContext.farmId)
    } catch { /* optional */ }
  }
  const id = route.params.id
  if (id) {
    await loadSensor(id)
    await fetchHistory()
    try {
      const r = await api.get(`/sensors/${id}/readings/latest`)
      store.readings[id] = r.data
    } catch { /* ignore */ }
  }
})
</script>
