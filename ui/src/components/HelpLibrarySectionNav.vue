<template>
  <nav
    class="flex flex-wrap gap-2"
    data-test="help-library-jump"
    aria-label="Library sections"
  >
    <a
      v-for="s in sections"
      :key="s.id"
      :href="`#help-section-${s.id}`"
      class="text-xs px-2.5 py-1 rounded-full border transition-colors"
      :class="activeSection === s.id
        ? 'border-green-700 text-green-300 bg-green-950/40'
        : 'border-zinc-700 text-zinc-400 hover:text-zinc-200'"
      :data-test="`help-library-jump-${s.id}`"
      @click.prevent="scrollToSection(s.id)"
    >
      {{ s.label }}
    </a>
  </nav>
</template>

<script setup>
import { nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { scrollToHelpLibrarySection } from '../lib/helpLibraryScroll.js'

const LEGACY_SECTIONS = new Set(['guide', 'knowledge', 'symptoms', 'catalog'])

const route = useRoute()
const router = useRouter()
const activeSection = ref('guide')

const sections = [
  { id: 'guide', label: 'How-to' },
  { id: 'knowledge', label: 'Search' },
  { id: 'symptoms', label: 'Symptoms' },
  { id: 'catalog', label: 'Import' },
]

function resolveSection() {
  const section = route.query.section
  if (typeof section === 'string' && LEGACY_SECTIONS.has(section)) {
    return section
  }
  const tab = route.query.tab
  if (typeof tab === 'string' && LEGACY_SECTIONS.has(tab)) {
    return tab
  }
  return 'guide'
}

function scrollToSection(id) {
  activeSection.value = id
  scrollToHelpLibrarySection(id)
  const query = { ...route.query, tab: 'library', section: id }
  delete query.fleet
  router.replace({ path: route.path, query })
}

async function applyDeepLink() {
  const section = resolveSection()
  activeSection.value = section
  await nextTick()
  if (typeof route.query.section === 'string' && LEGACY_SECTIONS.has(route.query.section)) {
    scrollToHelpLibrarySection(section, { smooth: false })
  }
}

onMounted(() => { void applyDeepLink() })

watch(
  () => [route.query.tab, route.query.section, route.query.cited_doc, route.query.crop_key],
  () => { void applyDeepLink() },
)
</script>
