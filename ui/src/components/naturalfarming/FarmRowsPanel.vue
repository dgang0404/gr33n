<template>
  <div class="space-y-4 max-w-6xl" data-test="nf-farm-rows">
    <div>
      <h2 class="text-lg font-semibold text-white">{{ NF_VOCAB.inputs }} &amp; {{ NF_VOCAB.batches.toLowerCase() }}</h2>
      <p class="text-sm text-zinc-500 mt-1">
        Edit farm rows — metadata, status, delete. Quantities, restock, and unit costs live in
        <router-link v-nav-hint="'/money'" :to="moneySuppliesLink" class="text-green-500 hover:underline">
          Money → Supplies on hand
        </router-link>.
      </p>
    </div>

    <div class="flex gap-1 bg-zinc-800 rounded-lg p-1 w-fit">
      <button
        v-for="t in tabs"
        :key="t.key"
        type="button"
        :class="[
          'px-4 py-1.5 text-sm rounded-md transition-colors font-medium',
          activeTab === t.key ? 'bg-green-600 text-white' : 'text-zinc-400 hover:text-white',
        ]"
        :data-test="`nf-manage-tab-${t.key}`"
        @click="selectSubTab(t.key)"
      >
        <span class="inline-flex items-center">
          {{ t.label }}
          <ConceptHelpTip v-if="t.conceptId" :concept-id="t.conceptId" position="bottom" />
        </span>
      </button>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm py-8 text-center">Loading…</div>

    <template v-else-if="activeTab === 'definitions'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ inputs.length }} {{ NF_VOCAB.input.toLowerCase() }}(s)</p>
        <button
          type="button"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg"
          data-test="nf-manage-new-input"
          @click="showInputForm = !showInputForm; editInput = null"
        >
          {{ showInputForm ? 'Cancel' : '+ New input' }}
        </button>
      </div>

      <form
        v-if="showInputForm"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3"
        @submit.prevent="submitInput"
      >
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
        <button
          type="submit"
          :disabled="saving"
          class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2"
        >
          {{ saving ? 'Saving…' : (editInput ? 'Update' : 'Create') }}
        </button>
      </form>

      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="input in inputs"
          :key="input.id"
          class="bg-zinc-800 border border-zinc-700 rounded-xl p-4 hover:border-green-700 transition-colors"
        >
          <div class="flex items-start justify-between mb-2">
            <h3 class="text-white font-semibold">{{ input.name }}</h3>
            <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-700 text-zinc-300">
              {{ formatCategory(input.category) }}
            </span>
          </div>
          <p v-if="input.description" class="text-zinc-400 text-sm mb-3 line-clamp-2">{{ input.description }}</p>
          <div class="flex gap-2 pt-2 border-t border-zinc-700">
            <button type="button" class="text-xs text-zinc-400 hover:text-zinc-200" @click="startEditInput(input)">Edit</button>
            <button type="button" class="text-xs text-red-500 hover:text-red-400" @click="confirmDeleteInput(input)">Delete</button>
          </div>
        </div>
        <div v-if="!inputs.length" class="col-span-full text-zinc-500 text-sm text-center py-8">
          No inputs yet — create one here or use Make a batch.
        </div>
      </div>
    </template>

    <template v-else-if="activeTab === 'batches'">
      <div class="flex items-center justify-between">
        <p class="text-zinc-400 text-sm">{{ batches.length }} {{ NF_VOCAB.batch.toLowerCase() }}(es)</p>
        <button
          type="button"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg"
          data-test="nf-manage-new-batch"
          @click="showBatchForm = !showBatchForm; editBatch = null"
        >
          {{ showBatchForm ? 'Cancel' : '+ New batch' }}
        </button>
      </div>

      <form
        v-if="showBatchForm"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3"
        @submit.prevent="submitBatch"
      >
        <select v-model.number="batchForm.input_definition_id" required class="input-field" :disabled="!!editBatch">
          <option value="" disabled>Select input</option>
          <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
        </select>
        <input v-model="batchForm.batch_identifier" placeholder="Batch ID (e.g. FPJ-2026-04)" class="input-field" />
        <select v-model="batchForm.status" required class="input-field">
          <option v-for="s in batchStatuses" :key="s" :value="s">{{ formatStatus(s) }}</option>
        </select>
        <template v-if="!editBatch">
          <input v-model="batchForm.creation_start_date" type="date" class="input-field" />
          <input v-model="batchForm.creation_end_date" type="date" class="input-field" />
          <input v-model="batchForm.expected_ready_date" type="date" class="input-field" />
          <input v-model.number="batchForm.quantity_produced" type="number" step="0.1" placeholder="Qty produced" class="input-field" />
          <input v-model.number="batchForm.current_quantity_remaining" type="number" step="0.1" placeholder="Qty remaining" class="input-field" />
          <input v-model="batchForm.storage_location" placeholder="Storage location" class="input-field" />
          <input v-model.number="batchForm.shelf_life_days" type="number" placeholder="Shelf life (days)" class="input-field" />
          <input v-model.number="batchForm.ph_value" type="number" step="0.1" placeholder="pH" class="input-field" />
          <input v-model.number="batchForm.ec_value_ms_cm" type="number" step="0.01" placeholder="EC (mS/cm)" class="input-field" />
          <input v-model="batchForm.ingredients_used" placeholder="Ingredients used" class="input-field lg:col-span-3" />
          <input v-model="batchForm.procedure_followed" placeholder="Procedure followed" class="input-field lg:col-span-3" />
        </template>
        <template v-else>
          <input v-model="batchForm.actual_ready_date" type="date" class="input-field" />
          <input v-model.number="batchForm.current_quantity_remaining" type="number" step="0.1" placeholder="Qty remaining" class="input-field" />
          <input v-model="batchForm.storage_location" placeholder="Storage location" class="input-field" />
        </template>
        <input
          v-model.number="batchForm.low_stock_threshold"
          type="number"
          step="0.1"
          placeholder="Low-stock threshold (optional)"
          class="input-field"
        />
        <input v-model="batchForm.observations_notes" placeholder="Notes / observations" class="input-field sm:col-span-2 lg:col-span-3" />
        <button type="submit" :disabled="saving" class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
          {{ saving ? 'Saving…' : (editBatch ? 'Update' : 'Create') }}
        </button>
      </form>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="text-left text-zinc-500 text-xs uppercase tracking-wide border-b border-zinc-700">
              <th class="pb-2 pr-4">Batch</th>
              <th class="pb-2 pr-4">Input</th>
              <th class="pb-2 pr-4">Status</th>
              <th class="pb-2 pr-4">Remaining</th>
              <th class="pb-2 pr-4">Mixes</th>
              <th class="pb-2 pr-4">Started</th>
              <th class="pb-2 pr-4">Storage</th>
              <th class="pb-2">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800">
            <tr
              v-for="b in batches"
              :key="b.id"
              class="transition-colors"
              :class="highlightedBatchId === b.id ? 'bg-green-950/30 ring-1 ring-green-800/60' : 'hover:bg-zinc-800/50'"
              :data-test="`nf-manage-batch-row-${b.id}`"
            >
              <td class="py-2.5 pr-4 font-semibold text-white">{{ b.batch_identifier ?? `#${b.id}` }}</td>
              <td class="py-2.5 pr-4 text-zinc-300">{{ inputName(b.input_definition_id) }}</td>
              <td class="py-2.5 pr-4"><span :class="statusClass(b.status)">{{ formatStatus(b.status) }}</span></td>
              <td class="py-2.5 pr-4 font-mono text-zinc-300">{{ b.current_quantity_remaining ?? '—' }}</td>
              <td class="py-2.5 pr-4">
                <router-link
                  v-if="batchMixCount(b.id)"
                  v-nav-hint="'/fertigation'"
                  :to="feedWaterFertigationRoute('mixing')"
                  class="text-xs text-green-600 hover:text-green-400"
                >
                  {{ batchMixCount(b.id) }} mix{{ batchMixCount(b.id) > 1 ? 'es' : '' }}
                </router-link>
                <span v-else class="text-xs text-zinc-600">—</span>
              </td>
              <td class="py-2.5 pr-4 text-zinc-400">{{ formatDate(b.creation_start_date) }}</td>
              <td class="py-2.5 pr-4 text-zinc-400">{{ b.storage_location ?? '—' }}</td>
              <td class="py-2.5">
                <div class="flex gap-2">
                  <button type="button" class="text-xs text-zinc-400 hover:text-zinc-200" @click="startEditBatch(b)">Edit</button>
                  <button type="button" class="text-xs text-red-500 hover:text-red-400" @click="confirmDeleteBatch(b)">Delete</button>
                </div>
              </td>
            </tr>
            <tr v-if="!batches.length">
              <td colspan="8" class="text-zinc-500 text-center py-8">No batches yet.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { feedWaterFertigationRoute, moneySuppliesRoute, naturalFarmingManageRoute } from '../../lib/workspaceRoutes.js'
import { useFarmStore } from '../../stores/farm.js'
import { useFarmContextStore } from '../../stores/farmContext.js'
import api from '../../api'
import { loadDomainEnums, enumValues, enumLabel } from '../../lib/domainEnums.js'
import ConceptHelpTip from '../ConceptHelpTip.vue'
import { NF_MANAGE_TAB_LABELS, NF_VOCAB } from '../../lib/naturalFarmingVocabulary.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const moneySuppliesLink = moneySuppliesRoute()

const mixingComponentsByBatch = ref({})
const tabs = [
  { key: 'definitions', label: NF_MANAGE_TAB_LABELS.definitions, conceptId: 'input_definition' },
  { key: 'batches', label: NF_MANAGE_TAB_LABELS.batches, conceptId: 'input_batch' },
]
const activeTab = ref('definitions')
const loading = ref(true)
const saving = ref(false)
const inputs = ref([])
const batches = ref([])
const domainEnums = ref(null)

const categories = computed(() => enumValues(domainEnums.value, 'input_definition_categories'))
const batchStatuses = computed(() => enumValues(domainEnums.value, 'batch_statuses'))

const showInputForm = ref(false)
const editInput = ref(null)
const inputForm = ref(emptyInputForm())
const showBatchForm = ref(false)
const editBatch = ref(null)
const batchForm = ref(emptyBatchForm())

const highlightedBatchId = computed(() => {
  const raw = route.query.batch_id
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(id) ? id : null
})

function emptyInputForm() {
  return {
    name: '', category: '', description: '', typical_ingredients: '',
    preparation_summary: '', storage_guidelines: '', safety_precautions: '', reference_source: '',
  }
}

function emptyBatchForm() {
  return {
    input_definition_id: '', batch_identifier: '', status: 'planning',
    creation_start_date: '', creation_end_date: '', expected_ready_date: '',
    quantity_produced: null, current_quantity_remaining: null, storage_location: '',
    shelf_life_days: null, ph_value: null, ec_value_ms_cm: null,
    ingredients_used: '', procedure_followed: '', observations_notes: '',
    actual_ready_date: '', low_stock_threshold: null,
  }
}

function batchMixCount(batchId) {
  return mixingComponentsByBatch.value[batchId] || 0
}

function applySubTabFromRoute() {
  const inv = route.query.inv
  if (inv === 'batches' || route.query.batch_id) activeTab.value = 'batches'
  else activeTab.value = 'definitions'
}

function selectSubTab(key) {
  activeTab.value = key
  router.replace(naturalFarmingManageRoute({
    inv: key,
    batchId: key === 'batches' ? route.query.batch_id : undefined,
  })).catch(() => {})
}

async function loadAll() {
  const fid = farmContext.farmId
  if (!fid) return
  const [i, b, mixEvents, enums] = await Promise.all([
    store.loadNfInputs(fid),
    store.loadNfBatches(fid),
    store.loadMixingEvents(fid),
    loadDomainEnums(api),
  ])
  domainEnums.value = enums
  inputs.value = i
  batches.value = b
  const counts = {}
  for (const me of mixEvents) {
    try {
      const comps = await store.loadMixingEventComponents(fid, me.id)
      for (const c of comps) {
        if (c.input_batch_id) counts[c.input_batch_id] = (counts[c.input_batch_id] || 0) + 1
      }
    } catch { /* skip */ }
  }
  mixingComponentsByBatch.value = counts
}

onMounted(async () => {
  try {
    applySubTabFromRoute()
    await loadAll()
  } finally {
    loading.value = false
  }
})

watch(() => [route.query.inv, route.query.batch_id], () => {
  applySubTabFromRoute()
})

watch(highlightedBatchId, (id) => {
  if (id && batches.value.some((b) => b.id === id)) {
    activeTab.value = 'batches'
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
  } finally {
    saving.value = false
  }
}

function startEditInput(input) {
  editInput.value = input
  inputForm.value = {
    name: input.name,
    category: input.category || '',
    description: input.description || '',
    typical_ingredients: input.typical_ingredients || '',
    preparation_summary: input.preparation_summary || '',
    storage_guidelines: input.storage_guidelines || '',
    safety_precautions: input.safety_precautions || '',
    reference_source: input.reference_source || '',
  }
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
        low_stock_threshold: batchForm.value.low_stock_threshold,
      })
    } else {
      await store.createNfBatch(farmContext.farmId, batchForm.value)
    }
    batches.value = await store.loadNfBatches(farmContext.farmId)
    showBatchForm.value = false
    editBatch.value = null
    batchForm.value = emptyBatchForm()
  } finally {
    saving.value = false
  }
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
    low_stock_threshold: batch.low_stock_threshold ?? null,
  }
  showBatchForm.value = true
}

async function confirmDeleteBatch(batch) {
  if (!confirm(`Delete batch "${batch.batch_identifier || '#' + batch.id}"?`)) return
  await store.deleteNfBatch(batch.id)
  batches.value = await store.loadNfBatches(farmContext.farmId)
}

function inputName(id) {
  return inputs.value.find((i) => i.id === id)?.name || `#${id}`
}

function formatCategory(c) {
  return enumLabel('input_definition_categories', c, domainEnums.value)
}

function formatStatus(s) {
  return enumLabel('batch_statuses', s, domainEnums.value)
}

function formatDate(d) {
  if (!d) return '—'
  try { return new Date(d).toLocaleDateString() } catch { return d }
}

function statusClass(s) {
  const base = 'text-xs font-semibold px-2 py-0.5 rounded-full'
  if (s === 'ready_for_use') return `${base} bg-green-900 text-green-300`
  if (s === 'fermenting_brewing' || s === 'maturing_aging') return `${base} bg-amber-900 text-amber-300`
  if (s === 'fully_used' || s === 'expired_discarded') return `${base} bg-zinc-700 text-zinc-400`
  return `${base} bg-zinc-700 text-zinc-300`
}
</script>
