<template>
  <div class="pb-16 space-y-10" data-test="help-library-hub">
    <HelpKnowledgeSurfacesMap />

    <section
      id="help-section-guide"
      class="scroll-mt-32 space-y-4"
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
      class="scroll-mt-32 border-t border-zinc-800/80 pt-8"
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
      class="scroll-mt-32 border-t border-zinc-800/80 pt-8"
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
      class="scroll-mt-32 border-t border-zinc-800/80 pt-8"
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
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'
import OperatorGuide from './OperatorGuide.vue'
import FarmKnowledge from './FarmKnowledge.vue'
import SymptomGuide from './SymptomGuide.vue'
import CommonsCatalog from './CommonsCatalog.vue'
</script>
