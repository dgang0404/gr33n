<template>
  <article
    class="farm-zone-tile relative flex flex-col h-full min-h-0 rounded-xl border-2 bg-zinc-950/90 backdrop-blur-sm shadow-lg transition-shadow"
    :class="tileClasses"
    :aria-label="ariaLabel"
    data-test="farm-canvas-zone-tile"
  >
    <header class="flex items-start justify-between gap-1 px-2 pt-2 pb-1 shrink-0">
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-1">
          <span class="text-sm" :aria-hidden="true">{{ typeIcon }}</span>
          <HelpTip v-if="showZoneTip" position="bottom" class="shrink-0">
            A zone is your grow area — a bed, tent, room, greenhouse section, or outdoor plot.
            You assign plants, lights, sensors, and watering to it.
          </HelpTip>
          <h4 class="text-xs font-semibold text-white truncate">{{ zone.name }}</h4>
        </div>
        <p class="text-[10px] text-zinc-500 truncate">{{ zoneTypeLabel }}</p>
      </div>
      <span
        v-if="attentionCount"
        class="text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0"
        :class="attentionBadgeClass"
        data-test="farm-tile-attention-badge"
      >
        {{ attentionCount }}
      </span>
    </header>

    <div class="flex-1 min-h-0 px-2 pb-2 space-y-1 text-[11px] leading-snug overflow-hidden">
      <p v-if="plantsLine" class="text-zinc-200 truncate" data-test="farm-tile-plants">
        <span class="text-zinc-500">Plants ·</span> {{ plantsLine }}
      </p>
      <p v-if="lightLine" class="text-zinc-300 truncate" data-test="farm-tile-light">
        <span class="text-zinc-500">Light ·</span> {{ lightLine }}
      </p>
      <p v-if="waterLine" class="text-zinc-300 truncate" data-test="farm-tile-water">
        <span class="text-zinc-500">Water ·</span> {{ waterLine }}
      </p>
      <p class="truncate" :class="sensorLineClass" data-test="farm-tile-sensors">
        <span class="text-zinc-500">Sensors ·</span> {{ status.sensors.summary }}
      </p>
      <p
        v-if="status.greenhouse?.insideTemp || greenhouseClimateLine"
        class="text-emerald-300/90 truncate"
        data-test="farm-tile-greenhouse"
      >
        <span class="text-zinc-500">Climate ·</span>
        {{ status.greenhouse?.insideTemp || greenhouseClimateLine }}
      </p>
    </div>

    <div
      v-if="arrangeMode"
      class="absolute bottom-1 right-1 w-3 h-3 rounded-sm bg-zinc-600 border border-zinc-400 cursor-se-resize"
      data-test="farm-tile-resize-handle"
      aria-hidden="true"
      @pointerdown.stop="$emit('resize-start', $event)"
    />
  </article>
</template>

<script setup>
import { computed } from 'vue'
import HelpTip from './HelpTip.vue'
import { formatZoneTypeLabel } from '../lib/farmVisualStatus.js'

const props = defineProps({
  zone: { type: Object, required: true },
  status: { type: Object, required: true },
  arrangeMode: { type: Boolean, default: false },
  focused: { type: Boolean, default: false },
  showZoneTip: { type: Boolean, default: false },
})

defineEmits(['resize-start'])

const zoneTypeLabel = computed(() => formatZoneTypeLabel(props.zone?.zone_type))

const typeIcon = computed(() => {
  const t = String(props.zone?.zone_type || '').toLowerCase()
  if (t.includes('greenhouse')) return '🪴'
  if (t.includes('outdoor')) return '🌱'
  return '🏠'
})

const tileClasses = computed(() => {
  const h = props.status?.health || 'ok'
  const border = {
    ok: 'border-green-800/70',
    warn: 'border-amber-600/80',
    alert: 'border-red-600/80',
    unconfigured: 'border-zinc-700',
  }[h] || 'border-zinc-700'
  const greenhouse = String(props.zone?.zone_type || '').toLowerCase().includes('greenhouse')
    ? 'ring-1 ring-emerald-900/40'
    : ''
  const focus = props.focused ? 'ring-2 ring-green-500' : ''
  return [border, greenhouse, focus]
})

const plantsLine = computed(() => {
  const p = props.status?.plants
  if (!p || p.state === 'empty') return p?.label || 'Empty — ready to plant'
  const parts = [p.cropName, p.stage].filter(Boolean)
  return parts.join(', ')
})

const lightLine = computed(() => {
  const l = props.status?.light
  if (!l || l.state === 'none') return null
  if (l.state === 'scheduled') return l.scheduleLabel ? `Scheduled · ${l.scheduleLabel}` : 'On a schedule'
  if (l.state === 'on') return 'On now'
  if (l.state === 'off') return 'Off now'
  return null
})

const waterLine = computed(() => {
  const w = props.status?.water
  if (!w || w.kind === 'none') return null
  const parts = [w.label]
  if (w.nextRun) parts.push(w.nextRun)
  return parts.join(' · ')
})

const sensorLineClass = computed(() => {
  const s = props.status?.sensors?.state
  if (s === 'attention' || s === 'mixed' && props.status?.sensors?.worst === 'attention') {
    return 'text-amber-300'
  }
  if (s === 'not_set_up') return 'text-zinc-500'
  return 'text-green-300/90'
})

const attentionCount = computed(() => props.status?.attention?.length || 0)

const attentionBadgeClass = computed(() => {
  const sev = props.status?.attention?.[0]?.severity
  if (sev === 'critical' || sev === 'high') return 'bg-red-900/70 text-red-200'
  return 'bg-amber-900/60 text-amber-200'
})

const greenhouseClimateLine = computed(() => {
  const g = props.status?.greenhouse
  if (!g) return null
  const parts = []
  if (g.ventState) parts.push(`Vent ${g.ventState}`)
  if (g.shadeState) parts.push(`Shade ${g.shadeState}`)
  return parts.length ? parts.join(' · ') : null
})

const ariaLabel = computed(() => {
  const parts = [props.zone?.name, plantsLine.value, props.status?.sensors?.summary].filter(Boolean)
  return parts.join('. ')
})
</script>

<style scoped>
.farm-zone-tile {
  touch-action: none;
}
</style>
