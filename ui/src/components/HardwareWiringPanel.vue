<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-semibold text-white inline-flex items-center">
        Hardware wiring
        <HelpTip position="top">
          Where this sensor is physically connected on the Pi — BCM GPIO pin, I2C channel, or serial port.
          The edge device uses this when reading values.
        </HelpTip>
      </h2>
      <button
        v-if="!editing"
        type="button"
        @click="beginEdit"
        class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 hover:text-white"
      >Edit</button>
    </div>

    <div v-if="!editing">
      <HardwareWiringBadge :wiring="displayWiring" show-empty />
      <p v-if="displayWiring?.notes" class="text-zinc-500 text-xs mt-2">{{ displayWiring.notes }}</p>
    </div>

    <form v-else class="space-y-3" @submit.prevent="save">
      <label class="block">
        <span class="text-xs text-zinc-400">Driver / source</span>
        <select
          v-model="form.source"
          required
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        >
          <option value="" disabled>Select driver…</option>
          <option v-for="opt in SENSOR_WIRING_SOURCES" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </label>

      <label v-if="needsGpio" class="block">
        <span class="text-xs text-zinc-400">BCM GPIO pin</span>
        <input
          v-model.number="form.gpio_pin"
          type="number"
          min="0"
          max="27"
          required
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        />
      </label>

      <label v-if="needsI2c" class="block">
        <span class="text-xs text-zinc-400">I2C channel (0–3)</span>
        <input
          v-model.number="form.i2c_channel"
          type="number"
          min="0"
          max="3"
          required
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        />
      </label>

      <label v-if="needsSerial" class="block">
        <span class="text-xs text-zinc-400">Serial port</span>
        <input
          v-model="form.serial_port"
          type="text"
          placeholder="/dev/ttyS0"
          required
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        />
      </label>

      <label class="block">
        <span class="text-xs text-zinc-400 inline-flex items-center gap-1">
          Edge device (Pi)
          <HelpTip position="top">
            Raspberry Pis registered for this farm (Settings or Connect edge device).
            Loaded from <code class="text-zinc-500">GET /farms/:id/devices</code> — name shown is the device record.
          </HelpTip>
        </span>
        <select
          v-model.number="form.device_id"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        >
          <option :value="null">(none)</option>
          <option v-for="d in devices" :key="d.id" :value="d.id">
            {{ d.name || d.device_uid || `Device ${d.id}` }}
          </option>
        </select>
        <p v-if="!devices.length" class="text-[10px] text-zinc-600 mt-1">
          No edge devices yet —
          <router-link
            v-if="deviceSetupRoute"
            :to="deviceSetupRoute"
            class="text-green-600 hover:text-green-400"
          >connect a Pi</router-link>.
        </p>
      </label>

      <label class="block">
        <span class="text-xs text-zinc-400">Notes (optional)</span>
        <input
          v-model="form.notes"
          type="text"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        />
      </label>

      <p v-if="conflictPreview" class="text-amber-400 text-xs">{{ conflictPreview.message }}</p>
      <p v-if="error" class="text-red-400 text-xs">{{ error }}</p>

      <div class="flex flex-wrap gap-2 pt-1">
        <button
          type="submit"
          :disabled="saving || !!conflictPreview"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-700 text-white hover:bg-green-600 disabled:opacity-40"
        >Save wiring</button>
        <button
          type="button"
          :disabled="saving"
          @click="clearWiring"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-400 hover:text-zinc-200 disabled:opacity-40"
        >Clear</button>
        <button
          type="button"
          :disabled="saving"
          @click="cancelEdit"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-400 hover:text-zinc-200"
        >Cancel</button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import HelpTip from './HelpTip.vue'
import HardwareWiringBadge from './HardwareWiringBadge.vue'
import { findWiringConflict, resolveWiring, SENSOR_WIRING_SOURCES } from '../lib/hardwareWiring.js'
import { useFarmContextStore } from '../stores/farmContext.js'

const props = defineProps({
  sensorId: { type: [String, Number], required: true },
  sensor: { type: Object, required: true },
  devices: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  autoEdit: { type: Boolean, default: false },
  presetWiring: { type: Object, default: null },
})

const emit = defineEmits(['updated'])

const farmContext = useFarmContextStore()
const deviceSetupRoute = computed(() => {
  const fid = farmContext.farmId
  return fid ? { path: `/farms/${fid}/devices/new` } : null
})

const editing = ref(false)
const saving = ref(false)
const error = ref('')
const form = ref(emptyForm())

const displayWiring = computed(() => resolveWiring(props.sensor))

const needsGpio = computed(() => ['dht22', 'gpio_digital'].includes(form.value.source))
const needsI2c = computed(() => form.value.source === 'ads1115')
const needsSerial = computed(() => form.value.source === 'mhz19')

const conflictPreview = computed(() => {
  if (!editing.value) return null
  return findWiringConflict({
    wiring: buildPayload(),
    entityType: 'sensor',
    entityId: props.sensorId,
    sensors: props.sensors,
    actuators: props.actuators,
  })
})

function emptyForm() {
  return {
    source: '',
    gpio_pin: null,
    i2c_channel: null,
    serial_port: '',
    device_id: null,
    notes: '',
  }
}

function beginEdit() {
  const w = displayWiring.value
  const preset = props.presetWiring || {}
  form.value = w
    ? {
        source: w.source || preset.source || '',
        gpio_pin: w.gpio_pin ?? preset.gpio_pin ?? null,
        i2c_channel: w.i2c_channel ?? preset.i2c_channel ?? null,
        serial_port: w.serial_port || preset.serial_port || '',
        device_id: w.device_id ?? preset.device_id ?? null,
        notes: w.notes || preset.notes || '',
      }
    : {
        ...emptyForm(),
        source: preset.source || '',
        gpio_pin: preset.gpio_pin ?? null,
        i2c_channel: preset.i2c_channel ?? null,
        serial_port: preset.serial_port || '',
        device_id: preset.device_id ?? null,
        notes: preset.notes || '',
      }
  error.value = ''
  editing.value = true
}

function cancelEdit() {
  editing.value = false
  error.value = ''
}

function buildPayload() {
  const f = form.value
  const wiring = {
    source: f.source,
    notes: f.notes || undefined,
  }
  if (f.device_id != null) wiring.device_id = f.device_id
  if (needsGpio.value && f.gpio_pin != null) wiring.gpio_pin = f.gpio_pin
  if (needsI2c.value && f.i2c_channel != null) wiring.i2c_channel = f.i2c_channel
  if (needsSerial.value && f.serial_port) wiring.serial_port = f.serial_port
  return wiring
}

async function save() {
  saving.value = true
  error.value = ''
  try {
    const r = await api.patch(`/sensors/${props.sensorId}/wiring`, { wiring: buildPayload() })
    emit('updated', r.data)
    editing.value = false
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to save wiring'
  } finally {
    saving.value = false
  }
}

async function clearWiring() {
  saving.value = true
  error.value = ''
  try {
    const r = await api.patch(`/sensors/${props.sensorId}/wiring`, { wiring: null })
    emit('updated', r.data)
    editing.value = false
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to clear wiring'
  } finally {
    saving.value = false
  }
}

watch(() => props.sensor, () => {
  if (!editing.value) error.value = ''
})

watch(
  () => [props.autoEdit, props.sensorId],
  ([auto]) => {
    if (auto) beginEdit()
  },
  { immediate: true },
)
</script>
