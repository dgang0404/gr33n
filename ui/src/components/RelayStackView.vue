<template>
  <div v-if="stacks.length" class="rounded-lg border border-zinc-800 bg-zinc-900/60 p-3" data-test="relay-stack-view">
    <h3 class="text-xs font-semibold text-zinc-300 mb-2">Relay stack (from your wiring)</h3>
    <p class="text-[10px] text-zinc-600 mb-3">
      Cards needed: {{ stacks.length }}. Set each card's DIP switch before stacking.
    </p>
    <div class="space-y-3">
      <div
        v-for="stack in stacks"
        :key="'stack-' + stack.level"
        class="rounded-lg border border-zinc-800 bg-zinc-950/50 p-3"
        :data-test="`relay-stack-${stack.level}`"
      >
        <div class="flex flex-wrap items-center gap-2 mb-2">
          <span class="text-xs font-semibold text-white">Stack {{ stack.level }}</span>
          <span class="text-[10px] font-mono text-zinc-500">I²C {{ stack.i2c }}</span>
          <span class="text-[10px] text-zinc-400">{{ stack.channelRange }}</span>
        </div>
        <p class="text-[10px] text-amber-300/80 mb-2">DIP: {{ stack.dipLabel }}</p>
        <div class="grid grid-cols-4 sm:grid-cols-8 gap-1">
          <button
            v-for="slot in stack.slots"
            :key="'slot-' + slot.channel"
            type="button"
            class="rounded border px-1 py-1.5 text-[10px] text-left transition-colors"
            :class="slotClass(slot)"
            :data-test="`relay-slot-ch-${slot.channel}`"
            @click="$emit('select-channel', slot)"
          >
            <span class="font-mono text-gr33n-400 block">ch{{ slot.channel }}</span>
            <span class="text-zinc-500">R{{ slot.relay }}</span>
            <span v-if="slot.assigned" class="text-zinc-300 truncate block">{{ slot.assigned.name }}</span>
            <span v-else class="text-zinc-600">free</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { buildRelayStacks } from '../lib/relayStack.js'

const props = defineProps({
  relayChannels: { type: Array, default: () => [] },
  conflictChannels: { type: Set, default: () => new Set() },
})

defineEmits(['select-channel'])

const stacks = computed(() => buildRelayStacks(props.relayChannels))

function slotClass(slot) {
  if (props.conflictChannels?.has(slot.channel)) {
    return 'border-red-800/80 bg-red-950/40'
  }
  if (slot.assigned) {
    return 'border-green-800/60 bg-green-950/20 hover:border-green-600'
  }
  return 'border-zinc-800 bg-zinc-900/40 hover:border-zinc-600'
}
</script>
