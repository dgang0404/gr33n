<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
      <div>
        <h2 class="text-lg font-semibold text-white">Grow economics</h2>
        <p class="text-zinc-500 text-sm mt-1 max-w-2xl">
          Compare harvests side by side or open a grow summary for cost-per-gram context.
        </p>
      </div>
      <router-link
        v-if="compareRoute"
        v-nav-hint="'/zones'"
        :to="compareRoute"
        class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 shrink-0"
        data-test="money-grows-compare"
      >
        Compare harvests →
      </router-link>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading grows…</div>
    <EmptyStateHint
      v-else-if="!cycles.length"
      reason="no_data"
      message="No crop cycles yet — start a grow from Zones → Plants or a zone Overview tab."
      action-label="Plants"
      :action-to="{ path: '/zones', query: { tab: 'plants' } }"
    />
    <div v-else class="space-y-5">
      <section v-for="group in groupedCycles" :key="group.key || 'unlinked'" class="space-y-2">
        <h3 class="text-xs uppercase tracking-wide text-zinc-500">
          {{ group.label }}
          <span class="text-zinc-600">({{ group.cycles.length }})</span>
        </h3>
        <div
          v-for="c in group.cycles"
          :key="c.id"
          class="flex items-center justify-between gap-3 bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3"
          data-test="money-grow-row"
        >
          <div class="min-w-0">
            <p class="text-sm text-zinc-200 truncate">{{ c.name || cycleBatchLabel(c) || `Grow #${c.id}` }}</p>
            <p class="text-[11px] text-zinc-500">
              {{ c.is_active ? 'Active' : 'Harvested' }}
              <span v-if="c.zone_id"> · zone #{{ c.zone_id }}</span>
              <span v-if="c.batch_label"> · {{ c.batch_label }}</span>
            </p>
          </div>
          <div class="flex items-center gap-3 shrink-0">
            <router-link
              :to="`/crop-cycles/${c.id}/summary`"
              class="text-xs text-green-500 hover:text-green-400"
            >
              Summary →
            </router-link>
            <router-link
              v-nav-hint="'/money'"
              :to="{ path: '/money', query: { tab: 'summary', cycle_id: c.id } }"
              class="text-xs text-zinc-500 hover:text-zinc-300"
            >
              Receipts
            </router-link>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { cycleBatchLabel } from '../lib/growHub.js'
import { groupCyclesByCropKey } from '../lib/cropAnalytics.js'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import EmptyStateHint from '../components/EmptyStateHint.vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const cycles = ref([])

const groupedCycles = computed(() => groupCyclesByCropKey(cycles.value))

const compareRoute = computed(() => {
  const fid = farmContext.farmId
  return fid ? { path: `/farms/${fid}/crop-cycles/compare` } : null
})

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) {
    cycles.value = []
    return
  }
  loading.value = true
  try {
    cycles.value = await store.loadCropCycles(fid)
  } finally {
    loading.value = false
  }
}

watch(() => farmContext.farmId, refresh, { immediate: true })
onMounted(refresh)
</script>
