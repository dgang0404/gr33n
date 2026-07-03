<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="actuator-wiring-panel">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-white inline-flex items-center gap-1.5">
        Hardware wiring
        <HelpTip position="top">
          For Sequent relay HATs set a channel (0–63). For direct GPIO relays set a BCM pin. Both need the Pi device assigned.
        </HelpTip>
      </h3>
      <button
        v-if="!editing"
        type="button"
        class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 hover:text-white"
        data-test="actuator-wiring-edit"
        @click="beginEdit"
      >Edit</button>
    </div>

    <!-- Read view -->
    <div v-if="!editing" class="space-y-1.5">
      <HardwareWiringBadge :entity="actuator" show-empty />
      <div v-if="channelLabel" class="text-xs text-zinc-400">{{ channelLabel }}</div>
      <div v-if="deviceLabel" class="text-[11px] text-zinc-600">{{ deviceLabel }}</div>
    </div>

    <!-- Edit form -->
    <form v-else class="space-y-3" @submit.prevent="save">
      <!-- Mode toggle -->
      <div class="flex gap-2">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg border transition-colors"
          :class="mode === 'relay_hat'
            ? 'bg-green-900/40 border-green-700 text-green-300'
            : 'border-zinc-700 text-zinc-400 hover:text-zinc-200'"
          @click="mode = 'relay_hat'"
        >Sequent relay HAT (channel)</button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg border transition-colors"
          :class="mode === 'gpio_relay'
            ? 'bg-green-900/40 border-green-700 text-green-300'
            : 'border-zinc-700 text-zinc-400 hover:text-zinc-200'"
          @click="mode = 'gpio_relay'"
        >Direct GPIO relay (pin)</button>
      </div>

      <!-- Relay HAT channel -->
      <label v-if="mode === 'relay_hat'" class="block">
        <span class="text-xs text-zinc-400">Relay channel (0 = relay 1 on card 0, 8 = relay 1 on card 1, …)</span>
        <input
          v-model.number="form.channel"
          type="number"
          min="0"
          max="63"
          required
          placeholder="e.g. 0"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
          data-test="actuator-channel-input"
        />
        <p class="text-[11px] text-zinc-600 mt-1">Stack {{ Math.floor((form.channel ?? 0) / 8) }}, relay {{ ((form.channel ?? 0) % 8) + 1 }}</p>
      </label>

      <!-- Direct GPIO pin -->
      <label v-else class="block">
        <span class="text-xs text-zinc-400">BCM GPIO pin (0–27)</span>
        <input
          v-model.number="form.gpioPin"
          type="number"
          min="0"
          max="27"
          required
          placeholder="e.g. 17"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
          data-test="actuator-gpio-input"
        />
      </label>

      <!-- Pi device -->
      <label class="block">
        <span class="text-xs text-zinc-400">Edge device (Pi)</span>
        <select
          v-model.number="form.deviceId"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        >
          <option :value="null">(none — assign later)</option>
          <option v-for="d in devices" :key="d.id" :value="d.id">
            {{ d.name || d.device_uid || `Device ${d.id}` }}
          </option>
        </select>
      </label>

      <!-- Notes (only for gpio_relay, stored in config.wiring) -->
      <label v-if="mode === 'gpio_relay'" class="block">
        <span class="text-xs text-zinc-400">Notes (optional)</span>
        <input
          v-model="form.notes"
          type="text"
          placeholder="e.g. NC contact for fail-safe fan"
          class="mt-1 w-full rounded-lg bg-zinc-950 border border-zinc-700 px-3 py-2 text-sm text-white focus:outline-none focus:border-green-600"
        />
      </label>

      <p v-if="error" class="text-xs text-red-400">{{ error }}</p>

      <div class="flex flex-wrap gap-2 pt-1">
        <button
          type="submit"
          :disabled="saving"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-700 text-white hover:bg-green-600 disabled:opacity-40"
          data-test="actuator-wiring-save"
        >{{ saving ? 'Saving…' : 'Save wiring' }}</button>
        <button
          type="button"
          :disabled="saving"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-400 hover:text-zinc-200 disabled:opacity-40"
          @click="clearWiring"
        >Clear</button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-400 hover:text-zinc-200"
          @click="editing = false"
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
import { resolveWiring } from '../lib/hardwareWiring.js'

const props = defineProps({
  actuator: { type: Object, required: true },
  devices:  { type: Array, default: () => [] },
  autoEdit: { type: Boolean, default: false },
  presetGpioPin: { type: Number, default: null },
  presetMode: { type: String, default: '' },
})

const emit = defineEmits(['updated'])

const editing = ref(false)
const saving  = ref(false)
const error   = ref('')

// mode: 'relay_hat' (hardware_identifier = channel) | 'gpio_relay' (config.wiring.gpio_pin)
const mode = ref('relay_hat')

const form = ref({ channel: null, gpioPin: null, deviceId: null, notes: '' })

const currentWiring   = computed(() => resolveWiring(props.actuator))
const currentChannel  = computed(() => {
  const hi = props.actuator.hardware_identifier
  if (hi == null) return null
  const n = parseInt(String(hi), 10)
  return Number.isFinite(n) && n >= 0 ? n : null
})

const channelLabel = computed(() => {
  const ch = currentChannel.value
  if (ch == null) return ''
  return `Relay HAT — channel ${ch} (stack ${Math.floor(ch / 8)}, relay ${(ch % 8) + 1})`
})

const deviceLabel = computed(() => {
  const did = props.actuator.device_id
  if (!did) return ''
  const d = props.devices.find(d => d.id === did)
  return d ? `Pi: ${d.name || d.device_uid || `Device ${did}`}` : `Device ${did}`
})

function beginEdit() {
  const w = currentWiring.value
  if (props.presetMode === 'gpio_relay' || w?.gpio_pin != null) {
    mode.value = 'gpio_relay'
    form.value = {
      channel: null,
      gpioPin: w?.gpio_pin ?? props.presetGpioPin ?? null,
      deviceId: props.actuator.device_id ?? null,
      notes: w?.notes || '',
    }
  } else {
    mode.value = 'relay_hat'
    form.value = {
      channel: currentChannel.value,
      gpioPin: null,
      deviceId: props.actuator.device_id ?? null,
      notes: '',
    }
  }
  error.value = ''
  editing.value = true
}

async function save() {
  error.value = ''
  saving.value = true
  try {
    let updated
    if (mode.value === 'relay_hat') {
      const ch = form.value.channel
      if (ch == null || ch < 0 || ch > 63) {
        error.value = 'Channel must be 0–63'
        return
      }
      const r = await api.patch(`/actuators/${props.actuator.id}/assign`, {
        device_id:           form.value.deviceId,
        hardware_identifier: String(ch),
      })
      updated = r.data
    } else {
      const pin = form.value.gpioPin
      if (pin == null || pin < 0 || pin > 27) {
        error.value = 'GPIO pin must be 0–27'
        return
      }
      const [r1, r2] = await Promise.all([
        api.patch(`/actuators/${props.actuator.id}/assign`, {
          device_id:           form.value.deviceId,
          hardware_identifier: null,
        }),
        api.patch(`/actuators/${props.actuator.id}/wiring`, {
          wiring: { source: 'gpio_relay', gpio_pin: pin, device_id: form.value.deviceId, notes: form.value.notes || undefined },
        }),
      ])
      updated = r2.data ?? r1.data
    }
    emit('updated', updated)
    editing.value = false
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to save wiring'
  } finally {
    saving.value = false
  }
}

async function clearWiring() {
  error.value = ''
  saving.value = true
  try {
    const [r1] = await Promise.all([
      api.patch(`/actuators/${props.actuator.id}/assign`, {
        device_id: null,
        hardware_identifier: null,
        clear_device: true,
      }),
      api.patch(`/actuators/${props.actuator.id}/wiring`, { wiring: null }),
    ])
    emit('updated', r1.data)
    editing.value = false
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to clear wiring'
  } finally {
    saving.value = false
  }
}

watch(
  () => [props.autoEdit, props.actuator?.id],
  ([auto]) => {
    if (auto) beginEdit()
  },
  { immediate: true },
)
</script>
