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
            v-for="link in sectionManageLinks"
            :key="link.label"
            v-nav-hint="link.to"
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
        <strong class="text-zinc-400">automation or feed timing</strong> →
        <strong class="text-zinc-400">pump/light/fan</strong> →
        <strong class="text-zinc-400">device</strong>
      </p>
    </div>

    <!-- Live sensors -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-white mb-3">Sensors</h3>
      <EmptyStateHint
        v-if="!needSensors.length"
        reason="no_telemetry"
        :message="`No ${meta.shortLabel.toLowerCase()} sensors in this zone yet.`"
        compact
      />
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <SensorTile
          v-for="s in needSensors"
          :key="s.id"
          :sensor="s"
          :reading="store.readings[s.id]"
        />
      </div>
    </div>

    <ZoneAutomationPanel
      :need="need"
      :zone-id="zoneId"
      :zone-name="zone?.name || ''"
      :sensors="sensors"
      :rules="rules"
      :schedules="schedules"
      :active-program="activeProgram"
      :lighting-programs="lightingPrograms"
      @rules-updated="$emit('rules-updated')"
    />

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

    <ZoneWaterGrowStory
      v-if="need === PLANT_NEEDS.water"
      :zone-id="zoneId"
      :farm-id="farmId"
      :active-program="activeProgram"
      :programs="programs"
      :schedules="schedules"
      :fertigation-events="fertigationEvents"
      :actuators="actuators"
      :ec-targets="ecTargets"
      :reservoirs="reservoirs"
      :zone-name="zone?.name || 'This zone'"
      :farm-timezone="farmTimezone"
      @refreshed="$emit('water-refreshed', $event)"
      @plan-updated="$emit('plan-updated')"
    />

    <!-- Lighting block (light only) -->
    <div v-if="need === PLANT_NEEDS.light" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h3 class="text-sm font-semibold text-white">Lighting program</h3>
        <router-link
          :to="{ path: '/lighting', query: { zone_id: String(zoneId) } }"
          class="text-xs text-green-600 hover:text-green-400"
        >Lighting →</router-link>
      </div>
      <EmptyStateHint
        v-if="!zoneLightingPrograms.length"
        reason="no_data"
        message="No lighting program linked to this zone."
        action-label="Lighting programs"
        :action-to="{ path: '/lighting', query: { zone_id: String(zoneId) } }"
        compact
      />
      <ul v-else class="space-y-2">
        <li
          v-for="lp in zoneLightingPrograms"
          :key="lp.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 text-sm text-zinc-300"
        >
          {{ lp.name }}
          <span class="text-zinc-500 text-xs block mt-1">
            {{ lightingSummary(lp) }}
          </span>
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

    <ZoneComfortTargets
      :need="need"
      :zone-id="zoneId"
      :farm-id="farmId"
      :sensors="sensors"
      :setpoints="setpoints"
      @updated="$emit('setpoints-updated')"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
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
import ZoneComfortTargets from './ZoneComfortTargets.vue'
import ZoneAutomationPanel from './ZoneAutomationPanel.vue'
import ZoneWaterGrowStory from './ZoneWaterGrowStory.vue'
import EmptyStateHint from './EmptyStateHint.vue'
import { formatLightingProgramSummary } from '../lib/lightingDisplay.js'
import { scheduleRunsLabel } from '../lib/cronHumanize.js'

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
  ecTargets: { type: Array, default: () => [] },
  reservoirs: { type: Array, default: () => [] },
  activeProgram: { type: Object, default: null },
  actuatorEvents: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
  toggling: { type: Object, default: () => ({}) },
})

defineEmits(['toggle-actuator', 'refresh-events', 'setpoints-updated', 'rules-updated', 'water-refreshed', 'plan-updated'])

const store = useFarmStore()

const farmTimezone = computed(() => store.farm?.timezone || 'America/New_York')

const meta = computed(() => NEED_META[props.need] || NEED_META[PLANT_NEEDS.air])

const sectionManageLinks = computed(() => {
  if (props.need === PLANT_NEEDS.water) {
    return [
      {
        to: { path: '/feeding', query: { zone_id: String(props.zoneId) } },
        label: 'Feed & water hub',
      },
      {
        to: { path: '/fertigation', query: { tab: 'programs', zone_id: String(props.zoneId) } },
        label: 'Advanced feeding',
      },
    ]
  }
  if (props.need === PLANT_NEEDS.light) {
    return [{
      to: { path: '/lighting', query: { zone_id: String(props.zoneId) } },
      label: 'Lighting programs',
    }]
  }
  return meta.value.manageLinks
})

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

function lightingSummary(lp) {
  return formatLightingProgramSummary(lp)
}

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
      targetLabel: sp ? formatSetpoint(sp) : 'Set comfort target below',
      automationLabel: 'Alerts when readings leave the comfort band',
      controlLabel: '—',
      lastEventLabel: '',
    })
  }

  if (props.need === PLANT_NEEDS.water && props.activeProgram) {
    const sched = props.schedules.find(s => s.id === props.activeProgram.schedule_id)
    const pump = needActuators.value[0]
    cards.unshift({
      title: props.activeProgram.name,
      subtitle: 'Feeding plan',
      manageTo: { path: '/feeding', query: { zone_id: String(props.zoneId) } },
      readingLabel: needSensors.value.length ? formatReading(needSensors.value[0]) : 'Add EC/pH sensor',
      targetLabel: props.activeProgram.run_duration_seconds
        ? `Pump on ${props.activeProgram.run_duration_seconds}s per run`
        : 'Volume-based feed',
      automationLabel: sched
        ? `${scheduleRunsLabel(sched)}${sched.is_active ? '' : ' · paused'}`
        : 'No feed timing linked',
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
        subtitle: 'Automation',
        manageTo: '/automation',
        readingLabel: needSensors.value[0] ? formatReading(needSensors.value[0]) : 'Add temp/humidity/lux sensor',
        targetLabel: 'Uses comfort targets when configured',
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
