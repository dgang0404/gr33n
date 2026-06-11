<template>
  <WorkspaceShell workspace-id="help">
    <template #default="{ activeTab }">
      <OperatorGuide v-if="activeTab === 'guide'" embedded />
      <PiSetupGuide v-else-if="activeTab === 'pi-setup'" embedded />
      <FarmKnowledge v-else-if="activeTab === 'knowledge'" embedded />
      <CommonsCatalog v-else-if="activeTab === 'catalog'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import OperatorGuide from '../OperatorGuide.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
import FarmKnowledge from '../FarmKnowledge.vue'
import CommonsCatalog from '../CommonsCatalog.vue'
</script>
