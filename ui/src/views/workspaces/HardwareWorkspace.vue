<template>
  <WorkspaceShell workspace-id="hardware">
    <template #default="{ activeTab }">
      <GpioBoard v-if="activeTab === 'board'" />
      <HardwareDevicesPanel v-else-if="activeTab === 'devices'" />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import GpioBoard from '../hardware/GpioBoard.vue'
import HardwareDevicesPanel from '../../components/workspaces/HardwareDevicesPanel.vue'
</script>
