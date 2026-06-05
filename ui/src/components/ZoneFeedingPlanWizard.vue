<template>
  <div
    class="bg-zinc-950 border border-zinc-800 rounded-xl p-4 space-y-4"
    data-test="feeding-plan-wizard"
  >
    <div>
      <h4 class="text-sm font-semibold text-white">Start feeding plan</h4>
      <p class="text-zinc-600 text-xs mt-0.5">Three quick steps — no technical setup screens.</p>
    </div>

    <div class="flex gap-2 text-[10px] uppercase tracking-wide text-zinc-500">
      <span :class="step === 1 ? 'text-green-400' : ''">1 Name</span>
      <span>›</span>
      <span :class="step === 2 ? 'text-green-400' : ''">2 Volume</span>
      <span>›</span>
      <span :class="step === 3 ? 'text-green-400' : ''">3 Daily time</span>
    </div>

    <div v-if="step === 1" class="space-y-2">
      <label class="block text-[11px] text-zinc-500">
        Plan name
        <input
          v-model="form.name"
          type="text"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="wizard-name"
        />
      </label>
    </div>

    <div v-else-if="step === 2" class="space-y-3">
      <label class="block text-[11px] text-zinc-500">
        Volume per feed (L)
        <input
          v-model.number="form.volumeLiters"
          type="number"
          step="0.1"
          min="0.1"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="wizard-volume"
        />
      </label>
      <label class="flex items-center gap-2 text-xs text-zinc-300">
        <input v-model="form.irrigationOnly" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
        Water only — plain irrigation, no nutrients
      </label>
      <div v-if="!form.irrigationOnly" class="grid grid-cols-2 gap-2">
        <label class="text-[11px] text-zinc-500">
          EC min
          <input v-model.number="form.ecMin" type="number" step="0.01" class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-white" />
        </label>
        <label class="text-[11px] text-zinc-500">
          EC max
          <input v-model.number="form.ecMax" type="number" step="0.01" class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-sm text-white" />
        </label>
      </div>
      <p v-if="!pickedReservoir" class="text-[11px] text-amber-400/90">
        No reservoir on this farm yet — link one in Advanced feeding for automated runs.
      </p>
    </div>

    <div v-else class="space-y-2">
      <label class="block text-[11px] text-zinc-500">
        What time should this room feed each day?
        <input
          v-model="form.dailyFeedTime"
          type="time"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="wizard-time"
        />
      </label>
      <p class="text-[11px] text-zinc-600">
        We store the schedule in the background — you only pick a daily time.
      </p>
    </div>

    <p v-if="error" class="text-xs text-red-400">{{ error }}</p>

    <div class="flex flex-wrap gap-2">
      <button
        v-if="step > 1"
        type="button"
        class="text-xs px-3 py-1.5 rounded-md border border-zinc-700 text-zinc-400 hover:text-zinc-200"
        @click="step -= 1"
      >
        Back
      </button>
      <button
        v-if="step < 3"
        type="button"
        class="text-xs px-3 py-1.5 rounded-md bg-zinc-800 text-zinc-200 hover:bg-zinc-700"
        @click="nextStep"
      >
        Next
      </button>
      <button
        v-else
        type="button"
        class="text-xs px-3 py-1.5 rounded-md bg-green-800 hover:bg-green-700 text-white disabled:opacity-50"
        :disabled="creating"
        data-test="wizard-create"
        @click="createPlan"
      >
        {{ creating ? 'Creating…' : 'Create feeding plan' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import {
  buildWizardEcTargetPayload,
  buildWizardProgramPayload,
  buildWizardSchedulePayload,
  pickReservoirForZone,
} from '../lib/feedingPlanEdit.js'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  zoneName: { type: String, default: 'This room' },
  reservoirs: { type: Array, default: () => [] },
  farmTimezone: { type: String, default: 'UTC' },
})

const emit = defineEmits(['created'])

const store = useFarmStore()
const step = ref(1)
const creating = ref(false)
const error = ref('')

const form = ref({
  name: `${props.zoneName} feeding`,
  volumeLiters: 0.3,
  irrigationOnly: false,
  ecMin: 1.1,
  ecMax: 1.3,
  dailyFeedTime: '06:00',
})

const pickedReservoir = computed(() => pickReservoirForZone(props.reservoirs, props.zoneId))

function nextStep() {
  error.value = ''
  if (step.value === 1 && !String(form.value.name || '').trim()) {
    error.value = 'Enter a plan name.'
    return
  }
  if (step.value === 2 && (!form.value.volumeLiters || form.value.volumeLiters <= 0)) {
    error.value = 'Enter a volume greater than zero.'
    return
  }
  step.value += 1
}

async function createPlan() {
  creating.value = true
  error.value = ''
  try {
    const schedule = await store.createSchedule(
      props.farmId,
      buildWizardSchedulePayload({
        zoneName: props.zoneName,
        farmTimezone: props.farmTimezone,
        dailyFeedTime: form.value.dailyFeedTime,
      }),
    )

    let ecTargetId = null
    if (!form.value.irrigationOnly && form.value.ecMin != null && form.value.ecMax != null) {
      const ecTarget = await store.createEcTarget(
        props.farmId,
        buildWizardEcTargetPayload({
          zoneId: props.zoneId,
          ecMin: form.value.ecMin,
          ecMax: form.value.ecMax,
        }),
      )
      ecTargetId = ecTarget.id
    }

    await store.createProgram(
      props.farmId,
      buildWizardProgramPayload({
        name: form.value.name.trim(),
        zoneId: props.zoneId,
        scheduleId: schedule.id,
        reservoirId: pickedReservoir.value?.id,
        ecTargetId,
        volumeLiters: form.value.volumeLiters,
        irrigationOnly: form.value.irrigationOnly,
      }),
    )

    emit('created')
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Could not create feeding plan'
  } finally {
    creating.value = false
  }
}
</script>
