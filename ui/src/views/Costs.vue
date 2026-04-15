<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold text-white">Costs</h1>
      <button type="button" @click="reload" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
    </div>

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
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="createForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Income
          </label>
          <button type="submit" :disabled="saving" class="px-4 py-2 bg-green-700 text-white text-sm rounded-lg disabled:opacity-50">
            {{ saving ? 'Saving…' : 'Add' }}
          </button>
        </form>
        <p v-if="formError" class="text-xs text-red-400 mt-2">{{ formError }}</p>
      </div>

      <div class="overflow-x-auto border border-zinc-800 rounded-xl">
        <table class="w-full text-sm">
          <thead class="bg-zinc-900 text-zinc-500 text-xs uppercase">
            <tr>
              <th class="text-left px-4 py-3">Date</th>
              <th class="text-left px-4 py-3">Category</th>
              <th class="text-left px-4 py-3">Description</th>
              <th class="text-right px-4 py-3">Amount</th>
              <th class="text-left px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800">
            <tr v-for="t in transactions" :key="t.id" class="bg-zinc-950 hover:bg-zinc-900/50">
              <td class="px-4 py-2 text-zinc-300 whitespace-nowrap">{{ isoDate(t.transaction_date) }}</td>
              <td class="px-4 py-2">
                <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-800 text-zinc-300">{{ formatCat(t.category) }}</span>
              </td>
              <td class="px-4 py-2 text-zinc-400 max-w-xs truncate">{{ t.description || '—' }}</td>
              <td class="px-4 py-2 text-right font-mono tabular-nums"
                :class="t.is_income ? 'text-green-400' : 'text-red-400'">
                {{ t.is_income ? '+' : '−' }}{{ fmtMoney(Number(t.amount)) }} {{ t.currency }}
              </td>
              <td class="px-4 py-2">
                <button type="button" class="text-xs text-zinc-500 hover:text-zinc-300 mr-2" @click="startEdit(t)">Edit</button>
                <button type="button" class="text-xs text-red-500 hover:text-red-400" @click="removeTx(t)">Delete</button>
              </td>
            </tr>
            <tr v-if="!transactions.length">
              <td colspan="5" class="px-4 py-8 text-center text-zinc-500">No transactions yet.</td>
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
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="editForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Income
          </label>
          <div class="flex gap-2 justify-end">
            <button type="button" class="text-sm text-zinc-500" @click="editRow = null">Cancel</button>
            <button type="submit" :disabled="saving" class="px-3 py-1.5 bg-green-700 text-white text-sm rounded-lg">Save</button>
          </div>
        </form>
      </div>
    </template>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const loading = ref(true)
const saving = ref(false)
const formError = ref('')
const summary = reactive({ total_income: 0, total_expenses: 0, net: 0 })
const transactions = ref([])
const editRow = ref(null)
const editForm = reactive({
  transaction_date: '',
  category: 'miscellaneous',
  subcategory: '',
  amount: 0,
  currency: 'USD',
  description: '',
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
  is_income: false,
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

async function reload() {
  loading.value = true
  formError.value = ''
  try {
    const fid = farmContext.farmId
    if (!fid) return
    if (!store.zones.length) await store.loadAll(fid)
    const [s, tx] = await Promise.all([
      store.loadCostSummary(fid),
      store.loadCosts(fid, { limit: 100, offset: 0 }),
    ])
    Object.assign(summary, s || { total_income: 0, total_expenses: 0, net: 0 })
    transactions.value = tx
  } finally {
    loading.value = false
  }
}

async function submitCreate() {
  formError.value = ''
  saving.value = true
  try {
    await store.createCost(farmContext.farmId, {
      transaction_date: createForm.transaction_date,
      category: createForm.category,
      subcategory: createForm.subcategory.trim() || undefined,
      amount: createForm.amount,
      currency: createForm.currency.trim().toUpperCase(),
      description: createForm.description.trim() || undefined,
      is_income: createForm.is_income,
    })
    createForm.amount = 0
    createForm.description = ''
    createForm.subcategory = ''
    await reload()
  } catch (e) {
    formError.value = e.response?.data?.error || e.message
  } finally {
    saving.value = false
  }
}

function startEdit(t) {
  editRow.value = t
  editForm.transaction_date = isoDate(t.transaction_date)
  editForm.category = t.category
  editForm.subcategory = t.subcategory || ''
  editForm.amount = Number(t.amount)
  editForm.currency = t.currency || 'USD'
  editForm.description = t.description || ''
  editForm.is_income = !!t.is_income
}

async function submitEdit() {
  saving.value = true
  try {
    await store.updateCost(editRow.value.id, {
      transaction_date: editForm.transaction_date,
      category: editForm.category,
      subcategory: editForm.subcategory.trim() || undefined,
      amount: editForm.amount,
      currency: editForm.currency.trim().toUpperCase(),
      description: editForm.description.trim() || undefined,
      is_income: editForm.is_income,
    })
    editRow.value = null
    await reload()
  } finally {
    saving.value = false
  }
}

async function removeTx(t) {
  if (!confirm('Delete this transaction?')) return
  await store.deleteCost(t.id)
  await reload()
}

onMounted(reload)
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
