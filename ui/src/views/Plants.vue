<template>
  <div :class="embedded ? '' : 'p-6'">
    <div v-if="!embedded" class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Plants
        <HelpTip position="bottom">
          Plants are species or strain definitions (e.g. "OG Kush", "Basil - Genovese"). Create a plant here, then use Crop Cycles in Fertigation to track an individual grow of that plant in a specific zone.
        </HelpTip>
      </h1>
      <div class="flex items-center gap-3">
        <button
          @click="openCreate"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        >
          + New Plant
        </button>
        <button @click="refresh" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
        <span class="text-xs text-zinc-500">{{ plants.length }} plants</span>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading plants…</div>

    <EmptyStateHint
      v-else-if="!plants.length"
      reason="no_data"
      message="No plants yet — add a strain definition, then start a grow in a zone."
      action-label="Add your first plant"
      @action="openCreate"
    />

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="p in plants"
        :key="p.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-zinc-700 transition-colors"
      >
        <div class="flex items-start justify-between gap-2 mb-2">
          <div class="min-w-0">
            <p class="text-white text-sm font-medium truncate">{{ p.display_name }}</p>
            <p v-if="p.variety_or_cultivar" class="text-zinc-500 text-xs mt-0.5">{{ p.variety_or_cultivar }}</p>
          </div>
          <span class="text-lg shrink-0">🌱</span>
        </div>
        <div v-if="metaSummary(p.meta)" class="mb-3">
          <p class="text-zinc-600 text-xs line-clamp-2">{{ metaSummary(p.meta) }}</p>
        </div>
        <p class="text-zinc-600 text-[11px] mb-3">Added {{ formatDate(p.created_at) }}</p>
        <div
          v-if="cyclesForPlant(p.id).length"
          class="mb-3 rounded-lg border border-zinc-800 bg-zinc-950/50 px-3 py-2"
          data-test="plant-active-cycles"
        >
          <p class="text-[10px] uppercase tracking-widest text-zinc-500 mb-1.5">Grows</p>
          <ul class="space-y-1">
            <li v-for="c in cyclesForPlant(p.id)" :key="c.id" class="text-xs">
              <router-link
                :to="`/crop-cycles/${c.id}/summary`"
                class="text-green-500 hover:text-green-300"
              >
                {{ c.name || c.strain_or_variety || 'Grow' }}
              </router-link>
              <span class="text-zinc-500"> · {{ c.is_active ? 'active' : 'harvested' }}</span>
            </li>
          </ul>
        </div>
        <div class="flex items-center gap-3 border-t border-zinc-800 pt-2 flex-wrap">
          <button
            type="button"
            v-nav-hint="'/zones'"
            class="text-xs text-green-500 hover:text-green-300 font-medium"
            data-test="plant-start-grow"
            @click="openStartGrow(p)"
          >
            Start a grow
          </button>
          <button @click="openEdit(p)" class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
          <button @click="confirmDelete(p)" class="text-xs text-red-500 hover:text-red-400">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create / Edit modal -->
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="showModal = false"
    >
      <div class="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4">
        <h2 class="text-white font-semibold">{{ editing ? 'Edit Plant' : 'New Plant' }}</h2>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Display name *</label>
          <input
            v-model="form.display_name"
            type="text"
            required
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
          />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Variety / cultivar</label>
          <input
            v-model="form.variety_or_cultivar"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
          />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Metadata (JSON)</label>
          <textarea
            v-model="form.metaStr"
            rows="3"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white font-mono focus:outline-none focus:ring-1 focus:ring-green-600"
            placeholder='e.g. {"species": "Cannabis sativa", "photoperiod": "short-day"}'
          />
          <p v-if="metaError" class="text-red-400 text-[11px] mt-1">{{ metaError }}</p>
        </div>
        <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-1">
          <button
            @click="showModal = false"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >
            Cancel
          </button>
          <button
            @click="submitForm"
            :disabled="submitting || !form.display_name.trim()"
            class="px-4 py-1.5 text-xs rounded-lg bg-green-700 hover:bg-green-600 text-white font-medium disabled:opacity-40"
          >
            {{ submitting ? 'Saving…' : editing ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete confirmation -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="deleteTarget = null"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-sm space-y-4">
        <h3 class="text-white font-semibold">Delete Plant</h3>
        <p class="text-sm text-zinc-300">
          Delete <span class="text-white font-medium">{{ deleteTarget.display_name }}</span>?
        </p>
        <div class="flex justify-end gap-3 pt-2">
          <button
            @click="deleteTarget = null"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >
            Cancel
          </button>
          <button
            @click="doDelete"
            :disabled="submitting"
            class="px-3 py-1.5 text-xs rounded bg-red-600 hover:bg-red-500 text-white font-medium disabled:opacity-50"
          >
            {{ submitting ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>

    <StartGrowWizard
      :open="showStartGrowWizard"
      :farm-id="farmContext.farmId"
      :zones="zones"
      :programs="programs"
      :plants="plants"
      :initial-strain="startGrowStrain"
      :initial-plant-id="startGrowPlantId"
      @close="showStartGrowWizard = false"
      @created="onGrowStarted"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import StartGrowWizard from '../components/StartGrowWizard.vue'
import { strainFromPlant } from '../lib/growHub.js'

defineProps({
  embedded: { type: Boolean, default: false },
})

const store = useFarmStore()
const farmContext = useFarmContextStore()
const route = useRoute()
const router = useRouter()

const plants = ref([])
const zones = ref([])
const programs = ref([])
const showStartGrowWizard = ref(false)
const startGrowStrain = ref('')
const startGrowPlantId = ref(null)
const cropCycles = ref([])
const loading = ref(false)
const showModal = ref(false)
const editing = ref(null)
const submitting = ref(false)
const formError = ref('')
const metaError = ref('')
const deleteTarget = ref(null)
const form = ref(emptyForm())

function emptyForm() {
  return { display_name: '', variety_or_cultivar: '', metaStr: '{}' }
}

function openCreate() {
  editing.value = null
  form.value = emptyForm()
  formError.value = ''
  metaError.value = ''
  showModal.value = true
}

function openEdit(plant) {
  editing.value = plant
  form.value = {
    display_name: plant.display_name || '',
    variety_or_cultivar: plant.variety_or_cultivar || '',
    metaStr: typeof plant.meta === 'string' ? plant.meta : JSON.stringify(plant.meta ?? {}, null, 2),
  }
  formError.value = ''
  metaError.value = ''
  showModal.value = true
}

function confirmDelete(plant) {
  deleteTarget.value = plant
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    const [p, z, prog, cycles] = await Promise.all([
      store.loadPlants(fid),
      store.zones.length ? Promise.resolve(store.zones) : store.loadAll(fid).then(() => store.zones),
      store.loadFertigationPrograms(fid),
      store.loadCropCycles(fid),
    ])
    plants.value = p
    zones.value = z
    programs.value = prog
    cropCycles.value = cycles || []
  } finally {
    loading.value = false
  }
}

function cyclesForPlant(plantId) {
  return cropCycles.value.filter((c) => Number(c.plant_id) === Number(plantId))
}

function openStartGrow(plant) {
  startGrowStrain.value = strainFromPlant(plant)
  startGrowPlantId.value = plant.id
  showStartGrowWizard.value = true
}

async function onGrowStarted(cycle) {
  showStartGrowWizard.value = false
  if (cycle?.zone_id) {
    router.push({ path: `/zones/${cycle.zone_id}` })
  }
}

async function submitForm() {
  formError.value = ''
  metaError.value = ''
  const fid = farmContext.farmId
  if (!fid) { formError.value = 'No farm selected'; return }
  const name = form.value.display_name.trim()
  if (!name) return

  let meta
  try {
    meta = JSON.parse(form.value.metaStr || '{}')
  } catch {
    metaError.value = 'Invalid JSON'
    return
  }

  const payload = {
    display_name: name,
    variety_or_cultivar: form.value.variety_or_cultivar.trim() || null,
    meta,
  }

  submitting.value = true
  try {
    if (editing.value) {
      await store.updatePlant(editing.value.id, payload)
    } else {
      await store.createPlant(fid, payload)
    }
    showModal.value = false
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Failed to save'
  } finally {
    submitting.value = false
  }
}

async function doDelete() {
  submitting.value = true
  try {
    await store.deletePlant(deleteTarget.value.id)
    deleteTarget.value = null
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to delete'
  } finally {
    submitting.value = false
  }
}

function metaSummary(meta) {
  if (!meta || meta === '{}') return ''
  try {
    const obj = typeof meta === 'string' ? JSON.parse(meta) : meta
    const keys = Object.keys(obj)
    if (!keys.length) return ''
    return keys.slice(0, 4).map(k => `${k}: ${obj[k]}`).join(' · ')
  } catch {
    return ''
  }
}

function formatDate(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

watch(
  () => route.query.start_grow,
  (flag) => {
    if (!flag) return
    startGrowStrain.value = typeof route.query.strain === 'string' ? route.query.strain : ''
    showStartGrowWizard.value = true
  },
  { immediate: true },
)

onMounted(refresh)
watch(() => farmContext.farmId, refresh)
</script>
