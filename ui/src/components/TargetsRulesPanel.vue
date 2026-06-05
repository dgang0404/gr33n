<template>
  <div class="space-y-4" data-test="targets-rules-panel">
    <ZoneContextBanner
      v-if="zoneContextId"
      :zone-id="zoneContextId"
      :zone-name="zoneName(zoneContextId)"
      page-label="Automation"
      :clear-route="{ path: '/comfort-targets', query: { tab: 'rules' } }"
    />

    <div v-if="loading" class="text-zinc-400 text-sm">Loading automation rules…</div>

    <EmptyStateHint
      v-else-if="!displayRules.length"
      reason="automation_off"
      message="No automation rules yet — apply a greenhouse template or create rules in Advanced."
      action-label="Advanced automations"
      action-to="/automation"
    />

    <div v-else class="space-y-3">
      <article
        v-for="rule in displayRules"
        :key="rule.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
        :data-test="`farmer-rule-${rule.id}`"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-sm text-white font-medium">{{ rule.name }}</p>
            <p class="text-xs text-zinc-400 mt-1">{{ summaryFor(rule) }}</p>
          </div>
          <button
            type="button"
            class="text-xs px-2 py-1 rounded border shrink-0"
            :class="rule.is_active ? 'border-green-700 text-green-400' : 'border-zinc-700 text-zinc-400'"
            :data-test="`toggle-rule-${rule.id}`"
            @click="toggleRule(rule)"
          >
            {{ rule.is_active ? 'On' : 'Paused' }}
          </button>
        </div>
      </article>
    </div>

    <div class="flex flex-wrap gap-3 text-xs">
      <router-link
        :to="{ path: '/zones', query: zoneContextId ? undefined : {} }"
        class="text-zinc-500 hover:text-green-400"
      >
        Greenhouse templates on zone Climate tab →
      </router-link>
      <router-link to="/automation" class="text-green-600 hover:text-green-400">
        Advanced automations →
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneContextBanner from './ZoneContextBanner.vue'
import EmptyStateHint from './EmptyStateHint.vue'
import { filterRulesForZone } from '../lib/zoneContext.js'
import { ruleSummary } from '../lib/ruleSummary.js'

const props = defineProps({
  zoneContextId: { type: Number, default: null },
})

const emit = defineEmits(['refresh'])

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const rules = ref([])
const ruleActions = ref({})

const displayRules = computed(() => {
  if (!props.zoneContextId) return rules.value
  const zone = store.zones.find((z) => z.id === props.zoneContextId)
  return filterRulesForZone(
    rules.value,
    props.zoneContextId,
    zone?.name || '',
    store.sensors,
  )
})

function zoneName(zoneId) {
  return store.zones.find((z) => z.id === zoneId)?.name || `Zone ${zoneId}`
}

function summaryFor(rule) {
  return ruleSummary(rule, {
    sensors: store.sensors,
    actuators: store.actuators,
    actions: ruleActions.value[rule.id] || [],
  })
}

async function loadData() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length) await store.loadAll(fid)
    const rs = await store.loadAutomationRules(fid)
    rules.value = rs
    const actionLists = await Promise.all(rs.map((r) => store.loadRuleActions(r.id)))
    const next = {}
    rs.forEach((r, i) => { next[r.id] = actionLists[i] })
    ruleActions.value = next
  } finally {
    loading.value = false
  }
}

async function toggleRule(rule) {
  const updated = await store.updateAutomationRuleActive(rule.id, !rule.is_active)
  const idx = rules.value.findIndex((r) => r.id === rule.id)
  if (idx >= 0) rules.value[idx] = updated
  emit('refresh')
}

onMounted(loadData)

defineExpose({ loadData })
</script>
