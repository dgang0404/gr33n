<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold text-white">Fertigation</h1>
      <button @click="refresh" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1">
      <button
        v-for="t in tabs" :key="t.id"
        @click="activeTab = t.id"
        class="px-4 py-2 text-sm rounded-md transition-colors"
        :class="activeTab === t.id
          ? 'bg-zinc-800 text-white font-medium'
          : 'text-zinc-400 hover:text-zinc-200'"
      >{{ t.label }}</button>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading…</div>

    <!-- Reservoirs -->
    <template v-else-if="activeTab === 'reservoirs'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ reservoirs.length }} reservoir(s)</p>
        <button @click="showReservoirForm = !showReservoirForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showReservoirForm ? 'Cancel' : '+ Add Reservoir' }}
        </button>
      </div>

      <!-- Create form -->
      <form v-if="showReservoirForm" @submit.prevent="submitReservoir"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input v-model="resForm.name" placeholder="Name" required
          class="input-field" />
        <select v-model="resForm.status" class="input-field">
          <option value="ready">Ready</option>
          <option value="mixing">Mixing</option>
          <option value="needs_top_up">Needs Top-Up</option>
          <option value="needs_flush">Needs Flush</option>
          <option value="flushing">Flushing</option>
          <option value="offline">Offline</option>
          <option value="empty">Empty</option>
        </select>
        <input v-model.number="resForm.capacity_liters" type="number" step="0.1" min="0"
          placeholder="Capacity (L)" required class="input-field" />
        <input v-model.number="resForm.current_volume_liters" type="number" step="0.1" min="0"
          placeholder="Current Volume (L)" required class="input-field" />
        <select v-model="resForm.zone_id" class="input-field">
          <option :value="null">No zone</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
          {{ saving ? 'Saving…' : 'Create Reservoir' }}
        </button>
      </form>

      <!-- Reservoir cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="r in reservoirs" :key="r.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-2">
          <div class="flex items-center justify-between">
            <p class="text-white text-sm font-medium">{{ r.name }}</p>
            <span class="text-xs px-2 py-0.5 rounded-full capitalize"
              :class="r.status === 'ready' ? 'bg-green-900/50 text-green-300' : r.status === 'offline' || r.status === 'empty' ? 'bg-red-900/50 text-red-300' : 'bg-yellow-900/50 text-yellow-300'">
              {{ r.status?.replace(/_/g, ' ') }}
            </span>
          </div>
          <div class="flex items-end gap-1">
            <span class="text-white text-lg font-mono">{{ r.current_volume_liters || 0 }}</span>
            <span class="text-zinc-500 text-sm mb-0.5">/ {{ r.capacity_liters || 0 }} L</span>
          </div>
          <div class="w-full bg-zinc-800 rounded-full h-2">
            <div class="bg-blue-500 h-2 rounded-full transition-all"
              :style="{ width: fillPct(r) + '%' }" />
          </div>
          <p v-if="r.last_ec_mscm" class="text-zinc-500 text-xs">
            EC {{ r.last_ec_mscm }} mS/cm · pH {{ r.last_ph || '—' }}
          </p>
          <p class="text-zinc-600 text-xs">{{ zoneLabel(r.zone_id) }}</p>
        </div>
      </div>
      <p v-if="!reservoirs.length" class="text-zinc-500 text-sm">No reservoirs configured yet.</p>
    </template>

    <!-- EC Targets -->
    <template v-else-if="activeTab === 'ec-targets'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ ecTargets.length }} target(s)</p>
        <button @click="showEcForm = !showEcForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showEcForm ? 'Cancel' : '+ Add EC Target' }}
        </button>
      </div>

      <form v-if="showEcForm" @submit.prevent="submitEcTarget"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <select v-model="ecForm.growth_stage" required class="input-field">
          <option value="" disabled>Growth stage</option>
          <option v-for="gs in growthStages" :key="gs" :value="gs">{{ gs }}</option>
        </select>
        <select v-model="ecForm.zone_id" class="input-field">
          <option :value="null">All zones</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <input v-model.number="ecForm.ec_min_mscm" type="number" step="0.01" placeholder="EC min (mS/cm)"
          required class="input-field" />
        <input v-model.number="ecForm.ec_max_mscm" type="number" step="0.01" placeholder="EC max (mS/cm)"
          required class="input-field" />
        <input v-model.number="ecForm.ph_min" type="number" step="0.1" placeholder="pH min"
          required class="input-field" />
        <input v-model.number="ecForm.ph_max" type="number" step="0.1" placeholder="pH max"
          required class="input-field" />
        <input v-model="ecForm.notes" placeholder="Notes (optional)" class="input-field sm:col-span-2 lg:col-span-2" />
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
          {{ saving ? 'Saving…' : 'Create Target' }}
        </button>
      </form>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-zinc-400 border-b border-zinc-800">
            <tr>
              <th class="py-2 pr-4">Stage</th>
              <th class="py-2 pr-4">Zone</th>
              <th class="py-2 pr-4">EC Range</th>
              <th class="py-2 pr-4">pH Range</th>
              <th class="py-2">Notes</th>
            </tr>
          </thead>
          <tbody class="text-zinc-300">
            <tr v-for="t in ecTargets" :key="t.id" class="border-b border-zinc-800/50">
              <td class="py-2 pr-4 capitalize">{{ t.growth_stage }}</td>
              <td class="py-2 pr-4">{{ zoneLabel(t.zone_id) }}</td>
              <td class="py-2 pr-4 font-mono">{{ t.ec_min_mscm }}–{{ t.ec_max_mscm }} mS/cm</td>
              <td class="py-2 pr-4 font-mono">{{ t.ph_min }}–{{ t.ph_max }}</td>
              <td class="py-2 text-zinc-500 truncate max-w-48">{{ t.notes || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-if="!ecTargets.length" class="text-zinc-500 text-sm">No EC targets configured yet.</p>
    </template>

    <!-- Programs -->
    <template v-else-if="activeTab === 'programs'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ programs.length }} program(s)</p>
        <button @click="showProgramForm = !showProgramForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showProgramForm ? 'Cancel' : '+ Add Program' }}
        </button>
      </div>

      <form v-if="showProgramForm" @submit.prevent="submitProgram"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input v-model="progForm.name" placeholder="Program name" required class="input-field" />
        <select v-model="progForm.target_zone_id" class="input-field">
          <option :value="null">No target zone</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <select v-model="progForm.reservoir_id" class="input-field">
          <option :value="null">No reservoir</option>
          <option v-for="r in reservoirs" :key="r.id" :value="r.id">{{ r.name }}</option>
        </select>
        <select v-model="progForm.ec_target_id" class="input-field">
          <option :value="null">No EC target</option>
          <option v-for="t in ecTargets" :key="t.id" :value="t.id">{{ t.growth_stage }} ({{ t.ec_min_mscm }}–{{ t.ec_max_mscm }})</option>
        </select>
        <select v-model="progForm.application_recipe_id" class="input-field">
          <option :value="null">No NF recipe</option>
          <option v-for="r in nfRecipes" :key="r.id" :value="r.id">{{ r.name }}</option>
        </select>
        <input v-model.number="progForm.total_volume_liters" type="number" step="0.1" placeholder="Total volume (L)"
          required class="input-field" />
        <label class="flex items-center gap-2 text-zinc-300 text-sm">
          <input type="checkbox" v-model="progForm.is_active" class="rounded bg-zinc-800 border-zinc-700" />
          Active
        </label>
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2">
          {{ saving ? 'Saving…' : 'Create Program' }}
        </button>
      </form>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div v-for="p in programs" :key="p.id"
          class="bg-zinc-900 border rounded-xl p-4 space-y-2"
          :class="p.is_active ? 'border-green-800/70' : 'border-zinc-800'">
          <div class="flex items-center justify-between">
            <p class="text-white text-sm font-medium">{{ p.name }}</p>
            <span class="text-xs px-2 py-0.5 rounded-full"
              :class="p.is_active ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-400'">
              {{ p.is_active ? 'Active' : 'Inactive' }}
            </span>
          </div>
          <p class="text-zinc-400 text-xs">{{ zoneLabel(p.target_zone_id) }} · {{ p.total_volume_liters || 0 }}L</p>
          <p v-if="p.description" class="text-zinc-500 text-xs">{{ p.description }}</p>
        </div>
      </div>
      <p v-if="!programs.length" class="text-zinc-500 text-sm">No programs configured yet.</p>
    </template>

    <!-- Crop cycles -->
    <template v-else-if="activeTab === 'crop-cycles'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ cropCycles.length }} cycle(s)</p>
        <button @click="toggleCycleForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showCycleForm ? 'Cancel' : '+ New cycle' }}
        </button>
      </div>

      <form v-if="showCycleForm" @submit.prevent="submitCycle"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <select v-model.number="cycleForm.zone_id" required class="input-field">
          <option value="" disabled>Zone</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <input v-model="cycleForm.name" placeholder="Cycle name" required class="input-field" />
        <input v-model="cycleForm.strain_or_variety" placeholder="Strain / variety" class="input-field" />
        <select v-model="cycleForm.current_stage" class="input-field">
          <option v-for="gs in growthStages" :key="gs" :value="gs">{{ gs.replace(/_/g, ' ') }}</option>
        </select>
        <input v-model="cycleForm.started_at" type="date" required class="input-field" />
        <select v-model.number="cycleForm.primary_program_id" class="input-field">
          <option :value="null">No primary program</option>
          <option v-for="p in programs" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <textarea v-model="cycleForm.cycle_notes" placeholder="Notes" class="input-field sm:col-span-2" rows="2" />
        <label class="flex items-center gap-2 text-zinc-300 text-sm sm:col-span-2">
          <input type="checkbox" v-model="cycleForm.is_active" class="rounded bg-zinc-800 border-zinc-700" />
          Active
        </label>
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2">
          {{ saving ? 'Saving…' : (editCycle ? 'Update cycle' : 'Create cycle') }}
        </button>
      </form>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div v-for="c in cropCycles" :key="c.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
          <div class="flex items-start justify-between gap-2">
            <div>
              <p class="text-white font-medium">{{ c.name }}</p>
              <p class="text-zinc-500 text-xs mt-1">{{ zoneLabel(c.zone_id) }}
                <span v-if="c.strain_or_variety"> · {{ c.strain_or_variety }}</span>
              </p>
            </div>
            <span class="text-xs px-2 py-0.5 rounded-full capitalize shrink-0"
              :class="c.is_active ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-500'">
              {{ c.is_active ? 'active' : 'inactive' }}
            </span>
          </div>
          <p class="text-xs text-zinc-400">Stage: <span class="text-zinc-200 capitalize">{{ cycleStage(c) }}</span></p>
          <p class="text-xs text-zinc-500">Started {{ isoDate(c.started_at) }}</p>
          <p v-if="c.primary_program_id" class="text-xs text-zinc-500">
            Program: {{ programName(c.primary_program_id) }}
          </p>
          <div v-if="cycleStage(c) === 'harvest'" class="text-xs text-zinc-400 space-y-1">
            <p v-if="c.yield_grams != null">Yield: {{ c.yield_grams }} g</p>
            <p v-if="c.yield_notes">{{ c.yield_notes }}</p>
          </div>
          <div class="flex flex-wrap gap-2 items-center">
            <select v-model="stageDraft[c.id]" class="input-field text-xs py-1 max-w-[10rem]">
              <option v-for="gs in growthStages" :key="gs" :value="gs">{{ gs.replace(/_/g, ' ') }}</option>
            </select>
            <button type="button" @click="patchStage(c)" :disabled="saving"
              class="text-xs px-2 py-1 rounded bg-zinc-800 text-zinc-300 hover:bg-zinc-700">Set stage</button>
            <button type="button" @click="startEditCycle(c)" class="text-xs text-zinc-500 hover:text-zinc-300">Edit</button>
            <button type="button" @click="deleteCycle(c)" class="text-xs text-red-500 hover:text-red-400">Deactivate</button>
          </div>
        </div>
      </div>
      <p v-if="!cropCycles.length" class="text-zinc-500 text-sm">No crop cycles yet.</p>
    </template>

    <!-- Events -->
    <template v-else-if="activeTab === 'events'">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
        <div class="flex flex-wrap items-center gap-3">
          <p class="text-zinc-400 text-sm">{{ fertigationEvents.length }} event(s)</p>
          <label class="flex items-center gap-2 text-xs text-zinc-500">
            <span>Filter by crop cycle</span>
            <select v-model="eventCropFilter" @change="reloadEventsOnly" class="input-field py-1 text-xs max-w-[14rem]">
              <option value="">All cycles</option>
              <option v-for="c in cropCycles" :key="c.id" :value="String(c.id)">{{ c.name }} ({{ zoneLabel(c.zone_id) }})</option>
            </select>
          </label>
        </div>
        <button @click="showEventForm = !showEventForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showEventForm ? 'Cancel' : '+ Log Event' }}
        </button>
      </div>

      <form v-if="showEventForm" @submit.prevent="submitEvent"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <select v-model.number="evForm.zone_id" required class="input-field">
          <option value="" disabled>Zone</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
        <select v-model="evForm.crop_cycle_id" class="input-field">
          <option :value="null">No crop cycle</option>
          <option v-for="c in cyclesForEventZone" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select v-model="evForm.program_id" class="input-field">
          <option :value="null">No program</option>
          <option v-for="p in programs" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
        <input v-model.number="evForm.volume_applied_liters" type="number" step="0.1" placeholder="Volume (L)"
          required class="input-field" />
        <input v-model.number="evForm.ec_before_mscm" type="number" step="0.01" placeholder="EC before" class="input-field" />
        <input v-model.number="evForm.ec_after_mscm" type="number" step="0.01" placeholder="EC after" class="input-field" />
        <input v-model.number="evForm.ph_before" type="number" step="0.1" placeholder="pH before" class="input-field" />
        <input v-model.number="evForm.ph_after" type="number" step="0.1" placeholder="pH after" class="input-field" />
        <input v-model="evForm.notes" placeholder="Notes (optional)" class="input-field sm:col-span-2" />
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
          {{ saving ? 'Saving…' : 'Log Event' }}
        </button>
      </form>

      <div class="overflow-x-auto">
        <table class="w-full text-sm text-left">
          <thead class="text-xs text-zinc-400 border-b border-zinc-800">
            <tr>
              <th class="py-2 pr-4">Time</th>
              <th class="py-2 pr-4">Zone</th>
              <th class="py-2 pr-4">Crop cycle</th>
              <th class="py-2 pr-4">Volume</th>
              <th class="py-2 pr-4">EC Before→After</th>
              <th class="py-2 pr-4">pH Before→After</th>
              <th class="py-2 pr-4">Trigger</th>
              <th class="py-2">Notes</th>
            </tr>
          </thead>
          <tbody class="text-zinc-300">
            <tr v-for="e in sortedEvents" :key="e.id" class="border-b border-zinc-800/50">
              <td class="py-2 pr-4 whitespace-nowrap">{{ formatDate(e.applied_at) }}</td>
              <td class="py-2 pr-4">{{ zoneLabel(e.zone_id) }}</td>
              <td class="py-2 pr-4 text-zinc-500 text-xs">{{ cycleLabel(e.crop_cycle_id) }}</td>
              <td class="py-2 pr-4 font-mono">{{ e.volume_applied_liters || 0 }}L</td>
              <td class="py-2 pr-4 font-mono">{{ e.ec_before_mscm || '—' }} → {{ e.ec_after_mscm || '—' }}</td>
              <td class="py-2 pr-4 font-mono">{{ e.ph_before || '—' }} → {{ e.ph_after || '—' }}</td>
              <td class="py-2 pr-4 text-xs capitalize">{{ (e.trigger_source || 'manual').replace(/_/g, ' ') }}</td>
              <td class="py-2 text-zinc-500 truncate max-w-48">{{ e.notes || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-if="!fertigationEvents.length" class="text-zinc-500 text-sm">No fertigation events recorded yet.</p>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('reservoirs')

const tabs = [
  { id: 'reservoirs', label: 'Reservoirs' },
  { id: 'ec-targets', label: 'EC Targets' },
  { id: 'programs', label: 'Programs' },
  { id: 'crop-cycles', label: 'Crop Cycles' },
  { id: 'events', label: 'Events' },
]

const growthStages = ['clone', 'seedling', 'early_veg', 'late_veg', 'transition', 'early_flower', 'mid_flower', 'late_flower', 'flush', 'harvest', 'dry_cure']

const zones = computed(() => store.zones)
const farmId = computed(() => farmContext.farmId)

const reservoirs = ref([])
const ecTargets = ref([])
const programs = ref([])
const nfRecipes = ref([])
const cropCycles = ref([])
const fertigationEvents = ref([])
const eventCropFilter = ref('')
const showCycleForm = ref(false)
const editCycle = ref(null)
const stageDraft = reactive({})
const cycleForm = ref({
  zone_id: '',
  name: '',
  strain_or_variety: '',
  current_stage: 'seedling',
  started_at: new Date().toISOString().slice(0, 10),
  is_active: true,
  cycle_notes: '',
  primary_program_id: null,
  harvested_at: '',
  yield_grams: null,
  yield_notes: '',
})

const sortedEvents = computed(() =>
  [...fertigationEvents.value].sort((a, b) => new Date(b.applied_at) - new Date(a.applied_at))
)

const showReservoirForm = ref(false)
const showEcForm = ref(false)
const showProgramForm = ref(false)
const showEventForm = ref(false)

const resForm = ref({ name: '', status: 'ready', capacity_liters: 0, current_volume_liters: 0, zone_id: null })
const ecForm = ref({ growth_stage: '', zone_id: null, ec_min_mscm: 0, ec_max_mscm: 0, ph_min: 0, ph_max: 0, notes: '' })
const progForm = ref({
  name: '',
  application_recipe_id: null,
  target_zone_id: null,
  reservoir_id: null,
  ec_target_id: null,
  total_volume_liters: 0,
  is_active: false,
  ec_trigger_low: 0,
  ph_trigger_low: 0,
  ph_trigger_high: 0,
})
const evForm = ref({
  zone_id: '',
  crop_cycle_id: null,
  program_id: null,
  volume_applied_liters: 0,
  ec_before_mscm: 0,
  ec_after_mscm: 0,
  ph_before: 0,
  ph_after: 0,
  notes: '',
  trigger_source: 'manual',
})

const cyclesForEventZone = computed(() => {
  const zid = evForm.value.zone_id
  if (!zid) return cropCycles.value
  return cropCycles.value.filter((c) => Number(c.zone_id) === Number(zid))
})

watch(
  () => evForm.value.zone_id,
  () => {
    evForm.value.crop_cycle_id = null
  }
)

async function refresh() {
  loading.value = true
  try {
    if (!store.zones.length && farmId.value) await store.loadAll(farmId.value)
    const fid = farmId.value
    const cropQ = eventCropFilter.value ? Number(eventCropFilter.value) : undefined
    const [r, ec, p, ev, cc, recipes] = await Promise.all([
      store.loadReservoirs(fid),
      store.loadEcTargets(fid),
      store.loadFertigationPrograms(fid),
      store.loadFertigationEvents(fid, { cropCycleId: cropQ }),
      store.loadCropCycles(fid),
      store.loadRecipes(fid),
    ])
    reservoirs.value = r
    ecTargets.value = ec
    programs.value = p
    fertigationEvents.value = ev
    cropCycles.value = cc
    nfRecipes.value = recipes
    for (const c of cropCycles.value) {
      if (stageDraft[c.id] == null) stageDraft[c.id] = cycleStageRaw(c)
    }
  } finally { loading.value = false }
}

async function reloadEventsOnly() {
  const fid = farmId.value
  if (!fid) return
  const cropQ = eventCropFilter.value ? Number(eventCropFilter.value) : undefined
  fertigationEvents.value = await store.loadFertigationEvents(fid, { cropCycleId: cropQ })
}

function cycleLabel(id) {
  if (id == null) return '—'
  const c = cropCycles.value.find((x) => x.id === id)
  return c ? c.name : `#${id}`
}

function cycleStageRaw(c) {
  const s = c.current_stage
  if (!s) return 'seedling'
  if (typeof s === 'object' && s.valid) return s.gr33nfertigation_growth_stage_enum
  if (typeof s === 'string') return s
  return 'seedling'
}

function cycleStage(c) {
  return String(cycleStageRaw(c)).replace(/_/g, ' ')
}

function isoDate(d) {
  if (!d) return '—'
  if (typeof d === 'string') return d.slice(0, 10)
  if (d.Time) return String(d.Time).slice(0, 10)
  return '—'
}

function programName(id) {
  return programs.value.find(p => p.id === id)?.name ?? `#${id}`
}

function emptyCycleForm() {
  return {
    zone_id: '',
    name: '',
    strain_or_variety: '',
    current_stage: 'seedling',
    started_at: new Date().toISOString().slice(0, 10),
    is_active: true,
    cycle_notes: '',
    primary_program_id: null,
    harvested_at: '',
    yield_grams: null,
    yield_notes: '',
  }
}

function toggleCycleForm() {
  showCycleForm.value = !showCycleForm.value
  if (!showCycleForm.value) {
    editCycle.value = null
    cycleForm.value = emptyCycleForm()
  }
}

function startEditCycle(c) {
  editCycle.value = c
  showCycleForm.value = true
  cycleForm.value = {
    zone_id: c.zone_id,
    name: c.name,
    strain_or_variety: c.strain_or_variety || '',
    current_stage: cycleStageRaw(c),
    started_at: isoDate(c.started_at) === '—' ? new Date().toISOString().slice(0, 10) : isoDate(c.started_at),
    is_active: !!c.is_active,
    cycle_notes: c.cycle_notes || '',
    primary_program_id: c.primary_program_id ?? null,
    harvested_at: c.harvested_at ? isoDate(c.harvested_at) : '',
    yield_grams: c.yield_grams != null ? Number(c.yield_grams) : null,
    yield_notes: c.yield_notes || '',
  }
}

async function submitCycle() {
  saving.value = true
  try {
    const fid = farmId.value
    const base = {
      name: cycleForm.value.name.trim(),
      strain_or_variety: cycleForm.value.strain_or_variety?.trim() || undefined,
      zone_id: Number(cycleForm.value.zone_id),
      is_active: cycleForm.value.is_active,
      cycle_notes: cycleForm.value.cycle_notes?.trim() || undefined,
      primary_program_id: cycleForm.value.primary_program_id,
    }
    if (editCycle.value) {
      const payload = {
        ...base,
        harvested_at: cycleForm.value.harvested_at || undefined,
        yield_grams: cycleForm.value.yield_grams != null ? cycleForm.value.yield_grams : undefined,
        yield_notes: cycleForm.value.yield_notes?.trim() || undefined,
      }
      await store.updateCropCycle(editCycle.value.id, payload)
    } else {
      await store.createCropCycle(fid, {
        zone_id: base.zone_id,
        name: base.name,
        strain_or_variety: base.strain_or_variety,
        current_stage: cycleForm.value.current_stage,
        started_at: cycleForm.value.started_at,
        is_active: base.is_active,
        cycle_notes: base.cycle_notes,
        primary_program_id: base.primary_program_id,
      })
    }
    showCycleForm.value = false
    editCycle.value = null
    cycleForm.value = emptyCycleForm()
    cropCycles.value = await store.loadCropCycles(fid)
  } finally { saving.value = false }
}

async function patchStage(c) {
  const next = stageDraft[c.id]
  if (!next) return
  saving.value = true
  try {
    await store.updateCropCycleStage(c.id, next)
    cropCycles.value = await store.loadCropCycles(farmId.value)
  } finally { saving.value = false }
}

async function deleteCycle(c) {
  if (!confirm(`Deactivate cycle "${c.name}"?`)) return
  await store.deleteCropCycle(c.id)
  cropCycles.value = await store.loadCropCycles(farmId.value)
}

onMounted(() => {
  if (route.query.tab === 'crop-cycles') activeTab.value = 'crop-cycles'
  if (route.query.recipe) {
    progForm.value.application_recipe_id = Number(route.query.recipe)
    activeTab.value = 'programs'
    showProgramForm.value = true
  }
  refresh()
})

async function submitReservoir() {
  saving.value = true
  try {
    await store.createReservoir(farmId.value, resForm.value)
    showReservoirForm.value = false
    resForm.value = { name: '', status: 'ready', capacity_liters: 0, current_volume_liters: 0, zone_id: null }
    reservoirs.value = await store.loadReservoirs(farmId.value)
  } finally { saving.value = false }
}

async function submitEcTarget() {
  saving.value = true
  try {
    await store.createEcTarget(farmId.value, ecForm.value)
    showEcForm.value = false
    ecForm.value = { growth_stage: '', zone_id: null, ec_min_mscm: 0, ec_max_mscm: 0, ph_min: 0, ph_max: 0, notes: '' }
    ecTargets.value = await store.loadEcTargets(farmId.value)
  } finally { saving.value = false }
}

async function submitProgram() {
  saving.value = true
  try {
    await store.createProgram(farmId.value, progForm.value)
    showProgramForm.value = false
    progForm.value = {
      name: '',
      application_recipe_id: null,
      target_zone_id: null,
      reservoir_id: null,
      ec_target_id: null,
      total_volume_liters: 0,
      is_active: false,
      ec_trigger_low: 0,
      ph_trigger_low: 0,
      ph_trigger_high: 0,
    }
    programs.value = await store.loadFertigationPrograms(farmId.value)
  } finally { saving.value = false }
}

async function submitEvent() {
  saving.value = true
  try {
    const payload = { ...evForm.value }
    if (!payload.crop_cycle_id) delete payload.crop_cycle_id
    await store.createFertigationEvent(farmId.value, payload)
    showEventForm.value = false
    evForm.value = {
      zone_id: '',
      crop_cycle_id: null,
      program_id: null,
      volume_applied_liters: 0,
      ec_before_mscm: 0,
      ec_after_mscm: 0,
      ph_before: 0,
      ph_after: 0,
      notes: '',
      trigger_source: 'manual',
    }
    await reloadEventsOnly()
  } finally { saving.value = false }
}

function zoneLabel(id) {
  if (!id) return 'All zones'
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}

function fillPct(r) {
  if (!r.capacity_liters || r.capacity_liters <= 0) return 0
  return Math.min(100, Math.round((r.current_volume_liters / r.capacity_liters) * 100))
}

function formatDate(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
