<template>
  <WorkspaceShell workspace-id="hardware">
    <template #default="{ activeTab }">
      <GpioBoard v-if="activeTab === 'board'" />
      <HardwareDevicesPanel v-else-if="activeTab === 'devices'" />
      <PiSetupGuide v-else-if="activeTab === 'reference'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import GpioBoard from '../hardware/GpioBoard.vue'
import HardwareDevicesPanel from '../../components/workspaces/HardwareDevicesPanel.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
</script>
