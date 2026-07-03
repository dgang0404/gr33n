<template>
  <div v-if="expanded || commands.length" class="rounded-lg border border-zinc-800/80 bg-zinc-950/30 px-3 py-2 space-y-2" data-test="device-command-queue">
    <button
      type="button"
      class="flex items-center justify-between w-full text-left text-[10px] uppercase tracking-wide text-zinc-500 hover:text-zinc-300"
      @click="expanded = !expanded"
    >
      <span>Command queue ({{ commands.length }})</span>
      <span>{{ expanded ? '▾' : '▸' }}</span>
    </button>
    <div v-if="expanded" class="space-y-1 max-h-40 overflow-y-auto">
      <p v-if="loading" class="text-[10px] text-zinc-600">Loading…</p>
      <p v-else-if="!commands.length" class="text-[10px] text-zinc-600">No recent commands.</p>
      <div
        v-for="cmd in commands"
        :key="cmd.id"
        class="flex items-start justify-between gap-2 text-[10px] border-b border-zinc-900 pb-1"
        data-test="device-command-row"
      >
        <div class="min-w-0">
          <span class="text-zinc-300 font-mono">#{{ cmd.id }}</span>
          <span class="text-zinc-600 ml-1">{{ cmd.command_type }}</span>
          <span class="ml-1 capitalize" :class="statusClass(cmd.status)">{{ cmd.status }}</span>
          <div class="text-zinc-600 truncate">{{ ageLabel(cmd.created_at) }}</div>
        </div>
        <button
          v-if="cmd.status === 'pending'"
          type="button"
          class="shrink-0 text-red-400 hover:text-red-300"
          data-test="cancel-command"
          @click="cancel(cmd.id)"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm'

const props = defineProps({
  deviceId: { type: Number, required: true },
})

const store = useFarmStore()
const commands = ref([])
const loading = ref(false)
const expanded = ref(false)

async function refresh() {
  if (!props.deviceId) return
  loading.value = true
  try {
    const data = await store.listDeviceCommands(props.deviceId)
    commands.value = Array.isArray(data?.commands) ? data.commands : []
    if (commands.value.some(c => c.status === 'pending' || c.status === 'in_progress')) {
      expanded.value = true
    }
  } catch {
    commands.value = []
  } finally {
    loading.value = false
  }
}

async function cancel(commandId) {
  await store.cancelDeviceCommand(props.deviceId, commandId)
  await refresh()
}

function ageLabel(iso) {
  if (!iso) return ''
  const mins = Math.round((Date.now() - new Date(iso).getTime()) / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  return `${Math.round(mins / 60)}h ago`
}

function statusClass(status) {
  if (status === 'completed') return 'text-emerald-500'
  if (status === 'failed') return 'text-red-400'
  if (status === 'pending' || status === 'in_progress') return 'text-amber-400'
  return 'text-zinc-500'
}

onMounted(refresh)
watch(() => props.deviceId, refresh)
</script>
