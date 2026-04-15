<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-bold text-white">JADAM Inputs & Batches</h2>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-zinc-800 rounded-lg p-1 w-fit">
      <button
        v-for="t in tabs" :key="t.key"
        @click="activeTab = t.key"
        :class="[
          'px-4 py-1.5 text-sm rounded-md transition-colors font-medium',
          activeTab === t.key
            ? 'bg-green-600 text-white'
            : 'text-zinc-400 hover:text-white',
        ]"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="text-zinc-500 text-sm py-8 text-center">Loading...</div>

    <!-- Definitions tab -->
    <template v-if="!loading && activeTab === 'definitions'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ inputs.length }} definition(s)</p>
        <button @click="showInputForm = !showInputForm; editInput = null"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showInputForm ? 'Cancel' : '+ New Input' }}
        </button>
      </div>

      <!-- Input form (create/edit) -->
      <form v-if="showInputForm" @submit.prevent="submitInput"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input v-model="inputForm.name" placeholder="Name" required class="input-field" />
        <select v-model="inputForm.category" required class="input-field">
          <option value="" disabled>Category</option>
          <option v-for="c in categories" :key="c" :value="c">{{ formatCategory(c) }}</option>
        </select>
        <input v-model="inputForm.description" placeholder="Description" class="input-field sm:col-span-2" />
        <input v-model="inputForm.typical_ingredients" placeholder="Typical ingredients" class="input-field sm:col-span-2" />
        <input v-model="inputForm.preparation_summary" placeholder="Preparation summary" class="input-field" />
        <input v-model="inputForm.storage_guidelines" placeholder="Storage guidelines" class="input-field" />
        <input v-model="inputForm.safety_precautions" placeholder="Safety precautions" class="input-field" />
        <input v-model="inputForm.reference_source" placeholder="Reference source" class="input-field" />
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2">
          {{ saving ? 'Saving…' : (editInput ? 'Update' : 'Create') }}
        </button>
      </form>

      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="input in inputs" :key="input.id"
          class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 hover:border-green-700 transition-colors"
        >
          <div class="flex items-start justify-between mb-2">
            <h3 class="text-white font-semibold">{{ input.name }}</h3>
            <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-700 text-zinc-300">
              {{ formatCategory(input.category) }}
            </span>
          </div>
          <p v-if="input.description" class="text-zinc-400 text-sm mb-3 line-clamp-2">
            {{ input.description }}
          </p>
          <div v-if="input.typical_ingredients" class="text-xs text-zinc-500 mb-3">
            <span class="text-zinc-600 uppercase tracking-wide">Ingredients:</span>
            {{ input.typical_ingredients }}
          </div>
          <div class="flex gap-2 pt-2 border-t border-zinc-700">
            <button @click="startEditInput(input)" class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
            <button @click="confirmDeleteInput(input)" class="text-xs text-red-500 hover:text-red-400">Delete</button>
          </div>
        </div>
        <div v-if="!inputs.length" class="col-span-full text-zinc-500 text-sm text-center py-8">
          No input definitions found.
        </div>
      </div>
    </template>

    <!-- Batches tab -->
    <template v-if="!loading && activeTab === 'batches'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ batches.length }} batch(es)</p>
        <button @click="showBatchForm = !showBatchForm; editBatch = null"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showBatchForm ? 'Cancel' : '+ New Batch' }}
        </button>
      </div>

      <!-- Batch form (create/edit) -->
      <form v-if="showBatchForm" @submit.prevent="submitBatch"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
        <select v-model.number="batchForm.input_definition_id" required class="input-field"
          :disabled="!!editBatch">
          <option value="" disabled>Select input</option>
          <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
        </select>
        <input v-model="batchForm.batch_identifier" placeholder="Batch ID (e.g. FPJ-2026-04)"
          class="input-field" />
        <select v-model="batchForm.status" required class="input-field">
          <option v-for="s in batchStatuses" :key="s" :value="s">{{ formatStatus(s) }}</option>
        </select>
        <template v-if="!editBatch">
          <input v-model="batchForm.creation_start_date" type="date" class="input-field" placeholder="Start date" />
          <input v-model="batchForm.creation_end_date" type="date" class="input-field" placeholder="End date" />
          <input v-model="batchForm.expected_ready_date" type="date" class="input-field" placeholder="Expected ready" />
          <input v-model.number="batchForm.quantity_produced" type="number" step="0.1" placeholder="Qty produced"
            class="input-field" />
          <input v-model.number="batchForm.current_quantity_remaining" type="number" step="0.1"
            placeholder="Qty remaining" class="input-field" />
          <input v-model="batchForm.storage_location" placeholder="Storage location" class="input-field" />
          <input v-model.number="batchForm.shelf_life_days" type="number" placeholder="Shelf life (days)"
            class="input-field" />
          <input v-model.number="batchForm.ph_value" type="number" step="0.1" placeholder="pH" class="input-field" />
          <input v-model.number="batchForm.ec_value_ms_cm" type="number" step="0.01" placeholder="EC (mS/cm)"
            class="input-field" />
          <input v-model="batchForm.ingredients_used" placeholder="Ingredients used" class="input-field lg:col-span-3" />
          <input v-model="batchForm.procedure_followed" placeholder="Procedure followed" class="input-field lg:col-span-3" />
        </template>
        <template v-else>
          <input v-model="batchForm.actual_ready_date" type="date" class="input-field" placeholder="Actual ready date" />
          <input v-model.number="batchForm.current_quantity_remaining" type="number" step="0.1"
            placeholder="Qty remaining" class="input-field" />
          <input v-model="batchForm.storage_location" placeholder="Storage location" class="input-field" />
        </template>
        <input v-model="batchForm.observations_notes" placeholder="Notes / observations"
          class="input-field sm:col-span-2 lg:col-span-3" />
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
          {{ saving ? 'Saving…' : (editBatch ? 'Update' : 'Create') }}
        </button>
      </form>

      <div class="card overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-zinc-500 text-xs uppercase tracking-wide border-b border-zinc-700">
              <th class="pb-2 pr-4">Batch</th>
              <th class="pb-2 pr-4">Input</th>
              <th class="pb-2 pr-4">Status</th>
              <th class="pb-2 pr-4">Qty Remaining</th>
              <th class="pb-2 pr-4">Started</th>
              <th class="pb-2 pr-4">Storage</th>
              <th class="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800">
            <tr v-for="b in batches" :key="b.id" class="hover:bg-zinc-800/50 transition-colors">
              <td class="py-2.5 pr-4 font-semibold text-white">{{ b.batch_identifier ?? `#${b.id}` }}</td>
              <td class="py-2.5 pr-4 text-zinc-300">{{ inputName(b.input_definition_id) }}</td>
              <td class="py-2.5 pr-4">
                <span :class="statusClass(b.status)">{{ formatStatus(b.status) }}</span>
              </td>
              <td class="py-2.5 pr-4 font-mono text-zinc-300">{{ b.current_quantity_remaining ?? '—' }}</td>
              <td class="py-2.5 pr-4 text-zinc-400">{{ formatDate(b.creation_start_date) }}</td>
              <td class="py-2.5 pr-4 text-zinc-400">{{ b.storage_location ?? '—' }}</td>
              <td class="py-2.5">
                <div class="flex gap-2">
                  <button @click="startEditBatch(b)" class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
                  <button @click="confirmDeleteBatch(b)" class="text-xs text-red-500 hover:text-red-400">Delete</button>
                </div>
              </td>
            </tr>
            <tr v-if="!batches.length">
              <td colspan="7" class="text-zinc-500 text-center py-8">No batches found.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const tabs = [
  { key: 'definitions', label: 'Input Definitions' },
  { key: 'batches',     label: 'Batches' },
]
const activeTab = ref('definitions')
const loading   = ref(true)
const saving    = ref(false)
const inputs    = ref([])
const batches   = ref([])

const showInputForm = ref(false)
const editInput = ref(null)
const inputForm = ref(emptyInputForm())

const showBatchForm = ref(false)
const editBatch = ref(null)
const batchForm = ref(emptyBatchForm())

const categories = [
  'microbial_inoculant', 'fermented_plant_juice', 'water_soluble_nutrient',
  'oriental_herbal_nutrient', 'fish_amino_acid', 'insect_attractant_repellent',
  'soil_conditioner', 'compost_tea_extract', 'biochar_preparation',
  'other_ferment', 'other_extract',
]

const batchStatuses = [
  'planning', 'ingredients_gathered', 'mixing_in_progress', 'fermenting_brewing',
  'maturing_aging', 'ready_for_use', 'partially_used', 'fully_used',
  'expired_discarded', 'failed_production',
]

function emptyInputForm() {
  return { name: '', category: '', description: '', typical_ingredients: '', preparation_summary: '', storage_guidelines: '', safety_precautions: '', reference_source: '' }
}
function emptyBatchForm() {
  return { input_definition_id: '', batch_identifier: '', status: 'planning', creation_start_date: '', creation_end_date: '', expected_ready_date: '', quantity_produced: null, current_quantity_remaining: null, storage_location: '', shelf_life_days: null, ph_value: null, ec_value_ms_cm: null, ingredients_used: '', procedure_followed: '', observations_notes: '', actual_ready_date: '' }
}

onMounted(async () => {
  try {
    const fid = farmContext.farmId
    const [i, b] = await Promise.all([
      store.loadNfInputs(fid),
      store.loadNfBatches(fid),
    ])
    inputs.value  = i
    batches.value = b
  } finally {
    loading.value = false
  }
})

async function submitInput() {
  saving.value = true
  try {
    if (editInput.value) {
      await store.updateNfInput(editInput.value.id, inputForm.value)
    } else {
      await store.createNfInput(farmContext.farmId, inputForm.value)
    }
    inputs.value = await store.loadNfInputs(farmContext.farmId)
    showInputForm.value = false
    editInput.value = null
    inputForm.value = emptyInputForm()
  } finally { saving.value = false }
}

function startEditInput(input) {
  editInput.value = input
  inputForm.value = { name: input.name, category: input.category || '', description: input.description || '', typical_ingredients: input.typical_ingredients || '', preparation_summary: input.preparation_summary || '', storage_guidelines: input.storage_guidelines || '', safety_precautions: input.safety_precautions || '', reference_source: input.reference_source || '' }
  showInputForm.value = true
}

async function confirmDeleteInput(input) {
  if (!confirm(`Delete input "${input.name}"?`)) return
  await store.deleteNfInput(input.id)
  inputs.value = await store.loadNfInputs(farmContext.farmId)
}

async function submitBatch() {
  saving.value = true
  try {
    if (editBatch.value) {
      await store.updateNfBatch(editBatch.value.id, {
        batch_identifier: batchForm.value.batch_identifier,
        status: batchForm.value.status,
        actual_ready_date: batchForm.value.actual_ready_date || null,
        current_quantity_remaining: batchForm.value.current_quantity_remaining,
        storage_location: batchForm.value.storage_location,
        observations_notes: batchForm.value.observations_notes,
      })
    } else {
      await store.createNfBatch(farmContext.farmId, batchForm.value)
    }
    batches.value = await store.loadNfBatches(farmContext.farmId)
    showBatchForm.value = false
    editBatch.value = null
    batchForm.value = emptyBatchForm()
  } finally { saving.value = false }
}

function startEditBatch(batch) {
  editBatch.value = batch
  batchForm.value = {
    ...emptyBatchForm(),
    input_definition_id: batch.input_definition_id,
    batch_identifier: batch.batch_identifier || '',
    status: batch.status || 'planning',
    actual_ready_date: batch.actual_ready_date || '',
    current_quantity_remaining: batch.current_quantity_remaining,
    storage_location: batch.storage_location || '',
    observations_notes: batch.observations_notes || '',
  }
  showBatchForm.value = true
}

async function confirmDeleteBatch(batch) {
  if (!confirm(`Delete batch "${batch.batch_identifier || '#' + batch.id}"?`)) return
  await store.deleteNfBatch(batch.id)
  batches.value = await store.loadNfBatches(farmContext.farmId)
}

const inputName = (id) => {
  const found = inputs.value.find(i => i.id === id)
  return found ? found.name : `#${id}`
}

const formatCategory = (c) =>
  c ? c.replace(/_/g, ' ') : ''

const formatStatus = (s) =>
  s ? s.replace(/_/g, ' ') : ''

const formatDate = (d) => {
  if (!d) return '—'
  try { return new Date(d).toLocaleDateString() } catch { return d }
}

const statusClass = (s) => {
  const base = 'text-xs font-semibold px-2 py-0.5 rounded-full'
  if (s === 'ready_for_use')     return `${base} bg-green-900 text-green-300`
  if (s === 'fermenting_brewing' || s === 'maturing_aging') return `${base} bg-amber-900 text-amber-300`
  if (s === 'fully_used' || s === 'expired_discarded')      return `${base} bg-zinc-700 text-zinc-400`
  return `${base} bg-zinc-700 text-zinc-300`
}
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
