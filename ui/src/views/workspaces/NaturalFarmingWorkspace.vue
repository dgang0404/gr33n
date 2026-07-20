<template>
  <WorkspaceShell workspace-id="naturalfarming">
    <template #default="{ activeTab }">
      <section
        v-if="activeTab === 'start'"
        class="p-4 max-w-3xl"
        data-test="nf-tab-start"
      >
        <h2 class="text-lg font-semibold mb-2">Start here</h2>
        <p class="text-muted">
          Switchover wizard for growers moving off bottle nutrients — ships in WS2.
        </p>
      </section>
      <section
        v-else-if="activeTab === 'library'"
        class="p-4 max-w-3xl"
        data-test="nf-tab-library"
      >
        <h2 class="text-lg font-semibold mb-2">Recipe library</h2>
        <p class="text-muted">
          Browse canonical inputs and application recipes from Phase 208 field guides — ships in WS3b.
        </p>
      </section>
      <section
        v-else-if="activeTab === 'batch'"
        class="p-4 max-w-3xl"
        data-test="nf-tab-batch"
      >
        <h2 class="text-lg font-semibold mb-2">Make a batch</h2>
        <p class="text-muted">
          Ferment JMS, JLF, FPJ, and other inputs with step cards from field guides — ships in WS3.
        </p>
      </section>
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
</script>
