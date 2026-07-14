<template>
  <div :class="embedded ? 'p-4 sm:p-6 max-w-4xl mx-auto space-y-6' : 'p-6 max-w-4xl'">
    <header v-if="!embedded" class="space-y-2">
      <h1 class="text-2xl font-bold text-green-400">Symptom guide</h1>
      <p class="text-sm text-zinc-500">
        Browse the agronomy symptom catalog — filter by crop or category. Linked from Farm Guardian citations.
      </p>
    </header>
    <p v-else class="text-xs text-zinc-500 leading-relaxed">
      Crop symptom lookup — filter by crop or category. Guardian citations deep-link here with filters applied.
    </p>

    <div class="flex flex-wrap gap-3 items-end" data-test="symptom-guide-filters">
      <div>
        <label class="block text-xs text-zinc-500 mb-1" for="symptom-crop">Crop</label>
        <select
          id="symptom-crop"
          v-model="cropKey"
          class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white min-w-[10rem]"
          data-test="symptom-crop-select"
          @change="onFilterChange"
        >
          <option value="">All crops</option>
          <option v-for="c in cropOptions" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-zinc-500 mb-1" for="symptom-category">Category</label>
        <select
          id="symptom-category"
          v-model="category"
          class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white min-w-[10rem]"
          data-test="symptom-category-select"
          @change="onFilterChange"
        >
          <option value="">All categories</option>
          <option v-for="c in categoryOptions" :key="c" :value="c">{{ c }}</option>
        </select>
      </div>
      <button
        type="button"
        class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800"
        data-test="symptom-guide-search"
        @click="load"
      >
        Search
      </button>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading…</div>
    <div v-else-if="error" class="text-red-400 text-sm">{{ error }}</div>
    <div v-else-if="symptoms.length === 0" class="text-zinc-500 text-sm">No symptoms match your filters.</div>
    <div v-else class="space-y-3" data-test="symptom-guide-list">
      <article
        v-for="s in symptoms"
        :key="s.id || s.symptom_key"
        class="bg-zinc-800 border border-zinc-700 rounded-xl p-4"
      >
        <div class="flex flex-wrap items-start justify-between gap-2 mb-2">
          <h2 class="text-white text-sm font-semibold">{{ s.display_name || s.symptom_key }}</h2>
          <span v-if="primaryCategory(s)" class="text-[10px] uppercase tracking-wide text-zinc-500">{{ primaryCategory(s) }}</span>
        </div>
        <p class="text-[11px] text-zinc-500 mb-1">Crops: {{ formatCropKeys(s.crop_keys) }}</p>
        <p v-if="s.body_md" class="text-zinc-300 text-sm whitespace-pre-wrap">{{ s.body_md }}</p>
        <p v-if="s.severity_hint" class="text-zinc-400 text-xs mt-2">
          <span class="text-zinc-500">Severity:</span> {{ s.severity_hint }}
        </p>
      </article>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { formatCropKeys, uniqueCategories, uniqueCropKeys } from '../lib/symptomGuideFilters.js'

defineProps({
  embedded: { type: Boolean, default: false },
})

const route = useRoute()
const router = useRouter()
const cropKey = ref('')
const category = ref('')
const symptoms = ref([])
const catalog = ref([])
const loading = ref(false)
const error = ref('')

const cropOptions = computed(() => uniqueCropKeys(catalog.value))
const categoryOptions = computed(() => uniqueCategories(catalog.value))

function primaryCategory(s) {
  const cats = s.categories || []
  return cats.length ? cats[0] : ''
}

function syncFromRoute() {
  cropKey.value = route.query.crop_key ? String(route.query.crop_key) : ''
  category.value = route.query.category ? String(route.query.category) : ''
}

function onFilterChange() {
  const query = { ...route.query }
  if (cropKey.value) query.crop_key = cropKey.value
  else delete query.crop_key
  if (category.value) query.category = category.value
  else delete query.category
  if (route.path === '/operator-guide') {
    query.tab = 'library'
    query.section = 'symptoms'
    router.replace({ path: '/operator-guide', query })
  } else {
    router.replace({ path: route.path, query })
  }
}

async function loadCatalog() {
  const r = await api.get('/commons/agronomy-symptoms')
  catalog.value = Array.isArray(r.data?.symptoms) ? r.data.symptoms : []
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params = {}
    if (cropKey.value.trim()) params.crop_key = cropKey.value.trim()
    if (category.value.trim()) params.category = category.value.trim()
    const r = await api.get('/commons/agronomy-symptoms', { params })
    symptoms.value = Array.isArray(r.data?.symptoms) ? r.data.symptoms : []
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load symptoms'
    symptoms.value = []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  syncFromRoute()
  try {
    await loadCatalog()
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load symptom catalog'
  }
  await load()
})

watch(
  () => [route.query.crop_key, route.query.category],
  async () => {
    syncFromRoute()
    await load()
  },
)
</script>
