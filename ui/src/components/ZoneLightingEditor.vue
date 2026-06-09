<template>
  <div class="space-y-3" data-test="zone-lighting-editor">
    <div class="flex items-center justify-between gap-2 flex-wrap">
      <p class="text-zinc-500 text-xs">Edit photoperiod for this zone — same endpoints as the fleet lighting admin.</p>
      <div class="flex gap-2">
        <button
          type="button"
          class="text-xs px-2.5 py-1 rounded bg-gr33n-700 hover:bg-gr33n-600 text-white"
          data-test="zone-lighting-create"
          @click="openCreate"
        >+ New program</button>
        <router-link
          v-nav-hint="'/zones'"
          :to="{ path: '/zones', query: { tab: 'fleet', fleet: 'lighting' } }"
          class="text-xs text-zinc-500 hover:text-green-400"
        >All zones (Fleet) →</router-link>
      </div>
    </div>

    <p v-if="error" class="text-xs text-red-400">{{ error }}</p>
    <p v-if="loading" class="text-zinc-500 text-sm">Loading lighting programs…</p>
    <EmptyStateHint
      v-else-if="!programs.length"
      reason="no_data"
      message="No lighting program for this zone yet."
      action-label="Create program"
      :action-to="null"
      compact
      @action="openCreate"
    />

    <ul v-else class="space-y-2">
      <li
        v-for="prog in programs"
        :key="prog.id"
        class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2"
        :data-test="`zone-lighting-program-${prog.id}`"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <p class="text-sm text-zinc-200 font-medium">{{ prog.name }}</p>
            <p class="text-xs text-zinc-500 mt-0.5">{{ summary(prog) }}</p>
          </div>
          <div class="flex flex-wrap gap-1 shrink-0">
            <button type="button" class="text-[10px] px-2 py-0.5 rounded border border-zinc-700 text-zinc-400" @click="openEdit(prog)">Edit</button>
            <button
              v-if="prog.is_active"
              type="button"
              class="text-[10px] px-2 py-0.5 rounded border border-yellow-800/50 text-yellow-400"
              @click="setActive(prog, false)"
            >Pause</button>
            <button
              v-else
              type="button"
              class="text-[10px] px-2 py-0.5 rounded border border-green-800/50 text-green-400"
              @click="setActive(prog, true)"
            >Activate</button>
          </div>
        </div>
      </li>
    </ul>

    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/60" @click="closeModal"></div>
      <div class="relative bg-zinc-950 border border-zinc-800 rounded-xl shadow-2xl p-6 w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
        <h2 class="text-white font-semibold text-base mb-4">{{ editTarget ? 'Edit lighting program' : 'New lighting program' }}</h2>
        <LightingProgramForm
          :form="form"
          :presets="presets"
          :actuators="lightActuators"
          :show-zone-select="false"
          :show-presets="!editTarget"
          @pick-preset="pickPreset"
          @clock-change="onClockChange"
        />
        <p v-if="modalError" class="mt-3 text-xs text-red-400">{{ modalError }}</p>
        <div class="flex justify-end gap-2 mt-5">
          <button type="button" class="px-4 py-2 text-sm text-zinc-400 border border-zinc-700 rounded-lg" @click="closeModal">Cancel</button>
          <button type="button" class="px-4 py-2 text-sm bg-gr33n-600 text-white rounded-lg disabled:opacity-50" :disabled="saving" @click="submitForm">
            {{ saving ? 'Saving…' : editTarget ? 'Save' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import api from '../api'
import EmptyStateHint from './EmptyStateHint.vue'
import LightingProgramForm from './LightingProgramForm.vue'
import { formatLightingProgramSummary } from '../lib/lightingDisplay.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
})

const emit = defineEmits(['updated'])

const programs = ref([])
const actuators = ref([])
const loading = ref(false)
const error = ref('')
const saving = ref(false)
const showModal = ref(false)
const editTarget = ref(null)
const modalError = ref('')

const presets = [
  { key: 'peas_22_2', label: '22/2 Peas', onHours: 22 },
  { key: 'veg_18_6', label: '18/6 Veg', onHours: 18 },
  { key: 'flower_12_12', label: '12/12 Flower', onHours: 12 },
  { key: 'seedling_16_8', label: '16/8 Seedling', onHours: 16 },
]

const blankForm = () => ({
  name: '',
  description: '',
  zoneId: props.zoneId,
  actuatorId: '',
  onHours: 18,
  lightsOnAt: '06:00',
  timezone: 'America/New_York',
  isActive: true,
  presetKey: '',
})
const form = ref(blankForm())

const lightActuators = computed(() =>
  actuators.value.filter((a) => !a.deleted_at && (a.actuator_type === 'light' || a.actuator_type === 'grow_light')),
)

function summary(prog) {
  return formatLightingProgramSummary(prog)
}

async function load() {
  if (!props.farmId) return
  loading.value = true
  error.value = ''
  try {
    const [p, a] = await Promise.all([
      api.get(`/farms/${props.farmId}/lighting-programs`),
      api.get(`/farms/${props.farmId}/actuators`),
    ])
    const all = p.data?.programs ?? p.data ?? []
    programs.value = all.filter((prog) => Number(prog.zone_id) === props.zoneId)
    actuators.value = a.data ?? []
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editTarget.value = null
  form.value = blankForm()
  if (programs.value.length) form.value.timezone = programs.value[0].timezone
  modalError.value = ''
  showModal.value = true
}

function openEdit(prog) {
  editTarget.value = prog
  form.value = {
    name: prog.name,
    description: prog.description ?? '',
    zoneId: prog.zone_id,
    actuatorId: prog.actuator_id,
    onHours: prog.on_hours,
    lightsOnAt: prog.lights_on_at,
    timezone: prog.timezone,
    isActive: prog.is_active,
    presetKey: '',
  }
  modalError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editTarget.value = null
}

function pickPreset(p) {
  form.value.presetKey = p.key
  form.value.onHours = p.onHours
  if (!form.value.name) form.value.name = p.label
}

function onClockChange({ lightsOnAt, onHours }) {
  form.value.lightsOnAt = lightsOnAt
  form.value.onHours = onHours
}

async function submitForm() {
  modalError.value = ''
  if (!form.value.name.trim()) { modalError.value = 'Name is required'; return }
  if (!form.value.actuatorId) { modalError.value = 'Actuator is required'; return }

  saving.value = true
  try {
    if (editTarget.value) {
      await api.patch(`/lighting-programs/${editTarget.value.id}`, {
        name: form.value.name,
        description: form.value.description || null,
        on_hours: form.value.onHours,
        off_hours: 24 - form.value.onHours,
        lights_on_at: form.value.lightsOnAt,
        timezone: form.value.timezone,
        is_active: form.value.isActive,
      })
    } else if (form.value.presetKey) {
      await api.post(`/farms/${props.farmId}/lighting-programs/from-preset`, {
        preset_key: form.value.presetKey,
        name: form.value.name,
        zone_id: props.zoneId,
        actuator_id: form.value.actuatorId,
        lights_on_at: form.value.lightsOnAt,
        timezone: form.value.timezone,
      })
    } else {
      await api.post(`/farms/${props.farmId}/lighting-programs`, {
        name: form.value.name,
        description: form.value.description || null,
        zone_id: props.zoneId,
        actuator_id: form.value.actuatorId,
        on_hours: form.value.onHours,
        off_hours: 24 - form.value.onHours,
        lights_on_at: form.value.lightsOnAt,
        timezone: form.value.timezone,
        is_active: form.value.isActive,
      })
    }
    closeModal()
    await load()
    emit('updated')
  } catch (e) {
    modalError.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

async function setActive(prog, active) {
  try {
    await api.post(`/lighting-programs/${prog.id}/${active ? 'activate' : 'deactivate'}`)
    await load()
    emit('updated')
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  }
}

onMounted(load)
watch(() => [props.zoneId, props.farmId], load)
</script>
