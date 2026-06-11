<template>
  <WorkspaceShell workspace-id="comfort">
    <template #default="{ activeTab }">
      <div class="p-4 sm:p-6 space-y-6">
        <OperatorConceptBanner :concept-ids="comfortConcepts" />

        <ComfortTargetsHub
          v-if="activeTab === 'comfort'"
          embedded
          section="comfort"
        />

        <template v-else-if="activeTab === 'schedules'">
          <ComfortTargetsHub embedded section="schedules" />
          <div
            id="comfort-advanced-schedules"
            class="border-t border-zinc-800 pt-6 scroll-mt-24"
            data-test="comfort-advanced-schedules"
          >
            <p class="text-xs text-zinc-500 mb-3 flex items-center gap-1">
              Cron expressions and preconditions — power users.
              <ConceptHelpTip concept-id="schedule" />
            </p>
            <Schedules embedded />
          </div>
        </template>

        <template v-else-if="activeTab === 'automations'">
          <ComfortTargetsHub embedded section="automations" />
          <div class="border-t border-zinc-800 pt-6">
            <p class="text-xs text-zinc-500 mb-3 flex items-center gap-1">
              Predicate JSON and full rule editor — power users.
              <ConceptHelpTip concept-id="rule" />
              <ConceptHelpTip concept-id="automation_run" />
            </p>
            <Automation embedded />
          </div>
        </template>

        <div v-else-if="activeTab === 'raw'" class="space-y-4">
          <p class="text-xs text-zinc-500 flex items-center gap-1">
            Same data as Comfort bands, table view.
            <ConceptHelpTip concept-id="setpoint" show-table />
          </p>
          <Setpoints embedded />
        </div>
      </div>
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import ComfortTargetsHub from '../ComfortTargetsHub.vue'
import Schedules from '../Schedules.vue'
import Automation from '../Automation.vue'
import Setpoints from '../Setpoints.vue'
import OperatorConceptBanner from '../../components/OperatorConceptBanner.vue'
import ConceptHelpTip from '../../components/ConceptHelpTip.vue'
import { COMFORT_WORKSPACE_CONCEPTS } from '../../lib/operatorConcepts.js'

const comfortConcepts = COMFORT_WORKSPACE_CONCEPTS
</script>
