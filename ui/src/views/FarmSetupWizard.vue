<template>
  <div class="p-6 max-w-3xl mx-auto space-y-6" data-test="farm-setup-wizard">
    <div>
      <h1 class="text-xl font-semibold text-white">Set up your farm</h1>
      <p class="text-zinc-500 text-sm mt-1">
        Choose how much starter data to load for <strong class="text-zinc-300">{{ farmLabel }}</strong>.
        Templates are idempotent — applying the same pack again does not duplicate rows.
      </p>
    </div>

    <div class="flex gap-2 text-[10px] uppercase tracking-wide text-zinc-500" aria-label="Wizard steps">
      <span :class="step === 'choose' ? 'text-green-400' : ''">1 Choose</span>
      <span>›</span>
      <span :class="step === 'preview' ? 'text-green-400' : ''">2 Preview</span>
      <span>›</span>
      <span :class="step === 'done' ? 'text-green-400' : ''">3 Finish</span>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>

    <!-- Step 1 — Choose -->
    <template v-if="step === 'choose'">
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <button
          v-for="card in primaryChoices"
          :key="card.id"
          type="button"
          class="text-left rounded-xl border p-4 transition-colors"
          :class="selectedChoice === card.id
            ? 'border-green-600 bg-green-950/30'
            : 'border-zinc-800 bg-zinc-900 hover:border-zinc-600'"
          :data-test="`setup-card-${card.id}`"
          @click="selectChoice(card.id)"
        >
          <div class="flex items-start justify-between gap-2 mb-2">
            <span class="text-2xl" aria-hidden="true">{{ card.icon }}</span>
            <span
              v-if="card.recommended"
              class="text-[10px] uppercase tracking-wide text-green-400 border border-green-800 rounded px-1.5 py-0.5"
            >Recommended</span>
          </div>
          <p class="text-sm font-medium text-white">{{ card.label }}</p>
          <p class="text-xs text-zinc-500 mt-1">{{ card.tagline }}</p>
        </button>
      </div>

      <details class="text-sm text-zinc-400">
        <summary class="cursor-pointer text-zinc-300 hover:text-white select-none">More starter packs</summary>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 mt-3">
          <button
            v-for="card in moreChoices"
            :key="card.id"
            type="button"
            class="text-left rounded-lg border border-zinc-800 bg-zinc-900/80 px-3 py-2 hover:border-zinc-600"
            :class="selectedChoice === card.id ? 'border-green-700' : ''"
            :data-test="`setup-card-${card.id}`"
            @click="selectChoice(card.id)"
          >
            <p class="text-sm text-zinc-200">{{ card.label }}</p>
            <p class="text-[11px] text-zinc-500">{{ card.tagline }}</p>
          </button>
        </div>
      </details>

      <div class="flex flex-wrap gap-2 pt-2">
        <button
          type="button"
          :disabled="!selectedChoice"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
          data-test="setup-continue-preview"
          @click="goPreview"
        >
          Continue
        </button>
        <router-link to="/settings" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200">
          Back to Settings
        </router-link>
      </div>
    </template>

    <!-- Step 2 — Preview & confirm -->
    <template v-else-if="step === 'preview'">
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">
          {{ preview.isBlank ? 'Blank farm' : 'What this template creates' }}
        </h2>
        <p class="text-xs text-zinc-500">{{ preview.title }}</p>
        <ul class="list-disc pl-5 text-sm text-zinc-300 space-y-1.5">
          <li v-for="(line, i) in preview.bullets" :key="i">{{ line }}</li>
        </ul>
        <p v-if="!preview.isBlank" class="text-[11px] text-amber-300/90 border border-amber-900/50 rounded-lg px-3 py-2 bg-amber-950/20">
          Farm admins only. If this pack was applied before, the API returns “already applied” and leaves data unchanged.
        </p>
      </section>

      <p v-if="applyError" class="text-sm text-red-400">{{ applyError }}</p>

      <div class="flex flex-wrap gap-2">
        <button
          v-if="preview.isBlank"
          type="button"
          :disabled="applying"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
          data-test="setup-finish-blank"
          @click="finishBlank"
        >
          Continue with blank farm
        </button>
        <button
          v-else
          type="button"
          :disabled="applying"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
          data-test="setup-apply-template"
          @click="applyTemplate"
        >
          {{ applying ? 'Applying…' : 'Apply starter pack' }}
        </button>
        <button
          type="button"
          class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200"
          @click="step = 'choose'"
        >
          Back
        </button>
      </div>
      <section class="pt-4 border-t border-zinc-800 space-y-2" data-test="farm-setup-guardian-help">
        <p class="text-[10px] uppercase tracking-widest text-zinc-500">Need help?</p>
        <GuardianStarterChips :starters="farmSetupStarters" />
      </section>
    </template>

    <!-- Step 3 — Done -->
    <template v-else>
      <section class="bg-zinc-900 border border-green-900/50 rounded-xl p-4 space-y-2">
        <p class="text-sm text-green-300 font-medium">{{ doneMessage }}</p>
        <p v-if="zoneCount > 0" class="text-xs text-zinc-500">
          {{ zoneCount }} zone{{ zoneCount === 1 ? '' : 's' }} on this farm now.
        </p>
      </section>
      <div class="flex flex-wrap gap-2">
        <router-link
          v-if="farmId"
          :to="`/farms/${farmId}/zones/new`"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
          data-test="setup-add-zone"
        >
          Add a zone
        </router-link>
        <router-link
          to="/zones"
          class="px-4 py-2 text-sm text-zinc-300 border border-zinc-700 rounded-lg"
          data-test="setup-go-zones"
        >
          Open My zones
        </router-link>
        <router-link
          to="/"
          class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200"
        >
          Go to Today
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import {
  FARM_SETUP_PRIMARY_CHOICES,
  FARM_SETUP_BLANK_ID,
  farmSetupMoreChoices,
  previewForSetupChoice,
  formatBootstrapApplyResult,
  markFarmSetupComplete,
} from '../lib/farmSetupWizard.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const step = ref('choose')
const selectedChoice = ref(FARM_SETUP_BLANK_ID)
const applying = ref(false)
const applyError = ref('')
const doneMessage = ref('')
const loadError = ref('')
const zoneCount = ref(0)

const farmId = computed(() => {
  const raw = route.params.id
  const n = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

const farmLabel = computed(() => {
  const f = farmContext.farms.find((x) => x.id === farmId.value)
  return f?.name || store.farm?.name || (farmId.value ? `Farm #${farmId.value}` : 'this farm')
})

const primaryChoices = FARM_SETUP_PRIMARY_CHOICES
const moreChoices = farmSetupMoreChoices()

const preview = computed(() => previewForSetupChoice(selectedChoice.value))

const farmSetupStarters = computed(() => buildSetupStarters({
  surface: 'farm_setup_wizard',
  farmId: farmId.value,
  zoneCount: store.zones?.length ?? zoneCount.value,
  zones: store.zones || [],
}))

function selectChoice(id) {
  selectedChoice.value = id
}

function goPreview() {
  if (!selectedChoice.value) return
  applyError.value = ''
  step.value = 'preview'
}

async function ensureFarmContext() {
  loadError.value = ''
  if (!farmId.value) {
    loadError.value = 'Invalid farm id in URL.'
    return false
  }
  if (!farmContext.farms.length) {
    try {
      await farmContext.fetchFarms()
    } catch (e) {
      loadError.value = e.response?.data?.error || 'Could not load farms'
      return false
    }
  }
  if (!farmContext.farms.some((f) => f.id === farmId.value)) {
    loadError.value = 'Farm not found or you do not have access.'
    return false
  }
  if (farmContext.farmId !== farmId.value) {
    await farmContext.selectFarm(farmId.value)
  }
  zoneCount.value = store.zones.length
  return true
}

function finishDone(message) {
  markFarmSetupComplete(farmId.value)
  doneMessage.value = message
  zoneCount.value = store.zones.length
  step.value = 'done'
}

async function finishBlank() {
  finishDone('Your farm is ready. Add zones when you are ready to start.')
}

async function applyTemplate() {
  if (!farmId.value || selectedChoice.value === FARM_SETUP_BLANK_ID) {
    await finishBlank()
    return
  }
  applying.value = true
  applyError.value = ''
  try {
    const data = await farmContext.applyBootstrapTemplate(farmId.value, selectedChoice.value)
    const result = formatBootstrapApplyResult(data?.bootstrap)
    if (!result.ok) {
      applyError.value = result.message
      return
    }
    finishDone(result.message)
  } catch (e) {
    applyError.value = e.response?.data?.error || e.message || 'Could not apply starter pack'
  } finally {
    applying.value = false
  }
}

onMounted(() => {
  void ensureFarmContext()
})

watch(farmId, () => {
  void ensureFarmContext()
})
</script>
