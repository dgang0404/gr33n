<template>
  <div class="space-y-4" data-test="virtual-pi-board">
    <div class="flex flex-wrap items-center gap-2 text-[10px] text-zinc-500">
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-red-900/80" /> Power</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-zinc-950 border border-zinc-600" /> GND</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-green-900/50 border border-green-700" /> Assigned</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-zinc-900 border border-zinc-700" /> Free GPIO</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-zinc-800 border border-zinc-600" /> Reserved</span>
    </div>

    <div class="rounded-xl border border-zinc-700 bg-zinc-950 p-3 max-w-md mx-auto">
      <p class="text-[10px] uppercase tracking-wide text-zinc-600 text-center mb-2">40-pin header · USB ↓</p>
      <div class="grid grid-cols-2 gap-x-2 gap-y-1">
        <template v-for="row in gridRows" :key="'row-' + row.row">
          <PinCell
            :pin="row.left"
            :assignments="assignmentsForPin(row.left?.physical)"
            :zone-names="zoneNames"
          />
          <PinCell
            :pin="row.right"
            :assignments="assignmentsForPin(row.right?.physical)"
            :zone-names="zoneNames"
          />
        </template>
      </div>
    </div>

    <div v-if="i2cAttachments.length" class="rounded-lg border border-zinc-800 bg-zinc-900/60 p-3">
      <h3 class="text-xs font-semibold text-zinc-300 mb-2">I²C bus (pins 3 &amp; 5)</h3>
      <ul class="space-y-1 text-xs">
        <li
          v-for="item in i2cAttachments"
          :key="'i2c-' + item.kind + '-' + item.id"
          class="flex flex-wrap items-center gap-2"
        >
          <span class="text-zinc-200">{{ item.label || item.name }}</span>
          <router-link
            v-if="item.zoneId && item.kind !== 'hat'"
            v-nav-hint="'/zones'"
            :to="zoneHardwareRoute(item.zoneId)"
            class="text-green-500 hover:text-green-400 text-[10px]"
          >
            {{ zoneLabel(item.zoneId) }} →
          </router-link>
        </li>
      </ul>
      <div v-if="relayChannels.length" class="mt-3 pt-3 border-t border-zinc-800">
        <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Relay channels</p>
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-1">
          <div
            v-for="r in relayChannels"
            :key="'relay-' + r.id"
            class="rounded bg-zinc-950 border border-zinc-800 px-2 py-1 text-[10px]"
          >
            <span class="font-mono text-gr33n-400">{{ r.label }}</span>
            <span class="text-zinc-400 truncate block">{{ r.name }}</span>
            <router-link
              v-if="r.zoneId"
              v-nav-hint="'/zones'"
              :to="zoneHardwareRoute(r.zoneId)"
              class="text-green-600 hover:text-green-400"
            >
              {{ zoneLabel(r.zoneId) }}
            </router-link>
          </div>
        </div>
      </div>
    </div>

    <div v-if="uartAttachments.length" class="rounded-lg border border-zinc-800 bg-zinc-900/60 p-3">
      <h3 class="text-xs font-semibold text-zinc-300 mb-2">Serial / UART (pins 8 &amp; 10)</h3>
      <ul class="space-y-1 text-xs text-zinc-300">
        <li v-for="u in uartAttachments" :key="'uart-' + u.id">
          {{ u.name }} · {{ u.label }}
          <router-link
            v-if="u.zoneId"
            v-nav-hint="'/zones'"
            :to="zoneHardwareRoute(u.zoneId)"
            class="text-green-500 hover:text-green-400 ml-1 text-[10px]"
          >
            {{ zoneLabel(u.zoneId) }} →
          </router-link>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  assignmentsForDevice,
  headerGridRows,
} from '../lib/piPinMap.js'
import { zoneHardwareRoute } from '../lib/workspaceRoutes.js'
import PinCell from './VirtualPiPinCell.vue'

const props = defineProps({
  deviceId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  zones: { type: Array, default: () => [] },
})

const gridRows = headerGridRows()

const assignmentBundle = computed(() =>
  assignmentsForDevice(props.deviceId, props.sensors, props.actuators),
)

const byPhysical = computed(() => assignmentBundle.value.byPhysical)
const i2cAttachments = computed(() => assignmentBundle.value.i2cAttachments)
const relayChannels = computed(() => assignmentBundle.value.relayChannels)
const uartAttachments = computed(() => assignmentBundle.value.uartAttachments)

const zoneNames = computed(() => {
  const map = {}
  for (const z of props.zones) map[z.id] = z.name
  return map
})

function assignmentsForPin(physical) {
  if (physical == null) return []
  return byPhysical.value.get(physical) || []
}

function zoneLabel(zoneId) {
  return zoneNames.value[zoneId] || `Zone ${zoneId}`
}
</script>
