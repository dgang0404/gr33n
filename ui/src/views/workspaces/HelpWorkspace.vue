<template>
  <WorkspaceShell workspace-id="help">
    <template #default="{ activeTab }">
      <HelpLibraryHub v-if="activeTab === 'library'" />
      <PiSetupGuide v-else-if="activeTab === 'pi-setup'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import HelpLibraryHub from '../HelpLibraryHub.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
</script>
