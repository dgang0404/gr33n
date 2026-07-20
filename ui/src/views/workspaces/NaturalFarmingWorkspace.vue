<template>
  <WorkspaceShell workspace-id="naturalfarming">
    <template #default="{ activeTab }">
      <SwitchoverWizard v-if="activeTab === 'start'" />
      <RecipeLibraryPanel v-else-if="activeTab === 'library'" />
      <MakeBatchPanel v-else-if="activeTab === 'batch'" />
      <RecipesApplyPanel v-else-if="activeTab === 'recipes'" />
      <OnHandPanel v-else-if="activeTab === 'stock'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import SwitchoverWizard from '../../components/naturalfarming/SwitchoverWizard.vue'
import MakeBatchPanel from '../../components/naturalfarming/MakeBatchPanel.vue'
import RecipeLibraryPanel from '../../components/naturalfarming/RecipeLibraryPanel.vue'
import RecipesApplyPanel from '../../components/naturalfarming/RecipesApplyPanel.vue'
import OnHandPanel from '../../components/naturalfarming/OnHandPanel.vue'
</script>
