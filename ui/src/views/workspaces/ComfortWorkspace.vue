<template>
  <WorkspaceShell workspace-id="comfort">
    <template #default="{ activeTab }">
      <div class="p-4 sm:p-6 space-y-6">
        <ComfortTargetsHub
          v-if="activeTab === 'comfort'"
          embedded
          section="comfort"
        />

        <template v-else-if="activeTab === 'schedules'">
          <ComfortTargetsHub embedded section="schedules" />
          <div class="border-t border-zinc-800 pt-6">
            <p class="text-xs text-zinc-500 mb-3">Cron expressions and preconditions — power users.</p>
            <Schedules embedded />
          </div>
        </template>

        <template v-else-if="activeTab === 'automations'">
          <ComfortTargetsHub embedded section="automations" />
          <div class="border-t border-zinc-800 pt-6">
            <p class="text-xs text-zinc-500 mb-3">Predicate JSON and full rule editor — power users.</p>
            <Automation embedded />
          </div>
        </template>

        <Setpoints v-else-if="activeTab === 'raw'" embedded />
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
</script>
