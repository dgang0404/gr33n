<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 flex justify-end"
    data-test="pin-wiring-drawer"
  >
    <div class="absolute inset-0 bg-black/60" @click="$emit('close')" />
    <aside
      class="relative flex h-full w-full max-w-xl flex-col border-l border-zinc-800 bg-zinc-900 shadow-xl"
      data-test="pin-wiring-drawer-panel"
    >
      <div class="shrink-0 border-b border-zinc-800 px-4 py-3">
        <div class="flex items-start justify-between gap-2">
          <div>
            <h2 class="text-sm font-semibold text-white">Pin wiring</h2>
            <p v-if="pin" class="mt-1 text-xs text-zinc-500">
              Physical {{ pin.physical }}
              <span v-if="pin.bcm != null">· BCM {{ pin.bcm }}</span>
              · {{ pin.label }}
            </p>
          </div>
          <button type="button" class="text-lg leading-none text-zinc-500 hover:text-white" @click="$emit('close')">×</button>
        </div>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-4 py-4">
        <DriverHookupSteps v-if="activeDriver" :driver="activeDriver" />

        <template v-if="!selectedEntity">
          <p class="mb-2 text-xs text-zinc-400">Pick hardware to wire to this pin:</p>
          <div v-if="!unwired.unwiredSensors.length && !unwired.unwiredActuators.length" class="mb-4 text-xs text-zinc-500">
            No unwired sensors or actuators. Add hardware on a zone page first.
          </div>
          <ul v-else class="mb-4 space-y-1.5">
            <li v-for="s in unwired.unwiredSensors" :key="'s-' + s.id">
              <button
                type="button"
                class="w-full rounded-lg border border-zinc-800 px-3 py-2.5 text-left text-sm hover:border-green-700"
                @click="selectEntity('sensor', s)"
              >
                <span class="text-zinc-200">🌡 {{ s.name || s.sensor_type }}</span>
                <span class="block capitalize text-zinc-600">{{ s.sensor_type }}</span>
              </button>
            </li>
            <li v-for="a in unwired.unwiredActuators" :key="'a-' + a.id">
              <button
                type="button"
                class="w-full rounded-lg border border-zinc-800 px-3 py-2.5 text-left text-sm hover:border-green-700"
                @click="selectEntity('actuator', a)"
              >
                <span class="text-zinc-200">⚡ {{ a.name }}</span>
                <span class="block capitalize text-zinc-600">{{ a.actuator_type }}</span>
              </button>
            </li>
          </ul>

          <p v-if="existingOnPin.length" class="mb-2 text-xs text-zinc-400">Or edit what's already here:</p>
          <ul v-if="existingOnPin.length" class="space-y-1.5">
            <li v-for="a in existingOnPin" :key="a.kind + '-' + a.id">
              <button
                type="button"
                class="w-full rounded-lg border border-zinc-700 px-3 py-2.5 text-left text-sm text-zinc-200 hover:border-green-600"
                @click="selectExisting(a)"
              >
                Edit {{ a.name }} ({{ a.kind }})
              </button>
            </li>
          </ul>
        </template>

        <HardwareWiringPanel
          v-else-if="selectedEntity?.kind === 'sensor'"
          :sensor-id="selectedEntity.entity.id"
          :sensor="selectedEntity.entity"
          :devices="devices"
          :sensors="sensors"
          :actuators="actuators"
          auto-edit
          :preset-wiring="presetSensorWiring"
          class="border-0 bg-transparent p-0"
          @updated="onUpdated"
        />

        <ActuatorWiringPanel
          v-else-if="selectedEntity?.kind === 'actuator'"
          :actuator="selectedEntity.entity"
          :devices="devices"
          auto-edit
          :preset-gpio-pin="presetGpioPin"
          :preset-mode="presetActuatorMode"
          class="border-0 bg-transparent p-0"
          @updated="onUpdated"
        />

        <button
          v-if="selectedEntity"
          type="button"
          class="mt-4 text-xs text-zinc-500 hover:text-zinc-300"
          @click="selectedEntity = null"
        >
          ← Back to list
        </button>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import HardwareWiringPanel from './HardwareWiringPanel.vue'
import ActuatorWiringPanel from './ActuatorWiringPanel.vue'
import DriverHookupSteps from './DriverHookupSteps.vue'
import { listUnwiredEntities } from '../lib/wiringConflicts.js'
import { hookupStepsForDriver, wiringSourceForEntity, driverHookupsFromTaxonomy } from '../lib/driverHookups.js'
import { getDeviceTaxonomy } from '../lib/deviceTaxonomy.js'

const props = defineProps({
  open: { type: Boolean, default: false },
  pin: { type: Object, default: null },
  deviceId: { type: Number, required: true },
  assignments: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  devices: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'updated', 'hookup-change'])

const selectedEntity = ref(null)

const unwired = computed(() => listUnwiredEntities(props.sensors, props.actuators))
const existingOnPin = computed(() => props.assignments || [])

const presetGpioPin = computed(() => (props.pin?.bcm != null ? props.pin.bcm : null))

const presetActuatorMode = computed(() => (props.pin?.role === 'gpio' ? 'gpio_relay' : 'relay_hat'))

const presetSensorWiring = computed(() => ({
  device_id: props.deviceId,
  gpio_pin: props.pin?.bcm ?? null,
  source: 'dht22',
}))

const activeDriver = computed(() => {
  if (selectedEntity.value) {
    return wiringSourceForEntity(selectedEntity.value.kind, selectedEntity.value.entity)
  }
  if (props.assignments.length === 1) {
    const a = props.assignments[0]
    if (a.kind === 'sensor') {
      const s = props.sensors.find((x) => x.id === a.id)
      return wiringSourceForEntity('sensor', s)
    }
    const act = props.actuators.find((x) => x.id === a.id)
    return wiringSourceForEntity('actuator', act)
  }
  return ''
})

function emitHookupHighlight() {
  const driver = activeDriver.value
  if (!driver) {
    emit('hookup-change', { roles: [], bcmPin: null })
    return
  }
  const hookups = driverHookupsFromTaxonomy(getDeviceTaxonomy())
  const steps = hookupStepsForDriver(hookups, driver)
  const roles = steps.map((s) => s.role)
  emit('hookup-change', { roles, bcmPin: props.pin?.bcm ?? null })
}

watch(activeDriver, () => emitHookupHighlight(), { immediate: true })
watch(() => props.pin, () => emitHookupHighlight())

watch(() => props.open, (isOpen) => {
  if (!isOpen) {
    selectedEntity.value = null
    emit('hookup-change', { roles: [], bcmPin: null })
    return
  }
  if (!props.pin && props.assignments.length === 1) {
    selectExisting(props.assignments[0])
  }
})

function selectEntity(kind, entity) {
  selectedEntity.value = { kind, entity }
}

function selectExisting(a) {
  if (a.kind === 'sensor') {
    const entity = props.sensors.find((s) => s.id === a.id)
    if (entity) selectedEntity.value = { kind: 'sensor', entity }
  } else {
    const entity = props.actuators.find((x) => x.id === a.id)
    if (entity) selectedEntity.value = { kind: 'actuator', entity }
  }
}

function onUpdated() {
  emit('updated')
  emit('close')
}
</script>
