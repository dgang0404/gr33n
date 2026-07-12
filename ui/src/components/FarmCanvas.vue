<template>
  <section class="space-y-3" data-test="farm-canvas">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <div>
        <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Your farm</h3>
        <p class="text-[11px] text-zinc-500 mt-0.5">
          {{ arrangeMode ? 'Drag zones to match your space — positions save automatically.' : 'Tap a zone for quick actions.' }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg border transition-colors"
          :class="arrangeMode ? 'bg-green-900/50 text-green-300 border-green-700' : 'bg-zinc-800 text-zinc-400 border-zinc-700 hover:border-zinc-600'"
          data-test="farm-canvas-arrange-toggle"
          @click="toggleArrange"
        >
          {{ arrangeMode ? 'Done arranging' : 'Arrange layout' }}
        </button>
        <label
          v-if="arrangeMode && farmId"
          class="text-xs px-3 py-1.5 rounded-lg bg-zinc-800 text-zinc-400 border border-zinc-700 hover:border-zinc-600 cursor-pointer"
        >
          {{ backgroundUploading ? 'Uploading…' : 'Background photo' }}
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            class="hidden"
            data-test="farm-canvas-background-input"
            @change="onBackgroundPick"
          />
        </label>
        <button
          v-if="arrangeMode && backgroundUrl"
          type="button"
          class="text-xs px-2 py-1.5 rounded-lg text-zinc-500 hover:text-red-400"
          data-test="farm-canvas-background-clear"
          @click="clearBackground"
        >
          Remove photo
        </button>
      </div>
    </div>

    <div
      v-if="!zones.length"
      class="rounded-xl border border-dashed border-zinc-700 bg-zinc-900/50 p-8 text-center"
      data-test="farm-canvas-empty"
    >
      <p class="text-sm text-zinc-300">Add your first zone to see your farm here.</p>
      <router-link
        v-nav-hint="'/zones'"
        to="/zones"
        class="inline-block mt-3 text-sm text-gr33n-400 hover:text-gr33n-300"
      >
        Go to My zones →
      </router-link>
    </div>

    <div
      v-else
      ref="canvasEl"
      class="farm-canvas-stage relative w-full rounded-xl border border-zinc-800 overflow-hidden bg-zinc-950 select-none"
      :class="{ 'ring-1 ring-green-800/40': arrangeMode }"
      tabindex="0"
      data-test="farm-canvas-stage"
      @keydown="onCanvasKeydown"
    >
      <img
        v-if="backgroundUrl"
        :src="backgroundUrl"
        alt=""
        class="absolute inset-0 w-full h-full object-cover opacity-35 pointer-events-none"
        data-test="farm-canvas-background"
      />
      <div
        v-else-if="arrangeMode"
        class="absolute inset-0 flex items-center justify-center pointer-events-none"
      >
        <p class="text-xs text-zinc-600 px-4 text-center">
          Add a photo or sketch of your yard or floor plan (optional)
        </p>
      </div>

      <div
        v-for="(entry, index) in zoneEntries"
        :key="entry.zone.id"
        class="absolute min-w-0 min-h-0"
        :style="entry.style"
        :data-test="`farm-canvas-zone-${entry.zone.id}`"
      >
        <div
          class="h-full w-full"
          :class="arrangeMode ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer'"
          :tabindex="arrangeMode ? 0 : -1"
          @pointerdown="(e) => onTilePointerDown(e, entry)"
          @click="(e) => onTileClick(e, entry.zone.id)"
        >
          <FarmCanvasZoneTile
            :zone="entry.zone"
            :status="entry.status"
            :arrange-mode="arrangeMode"
            :focused="focusedZoneId === entry.zone.id"
            :show-zone-tip="index === 0"
            @resize-start="(e) => onResizeStart(e, entry)"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import FarmCanvasZoneTile from './FarmCanvasZoneTile.vue'
import { computeZoneVisualStatus, resolveZoneLayout } from '../lib/farmVisualStatus.js'
import {
  layoutToStyle,
  nudgeLayout,
  pointerDeltaToLayout,
  resizeLayout,
} from '../lib/farmCanvasLayout.js'
import { sortZonesForStack, zoneHasTasksDueToday } from '../lib/zoneQuickActions.js'

const props = defineProps({
  farmId: { type: Number, default: null },
  zones: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
  actuators: { type: Array, default: () => [] },
  tasks: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  cropCycles: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
  backgroundUrl: { type: String, default: null },
})

const emit = defineEmits(['select-zone'])

const store = useFarmStore()

const arrangeMode = ref(false)
const canvasEl = ref(null)
const canvasSize = ref({ width: 1, height: 1 })
const localLayouts = ref({})
const focusedZoneId = ref(null)
const backgroundUploading = ref(false)
const saveTimers = ref({})

const dragState = ref(null)
const resizeState = ref(null)

function statusForZone(zone) {
  return computeZoneVisualStatus({
    zone,
    sensors: props.sensors,
    readings: props.readings,
    actuators: props.actuators,
    tasks: props.tasks,
    alerts: props.alerts,
    schedules: props.schedules,
    programs: props.programs,
    cropCycles: props.cropCycles,
    fertigationEvents: props.fertigationEvents,
  })
}

const sortedZones = computed(() =>
  sortZonesForStack(
    props.zones,
    statusForZone,
    (zoneId) => zoneHasTasksDueToday(props.tasks, zoneId),
  ),
)

const zoneEntries = computed(() =>
  sortedZones.value.map((zone, index) => {
    const layout = localLayouts.value[zone.id]
      ?? resolveZoneLayout(zone, (id) => store.zoneLayout(id), index)
    const status = statusForZone(zone)
    return {
      zone,
      layout,
      status,
      style: layoutToStyle(layout, canvasSize.value),
    }
  }),
)

function syncLayoutsFromStore() {
  const next = {}
  for (const z of props.zones) {
    const saved = store.zoneLayout(z.id)
    if (saved) next[z.id] = saved
  }
  localLayouts.value = next
}

function measureCanvas() {
  const el = canvasEl.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  canvasSize.value = { width: rect.width || 1, height: rect.height || 1 }
}

function toggleArrange() {
  arrangeMode.value = !arrangeMode.value
  if (!arrangeMode.value) focusedZoneId.value = null
}

function scheduleSave(zoneId, layout) {
  if (saveTimers.value[zoneId]) clearTimeout(saveTimers.value[zoneId])
  saveTimers.value[zoneId] = setTimeout(async () => {
    try {
      await store.saveZoneLayout(zoneId, layout)
    } catch { /* best effort */ }
  }, 400)
}

function onTilePointerDown(event, entry) {
  if (!arrangeMode.value || event.button !== 0) return
  focusedZoneId.value = entry.zone.id
  const rect = canvasEl.value?.getBoundingClientRect()
  if (!rect) return
  dragState.value = {
    zoneId: entry.zone.id,
    startX: event.clientX,
    startY: event.clientY,
    startLayout: { ...entry.layout },
    rect,
  }
  event.currentTarget?.setPointerCapture?.(event.pointerId)
  event.preventDefault()
}

function onResizeStart(event, entry) {
  if (!arrangeMode.value) return
  focusedZoneId.value = entry.zone.id
  resizeState.value = {
    zoneId: entry.zone.id,
    startX: event.clientX,
    startY: event.clientY,
    startLayout: { ...entry.layout },
    rect: canvasEl.value?.getBoundingClientRect(),
  }
  event.target?.setPointerCapture?.(event.pointerId)
  event.preventDefault()
  event.stopPropagation()
}

function onPointerMove(event) {
  if (dragState.value) {
    const next = pointerDeltaToLayout(
      event.clientX,
      event.clientY,
      dragState.value.startX,
      dragState.value.startY,
      dragState.value.rect,
      dragState.value.startLayout,
    )
    localLayouts.value = { ...localLayouts.value, [dragState.value.zoneId]: next }
  }
  if (resizeState.value && resizeState.value.rect) {
    const dx = (event.clientX - resizeState.value.startX) / resizeState.value.rect.width
    const dy = (event.clientY - resizeState.value.startY) / resizeState.value.rect.height
    const next = resizeLayout(resizeState.value.startLayout, 'se', dx, dy)
    localLayouts.value = { ...localLayouts.value, [resizeState.value.zoneId]: next }
  }
}

function onPointerUp() {
  if (dragState.value) {
    const id = dragState.value.zoneId
    const layout = localLayouts.value[id]
    if (layout) scheduleSave(id, layout)
    dragState.value = null
  }
  if (resizeState.value) {
    const id = resizeState.value.zoneId
    const layout = localLayouts.value[id]
    if (layout) scheduleSave(id, layout)
    resizeState.value = null
  }
}

function onTileClick(event, zoneId) {
  if (arrangeMode.value) {
    event.preventDefault()
    return
  }
  const zone = props.zones.find((z) => z.id === zoneId)
  if (!zone) return
  const entry = zoneEntries.value.find((e) => e.zone.id === zoneId)
  emit('select-zone', zone, entry?.status)
}

function onCanvasKeydown(event) {
  if (!arrangeMode.value || !focusedZoneId.value) return
  const map = {
    ArrowLeft: 'left',
    ArrowRight: 'right',
    ArrowUp: 'up',
    ArrowDown: 'down',
  }
  const dir = map[event.key]
  if (!dir) return
  event.preventDefault()
  const current = localLayouts.value[focusedZoneId.value]
    ?? resolveZoneLayout(
      props.zones.find((z) => z.id === focusedZoneId.value),
      (id) => store.zoneLayout(id),
      0,
    )
  const next = nudgeLayout(current, dir)
  localLayouts.value = { ...localLayouts.value, [focusedZoneId.value]: next }
  scheduleSave(focusedZoneId.value, next)
}

async function onBackgroundPick(event) {
  const file = event.target.files?.[0]
  if (!file || !props.farmId) return
  backgroundUploading.value = true
  try {
    await store.uploadLayoutBackground(props.farmId, file)
  } finally {
    backgroundUploading.value = false
    event.target.value = ''
  }
}

async function clearBackground() {
  if (!props.farmId) return
  await store.clearLayoutBackground(props.farmId)
}

let resizeObserver
onMounted(() => {
  measureCanvas()
  syncLayoutsFromStore()
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
  if (typeof ResizeObserver !== 'undefined' && canvasEl.value) {
    resizeObserver = new ResizeObserver(measureCanvas)
    resizeObserver.observe(canvasEl.value)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  resizeObserver?.disconnect()
  Object.values(saveTimers.value).forEach((t) => clearTimeout(t))
})

watch(() => props.zones, syncLayoutsFromStore, { deep: true })
</script>

<style scoped>
.farm-canvas-stage {
  aspect-ratio: 16 / 10;
  min-height: 280px;
}
</style>
