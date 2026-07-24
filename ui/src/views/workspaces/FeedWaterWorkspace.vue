<template>
  <WorkspaceShell workspace-id="feedwater">
    <template #default="{ activeTab }">
      <div class="space-y-4">
        <OperatorConceptBanner :concept-ids="feedWaterConcepts" :relationships="feedWaterRelationships" />

        <FeedingHub v-if="activeTab === 'daily'" />
        <FeedingAdminHub v-else-if="activeTab === 'programs'" />
        <Fertigation v-else-if="activeTab === 'nutrients' || activeTab === 'advanced'" />
      </div>
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import FeedingHub from '../FeedingHub.vue'
import FeedingAdminHub from '../FeedingAdminHub.vue'
import Fertigation from '../Fertigation.vue'
import OperatorConceptBanner from '../../components/OperatorConceptBanner.vue'
import { FEED_WATER_CONCEPT_RELATIONSHIPS, FEED_WATER_WORKSPACE_CONCEPTS } from '../../lib/operatorConcepts.js'

const feedWaterConcepts = FEED_WATER_WORKSPACE_CONCEPTS
const feedWaterRelationships = FEED_WATER_CONCEPT_RELATIONSHIPS
</script>
