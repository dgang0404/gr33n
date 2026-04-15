<template>
  <div class="p-6 space-y-6">
    <div>
      <router-link to="/sensors" class="text-xs text-zinc-500 hover:text-zinc-300">&larr; Back to sensors</router-link>
      <div v-if="loadError" class="mt-4 text-red-400 text-sm">{{ loadError }}</div>
      <div v-else-if="!sensor" class="mt-4 text-zinc-500 text-sm">Loading sensor…</div>
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

        <div class="flex flex-wrap gap-2 mt-4">
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
        </div>

        <div v-if="statsLoading" class="text-zinc-500 text-sm">Loading stats…</div>
        <div v-else class="grid grid-cols-2 md:grid-cols-4 gap-3 mt-2">
          <div v-for="card in statCards" :key="card.label" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
            <p class="text-zinc-500 text-xs mb-1">{{ card.label }}</p>
            <p class="text-white text-lg font-mono tabular-nums">{{ card.value }}</p>
          </div>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h2 class="text-sm font-semibold text-white mb-3">History</h2>
          <p v-if="chartLoading" class="text-zinc-500 text-sm">Loading chart…</p>
          <p v-else-if="!chartData.datasets[0]?.data?.length" class="text-zinc-500 text-sm">No readings in this range.</p>
          <div v-else class="h-80 w-full">
            <Line :data="chartData" :options="chartOptions" />
          </div>
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

ChartJS.register(LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const sensor = ref(null)
const stats = ref(null)
const readings = ref([])
const loadError = ref(null)
const statsLoading = ref(true)
const chartLoading = ref(true)
const rangeHours = ref(24)

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

async function fetchHistory() {
  const id = route.params.id
  if (!id) return
  const { since, until } = rangeToSinceUntil(rangeHours.value)
  statsLoading.value = true
  chartLoading.value = true
  try {
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
    statsLoading.value = false
    chartLoading.value = false
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
