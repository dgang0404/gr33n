<template>
  <div class="pb-16 space-y-10" data-test="help-library-hub">
    <HelpKnowledgeSurfacesMap />

    <nav
      class="sticky top-[7.5rem] z-10 -mx-4 sm:-mx-6 px-4 sm:px-6 py-2 bg-zinc-950 border-b border-zinc-800/80 flex flex-wrap gap-2"
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

    <section
      id="help-section-guide"
      class="scroll-mt-36 space-y-4"
      data-test="help-library-section-guide"
    >
      <header class="space-y-1 px-4 sm:px-6">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Browse how-to</h2>
        <p class="text-xs text-zinc-500">Glossary, suggested click path, and platform doc citations.</p>
      </header>
      <OperatorGuide embedded hide-surfaces-map />
    </section>

    <section
      id="help-section-knowledge"
      class="scroll-mt-36 border-t border-zinc-800/80 pt-8"
      data-test="help-library-section-knowledge"
    >
      <header class="space-y-1 px-4 sm:px-6 mb-4">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Search your farm</h2>
        <p class="text-xs text-zinc-500">Semantic search and browsable field guides — plain language works.</p>
      </header>
      <FarmKnowledge embedded />
    </section>

    <section
      id="help-section-symptoms"
      class="scroll-mt-36 border-t border-zinc-800/80 pt-8"
      data-test="help-library-section-symptoms"
    >
      <header class="space-y-1 px-4 sm:px-6 mb-2">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Diagnose a symptom</h2>
        <p class="text-xs text-zinc-500">Crop symptom lookup — filter by crop or category.</p>
      </header>
      <SymptomGuide embedded />
    </section>

    <section
      id="help-section-catalog"
      class="scroll-mt-36 border-t border-zinc-800/80 pt-8"
      data-test="help-library-section-catalog"
    >
      <header class="space-y-1 px-4 sm:px-6 mb-4">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Import packs</h2>
        <p class="text-xs text-zinc-500">Commons recipes and seed packs — separate from search.</p>
      </header>
      <CommonsCatalog embedded />
    </section>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'
import OperatorGuide from './OperatorGuide.vue'
import FarmKnowledge from './FarmKnowledge.vue'
import SymptomGuide from './SymptomGuide.vue'
import CommonsCatalog from './CommonsCatalog.vue'

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
  const el = document.getElementById(`help-section-${id}`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  const query = { ...route.query, tab: 'library', section: id }
  delete query.fleet
  router.replace({ path: route.path, query })
}

async function applyDeepLink() {
  const section = resolveSection()
  activeSection.value = section
  await nextTick()
  if (section !== 'guide') {
    document.getElementById(`help-section-${section}`)?.scrollIntoView({ block: 'start' })
  }
}

onMounted(() => { void applyDeepLink() })

watch(
  () => [route.query.tab, route.query.section, route.query.cited_doc, route.query.crop_key],
  () => { void applyDeepLink() },
)
</script>
