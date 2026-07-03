<template>
  <div class="space-y-4" data-test="virtual-pi-board">
    <div
      v-if="conflictReport.conflicts.length"
      class="rounded-lg border border-red-800/70 bg-red-950/30 px-3 py-2 text-xs text-red-200"
      data-test="virtual-pi-conflicts"
    >
      <p class="font-semibold mb-1">{{ conflictReport.conflicts.length }} wiring conflict(s)</p>
      <ul class="space-y-1 text-[11px] text-red-200/90">
        <li v-for="(c, i) in conflictReport.conflicts" :key="i">{{ c.message }}</li>
      </ul>
    </div>

    <div class="flex flex-wrap items-center gap-2 text-[10px] text-zinc-500">
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-red-900/80" /> Power</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-zinc-950 border border-zinc-600" /> GND</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-green-900/50 border border-green-700" /> Assigned</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-zinc-900 border border-zinc-700" /> Free GPIO</span>
      <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-sm bg-red-950 border border-red-700" /> Conflict</span>
    </div>

    <div class="rounded-xl border border-zinc-700 bg-zinc-950 p-3 max-w-md mx-auto">
      <p class="text-[10px] uppercase tracking-wide text-zinc-600 text-center mb-2">40-pin header · USB ↓ · tap GPIO to wire</p>
      <div class="grid grid-cols-2 gap-x-2 gap-y-1">
        <template v-for="row in gridRows" :key="'row-' + row.row">
          <PinCell
            :pin="row.left"
            :assignments="assignmentsForPin(row.left?.physical)"
            :zone-names="zoneNames"
            :conflicted="isConflictPin(row.left?.physical)"
            :highlighted="isHookupPin(row.left?.physical)"
            :clickable="isClickablePin(row.left)"
            @pin-click="onPinClick"
          />
          <PinCell
            :pin="row.right"
            :assignments="assignmentsForPin(row.right?.physical)"
            :zone-names="zoneNames"
            :conflicted="isConflictPin(row.right?.physical)"
            :highlighted="isHookupPin(row.right?.physical)"
            :clickable="isClickablePin(row.right)"
            @pin-click="onPinClick"
          />
        </template>
      </div>
    </div>

    <RelayStackView
      :relay-channels="relayChannels"
      :conflict-channels="conflictReport.conflictChannels"
      @select-channel="onRelayChannelClick"
    />

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

    <PinWiringDrawer
      :open="drawerOpen"
      :pin="drawerPin"
      :device-id="deviceId"
      :assignments="drawerAssignments"
      :sensors="sensors"
      :actuators="actuators"
      :devices="devices"
      @close="closeDrawer"
      @updated="$emit('updated')"
      @hookup-change="onHookupChange"
    />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import {
  assignmentsForDevice,
  headerGridRows,
  physicalPinsForHookupRoles,
} from '../lib/piPinMap.js'
import { collectDeviceWiringConflicts } from '../lib/wiringConflicts.js'
import { zoneHardwareRoute } from '../lib/workspaceRoutes.js'
import PinCell from './VirtualPiPinCell.vue'
import PinWiringDrawer from './PinWiringDrawer.vue'
import RelayStackView from './RelayStackView.vue'

const props = defineProps({
  deviceId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  zones: { type: Array, default: () => [] },
  devices: { type: Array, default: () => [] },
})

defineEmits(['updated'])

const gridRows = headerGridRows()
const drawerOpen = ref(false)
const drawerPin = ref(null)
const drawerAssignments = ref([])
const hookupHighlight = ref({ roles: [], bcmPin: null })

const hookupPhysicalPins = computed(() =>
  physicalPinsForHookupRoles(hookupHighlight.value.roles, hookupHighlight.value.bcmPin),
)

const assignmentBundle = computed(() =>
  assignmentsForDevice(props.deviceId, props.sensors, props.actuators),
)

const byPhysical = computed(() => assignmentBundle.value.byPhysical)
const i2cAttachments = computed(() => assignmentBundle.value.i2cAttachments)
const relayChannels = computed(() => assignmentBundle.value.relayChannels)
const uartAttachments = computed(() => assignmentBundle.value.uartAttachments)

const conflictReport = computed(() =>
  collectDeviceWiringConflicts(props.deviceId, props.sensors, props.actuators),
)

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

function isConflictPin(physical) {
  if (physical == null) return false
  return conflictReport.value.conflictPhysicalPins.has(physical)
}

function isHookupPin(physical) {
  if (physical == null) return false
  return hookupPhysicalPins.value.has(physical)
}

function onHookupChange(payload) {
  hookupHighlight.value = payload || { roles: [], bcmPin: null }
}

function isClickablePin(pin) {
  return pin?.role === 'gpio' && pin.bcm != null
}

function onPinClick(pin) {
  if (!pin || pin.role !== 'gpio') return
  drawerPin.value = pin
  drawerAssignments.value = assignmentsForPin(pin.physical)
  drawerOpen.value = true
}

function onRelayChannelClick(slot) {
  if (!slot.assigned) return
  const entity = props.actuators.find((a) => a.id === slot.assigned.id)
  if (!entity) return
  drawerPin.value = null
  drawerAssignments.value = [{ kind: 'actuator', id: entity.id, name: entity.name }]
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
  drawerPin.value = null
  drawerAssignments.value = []
  hookupHighlight.value = { roles: [], bcmPin: null }
}
</script>
