<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Schedules</h1>
      <span class="text-xs text-zinc-500">{{ tasks.length }} tasks</span>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading tasks…</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-4 items-start">
      <div
        v-for="col in COLUMNS"
        :key="col.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 flex flex-col gap-3"
      >
        <div class="flex items-center justify-between pb-1 border-b border-zinc-800">
          <span class="flex items-center gap-2 font-medium text-white">
            <span>{{ col.icon }}</span>{{ col.label }}
          </span>
          <span class="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded-full">
            {{ colTasks(col).length }}
          </span>
        </div>

        <p v-if="!colTasks(col).length" class="text-zinc-700 text-sm py-4 text-center">
          No tasks
        </p>

        <div
          v-for="task in colTasks(col)"
          :key="task.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg p-3 hover:border-zinc-700 transition-colors"
        >
          <div class="flex items-start justify-between gap-2 mb-1.5">
            <p class="text-white text-sm font-medium leading-snug">{{ task.title }}</p>
            <span :class="priorityBadge(task.priority)"
              class="shrink-0 text-xs px-1.5 py-0.5 rounded">
              {{ PRIORITY_LABELS[task.priority] ?? 'normal' }}
            </span>
          </div>
          <p v-if="task.description" class="text-zinc-500 text-xs line-clamp-2 mb-2">
            {{ task.description }}
          </p>
          <div class="flex items-center justify-between text-xs text-zinc-600 mb-2">
            <span>{{ zoneName(task.zone_id) }}</span>
            <span v-if="task.due_date">Due {{ task.due_date }}</span>
          </div>
          <button
            v-if="col.next"
            @click="advance(task, col.next)"
            class="text-xs text-green-600 hover:text-green-400 transition-colors"
          >
            → {{ col.nextLabel }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'

const store = useFarmStore()
const tasks = ref([])
const loading = ref(false)

onMounted(async () => {
  if (!store.zones.length) await store.loadAll()
  loading.value = true
  try { tasks.value = await store.loadTasks() }
  finally { loading.value = false }
})

const COLUMNS = [
  { id: 'scheduled', label: 'Scheduled', icon: '📋',
    statuses: ['todo', 'on_hold', 'blocked_requires_input', 'pending_review'],
    next: 'in_progress', nextLabel: 'Start' },
  { id: 'running', label: 'Running', icon: '⚡',
    statuses: ['in_progress'],
    next: 'completed', nextLabel: 'Mark Done' },
  { id: 'done', label: 'Done', icon: '✅',
    statuses: ['completed', 'cancelled'],
    next: null, nextLabel: null },
]
function colTasks(col) { return tasks.value.filter(t => col.statuses.includes(t.status)) }
async function advance(task, nextStatus) {
  await store.updateTaskStatus(task.id, nextStatus)
  task.status = nextStatus
}
function zoneName(id) {
  if (!id) return ''
  return store.zones.find(z => z.id === id)?.name ?? `Zone ${id}`
}
const PRIORITY_LABELS = { 0: 'low', 1: 'normal', 2: 'high', 3: 'urgent' }
const PRIORITY_BADGE = {
  0: 'bg-zinc-800 text-zinc-400', 1: 'bg-blue-900/50 text-blue-400',
  2: 'bg-yellow-900/50 text-yellow-400', 3: 'bg-red-900/50 text-red-400',
}
function priorityBadge(p) { return PRIORITY_BADGE[p] ?? PRIORITY_BADGE[1] }
</script>
