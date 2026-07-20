<template>
  <div class="space-y-6 max-w-4xl p-4" data-test="nf-recipe-library">
    <div>
      <h2 class="text-lg font-semibold text-white">Recipe library</h2>
      <p class="text-sm text-zinc-500 mt-1">
        Read-only canon from Phase 208 — inputs, application recipes, and bootstrap programs. Ratios come from
        recipe canon and field guides, not hardcoded UI copy.
      </p>
    </div>

    <div class="flex gap-1 bg-zinc-800 rounded-lg p-1 w-fit" data-test="nf-library-subtabs">
      <button
        v-for="t in LIBRARY_TABS"
        :key="t.id"
        type="button"
        class="px-3 py-1.5 text-sm rounded-md transition-colors font-medium"
        :class="libraryTab === t.id ? 'bg-green-600 text-white' : 'text-zinc-400 hover:text-white'"
        :data-test="`nf-library-tab-${t.id}`"
        @click="selectLibraryTab(t.id)"
      >
        {{ t.label }}
        <span class="text-[10px] opacity-80 ml-1">({{ tabCounts[t.id] }})</span>
      </button>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>
    <p v-else-if="loading" class="text-sm text-zinc-500">Loading canonical recipes…</p>

    <template v-else>
      <div class="grid grid-cols-1 lg:grid-cols-[minmax(0,14rem)_1fr] gap-4">
        <!-- Card list -->
        <div class="space-y-2 max-h-[min(60vh,28rem)] overflow-y-auto pr-1" data-test="nf-library-list">
          <template v-if="libraryTab === 'inputs'">
            <button
              v-for="item in canonInputs"
              :key="item.seed_name"
              type="button"
              class="w-full text-left rounded-lg border px-3 py-2 transition-colors"
              :class="selectedInput?.seed_name === item.seed_name
                ? 'border-green-600 bg-green-950/30'
                : 'border-zinc-800 bg-zinc-950/60 hover:border-zinc-600'"
              :data-test="`nf-library-input-${libraryCardSlug(item.seed_name)}`"
              @click="selectInput(item)"
            >
              <p class="text-sm text-zinc-100 leading-snug">{{ item.seed_name }}</p>
              <p v-if="traditionBadge(item.tradition)" class="text-[10px] uppercase mt-1" :class="traditionBadge(item.tradition).class">
                {{ traditionBadge(item.tradition).text }}
              </p>
            </button>
          </template>

          <template v-else-if="libraryTab === 'application'">
            <button
              v-for="item in applicationRecipes"
              :key="item.seed_name"
              type="button"
              class="w-full text-left rounded-lg border px-3 py-2 transition-colors"
              :class="selectedRecipe?.seed_name === item.seed_name
                ? 'border-green-600 bg-green-950/30'
                : 'border-zinc-800 bg-zinc-950/60 hover:border-zinc-600'"
              :data-test="`nf-library-recipe-${libraryCardSlug(item.seed_name)}`"
              @click="selectRecipe(item)"
            >
              <p class="text-sm text-zinc-100 leading-snug">{{ item.seed_name }}</p>
              <p class="text-[11px] text-green-400/80 font-mono mt-1 truncate">{{ item.dilution }}</p>
            </button>
          </template>

          <template v-else>
            <button
              v-for="prog in LIBRARY_PROGRAMS"
              :key="prog.id"
              type="button"
              class="w-full text-left rounded-lg border px-3 py-2 transition-colors"
              :class="selectedProgram?.id === prog.id
                ? 'border-green-600 bg-green-950/30'
                : 'border-zinc-800 bg-zinc-950/60 hover:border-zinc-600'"
              :data-test="`nf-library-program-${prog.id}`"
              @click="selectProgram(prog)"
            >
              <p class="text-sm text-zinc-100">{{ prog.title }}</p>
              <p class="text-[11px] text-zinc-500 mt-1">{{ prog.summary }}</p>
            </button>
          </template>
        </div>

        <!-- Detail -->
        <div v-if="detailLoading" class="text-sm text-zinc-500">Loading field guide…</div>
        <div v-else-if="detailError" class="text-sm text-red-400">{{ detailError }}</div>

        <article
          v-else-if="libraryTab === 'inputs' && selectedInput"
          class="rounded-xl border border-zinc-800 bg-zinc-900 p-4 space-y-4"
          data-test="nf-library-input-detail"
        >
          <header class="space-y-1">
            <h3 class="text-base font-semibold text-white">{{ selectedInput.seed_name }}</h3>
            <p v-if="canonDilutionHint(selectedInput)" class="text-xs text-green-400/90 font-mono">
              {{ canonDilutionHint(selectedInput) }}
            </p>
            <p v-if="selectedInput.reference_source" class="text-xs text-zinc-500">
              {{ selectedInput.reference_source }}
            </p>
          </header>
          <GuideStepCards :cards="stepCards" />
          <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
            <router-link
              :to="{ path: '/natural-farming', query: { tab: 'batch', process: selectedInput.process_type } }"
              class="text-xs px-3 py-1.5 rounded-lg bg-green-800/50 text-green-300 border border-green-700 hover:bg-green-800/70"
              data-test="nf-library-make-batch-link"
            >
              Make this batch →
            </router-link>
            <LearnHowExpander v-if="selectedInput.guide" :guide-file="selectedInput.guide" />
          </div>
        </article>

        <article
          v-else-if="libraryTab === 'application' && selectedRecipe"
          class="rounded-xl border border-zinc-800 bg-zinc-900 p-4 space-y-4"
          data-test="nf-library-recipe-detail"
        >
          <header class="space-y-1">
            <h3 class="text-base font-semibold text-white">{{ selectedRecipe.seed_name }}</h3>
            <p class="text-xs text-zinc-400">
              {{ formatApplicationType(selectedRecipe.target_application_type) }}
            </p>
            <p class="text-sm text-green-400/90 font-mono">{{ selectedRecipe.dilution }}</p>
          </header>
          <div>
            <h4 class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Components</h4>
            <ul class="text-sm text-zinc-300 space-y-1">
              <li v-for="(c, i) in selectedRecipe.components || []" :key="i">· {{ c }}</li>
            </ul>
          </div>
          <div v-if="linkedBatches.length">
            <h4 class="text-xs uppercase tracking-wide text-zinc-500 mb-2">Ready batches on farm</h4>
            <ul class="text-sm text-zinc-300 space-y-1">
              <li v-for="row in linkedBatches" :key="row.batch.id">
                {{ row.inputName }} — {{ row.batch.batch_identifier || `#${row.batch.id}` }}
                ({{ row.batch.status }})
              </li>
            </ul>
          </div>
          <p v-else-if="farmId" class="text-xs text-zinc-500">No ready batches for these inputs yet.</p>
          <GuideStepCards v-if="stepCards.length" :cards="stepCards.slice(0, 3)" />
          <LearnHowExpander v-if="selectedRecipe.guide" :guide-file="selectedRecipe.guide" />
        </article>

        <article
          v-else-if="libraryTab === 'programs' && selectedProgram"
          class="rounded-xl border border-zinc-800 bg-zinc-900 p-4 space-y-4"
          data-test="nf-library-program-detail"
        >
          <header class="space-y-1">
            <h3 class="text-base font-semibold text-white">{{ selectedProgram.title }}</h3>
            <p class="text-xs text-zinc-500 font-mono">{{ selectedProgram.bootstrapTemplate }}</p>
            <p class="text-sm text-zinc-400">{{ selectedProgram.summary }}</p>
          </header>
          <GuideStepCards :cards="stepCards" />
          <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
            <router-link
              :to="{ path: '/feed-water', query: { tab: 'programs' } }"
              class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-white"
              data-test="nf-library-feed-water-link"
            >
              Open Feed &amp; water → Programs
            </router-link>
            <LearnHowExpander :guide-file="selectedProgram.guide" />
          </div>
        </article>

        <p v-else class="text-sm text-zinc-500">Select an item to view instructions.</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useFarmContextStore } from '../../stores/farmContext.js'
import { useFarmStore } from '../../stores/farm.js'
import { loadRecipeCanon } from '../../lib/naturalFarmingCanon.js'
import { batchStepCards } from '../../lib/naturalFarmingGuideSections.js'
import {
  canonDilutionHint,
  loadFieldGuideBody,
} from '../../lib/naturalFarmingBatchFlow.js'
import {
  LIBRARY_PROGRAMS,
  LIBRARY_TABS,
  formatApplicationType,
  libraryCardSlug,
  readyBatchesForComponents,
  traditionBadge,
} from '../../lib/naturalFarmingLibrary.js'
import GuideStepCards from './GuideStepCards.vue'
import LearnHowExpander from './LearnHowExpander.vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)

const canon = ref(null)
const loading = ref(true)
const loadError = ref('')
const libraryTab = ref('inputs')
const selectedInput = ref(null)
const selectedRecipe = ref(null)
const selectedProgram = ref(null)
const guideBody = ref('')
const detailLoading = ref(false)
const detailError = ref('')
const farmInputs = ref([])
const farmBatches = ref([])

const canonInputs = computed(() => /** @type {Array<Record<string, unknown>>} */ (canon.value?.inputs ?? []))
const applicationRecipes = computed(
  () => /** @type {Array<Record<string, unknown>>} */ (canon.value?.application_recipes ?? []),
)
const stepCards = computed(() => batchStepCards(guideBody.value))
const tabCounts = computed(() => ({
  inputs: canonInputs.value.length,
  application: applicationRecipes.value.length,
  programs: LIBRARY_PROGRAMS.length,
}))
const linkedBatches = computed(() => {
  if (!selectedRecipe.value) return []
  return readyBatchesForComponents(
    selectedRecipe.value.components,
    farmInputs.value,
    farmBatches.value,
  )
})

function selectLibraryTab(id) {
  libraryTab.value = id
  selectedInput.value = null
  selectedRecipe.value = null
  selectedProgram.value = null
  guideBody.value = ''
  if (id === 'inputs' && canonInputs.value.length) selectInput(canonInputs.value[0])
  if (id === 'application' && applicationRecipes.value.length) selectRecipe(applicationRecipes.value[0])
  if (id === 'programs' && LIBRARY_PROGRAMS.length) selectProgram(LIBRARY_PROGRAMS[0])
}

async function loadGuideForFile(guideFile) {
  detailLoading.value = true
  detailError.value = ''
  guideBody.value = ''
  try {
    guideBody.value = await loadFieldGuideBody(guideFile)
  } catch (err) {
    detailError.value = err?.response?.data?.error || err?.message || 'Could not load field guide'
  } finally {
    detailLoading.value = false
  }
}

function selectInput(item) {
  selectedInput.value = item
  loadGuideForFile(item.guide)
}

function selectRecipe(item) {
  selectedRecipe.value = item
  loadGuideForFile(item.guide)
}

function selectProgram(prog) {
  selectedProgram.value = prog
  loadGuideForFile(prog.guide)
}

async function loadFarmInventory() {
  if (!farmId.value) {
    farmInputs.value = []
    farmBatches.value = []
    return
  }
  farmInputs.value = await store.loadNfInputs(farmId.value)
  farmBatches.value = await store.loadNfBatches(farmId.value)
}

onMounted(async () => {
  try {
    canon.value = await loadRecipeCanon()
    await loadFarmInventory()
    if (canonInputs.value.length) selectInput(canonInputs.value[0])
  } catch (err) {
    loadError.value = err?.message || 'Could not load recipe canon'
  } finally {
    loading.value = false
  }
})

watch(farmId, loadFarmInventory)
</script>
