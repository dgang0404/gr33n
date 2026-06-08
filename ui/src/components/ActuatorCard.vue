<template>
  <div class="space-y-2" data-test="actuator-card-wrap">
    <div class="card flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <span class="text-2xl">{{ icon }}</span>
        <div class="min-w-0">
          <div class="text-sm font-semibold text-white truncate">{{ device.name }}</div>
          <div class="text-xs text-gray-500">{{ device.device_type }} · Zone {{ device.zone_id }}</div>
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
        <button @click="toggle"
          :class="isOn
            ? 'bg-gr33n-600 hover:bg-gr33n-700 text-white'
            : 'bg-gray-800 hover:bg-gray-700 text-gray-400'"
          class="px-4 py-1.5 rounded-full text-sm font-semibold transition-colors">
          {{ isOn ? 'ON' : 'OFF' }}
        </button>
      </div>
    </div>
    <DeviceApiKeyPanel v-if="showKeys" :device-id="device.id" />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmStore } from '../stores/farm'
import { configSyncBadge } from '../lib/deviceConfigSync'
import DeviceApiKeyPanel from './DeviceApiKeyPanel.vue'

const props = defineProps({ device: Object })
const store = useFarmStore()
const showKeys = ref(false)

const ICONS = { light: '💡', irrigation: '💧', fan: '🌀', pump: '⚙️', heater: '🔥' }
const icon  = computed(() => ICONS[props.device?.device_type] ?? '⚡')
const isOn  = computed(() => props.device?.status === 'online')
const syncBadge = computed(() => configSyncBadge(props.device))
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

function toggle() {
  store.toggleDevice(props.device.id, props.device.status)
}
</script>
