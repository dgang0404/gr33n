<template>
  <section class="md:hidden space-y-3" data-test="farm-zone-stack">
    <div>
      <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-widest">Your farm</h3>
      <p class="text-[11px] text-zinc-500 mt-0.5">
        <template v-if="filterLabel">Showing {{ entries.length }} of {{ totalZoneCountResolved }} zones · {{ filterLabel }}</template>
        <template v-else>Tap a zone for quick actions.</template>
      </p>
    </div>

    <div
      v-if="!entries.length && filterLabel"
      class="rounded-xl border border-dashed border-zinc-700 bg-zinc-900/50 p-6 text-center"
      data-test="farm-zone-stack-filter-empty"
    >
      <p class="text-sm text-zinc-300">No zones match "{{ filterLabel }}" right now.</p>
    </div>

    <div
      v-else-if="!entries.length"
      class="rounded-xl border border-dashed border-zinc-700 bg-zinc-900/50 p-6 text-center"
      data-test="farm-zone-stack-empty"
    >
      <p class="text-sm text-zinc-300">Add your first zone to see your farm here.</p>
      <router-link
        v-nav-hint="'/zones'"
        to="/zones"
        class="inline-block mt-3 text-sm text-gr33n-400 hover:text-gr33n-300 min-h-[44px] leading-[44px]"
      >
        Go to My zones →
      </router-link>
    </div>

    <div v-else class="space-y-3">
      <button
        v-for="entry in pagedEntries"
        :key="entry.zone.id"
        type="button"
        class="w-full text-left min-h-[44px] rounded-xl focus:outline-none focus:ring-2 focus:ring-green-600"
        :data-test="`farm-zone-stack-card-${entry.zone.id}`"
        :aria-label="`Open actions for ${entry.zone.name}`"
        @click="$emit('select-zone', entry.zone, entry.status)"
      >
        <FarmCanvasZoneTile
          :zone="entry.zone"
          :status="entry.status"
          :arrange-mode="false"
          class="min-h-[120px]"
        />
      </button>

      <div
        v-if="showPager"
        class="flex items-center justify-between pt-1"
        data-test="farm-zone-stack-pager"
      >
        <button
          type="button"
          class="min-h-[44px] px-3 text-sm text-zinc-400 hover:text-white disabled:opacity-40 disabled:hover:text-zinc-400"
          data-test="farm-zone-stack-prev"
          :disabled="page === 0"
          @click="page -= 1"
        >
          ← Previous
        </button>
        <span class="text-[11px] text-zinc-500">Page {{ page + 1 }} of {{ totalPages }}</span>
        <button
          type="button"
          class="min-h-[44px] px-3 text-sm text-zinc-400 hover:text-white disabled:opacity-40 disabled:hover:text-zinc-400"
          data-test="farm-zone-stack-next"
          :disabled="page >= totalPages - 1"
          @click="page += 1"
        >
          Next →
        </button>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import FarmCanvasZoneTile from './FarmCanvasZoneTile.vue'
import { computeZoneVisualStatus } from '../lib/farmVisualStatus.js'
import { sortZonesForStack, zoneHasTasksDueToday } from '../lib/zoneQuickActions.js'
import {
  MOBILE_PAGE_SIZE,
  paginateZones,
  shouldPageZoneStack,
  totalZonePages,
} from '../lib/farmTodayZoneFilter.js'

const props = defineProps({
  zones: { type: Array, default: () => [] },
  totalZoneCount: { type: Number, default: null },
  filterLabel: { type: String, default: '' },
  sensors: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
  actuators: { type: Array, default: () => [] },
  tasks: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  cropCycles: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
})

defineEmits(['select-zone'])

const statusCtx = computed(() => ({
  sensors: props.sensors,
  readings: props.readings,
  actuators: props.actuators,
  tasks: props.tasks,
  alerts: props.alerts,
  schedules: props.schedules,
  programs: props.programs,
  cropCycles: props.cropCycles,
  fertigationEvents: props.fertigationEvents,
}))

function statusFor(zone) {
  return computeZoneVisualStatus({ zone, ...statusCtx.value })
}

const entries = computed(() => {
  const sorted = sortZonesForStack(
    props.zones,
    statusFor,
    (zoneId) => zoneHasTasksDueToday(props.tasks, zoneId),
  )
  return sorted.map((zone) => ({
    zone,
    status: statusFor(zone),
  }))
})

const totalZoneCountResolved = computed(() => props.totalZoneCount ?? props.zones.length)

const page = ref(0)
const showPager = computed(() => shouldPageZoneStack(entries.value.length, MOBILE_PAGE_SIZE))
const totalPages = computed(() => totalZonePages(entries.value.length, MOBILE_PAGE_SIZE))
const pagedEntries = computed(() =>
  showPager.value ? paginateZones(entries.value, page.value, MOBILE_PAGE_SIZE) : entries.value,
)

watch(() => props.zones, () => { page.value = 0 })
</script>
