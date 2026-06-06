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

    <GuardianStarterChips :starters="moneyStarters" />

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
            <option v-for="c in spendCategories" :key="c.value" :value="c.value">{{ c.label }}</option>
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
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="receiptForm.is_income" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Money in (not an expense)
          </label>
          <button
            type="submit"
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

      <section class="space-y-3">
        <h2 class="text-xs text-zinc-500 uppercase tracking-widest">Recent activity</h2>
        <EmptyStateHint
          v-if="!recentRows.length"
          reason="no_data"
          message="No spending logged yet — save your first receipt above."
        />
        <div v-else class="space-y-2">
          <div
            v-for="row in recentRows"
            :key="row.clientCostId || row.id"
            class="flex items-center justify-between gap-3 bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3"
            :data-test="`money-row-${row.id || row.clientCostId}`"
          >
            <div class="min-w-0">
              <p class="text-sm text-zinc-200 truncate">{{ row.label }}</p>
              <p class="text-[11px] text-zinc-500 mt-0.5">
                {{ row.dateLabel }} · {{ row.categoryLabel }}
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
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import api from '../api.js'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import EmptyStateHint from '../components/EmptyStateHint.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { buildMoneyHubStarters } from '../lib/guardianStarters.js'
import {
  FARMER_SPEND_CATEGORIES,
  computeMonthSummary,
  buildRecentMoneyRows,
  formatMoney,
} from '../lib/moneyHub.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const saving = ref(false)
const formError = ref('')
const formSuccess = ref('')
const transactions = ref([])
const receiptFile = ref(null)
const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)

const spendCategories = FARMER_SPEND_CATEGORIES

const receiptForm = reactive({
  transaction_date: new Date().toISOString().slice(0, 10),
  amount: 0,
  category: 'miscellaneous',
  description: '',
  counterparty: '',
  currency: 'USD',
  is_income: false,
})

const moneyStarters = buildMoneyHubStarters()

const monthSummary = computed(() => computeMonthSummary(transactions.value))

const recentRows = computed(() => buildRecentMoneyRows(transactions.value))

const costPendingWrites = computed(() => {
  const fid = farmContext.farmId
  if (!fid) return 0
  return store.taskWriteQueue.filter((i) => i.farmId === fid && i.type === 'create_cost' && i.state !== 'synced').length
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
    await store.createCost(
      fid,
      {
        transaction_date: receiptForm.transaction_date,
        category: receiptForm.category,
        amount: receiptForm.amount,
        currency: receiptForm.currency.trim().toUpperCase(),
        description: receiptForm.description.trim() || undefined,
        counterparty: receiptForm.counterparty.trim() || undefined,
        is_income: receiptForm.is_income,
        document_type: receiptFile.value ? 'receipt' : undefined,
      },
      { receiptFile: receiptFile.value || undefined },
    )
    receiptFile.value = null
    receiptForm.amount = 0
    receiptForm.description = ''
    receiptForm.counterparty = ''
    formSuccess.value = 'Receipt saved.'
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
    transactions.value = await store.loadCosts(fid, { limit: 100, offset: 0 })
  } finally {
    loading.value = false
  }
}

function onConnectionChange() {
  isOnline.value = navigator.onLine
  if (isOnline.value) void store.flushTaskWriteQueue({ farmId: farmContext.farmId }).then(refresh)
}

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
