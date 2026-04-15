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
    <div v-else-if="activeTab === 'definitions'" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
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
        <div v-if="input.typical_ingredients" class="text-xs text-zinc-500">
          <span class="text-zinc-600 uppercase tracking-wide">Ingredients:</span>
          {{ input.typical_ingredients }}
        </div>
      </div>
      <div v-if="!inputs.length" class="col-span-full text-zinc-500 text-sm text-center py-8">
        No input definitions found.
      </div>
    </div>

    <!-- Batches tab -->
    <div v-else class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-left text-zinc-500 text-xs uppercase tracking-wide border-b border-zinc-700">
            <th class="pb-2 pr-4">Batch</th>
            <th class="pb-2 pr-4">Input</th>
            <th class="pb-2 pr-4">Status</th>
            <th class="pb-2 pr-4">Qty Remaining</th>
            <th class="pb-2 pr-4">Started</th>
            <th class="pb-2">Storage</th>
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
            <td class="py-2.5 text-zinc-400">{{ b.storage_location ?? '—' }}</td>
          </tr>
          <tr v-if="!batches.length">
            <td colspan="6" class="text-zinc-500 text-center py-8">No batches found.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'

const store = useFarmStore()

const tabs = [
  { key: 'definitions', label: 'Input Definitions' },
  { key: 'batches',     label: 'Batches' },
]
const activeTab = ref('definitions')
const loading   = ref(true)
const inputs    = ref([])
const batches   = ref([])

onMounted(async () => {
  try {
    const [i, b] = await Promise.all([
      store.loadNfInputs(),
      store.loadNfBatches(),
    ])
    inputs.value  = i
    batches.value = b
  } finally {
    loading.value = false
  }
})

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
