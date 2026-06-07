<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <router-link v-nav-hint="'/zones'" to="/zones" class="text-xs text-zinc-500 hover:text-zinc-300">&larr; Back to zones</router-link>
        <h1 class="text-xl font-semibold text-white mt-1">{{ zone?.name || 'Zone' }}</h1>
        <p class="text-zinc-500 text-sm">{{ zone?.description || 'No description' }}</p>
        <p class="text-zinc-600 text-xs mt-1">
          What this zone needs: water & feeding, light, and air/climate — use the tabs below.
        </p>
      </div>
      <div class="flex flex-col items-end gap-2">
        <span :class="zoneBadge(zone?.zone_type)" class="text-xs font-medium px-2 py-1 rounded-full capitalize">
          {{ zone?.zone_type || 'unknown' }}
        </span>
        <AskGuardianButton
          v-if="zone"
          variant="primary"
          size="sm"
          :prefilled-message="zoneGuardianPrompt"
          :context-ref="zoneGuardianContextRef"
        />
        <GuardianStarterChips v-if="zone" :starters="zoneStarters" />
      </div>
    </div>

    <div v-if="!zone" class="text-zinc-500 text-sm">Zone not found.</div>

    <template v-else>
      <ZoneAdvancedHint class="mb-2" />

      <div class="flex flex-wrap gap-1 border-b border-zinc-800">
        <button
          v-for="tab in zoneTabs"
          :key="tab.id"
          type="button"
          class="px-4 py-2 text-sm font-medium rounded-t-lg transition-colors"
          :class="activeTab === tab.id ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-zinc-300'"
          @click="activeTab = tab.id"
        >
          {{ tab.icon }} {{ tab.label }}
        </button>
      </div>

      <ZoneNeedSection
        v-if="activeTab === PLANT_NEEDS.water"
        :need="PLANT_NEEDS.water"
        :zone-id="zoneId"
        :farm-id="farmId"
        :zone="zone"
        :sensors="sensors"
        :actuators="actuators"
        :setpoints="zoneSetpoints"
        :schedules="schedules"
        :rules="rules"
        :programs="programs"
        :active-program="activeProgram"
        :ec-targets="ecTargets"
        :reservoirs="reservoirs"
        :actuator-events="actuatorEvents"
        :fertigation-events="zoneEvents"
        :toggling="toggling"
        @toggle-actuator="toggleActuator"
        @refresh-events="loadEvents"
        @setpoints-updated="loadSetpoints"
        @rules-updated="loadRules"
        @water-refreshed="onWaterRefreshed"
        @plan-updated="refreshFeedingPlan"
      />

      <ZoneNeedSection
        v-else-if="activeTab === PLANT_NEEDS.light"
        :need="PLANT_NEEDS.light"
        :zone-id="zoneId"
        :farm-id="farmId"
        :zone="zone"
        :sensors="sensors"
        :actuators="actuators"
        :setpoints="zoneSetpoints"
        :schedules="schedules"
        :rules="rules"
        :lighting-programs="lightingPrograms"
        :actuator-events="actuatorEvents"
        :toggling="toggling"
        @toggle-actuator="toggleActuator"
        @refresh-events="loadEvents"
        @setpoints-updated="loadSetpoints"
        @rules-updated="loadRules"
      />

      <ZoneNeedSection
        v-else-if="activeTab === PLANT_NEEDS.air"
        :need="PLANT_NEEDS.air"
        :zone-id="zoneId"
        :farm-id="farmId"
        :zone="zone"
        :is-greenhouse="isGreenhouse"
        :sensors="sensors"
        :actuators="actuators"
        :setpoints="zoneSetpoints"
        :schedules="schedules"
        :rules="rules"
        :actuator-events="actuatorEvents"
        :toggling="toggling"
        @toggle-actuator="toggleActuator"
        @refresh-events="loadEvents"
        @setpoints-updated="loadSetpoints"
        @rules-updated="loadRules"
      />

      <template v-else>
        <ZoneTodayStrip :chips="todaySnapshot.chips" />

        <ZoneAlertsPanel
          v-if="zone"
          :zone-id="zoneId"
          :zone-name="zone.name"
          :sensors="sensors"
          :alerts="store.alerts"
          @refresh="refreshZoneAlerts"
        />

        <ZoneTasksPanel
          v-if="zone"
          :zone-id="zoneId"
          :tasks="tasks"
          @refresh="refreshZoneTasks"
        />

        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <button
            type="button"
            class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 text-left hover:border-green-800/50 transition-colors"
            @click="activeTab = PLANT_NEEDS.water"
          >
            <p class="text-zinc-400 text-xs mb-1">💧 Water</p>
            <p class="text-white text-sm">{{ waterSummary }}</p>
          </button>
          <button
            type="button"
            class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 text-left hover:border-green-800/50 transition-colors"
            @click="activeTab = PLANT_NEEDS.light"
          >
            <p class="text-zinc-400 text-xs mb-1">💡 Light</p>
            <p class="text-white text-sm">{{ lightSummary }}</p>
          </button>
          <button
            type="button"
            class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 text-left hover:border-green-800/50 transition-colors"
            @click="activeTab = PLANT_NEEDS.air"
          >
            <p class="text-zinc-400 text-xs mb-1">🌬️ Climate</p>
            <p class="text-white text-sm">{{ airSummary }}</p>
          </button>
          <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
            <p class="text-zinc-400 text-xs mb-1">Due today</p>
            <p class="text-white text-2xl font-semibold">{{ zoneTasksDueToday.length }}</p>
          </div>
        </div>

        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-white">Reference photos</h2>
            <label class="text-xs text-green-600 hover:text-green-400 cursor-pointer">
              <input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" :disabled="photoUploading" @change="onPhotoSelected" />
              {{ photoUploading ? 'Uploading…' : '+ Add photo' }}
            </label>
          </div>
          <p v-if="photoError" class="text-red-400 text-xs mb-2">{{ photoError }}</p>
          <p v-else-if="photosLoading" class="text-zinc-500 text-sm">Loading photos…</p>
          <p v-else-if="!zonePhotos.length" class="text-zinc-500 text-sm">Walkthrough or crop reference photos for this zone.</p>
          <div v-else class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
            <div v-for="p in zonePhotos" :key="p.id" class="bg-zinc-950 border border-zinc-800 rounded-lg overflow-hidden group">
              <button type="button" class="block w-full aspect-square" @click="openPhoto(p)">
                <img :src="photoThumbUrl(p)" :alt="p.file_name || 'Zone photo'" class="w-full h-full object-cover" loading="lazy" />
              </button>
              <div class="px-2 py-1.5 flex items-center justify-between gap-1">
                <p class="text-zinc-500 text-[10px] truncate flex-1">{{ p.file_name }}</p>
                <button type="button" class="text-zinc-600 hover:text-red-400 text-[10px] shrink-0" :disabled="photoDeleting[p.id]" @click="removePhoto(p)">Remove</button>
              </div>
            </div>
          </div>
        </div>

      </template>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import {
  PLANT_NEEDS,
  NEED_META,
  sensorPlantNeed,
  actuatorPlantNeed,
} from '../lib/plantNeeds.js'
import AskGuardianButton from '../components/AskGuardianButton.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import ZoneNeedSection from '../components/ZoneNeedSection.vue'
import ZoneTodayStrip from '../components/ZoneTodayStrip.vue'
import ZoneAlertsPanel from '../components/ZoneAlertsPanel.vue'
import ZoneTasksPanel from '../components/ZoneTasksPanel.vue'
import ZoneAdvancedHint from '../components/ZoneAdvancedHint.vue'
import { zoneTasksDueToday as filterZoneTasksDueToday } from '../lib/zoneTasks.js'
import {
  buildZoneGuardianContextRef,
  buildZoneGuardianPrompt,
} from '../lib/guardianContextPrompts.js'
import { buildSetupStarters, buildZoneStarters } from '../lib/guardianStarters.js'
import { computeZoneTodaySnapshot, pickNextZoneSchedule } from '../lib/zoneGrowSummary.js'

const route = useRoute()
const activeTab = ref('overview')
const store = useFarmStore()
const farmContext = useFarmContextStore()
const toggling = ref({})
const zonePhotos = ref([])
const photoThumbUrls = ref({})
const photosLoading = ref(false)
const photoUploading = ref(false)
const photoDeleting = ref({})
const photoError = ref('')

const programs = ref([])
const events = ref([])
const ecTargets = ref([])
const reservoirs = ref([])
const actuatorEvents = ref([])
const eventsLoading = ref(false)
const schedules = ref([])
const tasks = ref([])
const setpoints = ref([])
const rules = ref([])
const lightingPrograms = ref([])
const cropCycles = ref([])
const waterQueueDepth = ref(0)

const zoneTabs = [
  { id: 'overview', icon: '📋', label: 'Overview' },
  { id: PLANT_NEEDS.water, icon: NEED_META[PLANT_NEEDS.water].icon, label: NEED_META[PLANT_NEEDS.water].shortLabel },
  { id: PLANT_NEEDS.light, icon: NEED_META[PLANT_NEEDS.light].icon, label: NEED_META[PLANT_NEEDS.light].shortLabel },
  { id: PLANT_NEEDS.air, icon: NEED_META[PLANT_NEEDS.air].icon, label: NEED_META[PLANT_NEEDS.air].shortLabel },
]

const farmId = computed(() => farmContext.farmId)
const zoneId = computed(() => Number(route.params.id))
const zone = computed(() => store.zones.find(z => z.id === zoneId.value))
const isGreenhouse = computed(() => String(zone.value?.zone_type || '').toLowerCase() === 'greenhouse')
const sensors = computed(() => store.sensorsByZone(zoneId.value))
const actuators = computed(() => store.actuatorsByZone(zoneId.value))
const zoneSetpoints = computed(() => setpoints.value)
const activeProgram = computed(() =>
  programs.value.find(p => p.target_zone_id === zoneId.value && p.is_active),
)
const zoneEvents = computed(() =>
  events.value.filter(e => e.zone_id === zoneId.value).sort((a, b) => new Date(b.applied_at) - new Date(a.applied_at)),
)
const zoneTasks = computed(() =>
  tasks.value.filter(t => t.zone_id === zoneId.value && t.status !== 'completed' && t.status !== 'cancelled'),
)
const zoneTasksDueToday = computed(() => filterZoneTasksDueToday(tasks.value, zoneId.value, 99))
const zoneDevices = computed(() => store.devicesByZone(zoneId.value))

const nextZoneSchedule = computed(() => {
  if (!zone.value) return null
  return pickNextZoneSchedule({
    zoneId: zoneId.value,
    zoneName: zone.value.name,
    schedules: schedules.value,
    activeProgram: activeProgram.value,
    lightingPrograms: lightingPrograms.value,
  })
})

const todaySnapshot = computed(() => {
  if (!zone.value) {
    return { chips: [], unreadAlerts: [], activeRulesCount: 0, queueDepth: 0, missingComfortTargets: 0 }
  }
  return computeZoneTodaySnapshot({
    zone: zone.value,
    sensors: sensors.value,
    devices: zoneDevices.value,
    alerts: store.alerts,
    rules: rules.value,
    schedules: schedules.value,
    setpoints: zoneSetpoints.value,
    activeProgram: activeProgram.value,
    lightingPrograms: lightingPrograms.value,
    queueDepth: waterQueueDepth.value,
    zoneTasks: zoneTasksDueToday.value,
  })
})

const guardianSnapshotCtx = computed(() => ({
  zone: zone.value,
  activeTab: activeTab.value,
  unreadAlerts: todaySnapshot.value.unreadAlerts,
  queueDepth: waterQueueDepth.value,
  missingComfortTargets: todaySnapshot.value.missingComfortTargets,
  offlineDevices: zoneDevices.value.filter(d => d.status !== 'online').length,
  nextSchedule: nextZoneSchedule.value,
  activeRulesCount: todaySnapshot.value.activeRulesCount,
  activeProgramName: activeProgram.value?.name,
}))

const zoneGuardianPrompt = computed(() =>
  zone.value ? buildZoneGuardianPrompt(guardianSnapshotCtx.value) : '',
)

const zoneGuardianContextRef = computed(() =>
  zone.value ? buildZoneGuardianContextRef(guardianSnapshotCtx.value) : null,
)

const zoneStarterSurface = computed(() => {
  const t = activeTab.value
  if (t === PLANT_NEEDS.water) return 'zone_water'
  if (t === PLANT_NEEDS.light) return 'zone_light'
  if (t === PLANT_NEEDS.air) return 'zone_climate'
  return 'zone_overview'
})

const hasActiveCropCycle = computed(() =>
  cropCycles.value.some(
    (c) => c.is_active && Number(c.zone_id) === Number(zoneId.value),
  ),
)

const zoneStarters = computed(() => {
  if (!zone.value) return []
  if (!hasActiveCropCycle.value) {
    return buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: farmId.value,
      zoneCount: 1,
      zones: [zone.value],
      zoneName: zone.value.name,
      activeCycles: cropCycles.value,
      unreadAlerts: todaySnapshot.value.unreadAlerts,
      deviceOffline: zoneDevices.value.some((d) => d.status !== 'online'),
    })
  }
  return buildZoneStarters(zoneStarterSurface.value, guardianSnapshotCtx.value)
})

const waterSummary = computed(() => {
  const n = sensors.value.filter(s => sensorPlantNeed(s.sensor_type) === PLANT_NEEDS.water).length
  return activeProgram.value?.name || `${n} sensor(s)`
})
const lightSummary = computed(() => {
  const lp = lightingPrograms.value.filter(l => l.zone_id === zoneId.value && l.is_active)
  return lp[0]?.name || `${actuators.value.filter(a => actuatorPlantNeed(a.actuator_type) === PLANT_NEEDS.light).length} light(s)`
})
const airSummary = computed(() => {
  const n = sensors.value.filter(s => sensorPlantNeed(s.sensor_type) === PLANT_NEEDS.air).length
  return isGreenhouse.value ? `Greenhouse · ${n} sensor(s)` : `${n} sensor(s)`
})

watch(() => route.query.tab, (t) => {
  if (t && zoneTabs.some(z => z.id === t)) activeTab.value = t
}, { immediate: true })

watch(() => activeProgram.value?.id, () => {
  loadWaterQueueDepth()
})

async function loadEvents() {
  eventsLoading.value = true
  try {
    const all = []
    for (const a of actuators.value) {
      const evts = await store.loadActuatorEvents(a.id, { limit: 10 })
      all.push(...evts)
    }
    all.sort((a, b) => new Date(b.event_time) - new Date(a.event_time))
    actuatorEvents.value = all.slice(0, 30)
  } catch {
    actuatorEvents.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function toggleActuator(a) {
  toggling.value[a.id] = true
  try {
    await store.toggleActuator(a.id, a.current_state_text || 'offline')
  } finally {
    toggling.value[a.id] = false
  }
}

onMounted(async () => {
  if (!store.zones.length && farmId.value) await store.loadAll(farmId.value)
  const fid = farmId.value
  const [p, e, s, ec, res, cycles] = await Promise.all([
    store.loadFertigationPrograms(fid),
    store.loadFertigationEvents(fid),
    store.loadSchedules(fid),
    store.loadEcTargets(fid),
    store.loadReservoirs(fid),
    store.loadCropCycles(fid),
  ])
  cropCycles.value = cycles
  programs.value = p
  events.value = e
  schedules.value = s
  ecTargets.value = ec
  reservoirs.value = res
  await store.loadTasks(fid)
  tasks.value = store.tasks
  await loadSetpoints()
  await loadRules()
  try {
    const lr = await api.get(`/farms/${fid}/lighting-programs`)
    lightingPrograms.value = lr.data?.programs ?? lr.data ?? []
  } catch {
    lightingPrograms.value = []
  }
  await Promise.all([loadEvents(), loadZonePhotos(), loadWaterQueueDepth(), store.loadAlerts(fid)])
})

async function loadRules() {
  const fid = farmId.value
  if (!fid) return
  try {
    rules.value = await store.loadAutomationRules(fid)
  } catch {
    rules.value = []
  }
}

async function loadSetpoints() {
  const fid = farmId.value
  if (!fid) return
  try {
    const res = await api.get(`/farms/${fid}/setpoints`, { params: { zone_id: zoneId.value } })
    setpoints.value = res.data ?? []
  } catch {
    setpoints.value = []
  }
}

async function refreshZoneAlerts() {
  const fid = farmId.value
  if (!fid) return
  await store.loadAlerts(fid)
}

async function refreshFeedingPlan() {
  const fid = farmId.value
  if (!fid) return
  const [p, s, ec, res] = await Promise.all([
    store.loadFertigationPrograms(fid),
    store.loadSchedules(fid),
    store.loadEcTargets(fid),
    store.loadReservoirs(fid),
  ])
  programs.value = p
  schedules.value = s
  ecTargets.value = ec
  reservoirs.value = res
  await loadWaterQueueDepth()
}

async function refreshZoneTasks() {
  const fid = farmId.value
  if (!fid) return
  await store.loadTasks(fid)
  tasks.value = store.tasks
}

function onWaterRefreshed(status) {
  waterQueueDepth.value = status?.queue_depth ?? 0
}

async function loadWaterQueueDepth() {
  waterQueueDepth.value = 0
  const prog = activeProgram.value
  if (!prog?.id) return
  try {
    const token = store.token || localStorage.getItem('token')
    const r = await fetch(`/fertigation/programs/${prog.id}/water-status`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (r.ok) {
      const data = await r.json()
      waterQueueDepth.value = data.queue_depth ?? 0
    }
  } catch {
    // non-fatal
  }
}

function photoThumbUrl(p) {
  return photoThumbUrls.value[p.id] || ''
}

function revokePhotoThumbUrls() {
  for (const url of Object.values(photoThumbUrls.value)) {
    if (url) URL.revokeObjectURL(url)
  }
  photoThumbUrls.value = {}
}

async function loadZonePhotos() {
  photosLoading.value = true
  photoError.value = ''
  revokePhotoThumbUrls()
  try {
    const r = await api.get(`/zones/${zoneId.value}/photos`)
    zonePhotos.value = r.data?.photos ?? []
    const thumbs = {}
    await Promise.all(zonePhotos.value.map(async (p) => {
      try {
        const img = await api.get(`/file-attachments/${p.id}/content`, { responseType: 'blob' })
        thumbs[p.id] = URL.createObjectURL(img.data)
      } catch { /* optional */ }
    }))
    photoThumbUrls.value = thumbs
  } catch (e) {
    zonePhotos.value = []
    photoError.value = e.response?.data?.error || e.message || 'Could not load photos'
  } finally {
    photosLoading.value = false
  }
}

onUnmounted(revokePhotoThumbUrls)

async function onPhotoSelected(ev) {
  const file = ev.target?.files?.[0]
  ev.target.value = ''
  if (!file || !zone.value) return
  photoUploading.value = true
  photoError.value = ''
  try {
    const fd = new FormData()
    fd.append('file', file)
    await api.post(`/zones/${zoneId.value}/photos`, fd)
    if (farmId.value) await store.loadAll(farmId.value)
    await loadZonePhotos()
  } catch (e) {
    photoError.value = e.response?.data?.error || e.message || 'Upload failed'
  } finally {
    photoUploading.value = false
  }
}

async function openPhoto(p) {
  try {
    const r = await api.get(`/file-attachments/${p.id}/download`)
    const url = String(r.data?.url || '')
    if (!url) throw new Error('Missing download URL')
    const finalUrl = url.startsWith('http://') || url.startsWith('https://')
      ? url
      : `${api.defaults.baseURL}${url}`
    window.open(finalUrl, '_blank', 'noopener')
  } catch (e) {
    photoError.value = e.response?.data?.error || e.message || 'Could not open photo'
  }
}

async function removePhoto(p) {
  if (!p?.id || photoDeleting.value[p.id]) return
  photoDeleting.value[p.id] = true
  photoError.value = ''
  try {
    await api.delete(`/zones/${zoneId.value}/photos/${p.id}`)
    if (farmId.value) await store.loadAll(farmId.value)
    await loadZonePhotos()
  } catch (e) {
    photoError.value = e.response?.data?.error || e.message || 'Could not remove photo'
  } finally {
    photoDeleting.value[p.id] = false
  }
}

const BADGE = {
  indoor: 'bg-indigo-900/60 text-indigo-300',
  outdoor: 'bg-emerald-900/60 text-emerald-300',
  greenhouse: 'bg-green-900/60 text-green-300',
}
function zoneBadge(type) {
  if (!type) return 'bg-zinc-800 text-zinc-400'
  const k = String(type).toLowerCase()
  for (const [name, cls] of Object.entries(BADGE)) {
    if (k.includes(name)) return cls
  }
  return 'bg-zinc-800 text-zinc-400'
}
</script>
