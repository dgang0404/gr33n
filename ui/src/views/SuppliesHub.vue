<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Supplies</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          What you have on hand, what is running low, and where to log a mix. Farm-wide stock — not tied to one zone.
        </p>
      </div>
      <div class="flex flex-wrap gap-2 shrink-0">
        <button
          type="button"
          class="px-3 py-1.5 text-xs font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
          data-test="supplies-new-batch"
          @click="openNewBatch()"
        >
          + New batch
        </button>
        <button
          type="button"
          class="text-xs text-zinc-400 hover:text-zinc-200"
          @click="refresh"
        >
          Refresh
        </button>
      </div>
    </div>

    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Supplies"
      back-to-zone-tab="water"
      :clear-route="{ path: '/operations/supplies' }"
    />

    <GuardianStarterChips :starters="suppliesStarters" />

    <p v-if="actionError" class="text-xs text-red-400" data-test="supplies-action-error">{{ actionError }}</p>
    <p v-if="actionSuccess" class="text-xs text-green-400" data-test="supplies-action-success">{{ actionSuccess }}</p>

    <div
      v-if="lowStockRows.length"
      class="rounded-xl border border-amber-800/80 bg-amber-950/40 px-4 py-3 space-y-2"
      data-test="supplies-low-stock-banner"
    >
      <p class="text-sm font-medium text-amber-200">
        {{ lowStockRows.length }} supply batch{{ lowStockRows.length === 1 ? '' : 'es' }} below the low-stock threshold
      </p>
      <ul class="text-sm text-amber-100/90 space-y-1">
        <li v-for="row in lowStockRows.slice(0, 5)" :key="row.batch.id" class="flex flex-wrap items-center gap-2">
          <span>
            <strong>{{ row.inputName }}</strong>
            — {{ formatQty(row.remaining) }} left (threshold {{ formatQty(row.threshold) }})
          </span>
          <button
            type="button"
            v-nav-hint="'/tasks'"
            class="text-[10px] px-2 py-0.5 rounded bg-amber-900/60 text-amber-100 hover:bg-amber-800/80"
            :disabled="refillTaskSaving === row.batch.id"
            data-test="supplies-refill-task"
            @click="createRefillTask(row)"
          >
            {{ refillTaskSaving === row.batch.id ? 'Creating…' : 'Create refill task' }}
          </button>
        </li>
      </ul>
      <p v-if="lowStockRows.length > 5" class="text-xs text-amber-300/80">
        + {{ lowStockRows.length - 5 }} more in the list below
      </p>
      <div class="flex flex-wrap gap-3 pt-1">
        <router-link
          v-if="lowStockAlertLink"
          v-nav-hint="'/alerts'"
          :to="lowStockAlertLink"
          class="inline-block text-xs text-amber-300 hover:text-amber-100 underline"
        >
          View low-stock alert →
        </router-link>
        <router-link
          v-nav-hint="'/tasks'"
          to="/tasks"
          class="inline-block text-xs text-amber-300/80 hover:text-amber-100 underline"
        >
          Open tasks →
        </router-link>
      </div>
    </div>

    <div class="flex flex-wrap gap-2">
      <router-link
        v-nav-hint="'/operations/feeding'"
        :to="logMixLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 transition-colors"
        data-test="supplies-log-mix"
      >
        Log a mix
      </router-link>
      <router-link
        :to="recipesLink"
        class="px-4 py-2 text-sm font-medium rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-700 hover:bg-zinc-700 transition-colors"
      >
        Mixing recipes ({{ recipes.length }})
      </router-link>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading supplies…</div>

    <EmptyStateHint
      v-else-if="!supplyRows.length"
      reason="no_data"
      :message="inputs.length ? 'No batches on hand — start one below or use the full editor.' : 'No supply batches yet — add inputs and batches in the full editor, or start from a demo farm.'"
      :action-label="inputs.length ? 'Create first batch' : 'Open full editor'"
      :action-to="inputs.length ? null : { path: '/inventory', query: { tab: 'batches' } }"
      @action="openNewBatch()"
    />

    <div v-else-if="supplyRows.length" class="space-y-3">
      <p class="text-xs text-zinc-500 uppercase tracking-widest">
        On hand — {{ supplyRows.length }} batch{{ supplyRows.length === 1 ? '' : 'es' }}
      </p>
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        <div
          v-for="row in supplyRows"
          :key="row.id"
          class="bg-zinc-900 border rounded-xl p-4 transition-colors"
          :class="row.lowStock ? 'border-amber-800/80' : 'border-zinc-800'"
          :data-test="`supply-row-${row.id}`"
        >
          <div class="flex items-start justify-between gap-2 mb-2">
            <div>
              <p class="text-white font-medium">{{ row.inputName }}</p>
              <p class="text-zinc-600 text-xs mt-0.5">{{ row.batchLabel }} · Farm-wide</p>
            </div>
            <span
              v-if="row.lowStock"
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-900/60 text-amber-200 font-semibold shrink-0"
            >Low</span>
            <span
              v-else
              class="text-[10px] px-1.5 py-0.5 rounded-full bg-zinc-800 text-zinc-500 capitalize shrink-0"
            >{{ formatStatus(row.status) }}</span>
          </div>

          <dl class="grid grid-cols-2 gap-2 text-xs mb-3">
            <div>
              <dt class="text-zinc-600">On hand</dt>
              <dd class="text-zinc-200 font-mono">{{ formatQty(row.remaining) }}</dd>
            </div>
            <div v-if="row.threshold != null">
              <dt class="text-zinc-600">Low at</dt>
              <dd class="text-zinc-400 font-mono">{{ formatQty(row.threshold) }}</dd>
            </div>
            <div v-if="row.storageLocation" class="col-span-2">
              <dt class="text-zinc-600">Storage</dt>
              <dd class="text-zinc-400">{{ row.storageLocation }}</dd>
            </div>
            <div class="col-span-2">
              <dt class="text-zinc-600">Unit cost</dt>
              <dd v-if="unitCostEditId !== row.inputDefinitionId" class="text-zinc-300">
                {{ row.unitCostLabel || 'Not set' }}
                <button
                  type="button"
                  class="ml-2 text-green-600 hover:text-green-400"
                  data-test="supplies-edit-unit-cost"
                  @click="startUnitCostEdit(row)"
                >
                  {{ row.unitCostLabel ? 'Edit' : 'Set cost' }}
                </button>
              </dd>
              <form
                v-else
                class="flex flex-wrap items-end gap-2 mt-1"
                data-test="supplies-unit-cost-form"
                @submit.prevent="submitUnitCost(row)"
              >
                <label class="flex flex-col gap-0.5">
                  <span class="text-[10px] text-zinc-500">$ per unit</span>
                  <input
                    v-model.number="unitCostForm.unit_cost"
                    type="number"
                    step="0.01"
                    min="0"
                    required
                    class="w-24 bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-zinc-200 font-mono text-xs"
                  />
                </label>
                <button type="submit" class="text-xs text-green-500" :disabled="saving">Save</button>
                <button type="button" class="text-xs text-zinc-500" @click="cancelUnitCostEdit">Cancel</button>
              </form>
            </div>
          </dl>

          <div
            v-if="restockBatchId === row.id"
            class="mb-3 p-2 rounded-lg border border-zinc-800 bg-zinc-950 space-y-2"
            data-test="supplies-restock-form"
          >
            <p class="text-[10px] text-zinc-500">Add quantity to on-hand (does not replace the total)</p>
            <div class="flex flex-wrap items-center gap-2">
              <input
                v-model.number="restockQty"
                type="number"
                step="0.1"
                min="0.1"
                placeholder="Qty to add"
                class="w-28 bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-white"
              />
              <button
                type="button"
                v-nav-hint="'/operations/supplies'"
                class="text-xs px-2 py-1 rounded bg-green-800 text-white disabled:opacity-50"
                :disabled="saving || !restockQty"
                data-test="supplies-restock-submit"
                @click="submitRestock(row)"
              >
                {{ saving ? 'Saving…' : 'Add' }}
              </button>
              <button type="button" class="text-xs text-zinc-500" @click="cancelRestock">Cancel</button>
            </div>
            <p v-if="restockPreview(row)" class="text-[10px] text-zinc-500">
              New on hand: <span class="text-zinc-300 font-mono">{{ restockPreview(row) }}</span>
            </p>
          </div>

          <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
            <button
              type="button"
              v-nav-hint="'/operations/supplies'"
              class="text-xs text-green-500 hover:text-green-400"
              data-test="supplies-restock-btn"
              @click="startRestock(row)"
            >
              + Add qty
            </button>
            <router-link
              v-if="mixCount(row.id)"
              v-nav-hint="'/operations/feeding'"
              :to="logMixLink"
              class="text-xs text-green-500 hover:text-green-400"
            >
              {{ mixCount(row.id) }} mix{{ mixCount(row.id) > 1 ? 'es' : '' }} logged
            </router-link>
            <button
              v-if="row.lowStock"
              type="button"
              v-nav-hint="'/tasks'"
              class="text-xs text-amber-400 hover:text-amber-200"
              :disabled="refillTaskSaving === row.id"
              data-test="supplies-refill-task-row"
              @click="createRefillTaskForRow(row)"
            >
              Refill task
            </button>
            <button
              type="button"
              class="text-xs text-zinc-500 hover:text-zinc-300"
              @click="openBatchEditor(row.id)"
            >
              Advanced editor →
            </button>
          </div>
        </div>
      </div>
    </div>

    <footer class="pt-2 border-t border-zinc-800">
      <router-link
        :to="{ path: '/inventory', query: zoneContextId ? { tab: 'definitions', zone_id: String(zoneContextId) } : { tab: 'definitions' } }"
        class="text-xs text-zinc-400 hover:text-green-400"
        data-test="supplies-advanced-footer"
      >
        Full inventory editor (definitions, recipes, batches) →
      </router-link>
    </footer>

    <!-- Quick new batch dialog -->
    <div
      v-if="showNewBatch"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      data-test="supplies-new-batch-modal"
      @click.self="closeNewBatch"
    >
      <form
        class="bg-zinc-900 border border-zinc-700 rounded-xl p-5 w-full max-w-md space-y-3"
        @submit.prevent="submitNewBatch"
      >
        <div class="flex items-center justify-between">
          <h2 class="text-sm font-semibold text-white">New supply batch</h2>
          <button type="button" class="text-xs text-zinc-500 hover:text-zinc-200" @click="closeNewBatch">Close</button>
        </div>
        <p class="text-[11px] text-zinc-500">Quick restock when a batch ran out or you received a delivery.</p>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Input</label>
          <select
            v-model.number="newBatchForm.input_definition_id"
            required
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          >
            <option :value="null" disabled>Select input</option>
            <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Batch label (optional)</label>
          <input
            v-model="newBatchForm.batch_identifier"
            type="text"
            placeholder="e.g. Delivery 2026-06"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Quantity on hand</label>
            <input
              v-model.number="newBatchForm.current_quantity_remaining"
              type="number"
              step="0.1"
              min="0"
              required
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            />
          </div>
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Low-stock alert at</label>
            <input
              v-model.number="newBatchForm.low_stock_threshold"
              type="number"
              step="0.1"
              min="0"
              placeholder="Optional"
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            />
          </div>
        </div>
        <p v-if="newBatchError" class="text-xs text-red-400">{{ newBatchError }}</p>
        <div class="flex justify-end gap-2 pt-1">
          <button type="button" class="text-xs text-zinc-400" @click="closeNewBatch">Cancel</button>
          <button
            type="submit"
            class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-700 text-white disabled:opacity-40"
            :disabled="saving"
          >
            {{ saving ? 'Saving…' : 'Create batch' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from '../components/ZoneContextBanner.vue'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { parseZoneIdQuery } from '../lib/zoneContext.js'
import { buildSuppliesHubStarters } from '../lib/guardianStarters.js'
import {
  buildRefillTaskPayload,
  buildSupplyRows,
  filterLowStockAlerts,
  listLowStockBatches,
  nextQuantityAfterRestock,
} from '../lib/suppliesHub.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const saving = ref(false)
const inputs = ref([])
const batches = ref([])
const recipes = ref([])
const alerts = ref([])
const programs = ref([])
const mixingComponentsByBatch = ref({})

const actionError = ref('')
const actionSuccess = ref('')
const restockBatchId = ref(null)
const restockQty = ref(null)
const unitCostEditId = ref(null)
const unitCostForm = reactive({ unit_cost: null })
const showNewBatch = ref(false)
const newBatchError = ref('')
const newBatchForm = reactive({
  input_definition_id: null,
  batch_identifier: '',
  current_quantity_remaining: null,
  low_stock_threshold: null,
})
const refillTaskSaving = ref(null)

const zoneContextId = computed(() => parseZoneIdQuery(route.query.zone_id))

const lowStockRows = computed(() => listLowStockBatches(batches.value, inputs.value))

const supplyRows = computed(() => buildSupplyRows(batches.value, inputs.value))

const lowStockAlerts = computed(() => filterLowStockAlerts(alerts.value))

const lowStockAlertLink = computed(() => {
  const first = lowStockAlerts.value[0]
  if (!first) return null
  return { path: '/alerts', query: { highlight: String(first.id) } }
})

const suppliesStarters = computed(() => buildSuppliesHubStarters({
  lowStockRows: lowStockRows.value,
  lowStockAlerts: lowStockAlerts.value,
  recipes: recipes.value,
  zones: store.zones,
  zoneContextId: zoneContextId.value,
  programs: programs.value,
  surface: zoneContextId.value ? 'supplies_hub_zone' : 'supplies_hub',
}))

const logMixLink = computed(() => {
  const q = { tab: 'mixing' }
  if (zoneContextId.value) q.zone_id = String(zoneContextId.value)
  return { path: '/operations/feeding', query: q }
})

const recipesLink = computed(() => ({
  path: '/inventory',
  query: { tab: 'recipes' },
}))

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function mixCount(batchId) {
  return mixingComponentsByBatch.value[batchId] || 0
}

function formatQty(value) {
  if (value == null || value === '') return '—'
  const n = Number(value)
  return Number.isFinite(n) ? String(n) : String(value)
}

function formatStatus(status) {
  return status ? String(status).replace(/_/g, ' ') : '—'
}

function clearActionFeedback() {
  actionError.value = ''
  actionSuccess.value = ''
}

function flashSuccess(msg) {
  actionSuccess.value = msg
  actionError.value = ''
  setTimeout(() => { actionSuccess.value = '' }, 4000)
}

function openBatchEditor(batchId) {
  router.push({ path: '/inventory', query: { tab: 'batches', batch_id: String(batchId) } })
}

function startRestock(row) {
  clearActionFeedback()
  restockBatchId.value = row.id
  restockQty.value = null
}

function cancelRestock() {
  restockBatchId.value = null
  restockQty.value = null
}

function restockPreview(row) {
  const next = nextQuantityAfterRestock(row.remaining, restockQty.value)
  return next == null ? null : formatQty(next)
}

async function submitRestock(row) {
  const next = nextQuantityAfterRestock(row.remaining, restockQty.value)
  if (next == null) {
    actionError.value = 'Enter a positive quantity to add.'
    return
  }
  saving.value = true
  clearActionFeedback()
  try {
    const batch = batches.value.find((b) => b.id === row.id)
    await store.updateNfBatch(row.id, {
      status: batch?.status || row.batchStatus || 'ready_for_use',
      current_quantity_remaining: next,
    })
    batches.value = await store.loadNfBatches(farmContext.farmId)
    cancelRestock()
    flashSuccess(`Added ${restockQty.value} — now ${formatQty(next)} on hand.`)
  } catch (e) {
    actionError.value = e.response?.data?.error || e.message || 'Restock failed'
  } finally {
    saving.value = false
  }
}

function startUnitCostEdit(row) {
  clearActionFeedback()
  unitCostEditId.value = row.inputDefinitionId
  unitCostForm.unit_cost = row.unitCost != null ? Number(row.unitCost) : null
}

function cancelUnitCostEdit() {
  unitCostEditId.value = null
}

async function submitUnitCost(row) {
  if (unitCostForm.unit_cost == null || Number(unitCostForm.unit_cost) < 0) {
    actionError.value = 'Enter a valid unit cost.'
    return
  }
  saving.value = true
  clearActionFeedback()
  try {
    const input = inputs.value.find((i) => i.id === row.inputDefinitionId)
    await store.updateNfInput(row.inputDefinitionId, {
      name: input?.name,
      category: input?.category,
      description: input?.description || '',
      typical_ingredients: input?.typical_ingredients || '',
      preparation_summary: input?.preparation_summary || '',
      storage_guidelines: input?.storage_guidelines || '',
      safety_precautions: input?.safety_precautions || '',
      reference_source: input?.reference_source || '',
      unit_cost: Number(unitCostForm.unit_cost),
      unit_cost_currency: input?.unit_cost_currency || 'USD',
      unit_cost_unit_id: input?.unit_cost_unit_id ?? null,
    })
    inputs.value = await store.loadNfInputs(farmContext.farmId)
    cancelUnitCostEdit()
    flashSuccess('Unit cost saved.')
  } catch (e) {
    actionError.value = e.response?.data?.error || e.message || 'Could not save unit cost'
  } finally {
    saving.value = false
  }
}

function openNewBatch(prefillInputId = null) {
  clearActionFeedback()
  newBatchError.value = ''
  newBatchForm.input_definition_id = prefillInputId
  newBatchForm.batch_identifier = ''
  newBatchForm.current_quantity_remaining = null
  newBatchForm.low_stock_threshold = null
  showNewBatch.value = true
}

function closeNewBatch() {
  showNewBatch.value = false
  newBatchError.value = ''
}

async function submitNewBatch() {
  const fid = farmContext.farmId
  if (!fid || !newBatchForm.input_definition_id) {
    newBatchError.value = 'Pick an input.'
    return
  }
  if (newBatchForm.current_quantity_remaining == null || Number(newBatchForm.current_quantity_remaining) < 0) {
    newBatchError.value = 'Enter quantity on hand.'
    return
  }
  saving.value = true
  newBatchError.value = ''
  try {
    await store.createNfBatch(fid, {
      input_definition_id: newBatchForm.input_definition_id,
      batch_identifier: newBatchForm.batch_identifier?.trim() || undefined,
      status: 'ready_for_use',
      current_quantity_remaining: Number(newBatchForm.current_quantity_remaining),
      low_stock_threshold: newBatchForm.low_stock_threshold ?? undefined,
    })
    batches.value = await store.loadNfBatches(fid)
    closeNewBatch()
    flashSuccess('Batch created.')
  } catch (e) {
    newBatchError.value = e.response?.data?.error || e.message || 'Could not create batch'
  } finally {
    saving.value = false
  }
}

function lowStockRowForBatch(row) {
  return lowStockRows.value.find((r) => r.batch.id === row.id)
}

function matchingLowStockAlert(row) {
  const name = row.inputName?.toLowerCase()
  return lowStockAlerts.value.find((a) => {
    const subj = String(a.subject_rendered || a.subject || '').toLowerCase()
    return name && subj.includes(name)
  })
}

async function createRefillTask(row) {
  const fid = farmContext.farmId
  if (!fid) return
  refillTaskSaving.value = row.batch.id
  clearActionFeedback()
  try {
    const alert = matchingLowStockAlert({
      inputName: row.inputName,
      id: row.batch.id,
    })
    if (alert) {
      await store.createTaskFromAlert(alert.id, buildRefillTaskPayload(row))
    } else {
      await store.createTask(fid, buildRefillTaskPayload(row))
    }
    flashSuccess('Refill task created — see Tasks.')
  } catch (e) {
    actionError.value = e.response?.data?.error || e.message || 'Could not create task'
  } finally {
    refillTaskSaving.value = null
  }
}

async function createRefillTaskForRow(row) {
  const ls = lowStockRowForBatch(row)
  if (!ls) return
  await createRefillTask(ls)
}

async function loadMixCounts(fid) {
  const counts = {}
  try {
    const mixEvents = await store.loadMixingEvents(fid)
    for (const me of mixEvents) {
      try {
        const comps = await store.loadMixingEventComponents(fid, me.id)
        for (const c of comps) {
          if (c.input_batch_id) counts[c.input_batch_id] = (counts[c.input_batch_id] || 0) + 1
        }
      } catch { /* skip */ }
    }
  } catch { /* skip */ }
  mixingComponentsByBatch.value = counts
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  clearActionFeedback()
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [i, b, r, a, p] = await Promise.all([
      store.loadNfInputs(fid),
      store.loadNfBatches(fid),
      store.loadRecipes(fid),
      store.loadAlerts(fid, { limit: 100 }),
      store.loadFertigationPrograms(fid),
    ])
    inputs.value = i
    batches.value = b
    recipes.value = r
    alerts.value = a
    programs.value = p
    await loadMixCounts(fid)
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>
