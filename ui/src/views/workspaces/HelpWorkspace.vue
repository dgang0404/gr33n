<template>
  <WorkspaceShell workspace-id="help" unified-header>
    <template v-if="isLibraryTab" #subnav-extra>
      <HelpLibrarySectionNav />
    </template>
    <template #default="{ activeTab }">
      <HelpLibraryHub v-if="activeTab === 'library'" />
      <PiSetupGuide v-else-if="activeTab === 'pi-setup'" embedded />
      <SymptomGuide v-else-if="activeTab === 'symptoms'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import HelpLibrarySectionNav from '../../components/HelpLibrarySectionNav.vue'
import HelpLibraryHub from '../HelpLibraryHub.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
import SymptomGuide from '../SymptomGuide.vue'
import { resolveWorkspaceTab } from '../../lib/workspaces.js'

const route = useRoute()
const router = useRouter()
const isLibraryTab = computed(() =>
  resolveWorkspaceTab('help', typeof route.query.tab === 'string' ? route.query.tab : null) === 'library',
)

watch(
  () => route.query.section,
  (section) => {
    if (section === 'symptoms') {
      const query = { ...route.query, tab: 'symptoms' }
      delete query.section
      router.replace({ path: '/operator-guide', query })
    }
  },
  { immediate: true },
)
</script>
