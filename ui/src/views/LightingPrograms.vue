<template>
  <div :class="embedded ? 'px-4 sm:px-6 pb-6' : 'p-6'">
    <div v-if="!embedded" class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-xl font-semibold text-white">Lighting Programs</h1>
        <p class="text-xs text-zinc-500 mt-0.5">Photoperiod schedules — one program owns the ON/OFF pair + actuator</p>
      </div>
      <div class="flex items-center gap-3">
        <button
          class="px-3 py-1.5 text-xs rounded bg-gr33n-600 hover:bg-gr33n-500 text-white font-medium"
          @click="openCreate"
        >+ New Program</button>
        <button class="text-xs text-zinc-400 hover:text-zinc-200" @click="load">Refresh</button>
      </div>
    </div>

    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Lighting"
      back-to-zone-tab="light"
      :clear-route="{ path: '/lighting' }"
    />

    <!-- Error banner -->
    <div v-if="error" class="mb-4 px-4 py-2 text-sm text-red-400 bg-red-900/20 border border-red-800/40 rounded-lg">{{ error }}</div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading…</div>

    <div v-else-if="!filteredPrograms.length" class="text-zinc-500 text-sm py-12 text-center">
      <EmptyStateHint
        :reason="zoneContextId ? 'no_data' : 'no_data'"
        :message="zoneContextId ? 'No lighting programs for this zone yet.' : 'No lighting programs yet.'"
        action-label="Create program"
        :action-to="null"
        @action="openCreate"
      />
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="prog in filteredPrograms"
        :key="prog.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-white font-medium text-sm">{{ prog.name }}</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full font-semibold"
                :class="prog.is_active
                  ? 'bg-green-900/50 text-green-400 border border-green-800/40'
                  : 'bg-zinc-800 text-zinc-500 border border-zinc-700'">
                {{ prog.is_active ? 'Active' : 'Inactive' }}
              </span>
              <span v-if="presetLabel(prog.metadata)" class="text-[10px] px-1.5 py-0.5 rounded bg-blue-900/30 text-blue-400 border border-blue-800/30">
                {{ presetLabel(prog.metadata) }}
              </span>
            </div>
            <p v-if="prog.description" class="text-zinc-500 text-xs mt-1">{{ prog.description }}</p>
            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 mt-2 text-xs text-zinc-400">
              <span>💡 <strong class="text-zinc-300">{{ prog.on_hours }}h ON</strong> / {{ prog.off_hours }}h OFF</span>
              <span>⏰ ON at <strong class="text-zinc-300">{{ prog.lights_on_at }}</strong></span>
              <span>OFF at <strong class="text-zinc-300">{{ computeOffTime(prog.lights_on_at, prog.on_hours) }}</strong></span>
              <span>🌐 {{ prog.timezone }}</span>
            </div>
            <div class="flex flex-wrap gap-2 mt-2">
              <router-link
                v-if="prog.zone_id"
                v-nav-hint="'/zones'"
                :to="{ path: `/zones/${prog.zone_id}`, query: { tab: 'light' } }"
                class="text-[11px] px-1.5 py-0.5 rounded bg-blue-900/30 text-blue-400 border border-blue-800/30 hover:bg-blue-900/50"
              >Open zone →</router-link>
              <router-link
                v-if="prog.schedule_on_id"
                v-nav-hint="'/schedules'"
                :to="{ path: '/schedules' }"
                class="text-[11px] px-1.5 py-0.5 rounded bg-green-900/30 text-green-400 border border-green-800/30 hover:bg-green-900/50"
              >ON sch #{{ prog.schedule_on_id }}</router-link>
              <router-link
                v-if="prog.schedule_off_id"
                v-nav-hint="'/schedules'"
                :to="{ path: '/schedules' }"
                class="text-[11px] px-1.5 py-0.5 rounded bg-yellow-900/30 text-yellow-400 border border-yellow-800/30 hover:bg-yellow-900/50"
              >OFF sch #{{ prog.schedule_off_id }}</router-link>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-1.5 shrink-0 flex-wrap justify-end">
            <button
              @click="openEdit(prog)"
              class="px-2 py-1 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
              title="Edit"
            >Edit</button>
            <button
              v-if="prog.is_active"
              @click="setActive(prog, false)"
              class="px-2 py-1 text-xs rounded border border-yellow-800/50 text-yellow-400 hover:text-yellow-300"
            >Deactivate</button>
            <button
              v-else
              @click="setActive(prog, true)"
              class="px-2 py-1 text-xs rounded border border-green-800/50 text-green-400 hover:text-green-300"
            >Activate</button>
            <button
              @click="confirmDelete(prog)"
              class="px-2 py-1 text-xs rounded border border-red-900/50 text-red-400 hover:text-red-300"
            >Delete</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Create / Edit modal ───────────────────────────────────────────── -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/60" @click="closeModal"></div>
      <div class="relative bg-zinc-950 border border-zinc-800 rounded-xl shadow-2xl p-6 w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
        <h2 class="text-white font-semibold text-base mb-4">
          {{ editTarget ? 'Edit Lighting Program' : 'New Lighting Program' }}
        </h2>

        <!-- Preset quick-start (only for create) -->
        <div v-if="!editTarget" class="mb-4">
          <label class="block text-xs text-zinc-400 font-medium mb-1.5 uppercase tracking-wide">Start from preset</label>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="p in presets"
              :key="p.key"
              type="button"
              class="px-2.5 py-1 text-xs rounded-full border"
              :class="form.presetKey === p.key
                ? 'border-gr33n-500 bg-gr33n-900/40 text-gr33n-300'
                : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
              @click="pickPreset(p)"
            >{{ p.label }}</button>
          </div>
        </div>

        <div class="space-y-3">
          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1">Name *</label>
            <input v-model="form.name" type="text" placeholder="e.g. Veg Room 18/6"
              class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500" />
          </div>

          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1">Zone *</label>
            <select v-model="form.zoneId"
              class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500">
              <option value="">— select zone —</option>
              <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1">Grow Light Actuator *</label>
            <select v-model="form.actuatorId"
              class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500">
              <option value="">— select actuator —</option>
              <option v-for="a in lightActuators" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </div>

          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1.5">Photoperiod</label>
            <PhotoperiodClockEditor
              v-model:model-lights-on-at="form.lightsOnAt"
              v-model:model-on-hours="form.onHours"
              :timezone="form.timezone"
              @change="onClockChange"
            />
          </div>

          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1">Timezone</label>
            <input v-model="form.timezone" type="text" placeholder="UTC or America/New_York"
              class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500" />
          </div>

          <div>
            <label class="block text-xs text-zinc-400 font-medium mb-1">Description</label>
            <textarea v-model="form.description" rows="2" placeholder="Optional notes"
              class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500 resize-none"></textarea>
          </div>

          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="form.isActive" class="accent-gr33n-500" />
            <span class="text-sm text-zinc-300">Active (enable schedules immediately)</span>
          </label>
        </div>

        <p v-if="modalError" class="mt-3 text-xs text-red-400">{{ modalError }}</p>

        <div class="flex justify-end gap-2 mt-5">
          <button @click="closeModal" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200 border border-zinc-700 rounded-lg">Cancel</button>
          <button
            @click="submitForm"
            :disabled="saving"
            class="px-4 py-2 text-sm bg-gr33n-600 hover:bg-gr33n-500 text-white rounded-lg disabled:opacity-50"
          >{{ saving ? 'Saving…' : editTarget ? 'Save Changes' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Delete confirmation -->
    <div v-if="deleteTarget" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/60" @click="deleteTarget = null"></div>
      <div class="relative bg-zinc-950 border border-zinc-800 rounded-xl shadow-2xl p-6 w-full max-w-sm mx-4">
        <h2 class="text-white font-semibold mb-2">Delete Lighting Program?</h2>
        <p class="text-zinc-400 text-sm mb-4">
          This will permanently delete <strong class="text-zinc-200">{{ deleteTarget.name }}</strong>
          and its paired ON/OFF schedules.
        </p>
        <div class="flex justify-end gap-2">
          <button @click="deleteTarget = null" class="px-4 py-2 text-sm text-zinc-400 border border-zinc-700 rounded-lg">Cancel</button>
          <button @click="doDelete" :disabled="saving" class="px-4 py-2 text-sm bg-red-700 hover:bg-red-600 text-white rounded-lg disabled:opacity-50">
            {{ saving ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api/index.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import PhotoperiodClockEditor from '../components/PhotoperiodClockEditor.vue'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import { parseZoneIdQuery } from '../lib/zoneContext.js'
import { computeOffTime } from '../lib/lightingDisplay.js'

const route = useRoute()
const farmContext = useFarmContextStore()

defineProps({
  embedded: { type: Boolean, default: false },
})

// ── farm context ──────────────────────────────────────────────────────────────
const farmId = computed(() => farmContext.farmId)

// ── state ─────────────────────────────────────────────────────────────────────
const programs = ref([])
const zones    = ref([])
const actuators= ref([])
const loading  = ref(false)
const error    = ref('')
const saving   = ref(false)

const showModal   = ref(false)
const editTarget  = ref(null)
const deleteTarget= ref(null)
const modalError  = ref('')

const PRESET_CHIPS = [
  { key: 'peas_22_2',     label: '22/2 Peas',    onHours: 22 },
  { key: 'veg_18_6',      label: '18/6 Veg',     onHours: 18 },
  { key: 'flower_12_12',  label: '12/12 Flower',  onHours: 12 },
  { key: 'seedling_16_8', label: '16/8 Seedling', onHours: 16 },
]
const presets = PRESET_CHIPS

const blankForm = () => ({
  name: '',
  description: '',
  zoneId: '',
  actuatorId: '',
  onHours: 18,
  lightsOnAt: '06:00',
  timezone: 'America/New_York',
  isActive: true,
  presetKey: '',
})
const form = ref(blankForm())

// ── computed ──────────────────────────────────────────────────────────────────
const lightActuators = computed(() =>
  actuators.value.filter(a => !a.deleted_at && (a.actuator_type === 'light' || a.actuator_type === 'grow_light'))
)

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

function zoneName(zoneId) {
  return zones.value.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

const filteredPrograms = computed(() => {
  if (!zoneContextId.value) return programs.value
  return programs.value.filter((p) => Number(p.zone_id) === zoneContextId.value)
})

// ── load ──────────────────────────────────────────────────────────────────────
async function load() {
  if (!farmId.value) return
  loading.value = true
  error.value = ''
  try {
    const [p, z, a] = await Promise.all([
      api.get(`/farms/${farmId.value}/lighting-programs`),
      api.get(`/farms/${farmId.value}/zones`),
      api.get(`/farms/${farmId.value}/actuators`),
    ])
    programs.value = p.data ?? []
    zones.value    = z.data ?? []
    actuators.value= a.data ?? []
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    loading.value = false
  }
}

onMounted(load)

watch(() => route.query.zone_id, () => {})

// ── helpers ───────────────────────────────────────────────────────────────────
function presetLabel(metadata) {
  try {
    const meta = typeof metadata === 'string' ? JSON.parse(metadata) : metadata
    if (meta?.preset_key) {
      const chip = PRESET_CHIPS.find(p => p.key === meta.preset_key)
      return chip ? chip.label : meta.preset_key
    }
  } catch {}
  return ''
}

function onClockChange({ lightsOnAt, onHours }) {
  form.value.lightsOnAt = lightsOnAt
  form.value.onHours    = onHours
}

// ── modal ─────────────────────────────────────────────────────────────────────
function openCreate() {
  editTarget.value = null
  form.value = blankForm()
  // Default timezone from first program or fallback.
  if (programs.value.length) form.value.timezone = programs.value[0].timezone
  modalError.value = ''
  showModal.value = true
}

function openEdit(prog) {
  editTarget.value = prog
  form.value = {
    name:        prog.name,
    description: prog.description ?? '',
    zoneId:      prog.zone_id,
    actuatorId:  prog.actuator_id,
    onHours:     prog.on_hours,
    lightsOnAt:  prog.lights_on_at,
    timezone:    prog.timezone,
    isActive:    prog.is_active,
    presetKey:   '',
  }
  modalError.value = ''
  showModal.value = true
}

function closeModal() {
  showModal.value  = false
  editTarget.value = null
}

function pickPreset(p) {
  form.value.presetKey = p.key
  form.value.onHours   = p.onHours
  if (!form.value.name) form.value.name = p.label
}

async function submitForm() {
  modalError.value = ''
  if (!form.value.name.trim()) { modalError.value = 'Name is required'; return }
  if (!form.value.zoneId)      { modalError.value = 'Zone is required'; return }
  if (!form.value.actuatorId)  { modalError.value = 'Actuator is required'; return }

  saving.value = true
  try {
    if (editTarget.value) {
      await api.patch(`/lighting-programs/${editTarget.value.id}`, {
        name:        form.value.name,
        description: form.value.description || null,
        on_hours:    form.value.onHours,
        off_hours:   24 - form.value.onHours,
        lights_on_at:form.value.lightsOnAt,
        timezone:    form.value.timezone,
        is_active:   form.value.isActive,
      })
    } else if (form.value.presetKey) {
      await api.post(`/farms/${farmId.value}/lighting-programs/from-preset`, {
        preset_key:   form.value.presetKey,
        name:         form.value.name,
        zone_id:      form.value.zoneId,
        actuator_id:  form.value.actuatorId,
        lights_on_at: form.value.lightsOnAt,
        timezone:     form.value.timezone,
      })
    } else {
      await api.post(`/farms/${farmId.value}/lighting-programs`, {
        name:         form.value.name,
        description:  form.value.description || null,
        zone_id:      form.value.zoneId,
        actuator_id:  form.value.actuatorId,
        on_hours:     form.value.onHours,
        off_hours:    24 - form.value.onHours,
        lights_on_at: form.value.lightsOnAt,
        timezone:     form.value.timezone,
        is_active:    form.value.isActive,
      })
    }
    closeModal()
    await load()
  } catch (e) {
    modalError.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}

// ── activate / deactivate ─────────────────────────────────────────────────────
async function setActive(prog, active) {
  try {
    await api.post(`/lighting-programs/${prog.id}/${active ? 'activate' : 'deactivate'}`)
    await load()
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  }
}

// ── delete ────────────────────────────────────────────────────────────────────
function confirmDelete(prog) {
  deleteTarget.value = prog
}

async function doDelete() {
  if (!deleteTarget.value) return
  saving.value = true
  try {
    await api.delete(`/lighting-programs/${deleteTarget.value.id}`)
    deleteTarget.value = null
    await load()
  } catch (e) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}
</script>
