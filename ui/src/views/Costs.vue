<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Costs</h1>
        <p
          v-if="syncStatusText"
          class="mt-1 text-[11px]"
          :class="syncStatusClass"
        >
          {{ syncStatusText }}
        </p>
      </div>
      <div class="flex items-center gap-2 flex-wrap justify-end">
        <span
          v-if="!isOnline"
          class="text-[11px] px-2 py-1 rounded border border-amber-700 bg-amber-900/40 text-amber-300"
        >
          Offline mode
        </span>
        <span
          v-if="costPendingWrites > 0"
          class="text-[11px] px-2 py-1 rounded border border-blue-700 bg-blue-900/40 text-blue-300"
        >
          {{ costPendingWrites }} queued cost{{ costPendingWrites === 1 ? '' : 's' }}
        </span>
        <button
          type="button"
          v-if="costPendingWrites > 0 || costQueueHasStale"
          :disabled="syncing"
          @click="syncNow"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-800 text-zinc-200 border border-zinc-700 hover:bg-zinc-700 disabled:opacity-40"
        >
          {{ syncing ? 'Syncing…' : 'Sync now' }}
        </button>
        <button
          type="button"
          v-if="costQueueItems.length > 0"
          @click="showQueueDetails = true"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-zinc-900 text-zinc-200 border border-zinc-700 hover:bg-zinc-800"
        >
          Queue details
        </button>
        <button type="button" @click="downloadCsv" :disabled="!farmContext.farmId"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 disabled:opacity-40">
          Export CSV
        </button>
        <button type="button" @click="downloadGlCsv" :disabled="!farmContext.farmId"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 disabled:opacity-40">
          Export GL CSV
        </button>
        <span class="flex items-center gap-1">
          <input
            v-model="summaryExportYear"
            type="number"
            min="1900"
            max="2100"
            placeholder="Year"
            class="w-[4.5rem] px-2 py-1.5 rounded-lg border border-zinc-700 bg-zinc-950 text-zinc-200 text-xs placeholder:text-zinc-600"
          />
          <button type="button" @click="downloadSummaryCsv" :disabled="!farmContext.farmId"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 disabled:opacity-40">
            Summary CSV
          </button>
        </span>
        <button type="button" @click="reload" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
      </div>
    </div>

    <p v-if="costQueueHasStale" class="text-xs text-amber-300">
      Some queued cost rows need review after a sync error. Retry or discard from the row actions or queue details.
    </p>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading…</div>

    <template v-else>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">Total income</p>
          <p class="text-green-400 text-2xl font-mono tabular-nums">{{ fmtMoney(summary.total_income) }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">Total expenses</p>
          <p class="text-red-400 text-2xl font-mono tabular-nums">{{ fmtMoney(summary.total_expenses) }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">Net</p>
          <p class="text-white text-2xl font-mono tabular-nums">{{ fmtMoney(summary.net) }}</p>
        </div>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-white">Add transaction</h2>
        </div>
        <form class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3" @submit.prevent="submitCreate">
          <input v-model="createForm.transaction_date" type="date" required class="input-field" />
          <select v-model="createForm.category" required class="input-field">
            <option v-for="c in costCategories" :key="c" :value="c">{{ formatCat(c) }}</option>
          </select>
          <input v-model="createForm.subcategory" placeholder="Subcategory" class="input-field" />
          <input v-model.number="createForm.amount" type="number" step="0.01" placeholder="Amount" required class="input-field" />
          <input v-model="createForm.currency" maxlength="3" placeholder="USD" required class="input-field uppercase" />
          <input v-model="createForm.description" placeholder="Description" class="input-field sm:col-span-2" />
          <input v-model="createForm.document_type" placeholder="Doc type (invoice, bill…)" class="input-field" />
          <input v-model="createForm.document_reference" placeholder="Invoice / ref #" class="input-field" />
          <input v-model="createForm.counterparty" placeholder="Vendor / customer" class="input-field sm:col-span-2" />
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="createForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Income
          </label>
          <label class="flex flex-col gap-1 text-zinc-400 text-xs sm:col-span-2">
            <span>Receipt (PDF or image, max 5 MB)</span>
            <input type="file" accept="application/pdf,image/jpeg,image/png,image/webp,.pdf"
              class="text-zinc-300 file:mr-2 file:rounded file:border-0 file:bg-zinc-800 file:px-2 file:py-1 file:text-xs"
              @change="onCreateReceiptPick" />
          </label>
          <button type="submit" :disabled="saving" class="px-4 py-2 bg-green-700 text-white text-sm rounded-lg disabled:opacity-50">
            {{ saving ? 'Saving…' : 'Add' }}
          </button>
        </form>
        <p v-if="formError" class="text-xs text-red-400 mt-2">{{ formError }}</p>
      </div>

      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold text-white">GL account mapping</h2>
          <div class="flex items-center gap-2">
            <button
              type="button"
              :disabled="coaSaving || !farmContext.farmId"
              @click="resetCoaAll"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-400 hover:border-zinc-500 disabled:opacity-40"
            >
              Reset all defaults
            </button>
            <button
              type="button"
              :disabled="coaSaving || !farmContext.farmId"
              @click="saveCoa"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-300 hover:border-zinc-500 disabled:opacity-40"
            >
              {{ coaSaving ? 'Saving…' : 'Save mapping' }}
            </button>
          </div>
        </div>
        <p class="text-xs text-zinc-500 mb-3">
          These account codes are used by `Export GL CSV` and can be customized per farm.
        </p>
        <div class="overflow-x-auto border border-zinc-800 rounded-lg">
          <table class="w-full text-xs">
            <thead class="bg-zinc-950 text-zinc-500 uppercase">
              <tr>
                <th class="text-left px-3 py-2">Category</th>
                <th class="text-left px-3 py-2">Account code</th>
                <th class="text-left px-3 py-2">Account name</th>
                <th class="text-left px-3 py-2">Source</th>
                <th class="text-left px-3 py-2">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-zinc-800">
              <tr v-for="m in coaMappings" :key="m.category" class="bg-zinc-900">
                <td class="px-3 py-2 text-zinc-300">{{ formatCat(m.category) }}</td>
                <td class="px-3 py-2">
                  <input v-model="m.account_code" class="input-field w-28" />
                </td>
                <td class="px-3 py-2">
                  <input v-model="m.account_name" class="input-field w-full min-w-[220px]" />
                </td>
                <td class="px-3 py-2">
                  <span class="text-[11px] px-2 py-0.5 rounded bg-zinc-800 text-zinc-400">{{ m.source }}</span>
                </td>
                <td class="px-3 py-2">
                  <button
                    type="button"
                    :disabled="coaSaving || m.source !== 'override'"
                    @click="resetCoaCategory(m.category)"
                    class="text-[11px] px-2 py-1 rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200 disabled:opacity-40"
                  >
                    Reset
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-if="coaError" class="text-xs text-red-400 mt-2">{{ coaError }}</p>
      </div>

      <div class="overflow-x-auto border border-zinc-800 rounded-xl">
        <table class="w-full text-sm">
          <thead class="bg-zinc-900 text-zinc-500 text-xs uppercase">
            <tr>
              <th class="text-left px-4 py-3">Date</th>
              <th class="text-left px-4 py-3">Category</th>
              <th class="text-left px-4 py-3">Description</th>
              <th class="text-left px-4 py-3 text-[11px]">Bookkeeping</th>
              <th class="text-right px-4 py-3">Amount</th>
              <th class="text-left px-4 py-3">Receipt</th>
              <th class="text-left px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800">
            <tr v-for="t in transactions" :key="t._offline?.clientCostId || t.id" class="bg-zinc-950 hover:bg-zinc-900/50">
              <td class="px-4 py-2 text-zinc-300 whitespace-nowrap">{{ isoDate(t.transaction_date) }}</td>
              <td class="px-4 py-2">
                <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-800 text-zinc-300">{{ formatCat(t.category) }}</span>
              </td>
              <td class="px-4 py-2 text-zinc-400 max-w-xs truncate">
                <span v-if="t._offline?.queued" class="block text-[11px] mb-0.5"
                  :class="t._offline?.stale ? 'text-amber-300' : 'text-blue-300'">
                  {{ t._offline?.stale ? `Sync: ${t._offline?.conflict || 'retry'}` : 'Queued' }}
                  <span v-if="t._offline?.receiptPending" class="text-zinc-500"> · receipt pending</span>
                </span>
                {{ t.description || '—' }}
              </td>
              <td class="px-4 py-2 text-[11px] text-zinc-500 max-w-[10rem]">
                <template v-if="t.document_type || t.document_reference || t.counterparty">
                  <span v-if="t.document_type" class="text-zinc-400">{{ t.document_type }}</span>
                  <span v-if="t.document_reference" class="block truncate" :title="t.document_reference">#{{ t.document_reference }}</span>
                  <span v-if="t.counterparty" class="block truncate text-zinc-600" :title="t.counterparty">{{ t.counterparty }}</span>
                </template>
                <span v-else>—</span>
              </td>
              <td class="px-4 py-2 text-right font-mono tabular-nums"
                :class="t.is_income ? 'text-green-400' : 'text-red-400'">
                {{ t.is_income ? '+' : '−' }}{{ fmtMoney(Number(t.amount)) }} {{ t.currency }}
              </td>
              <td class="px-4 py-2">
                <button v-if="t.receipt_file_id" type="button"
                  class="text-xs text-green-500 hover:text-green-400"
                  @click="openReceipt(t.receipt_file_id)">
                  View
                </button>
                <span v-else class="text-zinc-600">—</span>
              </td>
              <td class="px-4 py-2">
                <template v-if="t._offline?.queueItemId">
                  <button type="button" class="text-[11px] px-2 py-1 rounded bg-zinc-800 border border-zinc-700 text-zinc-200 mr-1" @click="retryCostQueue(t._offline.queueItemId)">Retry</button>
                  <button type="button" class="text-[11px] px-2 py-1 rounded bg-zinc-900 border border-zinc-700 text-zinc-400" @click="discardCostQueue(t._offline.queueItemId)">Discard</button>
                </template>
                <template v-else>
                  <button type="button" class="text-xs text-zinc-500 hover:text-zinc-300 mr-2" @click="startEdit(t)">Edit</button>
                  <button type="button" class="text-xs text-red-500 hover:text-red-400" @click="removeTx(t)">Delete</button>
                </template>
              </td>
            </tr>
            <tr v-if="!transactions.length">
              <td colspan="7" class="px-4 py-8 text-center text-zinc-500">No transactions yet.</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="editRow" class="fixed inset-0 bg-black/60 flex items-center justify-center p-4 z-50" @click.self="editRow = null">
        <form class="bg-zinc-900 border border-zinc-700 rounded-xl p-4 w-full max-w-md space-y-3" @submit.prevent="submitEdit">
          <h3 class="text-white font-medium">Edit transaction</h3>
          <input v-model="editForm.transaction_date" type="date" required class="input-field w-full" />
          <select v-model="editForm.category" required class="input-field w-full">
            <option v-for="c in costCategories" :key="c" :value="c">{{ formatCat(c) }}</option>
          </select>
          <input v-model="editForm.subcategory" placeholder="Subcategory" class="input-field w-full" />
          <input v-model.number="editForm.amount" type="number" step="0.01" required class="input-field w-full" />
          <input v-model="editForm.currency" maxlength="3" required class="input-field w-full uppercase" />
          <input v-model="editForm.description" placeholder="Description" class="input-field w-full" />
          <input v-model="editForm.document_type" placeholder="Doc type" class="input-field w-full" />
          <input v-model="editForm.document_reference" placeholder="Invoice / ref #" class="input-field w-full" />
          <input v-model="editForm.counterparty" placeholder="Vendor / customer" class="input-field w-full" />
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="editForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Income
          </label>
          <label class="flex flex-col gap-1 text-zinc-400 text-xs">
            <span>Replace receipt (optional)</span>
            <input type="file" accept="application/pdf,image/jpeg,image/png,image/webp,.pdf"
              class="text-zinc-300 file:mr-2 file:rounded file:border-0 file:bg-zinc-800 file:px-2 file:py-1 file:text-xs"
              @change="onEditReceiptPick" />
          </label>
          <div class="flex gap-2 justify-end">
            <button type="button" class="text-sm text-zinc-500" @click="editRow = null">Cancel</button>
            <button type="submit" :disabled="saving" class="px-3 py-1.5 bg-green-700 text-white text-sm rounded-lg">Save</button>
          </div>
        </form>
      </div>

      <div
        v-if="showQueueDetails"
        class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
        @click.self="showQueueDetails = false"
      >
        <div class="w-full max-w-2xl bg-zinc-900 border border-zinc-700 rounded-xl p-4">
          <div class="flex items-center justify-between mb-3">
            <h2 class="text-sm font-semibold text-white">Queued cost writes</h2>
            <button type="button" class="text-xs text-zinc-400 hover:text-zinc-200" @click="showQueueDetails = false">Close</button>
          </div>
          <p v-if="costQueueItems.length === 0" class="text-xs text-zinc-500">No queued costs.</p>
          <div v-else class="space-y-2 max-h-[60vh] overflow-auto pr-1">
            <div v-for="item in costQueueItems" :key="item.id" class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs font-medium text-zinc-100">{{ costQueueLabel(item) }}</span>
                <span class="text-[11px] px-2 py-0.5 rounded" :class="costQueueStateClass(item.state)">{{ item.state }}</span>
              </div>
              <p class="text-[11px] text-zinc-400 mb-1">attempts: {{ item.attempts }} · updated: {{ formatQueueTime(item.updatedAt) }}</p>
              <p v-if="item.lastError" class="text-[11px] text-amber-300 mb-2">{{ item.lastError }}</p>
              <div class="flex gap-2">
                <button type="button" class="text-[11px] px-2 py-1 rounded bg-zinc-800 border border-zinc-700 text-zinc-200" @click="retryCostQueue(item.id)">Retry</button>
                <button type="button" class="text-[11px] px-2 py-1 rounded bg-zinc-900 border border-zinc-700 text-zinc-400" @click="discardCostQueue(item.id)">Discard</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import api from '../api'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
const syncing = ref(false)
const showQueueDetails = ref(false)
const loading = ref(true)
const saving = ref(false)
const formError = ref('')
const summary = reactive({ total_income: 0, total_expenses: 0, net: 0 })
const transactions = ref([])
const editRow = ref(null)
const createReceiptFile = ref(null)
const editReceiptFile = ref(null)
const coaMappings = ref([])
const coaSaving = ref(false)
const coaError = ref('')
const editForm = reactive({
  transaction_date: '',
  category: 'miscellaneous',
  subcategory: '',
  amount: 0,
  currency: 'USD',
  description: '',
  document_type: '',
  document_reference: '',
  counterparty: '',
  is_income: false,
})

const costCategories = [
  'seeds_plants', 'fertilizers_soil_amendments', 'pest_disease_control', 'water_irrigation',
  'labor_wages', 'equipment_purchase_rental', 'equipment_maintenance_fuel', 'utilities_electricity_gas',
  'land_rent_mortgage', 'insurance', 'licenses_permits', 'feed_livestock', 'veterinary_services',
  'packaging_supplies', 'transportation_logistics', 'marketing_sales', 'training_consultancy', 'miscellaneous',
]

const createForm = reactive({
  transaction_date: new Date().toISOString().slice(0, 10),
  category: 'miscellaneous',
  subcategory: '',
  amount: 0,
  currency: 'USD',
  description: '',
  document_type: '',
  document_reference: '',
  counterparty: '',
  is_income: false,
})
const summaryExportYear = ref('')

const costPendingWrites = computed(() => {
  const fid = farmContext.farmId
  if (!fid) return 0
  return store.taskWriteQueue.filter((i) => i.farmId === fid && i.type === 'create_cost' && i.state !== 'synced').length
})
const costQueueItems = computed(() => {
  const fid = farmContext.farmId
  return store.taskWriteQueue.filter((i) => i.farmId === fid && i.type === 'create_cost')
})
const costQueueHasStale = computed(() =>
  transactions.value.some((t) => t._offline?.stale && t._offline?.clientCostId),
)
const syncStatusText = computed(() => {
  const status = store.taskSyncStatus
  if (!status?.lastAttemptAt) return ''
  const when = new Date(status.lastAttemptAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (status.lastResult === 'running') return `Last sync ${when}: running`
  if (status.lastResult === 'partial_error') return `Last sync ${when}: ${status.lastMessage || 'needs review'}`
  if (status.lastResult === 'ok') return `Last sync ${when}: ${status.lastMessage || 'ok'}`
  return `Last sync ${when}`
})
const syncStatusClass = computed(() => {
  const result = store.taskSyncStatus?.lastResult
  if (result === 'partial_error') return 'text-amber-300'
  if (result === 'ok') return 'text-emerald-300'
  return 'text-zinc-400'
})

function fmtMoney(n) {
  if (n == null || Number.isNaN(n)) return '0.00'
  return Number(n).toFixed(2)
}

function isoDate(d) {
  if (!d) return '—'
  if (typeof d === 'string') return d.slice(0, 10)
  if (d.Time) return String(d.Time).slice(0, 10)
  return '—'
}

function formatCat(c) {
  return c ? c.replace(/_/g, ' ') : ''
}

function onCreateReceiptPick(e) {
  createReceiptFile.value = e.target.files?.[0] ?? null
}

function onEditReceiptPick(e) {
  editReceiptFile.value = e.target.files?.[0] ?? null
}

function costQueueLabel(item) {
  const p = item.payload || {}
  return `Cost ${p.category || '?'} · ${p.amount ?? ''} ${p.currency || ''}`.trim()
}
function costQueueStateClass(state) {
  if (state === 'failed') return 'bg-amber-900/40 border border-amber-700 text-amber-300'
  if (state === 'pending') return 'bg-blue-900/40 border border-blue-700 text-blue-300'
  return 'bg-zinc-800 text-zinc-300'
}
function formatQueueTime(value) {
  if (!value) return 'n/a'
  try {
    return new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch {
    return value
  }
}

async function syncNow() {
  const fid = farmContext.farmId
  if (!fid || syncing.value) return
  syncing.value = true
  try {
    await store.flushTaskWriteQueue({ farmId: fid })
    await reload()
  } finally {
    syncing.value = false
  }
}

function onConnectionChange() {
  isOnline.value = navigator.onLine
  if (isOnline.value) void syncNow()
}

async function retryCostQueue(queueItemId) {
  if (!queueItemId) return
  store.retryTaskQueueItem(queueItemId)
  await syncNow()
}

async function discardCostQueue(queueItemId) {
  if (!queueItemId) return
  store.discardTaskQueueItem(queueItemId)
  showQueueDetails.value = false
  await reload()
}

async function openReceipt(fileId) {
  try {
    const r = await api.get(`/file-attachments/${fileId}/download`)
    const url = String(r.data?.url || '')
    if (!url) throw new Error('Missing receipt URL')
    const finalUrl = url.startsWith('http://') || url.startsWith('https://')
      ? url
      : `${api.defaults.baseURL}${url}`
    window.open(finalUrl, '_blank', 'noopener')
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Could not open receipt'
  }
}

async function downloadCsv() {
  const fid = farmContext.farmId
  if (!fid) return
  try {
    const r = await api.get(`/farms/${fid}/costs/export`, {
      params: { format: 'csv' },
      responseType: 'blob',
    })
    const url = URL.createObjectURL(r.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `farm-${fid}-costs.csv`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Export failed'
  }
}

async function downloadGlCsv() {
  const fid = farmContext.farmId
  if (!fid) return
  try {
    const r = await api.get(`/farms/${fid}/costs/export`, {
      params: { format: 'gl_csv' },
      responseType: 'blob',
    })
    const url = URL.createObjectURL(r.data)
    const a = document.createElement('a')
    a.href = url
    a.download = `farm-${fid}-costs-gl.csv`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'GL export failed'
  }
}

async function downloadSummaryCsv() {
  const fid = farmContext.farmId
  if (!fid) return
  const params = { format: 'summary_csv' }
  const y = String(summaryExportYear.value ?? '').trim()
  if (y) params.year = y
  try {
    const r = await api.get(`/farms/${fid}/costs/export`, {
      params,
      responseType: 'blob',
    })
    const url = URL.createObjectURL(r.data)
    const a = document.createElement('a')
    a.href = url
    const suffix = y ? `-${y}` : ''
    a.download = `farm-${fid}-costs-summary${suffix}.csv`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Summary export failed'
  }
}

async function reload() {
  loading.value = true
  formError.value = ''
  try {
    const fid = farmContext.farmId
    if (!fid) return
    if (!store.zones.length) await store.loadAll(fid)
    const [s, tx, coa] = await Promise.all([
      store.loadCostSummary(fid),
      store.loadCosts(fid, { limit: 100, offset: 0 }),
      store.loadCoaMappings(fid),
    ])
    Object.assign(summary, s || { total_income: 0, total_expenses: 0, net: 0 })
    transactions.value = tx
    coaMappings.value = coa
  } finally {
    loading.value = false
  }
}

async function saveCoa() {
  const fid = farmContext.farmId
  if (!fid) return
  coaError.value = ''
  coaSaving.value = true
  try {
    coaMappings.value = await store.saveCoaMappings(fid, coaMappings.value.map((m) => ({
      category: m.category,
      account_code: String(m.account_code || '').trim(),
      account_name: String(m.account_name || '').trim(),
    })))
  } catch (e) {
    coaError.value = e.response?.data?.error || e.message || 'Could not save mapping'
  } finally {
    coaSaving.value = false
  }
}

async function resetCoaCategory(category) {
  const fid = farmContext.farmId
  if (!fid || !category) return
  coaError.value = ''
  coaSaving.value = true
  try {
    coaMappings.value = await store.resetCoaMappingCategory(fid, category)
  } catch (e) {
    coaError.value = e.response?.data?.error || e.message || 'Could not reset category mapping'
  } finally {
    coaSaving.value = false
  }
}

async function resetCoaAll() {
  const fid = farmContext.farmId
  if (!fid) return
  if (!confirm('Reset all GL mappings back to defaults for this farm?')) return
  coaError.value = ''
  coaSaving.value = true
  try {
    coaMappings.value = await store.resetCoaMappingsAll(fid)
  } catch (e) {
    coaError.value = e.response?.data?.error || e.message || 'Could not reset mappings'
  } finally {
    coaSaving.value = false
  }
}

async function submitCreate() {
  formError.value = ''
  saving.value = true
  try {
    const fid = farmContext.farmId
    const receipt = createReceiptFile.value
    const row = await store.createCost(
      fid,
      {
        transaction_date: createForm.transaction_date,
        category: createForm.category,
        subcategory: createForm.subcategory.trim() || undefined,
        amount: createForm.amount,
        currency: createForm.currency.trim().toUpperCase(),
        description: createForm.description.trim() || undefined,
        is_income: createForm.is_income,
        document_type: createForm.document_type.trim() || undefined,
        document_reference: createForm.document_reference.trim() || undefined,
        counterparty: createForm.counterparty.trim() || undefined,
      },
      { receiptFile: receipt || undefined },
    )
    createReceiptFile.value = null
    createForm.amount = 0
    createForm.description = ''
    createForm.subcategory = ''
    createForm.document_type = ''
    createForm.document_reference = ''
    createForm.counterparty = ''
    await reload()
  } catch (e) {
    formError.value = e.response?.data?.error || e.message
  } finally {
    saving.value = false
  }
}

function startEdit(t) {
  editRow.value = t
  editReceiptFile.value = null
  editForm.transaction_date = isoDate(t.transaction_date)
  editForm.category = t.category
  editForm.subcategory = t.subcategory || ''
  editForm.amount = Number(t.amount)
  editForm.currency = t.currency || 'USD'
  editForm.description = t.description || ''
  editForm.document_type = t.document_type || ''
  editForm.document_reference = t.document_reference || ''
  editForm.counterparty = t.counterparty || ''
  editForm.is_income = !!t.is_income
}

async function submitEdit() {
  saving.value = true
  try {
    const fid = farmContext.farmId
    await store.updateCost(editRow.value.id, {
      transaction_date: editForm.transaction_date,
      category: editForm.category,
      subcategory: editForm.subcategory.trim() || undefined,
      amount: editForm.amount,
      currency: editForm.currency.trim().toUpperCase(),
      description: editForm.description.trim() || undefined,
      is_income: editForm.is_income,
      document_type: editForm.document_type.trim(),
      document_reference: editForm.document_reference.trim(),
      counterparty: editForm.counterparty.trim(),
    })
    if (editReceiptFile.value) {
      await store.uploadCostReceipt(fid, editReceiptFile.value, editRow.value.id)
      editReceiptFile.value = null
    }
    editRow.value = null
    await reload()
  } finally {
    saving.value = false
  }
}

async function removeTx(t) {
  if (t._offline?.queueItemId) {
    if (!confirm('Discard this queued cost? It was not saved on the server yet.')) return
    store.discardTaskQueueItem(t._offline.queueItemId)
    await reload()
    return
  }
  if (!confirm('Delete this transaction?')) return
  await store.deleteCost(t.id)
  await reload()
}

onMounted(() => {
  window.addEventListener('online', onConnectionChange)
  window.addEventListener('offline', onConnectionChange)
  void reload()
})

onUnmounted(() => {
  window.removeEventListener('online', onConnectionChange)
  window.removeEventListener('offline', onConnectionChange)
})
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
