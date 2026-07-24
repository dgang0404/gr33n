<template>
  <div class="space-y-6 max-w-4xl" data-test="nf-recipes-apply">
    <div>
      <h2 class="text-lg font-semibold text-white flex items-center gap-1">
        {{ NF_VOCAB.applyRecipes }}
        <ConceptHelpTip concept-id="application_recipe" position="bottom" />
      </h2>
      <p class="text-sm text-zinc-500 mt-1">
        Your farm {{ NF_VOCAB.applyRecipes.toLowerCase() }} — create, edit, and link to Feed &amp; water programs per zone.
      </p>
    </div>

    <p v-if="loading" class="text-sm text-zinc-500">Loading recipes…</p>
    <p v-else-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>

    <template v-else>
      <div class="flex items-center justify-between gap-3">
        <p class="text-zinc-400 text-sm">{{ recipes.length }} recipe(s)</p>
        <button
          type="button"
          class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg"
          data-test="nf-recipe-new"
          @click="toggleRecipeForm"
        >
          {{ showRecipeForm ? 'Cancel' : '+ New recipe' }}
        </button>
      </div>

      <form
        v-if="showRecipeForm"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 gap-3"
        data-test="nf-recipe-form"
        @submit.prevent="submitRecipe"
      >
        <input v-model="recipeForm.name" placeholder="Recipe name" required class="input-field" />
        <select v-model="recipeForm.target_application_type" required class="input-field">
          <option v-for="t in applicationTargets" :key="t" :value="t">{{ formatApplicationType(t) }}</option>
        </select>
        <select v-model.number="recipeForm.input_definition_id" class="input-field">
          <option :value="null">Primary input (optional)</option>
          <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
        </select>
        <input v-model="recipeForm.dilution_ratio" placeholder="Dilution ratio" class="input-field" />
        <textarea v-model="recipeForm.description" placeholder="Description" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.instructions" placeholder="Instructions" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.frequency_guidelines" placeholder="Frequency" class="input-field sm:col-span-2" rows="2" />
        <textarea v-model="recipeForm.notes" placeholder="Notes" class="input-field sm:col-span-2" rows="2" />
        <button
          type="submit"
          :disabled="saving"
          class="px-4 py-2 bg-green-700 text-white text-sm rounded-lg disabled:opacity-50 sm:col-span-2"
        >
          {{ saving ? 'Saving…' : (editRecipe ? 'Update recipe' : 'Create recipe') }}
        </button>
      </form>

      <section
        v-if="applyRecipe"
        ref="applyPanelEl"
        class="bg-zinc-900 border border-green-800/60 rounded-xl p-4 space-y-4 scroll-mt-24"
        data-test="nf-recipe-apply-panel"
      >
        <div class="flex items-start justify-between gap-2">
          <div>
            <h3 class="text-white font-medium">Apply — {{ applyRecipe.name }}</h3>
            <p class="text-xs text-zinc-500 mt-1">
              {{ formatApplicationType(applyRecipe.target_application_type) }}
              · {{ applyRecipe.dilution_ratio || 'no dilution set' }}
              · stages: {{ formatGrowthStages(applyRecipe.target_growth_stages, domainEnums) }}
            </p>
          </div>
          <button type="button" class="text-xs text-zinc-500 hover:text-zinc-300" @click="closeApply">
            Close
          </button>
        </div>

        <p class="text-xs text-zinc-500">
          Links this recipe to zones and Feed &amp; water programs — it does not run a mix or start a pump.
          Use <strong class="text-zinc-400">Open Feed &amp; water</strong> to wire schedules; Pi mix plans dose concentrates into your reservoir.
        </p>

        <label class="block text-xs text-zinc-500 max-w-xs">
          Zone
          <select v-model.number="applyZoneId" class="input-field mt-1 w-full" data-test="nf-apply-zone">
            <option :value="null">All zones</option>
            <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
          </select>
        </label>

        <div class="space-y-2">
          <p class="text-xs text-zinc-500 uppercase tracking-wide">Programs using this recipe</p>
          <ul v-if="linkedPrograms.length" class="space-y-2 text-sm">
            <li
              v-for="p in linkedPrograms"
              :key="p.id"
              class="flex flex-wrap items-center justify-between gap-2 border-b border-zinc-800 pb-2"
            >
              <span class="text-zinc-200">{{ p.name }}</span>
              <span class="text-xs text-zinc-500">{{ zoneLabel(p.target_zone_id) }}</span>
            </li>
          </ul>
          <p v-else class="text-sm text-zinc-600">
            No programs linked yet{{ applyZoneId ? ' in this zone' : '' }}.
            <router-link
              v-nav-hint="'/fertigation'"
              :to="feedWaterProgramLink(applyRecipe.id, { zoneId: applyZoneId })"
              class="text-green-400 hover:text-green-300 ml-1"
            >
              Create or link in Feed &amp; water →
            </router-link>
          </p>
        </div>

        <div class="flex flex-wrap gap-2 pt-2 border-t border-zinc-800">
          <router-link
            v-nav-hint="'/fertigation'"
            :to="feedWaterProgramLink(applyRecipe.id, { zoneId: applyZoneId })"
            class="text-xs px-3 py-1.5 rounded-lg bg-green-800/60 border border-green-700 text-green-200 hover:bg-green-800"
            data-test="nf-apply-feed-water-link"
          >
            Open Feed &amp; water → Programs
          </router-link>
          <router-link
            v-if="showAnimalsLink"
            v-nav-hint="'/animals'"
            to="/animals"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-white"
            data-test="nf-apply-animals-link"
          >
            Open Animals workspace
          </router-link>
        </div>
      </section>

      <div
        v-if="historyRecipe"
        ref="historyPanelEl"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 scroll-mt-24"
        data-test="nf-recipe-history"
      >
        <div class="flex items-center justify-between">
          <h3 class="text-white font-medium">Formula history — {{ historyRecipe.name }}</h3>
          <button type="button" class="text-xs text-zinc-500" @click="historyRecipe = null">Close</button>
        </div>
        <p class="text-xs text-zinc-500">
          Immutable snapshots from edits and component changes. Restore copies a revision onto the live recipe as a new revision — history is never deleted.
        </p>
        <p v-if="historyLoading" class="text-sm text-zinc-500">Loading history…</p>
        <ul v-else-if="recipeRevisions.length" class="space-y-2 text-sm">
          <li
            v-for="rev in recipeRevisions"
            :key="rev.id"
            class="flex flex-wrap items-center justify-between gap-2 border-b border-zinc-800 pb-2"
            :data-test="`nf-recipe-revision-${rev.id}`"
          >
            <div>
              <span class="text-zinc-200 font-medium">Rev {{ rev.revision_number }}</span>
              <span class="text-zinc-500 text-xs ml-2">{{ formatRevisionWhen(rev.created_at) }}</span>
              <p class="text-xs text-zinc-500 mt-0.5">
                {{ revisionSummaryLine(rev) }}
              </p>
              <p v-if="rev.change_summary" class="text-[10px] text-zinc-600">{{ rev.change_summary }}</p>
            </div>
            <button
              type="button"
              class="text-xs px-2 py-1 rounded border border-zinc-700 text-zinc-300 hover:text-white hover:border-zinc-500 disabled:opacity-40"
              :disabled="saving || rev.revision_number === latestRevisionNumber"
              :data-test="`nf-recipe-restore-${rev.id}`"
              @click="restoreRevision(rev)"
            >
              {{ rev.revision_number === latestRevisionNumber ? 'Current' : 'Restore' }}
            </button>
          </li>
        </ul>
        <p v-else class="text-sm text-zinc-600">No revisions yet.</p>
      </div>

      <div
        v-if="componentRecipe"
        ref="componentsPanelEl"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 scroll-mt-24"
        data-test="nf-recipe-components"
      >
        <div class="flex items-center justify-between">
          <h3 class="text-white font-medium">Components — {{ componentRecipe.name }}</h3>
          <button type="button" class="text-xs text-zinc-500" @click="componentRecipe = null">Close</button>
        </div>
        <form class="flex flex-wrap gap-2 items-end" @submit.prevent="addComponent">
          <select v-model.number="compForm.input_definition_id" required class="input-field">
            <option value="" disabled>Input</option>
            <option v-for="i in inputs" :key="i.id" :value="i.id">{{ i.name }}</option>
          </select>
          <input
            v-model.number="compForm.part_value"
            type="number"
            step="0.001"
            placeholder="Parts"
            required
            class="input-field w-28"
          />
          <button type="submit" :disabled="saving" class="px-3 py-2 bg-green-700 text-white text-xs rounded-lg">
            Add
          </button>
        </form>
        <ul class="space-y-2 text-sm">
          <li
            v-for="c in recipeComponents"
            :key="c.input_definition_id"
            class="flex justify-between gap-2 text-zinc-300 border-b border-zinc-800 pb-2"
          >
            <span>{{ c.input_name }} — {{ c.part_value }} parts</span>
            <button type="button" class="text-xs text-red-500" @click="removeComponentRow(c)">Remove</button>
          </li>
          <li v-if="!recipeComponents.length" class="text-zinc-600">No extra components.</li>
        </ul>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div
          v-for="rec in recipes"
          :key="rec.id"
          class="bg-zinc-800 border rounded-xl p-4 space-y-2 transition-colors"
          :class="applyRecipe?.id === rec.id ? 'border-green-600 ring-1 ring-green-900/40' : 'border-zinc-700'"
          :data-test="`nf-recipe-card-${rec.id}`"
        >
          <div class="flex items-start justify-between gap-2">
            <h3 class="text-white font-semibold">{{ rec.name }}</h3>
            <span class="text-xs px-2 py-0.5 rounded-full bg-zinc-700 text-zinc-300">
              {{ formatApplicationType(rec.target_application_type) }}
            </span>
          </div>
          <dl class="text-xs text-zinc-500 space-y-1">
            <div>
              <span class="text-zinc-600">Primary:</span>
              {{ rec.input_definition_id ? inputName(rec.input_definition_id) : '—' }}
            </div>
            <div>
              <span class="text-zinc-600">Dilution:</span>
              {{ rec.dilution_ratio || '—' }}
            </div>
            <div>
              <span class="text-zinc-600">Growth stages:</span>
              {{ formatGrowthStages(rec.target_growth_stages, domainEnums) }}
            </div>
          </dl>
          <p v-if="rec.description" class="text-zinc-400 text-sm line-clamp-2">{{ rec.description }}</p>
          <div class="flex flex-wrap gap-2 pt-2">
            <button
              type="button"
              class="text-xs font-medium"
              :class="applyRecipe?.id === rec.id ? 'text-green-300' : 'text-green-500 hover:text-green-400'"
              :data-test="`nf-recipe-apply-${rec.id}`"
              @click="openApply(rec)"
            >
              {{ applyRecipe?.id === rec.id ? 'Applying ▲' : 'Apply' }}
            </button>
            <button type="button" class="text-xs text-zinc-400 hover:text-white" @click="openRecipeComponents(rec)">
              Components
            </button>
            <button
              type="button"
              class="text-xs text-zinc-400 hover:text-white"
              :data-test="`nf-recipe-history-${rec.id}`"
              @click="openRecipeHistory(rec)"
            >
              History
            </button>
            <router-link
              v-nav-hint="'/fertigation'"
              :to="feedWaterProgramLink(rec.id)"
              class="text-xs text-green-500 hover:text-green-400"
              :data-test="`nf-recipe-feed-water-${rec.id}`"
            >
              Use in program
            </router-link>
            <button
              v-if="canWriteRecipes"
              type="button"
              class="text-xs text-zinc-400 hover:text-white"
              @click="startEditRecipe(rec)"
            >
              Edit
            </button>
            <button
              v-if="canDeleteRecipes"
              type="button"
              class="text-xs text-red-500 hover:text-red-400"
              @click="deleteRecipe(rec)"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
      <p v-if="!recipes.length" class="text-zinc-500 text-sm">No recipes yet — create one or import bootstrap.</p>

      <section class="pt-6 border-t border-zinc-800/80">
        <CommonsRecipePackImport />
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFarmStore } from '../../stores/farm.js'
import { useFarmContextStore } from '../../stores/farmContext.js'
import { useFarmCaps } from '../../composables/useFarmCaps.js'
import { FARM_SCOPES } from '../../lib/farmScopes.js'
import api from '../../api'
import { enumValues, loadDomainEnums } from '../../lib/domainEnums.js'
import { formatApplicationType } from '../../lib/naturalFarmingLibrary.js'
import {
  emptyRecipeForm,
  feedWaterProgramLink,
  formatGrowthStages,
  inputsByIdMap,
  isLivestockRecipe,
  programsForZone,
  programsUsingRecipe,
} from '../../lib/naturalFarmingRecipes.js'
import { isModuleEnabled, MODULE_SCHEMA, moduleMapFromRows } from '../../lib/farmModules.js'
import CommonsRecipePackImport from './CommonsRecipePackImport.vue'
import ConceptHelpTip from '../ConceptHelpTip.vue'
import { NF_VOCAB } from '../../lib/naturalFarmingVocabulary.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)
const { has: hasScope } = useFarmCaps(farmId)
const canWriteRecipes = computed(() => hasScope(FARM_SCOPES.nfRecipesWrite))
const canDeleteRecipes = computed(() => hasScope(FARM_SCOPES.nfRecipesDelete))

const loading = ref(true)
const loadError = ref('')
const saving = ref(false)
const inputs = ref([])
const recipes = ref([])
const programs = ref([])
const cropCycles = ref([])
const domainEnums = ref(null)

const showRecipeForm = ref(false)
const editRecipe = ref(null)
const recipeForm = ref(emptyRecipeForm())
const componentRecipe = ref(null)
const recipeComponents = ref([])
const compForm = ref({ input_definition_id: '', part_value: 1 })

const historyRecipe = ref(null)
const recipeRevisions = ref([])
const historyLoading = ref(false)
const historyPanelEl = ref(null)

const applyRecipe = ref(null)
const applyZoneId = ref(null)
const applyPanelEl = ref(null)
const componentsPanelEl = ref(null)

const zones = computed(() => store.zones)
const applicationTargets = computed(() => enumValues(domainEnums.value, 'application_targets'))
const inputMap = computed(() => inputsByIdMap(inputs.value))
const farmModules = computed(() => moduleMapFromRows(store.farmModules))

const linkedPrograms = computed(() => {
  if (!applyRecipe.value) return []
  const inZone = programsForZone(applyZoneId.value, programs.value, cropCycles.value)
  return programsUsingRecipe(applyRecipe.value.id, inZone)
})

const showAnimalsLink = computed(() => {
  if (!applyRecipe.value) return false
  if (!isModuleEnabled(farmModules.value, MODULE_SCHEMA.animals)) return false
  return isLivestockRecipe(applyRecipe.value, inputMap.value)
})

const latestRevisionNumber = computed(() => {
  if (!recipeRevisions.value.length) return 0
  return Math.max(...recipeRevisions.value.map((r) => Number(r.revision_number) || 0))
})

function formatRevisionWhen(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return String(iso)
  }
}

function revisionSummaryLine(rev) {
  try {
    const snap = typeof rev.snapshot === 'string' ? JSON.parse(rev.snapshot) : rev.snapshot
    const dilution = snap?.recipe?.dilution_ratio || '—'
    const count = Array.isArray(snap?.components) ? snap.components.length : 0
    return `${dilution} · ${count} component${count === 1 ? '' : 's'}`
  } catch {
    return '—'
  }
}

function zoneLabel(zoneId) {
  if (!zoneId) return 'All zones'
  const z = zones.value.find((row) => Number(row.id) === Number(zoneId))
  return z?.name || `Zone #${zoneId}`
}

function inputName(id) {
  const found = inputs.value.find((i) => i.id === id)
  return found ? found.name : `#${id}`
}

function scrollPanelIntoView(el) {
  nextTick(() => {
    el?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
  })
}

function closeApply() {
  applyRecipe.value = null
  applyZoneId.value = null
}

function openApply(rec) {
  if (applyRecipe.value?.id === rec.id) {
    closeApply()
    return
  }
  componentRecipe.value = null
  applyRecipe.value = rec
  applyZoneId.value = null
  scrollPanelIntoView(applyPanelEl.value)
}

function applyRecipeFromRoute() {
  const raw = route.query.recipe
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isFinite(id)) return
  const hit = recipes.value.find((r) => Number(r.id) === id)
  if (hit) {
    applyRecipe.value = hit
    showRecipeForm.value = false
    scrollPanelIntoView(applyPanelEl.value)
  }
}

async function loadAll() {
  if (!farmId.value) return
  try {
    const [i, r, p, cycles, enums] = await Promise.all([
      store.loadNfInputs(farmId.value),
      store.loadRecipes(farmId.value),
      store.loadFertigationPrograms(farmId.value),
      store.loadCropCycles(farmId.value),
      loadDomainEnums(api),
    ])
    inputs.value = i
    recipes.value = r
    programs.value = p
    cropCycles.value = cycles
    domainEnums.value = enums
    applyRecipeFromRoute()
  } catch (e) {
    loadError.value = e?.message || 'Failed to load recipes'
  }
}

onMounted(async () => {
  try {
    if (farmId.value) {
      await Promise.all([store.loadFarmModules(farmId.value), loadAll()])
    } else {
      await loadAll()
    }
  } finally {
    loading.value = false
  }
})

watch(() => route.query.recipe, () => applyRecipeFromRoute())
watch(farmId, () => { loading.value = true; loadAll().finally(() => { loading.value = false }) })

function toggleRecipeForm() {
  showRecipeForm.value = !showRecipeForm.value
  if (!showRecipeForm.value) {
    editRecipe.value = null
    recipeForm.value = emptyRecipeForm()
  }
}

async function submitRecipe() {
  saving.value = true
  try {
    const payload = { ...recipeForm.value }
    if (payload.input_definition_id === '' || payload.input_definition_id === null) {
      delete payload.input_definition_id
    }
    if (editRecipe.value) {
      await store.updateRecipe(editRecipe.value.id, payload)
    } else {
      await store.createRecipe(farmId.value, payload)
    }
    showRecipeForm.value = false
    editRecipe.value = null
    recipeForm.value = emptyRecipeForm()
    recipes.value = await store.loadRecipes(farmId.value)
  } finally {
    saving.value = false
  }
}

function startEditRecipe(rec) {
  editRecipe.value = rec
  showRecipeForm.value = true
  recipeForm.value = {
    name: rec.name,
    target_application_type: rec.target_application_type,
    input_definition_id: rec.input_definition_id ?? null,
    description: rec.description || '',
    dilution_ratio: rec.dilution_ratio || '',
    instructions: rec.instructions || '',
    frequency_guidelines: rec.frequency_guidelines || '',
    notes: rec.notes || '',
  }
}

async function deleteRecipe(rec) {
  if (!confirm(`Delete recipe "${rec.name}"?`)) return
  await store.deleteRecipe(rec.id)
  if (componentRecipe.value?.id === rec.id) componentRecipe.value = null
  if (applyRecipe.value?.id === rec.id) applyRecipe.value = null
  recipes.value = await store.loadRecipes(farmId.value)
}

async function openRecipeComponents(rec) {
  closeApply()
  historyRecipe.value = null
  componentRecipe.value = rec
  recipeComponents.value = await store.loadRecipeComponents(rec.id)
  scrollPanelIntoView(componentsPanelEl.value)
}

async function openRecipeHistory(rec) {
  closeApply()
  componentRecipe.value = null
  historyRecipe.value = rec
  historyLoading.value = true
  recipeRevisions.value = []
  scrollPanelIntoView(historyPanelEl.value)
  try {
    recipeRevisions.value = await store.loadRecipeRevisions(rec.id)
  } finally {
    historyLoading.value = false
  }
}

async function restoreRevision(rev) {
  if (!historyRecipe.value || !confirm(`Restore revision ${rev.revision_number}? Live recipe will match that snapshot as a new revision.`)) {
    return
  }
  saving.value = true
  try {
    const result = await store.restoreRecipeRevision(historyRecipe.value.id, rev.id)
    if (result?.recipe) {
      const idx = recipes.value.findIndex((r) => r.id === historyRecipe.value.id)
      if (idx >= 0) recipes.value[idx] = result.recipe
      historyRecipe.value = result.recipe
    } else {
      recipes.value = await store.loadRecipes(farmId.value)
      historyRecipe.value = recipes.value.find((r) => r.id === historyRecipe.value.id) || historyRecipe.value
    }
    recipeRevisions.value = await store.loadRecipeRevisions(historyRecipe.value.id)
  } finally {
    saving.value = false
  }
}

async function addComponent() {
  if (!componentRecipe.value || !compForm.value.input_definition_id) return
  saving.value = true
  try {
    await store.addRecipeComponent(componentRecipe.value.id, {
      input_definition_id: compForm.value.input_definition_id,
      part_value: compForm.value.part_value,
    })
    compForm.value = { input_definition_id: '', part_value: 1 }
    recipeComponents.value = await store.loadRecipeComponents(componentRecipe.value.id)
  } finally {
    saving.value = false
  }
}

async function removeComponentRow(c) {
  if (!componentRecipe.value) return
  await store.removeRecipeComponent(componentRecipe.value.id, c.input_definition_id)
  recipeComponents.value = await store.loadRecipeComponents(componentRecipe.value.id)
}
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
