<template>
  <WorkspaceShell workspace-id="feedwater">
    <template #default="{ activeTab }">
      <FeedingHub v-if="activeTab === 'daily'" />
      <FeedingAdminHub v-else-if="activeTab === 'programs'" />
      <Fertigation v-else-if="activeTab === 'nutrients' || activeTab === 'advanced'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import FeedingHub from '../FeedingHub.vue'
import FeedingAdminHub from '../FeedingAdminHub.vue'
import Fertigation from '../Fertigation.vue'
</script>
