<template>
  <section class="rounded-xl border border-zinc-800 bg-zinc-950/40 p-4 space-y-4" data-test="nf-commons-import">
    <div>
      <h3 class="text-sm font-medium text-zinc-200">Import a recipe pack</h3>
      <p class="text-xs text-zinc-500 mt-1 max-w-2xl">
        Browse Commons packs tagged <span class="text-zinc-400">natural_farming</span> — audited inputs,
        application recipes, and components import idempotently onto this farm.
      </p>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>
    <p v-else-if="loading" class="text-sm text-zinc-500">Loading Commons catalog…</p>

    <template v-else>
      <div v-if="!entries.length" class="text-sm text-zinc-500">
        No natural farming packs published yet. Run migrations or publish from Settings → Commons.
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-[minmax(0,14rem)_1fr] gap-4">
        <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
          <button
            v-for="entry in entries"
            :key="entry.slug"
            type="button"
            class="w-full text-left rounded-lg border px-3 py-2 transition-colors"
            :class="selectedSlug === entry.slug
              ? 'border-green-600 bg-green-950/30'
              : 'border-zinc-800 bg-zinc-900/80 hover:border-zinc-600'"
            :data-test="`nf-commons-pack-${entry.slug}`"
            @click="selectEntry(entry)"
          >
            <p class="text-sm text-zinc-100 leading-snug">{{ entry.title }}</p>
            <p class="text-[11px] text-zinc-500 mt-1 line-clamp-2">{{ entry.summary }}</p>
          </button>
        </div>

        <div v-if="!selectedSlug" class="text-sm text-zinc-500 flex items-center">
          Select a pack to preview inputs and recipes.
        </div>
        <div v-else-if="detailLoading" class="text-sm text-zinc-500">Loading pack…</div>
        <div v-else-if="detailError" class="text-sm text-red-400">{{ detailError }}</div>
        <article v-else-if="preview" class="space-y-3">
          <header class="space-y-1">
            <h4 class="text-sm font-semibold text-white">{{ detail?.title }}</h4>
            <p class="text-xs text-zinc-500">{{ detail?.summary }}</p>
          </header>

          <dl class="grid grid-cols-3 gap-2 text-center text-xs">
            <div class="rounded-lg border border-zinc-800 bg-zinc-900/80 p-2">
              <dt class="text-zinc-500">Inputs</dt>
              <dd class="text-lg font-semibold text-white">{{ preview.inputCount }}</dd>
            </div>
            <div class="rounded-lg border border-zinc-800 bg-zinc-900/80 p-2">
              <dt class="text-zinc-500">Recipes</dt>
              <dd class="text-lg font-semibold text-white">{{ preview.recipeCount }}</dd>
            </div>
            <div class="rounded-lg border border-zinc-800 bg-zinc-900/80 p-2">
              <dt class="text-zinc-500">Components</dt>
              <dd class="text-lg font-semibold text-white">{{ preview.componentCount }}</dd>
            </div>
          </dl>

          <div v-if="preview.inputNames.length" class="space-y-1">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500">Input definitions</p>
            <ul class="text-xs text-zinc-300 space-y-0.5 max-h-24 overflow-y-auto">
              <li v-for="name in preview.inputNames" :key="name">· {{ name }}</li>
            </ul>
          </div>
          <div v-if="preview.recipeNames.length" class="space-y-1">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500">Application recipes</p>
            <ul class="text-xs text-zinc-300 space-y-0.5 max-h-24 overflow-y-auto">
              <li v-for="name in preview.recipeNames" :key="name">· {{ name }}</li>
            </ul>
          </div>

          <div class="flex flex-wrap gap-2 pt-1 border-t border-zinc-800">
            <button
              type="button"
              class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
              :disabled="importing || !farmId"
              data-test="nf-commons-import-btn"
              @click="doImport"
            >
              {{ importing ? 'Importing…' : 'Import to farm' }}
            </button>
            <router-link
              v-if="importOk"
              :to="{ path: '/natural-farming', query: firstBatchQuery }"
              class="px-4 py-2 text-sm font-medium rounded-lg border border-zinc-600 text-zinc-200 hover:border-zinc-400"
              data-test="nf-commons-make-first-batch"
            >
              Make first batch
            </router-link>
          </div>
          <p v-if="importMessage" class="text-sm" :class="importOk ? 'text-green-400' : 'text-red-400'">
            {{ importMessage }}
          </p>
        </article>
        <p v-else class="text-sm text-amber-400/90">
          This catalog entry is not a natural farming recipe pack.
        </p>
      </div>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFarmStore } from '../../stores/farm.js'
import { useFarmContextStore } from '../../stores/farmContext.js'
import {
  NF_COMMONS_TAG,
  firstBatchQueryForPack,
  isNaturalFarmingCatalogEntry,
  parseNaturalFarmingPackBody,
} from '../../lib/naturalFarmingCommonsImport.js'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)

const entries = ref([])
const loading = ref(true)
const loadError = ref('')
const selectedSlug = ref('')
const detail = ref(null)
const detailLoading = ref(false)
const detailError = ref('')
const importing = ref(false)
const importMessage = ref('')
const importOk = ref(false)

const preview = computed(() => parseNaturalFarmingPackBody(detail.value?.body))
const firstBatchQuery = computed(() => firstBatchQueryForPack(preview.value))

async function loadEntries() {
  loading.value = true
  loadError.value = ''
  try {
    const rows = await store.loadCatalog({ q: NF_COMMONS_TAG })
    entries.value = rows.filter(isNaturalFarmingCatalogEntry)
  } catch (err) {
    loadError.value = err?.response?.data?.error || err?.message || 'Could not load catalog'
  } finally {
    loading.value = false
  }
}

async function selectEntry(entry) {
  selectedSlug.value = entry.slug
  detail.value = null
  detailError.value = ''
  importMessage.value = ''
  importOk.value = false
  detailLoading.value = true
  try {
    detail.value = await store.getCatalogEntry(entry.slug)
  } catch (err) {
    detailError.value = err?.response?.data?.error || err?.message || 'Could not load pack'
  } finally {
    detailLoading.value = false
  }
}

async function doImport() {
  if (!farmId.value || !selectedSlug.value) {
    importOk.value = false
    importMessage.value = 'Select a farm first (top bar).'
    return
  }
  importing.value = true
  importMessage.value = ''
  try {
    const out = await store.importCatalogEntry(farmId.value, selectedSlug.value)
    if (out.error) {
      importOk.value = false
      importMessage.value = out.error
    } else {
      importOk.value = true
      importMessage.value = out.apply?.message || 'Recipe pack imported.'
      await store.loadAll(farmId.value)
    }
  } catch (err) {
    importOk.value = false
    importMessage.value = err?.response?.data?.error || err?.message || 'Import failed'
  } finally {
    importing.value = false
  }
}

onMounted(loadEntries)
watch(farmId, () => {
  importMessage.value = ''
  importOk.value = false
})
</script>
