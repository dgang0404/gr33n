<template>
  <section
    class="bg-zinc-900 border border-green-900/50 rounded-xl p-4 space-y-3"
    data-test="first-run-checklist"
  >
    <div class="flex items-start justify-between gap-3">
      <div>
        <h2 class="text-sm font-semibold text-green-400">Getting started</h2>
        <p class="text-xs text-zinc-500 mt-0.5">
          Finish these steps to get readings, feeding, and automations online.
        </p>
      </div>
      <button
        type="button"
        class="text-[10px] text-zinc-500 hover:text-zinc-300 shrink-0 min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 inline-flex items-center justify-center"
        aria-label="Hide getting started checklist"
        data-test="first-run-dismiss"
        @click="onDismiss"
      >
        Hide for now
      </button>
    </div>

    <ul class="space-y-2">
      <li
        v-for="item in items"
        :key="item.id"
        class="flex items-center gap-3 rounded-lg border px-3 py-2 text-sm"
        :class="item.done ? 'border-green-900/40 bg-green-950/20' : 'border-zinc-800 bg-zinc-950/50'"
        :data-test="`first-run-item-${item.id}`"
      >
        <span
          class="text-base leading-none shrink-0"
          :class="item.done ? 'text-green-500' : 'text-zinc-600'"
          aria-hidden="true"
        >
          {{ item.done ? '☑' : '☐' }}
        </span>
        <router-link
          v-if="!item.done"
          :to="item.to"
          class="text-zinc-200 hover:text-green-300 transition-colors min-h-[44px] sm:min-h-0 inline-flex items-center"
          :aria-label="`Getting started: ${item.label}`"
          :data-test="`first-run-link-${item.id}`"
        >
          {{ item.label }}
        </router-link>
        <span v-else class="text-zinc-400">{{ item.label }}</span>
        <span v-if="item.done" class="ml-auto text-[10px] text-green-600/80 uppercase tracking-wide">Done</span>
      </li>
    </ul>

    <div v-if="starters.length" class="pt-1 border-t border-zinc-800 space-y-1.5">
      <p class="text-[10px] uppercase tracking-widest text-zinc-500">Ask Guardian</p>
      <GuardianStarterChips :starters="starters" />
    </div>
  </section>
</template>

<script setup>
import GuardianStarterChips from './GuardianStarterChips.vue'
import { dismissFirstRunChecklist } from '../lib/firstRunChecklist.js'

const props = defineProps({
  items: { type: Array, default: () => [] },
  farmId: { type: Number, default: null },
  starters: { type: Array, default: () => [] },
})

const emit = defineEmits(['dismiss'])

function onDismiss() {
  if (props.farmId) dismissFirstRunChecklist(props.farmId)
  emit('dismiss')
}
</script>
