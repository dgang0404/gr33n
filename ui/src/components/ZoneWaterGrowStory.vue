<template>
  <div
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-4 scroll-mt-24"
    data-test="zone-water-grow-story"
    id="zone-water-plan"
  >
    <div class="flex items-start justify-between gap-2 flex-wrap">
      <div>
        <h3 class="text-sm font-semibold text-white">How this zone gets water</h3>
        <p class="text-zinc-600 text-xs mt-0.5">Daily feeding plan, last run, and what the Pi is waiting on.</p>
      </div>
      <router-link
        v-nav-hint="advancedFeedingLink"
        :to="advancedFeedingLink"
        class="text-xs text-zinc-500 hover:text-green-400"
        data-test="feeding-advanced-link"
      >
        Advanced feeding →
      </router-link>
    </div>

    <p
      class="text-sm font-medium"
      :class="plan.hasPlan ? 'text-zinc-100' : 'text-zinc-500'"
      data-test="feeding-status-line"
    >
      {{ plan.statusLine }}
    </p>

    <ZoneCropStageTargetHint :zone-id="zoneId" :farm-id="farmId" />

    <div
      v-if="programMismatchSummaryText"
      class="rounded-lg border border-amber-800/60 bg-amber-950/30 px-3 py-2 text-[11px] text-amber-200/90"
      data-test="water-program-mismatch"
    >
      {{ programMismatchSummaryText }}
      <router-link
        v-nav-hint="advancedFeedingLink"
        :to="advancedFeedingLink"
        class="ml-1 text-amber-100/90 underline hover:text-white"
      >
        Edit program →
      </router-link>
    </div>

    <ZoneFeedingPlanWizard
      v-if="!plan.hasPlan"
      :zone-id="zoneId"
      :farm-id="farmId"
      :zone-name="zoneName"
      :reservoirs="reservoirs"
      :farm-timezone="farmTimezone"
      data-test="feeding-plan-wizard-wrap"
      @created="$emit('plan-updated')"
    />

    <template v-else>
      <div
        class="bg-zinc-950 border border-zinc-800 rounded-xl p-4 space-y-3"
        data-test="feeding-plan-card"
      >
        <div class="flex items-start justify-between gap-2 flex-wrap">
          <div>
            <p class="text-sm text-white font-medium">{{ plan.programName }}</p>
            <div class="flex flex-wrap gap-1.5 mt-1">
              <span
                v-if="plan.irrigationOnly"
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-sky-900/50 text-sky-300 font-semibold"
                data-test="feeding-water-only-badge"
              >Water only</span>
              <span
                v-if="plan.schedule && !plan.scheduleActive"
                class="text-[10px] px-1.5 py-0.5 rounded-full bg-zinc-800 text-zinc-400"
              >Feeding paused</span>
            </div>
          </div>
        </div>

        <dl class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div>
            <dt class="text-zinc-500 uppercase tracking-wide text-[10px]">Next feed</dt>
            <dd class="text-zinc-200 mt-0.5">{{ plan.nextRunLabel || '—' }}</dd>
          </div>
          <div>
            <dt class="text-zinc-500 uppercase tracking-wide text-[10px]">Volume</dt>
            <dd class="text-zinc-200 mt-0.5">{{ plan.volumeLiters != null ? `${plan.volumeLiters}L` : '—' }}</dd>
          </div>
          <div>
            <dt class="text-zinc-500 uppercase tracking-wide text-[10px]">EC target</dt>
            <dd class="text-zinc-200 mt-0.5">{{ plan.ecRange?.label || (plan.irrigationOnly ? 'Plain water' : '—') }}</dd>
          </div>
          <div>
            <dt class="text-zinc-500 uppercase tracking-wide text-[10px]">Pump run</dt>
            <dd class="text-zinc-200 mt-0.5">
              {{ plan.runDurationSeconds ? `${plan.runDurationSeconds}s` : '—' }}
            </dd>
          </div>
        </dl>

        <ZoneFeedingPlanEditor
          :active-program="activeProgram"
          :schedule="plan.schedule"
          :ec-targets="ecTargets"
          @saved="$emit('plan-updated')"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="bg-zinc-950 border border-zinc-800 rounded-lg p-3" data-test="grow-story-last" id="zone-feed-history">
          <div class="flex items-center justify-between gap-2 mb-1">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500">Last feed</p>
            <router-link
              v-nav-hint="feedHistoryLink"
              :to="feedHistoryLink"
              class="text-[10px] text-green-600 hover:text-green-400"
              data-test="feeding-history-link"
            >
              See history →
            </router-link>
          </div>
          <p class="text-sm text-zinc-200">{{ plan.lastEventSummary }}</p>
        </div>

        <div
          class="bg-zinc-950 border rounded-lg p-3 flex flex-col justify-center"
          :class="reservoirChipClass"
          data-test="feeding-reservoir-chip"
        >
          <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">Reservoir</p>
          <p class="text-sm">{{ plan.reservoirLabel }}</p>
        </div>
      </div>

      <div
        class="bg-zinc-950 border rounded-lg p-3"
        :class="plan.queueDepth > 0 ? 'border-amber-800/60' : 'border-zinc-800'"
        data-test="grow-story-edge"
      >
        <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-1">Pi queue</p>
        <p class="text-sm" :class="plan.queueDepth > 0 ? 'text-amber-200' : 'text-zinc-200'">
          {{ plan.queueLine }}
        </p>
      </div>

      <ActuatorPulseControl
        v-if="deliveryActuator"
        :actuator="deliveryActuator"
        :default-seconds="plan.runDurationSeconds || 30"
        data-test="feeding-delivery-pulse"
      />

      <p v-if="runNowFeedback" class="text-xs" :class="runNowOk ? 'text-green-400' : 'text-amber-400'">
        {{ runNowFeedback }}
      </p>

      <div class="flex flex-wrap items-center gap-2 border-t border-zinc-800 pt-3">
        <button
          type="button"
          class="text-xs px-3 py-2 min-h-[44px] sm:min-h-0 rounded-md bg-amber-900/70 text-amber-100 hover:bg-amber-800/85 disabled:opacity-50 font-medium"
          :class="focusRingClass"
          :disabled="runNowBusy"
          :aria-label="runNowLabel"
          data-test="grow-story-run-now"
          @click="runActiveProgramNow"
        >
          {{ runNowBusy ? 'Running…' : 'Run feed now' }}
        </button>
        <button
          v-if="plan.mixRequired && !plan.irrigationOnly"
          type="button"
          class="text-xs px-2.5 py-1 rounded-md border border-zinc-700 text-zinc-300 hover:border-green-700 hover:text-green-300 disabled:opacity-50"
          :disabled="mixPreviewLoading"
          data-test="grow-story-preview-mix"
          @click="loadMixPreview"
        >
          {{ mixPreviewLoading ? 'Calculating…' : 'Preview mix' }}
        </button>
        <router-link
          v-nav-hint="'/operations/feeding'"
          :to="logFeedLink"
          class="text-xs px-2.5 py-1 rounded-md border border-zinc-700 text-zinc-400 hover:text-green-300"
          data-test="grow-story-log-feed"
        >
          Log a feed
        </router-link>
      </div>

      <div v-if="waterStatus?.last_mixing_event" class="flex items-center gap-2 text-xs text-zinc-400">
        <span>Last mix:</span>
        <span class="text-zinc-200">{{ formatMixDate(waterStatus.last_mixing_event.mixed_at) }}</span>
        <span
          v-if="waterStatus.last_mixing_event.ec_target_met === true"
          class="px-1.5 py-0.5 rounded bg-green-900 text-green-300 text-[10px] font-semibold"
        >EC met ✓</span>
        <span
          v-else-if="waterStatus.last_mixing_event.ec_target_met === false"
          class="px-1.5 py-0.5 rounded bg-red-900 text-red-300 text-[10px] font-semibold"
        >EC not met</span>
      </div>

      <div v-if="showMixPreview && waterStatus?.mix_preview" class="space-y-1 border-t border-zinc-800 pt-3">
        <div class="flex items-center justify-between">
          <p class="text-xs text-zinc-400 font-semibold">Mix plan — {{ waterStatus.mix_preview.dilution_ratio }}</p>
          <button type="button" class="text-[10px] text-zinc-600 hover:text-zinc-400" @click="showMixPreview = false">
            hide
          </button>
        </div>
        <p class="text-[11px] text-zinc-500">
          {{ waterStatus.mix_preview.water_volume_liters }}L ·
          base {{ waterStatus.mix_preview.water_ec_mscm }} mS/cm →
          est. {{ waterStatus.mix_preview.estimated_final_ec_mscm }} mS/cm
        </p>
        <div
          v-for="step in waterStatus.mix_preview.steps"
          :key="step.step"
          class="flex items-center gap-2 text-[11px] text-zinc-300"
        >
          <span class="w-4 text-zinc-600 text-right">{{ step.step }}.</span>
          <span class="flex-1">{{ step.input_name }}</span>
          <span class="text-zinc-500">{{ step.volume_ml }} ml · {{ step.run_seconds }}s</span>
        </div>
        <p v-if="waterStatus.mix_preview.warnings?.length" class="text-[10px] text-amber-500 mt-1 leading-tight">
          ⚠ {{ waterStatus.mix_preview.warnings[0] }}
        </p>
      </div>
      <p
        v-else-if="waterStatus?.mix_preview_error && !plan.mixRequired"
        class="text-[11px] text-zinc-600 italic"
      >
        {{ waterStatus.mix_preview_error }}
      </p>
    </template>

    <ZoneGrowConnectionLine
      :zone-id="zoneId"
      :farm-id="farmId"
      :active-program="activeProgram"
    />

    <ZoneGrowCostPeek :zone-id="zoneId" :farm-id="farmId" />

    <div class="border-t border-zinc-800 pt-3 flex flex-wrap gap-3">
      <router-link
        v-nav-hint="'/operations/supplies'"
        :to="suppliesForRoomLink"
        class="text-xs text-zinc-400 hover:text-green-400"
        data-test="zone-water-supplies-link"
      >
        Stock &amp; recipes for this zone →
      </router-link>
      <router-link
        v-nav-hint="moneyTabRoute('summary')"
        :to="moneyTabRoute('summary')"
        class="text-xs text-zinc-500 hover:text-green-400"
      >
        Farm money →
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import { useFarmStore } from '../stores/farm.js'
import { buildZoneFeedingPlan } from '../lib/zoneFeedingPlan.js'
import { programMismatchSummary } from '../lib/programFit.js'
import { supportsPulseCommand } from '../lib/plantNeeds.js'
import ActuatorPulseControl from './ActuatorPulseControl.vue'
import ZoneFeedingPlanEditor from './ZoneFeedingPlanEditor.vue'
import ZoneFeedingPlanWizard from './ZoneFeedingPlanWizard.vue'
import ZoneGrowCostPeek from './ZoneGrowCostPeek.vue'
import ZoneGrowConnectionLine from './ZoneGrowConnectionLine.vue'
import ZoneCropStageTargetHint from './ZoneCropStageTargetHint.vue'
import { FARMER_FOCUS_RING, runFeedNowAriaLabel } from '../lib/farmerA11y.js'
import {
  moneyTabRoute,
  zoneTabRoute,
  zoneWaterPlanRoute,
} from '../lib/workspaceRoutes.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  zoneName: { type: String, default: 'This zone' },
  farmTimezone: { type: String, default: 'UTC' },
  activeProgram: { type: Object, default: null },
  programs: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  fertigationEvents: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  ecTargets: { type: Array, default: () => [] },
  reservoirs: { type: Array, default: () => [] },
  /** Active crop cycle in this zone — for program fit warnings (Phase 96). */
  activeCropCycle: { type: Object, default: null },
  /** { cropKey, stage } from active grow */
  growFitContext: { type: Object, default: () => ({ cropKey: '', stage: '' }) },
})

const emit = defineEmits(['refreshed', 'plan-updated'])

const store = useFarmStore()
const focusRingClass = FARMER_FOCUS_RING
const waterStatus = ref(null)
const queueHead = ref(null)
const mixPreviewLoading = ref(false)
const showMixPreview = ref(false)
const runNowBusy = ref(false)
const runNowFeedback = ref('')
const runNowOk = ref(true)
const deliveryDeviceId = ref(null)

const plan = computed(() => buildZoneFeedingPlan({
  zoneId: props.zoneId,
  activeProgram: props.activeProgram,
  programs: props.programs,
  schedules: props.schedules,
  events: props.fertigationEvents,
  ecTargets: props.ecTargets,
  reservoirs: props.reservoirs,
  waterStatus: waterStatus.value,
  queueHead: queueHead.value,
}))

const programForFit = computed(() => {
  if (props.activeProgram) return props.activeProgram
  const pid = props.activeCropCycle?.primary_program_id
  if (!pid) return null
  return props.programs.find((p) => Number(p.id) === Number(pid)) || null
})

const programMismatchSummaryText = computed(() =>
  programMismatchSummary(programForFit.value, props.growFitContext || {}),
)

const runNowLabel = computed(() =>
  runFeedNowAriaLabel(props.zoneName, props.activeProgram?.name || plan.value?.programName),
)

const deliveryActuator = computed(() => {
  const res = plan.value.reservoir
  if (!res?.delivery_actuator_id) return null
  const act = props.actuators.find((a) => a.id === res.delivery_actuator_id)
    || store.actuators.find((a) => a.id === res.delivery_actuator_id)
  if (!act || !supportsPulseCommand(act.actuator_type)) return act
  return act
})

const reservoirChipClass = computed(() => {
  if (plan.value.reservoirTone === 'warn') return 'border-amber-800/60 bg-amber-950/20'
  if (plan.value.reservoirTone === 'ok') return 'border-green-900/50 bg-green-950/15'
  return 'border-zinc-800'
})

const feedHistoryLink = computed(() =>
  zoneTabRoute(props.zoneId, 'water', '#zone-feed-history'),
)

const logFeedLink = computed(() => zoneWaterPlanRoute(props.zoneId))

const advancedFeedingLink = computed(() => zoneWaterPlanRoute(props.zoneId))

const suppliesForRoomLink = computed(() =>
  moneyTabRoute('supplies', { zoneId: props.zoneId }),
)

async function resolveDeliveryDevice() {
  deliveryDeviceId.value = null
  const res = plan.value.reservoir
  if (!res?.delivery_actuator_id) return
  const act = props.actuators.find((a) => a.id === res.delivery_actuator_id)
    || store.actuators.find((a) => a.id === res.delivery_actuator_id)
  deliveryDeviceId.value = act?.device_id || null
}

async function loadQueueHead() {
  queueHead.value = null
  if (!deliveryDeviceId.value) return
  try {
    const r = await api.get(`/devices/${deliveryDeviceId.value}/commands`, {
      params: { status: 'pending' },
    })
    const list = r.data?.commands ?? r.data ?? []
    queueHead.value = Array.isArray(list) && list.length ? list[0] : null
  } catch {
    // non-fatal
  }
}

async function loadWaterStatus() {
  if (!props.activeProgram?.id) {
    waterStatus.value = null
    queueHead.value = null
    return
  }
  try {
    const r = await api.get(`/fertigation/programs/${props.activeProgram.id}/water-status`)
    waterStatus.value = r.data
  } catch {
    waterStatus.value = null
  }
  await loadQueueHead()
  emit('refreshed', waterStatus.value)
}

async function loadMixPreview() {
  mixPreviewLoading.value = true
  await loadWaterStatus()
  mixPreviewLoading.value = false
  showMixPreview.value = true
}

async function runActiveProgramNow() {
  if (!props.farmId || !props.activeProgram?.id) return
  runNowBusy.value = true
  runNowFeedback.value = ''
  try {
    const res = await store.runFertigationProgramNow(props.farmId, props.activeProgram.id)
    runNowOk.value = true
    runNowFeedback.value = res.duplicate
      ? 'Already ran this minute — no duplicate commands queued.'
      : (res.message || 'Feed queued on the Pi.')
    await loadWaterStatus()
  } catch (e) {
    runNowOk.value = false
    runNowFeedback.value = e?.response?.data?.error || e?.message || 'Run now failed'
  } finally {
    runNowBusy.value = false
  }
}

function formatMixDate(iso) {
  if (!iso) return '—'
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  } catch {
    return iso
  }
}

watch(
  () => [props.activeProgram?.id, props.activeProgram?.reservoir_id, props.reservoirs],
  async () => {
    await resolveDeliveryDevice()
    await loadWaterStatus()
  },
  { immediate: true, deep: true },
)
</script>
