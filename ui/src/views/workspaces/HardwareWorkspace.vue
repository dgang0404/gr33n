<template>
  <WorkspaceShell workspace-id="hardware">
    <template #default="{ activeTab }">
      <HardwareBoardStub v-if="activeTab === 'board'" class="p-4 sm:p-6" />
      <HardwareDevicesPanel v-else-if="activeTab === 'devices'" />
      <PiSetupGuide v-else-if="activeTab === 'reference'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import HardwareBoardStub from '../../components/workspaces/HardwareBoardStub.vue'
import HardwareDevicesPanel from '../../components/workspaces/HardwareDevicesPanel.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
</script>
