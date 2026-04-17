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

    <!-- Recipes tab -->
    <template v-if="!loading && activeTab === 'recipes'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ recipes.length }} recipe(s)</p>
        <button @click="toggleRecipeForm"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
          {{ showRecipeForm ? 'Cancel' : '+ New recipe' }}
        </button>
      </div>

      <form v-if="showRecipeForm" @submit.prevent="submitRecipe"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input v-model="recipeForm.name" placeholder="Recipe name" required class="input-field" />
        <select v-model="recipeForm.target_application_type" required class="input-field">
          <option v-for="t in applicationTargets" :key="t" :value="t">{{ t.replace(/_/g, ' ') }}</option>
        </select>
        <select v-model.number="recipeForm.input_definition_id" class="input-field">
          <option :value="null">Primary input (optional)</option>
          <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
        </select>
        <input v-model="recipeForm.dilution_ratio" placeholder="Dilution ratio" class="input-field" />
        <textarea v-model="recipeForm.description" placeholder="Description" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.instructions" placeholder="Instructions" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.frequency_guidelines" placeholder="Frequency" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.notes" placeholder="Notes" class="input-field sm:col-span-2" rows="2" />
        <button type="submit" :disabled="saving"
          class="px-4 py-2 bg-green-700 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2">
          {{ saving ? 'Saving…' : (editRecipe ? 'Update recipe' : 'Create recipe') }}
        </button>
      </form>

      <div class="grid gap-4 sm:grid-cols-2">
        <div v-for="rec in recipes" :key="rec.id"
          class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 space-y-2">
          <div class="flex items-start justify-between gap-2">
            <h3 class="text-white font-semibold">{{ rec.name }}</h3>
            <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-700 text-zinc-300 capitalize">
              {{ String(rec.target_application_type || '').replace(/_/g, ' ') }}
            </span>
          </div>
          <p class="text-zinc-500 text-xs">
            Primary: {{ rec.input_definition_id ? inputName(rec.input_definition_id) : '—' }}
            <span v-if="rec.dilution_ratio"> · {{ rec.dilution_ratio }}</span>
          </p>
          <p v-if="rec.description" class="text-zinc-400 text-sm line-clamp-2">{{ rec.description }}</p>
          <div class="flex flex-wrap gap-2 pt-2">
            <button type="button" @click="openRecipeComponents(rec)" class="text-xs text-zinc-400 hover:text-white">Components</button>
            <router-link
              :to="{ path: '/fertigation', query: { tab: 'programs', recipe: rec.id } }"
              class="text-xs text-green-500 hover:text-green-400"
            >Use in program</router-link>
            <button type="button" @click="startEditRecipe(rec)" class="text-xs text-zinc-400 hover:text-white">Edit</button>
            <button type="button" @click="deleteRecipe(rec)" class="text-xs text-red-500 hover:text-red-400">Delete</button>
          </div>
        </div>
      </div>
      <p v-if="!recipes.length" class="text-zinc-500 text-sm">No recipes yet.</p>

      <div v-if="componentRecipe" class="mt-6 bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-white font-medium">Components — {{ componentRecipe.name }}</h3>
          <button type="button" class="text-xs text-zinc-500" @click="componentRecipe = null">Close</button>
        </div>
        <form class="flex flex-wrap gap-2 items-end" @submit.prevent="addComponent">
          <select v-model.number="compForm.input_definition_id" required class="input-field">
            <option value="" disabled>Input</option>
            <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
          </select>
          <input v-model.number="compForm.part_value" type="number" step="0.001" placeholder="Parts" required class="input-field w-28" />
          <button type="submit" :disabled="saving" class="px-3 py-2 bg-green-700 text-white text-xs rounded-lg">Add</button>
        </form>
        <ul class="space-y-2 text-sm">
          <li v-for="c in recipeComponents" :key="c.input_definition_id"
            class="flex justify-between gap-2 text-zinc-300 border-b border-zinc-800 pb-2">
            <span>{{ c.input_name }} — {{ c.part_value }} parts</span>
            <button type="button" class="text-xs text-red-500" @click="removeComponentRow(c)">Remove</button>
          </li>
          <li v-if="!recipeComponents.length" class="text-zinc-600">No extra components.</li>
        </ul>
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
              <th class="pb-2 pr-4">Used in</th>
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
              <td class="py-2.5 pr-4">
                <router-link v-if="batchMixCount(b.id)"
                  :to="{ path: '/fertigation', query: { tab: 'mixing' } }"
                  class="text-xs text-green-600 hover:text-green-400">
                  {{ batchMixCount(b.id) }} mix{{ batchMixCount(b.id) > 1 ? 'es' : '' }}
                </router-link>
                <span v-else class="text-xs text-zinc-600">—</span>
              </td>
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
              <td colspan="8" class="text-zinc-500 text-center py-8">No batches found.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const route = useRoute()

const store = useFarmStore()
const farmContext = useFarmContextStore()

const mixingComponentsByBatch = ref({})

function batchMixCount(batchId) {
  return mixingComponentsByBatch.value[batchId] || 0
}

const tabs = [
  { key: 'definitions', label: 'Input Definitions' },
  { key: 'batches',     label: 'Batches' },
  { key: 'recipes',     label: 'Recipes' },
]
const activeTab = ref('definitions')
const loading   = ref(true)
const saving    = ref(false)
const inputs    = ref([])
const batches   = ref([])
const recipes   = ref([])
const showRecipeForm = ref(false)
const editRecipe = ref(null)
const recipeForm = ref(emptyRecipeForm())
const componentRecipe = ref(null)
const recipeComponents = ref([])
const compForm = ref({ input_definition_id: '', part_value: 1 })

const applicationTargets = [
  'soil_drench', 'foliar_spray', 'seed_treatment', 'compost_pile_inoculant',
  'livestock_water_supplement', 'other',
]

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

function emptyRecipeForm() {
  return {
    name: '',
    target_application_type: 'soil_drench',
    input_definition_id: null,
    description: '',
    dilution_ratio: '',
    instructions: '',
    frequency_guidelines: '',
    notes: '',
  }
}

async function loadRecipesList() {
  const fid = farmContext.farmId
  if (!fid) return
  recipes.value = await store.loadRecipes(fid)
}

onMounted(async () => {
  try {
    const fid = farmContext.farmId
    const [i, b, mixEvents] = await Promise.all([
      store.loadNfInputs(fid),
      store.loadNfBatches(fid),
      store.loadMixingEvents(fid),
    ])
    inputs.value  = i
    batches.value = b
    if (route.query.tab === 'batches') activeTab.value = 'batches'
    else if (route.query.tab === 'recipes') activeTab.value = 'recipes'
    await loadRecipesList()
    const counts = {}
    for (const me of mixEvents) {
      try {
        const comps = await store.loadMixingEventComponents(fid, me.id)
        for (const c of comps) {
          if (c.input_batch_id) counts[c.input_batch_id] = (counts[c.input_batch_id] || 0) + 1
        }
      } catch { /* skip if endpoint fails */ }
    }
    mixingComponentsByBatch.value = counts
  } finally {
    loading.value = false
  }
})

watch(() => activeTab.value, (k) => {
  if (k === 'recipes') loadRecipesList()
})

function toggleRecipeForm() {
  showRecipeForm.value = !showRecipeForm.value
  if (!showRecipeForm.value) {
    editRecipe.value = null
    recipeForm.value = emptyRecipeForm()
  }
}

async function submitRecipe() {
  saving.value = true
  try {
    const fid = farmContext.farmId
    const payload = { ...recipeForm.value }
    if (payload.input_definition_id === '' || payload.input_definition_id === null) {
      delete payload.input_definition_id
    }
    if (editRecipe.value) {
      await store.updateRecipe(editRecipe.value.id, payload)
    } else {
      await store.createRecipe(fid, payload)
    }
    showRecipeForm.value = false
    editRecipe.value = null
    recipeForm.value = emptyRecipeForm()
    await loadRecipesList()
  } finally { saving.value = false }
}

function startEditRecipe(rec) {
  editRecipe.value = rec
  showRecipeForm.value = true
  recipeForm.value = {
    name: rec.name,
    target_application_type: rec.target_application_type,
    input_definition_id: rec.input_definition_id ?? null,
    description: rec.description || '',
    dilution_ratio: rec.dilution_ratio || '',
    instructions: rec.instructions || '',
    frequency_guidelines: rec.frequency_guidelines || '',
    notes: rec.notes || '',
  }
}

async function deleteRecipe(rec) {
  if (!confirm(`Delete recipe "${rec.name}"?`)) return
  await store.deleteRecipe(rec.id)
  if (componentRecipe.value?.id === rec.id) componentRecipe.value = null
  await loadRecipesList()
}

async function openRecipeComponents(rec) {
  componentRecipe.value = rec
  recipeComponents.value = await store.loadRecipeComponents(rec.id)
}

async function addComponent() {
  if (!componentRecipe.value || !compForm.value.input_definition_id) return
  saving.value = true
  try {
    await store.addRecipeComponent(componentRecipe.value.id, {
      input_definition_id: compForm.value.input_definition_id,
      part_value: compForm.value.part_value,
    })
    compForm.value = { input_definition_id: '', part_value: 1 }
    recipeComponents.value = await store.loadRecipeComponents(componentRecipe.value.id)
  } finally { saving.value = false }
}

async function removeComponentRow(c) {
  if (!componentRecipe.value) return
  await store.removeRecipeComponent(componentRecipe.value.id, c.input_definition_id)
  recipeComponents.value = await store.loadRecipeComponents(componentRecipe.value.id)
}

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
