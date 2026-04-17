<template>
  <div class="p-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-6">
      <h1 class="text-xl font-semibold text-white">Commons Catalog</h1>
      <div class="flex items-center gap-2">
        <button
          @click="tab = 'browse'"
          class="text-xs font-medium px-3 py-1.5 rounded-lg border transition-colors"
          :class="tab === 'browse' ? 'bg-green-900/50 text-green-400 border-green-800' : 'bg-zinc-800 text-zinc-400 border-zinc-700 hover:text-zinc-200'"
        >
          Browse Catalog
        </button>
        <button
          @click="tab = 'imports'"
          class="text-xs font-medium px-3 py-1.5 rounded-lg border transition-colors"
          :class="tab === 'imports' ? 'bg-green-900/50 text-green-400 border-green-800' : 'bg-zinc-800 text-zinc-400 border-zinc-700 hover:text-zinc-200'"
        >
          Farm Imports
        </button>
      </div>
    </div>

    <!-- Browse tab -->
    <template v-if="tab === 'browse'">
      <div class="mb-4">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search catalog…"
          class="w-full max-w-md bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
          @input="debouncedSearch"
        />
      </div>

      <div v-if="loading" class="text-zinc-400 text-sm">Loading catalog…</div>
      <div v-else-if="!entries.length" class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center">
        No catalog entries found{{ searchQuery ? ' matching your search' : '' }}.
      </div>

      <div v-else class="flex flex-col lg:flex-row gap-4">
        <!-- List -->
        <div class="lg:w-1/2 space-y-2 max-h-[70vh] overflow-y-auto pr-1">
          <div
            v-for="e in entries"
            :key="e.id"
            @click="selectEntry(e)"
            class="bg-zinc-900 border rounded-xl p-4 cursor-pointer transition-colors"
            :class="selected?.slug === e.slug ? 'border-green-700 bg-zinc-800' : 'border-zinc-800 hover:border-zinc-700'"
          >
            <div class="flex items-start justify-between gap-2 mb-1">
              <p class="text-white text-sm font-medium">{{ e.title }}</p>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-700 text-zinc-300 shrink-0">
                {{ e.license_spdx }}
              </span>
            </div>
            <p class="text-zinc-500 text-xs line-clamp-2 mb-2">{{ e.summary }}</p>
            <div class="flex flex-wrap gap-1.5">
              <span
                v-for="tag in (e.tags || [])"
                :key="tag"
                class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700"
              >
                {{ tag }}
              </span>
            </div>
            <p v-if="e.contributor_display" class="text-zinc-600 text-[11px] mt-2">by {{ e.contributor_display }}</p>
          </div>
        </div>

        <!-- Detail pane -->
        <div class="lg:w-1/2">
          <div v-if="!selected" class="text-zinc-600 text-sm bg-zinc-900 border border-zinc-800 rounded-xl p-8 text-center">
            Select an entry to see details.
          </div>
          <div v-else-if="detailLoading" class="text-zinc-400 text-sm p-4">Loading…</div>
          <div v-else-if="detail" class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 sticky top-6">
            <h2 class="text-white text-lg font-semibold mb-1">{{ detail.title }}</h2>
            <p class="text-zinc-500 text-sm mb-3">{{ detail.summary }}</p>

            <div class="flex flex-wrap gap-1.5 mb-3">
              <span
                v-for="tag in (detail.tags || [])"
                :key="tag"
                class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700"
              >
                {{ tag }}
              </span>
            </div>

            <div class="flex items-center gap-3 text-xs text-zinc-500 mb-4">
              <span v-if="detail.contributor_display">{{ detail.contributor_display }}</span>
              <span v-if="detail.contributor_uri">
                <a :href="detail.contributor_uri" target="_blank" class="text-green-500 hover:text-green-400 underline">Link</a>
              </span>
              <span class="px-1.5 py-0.5 rounded bg-zinc-800 border border-zinc-700">{{ detail.license_spdx }}</span>
            </div>

            <!-- Body / readme -->
            <div v-if="readmeText" class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 mb-4 max-h-64 overflow-y-auto">
              <pre class="text-xs text-zinc-300 whitespace-pre-wrap font-mono">{{ readmeText }}</pre>
            </div>

            <!-- Import button -->
            <div class="flex items-center gap-3">
              <button
                @click="doImport(detail.slug)"
                :disabled="importing"
                class="text-xs font-medium px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
              >
                {{ importing ? 'Importing…' : 'Import to Farm' }}
              </button>
              <span v-if="importMsg" class="text-xs" :class="importErr ? 'text-red-400' : 'text-green-400'">
                {{ importMsg }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Imports tab -->
    <template v-if="tab === 'imports'">
      <div v-if="importsLoading" class="text-zinc-400 text-sm">Loading imports…</div>
      <div v-else-if="!imports.length" class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center">
        No catalog entries imported to this farm yet.
      </div>
      <div v-else class="space-y-2">
        <div
          v-for="imp in imports"
          :key="imp.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-white text-sm font-medium">{{ imp.title }}</p>
              <p class="text-zinc-500 text-xs mt-0.5">{{ imp.summary }}</p>
            </div>
            <div class="text-right shrink-0">
              <p class="text-zinc-600 text-[11px]">Imported {{ formatDate(imp.imported_at) }}</p>
              <p v-if="imp.note" class="text-zinc-500 text-[11px] mt-0.5 italic">{{ imp.note }}</p>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const tab = ref('browse')
const searchQuery = ref('')
const entries = ref([])
const loading = ref(false)
const selected = ref(null)
const detail = ref(null)
const detailLoading = ref(false)
const importing = ref(false)
const importMsg = ref('')
const importErr = ref(false)
const imports = ref([])
const importsLoading = ref(false)

let searchTimer = null
function debouncedSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => loadCatalog(), 300)
}

async function loadCatalog() {
  loading.value = true
  try {
    entries.value = await store.loadCatalog({ q: searchQuery.value })
  } finally {
    loading.value = false
  }
}

async function selectEntry(entry) {
  selected.value = entry
  detailLoading.value = true
  try {
    detail.value = await store.getCatalogEntry(entry.slug)
  } finally {
    detailLoading.value = false
  }
}

const readmeText = ref('')
watch(detail, (d) => {
  if (!d?.body) { readmeText.value = ''; return }
  const b = typeof d.body === 'string' ? JSON.parse(d.body) : d.body
  readmeText.value = b?.readme_md || b?.readme || JSON.stringify(b, null, 2)
})

async function doImport(slug) {
  const fid = farmContext.farmId
  if (!fid) { importMsg.value = 'No farm selected'; importErr.value = true; return }
  importing.value = true
  importMsg.value = ''
  importErr.value = false
  try {
    await store.importCatalogEntry(fid, slug)
    importMsg.value = 'Imported successfully'
    importErr.value = false
  } catch (e) {
    importMsg.value = e.response?.data?.error || e.message || 'Import failed'
    importErr.value = true
  } finally {
    importing.value = false
  }
}

async function loadImports() {
  const fid = farmContext.farmId
  if (!fid) return
  importsLoading.value = true
  try {
    imports.value = await store.loadCatalogImports(fid)
  } finally {
    importsLoading.value = false
  }
}

function formatDate(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

watch(tab, (t) => {
  if (t === 'imports') loadImports()
})

onMounted(loadCatalog)
watch(() => farmContext.farmId, () => {
  loadCatalog()
  if (tab.value === 'imports') loadImports()
})
</script>
