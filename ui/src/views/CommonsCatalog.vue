<template>
  <div :class="embedded ? 'p-4 sm:p-6' : 'p-6'">
    <div v-if="!embedded" class="mb-4">
      <h1 class="text-xl font-semibold text-white mb-2">Commons Catalog</h1>
      <p class="text-zinc-400 text-sm max-w-3xl leading-relaxed">
        Share and install <strong class="text-zinc-300 font-medium">starter packs</strong> on this server —
        fertigation recipes, agronomy checklists, and documentation. Browse packs published here (by gr33n or your team),
        then <strong class="text-zinc-300 font-medium">Import to Farm</strong> to apply them automatically.
        Use <strong class="text-zinc-300 font-medium">Publish</strong> to export this farm&apos;s fertigation programs for other farms on the same deployment.
      </p>
    </div>

    <div class="flex flex-wrap items-center gap-2 mb-6">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="text-xs font-medium px-3 py-1.5 rounded-lg border transition-colors"
        :class="tab === t.id ? 'bg-green-900/50 text-green-400 border-green-800' : 'bg-zinc-800 text-zinc-400 border-zinc-700 hover:text-zinc-200'"
        @click="tab = t.id"
      >
        {{ t.label }}
      </button>
    </div>

    <!-- Browse -->
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
        <div class="lg:w-1/2 space-y-2 max-h-[70vh] overflow-y-auto pr-1">
          <div
            v-for="e in entries"
            :key="e.id"
            class="bg-zinc-900 border rounded-xl p-4 cursor-pointer transition-colors"
            :class="selected?.slug === e.slug ? 'border-green-700 bg-zinc-800' : 'border-zinc-800 hover:border-zinc-700'"
            @click="selectEntry(e)"
          >
            <div class="flex items-start justify-between gap-2 mb-1">
              <p class="text-white text-sm font-medium">{{ e.title }}</p>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-700 text-zinc-300 shrink-0">{{ e.license_spdx }}</span>
            </div>
            <p class="text-zinc-500 text-xs line-clamp-2 mb-2">{{ e.summary }}</p>
            <div class="flex flex-wrap gap-1.5">
              <span v-for="tag in (e.tags || [])" :key="tag" class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700">{{ tag }}</span>
            </div>
            <p v-if="e.contributor_display" class="text-zinc-600 text-[11px] mt-2">by {{ e.contributor_display }}</p>
          </div>
        </div>

        <div class="lg:w-1/2">
          <div v-if="!selected" class="text-zinc-600 text-sm bg-zinc-900 border border-zinc-800 rounded-xl p-8 text-center">
            Select a pack to preview and import.
          </div>
          <div v-else-if="detailLoading" class="text-zinc-400 text-sm p-4">Loading…</div>
          <div v-else-if="detail" class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 sticky top-6">
            <h2 class="text-white text-lg font-semibold mb-1">{{ detail.title }}</h2>
            <p class="text-zinc-500 text-sm mb-3">{{ detail.summary }}</p>
            <p v-if="packKind" class="text-[10px] uppercase tracking-wide text-green-500/90 mb-3">Kind: {{ packKind.replace(/_/g, ' ') }}</p>

            <div v-if="packCropKeys.length" class="flex flex-wrap items-center gap-1.5 mb-3">
              <span class="text-[10px] text-zinc-500 uppercase tracking-wide">Crops</span>
              <span v-for="key in packCropKeys" :key="key" class="text-[10px] px-1.5 py-0.5 rounded bg-green-950/60 text-green-400 border border-green-900/50">{{ key }}</span>
            </div>

            <div v-if="readmeText" class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 mb-4 max-h-64 overflow-y-auto">
              <pre class="text-xs text-zinc-300 whitespace-pre-wrap font-mono">{{ readmeText }}</pre>
            </div>

            <p class="text-zinc-500 text-xs mb-3">
              Import applies this pack to <strong class="text-zinc-400">{{ farmContext.selectedFarm?.name || 'the selected farm' }}</strong>
              — recipe packs create fertigation programs (inactive until you enable them).
            </p>

            <div class="flex flex-wrap items-center gap-3">
              <button
                type="button"
                :disabled="importing || !farmContext.farmId"
                class="text-xs font-medium px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
                @click="doImport(detail.slug)"
              >
                {{ importing ? 'Importing…' : 'Import to Farm' }}
              </button>
            </div>

            <div v-if="applyResult" class="mt-4 p-3 rounded-lg border text-xs" :class="applyErr ? 'border-red-800 bg-red-950/30 text-red-200' : 'border-green-900/50 bg-green-950/20 text-green-100'">
              <p class="font-semibold mb-1">{{ applyResult.message || applyResult.status }}</p>
              <p v-if="applyResult.programs_created">Created {{ applyResult.programs_created }} program(s).</p>
              <p v-if="applyResult.programs_updated">Updated metadata on {{ applyResult.programs_updated }} existing program(s).</p>
              <p v-if="applyResult.programs_skipped">Skipped {{ applyResult.programs_skipped }} existing program(s) (by name).</p>
              <ul v-if="applyResult.next_steps?.length" class="mt-2 list-disc pl-4 text-zinc-300 space-y-0.5">
                <li v-for="(s, i) in applyResult.next_steps" :key="i">{{ s }}</li>
              </ul>
              <p v-if="applyErr" class="mt-2 text-red-300">{{ applyErr }}</p>
            </div>
            <p v-else-if="importMsg && !applyResult" class="text-xs mt-2" :class="importErr ? 'text-red-400' : 'text-green-400'">{{ importMsg }}</p>
          </div>
        </div>
      </div>
    </template>

    <!-- Imports history -->
    <template v-if="tab === 'imports'">
      <div v-if="importsLoading" class="text-zinc-400 text-sm">Loading imports…</div>
      <div v-else-if="!imports.length" class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center">
        No packs imported to this farm yet. Browse the catalog and click Import to Farm.
      </div>
      <div v-else class="space-y-2">
        <div v-for="imp in imports" :key="imp.id" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <p class="text-white text-sm font-medium">{{ imp.title }}</p>
          <p class="text-zinc-500 text-xs mt-0.5">{{ imp.summary }}</p>
          <p class="text-zinc-600 text-[11px] mt-2">Imported {{ formatDate(imp.imported_at) }}</p>
        </div>
      </div>
    </template>

    <!-- Publish -->
    <template v-if="tab === 'publish'">
      <div class="max-w-xl space-y-4">
        <p class="text-zinc-400 text-sm">
          Export all fertigation programs from this farm as a recipe pack other farms on <em>this server</em> can import.
          Programs are published as <strong class="text-zinc-300">inactive</strong> for safety.
        </p>

        <label class="block">
          <span class="text-zinc-400 text-xs uppercase tracking-wide">Pack slug</span>
          <input v-model="publishSlug" type="text" placeholder="my-farm-lettuce-recipes" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs uppercase tracking-wide">Title</span>
          <input v-model="publishTitle" type="text" placeholder="My farm — lettuce recipes" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs uppercase tracking-wide">Summary</span>
          <textarea v-model="publishSummary" rows="2" class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
        </label>

        <button
          type="button"
          :disabled="publishBusy || !farmContext.farmId || !publishSlug.trim() || !publishTitle.trim()"
          class="text-xs font-medium px-4 py-2 rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
          @click="doExportPublish"
        >
          {{ publishBusy ? 'Publishing…' : 'Export farm programs & publish' }}
        </button>
        <p v-if="publishMsg" class="text-xs" :class="publishErr ? 'text-red-400' : 'text-green-400'">{{ publishMsg }}</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

defineProps({
  embedded: { type: Boolean, default: false },
})

const store = useFarmStore()
const farmContext = useFarmContextStore()

const tabs = [
  { id: 'browse', label: 'Browse Catalog' },
  { id: 'imports', label: 'Farm Imports' },
  { id: 'publish', label: 'Publish from Farm' },
]

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
const applyResult = ref(null)
const applyErr = ref('')
const imports = ref([])
const importsLoading = ref(false)

const publishSlug = ref('')
const publishTitle = ref('')
const publishSummary = ref('')
const publishBusy = ref(false)
const publishMsg = ref('')
const publishErr = ref(false)

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
  applyResult.value = null
  applyErr.value = ''
  importMsg.value = ''
  detailLoading.value = true
  try {
    detail.value = await store.getCatalogEntry(entry.slug)
  } finally {
    detailLoading.value = false
  }
}

const readmeText = ref('')
const packKind = ref('')
const packCropKeys = computed(() => {
  const d = detail.value
  if (!d?.body) return []
  const b = typeof d.body === 'string' ? JSON.parse(d.body) : d.body
  if (b?.kind !== 'fertigation_recipe_pack') return []
  const keys = new Set()
  for (const p of b.programs || []) {
    for (const k of p.recommended_crop_keys || []) keys.add(k)
  }
  return [...keys].sort()
})

watch(detail, (d) => {
  if (!d?.body) {
    readmeText.value = ''
    packKind.value = ''
    return
  }
  const b = typeof d.body === 'string' ? JSON.parse(d.body) : d.body
  packKind.value = b?.kind || ''
  readmeText.value = b?.readme_md || b?.readme || ''
})

async function doImport(slug) {
  const fid = farmContext.farmId
  if (!fid) {
    importMsg.value = 'No farm selected'
    importErr.value = true
    return
  }
  importing.value = true
  importMsg.value = ''
  importErr.value = false
  applyResult.value = null
  applyErr.value = ''
  try {
    const out = await store.importCatalogEntry(fid, slug)
    applyResult.value = out.apply || null
    if (out.error) {
      applyErr.value = out.error
      importErr.value = true
    } else {
      importMsg.value = out.apply?.message || 'Imported successfully'
      importErr.value = false
    }
  } catch (e) {
    importMsg.value = e.response?.data?.error || e.message || 'Import failed'
    importErr.value = true
  } finally {
    importing.value = false
  }
}

async function doExportPublish() {
  const fid = farmContext.farmId
  if (!fid) return
  publishBusy.value = true
  publishMsg.value = ''
  publishErr.value = false
  try {
    const out = await store.exportFarmRecipePack(fid, {
      slug: publishSlug.value.trim(),
      title: publishTitle.value.trim(),
      summary: publishSummary.value.trim(),
      publish: true,
    })
    publishMsg.value = out.message || `Published ${out.programs_exported} programs.`
    publishErr.value = false
    await loadCatalog()
    tab.value = 'browse'
  } catch (e) {
    publishMsg.value = e.response?.data?.error || e.message || 'Publish failed'
    publishErr.value = true
  } finally {
    publishBusy.value = false
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
