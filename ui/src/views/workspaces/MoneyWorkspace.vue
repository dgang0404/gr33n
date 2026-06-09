<template>
  <WorkspaceShell workspace-id="money">
    <template #default="{ activeTab }">
      <MoneyHub v-if="activeTab === 'summary'" />
      <Costs v-else-if="activeTab === 'ledger'" />
      <SuppliesHub v-else-if="activeTab === 'supplies'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import MoneyHub from '../MoneyHub.vue'
import Costs from '../Costs.vue'
import SuppliesHub from '../SuppliesHub.vue'
</script>
