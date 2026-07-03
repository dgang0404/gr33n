<template>
  <component
    :is="clickable ? 'button' : 'div'"
    type="button"
    class="rounded border px-1.5 py-1 min-h-[2.75rem] flex flex-col gap-0.5 text-left w-full"
    :class="cellClass"
    :data-test="pin ? `virtual-pi-pin-${pin.physical}` : 'virtual-pi-pin-empty'"
    :title="tooltip"
    :disabled="!clickable"
    @click="clickable ? $emit('pin-click', pin) : undefined"
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
    <span v-else-if="clickable && pin?.role === 'gpio'" class="text-[9px] text-zinc-600">tap to wire</span>
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { pinRoleClass } from '../lib/piPinMap.js'
import { zoneHardwareRoute } from '../lib/workspaceRoutes.js'

const props = defineProps({
  pin: { type: Object, default: null },
  assignments: { type: Array, default: () => [] },
  zoneNames: { type: Object, default: () => ({}) },
  conflicted: { type: Boolean, default: false },
  clickable: { type: Boolean, default: false },
})

defineEmits(['pin-click'])

const cellClass = computed(() => {
  if (props.conflicted) {
    return 'bg-red-950/50 border-red-700 ring-1 ring-red-800/80 cursor-pointer'
  }
  if (!props.pin) return 'bg-zinc-900 border-zinc-800'
  if (props.assignments.length) {
    return `${props.clickable ? 'cursor-pointer hover:border-green-500 ' : ''}bg-green-950/40 border-green-800/70`
  }
  if (props.pin.role === 'gpio') {
    const base = props.clickable
      ? 'cursor-pointer hover:border-zinc-500 hover:bg-zinc-800/80 '
      : ''
    if (props.pin.buses?.length) {
      return `${base}${pinRoleClass(props.pin)} ring-1 ring-blue-900/40`
    }
    return `${base}${pinRoleClass(props.pin)}`
  }
  return pinRoleClass(props.pin)
})

const tooltip = computed(() => {
  if (!props.pin) return ''
  const parts = [`Physical ${props.pin.physical}`]
  if (props.pin.bcm != null) parts.push(`BCM ${props.pin.bcm}`)
  if (props.conflicted) parts.push('CONFLICT — double-booked')
  if (props.assignments.length) {
    parts.push(props.assignments.map((a) => a.name).join(', '))
  } else if (props.clickable && props.pin.role === 'gpio') {
    parts.push('Click to assign wiring')
  }
  return parts.join(' · ')
})
</script>
