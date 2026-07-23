<template>
  <div class="space-y-6" data-test="nf-switchover-wizard">
    <nav
      class="flex flex-wrap gap-x-2 gap-y-1 text-xs uppercase tracking-wide text-zinc-500 border-b border-zinc-800/80 pb-3"
      aria-label="Switchover steps"
      data-test="nf-switchover-step-rail"
    >
      <span
        v-for="(id, idx) in steps"
        :key="id"
        class="inline-flex items-center gap-1"
        :class="step === id ? 'text-green-400 font-medium' : ''"
        :aria-current="step === id ? 'step' : undefined"
      >
        <span class="tabular-nums">{{ idx + 1 }}</span>
        <span>{{ stepLabels[id] }}</span>
        <span v-if="idx < steps.length - 1" class="text-zinc-700 px-1" aria-hidden="true">·</span>
      </span>
    </nav>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>
    <p v-else-if="loading" class="text-sm text-zinc-500">Loading canonical recipes…</p>

    <template v-else>
      <!-- Step 1 — context -->
      <section v-if="step === 'context'" class="space-y-4" data-test="nf-switchover-step-context">
        <div>
          <h2 class="text-lg font-semibold text-white">Switch from bottle nutrients</h2>
          <p class="text-sm text-zinc-500 mt-1 max-w-2xl">
            Map what you already run (EC, A+B, salts) to audited natural farming recipes — dilutions come from
            canonical seed data, not guesses.
          </p>
        </div>
        <h3 class="text-sm font-medium text-zinc-200">What are you growing today?</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <button
            v-for="opt in CONTEXT_OPTIONS"
            :key="opt.id"
            type="button"
            class="text-left rounded-xl border p-4 transition-colors"
            :class="contextId === opt.id
              ? 'border-green-600 bg-green-950/30'
              : 'border-zinc-800 bg-zinc-900 hover:border-zinc-600'"
            :data-test="`nf-context-${opt.id}`"
            @click="contextId = opt.id"
          >
            <p class="text-sm font-medium text-white">{{ opt.label }}</p>
            <p class="text-xs text-zinc-500 mt-1">{{ opt.hint }}</p>
          </button>
        </div>
        <LearnHowExpander :guide-file="learnGuideForStep('context', contextId)" />
        <div class="flex gap-2">
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
            :disabled="!contextId"
            data-test="nf-switchover-next"
            @click="step = 'pattern'"
          >
            Continue
          </button>
        </div>
      </section>

      <!-- Step 2 — commercial pattern -->
      <section v-else-if="step === 'pattern'" class="space-y-4" data-test="nf-switchover-step-pattern">
        <h3 class="text-sm font-medium text-zinc-200">What commercial pattern do you use now?</h3>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <button
            v-for="opt in COMMERCIAL_PATTERNS"
            :key="opt.id"
            type="button"
            class="text-left rounded-xl border p-4 transition-colors"
            :class="patternId === opt.id
              ? 'border-green-600 bg-green-950/30'
              : 'border-zinc-800 bg-zinc-900 hover:border-zinc-600'"
            :data-test="`nf-pattern-${opt.id}`"
            @click="patternId = opt.id"
          >
            <p class="text-sm font-medium text-white">{{ opt.label }}</p>
            <p class="text-xs text-zinc-500 mt-1">{{ opt.hint }}</p>
          </button>
        </div>
        <LearnHowExpander :guide-file="learnGuideForStep('pattern', contextId, patternId)" />
        <div class="flex flex-wrap gap-2">
          <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'context'">
            Back
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
            :disabled="!patternId"
            data-test="nf-switchover-next"
            @click="step = 'mapping'"
          >
            Continue
          </button>
        </div>
      </section>

      <!-- Step 3 — mapped program -->
      <section v-else-if="step === 'mapping'" class="space-y-4" data-test="nf-switchover-step-mapping">
        <h3 class="text-sm font-medium text-zinc-200">Your natural farming equivalent</h3>
        <p class="text-xs text-zinc-500">
          Replacing: <strong class="text-zinc-300">{{ mapping.commercialLabel }}</strong>
        </p>
        <ul class="space-y-3">
          <li
            v-for="(row, i) in mapping.naturalEquivalent"
            :key="i"
            class="rounded-xl border border-zinc-800 bg-zinc-900/80 p-4"
            :data-test="`nf-mapped-recipe-${i}`"
          >
            <p class="text-sm font-medium text-white">{{ row.recipe }}</p>
            <p v-if="row.frequency" class="text-xs text-zinc-400 mt-1">{{ row.frequency }}</p>
            <p v-if="row.dilution" class="text-xs text-green-400/90 mt-1 font-mono">{{ row.dilution }}</p>
          </li>
        </ul>
        <LearnHowExpander :guide-file="mapping.naturalEquivalent[0]?.guide || mapping.summaryGuide" />
        <div class="flex flex-wrap gap-2">
          <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'pattern'">
            Back
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
            data-test="nf-switchover-next"
            @click="step = 'first-batch'"
          >
            Continue
          </button>
        </div>
      </section>

      <!-- Step 4 — first batch -->
      <section v-else-if="step === 'first-batch'" class="space-y-4" data-test="nf-switchover-step-first-batch">
        <h3 class="text-sm font-medium text-zinc-200">Pick your first batch</h3>
        <p class="text-xs text-zinc-500">
          Ferment one concentrate first — you'll need it before combined drenches. Click a card to select, then continue.
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <button
            v-for="input in firstBatches"
            :key="input.seed_name"
            type="button"
            class="text-left rounded-xl border p-4 transition-colors"
            :class="selectedFirstBatchSeed === input.seed_name
              ? 'border-green-600 bg-green-950/30 ring-1 ring-green-900/40'
              : 'border-zinc-800 bg-zinc-900/80 hover:border-zinc-600'"
            :data-test="`nf-first-batch-${input.process_type}`"
            :aria-pressed="selectedFirstBatchSeed === input.seed_name"
            @click="selectedFirstBatchSeed = input.seed_name"
          >
            <p class="text-sm font-medium text-white">{{ input.seed_name }}</p>
            <p class="text-xs text-zinc-500 mt-1 capitalize">{{ input.tradition }} · {{ input.process_type }}</p>
            <p v-if="input.dilution_start" class="text-xs text-zinc-400 mt-2">
              Apply from {{ input.dilution_start }}
            </p>
          </button>
        </div>
        <LearnHowExpander :guide-file="learnGuideForInput(selectedFirstBatchInput)" />
        <div class="flex flex-wrap gap-2">
          <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'mapping'">
            Back
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
            :disabled="!selectedFirstBatchInput"
            data-test="nf-switchover-next"
            @click="step = 'actions'"
          >
            Continue
          </button>
        </div>
      </section>

      <!-- Step 5 — CTAs -->
      <section v-else class="space-y-4" data-test="nf-switchover-step-actions">
        <h3 class="text-sm font-medium text-zinc-200">Ready to start</h3>
        <p class="text-xs text-zinc-500 max-w-xl">
          {{ actionsStepIntro(contextId) }}
        </p>
        <div class="flex flex-col sm:flex-row flex-wrap gap-2">
          <button
            v-if="selectedFirstBatchInput"
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
            data-test="nf-cta-make-batch"
            @click="goMakeBatch(selectedFirstBatchInput)"
          >
            Make {{ selectedFirstBatchShortName }} batch
          </button>
          <button
            v-for="input in otherFirstBatches"
            :key="input.seed_name"
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg border border-zinc-600 text-zinc-200 hover:border-zinc-400"
            :data-test="`nf-cta-make-${input.process_type}`"
            @click="goMakeBatch(input)"
          >
            Make {{ shortBatchName(input) }} batch
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg border border-green-700/80 text-green-100 hover:bg-green-950/40 disabled:opacity-40"
            :disabled="applyingPack || !farmId || !mapping.switchoverPackKey"
            data-test="nf-cta-apply-switchover-pack"
            :title="mapping.switchoverPackKey ? 'Imports curated recipes and inputs onto this farm' : 'No pack for this bottle pattern'"
            @click="applySwitchoverPack"
          >
            {{ applyingPack ? 'Applying…' : 'Apply switchover pack to farm' }}
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium rounded-lg border border-amber-700/80 text-amber-100 hover:bg-amber-950/40 disabled:opacity-40"
            :disabled="applyingBootstrap || !farmId"
            data-test="nf-cta-apply-bootstrap"
            title="Seeds zones, programs, and application recipes on this farm — does not dose a mixing tank"
            @click="applyBootstrap"
          >
            {{ applyingBootstrap ? 'Applying…' : bootstrapApplyButtonLabel(contextId) }}
          </button>
        </div>
        <p v-if="packMessage" class="text-sm" :class="packOk ? 'text-green-400' : 'text-red-400'">
          {{ packMessage }}
        </p>
        <p v-if="bootstrapMessage" class="text-sm" :class="bootstrapOk ? 'text-green-400' : 'text-red-400'">
          {{ bootstrapMessage }}
        </p>
        <LearnHowExpander :guide-file="learnGuideForStep('actions', contextId)" />
        <button type="button" class="text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'first-batch'">
          Back
        </button>
      </section>
    </template>

    <section v-if="!loading && !loadError" class="space-y-4 pt-4 border-t border-zinc-800/80">
      <CommonsRecipePackImport />
      <router-link
        :to="{ path: '/natural-farming', query: { tab: 'library' } }"
        class="inline-block text-xs text-green-400 hover:text-green-300"
        data-test="nf-switchover-library-link"
      >
        Browse full recipe library →
      </router-link>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useFarmContextStore } from '../../stores/farmContext.js'
import { loadRecipeCanon } from '../../lib/naturalFarmingCanon.js'
import {
  CONTEXT_OPTIONS,
  COMMERCIAL_PATTERNS,
  SWITCHOVER_STEPS,
  batchTabQueryForInput,
  bootstrapTemplateForContext,
  firstBatchSuggestions,
  learnGuideForInput,
  learnGuideForStep,
  actionsStepIntro,
  bootstrapApplyButtonLabel,
  resolveSwitchoverMapping,
} from '../../lib/naturalFarmingSwitchover.js'
import { formatBootstrapApplyResult } from '../../lib/farmSetupWizard.js'
import LearnHowExpander from './LearnHowExpander.vue'
import CommonsRecipePackImport from './CommonsRecipePackImport.vue'

const router = useRouter()
const farmContext = useFarmContextStore()
const { farmId } = storeToRefs(farmContext)

const steps = SWITCHOVER_STEPS
const stepLabels = {
  context: 'Where you grow',
  pattern: 'Bottle program',
  mapping: 'Natural match',
  'first-batch': 'First batch',
  actions: 'Apply',
}

const step = ref('context')
const contextId = ref('')
const patternId = ref('')
const canon = ref(null)
const loading = ref(true)
const loadError = ref('')
const applyingBootstrap = ref(false)
const bootstrapMessage = ref('')
const bootstrapOk = ref(true)
const applyingPack = ref(false)
const packMessage = ref('')
const packOk = ref(true)
const selectedFirstBatchSeed = ref('')

const mapping = computed(() =>
  resolveSwitchoverMapping(contextId.value, patternId.value, canon.value ?? {}),
)

const firstBatches = computed(() => firstBatchSuggestions(canon.value ?? {}))

const selectedFirstBatchInput = computed(() =>
  firstBatches.value.find((b) => b.seed_name === selectedFirstBatchSeed.value) ?? null,
)

const otherFirstBatches = computed(() =>
  firstBatches.value.filter((b) => b.seed_name !== selectedFirstBatchSeed.value),
)

const selectedFirstBatchShortName = computed(() => shortBatchName(selectedFirstBatchInput.value))

watch(
  () => step.value,
  (s) => {
    if (s === 'first-batch' && !selectedFirstBatchSeed.value && firstBatches.value[0]) {
      selectedFirstBatchSeed.value = firstBatches.value[0].seed_name
    }
  },
)

watch(firstBatches, (list) => {
  if (!selectedFirstBatchSeed.value && list[0]) {
    selectedFirstBatchSeed.value = list[0].seed_name
  }
})

function shortBatchName(input) {
  const pt = String(input?.process_type ?? '').toUpperCase()
  if (pt) return pt
  const name = String(input?.seed_name ?? 'batch')
  return name.split('(')[0].trim().split(/\s+/)[0] || 'batch'
}

onMounted(async () => {
  try {
    canon.value = await loadRecipeCanon()
  } catch (err) {
    loadError.value = err?.message || 'Could not load recipe canon'
  } finally {
    loading.value = false
  }
})

function goMakeBatch(input) {
  router.push({ path: '/natural-farming', query: batchTabQueryForInput(input) })
}

async function applySwitchoverPack() {
  const packKey = mapping.value.switchoverPackKey
  if (!farmId.value) {
    packOk.value = false
    packMessage.value = 'Select a farm first (top bar).'
    return
  }
  if (!packKey) {
    packOk.value = false
    packMessage.value = 'No switchover pack for this pattern — use starter bootstrap or import full catalog pack.'
    return
  }
  applyingPack.value = true
  packMessage.value = ''
  try {
    const data = await farmContext.applyNaturalFarmingPack(farmId.value, packKey)
    packOk.value = data?.status === 'applied' || data?.status === 'already_applied' || data?.status === 'noop'
    packMessage.value = data?.message || 'Switchover pack applied.'
  } catch (err) {
    packOk.value = false
    packMessage.value = err?.response?.data?.error || err?.message || 'Apply failed'
  } finally {
    applyingPack.value = false
  }
}

async function applyBootstrap() {
  if (!farmId.value) {
    bootstrapOk.value = false
    bootstrapMessage.value = 'Select a farm first (top bar).'
    return
  }
  applyingBootstrap.value = true
  bootstrapMessage.value = ''
  try {
    const template = bootstrapTemplateForContext(contextId.value)
    const data = await farmContext.applyBootstrapTemplate(farmId.value, template)
    const result = formatBootstrapApplyResult(data)
    bootstrapOk.value = result.ok
    bootstrapMessage.value = result.message
  } catch (err) {
    bootstrapOk.value = false
    bootstrapMessage.value = err?.response?.data?.error || err?.message || 'Apply failed'
  } finally {
    applyingBootstrap.value = false
  }
}
</script>
