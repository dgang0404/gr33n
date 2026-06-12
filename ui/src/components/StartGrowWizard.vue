<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
    data-test="start-grow-wizard"
    @click.self="close"
  >
    <div class="w-full max-w-lg bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4">
      <div class="flex items-start justify-between gap-2">
        <div>
          <h2 class="text-white font-semibold">Start a grow</h2>
          <p class="text-zinc-500 text-xs mt-0.5">Pick plant type, zone, and optional feeding program.</p>
        </div>
        <button type="button" class="text-zinc-500 hover:text-zinc-300 text-sm" @click="close">✕</button>
      </div>

      <div class="space-y-3">
        <div>
          <CropLibraryPicker
            :farm-id="farmId"
            v-model="form.cropProfileId"
            label="Plant type (knowledge base)"
            required
            data-test="start-grow-crop-profile"
            @select="onCropProfileSelect"
          />
          <p class="text-[10px] text-zinc-600 mt-1">
            EC, DLI, and photoperiod targets for Guardian and the zone grow strip.
            Tune per crop in Settings → Crops &amp; targets.
          </p>
        </div>

        <div>
          <label class="block text-xs text-zinc-500 mb-1">Batch label (optional — variety or room note)</label>
          <input
            v-model="form.strain"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
            placeholder="e.g. Veg Room A, Roma variety, Batch 3"
            data-test="start-grow-strain"
          />
          <select
            v-if="plants.length"
            v-model="plantPickId"
            class="w-full mt-2 bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="start-grow-plant-pick"
            @change="onPlantPick"
          >
            <option :value="''">Or pick from Plants…</option>
            <option v-for="p in plants" :key="p.id" :value="p.id">{{ p.display_name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-xs text-zinc-500 mb-1">Zone *</label>
          <select
            v-model.number="form.zoneId"
            required
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="start-grow-zone"
          >
            <option :value="null" disabled>Select zone</option>
            <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
          </select>
        </div>

        <div>
          <label class="block text-xs text-zinc-500 mb-1">Cycle name</label>
          <input
            v-model="form.name"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            placeholder="Auto-filled from label + zone"
            data-test="start-grow-name"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Starting stage</label>
            <select v-model="form.stage" class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white">
              <option v-for="gs in growthStages" :key="gs" :value="gs">{{ formatStageLabel(gs) }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Started on</label>
            <input v-model="form.startedAt" type="date" class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white" />
          </div>
        </div>

        <div>
          <label class="block text-xs text-zinc-500 mb-1">Feeding program (optional)</label>
          <div
            v-if="programMismatchWarning"
            class="mb-2 rounded-lg border border-amber-800/60 bg-amber-950/30 px-3 py-2 text-[11px] text-amber-200/90"
            data-test="start-grow-program-mismatch"
          >
            {{ programMismatchWarning }}
          </div>
          <select
            v-model.number="form.programId"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="start-grow-program"
          >
            <option :value="null">None — assign later</option>
            <option
              v-for="p in sortedZonePrograms"
              :key="p.id"
              :value="p.id"
            >
              {{ p.name }}{{ programOptionSuffix(p, programFitContext) }}
            </option>
          </select>
          <p v-if="selectedProgramBand" class="text-[10px] text-zinc-600 mt-1">{{ selectedProgramBand }}</p>
        </div>
      </div>

      <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>

      <div class="flex justify-end gap-3 pt-1">
        <button
          type="button"
          class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          @click="close"
        >
          Cancel
        </button>
        <button
          type="button"
          class="px-4 py-1.5 text-xs rounded-lg bg-green-700 hover:bg-green-600 text-white font-medium disabled:opacity-40"
          :disabled="submitting || !canSubmit"
          data-test="start-grow-submit"
          @click="submit"
        >
          {{ submitting ? 'Starting…' : 'Start grow' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import api from '../api/index.js'
import {
  GROWTH_STAGES,
  buildStartGrowPayload,
  defaultCycleName,
  formatStageLabel,
  getGrowthStages,
  strainFromPlant,
} from '../lib/growHub.js'
import { loadDomainEnums } from '../lib/domainEnums.js'
import {
  programFitResult,
  programOptionSuffix,
  sortProgramsByFit,
  parseProgramMeta,
} from '../lib/programFit.js'
import CropLibraryPicker from './CropLibraryPicker.vue'

const props = defineProps({
  open: { type: Boolean, default: false },
  farmId: { type: Number, required: true },
  zones: { type: Array, default: () => [] },
  programs: { type: Array, default: () => [] },
  plants: { type: Array, default: () => [] },
  initialZoneId: { type: Number, default: null },
  initialStrain: { type: String, default: '' },
  initialPlantId: { type: Number, default: null },
})

const emit = defineEmits(['close', 'created'])

const store = useFarmStore()
const growthStages = ref([...GROWTH_STAGES])
const submitting = ref(false)
const formError = ref('')
const plantPickId = ref('')

function onCropProfileSelect(item) {
  if (!item?.crop_key) return
  form.value.cropKey = item.crop_key
  if (!form.value.strain?.trim()) {
    form.value.strain = item.display_name || item.crop_key
  }
}

const form = ref(emptyForm())

function emptyForm() {
  return {
    strain: '',
    zoneId: null,
    name: '',
    stage: 'seedling',
    startedAt: new Date().toISOString().slice(0, 10),
    programId: null,
    cropProfileId: null,
    cropKey: '',
  }
}

const zonePrograms = computed(() => {
  if (!form.value.zoneId) return props.programs
  return props.programs.filter((p) => Number(p.target_zone_id) === Number(form.value.zoneId))
})

const programFitContext = computed(() => ({
  cropKey: form.value.cropKey,
  stage: form.value.stage,
}))

const sortedZonePrograms = computed(() =>
  sortProgramsByFit(zonePrograms.value, programFitContext.value),
)

const selectedProgram = computed(() =>
  sortedZonePrograms.value.find((p) => Number(p.id) === Number(form.value.programId)) || null,
)

const programMismatchWarning = computed(() => {
  if (!selectedProgram.value) return ''
  const fit = programFitResult(selectedProgram.value, programFitContext.value)
  return fit.warnings[0] || ''
})

const selectedProgramBand = computed(() => {
  const band = parseProgramMeta(selectedProgram.value?.metadata).ec_band_mscm
  if (!band || band.min == null) return ''
  return `Program EC band: ${band.min}–${band.max} mS/cm (from catalog profile)`
})

const selectedZone = computed(() =>
  props.zones.find((z) => Number(z.id) === Number(form.value.zoneId)),
)

const canSubmit = computed(() =>
  Boolean(
    props.farmId &&
      form.value.zoneId &&
      form.value.cropProfileId &&
      form.value.cropKey &&
      (form.value.strain?.trim() || form.value.name?.trim()),
  ),
)

watch(
  () => [form.value.strain, form.value.zoneId],
  () => {
    if (!form.value.name || form.value.name === lastAutoName.value) {
      const auto = defaultCycleName(form.value.strain, selectedZone.value?.name)
      form.value.name = auto
      lastAutoName.value = auto
    }
  },
)

const lastAutoName = ref('')

watch(
  () => props.open,
  async (isOpen) => {
    if (!isOpen) return
    void loadDomainEnums(api).then((enums) => {
      growthStages.value = getGrowthStages(enums)
    })
    formError.value = ''
    plantPickId.value = ''
    const f = emptyForm()
    if (props.initialStrain) f.strain = props.initialStrain
    if (props.initialZoneId) f.zoneId = props.initialZoneId
    if (props.initialPlantId) plantPickId.value = String(props.initialPlantId)
    f.name = defaultCycleName(f.strain, props.zones.find((z) => z.id === f.zoneId)?.name)
    lastAutoName.value = f.name
    form.value = f
  },
)

watch(
  () => form.value.zoneId,
  () => {
    if (zonePrograms.value.length === 1) {
      form.value.programId = sortedZonePrograms.value[0]?.id ?? zonePrograms.value[0].id
    } else if (!zonePrograms.value.some((p) => p.id === form.value.programId)) {
      form.value.programId = null
    }
  },
)

function onPlantPick() {
  const plant = props.plants.find((p) => Number(p.id) === Number(plantPickId.value))
  if (plant) {
    form.value.strain = strainFromPlant(plant)
    if (plant.crop_profile_id) form.value.cropProfileId = plant.crop_profile_id
    if (plant.crop_key) form.value.cropKey = plant.crop_key
  }
}

function close() {
  emit('close')
}

async function submit() {
  formError.value = ''
  if (!form.value.cropProfileId || !form.value.cropKey) {
    formError.value = 'Choose a plant type from the knowledge base'
    return
  }
  if (!canSubmit.value) return
  submitting.value = true
  try {
    let plantId = plantPickId.value ? Number(plantPickId.value) : null
    if (!plantId) {
      const createdPlant = await store.createPlant(props.farmId, {
        crop_key: form.value.cropKey,
        variety_or_cultivar: form.value.strain?.trim() || null,
      })
      plantId = createdPlant.id
    } else {
      const plant = props.plants.find((p) => Number(p.id) === plantId)
      if (plant?.crop_key && plant.crop_key !== form.value.cropKey) {
        formError.value = 'Selected plant does not match the crop type — pick another plant or clear the selection'
        return
      }
    }
    if (!plantId) {
      formError.value = 'Could not link a catalog plant to this grow'
      return
    }
    const payload = buildStartGrowPayload({
      zoneId: form.value.zoneId,
      strain: form.value.strain,
      name: form.value.name,
      stage: form.value.stage,
      startedAt: form.value.startedAt,
      programId: form.value.programId,
      plantId,
    })
    const created = await store.createCropCycle(props.farmId, payload)
    emit('created', created)
    close()
  } catch (e) {
    formError.value = e?.response?.data?.error || e?.message || 'Failed to start grow'
  } finally {
    submitting.value = false
  }
}
</script>
