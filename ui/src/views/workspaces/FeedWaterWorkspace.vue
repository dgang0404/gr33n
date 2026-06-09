<template>
  <WorkspaceShell workspace-id="feedwater">
    <template #default="{ activeTab }">
      <FeedingHub v-if="activeTab === 'daily'" />
      <FeedingAdminHub v-else-if="activeTab === 'programs'" />
      <Inventory v-else-if="activeTab === 'nutrients'" />
      <Fertigation v-else-if="activeTab === 'advanced'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import FeedingHub from '../FeedingHub.vue'
import FeedingAdminHub from '../FeedingAdminHub.vue'
import Inventory from '../Inventory.vue'
import Fertigation from '../Fertigation.vue'
</script>
