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
            class="text-xs text-zinc-500 hover:text-green-400"
          >
            {{ link.label }} →
          </router-link>
        </div>
      </div>
      <ZoneConnectionPipeline :need="need" :devices="zoneDevices" />
    </div>

    <!-- Live sensors -->
    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h3 class="text-sm font-semibold text-white mb-3">Sensors</h3>
      <EmptyStateHint
        v-if="!needSensors.length"
        reason="no_telemetry"
        :message="needSensorsEmptyMessage"
        :action-label="otherSensorsInZone ? 'GPIO & wiring below' : undefined"
        :action-to="otherSensorsInZone ? zoneHardwareRoute(zoneId) : undefined"
        compact
      />
      <div v-else class="space-y-3">
        <div
          v-for="s in needSensors"
          :key="s.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
          :data-test="`zone-sensor-${s.id}`"
        >
          <div class="flex items-start justify-between gap-2 flex-wrap mb-1">
            <SensorTile :sensor="s" :reading="store.readings[s.id]" class="flex-1 min-w-0" />
          </div>
          <HardwareWiringBadge
            :entity="s"
            show-empty
            :hint-path="zoneHintPath"
            class="mt-1"
          />
          <button
            type="button"
            class="mt-2 text-[10px] text-zinc-500 hover:text-zinc-300"
            :data-test="`zone-sensor-wiring-toggle-${s.id}`"
            @click="toggleSensorWiring(s.id)"
          >
            {{ sensorWiringOpen[s.id] ? '▾ Hide wiring' : '▸ Edit wiring' }}
          </button>
          <HardwareWiringPanel
            v-if="sensorWiringOpen[s.id]"
            :sensor-id="s.id"
            :sensor="s"
            :devices="store.devices"
            :sensors="sensors"
            :actuators="actuators"
            class="mt-2 border-0 bg-transparent p-0"
            data-test="zone-sensor-wiring-panel"
            @updated="onHardwareUpdated"
          />
        </div>
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
      @schedules-updated="$emit('schedules-updated')"
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
      :active-crop-cycle="activeCropCycle"
      :grow-fit-context="growFitContext"
      @refreshed="$emit('water-refreshed', $event)"
      @plan-updated="$emit('plan-updated')"
    />

    <!-- Lighting inline editor (light only) -->
    <div v-if="need === PLANT_NEEDS.light" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-4">
      <h3 class="text-sm font-semibold text-white">Lighting program</h3>
      <ZoneCropStageTargetHint :zone-id="zoneId" :farm-id="farmId" />
      <ZoneLightingEditor
        :zone-id="zoneId"
        :farm-id="farmId"
        @updated="$emit('lighting-updated')"
      />
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
              <p class="text-zinc-500 text-xs capitalize">{{ a.actuator_type }} · {{ a.current_state_text || 'offline' }}</p>
              <HardwareWiringBadge
                :entity="a"
                show-empty
                :hint-path="zoneHintPath"
                class="mt-1"
              />
            </div>
            <div v-if="isOpenCloseActuator(a.actuator_type)" class="flex gap-1 shrink-0">
              <button
                v-for="cmd in ['open', 'close']"
                :key="cmd"
                type="button"
                class="px-3 py-1.5 rounded-lg text-xs font-medium capitalize disabled:opacity-40"
                :class="cmd === 'open' ? 'bg-gr33n-700 hover:bg-gr33n-600 text-white' : 'bg-zinc-800 hover:bg-zinc-700 text-zinc-300'"
                :disabled="commandBusy[a.id]"
                :data-test="`actuator-${cmd}-${a.id}`"
                @click="queueActuatorCommand(a, cmd)"
              >
                {{ cmd }}
              </button>
            </div>
            <div v-else-if="isDispenseActuator(a.actuator_type)" class="shrink-0">
              <button
                type="button"
                class="px-3 py-1.5 rounded-lg text-xs font-medium bg-gr33n-700 hover:bg-gr33n-600 text-white disabled:opacity-40"
                :disabled="commandBusy[a.id]"
                data-test="actuator-dispense"
                @click="queueActuatorCommand(a, 'dispense')"
              >
                Dispense
              </button>
            </div>
            <button
              v-else
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
          <p v-if="commandFeedback[a.id]" class="text-[10px] text-green-400 mt-1">{{ commandFeedback[a.id] }}</p>
          <ActuatorPulseControl v-if="!isOpenCloseActuator(a.actuator_type)" :actuator="a" />
          <button
            type="button"
            class="mt-2 text-[10px] text-zinc-500 hover:text-zinc-300"
            :data-test="`zone-actuator-wiring-toggle-${a.id}`"
            @click="toggleActuatorWiring(a.id)"
          >
            {{ actuatorWiringOpen[a.id] ? '▾ Hide wiring' : '▸ Edit wiring' }}
          </button>
          <ActuatorWiringPanel
            v-if="actuatorWiringOpen[a.id]"
            :actuator="a"
            :devices="store.devices"
            class="mt-2"
            data-test="zone-actuator-wiring-panel"
            @updated="onHardwareUpdated"
          />
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
import { computed, reactive } from 'vue'
import {
  PLANT_NEEDS,
  NEED_META,
  sensorPlantNeed,
  actuatorPlantNeed,
} from '../lib/plantNeeds.js'
import { isDispenseActuator, isOpenCloseActuator } from '../lib/actuatorControls.js'
import { useActuatorCommands } from '../composables/useActuatorCommands.js'
import { useFarmStore } from '../stores/farm.js'
import SensorTile from './SensorTile.vue'
import ZoneNeedConnectionCard from './ZoneNeedConnectionCard.vue'
import ActuatorPulseControl from './ActuatorPulseControl.vue'
import ActuatorWiringPanel from './ActuatorWiringPanel.vue'
import HardwareWiringPanel from './HardwareWiringPanel.vue'
import HardwareWiringBadge from './HardwareWiringBadge.vue'
import ZoneGreenhouseTab from './ZoneGreenhouseTab.vue'
import ZoneComfortTargets from './ZoneComfortTargets.vue'
import ZoneAutomationPanel from './ZoneAutomationPanel.vue'
import ZoneWaterGrowStory from './ZoneWaterGrowStory.vue'
import ZoneCropStageTargetHint from './ZoneCropStageTargetHint.vue'
import ZoneLightingEditor from './ZoneLightingEditor.vue'
import EmptyStateHint from './EmptyStateHint.vue'
import ZoneConnectionPipeline from './ZoneConnectionPipeline.vue'
import { scheduleRunsLabel } from '../lib/cronHumanize.js'
import { formatEntityHardwareLabel } from '../lib/hardwareWiring.js'
import {
  comfortTabRoute,
  zoneHardwareRoute,
  zoneWaterPlanRoute,
  zonesWorkspaceTabRoute,
} from '../lib/workspaceRoutes.js'

const hardwareFleetLink = (fleet) => ({
  to: zonesWorkspaceTabRoute('fleet', { fleet }),
  label: 'Farm hardware',
})

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
  activeCropCycle: { type: Object, default: null },
  growFitContext: { type: Object, default: () => ({ cropKey: '', stage: '' }) },
  actuatorEvents: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
  toggling: { type: Object, default: () => ({}) },
})

const emit = defineEmits([
  'toggle-actuator',
  'refresh-events',
  'setpoints-updated',
  'rules-updated',
  'schedules-updated',
  'water-refreshed',
  'plan-updated',
  'lighting-updated',
  'hardware-updated',
])

const store = useFarmStore()
const { sendCommand } = useActuatorCommands()
const sensorWiringOpen = reactive({})
const actuatorWiringOpen = reactive({})
const commandBusy = reactive({})
const commandFeedback = reactive({})

function toggleSensorWiring(id) {
  sensorWiringOpen[id] = !sensorWiringOpen[id]
}

function toggleActuatorWiring(id) {
  actuatorWiringOpen[id] = !actuatorWiringOpen[id]
}

async function queueActuatorCommand(actuator, command) {
  commandBusy[actuator.id] = true
  commandFeedback[actuator.id] = ''
  try {
    const ok = await sendCommand(actuator, command, `${props.zone?.name || 'Zone'}: ${command}`)
    if (ok) commandFeedback[actuator.id] = `Queued ${command}`
  } finally {
    commandBusy[actuator.id] = false
  }
}

function onHardwareUpdated() {
  emit('hardware-updated')
}

const zoneDevices = computed(() => store.devicesByZone(props.zoneId))

const farmTimezone = computed(() => store.farm?.timezone || 'America/New_York')

const meta = computed(() => NEED_META[props.need] || NEED_META[PLANT_NEEDS.air])

const zoneHintPath = computed(() => `/zones/${props.zoneId}`)

const otherSensorsInZone = computed(() =>
  props.sensors.some((s) => sensorPlantNeed(s.sensor_type) !== props.need),
)

const needSensorsEmptyMessage = computed(() => {
  const label = meta.value.shortLabel.toLowerCase()
  if (otherSensorsInZone.value) {
    return `No ${label} sensors on this tab — other sensors in this zone are listed under Sensors & controls below.`
  }
  return `No ${label} sensors in this zone yet.`
})

const sectionManageLinks = computed(() => {
  if (props.need === PLANT_NEEDS.water) {
    return [hardwareFleetLink('controls')]
  }
  if (props.need === PLANT_NEEDS.light) {
    return [hardwareFleetLink('lighting')]
  }
  if (props.need === PLANT_NEEDS.air) {
    return [
      ...(meta.value.manageLinks || []),
      hardwareFleetLink('sensors'),
    ]
  }
  return [
    ...(meta.value.manageLinks || []),
    hardwareFleetLink(props.need === PLANT_NEEDS.water ? 'controls' : 'sensors'),
  ].filter((link, i, arr) => arr.findIndex((x) => JSON.stringify(x.to) === JSON.stringify(link.to)) === i)
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
      manageTo: zoneWaterPlanRoute(props.zoneId),
      readingLabel: needSensors.value.length ? formatReading(needSensors.value[0]) : 'Add EC/pH sensor',
      targetLabel: props.activeProgram.run_duration_seconds
        ? `Pump on ${props.activeProgram.run_duration_seconds}s per run`
        : 'Volume-based feed',
      automationLabel: sched
        ? `${scheduleRunsLabel(sched)}${sched.is_active ? '' : ' · paused'}`
        : 'No feed timing linked',
      controlLabel: pump ? `${pump.name} — ${pump.current_state_text || 'offline'}` : 'Assign a pump actuator',
      controlHardwareLabel: pump ? formatEntityHardwareLabel(pump) : '',
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
        manageTo: zonesWorkspaceTabRoute('fleet', { fleet: 'lighting' }),
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
        manageTo: comfortTabRoute('automations'),
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
