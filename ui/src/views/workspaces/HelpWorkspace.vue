<template>
  <WorkspaceShell workspace-id="help" unified-header>
    <template v-if="isLibraryTab" #subnav-extra>
      <HelpLibrarySectionNav />
    </template>
    <template #default="{ activeTab }">
      <HelpLibraryHub v-if="activeTab === 'library'" />
      <PiSetupGuide v-else-if="activeTab === 'pi-setup'" embedded />
      <FarmKnowledge v-else-if="activeTab === 'knowledge'" embedded />
      <SymptomGuide v-else-if="activeTab === 'symptoms'" embedded />
      <CommonsCatalog v-else-if="activeTab === 'catalog'" embedded />
    </template>
  </WorkspaceShell>
</template>

<script setup>
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import WorkspaceShell from '../../components/WorkspaceShell.vue'
import HelpLibrarySectionNav from '../../components/HelpLibrarySectionNav.vue'
import HelpLibraryHub from '../HelpLibraryHub.vue'
import PiSetupGuide from '../PiSetupGuide.vue'
import FarmKnowledge from '../FarmKnowledge.vue'
import SymptomGuide from '../SymptomGuide.vue'
import CommonsCatalog from '../CommonsCatalog.vue'
import { resolveWorkspaceTab } from '../../lib/workspaces.js'

/** Legacy ?section= deep links from before Phase 201 tabs. */
const LEGACY_SECTION_TAB = {
  knowledge: 'knowledge',
  catalog: 'catalog',
  symptoms: 'symptoms',
}

const route = useRoute()
const router = useRouter()
const isLibraryTab = computed(() =>
  resolveWorkspaceTab('help', typeof route.query.tab === 'string' ? route.query.tab : null) === 'library',
)

watch(
  () => route.query.section,
  (section) => {
    if (typeof section !== 'string') return
    const tab = LEGACY_SECTION_TAB[section]
    if (!tab) return
    const query = { ...route.query, tab }
    delete query.section
    router.replace({ path: '/operator-guide', query })
  },
  { immediate: true },
)
</script>
