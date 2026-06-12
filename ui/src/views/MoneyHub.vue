<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-white">Money</h1>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          What you spent this month, save receipts, and review recent farm spending — without ledger jargon on the first screen.
        </p>
      </div>
      <button
        type="button"
        class="text-xs text-zinc-400 hover:text-zinc-200 shrink-0"
        @click="refresh"
      >
        Refresh
      </button>
    </div>

    <p v-if="!isOnline" class="text-xs text-amber-300 rounded-lg border border-amber-800/60 bg-amber-950/30 px-3 py-2">
      Offline — new receipts queue until you are back online.
    </p>
    <p v-if="costPendingWrites > 0" class="text-xs text-blue-300">
      {{ costPendingWrites }} receipt{{ costPendingWrites === 1 ? '' : 's' }} waiting to sync.
    </p>

    <div
      v-if="!energyPrices.length && !loading"
      class="rounded-lg border border-zinc-700 bg-zinc-950 px-4 py-3 flex flex-wrap items-center justify-between gap-2"
      data-test="money-energy-nudge"
    >
      <p class="text-xs text-zinc-400">
        No electricity price set — automatic power-cost logging stays off until you add a $/kWh rate.
      </p>
      <router-link
        v-nav-hint="'/costs'"
        to="/costs"
        class="text-xs text-green-500 hover:text-green-400 shrink-0"
      >
        Set energy price →
      </router-link>
    </div>

    <GuardianStarterChips :starters="moneyStarters" />

    <div
      v-if="filterCycleId"
      class="rounded-lg border border-emerald-800/60 bg-emerald-950/30 px-4 py-3 flex flex-wrap items-center justify-between gap-2"
      data-test="money-grow-filter"
    >
      <p class="text-xs text-emerald-200">
        Showing receipts tagged to grow #{{ filterCycleId }}
        <span v-if="filterCycleLabel" class="text-emerald-400">({{ filterCycleLabel }})</span>
      </p>
      <router-link
        to="/operations/money"
        class="text-xs text-zinc-400 hover:text-zinc-200"
      >
        Clear filter
      </router-link>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading money summary…</div>

    <template v-else>
      <section class="grid grid-cols-1 sm:grid-cols-3 gap-4" data-test="money-month-summary">
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">{{ monthSummary.monthLabel }} — spent</p>
          <p class="text-red-400 text-2xl font-mono tabular-nums">${{ formatMoney(monthSummary.expenses) }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">{{ monthSummary.monthLabel }} — received</p>
          <p class="text-green-400 text-2xl font-mono tabular-nums">${{ formatMoney(monthSummary.income) }}</p>
        </div>
        <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-zinc-500 text-xs mb-1">Net this month</p>
          <p class="text-white text-2xl font-mono tabular-nums">${{ formatMoney(monthSummary.net) }}</p>
          <p class="text-zinc-600 text-[11px] mt-1">{{ monthSummary.count }} transaction{{ monthSummary.count === 1 ? '' : 's' }}</p>
        </div>
      </section>

      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Save receipt</h2>
        <p class="text-xs text-zinc-500">Attach a photo or PDF and record what you paid — plain language, no chart of accounts.</p>
        <form class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3" @submit.prevent="submitReceipt">
          <input v-model="receiptForm.transaction_date" type="date" required class="input-field" />
          <input
            v-model.number="receiptForm.amount"
            type="number"
            step="0.01"
            min="0"
            placeholder="Amount"
            required
            class="input-field"
          />
          <select v-model="receiptForm.category" required class="input-field">
            <option v-for="c in receiptCategories" :key="c.value + c.label" :value="c.value">{{ c.label }}</option>
          </select>
          <input v-model="receiptForm.description" placeholder="What was this for?" class="input-field sm:col-span-2" />
          <input v-model="receiptForm.counterparty" placeholder="Vendor (optional)" class="input-field" />
          <label class="flex flex-col gap-1 text-zinc-400 text-xs sm:col-span-2">
            <span>Receipt photo or PDF</span>
            <input
              type="file"
              accept="application/pdf,image/jpeg,image/png,image/webp,.pdf"
              class="text-zinc-300 file:mr-2 file:rounded file:border-0 file:bg-zinc-800 file:px-2 file:py-1 file:text-xs"
              @change="onReceiptPick"
            />
          </label>
          <label class="flex items-center gap-2 text-zinc-300 text-sm sm:col-span-2">
            <input v-model="receiptForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Money in (sold harvest, grant, or other income)
          </label>
          <div class="sm:col-span-2 lg:col-span-3 space-y-2 border-t border-zinc-800 pt-3">
            <p class="text-[10px] uppercase tracking-widest text-zinc-500">Tag to a grow (optional)</p>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <select
                v-model.number="receiptForm.tagZoneId"
                class="input-field"
                data-test="money-tag-zone"
              >
                <option :value="null">No room</option>
                <option v-for="z in store.zones" :key="z.id" :value="z.id">{{ z.name }}</option>
              </select>
              <select
                v-model.number="receiptForm.tagCycleId"
                class="input-field"
                :disabled="!tagCycleOptions.length"
                data-test="money-tag-cycle"
              >
                <option :value="null">{{ tagCycleOptions.length ? 'Pick active grow' : 'No active grow in room' }}</option>
                <option v-for="c in tagCycleOptions" :key="c.id" :value="c.id">{{ formatCycleOptionLabel(c) }}</option>
              </select>
            </div>
          </div>
          <button
            type="submit"
            v-nav-hint="'/operations/money'"
            :disabled="saving"
            class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50"
            data-test="money-save-receipt"
          >
            {{ saving ? 'Saving…' : 'Save receipt' }}
          </button>
        </form>
        <p v-if="formError" class="text-xs text-red-400">{{ formError }}</p>
        <p v-if="formSuccess" class="text-xs text-green-400">{{ formSuccess }}</p>
      </section>

      <section v-if="autologRows.length" class="space-y-3" data-test="money-autolog-section">
        <h2 class="text-xs text-zinc-500 uppercase tracking-widest">Logged automatically</h2>
        <p class="text-[11px] text-zinc-600">Mixes, labor, supplies, and electricity the farm recorded for you.</p>
        <div class="space-y-2">
          <div
            v-for="row in autologRows"
            :key="row.id || row.clientCostId"
            class="flex items-center justify-between gap-3 bg-zinc-900/60 border border-zinc-800/80 rounded-xl px-4 py-3"
            :data-test="`money-autolog-${row.id || row.clientCostId}`"
          >
            <div class="min-w-0">
              <p class="text-sm text-zinc-300 truncate">{{ row.label }}</p>
              <p class="text-[11px] text-zinc-500 mt-0.5">
                {{ row.dateLabel }} · {{ row.categoryLabel }}
                <span class="text-zinc-600"> · auto</span>
              </p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span class="text-sm font-mono tabular-nums text-red-400/90">
                −${{ formatMoney(row.amount) }}
              </span>
              <router-link
                v-if="row.autologLink"
                v-nav-hint="row.autologLink.path"
                :to="row.autologLink"
                class="text-xs text-green-500 hover:text-green-400"
              >
                View →
              </router-link>
            </div>
          </div>
        </div>
      </section>

      <section class="space-y-3">
        <h2 class="text-xs text-zinc-500 uppercase tracking-widest">Your receipts</h2>
        <EmptyStateHint
          v-if="!manualRows.length"
          reason="no_data"
          message="No receipts saved yet — log spending above."
        />
        <div v-else class="space-y-2">
          <div
            v-for="row in manualRows"
            :key="row.clientCostId || row.id"
            class="flex items-center justify-between gap-3 bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3"
            :data-test="`money-row-${row.id || row.clientCostId}`"
          >
            <div class="min-w-0">
              <p class="text-sm text-zinc-200 truncate">{{ row.label }}</p>
              <p class="text-[11px] text-zinc-500 mt-0.5">
                {{ row.dateLabel }} · {{ row.categoryLabel }}
                <span v-if="row.cropCycleId" class="text-zinc-600"> · tagged to grow</span>
                <span v-if="row.queued" class="text-blue-300"> · queued</span>
                <span v-if="row.stale" class="text-amber-300"> · sync issue</span>
              </p>
            </div>
            <div class="flex items-center gap-2 shrink-0">
              <span
                class="text-sm font-mono tabular-nums"
                :class="row.isIncome ? 'text-green-400' : 'text-red-400'"
              >
                {{ row.isIncome ? '+' : '−' }}${{ formatMoney(row.amount) }}
              </span>
              <button
                v-if="row.hasReceipt"
                type="button"
                class="text-xs text-green-500 hover:text-green-400"
                @click="openReceipt(row.receiptFileId)"
              >
                Receipt
              </button>
              <router-link
                v-if="row.advancedLink"
                v-nav-hint="'/fertigation'"
                :to="row.advancedLink"
                class="text-xs text-zinc-500 hover:text-zinc-300"
              >
                Details →
              </router-link>
            </div>
          </div>
        </div>
      </section>
    </template>

    <footer class="pt-2 border-t border-zinc-800">
      <router-link
        v-nav-hint="'/operations/money'"
        to="/costs"
        class="text-xs text-zinc-400 hover:text-green-400"
        data-test="money-advanced-footer"
      >
        Full costs editor (GL mapping, exports, energy prices) →
      </router-link>
    </footer>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { buildMoneyHubStarters } from '../lib/guardianStarters.js'
import { loadDomainEnums } from '../lib/domainEnums.js'
import {
  FARMER_INCOME_CATEGORIES,
  spendCategoryOptions,
  computeMonthSummary,
  buildAutologMoneyRows,
  buildManualMoneyRows,
  formatMoney,
  activeCyclesForZone,
  formatCycleOptionLabel,
} from '../lib/moneyHub.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const route = useRoute()

const filterCycleId = computed(() => {
  const raw = route.query.cycle_id
  if (!raw) return null
  const n = Number(raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

const filterCycleLabel = computed(() => {
  if (!filterCycleId.value) return ''
  const c = cropCycles.value.find((row) => Number(row.id) === filterCycleId.value)
  return c?.name || cycleBatchLabel(c) || ''
})

const loading = ref(false)
const saving = ref(false)
const formError = ref('')
const formSuccess = ref('')
const transactions = ref([])
const cropCycles = ref([])
const energyPrices = ref([])
const domainEnums = ref(null)
const receiptFile = ref(null)
const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)

const receiptForm = reactive({
  transaction_date: new Date().toISOString().slice(0, 10),
  amount: 0,
  category: 'miscellaneous',
  description: '',
  counterparty: '',
  currency: 'USD',
  is_income: false,
  tagZoneId: null,
  tagCycleId: null,
})

const moneyStarters = computed(() => buildMoneyHubStarters())

const monthSummary = computed(() => computeMonthSummary(transactions.value))

const autologRows = computed(() => buildAutologMoneyRows(transactions.value))

const manualRows = computed(() => buildManualMoneyRows(transactions.value))

const receiptCategories = computed(() =>
  receiptForm.is_income ? FARMER_INCOME_CATEGORIES : spendCategoryOptions(domainEnums.value),
)

const tagCycleOptions = computed(() => {
  if (!receiptForm.tagZoneId) return []
  return activeCyclesForZone(cropCycles.value, receiptForm.tagZoneId)
})

const costPendingWrites = computed(() => {
  const fid = farmContext.farmId
  if (!fid) return 0
  return store.taskWriteQueue.filter((i) => i.farmId === fid && i.type === 'create_cost' && i.state !== 'synced').length
})

watch(() => receiptForm.tagZoneId, () => {
  receiptForm.tagCycleId = tagCycleOptions.value[0]?.id ?? null
})

watch(() => receiptForm.is_income, (income) => {
  receiptForm.category = 'miscellaneous'
  if (!income && receiptForm.description.toLowerCase().includes('sold')) {
    receiptForm.description = ''
  }
})

function onReceiptPick(e) {
  receiptFile.value = e.target.files?.[0] ?? null
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

async function submitReceipt() {
  formError.value = ''
  formSuccess.value = ''
  saving.value = true
  try {
    const fid = farmContext.farmId
    if (!fid) return
    const payload = {
      transaction_date: receiptForm.transaction_date,
      category: receiptForm.category,
      amount: receiptForm.amount,
      currency: receiptForm.currency.trim().toUpperCase(),
      description: receiptForm.description.trim() || undefined,
      counterparty: receiptForm.counterparty.trim() || undefined,
      is_income: receiptForm.is_income,
      document_type: receiptFile.value ? 'receipt' : undefined,
    }
    const taggedCycleId = receiptForm.tagCycleId
    if (taggedCycleId) {
      payload.crop_cycle_id = taggedCycleId
    }
    await store.createCost(fid, payload, { receiptFile: receiptFile.value || undefined })
    receiptFile.value = null
    receiptForm.amount = 0
    receiptForm.description = ''
    receiptForm.counterparty = ''
    receiptForm.tagCycleId = null
    formSuccess.value = taggedCycleId
      ? 'Receipt saved and tagged to this grow.'
      : 'Receipt saved.'
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Could not save receipt'
  } finally {
    saving.value = false
  }
}

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  formError.value = ''
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const [costs, cycles, prices, enums] = await Promise.all([
      store.loadCosts(fid, {
        limit: 100,
        offset: 0,
        cropCycleId: filterCycleId.value,
      }),
      store.loadCropCycles(fid),
      api.get(`/farms/${fid}/energy-prices`).then((r) => r.data).catch(() => []),
      loadDomainEnums(api),
    ])
    transactions.value = costs
    cropCycles.value = cycles
    energyPrices.value = Array.isArray(prices) ? prices : []
    domainEnums.value = enums
  } finally {
    loading.value = false
  }
}

function onConnectionChange() {
  isOnline.value = navigator.onLine
  if (isOnline.value) void store.flushTaskWriteQueue({ farmId: farmContext.farmId }).then(refresh)
}

watch(filterCycleId, () => {
  void refresh()
})

onMounted(() => {
  refresh()
  if (typeof window !== 'undefined') {
    window.addEventListener('online', onConnectionChange)
    window.addEventListener('offline', onConnectionChange)
  }
})

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('online', onConnectionChange)
    window.removeEventListener('offline', onConnectionChange)
  }
})
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
