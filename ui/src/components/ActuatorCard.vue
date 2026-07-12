<template>
  <div class="space-y-2" data-test="actuator-card-wrap">
    <div class="card flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <span class="text-2xl">{{ icon }}</span>
        <div class="min-w-0">
          <div class="text-sm font-semibold text-white truncate">{{ device.name }}</div>
          <div class="text-xs text-gray-500">
            {{ device.device_type }} · Zone {{ device.zone_id }}
            <span v-if="deviceStatusLabel" class="ml-1" :class="deviceStatusClass">{{ deviceStatusLabel }}</span>
          </div>
          <div v-if="telemetryLine" class="text-[10px] text-zinc-500 mt-0.5">{{ telemetryLine }}</div>
          <span
            v-if="syncBadge"
            v-nav-hint="syncBadgeNavHint"
            class="inline-block cursor-default"
            :title="syncBadgeNavHint ? 'See Pi + HAT setup in sidebar' : undefined"
          >
            <span
              class="text-[10px] mt-0.5 font-medium"
              :class="syncBadgeClass"
              data-test="device-config-sync-badge"
            >
              {{ syncBadge.label }}
            </span>
          </span>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button
          type="button"
          class="text-[10px] uppercase tracking-wide text-zinc-500 hover:text-zinc-300 px-2 py-1 rounded border border-zinc-800"
          data-test="device-key-toggle"
          @click="showKeys = !showKeys"
        >
          {{ showKeys ? 'Hide key' : 'API key' }}
        </button>
      </div>
    </div>

    <div v-if="deviceActuators.length" class="space-y-2 pl-1">
      <div
        v-for="actuator in deviceActuators"
        :key="actuator.id"
        class="rounded-lg border border-zinc-800 bg-zinc-950/50 px-3 py-2"
        data-test="device-actuator-row"
      >
        <div class="flex items-center justify-between gap-2">
          <div class="min-w-0">
            <div class="text-xs font-medium text-white truncate">{{ actuator.name }}</div>
            <div class="text-[10px] text-zinc-500 capitalize">{{ actuator.actuator_type }}</div>
          </div>
          <div class="flex items-center gap-1 shrink-0">
            <button
              type="button"
              class="px-3 py-1 rounded-full text-xs font-semibold bg-gr33n-600 hover:bg-gr33n-700 text-white disabled:opacity-40"
              :disabled="busyId === actuator.id"
              data-test="actuator-on"
              @click="sendCommand(actuator, 'on')"
            >
              ON
            </button>
            <button
              type="button"
              class="px-3 py-1 rounded-full text-xs font-semibold bg-gray-800 hover:bg-gray-700 text-gray-300 disabled:opacity-40"
              :disabled="busyId === actuator.id"
              data-test="actuator-off"
              @click="sendCommand(actuator, 'off')"
            >
              OFF
            </button>
          </div>
        </div>
        <p v-if="queueHint[actuator.id]" class="text-[10px] text-amber-400 mt-1">{{ queueHint[actuator.id] }}</p>
        <ActuatorPulseControl :actuator="actuator" />
      </div>
    </div>
    <p v-else class="text-[10px] text-zinc-600 px-1">No actuators bound to this device.</p>

    <DeviceCommandQueue :device-id="device.id" />

    <DeviceApiKeyPanel v-if="showKeys" :device-id="device.id" />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmStore } from '../stores/farm'
import { configSyncBadge } from '../lib/deviceConfigSync'
import { useActuatorCommands } from '../composables/useActuatorCommands'
import DeviceApiKeyPanel from './DeviceApiKeyPanel.vue'
import DeviceCommandQueue from './DeviceCommandQueue.vue'
import ActuatorPulseControl from './ActuatorPulseControl.vue'

const props = defineProps({ device: Object })
const store = useFarmStore()
const { busyId, feedback, sendCommand: queueCommand } = useActuatorCommands()
const showKeys = ref(false)
const queueHint = ref({})

const ICONS = { light: '💡', irrigation: '💧', fan: '🌀', pump: '⚙️', heater: '🔥' }
const icon = computed(() => ICONS[props.device?.device_type] ?? '⚡')
const syncBadge = computed(() => configSyncBadge(props.device))
const deviceActuators = computed(() => store.actuatorsByDevice(props.device?.id))

const deviceConfig = computed(() => {
  const c = props.device?.config
  if (!c) return {}
  if (typeof c === 'object') return c
  return {}
})

const telemetryLine = computed(() => {
  const parts = []
  const ver = deviceConfig.value.client_version || props.device?.firmware_version
  if (ver) parts.push(`client ${ver}`)
  if (props.device?.last_heartbeat) {
    const ageMin = Math.round((Date.now() - new Date(props.device.last_heartbeat).getTime()) / 60000)
    parts.push(`heartbeat ${ageMin}m ago`)
  }
  return parts.join(' · ')
})

const deviceStatusLabel = computed(() => {
  const s = props.device?.status
  if (!s) return ''
  return s === 'online' ? '● online' : `● ${s}`
})

const deviceStatusClass = computed(() => {
  const s = props.device?.status
  if (s === 'online') return 'text-emerald-500'
  if (s === 'offline') return 'text-zinc-500'
  return 'text-amber-400'
})

const syncBadgeNavHint = computed(() => {
  const tone = syncBadge.value?.tone
  return tone === 'warn' || tone === 'muted' ? '/pi-setup' : null
})
const syncBadgeClass = computed(() => {
  const tone = syncBadge.value?.tone
  if (tone === 'ok') return 'text-emerald-500/90'
  if (tone === 'warn') return 'text-amber-400'
  return 'text-gray-500'
})

async function sendCommand(actuator, command) {
  const ok = await queueCommand(actuator, command, `Dashboard: ${command}`)
  queueHint.value = {
    ...queueHint.value,
    [actuator.id]: ok
      ? feedback.value
      : feedback.value,
  }
}
</script>
