<template>
  <WorkspaceShell workspace-id="naturalfarming">
    <template #default="{ activeTab }">
      <div class="space-y-4">
        <OperatorConceptBanner
          :concept-ids="nfConcepts"
          :relationships="nfRelationships"
        />
        <MakeBatchPanel v-if="activeTab === 'batch'" />
        <RecipeLibraryPanel v-else-if="activeTab === 'library'" />
        <RecipesApplyPanel v-else-if="activeTab === 'recipes'" />
        <FarmRowsPanel v-else-if="activeTab === 'manage'" />
      </div>
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import OperatorConceptBanner from '../../components/OperatorConceptBanner.vue'
import MakeBatchPanel from '../../components/naturalfarming/MakeBatchPanel.vue'
import RecipeLibraryPanel from '../../components/naturalfarming/RecipeLibraryPanel.vue'
import RecipesApplyPanel from '../../components/naturalfarming/RecipesApplyPanel.vue'
import FarmRowsPanel from '../../components/naturalfarming/FarmRowsPanel.vue'
import {
  NATURAL_FARMING_CONCEPT_RELATIONSHIPS,
  NATURAL_FARMING_WORKSPACE_CONCEPTS,
} from '../../lib/operatorConcepts.js'

const nfConcepts = NATURAL_FARMING_WORKSPACE_CONCEPTS
const nfRelationships = NATURAL_FARMING_CONCEPT_RELATIONSHIPS
</script>
