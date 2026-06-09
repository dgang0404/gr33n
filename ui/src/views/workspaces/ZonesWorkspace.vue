<template>
  <WorkspaceShell workspace-id="zones">
    <template #default="{ activeTab }">
      <div v-if="activeTab === 'rooms'" class="p-0">
        <Zones />
      </div>
      <div v-else-if="activeTab === 'fleet'" class="flex flex-col min-h-[50vh]">
        <div class="px-4 sm:px-6 pt-3 pb-2 border-b border-zinc-800/60 flex flex-wrap gap-2">
          <button
            v-for="sub in fleetSubTabs"
            :key="sub.id"
            type="button"
            class="px-3 py-1.5 text-xs rounded-lg border transition-colors"
            :class="fleetTab === sub.id
              ? 'border-green-700/60 bg-green-950/40 text-green-400'
              : 'border-zinc-800 text-zinc-500 hover:text-zinc-300'"
            @click="selectFleet(sub.id)"
          >
            {{ sub.label }}
          </button>
        </div>
        <Sensors v-if="fleetTab === 'sensors'" embedded group-by-zone />
        <Actuators v-else-if="fleetTab === 'controls'" embedded group-by-zone />
        <LightingPrograms v-else-if="fleetTab === 'lighting'" embedded />
      </div>
      <div v-else-if="activeTab === 'strains'" class="p-0">
        <Plants embedded />
      </div>
    </template>
  </WorkspaceShell>
</template>

<script setup>
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import Zones from '../Zones.vue'
import Sensors from '../Sensors.vue'
import Actuators from '../Actuators.vue'
import LightingPrograms from '../LightingPrograms.vue'
import Plants from '../Plants.vue'
import { FLEET_SUB_TABS, resolveFleetSubTab } from '../../lib/workspaces.js'

const route = useRoute()
const router = useRouter()

const fleetSubTabs = FLEET_SUB_TABS

const fleetTab = computed(() =>
  resolveFleetSubTab(typeof route.query.fleet === 'string' ? route.query.fleet : null),
)

function selectFleet(id) {
  router.replace({ path: route.path, query: { ...route.query, tab: 'fleet', fleet: id } })
}

watch(
  () => [route.query.tab, route.query.fleet],
  () => {
    if (route.query.tab === 'fleet') {
      const resolved = resolveFleetSubTab(typeof route.query.fleet === 'string' ? route.query.fleet : null)
      if (route.query.fleet !== resolved) {
        router.replace({ path: route.path, query: { ...route.query, tab: 'fleet', fleet: resolved } })
      }
    }
  },
  { immediate: true },
)
</script>
