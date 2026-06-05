<template>
  <form
    class="border-t border-zinc-800 pt-3 space-y-3"
    data-test="feeding-plan-editor"
    @submit.prevent="save"
  >
    <div class="flex items-center justify-between gap-2">
      <p class="text-xs font-semibold text-zinc-300">Edit feeding plan</p>
      <button
        type="button"
        class="text-[11px] text-zinc-500 hover:text-zinc-300"
        @click="resetDraft"
      >
        Reset
      </button>
    </div>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <label class="text-[11px] text-zinc-500">
        Volume per feed (L)
        <input
          v-model.number="draft.volumeLiters"
          type="number"
          step="0.1"
          min="0"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="feeding-edit-volume"
        />
      </label>

      <label class="text-[11px] text-zinc-500">
        Daily feed time
        <input
          v-model="draft.dailyFeedTime"
          type="time"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="feeding-edit-time"
          :disabled="!schedule"
        />
      </label>

      <label v-if="!draft.irrigationOnly && !draft.ecFromTarget" class="text-[11px] text-zinc-500">
        EC target min (mS/cm)
        <input
          v-model.number="draft.ecMin"
          type="number"
          step="0.01"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="feeding-edit-ec-min"
        />
      </label>

      <label v-if="!draft.irrigationOnly && !draft.ecFromTarget" class="text-[11px] text-zinc-500">
        EC target max (mS/cm)
        <input
          v-model.number="draft.ecMax"
          type="number"
          step="0.01"
          class="mt-1 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white"
          data-test="feeding-edit-ec-max"
        />
      </label>
    </div>

    <p v-if="draft.ecFromTarget && !draft.irrigationOnly" class="text-[11px] text-zinc-600">
      EC band comes from a linked target — change it under Advanced feeding if needed.
    </p>

    <div class="flex flex-wrap gap-4 text-xs text-zinc-300">
      <label class="flex items-center gap-2">
        <input v-model="draft.irrigationOnly" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
        Water only (no nutrients)
      </label>
      <label v-if="schedule" class="flex items-center gap-2">
        <input v-model="draft.feedingPaused" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
        Pause daily feeding
      </label>
    </div>

    <p v-if="!schedule" class="text-[11px] text-amber-400/90">
      No daily feed time linked — set one up in Advanced feeding or start a new plan below.
    </p>

    <p v-if="error" class="text-xs text-red-400">{{ error }}</p>
    <p v-if="savedMsg" class="text-xs text-green-400">{{ savedMsg }}</p>

    <button
      type="submit"
      class="text-xs px-3 py-1.5 rounded-md bg-green-800 hover:bg-green-700 text-white disabled:opacity-50"
      :disabled="saving"
      data-test="feeding-edit-save"
    >
      {{ saving ? 'Saving…' : 'Save feeding plan' }}
    </button>
  </form>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { buildDailyFeedCron } from '../lib/dailyFeedSchedule.js'
import {
  buildFeedingPlanDraft,
  buildProgramPatch,
  buildSchedulePatch,
} from '../lib/feedingPlanEdit.js'

const props = defineProps({
  activeProgram: { type: Object, required: true },
  schedule: { type: Object, default: null },
  ecTargets: { type: Array, default: () => [] },
})

const emit = defineEmits(['saved'])

const store = useFarmStore()
const draft = ref(buildFeedingPlanDraft({
  activeProgram: props.activeProgram,
  schedule: props.schedule,
  ecTargets: props.ecTargets,
}))
const saving = ref(false)
const error = ref('')
const savedMsg = ref('')

function resetDraft() {
  error.value = ''
  savedMsg.value = ''
  draft.value = buildFeedingPlanDraft({
    activeProgram: props.activeProgram,
    schedule: props.schedule,
    ecTargets: props.ecTargets,
  })
}

watch(
  () => [props.activeProgram, props.schedule, props.ecTargets],
  () => resetDraft(),
  { deep: true },
)

async function save() {
  saving.value = true
  error.value = ''
  savedMsg.value = ''
  try {
    await store.updateProgram(
      props.activeProgram.id,
      buildProgramPatch(props.activeProgram, draft.value),
    )

    if (props.schedule) {
      await store.updateSchedule(
        props.schedule.id,
        buildSchedulePatch(
          props.schedule,
          draft.value,
          buildDailyFeedCron(draft.value.dailyFeedTime),
        ),
      )
    }

    savedMsg.value = 'Feeding plan saved.'
    emit('saved')
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Could not save feeding plan'
  } finally {
    saving.value = false
  }
}
</script>
