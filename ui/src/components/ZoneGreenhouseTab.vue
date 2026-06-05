<template>
  <div class="space-y-6">
    <p v-if="saveError" class="text-red-400 text-sm">{{ saveError }}</p>
    <p v-if="saveOk" class="text-green-400 text-sm">Climate profile saved.</p>
    <p v-if="commandError" class="text-red-400 text-sm">{{ commandError }}</p>
    <p v-if="commandOk" class="text-green-400 text-sm">{{ commandOk }}</p>

    <div
      v-if="missingLuxBanner"
      class="bg-amber-950/40 border border-amber-800/60 rounded-xl px-4 py-3 text-sm text-amber-200"
    >
      <strong>Auto shade disabled</strong> — no <strong>lux</strong> or <strong>par</strong> sensor in this zone.
      Add a sensor, set policy to <strong>manual</strong>, or apply GH templates with
      “skip high-lux without sensor”.
    </div>
    <div
      v-if="missingTempBanner"
      class="bg-amber-950/40 border border-amber-800/60 rounded-xl px-4 py-3 text-sm text-amber-200"
    >
      High-temp fan and night-retract templates need a <strong>temperature</strong> sensor in this zone.
    </div>
    <div
      v-if="profile.automation_policy === 'auto' && !profile.shade_actuator_id && !(profile.fan_actuator_ids || []).length"
      class="bg-zinc-900 border border-zinc-700 rounded-xl px-4 py-3 text-sm text-zinc-400"
    >
      <code>automation_policy=auto</code> but no shade or fan actuators linked — link actuators below
      before enabling GH rules.
    </div>

    <!-- Climate profile -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-4">
      <h2 class="text-sm font-semibold text-white">Greenhouse climate profile</h2>
      <p class="text-zinc-500 text-xs">
        Block sun and ventilation are separate from supplemental lighting
        (<router-link to="/lighting" class="text-green-600 hover:text-green-400">Lighting programs</router-link>).
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <label class="block">
          <span class="text-zinc-400 text-xs">Cover type</span>
          <select v-model="profile.cover_type" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option value="">—</option>
            <option value="glass">Glass</option>
            <option value="polycarbonate">Polycarbonate</option>
            <option value="film">Film</option>
          </select>
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Automation policy</span>
          <select v-model="profile.automation_policy" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option value="">—</option>
            <option value="auto">Auto (sensor rules)</option>
            <option value="manual">Manual only</option>
            <option value="schedule_only">Schedule only</option>
          </select>
        </label>
        <label class="block md:col-span-2">
          <span class="text-zinc-400 text-xs">Notes</span>
          <textarea v-model="profile.notes" rows="2" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" placeholder="Glazing, orientation, shade brand…" />
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Shade actuator</span>
          <select v-model.number="profile.shade_actuator_id" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option :value="null">—</option>
            <option v-for="a in shadeActuatorOptions" :key="a.id" :value="a.id">{{ a.name }} ({{ a.actuator_type }})</option>
          </select>
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Vent actuator</span>
          <select v-model.number="profile.vent_actuator_id" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
            <option :value="null">—</option>
            <option v-for="a in ventActuatorOptions" :key="a.id" :value="a.id">{{ a.name }} ({{ a.actuator_type }})</option>
          </select>
        </label>
        <div class="block md:col-span-2">
          <span class="text-zinc-400 text-xs">Fan actuators</span>
          <div class="mt-2 flex flex-wrap gap-3">
            <label v-for="a in fanActuatorOptions" :key="a.id" class="flex items-center gap-2 text-sm text-zinc-300">
              <input type="checkbox" :value="a.id" v-model="profile.fan_actuator_ids" class="rounded border-zinc-600" />
              {{ a.name }}
            </label>
            <p v-if="!fanActuatorOptions.length" class="text-zinc-600 text-xs">No exhaust/circulation fans in this zone.</p>
          </div>
        </div>
      </div>
      <button
        type="button"
        class="px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-white text-sm font-medium disabled:opacity-40"
        :disabled="saving"
        @click="saveProfile"
      >
        {{ saving ? 'Saving…' : 'Save profile' }}
      </button>
    </div>

    <!-- Climate sensors -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h2 class="text-sm font-semibold text-white mb-3">Climate sensors</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <SensorTile
          v-for="s in climateSensors" :key="s.id"
          :sensor="s"
          :reading="store.readings[s.id]"
        />
        <div
          v-for="slot in climatePlaceholders" :key="slot.type"
          class="bg-zinc-950 border border-dashed border-zinc-700 rounded-xl p-4 text-zinc-500 text-sm"
        >
          <p class="text-xs uppercase tracking-wide">{{ slot.label }}</p>
          <p class="mt-2">No {{ slot.label.toLowerCase() }} sensor</p>
        </div>
      </div>
    </div>

    <!-- Typed manual controls -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h2 class="text-sm font-semibold text-white mb-3">Manual controls</h2>
      <p class="text-zinc-500 text-xs mb-3">
        Commands queue on the Pi via <code class="text-zinc-400">pending_command</code> (same path as automation).
      </p>
      <p v-if="!climateActuators.length" class="text-zinc-500 text-sm">No greenhouse actuators in this zone.</p>
      <div v-else class="space-y-3">
        <div
          v-for="a in climateActuators" :key="a.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
        >
          <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
            <div>
              <p class="text-white text-sm font-medium">{{ a.name }}</p>
              <p class="text-zinc-500 text-xs capitalize">{{ a.actuator_type }}
                <span v-if="!a.device_id" class="text-amber-500"> · no device linked</span>
              </p>
            </div>
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="cmd in commandsFor(a)"
              :key="cmd"
              type="button"
              class="px-3 py-1.5 rounded-lg text-xs font-medium border border-zinc-700 hover:border-green-600 hover:text-green-300 text-zinc-300 disabled:opacity-40"
              :disabled="!a.device_id || commanding[a.id]"
              @click="sendCommand(a, cmd)"
            >
              {{ cmd }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- GH automations -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex flex-wrap items-center justify-between gap-2 mb-3">
        <h2 class="text-sm font-semibold text-white">Greenhouse automations</h2>
        <div class="flex flex-wrap items-center gap-2">
          <label v-if="canApplyTemplates" class="flex items-center gap-1.5 text-xs text-zinc-400">
            <input v-model="allowMissingLux" type="checkbox" class="rounded border-zinc-600" />
            Skip high-lux if no PAR sensor
          </label>
          <button
            v-if="canApplyTemplates"
            type="button"
            class="px-3 py-1.5 rounded-lg text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 disabled:opacity-40"
            :disabled="applyingTemplates"
            @click="applyGhTemplates"
          >
            {{ applyingTemplates ? 'Applying…' : 'Clone GH templates' }}
          </button>
          <router-link to="/automation" class="text-xs text-green-600 hover:text-green-400">Automation →</router-link>
        </div>
      </div>
      <p v-if="templateMsg" class="text-xs mb-2" :class="templateError ? 'text-red-400' : 'text-green-400'">{{ templateMsg }}</p>
      <p v-if="rulesLoading" class="text-zinc-500 text-sm">Loading rules…</p>
      <p v-else-if="!ghRules.length" class="text-zinc-500 text-sm">
        No <code class="text-zinc-400">GH —</code> rules yet. Apply bootstrap
        <code class="text-zinc-400">greenhouse_climate_v1</code> or clone templates from Settings.
      </p>
      <ul v-else class="space-y-2">
        <li
          v-for="r in ghRules" :key="r.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 flex items-center justify-between gap-2"
        >
          <div class="min-w-0">
            <p class="text-zinc-200 text-sm truncate">{{ r.name }}</p>
            <p class="text-zinc-600 text-xs line-clamp-2">{{ r.description || '—' }}</p>
          </div>
          <span
            class="shrink-0 text-xs px-2 py-0.5 rounded capitalize"
            :class="r.is_active ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-500'"
          >
            {{ r.is_active ? 'active' : 'inactive' }}
          </span>
        </li>
      </ul>
    </div>

    <!-- Recent shade/fan events -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h2 class="text-sm font-semibold text-white mb-3">Recent shade / vent / fan events</h2>
      <p v-if="!ghEvents.length" class="text-zinc-500 text-sm">No recent events for greenhouse actuators.</p>
      <div v-else class="space-y-2 max-h-48 overflow-y-auto">
        <div
          v-for="ev in ghEvents" :key="ev.event_time + '-' + ev.actuator_id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-xs"
        >
          <span class="text-zinc-200">{{ actuatorName(ev.actuator_id) }}</span>
          <span class="text-zinc-500"> → {{ ev.command_sent }}</span>
          <span class="text-zinc-600 ml-2">{{ formatTime(ev.event_time) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import api from '../api'
import { useFarmStore } from '../stores/farm'
import SensorTile from './SensorTile.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  zone: { type: Object, required: true },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  actuatorEvents: { type: Array, default: () => [] },
})

const emit = defineEmits(['refresh-events'])

const store = useFarmStore()
const saving = ref(false)
const saveError = ref('')
const saveOk = ref(false)
const commandError = ref('')
const commandOk = ref('')
const commanding = ref({})
const rulesLoading = ref(false)
const ghRules = ref([])
const applyingTemplates = ref(false)
const allowMissingLux = ref(false)
const templateMsg = ref('')
const templateError = ref(false)

const GH_ACTUATOR_TYPES = new Set([
  'shade_screen', 'ridge_vent', 'exhaust_fan', 'circulation_fan',
  'glazing_panel', 'shade_cloth_motor',
])

const CLIMATE_SENSOR_TYPES = [
  { match: t => t === 'lux' || t.includes('lux'), label: 'Lux' },
  { match: t => t === 'par' || t.includes('par'), label: 'PAR' },
  { match: t => t.includes('temp'), label: 'Temperature' },
  { match: t => t.includes('humid') || t === 'rh', label: 'Humidity' },
]

function parseMeta(meta) {
  if (!meta) return {}
  if (typeof meta === 'string') {
    try { return JSON.parse(meta) } catch { return {} }
  }
  return meta
}

function defaultProfile() {
  return {
    cover_type: '',
    automation_policy: '',
    notes: '',
    shade_actuator_id: null,
    vent_actuator_id: null,
    fan_actuator_ids: [],
  }
}

const profile = reactive(defaultProfile())

function loadProfileFromZone() {
  const m = parseMeta(props.zone?.meta_data)
  const gc = m.greenhouse_climate || {}
  profile.cover_type = gc.cover_type || ''
  profile.automation_policy = gc.automation_policy || ''
  profile.notes = gc.notes || ''
  profile.shade_actuator_id = gc.shade_actuator_id ?? null
  profile.vent_actuator_id = gc.vent_actuator_id ?? null
  profile.fan_actuator_ids = Array.isArray(gc.fan_actuator_ids) ? [...gc.fan_actuator_ids] : []
}

watch(() => props.zone, loadProfileFromZone, { immediate: true, deep: true })

const farmActuators = computed(() => store.actuators.filter(a => !a.deleted_at))
const shadeActuatorOptions = computed(() =>
  farmActuators.value.filter(a => a.zone_id === props.zoneId &&
    ['shade_screen', 'shade_cloth_motor'].includes(a.actuator_type)))
const ventActuatorOptions = computed(() =>
  farmActuators.value.filter(a => a.zone_id === props.zoneId &&
    ['ridge_vent', 'glazing_panel'].includes(a.actuator_type)))
const fanActuatorOptions = computed(() =>
  farmActuators.value.filter(a => a.zone_id === props.zoneId &&
    ['exhaust_fan', 'circulation_fan'].includes(a.actuator_type)))

const climateActuators = computed(() =>
  props.actuators.filter(a => GH_ACTUATOR_TYPES.has(a.actuator_type)))

const climateSensors = computed(() => {
  const used = new Set()
  const out = []
  for (const slot of CLIMATE_SENSOR_TYPES) {
    const s = props.sensors.find(s => {
      const t = String(s.sensor_type || s.type || '').toLowerCase()
      return slot.match(t)
    })
    if (s) {
      used.add(s.id)
      out.push(s)
    }
  }
  return out
})

const climatePlaceholders = computed(() => {
  const have = new Set(climateSensors.value.map(s => {
    const t = String(s.sensor_type || s.type || '').toLowerCase()
    if (t.includes('lux')) return 'lux'
    if (t.includes('par')) return 'par'
    if (t.includes('temp')) return 'temp'
    if (t.includes('humid') || t === 'rh') return 'humid'
    return t
  }))
  const slots = []
  if (!have.has('lux') && !have.has('par')) slots.push({ type: 'lux', label: 'Lux / PAR' })
  if (!have.has('temp')) slots.push({ type: 'temp', label: 'Temperature' })
  if (!have.has('humid')) slots.push({ type: 'humid', label: 'Humidity' })
  return slots
})

function zoneHasLuxSensor() {
  return props.sensors.some(s => {
    const t = String(s.sensor_type || s.type || '').toLowerCase()
    return t === 'lux' || t.includes('lux') || t === 'par' || t.includes('par')
  })
}

function zoneHasTempSensor() {
  return props.sensors.some(s => {
    const t = String(s.sensor_type || s.type || '').toLowerCase()
    return t.includes('temp')
  })
}

const missingLuxBanner = computed(() => profile.automation_policy === 'auto' && !zoneHasLuxSensor())
const missingTempBanner = computed(() => profile.automation_policy === 'auto' && !zoneHasTempSensor())

const canApplyTemplates = computed(() =>
  profile.shade_actuator_id || (profile.fan_actuator_ids || []).length)

function firstZoneLuxSensorId() {
  const s = props.sensors.find(s => {
    const t = String(s.sensor_type || s.type || '').toLowerCase()
    return t === 'lux' || t.includes('lux') || t === 'par' || t.includes('par')
  })
  return s?.id ?? null
}

function firstZoneTempSensorId() {
  const s = props.sensors.find(s => {
    const t = String(s.sensor_type || s.type || '').toLowerCase()
    return t.includes('temp')
  })
  return s?.id ?? null
}

async function applyGhTemplates() {
  applyingTemplates.value = true
  templateMsg.value = ''
  templateError.value = false
  try {
    const fanIds = profile.fan_actuator_ids || []
    const body = {
      zone_id: props.zoneId,
      shade_actuator_id: profile.shade_actuator_id || null,
      fan_actuator_id: fanIds[0] || null,
      lux_sensor_id: firstZoneLuxSensorId(),
      temp_sensor_id: firstZoneTempSensorId(),
      allow_missing_lux_sensor: allowMissingLux.value,
      allow_missing_temp_sensor: !firstZoneTempSensorId(),
    }
    const res = await api.post(`/farms/${props.farmId}/automation/rule-templates/greenhouse`, body)
    const skipped = res.data?.skipped_rule_families || []
    const n = res.data?.rules_created ?? 0
    templateMsg.value = skipped.length
      ? `Created ${n} rule(s); skipped: ${skipped.join(', ')}. Enable on Automation when sensors are ready.`
      : `Created ${n} GH rule(s). They start inactive — review on Automation.`
    await loadGhRules()
  } catch (e) {
    templateError.value = true
    templateMsg.value = e.response?.data?.error || e.message || 'Template apply failed'
  } finally {
    applyingTemplates.value = false
  }
}

const ghEvents = computed(() => {
  const ids = new Set(climateActuators.value.map(a => a.id))
  return props.actuatorEvents.filter(ev => ids.has(ev.actuator_id)).slice(0, 15)
})

function commandsFor(a) {
  if (Array.isArray(a.valid_commands) && a.valid_commands.length) return a.valid_commands
  if (['shade_screen', 'shade_cloth_motor'].includes(a.actuator_type)) return ['deploy', 'retract', 'stop']
  if (['ridge_vent', 'glazing_panel'].includes(a.actuator_type)) return ['open', 'close', 'stop']
  return ['on', 'off']
}

async function saveProfile() {
  saving.value = true
  saveError.value = ''
  saveOk.value = false
  try {
    const meta = { ...parseMeta(props.zone.meta_data) }
    const gc = {
      cover_type: profile.cover_type || undefined,
      automation_policy: profile.automation_policy || undefined,
      notes: profile.notes || undefined,
      shade_actuator_id: profile.shade_actuator_id || undefined,
      vent_actuator_id: profile.vent_actuator_id || undefined,
      fan_actuator_ids: profile.fan_actuator_ids?.length ? profile.fan_actuator_ids : undefined,
    }
    Object.keys(gc).forEach(k => { if (gc[k] === undefined) delete gc[k] })
    meta.greenhouse_climate = gc
    await api.put(`/zones/${props.zoneId}`, {
      name: props.zone.name,
      description: props.zone.description ?? null,
      zone_type: props.zone.zone_type,
      meta_data: meta,
    })
    await store.loadAll(props.farmId)
    saveOk.value = true
    setTimeout(() => { saveOk.value = false }, 3000)
  } catch (e) {
    saveError.value = e.response?.data?.error || e.message || 'Save failed'
  } finally {
    saving.value = false
  }
}

async function sendCommand(a, cmd) {
  commanding.value[a.id] = true
  commandError.value = ''
  commandOk.value = ''
  try {
    const res = await store.enqueueActuatorCommand(a.id, cmd, `Zone ${props.zone?.name}: ${cmd}`)
    commandOk.value = `Queued "${cmd}" on ${res.device_name || 'device'} — Pi will execute on next poll.`
    emit('refresh-events')
    setTimeout(() => { commandOk.value = '' }, 5000)
  } catch (e) {
    commandError.value = e.response?.data?.error || e.message || 'Command failed'
  } finally {
    commanding.value[a.id] = false
  }
}

async function loadGhRules() {
  if (!props.farmId) return
  rulesLoading.value = true
  try {
    const rules = await store.loadAutomationRules(props.farmId)
    ghRules.value = (rules || []).filter(r => {
      const name = String(r.name || '')
      if (name.startsWith('GH —') || name.startsWith('GH -')) return true
      const zid = r.trigger_configuration?.zone_id
      return zid != null && Number(zid) === props.zoneId
    })
  } catch {
    ghRules.value = []
  } finally {
    rulesLoading.value = false
  }
}

function actuatorName(id) {
  return store.actuators.find(a => a.id === id)?.name || `Actuator ${id}`
}

function formatTime(ts) {
  if (!ts) return '—'
  const d = new Date(ts)
  const mins = Math.floor((Date.now() - d) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  return d.toLocaleString()
}

onMounted(loadGhRules)
</script>
