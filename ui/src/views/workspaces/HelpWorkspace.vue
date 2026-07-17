<template>
  <WorkspaceShell workspace-id="help" unified-header>
    <template v-if="isLibraryTab" #subnav-extra>
      <HelpLibrarySectionNav />
    </template>
    <template #default="{ activeTab }">
      <HelpLibraryHub v-if="activeTab === 'library'" />
      <PiSetupGuide v-else-if="activeTab === 'pi-setup'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import HelpLibrarySectionNav from '../../components/HelpLibrarySectionNav.vue'
import HelpLibraryHub from '../HelpLibraryHub.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
import { resolveWorkspaceTab } from '../../lib/workspaces.js'

const route = useRoute()
const isLibraryTab = computed(() =>
  resolveWorkspaceTab('help', typeof route.query.tab === 'string' ? route.query.tab : null) === 'library',
)
</script>
