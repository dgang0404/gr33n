<template>
  <header
    class="flex flex-col sm:flex-row sm:items-start justify-between gap-3"
    data-test="farm-today-header"
  >
    <div class="min-w-0">
      <h2 class="text-xl font-bold text-white truncate">
        {{ farmName || 'Loading...' }}
        <HelpTip position="bottom">
          <strong>How it all connects:</strong> Your farm has <em>zones</em> (grow areas), each with <em>sensors</em>
          (reading temp, humidity, EC) and <em>controls</em> (pumps, lights, fans). <em>Feeding plans</em> say when each zone
          gets water and nutrients. <em>Automations</em> react to readings. <em>Tasks</em> are your daily to-do list.
          Open <router-link v-nav-hint="'/operator-guide'" to="/operator-guide" class="text-gr33n-400 underline">Guide</router-link> for a suggested click path.
        </HelpTip>
      </h2>
      <p class="text-sm text-zinc-500 mt-0.5">{{ subtitle }}</p>
      <div class="flex flex-wrap items-center gap-2 mt-2" data-test="farm-today-header-pills">
        <span
          v-if="rollup.healthy"
          class="text-xs px-2.5 py-1 rounded-full bg-green-950/50 text-green-300 border border-green-900/60"
          data-test="farm-today-pill-healthy"
        >
          {{ rollup.healthy }} healthy
        </span>
        <button
          v-if="rollup.attention"
          type="button"
          class="text-xs px-2.5 py-1 rounded-full bg-amber-950/50 text-amber-200 border border-amber-900/60 hover:border-amber-700 transition-colors"
          data-test="farm-today-pill-attention"
          @click="$emit('filter-attention')"
        >
          {{ rollup.attention }} need attention
        </button>
        <router-link
          v-if="rollup.tasksTodayCount"
          v-nav-hint="tasksLink"
          :to="tasksLink"
          class="text-xs px-2.5 py-1 rounded-full bg-zinc-900 text-zinc-300 border border-zinc-700 hover:border-zinc-600 transition-colors"
          data-test="farm-today-pill-tasks"
        >
          {{ rollup.tasksTodayCount }} task{{ rollup.tasksTodayCount === 1 ? '' : 's' }} today
          <span v-if="rollup.overdueTaskCount" class="text-amber-400 ml-1">
            ({{ rollup.overdueTaskCount }} overdue)
          </span>
        </router-link>
        <router-link
          v-if="rollup.unreadAlerts"
          v-nav-hint="alertsLink"
          :to="alertsLink"
          class="text-xs px-2.5 py-1 rounded-full bg-red-950/40 text-red-200 border border-red-900/60 hover:border-red-700 transition-colors"
          data-test="farm-today-pill-alerts"
        >
          {{ rollup.unreadAlerts }} alert{{ rollup.unreadAlerts === 1 ? '' : 's' }}
        </router-link>
      </div>
    </div>
    <button
      type="button"
      class="text-xs text-gr33n-400 hover:text-gr33n-300 transition-colors self-start sm:self-auto shrink-0"
      data-test="farm-today-refresh"
      @click="$emit('refresh')"
    >
      &#x21bb; Refresh
    </button>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import HelpTip from './HelpTip.vue'
import { buildFarmTodayRollup, todayHeaderSubtitle } from '../lib/farmTodayHeader.js'

const props = defineProps({
  farmName: { type: String, default: '' },
  zones: { type: Array, default: () => [] },
  getStatus: { type: Function, required: true },
  tasksTodayCount: { type: Number, default: 0 },
  unreadAlerts: { type: Number, default: 0 },
  overdueTaskCount: { type: Number, default: 0 },
  tasksLink: { type: [String, Object], required: true },
  alertsLink: { type: [String, Object], required: true },
  siteWeather: { type: Object, default: null },
})

defineEmits(['refresh', 'filter-attention'])

const rollup = computed(() => buildFarmTodayRollup({
  zones: props.zones,
  getStatus: props.getStatus,
  tasksTodayCount: props.tasksTodayCount,
  unreadAlerts: props.unreadAlerts,
  overdueTaskCount: props.overdueTaskCount,
}))

const subtitle = computed(() => todayHeaderSubtitle(props.siteWeather))
</script>
