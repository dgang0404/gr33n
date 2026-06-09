<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-tasks-panel">
    <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
      <div class="flex items-center gap-2">
        <h2 class="text-sm font-semibold text-white">Due today in this zone</h2>
        <span
          v-if="dueToday.length"
          class="text-[10px] font-semibold px-2 py-0.5 rounded-full bg-blue-900/50 text-blue-200"
        >
          {{ dueToday.length }}
        </span>
      </div>
      <router-link v-nav-hint="'/zones'" :to="tasksLink" class="text-xs text-green-600 hover:text-green-400">
        See all in Ops →
      </router-link>
    </div>

    <p v-if="!dueToday.length" class="text-zinc-500 text-sm">
      <EmptyStateHint reason="no_data" message="No tasks due today for this zone." compact />
    </p>

    <ul v-else class="space-y-2">
      <li
        v-for="task in dueToday"
        :key="task.id"
        class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 flex flex-col sm:flex-row sm:items-center gap-2 sm:justify-between"
        :data-test="`zone-task-${task.id}`"
      >
        <div class="min-w-0 flex-1">
          <p class="text-sm text-zinc-200 font-medium truncate">{{ task.title }}</p>
          <p class="text-[11px] mt-0.5" :class="dueClass(task.due_date)">
            {{ formatTaskDueLabel(task.due_date) }}
            <span v-if="task.task_type" class="text-zinc-600"> · {{ task.task_type }}</span>
          </p>
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            type="button"
            class="text-xs px-2 py-1 rounded border border-green-800 text-green-400 hover:bg-green-900/30 disabled:opacity-50"
            :disabled="busyId === task.id"
            :data-test="`zone-task-complete-${task.id}`"
            @click="openComplete(task)"
          >
            Done
          </button>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200 disabled:opacity-50"
            :disabled="busyId === task.id"
            :data-test="`zone-task-snooze-${task.id}`"
            @click="snoozeTask(task)"
          >
            Snooze
          </button>
        </div>
      </li>
    </ul>

    <TaskCompleteSheet
      :open="!!completeTarget"
      :task="completeTarget"
      :batches="nfBatches"
      :inputs="nfInputs"
      @cancel="completeTarget = null"
      @complete="onCompleteSheet"
    />
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmContextStore } from '../stores/farmContext.js'
import { useFarmStore } from '../stores/farm.js'
import {
  zoneTasksDueToday,
  formatTaskDueLabel,
  snoozeDueDateToTomorrow,
  todayDateIso,
} from '../lib/zoneTasks.js'
import EmptyStateHint from './EmptyStateHint.vue'
import TaskCompleteSheet from './TaskCompleteSheet.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  tasks: { type: Array, default: () => [] },
  limit: { type: Number, default: 5 },
})

const emit = defineEmits(['refresh'])

const store = useFarmStore()
const farmContext = useFarmContextStore()
const busyId = ref(null)
const completeTarget = ref(null)
const nfBatches = ref([])
const nfInputs = ref([])

const dueToday = computed(() => zoneTasksDueToday(props.tasks, props.zoneId, props.limit))

const tasksLink = computed(() => ({
  path: `/zones/${props.zoneId}`,
  query: { tab: 'ops', ops: 'tasks' },
}))

function dueClass(dueDate) {
  const d = String(dueDate || '').slice(0, 10)
  if (d && d < todayDateIso()) return 'text-red-400'
  if (d === todayDateIso()) return 'text-amber-300/90'
  return 'text-zinc-500'
}

async function ensureSuppliesLoaded() {
  const fid = farmContext.farmId
  if (!fid) return
  if (!nfBatches.value.length) nfBatches.value = await store.loadNfBatches(fid)
  if (!nfInputs.value.length) nfInputs.value = await store.loadNfInputs(fid)
}

async function openComplete(task) {
  await ensureSuppliesLoaded()
  completeTarget.value = task
}

async function onCompleteSheet({ task, consumption }) {
  if (!task) return
  busyId.value = task.id
  try {
    await store.updateTaskStatus(task.id, 'completed')
    if (consumption) {
      await store.recordTaskConsumption(task.id, consumption)
      const fid = farmContext.farmId
      if (fid) {
        nfBatches.value = await store.loadNfBatches(fid)
        await store.loadFarmTaskConsumptions(fid)
      }
    }
    completeTarget.value = null
    emit('refresh')
  } finally {
    busyId.value = null
  }
}

async function snoozeTask(task) {
  busyId.value = task.id
  try {
    await store.updateTask(task.id, {
      due_date: snoozeDueDateToTomorrow(task.due_date),
    })
    emit('refresh')
  } finally {
    busyId.value = null
  }
}
</script>
