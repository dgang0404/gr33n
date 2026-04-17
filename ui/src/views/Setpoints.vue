<!--
  Setpoints view — Phase 20.6 WS4.

  One page that answers "what *should* this zone/cycle look like right
  now, per stage?" and lets operators edit the answer. Rules built with
  a setpoint-typed predicate (see RuleForm.vue "Use setpoint from
  zone/cycle" toggle) resolve these rows at every tick.
-->
<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <h1 class="text-xl font-semibold text-white">Setpoints</h1>
        <HelpTip position="bottom">
          Setpoints express the ideal environment for a zone or a crop cycle at a growth stage
          as first-class data. The rule engine resolves them in this precedence order:
          <strong>cycle + stage → cycle (any stage) → zone + stage → zone (any stage)</strong>.
          Rules written once auto-adjust as cycles advance — change <em>mid_flower</em>'s
          dew point target and every setpoint-typed rule for that cycle picks it up on
          the next tick.
        </HelpTip>
      </div>
      <button
        @click="refresh"
        class="text-xs text-zinc-400 hover:text-zinc-200"
      >Refresh</button>
    </div>

    <div v-if="!farmContext.farmId" class="text-zinc-400 text-sm">
      Select a farm to manage its setpoints.
    </div>

    <template v-else>
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">
          {{ filtered.length }} setpoint(s)
          <span v-if="filterZone || filterCycle || filterSensorType"> (filtered)</span>
        </p>
        <button
          @click="showNewForm = !showNewForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg"
        >{{ showNewForm ? 'Cancel' : '+ New Setpoint' }}</button>
      </div>

      <!-- Filters -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
        <select v-model.number="filterZone" class="input-field text-sm">
          <option :value="null">All zones</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <select v-model.number="filterCycle" class="input-field text-sm">
          <option :value="null">All crop cycles</option>
          <option v-for="c in cropCycles" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <input
          v-model="filterSensorType"
          type="text"
          placeholder="Filter by sensor_type (substring)"
          class="input-field text-sm"
        />
      </div>

      <!-- Create form -->
      <SetpointRow
        v-if="showNewForm"
        :value="newRow"
        :zones="zones"
        :crop-cycles="cropCycles"
        :busy="saving"
        @save="onCreate"
      />

      <!-- List -->
      <div v-if="loading" class="text-zinc-400 text-sm">Loading…</div>
      <div v-else-if="filtered.length === 0" class="text-zinc-500 text-sm">
        No setpoints match the current filters. Start with the +New button above to tune an
        "ideal" for this zone; add a cycle-scoped override later for stage-specific deltas.
      </div>
      <div v-else class="space-y-2">
        <SetpointRow
          v-for="sp in filtered"
          :key="sp.id"
          :value="sp"
          :zones="zones"
          :crop-cycles="cropCycles"
          :busy="saving"
          @save="onUpdate"
          @delete="onDelete"
        />
      </div>

      <p v-if="errorMsg" class="text-red-400 text-sm">{{ errorMsg }}</p>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import api from '../api'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'
import SetpointRow from '../components/SetpointRow.vue'

const farmContext = useFarmContextStore()

const setpoints = ref([])
const zones = ref([])
const cropCycles = ref([])

const loading = ref(false)
const saving = ref(false)
const errorMsg = ref('')

const showNewForm = ref(false)
const newRow = ref(emptyRow())

const filterZone = ref(null)
const filterCycle = ref(null)
const filterSensorType = ref('')

const filtered = computed(() => {
  return setpoints.value.filter(sp => {
    if (filterZone.value && sp.zone_id !== filterZone.value) return false
    if (filterCycle.value && sp.crop_cycle_id !== filterCycle.value) return false
    if (filterSensorType.value && !sp.sensor_type?.toLowerCase().includes(filterSensorType.value.toLowerCase())) return false
    return true
  })
})

function emptyRow() {
  return {
    id: null,
    zone_id: null,
    crop_cycle_id: null,
    stage: null,
    sensor_type: '',
    min_value: null,
    max_value: null,
    ideal_value: null,
  }
}

async function refresh() {
  if (!farmContext.farmId) return
  loading.value = true
  errorMsg.value = ''
  try {
    const [spRes, zRes, ccRes] = await Promise.all([
      api.get(`/farms/${farmContext.farmId}/setpoints`),
      api.get(`/farms/${farmContext.farmId}/zones`),
      api.get(`/farms/${farmContext.farmId}/crop-cycles`).catch(() => ({ data: [] })),
    ])
    setpoints.value = spRes.data ?? []
    zones.value = zRes.data ?? []
    cropCycles.value = ccRes.data ?? []
    if (newRow.value.zone_id == null && zones.value.length) {
      newRow.value.zone_id = zones.value[0].id
    }
  } catch (e) {
    errorMsg.value = e.response?.data?.error ?? e.message
  } finally {
    loading.value = false
  }
}

async function onCreate({ payload }) {
  if (!farmContext.farmId) return
  saving.value = true
  errorMsg.value = ''
  try {
    const { data } = await api.post(`/farms/${farmContext.farmId}/setpoints`, payload)
    setpoints.value.push(data)
    showNewForm.value = false
    newRow.value = emptyRow()
    if (zones.value.length) newRow.value.zone_id = zones.value[0].id
  } catch (e) {
    errorMsg.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

async function onUpdate({ id, payload }) {
  saving.value = true
  errorMsg.value = ''
  try {
    const { data } = await api.put(`/setpoints/${id}`, payload)
    const idx = setpoints.value.findIndex(s => s.id === id)
    if (idx !== -1) setpoints.value[idx] = data
  } catch (e) {
    errorMsg.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

async function onDelete(id) {
  saving.value = true
  errorMsg.value = ''
  try {
    await api.delete(`/setpoints/${id}`)
    setpoints.value = setpoints.value.filter(s => s.id !== id)
  } catch (e) {
    errorMsg.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

onMounted(refresh)
watch(() => farmContext.farmId, refresh)
</script>
