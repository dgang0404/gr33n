<template>
  <div
    class="rounded border px-1.5 py-1 min-h-[2.75rem] flex flex-col gap-0.5"
    :class="cellClass"
    :data-test="pin ? `virtual-pi-pin-${pin.physical}` : 'virtual-pi-pin-empty'"
    :title="tooltip"
  >
    <div class="flex items-center justify-between gap-1">
      <span class="text-[9px] font-mono text-zinc-500">{{ pin?.physical ?? '—' }}</span>
      <span v-if="pin?.bcm != null" class="text-[9px] font-mono text-zinc-600">BCM {{ pin.bcm }}</span>
    </div>
    <span class="text-[9px] text-zinc-400 truncate">{{ pin?.label || '—' }}</span>
    <template v-if="assignments.length">
      <router-link
        v-for="a in assignments.slice(0, 2)"
        :key="a.kind + '-' + a.id"
        v-nav-hint="'/zones'"
        :to="a.zoneId ? zoneHardwareRoute(a.zoneId) : '/zones'"
        class="text-[9px] text-green-400 hover:text-green-300 truncate"
        @click.stop
      >
        {{ a.name }}
      </router-link>
      <span v-if="assignments.length > 2" class="text-[9px] text-zinc-600">+{{ assignments.length - 2 }} more</span>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { pinRoleClass } from '../lib/piPinMap.js'
import { zoneHardwareRoute } from '../lib/workspaceRoutes.js'

const props = defineProps({
  pin: { type: Object, default: null },
  assignments: { type: Array, default: () => [] },
  zoneNames: { type: Object, default: () => ({}) },
})

const cellClass = computed(() => {
  if (!props.pin) return 'bg-zinc-900 border-zinc-800'
  if (props.assignments.length) {
    return 'bg-green-950/40 border-green-800/70'
  }
  if (props.pin.role === 'gpio' && props.pin.buses?.length) {
    return `${pinRoleClass(props.pin)} ring-1 ring-blue-900/40`
  }
  return pinRoleClass(props.pin)
})

const tooltip = computed(() => {
  if (!props.pin) return ''
  const parts = [`Physical ${props.pin.physical}`]
  if (props.pin.bcm != null) parts.push(`BCM ${props.pin.bcm}`)
  if (props.pin.buses?.length) parts.push(props.pin.buses.join(', '))
  if (props.assignments.length) {
    parts.push(props.assignments.map((a) => a.name).join(', '))
  }
  return parts.join(' · ')
})
</script>
