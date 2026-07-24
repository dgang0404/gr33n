<template>
  <div class="space-y-6 max-w-3xl" data-test="nf-make-batch">
    <div>
      <h2 class="text-lg font-semibold text-white flex items-center gap-1">
        {{ NF_VOCAB.makeBatch }}
        <ConceptHelpTip concept-id="input_batch" position="bottom" />
      </h2>
      <p class="text-sm text-zinc-500 mt-1">
        Pick an {{ NF_VOCAB.input.toLowerCase() }}, follow the field guide, then record a {{ NF_VOCAB.batch.toLowerCase() }} on this farm.
      </p>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>
    <p v-else-if="loading" class="text-sm text-zinc-500">Loading recipe canon…</p>

    <template v-else>
      <!-- Process type -->
      <section class="space-y-3" data-test="nf-batch-process-picker">
        <h3 class="text-sm font-medium text-zinc-200">1. Process type</h3>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="pt in processTypes"
            :key="pt.id"
            type="button"
            class="px-3 py-2 rounded-lg border text-sm transition-colors"
            :class="processType === pt.id
              ? 'border-green-600 bg-green-950/40 text-white'
              : 'border-zinc-800 bg-zinc-900 text-zinc-300 hover:border-zinc-600'"
            :data-test="`nf-batch-process-${pt.id}`"
            @click="selectProcess(pt.id)"
          >
            <span class="font-medium">{{ pt.label }}</span>
            <span class="text-zinc-500 text-xs ml-1">{{ pt.expand }}</span>
            <span
              v-if="pt.tradition === 'knf'"
              class="ml-1 text-[10px] uppercase text-amber-400/90"
            >KNF</span>
          </button>
        </div>
      </section>

      <!-- Variant -->
      <section v-if="processType && variants.length" class="space-y-3" data-test="nf-batch-variant-picker">
        <h3 class="text-sm font-medium text-zinc-200">2. Variant</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
          <button
            v-for="v in variants"
            :key="v.seed_name"
            type="button"
            class="text-left rounded-xl border p-3 transition-colors"
            :class="selectedVariant?.seed_name === v.seed_name
              ? 'border-green-600 bg-green-950/30'
              : 'border-zinc-800 bg-zinc-900 hover:border-zinc-600'"
            :data-test="`nf-batch-variant-${v.process_type}-${slugify(v.seed_name)}`"
            @click="selectVariant(v)"
          >
            <p class="text-sm font-medium text-white">{{ v.seed_name }}</p>
            <p v-if="canonDilutionHint(v)" class="text-xs text-green-400/90 mt-1 font-mono">
              {{ canonDilutionHint(v) }}
            </p>
          </button>
        </div>
        <LearnHowExpander v-if="selectedVariant?.guide" :guide-file="selectedVariant.guide" />
      </section>

      <!-- Step cards -->
      <section v-if="selectedVariant" class="space-y-3" data-test="nf-batch-step-cards">
        <h3 class="text-sm font-medium text-zinc-200">3. Follow the guide</h3>
        <p v-if="guideLoading" class="text-xs text-zinc-500">Loading field guide…</p>
        <p v-else-if="guideError" class="text-xs text-red-400">{{ guideError }}</p>
        <div v-else class="space-y-3">
          <GuideStepCards :cards="stepCards" />
          <p v-if="!stepCards.length" class="text-sm text-zinc-500">
            Guide sections not loaded — re-ingest field guides or open Help → Library.
          </p>
        </div>
      </section>

      <!-- Create batch -->
      <section
        v-if="selectedVariant && stepCards.length"
        class="space-y-3 border-t border-zinc-800 pt-4"
        data-test="nf-batch-create-form"
      >
        <h3 class="text-sm font-medium text-zinc-200">4. Start batch on farm</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <label class="block text-xs text-zinc-500">
            Batch label
            <input
              v-model="form.batch_identifier"
              type="text"
              class="input-field mt-1 w-full"
              placeholder="e.g. JMS March week 1"
              data-test="nf-batch-identifier"
            />
          </label>
          <label class="block text-xs text-zinc-500">
            Status
            <select v-model="form.status" class="input-field mt-1 w-full" data-test="nf-batch-status">
              <option value="planning">Planning</option>
              <option value="fermenting_brewing">Fermenting</option>
              <option value="ready_for_use">Ready for use</option>
            </select>
          </label>
          <label class="block text-xs text-zinc-500 sm:col-span-2">
            Notes
            <input
              v-model="form.observations_notes"
              type="text"
              class="input-field mt-1 w-full"
              placeholder="Optional observations"
            />
          </label>
          <label class="flex items-center gap-2 text-xs text-zinc-400 sm:col-span-2">
            <input v-model="form.create_prep_task" type="checkbox" data-test="nf-batch-prep-task" />
            Also create a prep task (jadam_prep)
          </label>
        </div>
        <p v-if="actionError" class="text-sm text-red-400">{{ actionError }}</p>
        <p v-if="actionSuccess" class="text-sm text-green-400">{{ actionSuccess }}</p>
        <div class="flex flex-wrap gap-2">
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
            :disabled="saving || !farmId"
            data-test="nf-batch-submit"
            @click="submitBatch"
          >
            {{ saving ? 'Saving…' : 'Create input & batch' }}
          </button>
          <router-link
            :to="moneySuppliesRoute()"
            class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200"
          >
            View in Money →
          </router-link>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFarmStore } from '../../stores/farm.js'
import { useFarmContextStore } from '../../stores/farmContext.js'
import { loadRecipeCanon } from '../../lib/naturalFarmingCanon.js'
import { batchStepCards } from '../../lib/naturalFarmingGuideSections.js'
import {
  batchCreatePayload,
  buildInputPayload,
  canonDilutionHint,
  findFarmInputByName,
  loadFieldGuideBody,
  prepTaskPayload,
  processTypesFromCanon,
  variantsForProcess,
} from '../../lib/naturalFarmingBatchFlow.js'
import LearnHowExpander from './LearnHowExpander.vue'
import GuideStepCards from './GuideStepCards.vue'
import ConceptHelpTip from '../ConceptHelpTip.vue'
import { NF_VOCAB } from '../../lib/naturalFarmingVocabulary.js'
import { moneySuppliesRoute } from '../../lib/workspaceRoutes.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)

const canon = ref(null)
const loading = ref(true)
const loadError = ref('')
const processType = ref('')
const selectedVariant = ref(null)
const guideBody = ref('')
const guideLoading = ref(false)
const guideError = ref('')
const farmInputs = ref([])
const saving = ref(false)
const actionError = ref('')
const actionSuccess = ref('')

const form = ref({
  batch_identifier: '',
  status: 'fermenting_brewing',
  observations_notes: '',
  create_prep_task: true,
})

const processTypes = computed(() => processTypesFromCanon(canon.value ?? {}))
const variants = computed(() =>
  processType.value ? variantsForProcess(processType.value, canon.value ?? {}) : [],
)
const stepCards = computed(() => batchStepCards(guideBody.value))

function slugify(name) {
  return String(name).toLowerCase().replace(/[^a-z0-9]+/g, '-').slice(0, 40)
}

function selectProcess(id) {
  processType.value = id
  selectedVariant.value = null
  guideBody.value = ''
  const vars = variantsForProcess(id, canon.value ?? {})
  if (vars.length === 1) selectVariant(vars[0])
}

async function selectVariant(variant) {
  selectedVariant.value = variant
  guideLoading.value = true
  guideError.value = ''
  actionError.value = ''
  actionSuccess.value = ''
  try {
    guideBody.value = await loadFieldGuideBody(variant.guide)
  } catch (err) {
    guideError.value = err?.response?.data?.error || err?.message || 'Could not load field guide'
    guideBody.value = ''
  } finally {
    guideLoading.value = false
  }
}

async function loadFarmInputs() {
  if (!farmId.value) return
  farmInputs.value = await store.loadNfInputs(farmId.value)
}

async function submitBatch() {
  const fid = farmId.value
  if (!fid || !selectedVariant.value) return
  saving.value = true
  actionError.value = ''
  actionSuccess.value = ''
  try {
    let inputDef = findFarmInputByName(farmInputs.value, selectedVariant.value.seed_name)
    if (!inputDef) {
      inputDef = await store.createNfInput(fid, buildInputPayload(selectedVariant.value, guideBody.value))
      farmInputs.value = await store.loadNfInputs(fid)
    }
    const batchPayload = {
      input_definition_id: inputDef.id,
      ...batchCreatePayload(selectedVariant.value, guideBody.value, form.value),
    }
    await store.createNfBatch(fid, batchPayload)
    if (form.value.create_prep_task) {
      await store.createTask(fid, prepTaskPayload(selectedVariant.value, guideBody.value))
    }
    actionSuccess.value = `Batch started for ${selectedVariant.value.seed_name}.`
    form.value.batch_identifier = ''
    form.value.observations_notes = ''
  } catch (err) {
    actionError.value = err?.response?.data?.error || err?.message || 'Could not create batch'
  } finally {
    saving.value = false
  }
}

function applyRouteProcess() {
  const raw = route.query.process
  const p = typeof raw === 'string' ? raw.trim() : ''
  if (!p || !canon.value) return
  if (processTypes.value.some((t) => t.id === p)) selectProcess(p)
}

onMounted(async () => {
  try {
    canon.value = await loadRecipeCanon()
    await loadFarmInputs()
    applyRouteProcess()
  } catch (err) {
    loadError.value = err?.message || 'Could not load recipe canon'
  } finally {
    loading.value = false
  }
})

watch(farmId, loadFarmInputs)
watch(() => route.query.process, applyRouteProcess)
</script>
