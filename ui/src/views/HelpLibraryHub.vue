<template>
  <div class="px-4 sm:px-6 pt-2 pb-12 space-y-8" data-test="help-library-hub">
    <HelpKnowledgeSurfacesMap />

    <section
      id="help-section-guide"
      class="scroll-mt-4 space-y-4"
      data-test="help-library-section-guide"
    >
      <header class="space-y-1">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">Browse how-to</h2>
        <p class="text-xs text-zinc-500">Glossary, suggested click path, and platform doc citations.</p>
      </header>
      <OperatorGuide embedded hide-surfaces-map />
    </section>
  </div>
</template>

<script setup>
import HelpKnowledgeSurfacesMap from '../components/HelpKnowledgeSurfacesMap.vue'
import OperatorGuide from './OperatorGuide.vue'
</script>
