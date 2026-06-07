<template>
  <div class="space-y-4" data-test="targets-schedules-panel">
    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="What runs when"
      :clear-route="{ path: '/comfort-targets', query: { tab: 'schedules' } }"
    />

    <div v-if="loading" class="text-zinc-400 text-sm">Loading schedules…</div>

    <EmptyStateHint
      v-else-if="!filteredSchedules.length"
      reason="automation_off"
      message="No schedules yet — add a daily run time below or use Advanced for full cron editing."
      action-label="Advanced schedules"
      action-to="/schedules"
    />

    <div v-else class="space-y-3">
      <article
        v-for="schedule in filteredSchedules"
        :key="schedule.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
        :data-test="`farmer-schedule-${schedule.id}`"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-sm text-white font-medium">{{ schedule.name }}</p>
            <p class="text-xs text-zinc-400 mt-1">{{ scheduleRunsLabel(schedule) }}</p>
            <p v-if="linkedProgram(schedule.id)" class="text-[11px] text-zinc-500 mt-1">
              Feeding plan: {{ linkedProgram(schedule.id).name }}
            </p>
            <p v-if="linkedLighting(schedule.id)" class="text-[11px] text-zinc-500 mt-1">
              Lighting: {{ linkedLighting(schedule.id) }}
            </p>
          </div>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded border shrink-0"
            :class="schedule.is_active ? 'border-green-700 text-green-400' : 'border-zinc-700 text-zinc-400'"
            :data-test="`toggle-schedule-${schedule.id}`"
            @click="toggleSchedule(schedule)"
          >
            {{ schedule.is_active ? 'On' : 'Paused' }}
          </button>
        </div>
      </article>
    </div>

    <details class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <summary class="text-sm text-white font-medium cursor-pointer">Add daily schedule</summary>
      <form class="mt-3 space-y-3 max-w-md" @submit.prevent="createDaily">
        <label class="block text-xs text-zinc-400">
          Name
          <input v-model="form.name" required class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" />
        </label>
        <label class="block text-xs text-zinc-400">
          Run every day at
          <input v-model="form.time" type="time" required class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" />
        </label>
        <label class="block text-xs text-zinc-400">
          Timezone
          <input v-model="form.timezone" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white" placeholder="UTC" />
        </label>
        <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>
        <button
          type="submit"
          class="text-xs px-3 py-1.5 rounded bg-green-800 hover:bg-green-700 text-white disabled:opacity-50"
          :disabled="saving"
        >
          {{ saving ? 'Creating…' : 'Create schedule' }}
        </button>
      </form>
    </details>

    <p class="text-xs text-zinc-600">
      Need cron strings or preconditions?
      <router-link v-nav-hint="'/schedules'" to="/schedules" class="text-green-600 hover:text-green-400">Advanced schedules →</router-link>
    </p>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import api from '../api'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from './ZoneContextBanner.vue'
import EmptyStateHint from './EmptyStateHint.vue'
import { filterSchedulesForZone } from '../lib/zoneContext.js'
import { buildDailyCron, scheduleRunsLabel } from '../lib/cronHumanize.js'

const props = defineProps({
  zoneContextId: { type: Number, default: null },
})

const emit = defineEmits(['refresh'])

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const saving = ref(false)
const formError = ref('')
const schedules = ref([])
const programs = ref([])
const tasks = ref([])
const lightingPrograms = ref([])
const form = ref({ name: '', time: '06:00', timezone: 'UTC' })

const filteredSchedules = computed(() => {
  if (!props.zoneContextId) return schedules.value
  const zone = store.zones.find((z) => z.id === props.zoneContextId)
  return filterSchedulesForZone(
    schedules.value,
    props.zoneContextId,
    zone?.name || '',
    programs.value,
    lightingPrograms.value,
    tasks.value,
  )
})

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function linkedProgram(scheduleId) {
  return programs.value.find((p) => Number(p.schedule_id) === Number(scheduleId) && p.is_active)
}

function linkedLighting(scheduleId) {
  const lp = lightingPrograms.value.find(
    (p) => Number(p.schedule_on_id) === Number(scheduleId) || Number(p.schedule_off_id) === Number(scheduleId),
  )
  return lp?.name || ''
}

async function loadData() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [s, p, lp] = await Promise.all([
      store.loadSchedules(fid),
      store.loadFertigationPrograms(fid),
      api.get(`/farms/${fid}/lighting-programs`).catch(() => ({ data: [] })),
    ])
    schedules.value = s
    programs.value = p
    lightingPrograms.value = lp.data?.programs ?? lp.data ?? []
    await store.loadTasks(fid)
    tasks.value = store.tasks
  } finally {
    loading.value = false
  }
}

async function toggleSchedule(schedule) {
  const updated = await store.updateScheduleActive(schedule.id, !schedule.is_active)
  const idx = schedules.value.findIndex((s) => s.id === schedule.id)
  if (idx >= 0) schedules.value[idx] = updated
  emit('refresh')
}

async function createDaily() {
  formError.value = ''
  const fid = farmContext.farmId
  if (!fid) return
  const [hourStr, minuteStr] = String(form.value.time || '06:00').split(':')
  saving.value = true
  try {
    await store.createSchedule(fid, {
      name: form.value.name.trim(),
      description: null,
      schedule_type: 'cron',
      cron_expression: buildDailyCron(Number(hourStr), Number(minuteStr)),
      timezone: form.value.timezone.trim() || 'UTC',
      is_active: true,
      preconditions: [],
    })
    form.value = { name: '', time: '06:00', timezone: form.value.timezone || 'UTC' }
    await loadData()
    emit('refresh')
  } catch (e) {
    formError.value = e?.response?.data?.error || e.message || 'Could not create schedule'
  } finally {
    saving.value = false
  }
}

onMounted(loadData)

defineExpose({ loadData })
</script>
