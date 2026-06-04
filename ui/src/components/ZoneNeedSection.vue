<template>
  <div class="space-y-6">
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-start justify-between gap-3 mb-2">
        <div>
          <h2 class="text-sm font-semibold text-white flex items-center gap-2">
            <span>{{ meta.icon }}</span>
            {{ meta.label }}
          </h2>
          <p class="text-zinc-500 text-xs mt-1">{{ meta.description }}</p>
        </div>
        <div class="flex flex-wrap gap-2 justify-end">
          <router-link
            v-for="link in meta.manageLinks"
            :key="link.to"
            :to="link.to"
            class="text-xs text-green-600 hover:text-green-400"
          >
            {{ link.label }} →
          </router-link>
        </div>
      </div>
      <p class="text-zinc-600 text-xs">
        How it connects: <strong class="text-zinc-400">sensor reading</strong> →
        <strong class="text-zinc-400">target band</strong> →
        <strong class="text-zinc-400">schedule or rule</strong> →
        <strong class="text-zinc-400">pump/light/fan</strong> →
        <strong class="text-zinc-400">device</strong>
      </p>
    </div>

    <!-- Live sensors -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-white mb-3">Sensors</h3>
      <p v-if="!needSensors.length" class="text-zinc-500 text-sm">No {{ meta.shortLabel.toLowerCase() }} sensors in this zone yet.</p>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <SensorTile
          v-for="s in needSensors"
          :key="s.id"
          :sensor="s"
          :reading="store.readings[s.id]"
        />
      </div>
    </div>

    <!-- Connection cards -->
    <div v-if="connectionCards.length" class="space-y-3">
      <h3 class="text-sm font-semibold text-white">What runs this</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <ZoneNeedConnectionCard
          v-for="(card, i) in connectionCards"
          :key="i"
          v-bind="card"
        />
      </div>
    </div>

    <!-- Fertigation block (water only) — Phase 39 WS7 enhanced -->
    <div v-if="need === PLANT_NEEDS.water" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">Feeding program</h3>
        <router-link to="/fertigation" class="text-xs text-green-600 hover:text-green-400">Fertigation →</router-link>
      </div>
      <p v-if="!activeProgram" class="text-zinc-500 text-sm">No active fertigation program for this zone.</p>
      <template v-else>
        <div class="flex items-start justify-between gap-2 flex-wrap">
          <div>
            <p class="text-zinc-200 text-sm">{{ activeProgram.name }}</p>
            <p class="text-zinc-600 text-xs mt-1">
              {{ activeProgram.total_volume_liters || '—' }}L
              <span v-if="activeProgram.run_duration_seconds"> · pump run {{ activeProgram.run_duration_seconds }}s</span>
            </p>
          </div>
          <!-- Queue depth chip -->
          <span
            v-if="waterStatus && waterStatus.queue_depth > 0"
            class="text-[10px] px-2 py-0.5 rounded-full bg-amber-900 text-amber-300 font-semibold shrink-0"
            title="Commands pending on delivery device"
          >
            {{ waterStatus.queue_depth }} queued
          </span>
        </div>

        <!-- Last mixing event badge -->
        <div v-if="waterStatus?.last_mixing_event" class="mt-3 flex items-center gap-2 text-xs text-zinc-400">
          <span>Last mix:</span>
          <span class="text-zinc-200">{{ formatMixDate(waterStatus.last_mixing_event.mixed_at) }}</span>
          <span
            v-if="waterStatus.last_mixing_event.ec_target_met === true"
            class="px-1.5 py-0.5 rounded bg-green-900 text-green-300 text-[10px] font-semibold"
          >EC met ✓</span>
          <span
            v-else-if="waterStatus.last_mixing_event.ec_target_met === false"
            class="px-1.5 py-0.5 rounded bg-red-900 text-red-300 text-[10px] font-semibold"
          >EC not met</span>
        </div>

        <!-- Preview mix plan -->
        <div class="mt-3">
          <button
            v-if="!showMixPreview && waterStatus?.mix_required"
            type="button"
            class="text-xs text-blue-400 hover:text-blue-300 underline"
            :disabled="mixPreviewLoading"
            @click="loadMixPreview"
          >
            {{ mixPreviewLoading ? 'Calculating…' : 'Preview mix plan →' }}
          </button>

          <div v-if="showMixPreview && waterStatus?.mix_preview" class="mt-2 space-y-1">
            <div class="flex items-center justify-between">
              <p class="text-xs text-zinc-400 font-semibold">Mix plan — {{ waterStatus.mix_preview.dilution_ratio }}</p>
              <button
                type="button"
                class="text-[10px] text-zinc-600 hover:text-zinc-400"
                @click="showMixPreview = false"
              >hide</button>
            </div>
            <p class="text-[11px] text-zinc-500">
              {{ waterStatus.mix_preview.water_volume_liters }}L ·
              base {{ waterStatus.mix_preview.water_ec_mscm }} mS/cm →
              est. {{ waterStatus.mix_preview.estimated_final_ec_mscm }} mS/cm
            </p>
            <div
              v-for="step in waterStatus.mix_preview.steps"
              :key="step.step"
              class="flex items-center gap-2 text-[11px] text-zinc-300"
            >
              <span class="w-4 text-zinc-600 text-right">{{ step.step }}.</span>
              <span class="flex-1">{{ step.input_name }}</span>
              <span class="text-zinc-500">{{ step.volume_ml }} ml · {{ step.run_seconds }}s</span>
            </div>
            <p v-if="waterStatus.mix_preview.warnings?.length"
               class="text-[10px] text-amber-500 mt-1 leading-tight">
              ⚠ {{ waterStatus.mix_preview.warnings[0] }}
            </p>
          </div>

          <p v-if="waterStatus?.mix_preview_error && !waterStatus?.mix_required"
             class="text-[11px] text-zinc-600 mt-1 italic">
            {{ waterStatus.mix_preview_error }}
          </p>
        </div>
      </template>
    </div>

    <!-- Lighting block (light only) -->
    <div v-if="need === PLANT_NEEDS.light" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">Lighting program</h3>
        <router-link to="/lighting" class="text-xs text-green-600 hover:text-green-400">Lighting →</router-link>
      </div>
      <p v-if="!zoneLightingPrograms.length" class="text-zinc-500 text-sm">No lighting program linked to this zone.</p>
      <ul v-else class="space-y-2">
        <li
          v-for="lp in zoneLightingPrograms"
          :key="lp.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-300"
        >
          {{ lp.name }}
          <span class="text-zinc-600 text-xs ml-2">{{ lp.is_active ? 'active' : 'inactive' }}</span>
        </li>
      </ul>
    </div>

    <!-- Greenhouse climate (air + greenhouse zone) -->
    <ZoneGreenhouseTab
      v-if="need === PLANT_NEEDS.air && isGreenhouse"
      :zone-id="zoneId"
      :farm-id="farmId"
      :zone="zone"
      :sensors="sensors"
      :actuators="actuators"
      :actuator-events="actuatorEvents"
      @refresh-events="$emit('refresh-events')"
    />

    <!-- Manual controls -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-white mb-3">Manual controls</h3>
      <p v-if="!needActuators.length" class="text-zinc-500 text-sm">No {{ meta.shortLabel.toLowerCase() }} actuators in this zone.</p>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div
          v-for="a in needActuators"
          :key="a.id"
          class="bg-zinc-950 border rounded-lg p-3"
          :class="a.current_state_text === 'online' ? 'border-green-800/70' : 'border-zinc-800'"
        >
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="text-white text-sm font-medium truncate">{{ a.name }}</p>
              <p class="text-zinc-500 text-xs capitalize">{{ a.actuator_type }}</p>
            </div>
            <button
              type="button"
              class="relative shrink-0 w-11 h-6 rounded-full transition-colors disabled:opacity-40"
              :class="a.current_state_text === 'online' ? 'bg-green-600' : 'bg-zinc-700'"
              :disabled="toggling[a.id]"
              @click="$emit('toggle-actuator', a)"
            >
              <span
                class="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200"
                :class="a.current_state_text === 'online' ? 'translate-x-5' : 'translate-x-0'"
              />
            </button>
          </div>
          <ActuatorPulseControl :actuator="a" />
        </div>
      </div>
    </div>

    <!-- Setpoints for this need -->
    <div v-if="needSetpoints.length" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">Targets</h3>
        <router-link to="/setpoints" class="text-xs text-green-600 hover:text-green-400">Setpoints →</router-link>
      </div>
      <div class="space-y-1">
        <div
          v-for="sp in needSetpoints"
          :key="sp.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 flex justify-between text-xs"
        >
          <span class="text-zinc-200">{{ sp.sensor_type }}</span>
          <span class="text-zinc-500">
            <span v-if="sp.min_value != null">min {{ sp.min_value }}</span>
            <span v-if="sp.ideal_value != null"> · ideal {{ sp.ideal_value }}</span>
            <span v-if="sp.max_value != null"> · max {{ sp.max_value }}</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import {
  PLANT_NEEDS,
  NEED_META,
  sensorPlantNeed,
  actuatorPlantNeed,
} from '../lib/plantNeeds.js'
import { useFarmStore } from '../stores/farm.js'
import SensorTile from './SensorTile.vue'
import ZoneNeedConnectionCard from './ZoneNeedConnectionCard.vue'
import ActuatorPulseControl from './ActuatorPulseControl.vue'
import ZoneGreenhouseTab from './ZoneGreenhouseTab.vue'

const props = defineProps({
  need: { type: String, required: true },
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  zone: { type: Object, default: null },
  isGreenhouse: { type: Boolean, default: false },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  setpoints: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  rules: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  lightingPrograms: { type: Array, default: () => [] },
  activeProgram: { type: Object, default: null },
  actuatorEvents: { type: Array, default: () => [] },
  toggling: { type: Object, default: () => ({}) },
})

defineEmits(['toggle-actuator', 'refresh-events'])

const store = useFarmStore()

// ── Phase 39 WS7: water status (mix preview, queue depth, last mix) ──────────
const waterStatus = ref(null)
const mixPreviewLoading = ref(false)
const showMixPreview = ref(false)

async function loadWaterStatus() {
  if (props.need !== PLANT_NEEDS.water || !props.activeProgram?.id) return
  try {
    const token = store.token || localStorage.getItem('token')
    const r = await fetch(`/fertigation/programs/${props.activeProgram.id}/water-status`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (r.ok) waterStatus.value = await r.json()
  } catch {
    // non-fatal
  }
}

async function loadMixPreview() {
  mixPreviewLoading.value = true
  await loadWaterStatus()
  mixPreviewLoading.value = false
  showMixPreview.value = true
}

// Load on mount when water tab is active
watch(
  () => [props.activeProgram?.id, props.need],
  () => { if (props.need === PLANT_NEEDS.water) loadWaterStatus() },
  { immediate: true },
)

function formatMixDate(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}
const meta = computed(() => NEED_META[props.need] || NEED_META[PLANT_NEEDS.air])

const needSensors = computed(() =>
  props.sensors.filter(s => sensorPlantNeed(s.sensor_type) === props.need),
)

const needActuators = computed(() =>
  props.actuators.filter(a => actuatorPlantNeed(a.actuator_type) === props.need),
)

const needSetpoints = computed(() => {
  const types = new Set(needSensors.value.map(s => s.sensor_type))
  return props.setpoints.filter(sp => types.has(sp.sensor_type))
})

const zoneLightingPrograms = computed(() =>
  props.lightingPrograms.filter(lp => lp.zone_id === props.zoneId),
)

function formatReading(sensor) {
  const r = store.readings[sensor.id]
  if (!r || r.value_raw == null) return 'No reading yet'
  const unit = sensor.unit_of_measure || ''
  return `${r.value_raw}${unit ? ` ${unit}` : ''}`
}

function formatSetpoint(sp) {
  const parts = []
  if (sp.min_value != null) parts.push(`min ${sp.min_value}`)
  if (sp.ideal_value != null) parts.push(`ideal ${sp.ideal_value}`)
  if (sp.max_value != null) parts.push(`max ${sp.max_value}`)
  return parts.length ? parts.join(' · ') : 'Not set'
}

function lastEventForActuator(actuatorId) {
  const ev = props.actuatorEvents.find(e => e.actuator_id === actuatorId)
  if (!ev) return ''
  const cmd = ev.command_sent || '—'
  const src = ev.source || ''
  return `${cmd} (${src})`
}

const connectionCards = computed(() => {
  const cards = []

  for (const s of needSensors.value.slice(0, 4)) {
    const sp = props.setpoints.find(x => x.sensor_type === s.sensor_type)
    cards.push({
      title: s.name || s.sensor_type,
      subtitle: s.sensor_type,
      manageTo: `/sensors/${s.id}`,
      readingLabel: formatReading(s),
      targetLabel: sp ? formatSetpoint(sp) : 'No setpoint — add under Setpoints',
      automationLabel: 'Alerts use sensor thresholds; rules use Setpoints page',
      controlLabel: '—',
      lastEventLabel: '',
    })
  }

  if (props.need === PLANT_NEEDS.water && props.activeProgram) {
    const sched = props.schedules.find(s => s.id === props.activeProgram.schedule_id)
    const pump = needActuators.value[0]
    cards.unshift({
      title: props.activeProgram.name,
      subtitle: 'Fertigation program',
      manageTo: '/fertigation',
      readingLabel: needSensors.value.length ? formatReading(needSensors.value[0]) : 'Add EC/pH sensor',
      targetLabel: props.activeProgram.run_duration_seconds
        ? `Pump on ${props.activeProgram.run_duration_seconds}s per run`
        : 'Volume-based feed',
      automationLabel: sched ? `${sched.name} (${sched.is_active ? 'active' : 'inactive'})` : 'No schedule linked',
      controlLabel: pump ? `${pump.name} — ${pump.current_state_text || 'offline'}` : 'Assign a pump actuator',
      controlOnline: pump?.current_state_text === 'online',
      lastEventLabel: pump ? lastEventForActuator(pump.id) : '',
    })
  }

  if (props.need === PLANT_NEEDS.light) {
    for (const lp of zoneLightingPrograms.value.slice(0, 2)) {
      const light = needActuators.value.find(a =>
        a.id === lp.actuator_id || (a.actuator_type || '').includes('light'),
      )
      cards.push({
        title: lp.name,
        subtitle: 'Lighting program',
        manageTo: '/lighting',
        readingLabel: needSensors.value[0] ? formatReading(needSensors.value[0]) : 'Optional lux/PAR sensor',
        targetLabel: `${lp.on_hours ?? '—'}h on / ${lp.off_hours ?? '—'}h off`,
        automationLabel: lp.is_active ? 'Active — ON/OFF schedules fire automatically' : 'Inactive',
        controlLabel: light ? `${light.name} — ${light.current_state_text || 'offline'}` : 'Link grow light actuator',
        controlOnline: light?.current_state_text === 'online',
        lastEventLabel: light ? lastEventForActuator(light.id) : '',
      })
    }
  }

  if (props.need === PLANT_NEEDS.air) {
    const zoneRules = props.rules.filter(r => {
      try {
        const tc = typeof r.trigger_configuration === 'string'
          ? JSON.parse(r.trigger_configuration)
          : r.trigger_configuration
        return tc?.zone_id === props.zoneId
      } catch {
        return false
      }
    })
    for (const r of zoneRules.slice(0, 3)) {
      const fan = needActuators.value[0]
      cards.push({
        title: r.name,
        subtitle: 'Automation rule',
        manageTo: '/automation',
        readingLabel: needSensors.value[0] ? formatReading(needSensors.value[0]) : 'Add temp/humidity/lux sensor',
        targetLabel: 'Uses setpoints when conditions type is setpoint',
        automationLabel: r.is_active ? 'Active' : 'Inactive',
        controlLabel: fan ? `${fan.name}` : 'Link fan/vent/shade actuator',
        controlOnline: fan?.current_state_text === 'online',
        lastEventLabel: fan ? lastEventForActuator(fan.id) : '',
      })
    }
  }

  return cards
})
</script>
