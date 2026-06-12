<template>
  <div class="p-6 max-w-2xl mx-auto space-y-6" data-test="zone-setup-wizard">
    <div>
      <h1 class="text-xl font-semibold text-white">Add a zone</h1>
      <p class="text-zinc-500 text-sm mt-1">
        Name the zone, pick its type, and optionally set greenhouse climate or a lighting photoperiod.
      </p>
    </div>

    <div class="flex gap-2 text-[10px] uppercase tracking-wide text-zinc-500" aria-label="Wizard steps">
      <span :class="step === 'basics' ? 'text-green-400' : ''" :aria-current="step === 'basics' ? 'step' : undefined">1 Basics</span>
      <span aria-hidden="true">›</span>
      <span :class="step === 'needs' ? 'text-green-400' : ''" :aria-current="step === 'needs' ? 'step' : undefined">2 Needs</span>
      <span aria-hidden="true">›</span>
      <span :class="step === 'extras' ? 'text-green-400' : ''" :aria-current="step === 'extras' ? 'step' : undefined">3 Extras</span>
      <span aria-hidden="true">›</span>
      <span :class="step === 'done' ? 'text-green-400' : ''" :aria-current="step === 'done' ? 'step' : undefined">4 Done</span>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>

    <!-- Step 1 — Basics -->
    <template v-if="step === 'basics'">
      <form class="space-y-4" @submit.prevent="goNeeds">
        <label class="block">
          <span class="text-zinc-400 text-xs">Zone name</span>
          <input
            v-model="form.name"
            type="text"
            required
            placeholder="e.g. Flower Room"
            class="input-field mt-1 w-full"
            data-test="zone-wizard-name"
          />
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Description (optional)</span>
          <input
            v-model="form.description"
            type="text"
            placeholder="Bench A, 4×8 tent…"
            class="input-field mt-1 w-full"
          />
        </label>
        <fieldset class="space-y-2">
          <legend class="text-zinc-400 text-xs mb-2">Zone type</legend>
          <label
            v-for="t in zoneTypes"
            :key="t.value"
            class="flex items-start gap-3 rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 cursor-pointer"
            :class="form.zoneType === t.value ? 'border-green-700' : ''"
          >
            <input v-model="form.zoneType" type="radio" :value="t.value" class="mt-1" />
            <span>
              <span class="text-sm text-zinc-200">{{ t.label }}</span>
              <span class="block text-[11px] text-zinc-500">{{ t.hint }}</span>
            </span>
          </label>
        </fieldset>
        <div class="flex flex-wrap gap-2">
          <button
            type="submit"
            class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
            data-test="zone-wizard-continue"
          >
            Continue
          </button>
          <router-link to="/zones" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200">
            Cancel
          </router-link>
        </div>
      </form>
    </template>

    <!-- Step 2 — Greenhouse profile (or skip) -->
    <template v-else-if="step === 'needs'">
      <section v-if="isGreenhouse" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Greenhouse climate profile</h2>
        <p class="text-xs text-zinc-500">
          Shade and ventilation are separate from supplemental lighting — tune actuators on the zone Climate tab later.
        </p>
        <label class="block">
          <span class="text-zinc-400 text-xs">Cover type</span>
          <select v-model="form.coverType" class="input-field mt-1 w-full">
            <option value="">—</option>
            <option v-for="c in coverTypes" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
        </label>
        <fieldset class="space-y-2">
          <legend class="text-zinc-400 text-xs">Automation policy</legend>
          <label
            v-for="p in automationPolicies"
            :key="p.value"
            class="flex items-start gap-2 text-sm text-zinc-300"
          >
            <input v-model="form.automationPolicy" type="radio" :value="p.value" class="mt-1" />
            <span>
              {{ p.label }}
              <span class="block text-[11px] text-zinc-500">{{ p.hint }}</span>
            </span>
          </label>
        </fieldset>
        <label class="block">
          <span class="text-zinc-400 text-xs">Notes (optional)</span>
          <textarea
            v-model="form.greenhouseNotes"
            rows="2"
            class="input-field mt-1 w-full"
            placeholder="Glazing, orientation, shade brand…"
          />
        </label>
      </section>
      <p v-else class="text-sm text-zinc-400">
        No extra climate profile for <strong class="text-zinc-300">{{ form.zoneType }}</strong> zones — continue to optional lighting.
      </p>
      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
          @click="goExtras"
        >
          Continue
        </button>
        <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'basics'">
          Back
        </button>
      </div>
    </template>

    <!-- Step 3 — Lighting + Pi note -->
    <template v-else-if="step === 'extras'">
      <section v-if="showLighting" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Lighting photoperiod (optional)</h2>
        <p class="text-xs text-zinc-500">Creates a lighting program from a preset — needs a grow light actuator on the farm.</p>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="p in lightingPresets"
            :key="p.key || 'skip'"
            type="button"
            class="text-xs px-3 py-1.5 rounded-full border transition-colors"
            :class="form.lightingPreset === p.key
              ? 'border-green-600 bg-green-950/40 text-green-300'
              : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
            @click="form.lightingPreset = p.key"
          >
            {{ p.label }}
          </button>
        </div>
        <template v-if="form.lightingPreset">
          <label class="block">
            <span class="text-zinc-400 text-xs">Grow light actuator</span>
            <select v-model.number="form.actuatorId" class="input-field mt-1 w-full" data-test="zone-wizard-actuator">
              <option :value="null">— Select —</option>
              <option v-for="a in lightActuators" :key="a.id" :value="a.id">
                {{ a.name }} (zone {{ a.zone_id || 'unassigned' }})
              </option>
            </select>
          </label>
          <label class="block">
            <span class="text-zinc-400 text-xs">Lights ON at</span>
            <input v-model="form.lightsOnAt" type="time" class="input-field mt-1 w-full" />
          </label>
          <p v-if="!lightActuators.length" class="text-xs text-amber-300/90">
            No grow light actuators yet — skip lighting or add actuators after connecting a Pi (device wizard, WS3).
          </p>
        </template>
      </section>

      <section class="bg-zinc-900/60 border border-zinc-800 rounded-xl p-4 space-y-2">
        <h2 class="text-sm font-semibold text-zinc-300">Edge device</h2>
        <p class="text-xs text-zinc-500">
          {{ unassignedDevices.length
            ? `${unassignedDevices.length} farm device(s) are not assigned to a zone yet.`
            : 'Connect a Pi when you are ready to wire sensors and pumps.' }}
        </p>
        <router-link
          v-if="farmId"
          :to="deviceWizardLink"
          class="inline-block text-xs text-green-400 hover:text-green-300 underline"
          data-test="zone-wizard-device-link"
        >
          Open edge device wizard →
        </router-link>
      </section>

      <p v-if="submitError" class="text-sm text-red-400">{{ submitError }}</p>

      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          :disabled="saving"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40"
          data-test="zone-wizard-create"
          @click="createZone"
        >
          {{ saving ? 'Creating…' : 'Create zone' }}
        </button>
        <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'needs'">
          Back
        </button>
      </div>
    </template>

    <!-- Step 4 — Done -->
    <template v-else>
      <section class="bg-zinc-900 border border-green-900/50 rounded-xl p-4 space-y-2">
        <p class="text-sm text-green-300 font-medium">{{ doneMessage }}</p>
        <p v-if="lightingNote" class="text-xs text-zinc-500">{{ lightingNote }}</p>
      </section>
      <div class="flex flex-wrap gap-2">
        <router-link
          v-if="createdZoneId"
          :to="`/zones/${createdZoneId}`"
          class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white"
          data-test="zone-wizard-open-room"
        >
          Open {{ form.name }}
        </router-link>
        <router-link to="/zones" class="px-4 py-2 text-sm text-zinc-300 border border-zinc-700 rounded-lg">
          All zones
        </router-link>
      </div>
      <section class="pt-4 border-t border-zinc-800 space-y-2" data-test="zone-wizard-guardian-help">
        <p class="text-[10px] uppercase tracking-widest text-zinc-500">Need help?</p>
        <GuardianStarterChips :starters="zoneWizardStarters" />
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import {
  buildZoneCreatePayload,
  buildLightingPresetRequest,
  filterLightActuators,
  listUnassignedDevices,
  isGreenhouseZoneType,
  supportsLightingPreset,
  zoneSetupTypeOptions,
  zoneSetupCoverTypes,
  zoneSetupAutomationPolicies,
} from '../lib/zoneSetupWizard.js'
import { loadDomainEnums } from '../lib/domainEnums.js'
import { LIGHTING_PRESET_SKIP, loadLightingPresets } from '../lib/lightingPresets.js'
import { deviceSetupRoute } from '../lib/deviceSetupWizard.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const step = ref('basics')
const saving = ref(false)
const submitError = ref('')
const loadError = ref('')
const doneMessage = ref('')
const lightingNote = ref('')
const createdZoneId = ref(null)

const form = reactive({
  name: '',
  description: '',
  zoneType: 'indoor',
  coverType: '',
  automationPolicy: 'manual',
  greenhouseNotes: '',
  lightingPreset: '',
  actuatorId: null,
  lightsOnAt: '06:00',
})

const farmId = computed(() => {
  const raw = route.params.id
  const n = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

const zoneTypes = computed(() => zoneSetupTypeOptions())
const coverTypes = computed(() => zoneSetupCoverTypes())
const automationPolicies = computed(() => zoneSetupAutomationPolicies())
const apiLightingPresets = ref([])
const lightingPresets = computed(() => [LIGHTING_PRESET_SKIP, ...apiLightingPresets.value])

const isGreenhouse = computed(() => isGreenhouseZoneType(form.zoneType))
const showLighting = computed(() => supportsLightingPreset(form.zoneType))
const lightActuators = computed(() => filterLightActuators(store.actuators))
const unassignedDevices = computed(() => listUnassignedDevices(store.devices))

const deviceWizardLink = computed(() => {
  if (!farmId.value) return '/settings'
  return deviceSetupRoute(farmId.value, createdZoneId.value)
})

const farmTimezone = computed(() =>
  farmContext.selectedFarm?.timezone || store.farm?.timezone || 'America/New_York',
)

const zoneWizardStarters = computed(() => buildSetupStarters({
  surface: 'zone_wizard',
  farmId: farmId.value,
  zoneCount: store.zones?.length ?? 0,
  zones: store.zones || [],
  zoneName: form.name,
}))

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
  return true
}

function goNeeds() {
  if (!form.name.trim()) return
  step.value = 'needs'
}

function goExtras() {
  step.value = 'extras'
}

async function createZone() {
  if (!farmId.value) return
  saving.value = true
  submitError.value = ''
  lightingNote.value = ''
  try {
    const payload = buildZoneCreatePayload(form)
    const zone = await store.createZone(farmId.value, payload)
    createdZoneId.value = zone.id
    doneMessage.value = `“${zone.name}” is ready.`

    const lightingReq = buildLightingPresetRequest({
      farmId: farmId.value,
      zoneId: zone.id,
      zoneName: zone.name,
      presetKey: form.lightingPreset,
      actuatorId: form.actuatorId,
      lightsOnAt: form.lightsOnAt,
      timezone: farmTimezone.value,
      presets: apiLightingPresets.value,
    })
    if (lightingReq && form.actuatorId) {
      try {
        await api.post(lightingReq.url, lightingReq.body)
        lightingNote.value = 'Lighting program created from preset — review on the zone Light tab or /lighting.'
      } catch (e) {
        lightingNote.value = `Room created, but lighting preset failed: ${e.response?.data?.error || e.message}`
      }
    } else if (form.lightingPreset && !form.actuatorId) {
      lightingNote.value = 'Skipped lighting preset — pick a grow light actuator or add one after Pi setup.'
    }

    step.value = 'done'
  } catch (e) {
    submitError.value = e.response?.data?.error || e.message || 'Could not create room'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadDomainEnums(api)
  void ensureFarmContext()
  void loadLightingPresets(api).then((rows) => {
    apiLightingPresets.value = rows ?? []
  })
})

watch(farmId, () => {
  void ensureFarmContext()
})
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
