<template>
  <div class="p-6 max-w-4xl">
    <h1 class="text-2xl font-bold text-green-400 mb-2">Symptom guide</h1>
    <p class="text-sm text-zinc-500 mb-6">
      Browse the agronomy symptom catalog — filter by crop or category. Linked from Farm Guardian citations.
    </p>

    <div class="flex flex-wrap gap-3 mb-6">
      <div>
        <label class="block text-xs text-zinc-500 mb-1">Crop key</label>
        <input
          v-model="cropKey"
          type="text"
          placeholder="e.g. tomato"
          class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white w-40"
          @keyup.enter="load"
        />
      </div>
      <div>
        <label class="block text-xs text-zinc-500 mb-1">Category</label>
        <input
          v-model="category"
          type="text"
          placeholder="e.g. deficiency"
          class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white w-40"
          @keyup.enter="load"
        />
      </div>
      <div class="flex items-end">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800"
          @click="load"
        >
          Search
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-500 text-sm">Loading…</div>
    <div v-else-if="error" class="text-red-400 text-sm">{{ error }}</div>
    <div v-else-if="symptoms.length === 0" class="text-zinc-500 text-sm">No symptoms match your filters.</div>
    <div v-else class="space-y-3">
      <article
        v-for="s in symptoms"
        :key="s.id"
        class="bg-zinc-800 border border-zinc-700 rounded-xl p-4"
      >
        <div class="flex flex-wrap items-start justify-between gap-2 mb-2">
          <h2 class="text-white text-sm font-semibold">{{ s.symptom_name || s.symptom_key }}</h2>
          <span v-if="s.category" class="text-[10px] uppercase tracking-wide text-zinc-500">{{ s.category }}</span>
        </div>
        <p v-if="s.crop_key" class="text-[11px] text-zinc-500 mb-1">Crop: {{ s.crop_key }}</p>
        <p v-if="s.description" class="text-zinc-300 text-sm">{{ s.description }}</p>
        <p v-if="s.likely_causes" class="text-zinc-400 text-xs mt-2">
          <span class="text-zinc-500">Likely causes:</span> {{ s.likely_causes }}
        </p>
      </article>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'

const route = useRoute()
const cropKey = ref('')
const category = ref('')
const symptoms = ref([])
const loading = ref(false)
const error = ref('')

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

onMounted(() => {
  if (route.query.crop_key) cropKey.value = String(route.query.crop_key)
  if (route.query.category) category.value = String(route.query.category)
  load()
})

watch(
  () => [route.query.crop_key, route.query.category],
  () => {
    if (route.query.crop_key) cropKey.value = String(route.query.crop_key)
    if (route.query.category) category.value = String(route.query.category)
    load()
  },
)
</script>
