<template>
  <WorkspaceShell workspace-id="naturalfarming">
    <template #default="{ activeTab }">
      <SwitchoverWizard v-if="activeTab === 'start'" />
      <RecipeLibraryPanel v-else-if="activeTab === 'library'" />
      <MakeBatchPanel v-else-if="activeTab === 'batch'" />
      <section
        v-else-if="activeTab === 'recipes'"
        class="p-4 max-w-3xl"
        data-test="nf-tab-recipes"
      >
        <h2 class="text-lg font-semibold mb-2">Recipes &amp; apply</h2>
        <p class="text-muted">
          Application recipes and fertigation program links — rehosts Inventory in WS4.
        </p>
      </section>
      <section
        v-else-if="activeTab === 'stock'"
        class="p-4 max-w-3xl"
        data-test="nf-tab-stock"
      >
        <h2 class="text-lg font-semibold mb-2">On hand</h2>
        <p class="text-muted">
          Ready batches and low-stock cues — bridges SuppliesHub in WS5.
        </p>
      </section>
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import SwitchoverWizard from '../../components/naturalfarming/SwitchoverWizard.vue'
import MakeBatchPanel from '../../components/naturalfarming/MakeBatchPanel.vue'
import RecipeLibraryPanel from '../../components/naturalfarming/RecipeLibraryPanel.vue'
</script>
