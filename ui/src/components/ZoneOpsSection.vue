<template>
  <div class="space-y-4" data-test="zone-ops-section">
    <div class="flex flex-wrap gap-2 border-b border-zinc-800 pb-2">
      <button
        v-for="sub in opsSubTabs"
        :key="sub.id"
        type="button"
        class="px-3 py-1.5 text-xs rounded-lg border transition-colors"
        :class="opsView === sub.id
          ? 'border-green-700/60 bg-green-950/40 text-green-400'
          : 'border-zinc-800 text-zinc-500 hover:text-zinc-300'"
        :data-test="`zone-ops-sub-${sub.id}`"
        @click="selectOps(sub.id)"
      >
        {{ sub.icon }} {{ sub.label }}
      </button>
    </div>

    <Alerts
      v-if="opsView === 'alerts'"
      embedded
      :lock-zone-id="zoneId"
    />
    <Tasks
      v-else
      embedded
      :lock-zone-id="zoneId"
      :auto-open-create="autoOpenCreate"
    />
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Alerts from '../views/Alerts.vue'
import Tasks from '../views/Tasks.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
})

const route = useRoute()
const router = useRouter()

const opsSubTabs = [
  { id: 'alerts', icon: '🔔', label: 'Alerts' },
  { id: 'tasks', icon: '✅', label: 'Tasks' },
]

const opsView = computed(() => {
  const raw = route.query.ops
  const id = typeof raw === 'string' ? raw : 'alerts'
  return opsSubTabs.some((t) => t.id === id) ? id : 'alerts'
})

const autoOpenCreate = computed(() => route.query.create === '1' || route.query.create === 'true')

function selectOps(id) {
  router.replace({
    path: route.path,
    query: { ...route.query, tab: 'ops', ops: id },
  })
}

watch(
  () => route.query.tab,
  (tab) => {
    if (tab === 'ops' && !route.query.ops) {
      router.replace({
        path: route.path,
        query: { ...route.query, tab: 'ops', ops: 'alerts' },
      })
    }
  },
  { immediate: true },
)
</script>
