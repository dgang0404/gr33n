<template>
  <nav
    class="flex flex-wrap gap-2"
    data-test="help-library-jump"
    aria-label="Library sections"
  >
    <a
      href="#help-section-guide"
      class="text-xs px-2.5 py-1 rounded-full border transition-colors"
      :class="activeSection === 'guide'
        ? 'border-green-700 text-green-300 bg-green-950/40'
        : 'border-zinc-700 text-zinc-400 hover:text-zinc-200'"
      data-test="help-library-jump-guide"
      @click.prevent="scrollToSection('guide')"
    >
      How-to
    </a>
  </nav>
</template>

<script setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { scrollToHelpLibrarySection } from '../lib/helpLibraryScroll.js'

const route = useRoute()
const router = useRouter()
const activeSection = ref('guide')

function scrollToSection(id) {
  activeSection.value = id
  scrollToHelpLibrarySection(id)
  const query = { ...route.query, tab: 'library', section: id }
  delete query.fleet
  router.replace({ path: route.path, query })
}

async function applyDeepLink() {
  activeSection.value = 'guide'
  await nextTick()
  if (route.query.section === 'guide') {
    scrollToHelpLibrarySection('guide', { smooth: false })
  }
}

onMounted(() => { void applyDeepLink() })

watch(
  () => [route.query.tab, route.query.section, route.query.cited_doc],
  () => { void applyDeepLink() },
)
</script>
