<template>
  <section
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4"
    data-test="help-knowledge-surfaces-map"
  >
    <div class="space-y-1">
      <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">
        What lives where
      </h2>
      <p class="text-xs text-zinc-500 leading-relaxed">
        Three related knowledge surfaces — each does something different. Browse here without waiting for a Guardian citation.
      </p>
    </div>
    <div class="grid gap-3 sm:grid-cols-2">
      <router-link
        v-for="card in cards"
        :key="card.id"
        :to="card.to"
        class="rounded-lg border border-zinc-800 bg-zinc-950/60 px-4 py-3 hover:border-green-800/60 hover:bg-green-950/20 transition-colors group"
        :data-test="`help-surface-card-${card.id}`"
      >
        <p class="text-sm font-semibold text-green-400 group-hover:text-green-300">{{ card.title }}</p>
        <p class="text-xs text-zinc-500 mt-1 leading-relaxed">{{ card.blurb }}</p>
      </router-link>
    </div>
  </section>
</template>

<script setup>
const cards = [
  {
    id: 'knowledge',
    title: 'Knowledge',
    blurb: 'Semantic search over your farm’s indexed notes, tasks, and ingested docs — plain language, not exact keywords.',
    to: { path: '/operator-guide', query: { tab: 'library', section: 'knowledge' } },
  },
  {
    id: 'catalog',
    title: 'Catalog',
    blurb: 'Import Commons packs (recipes, seed packs) — separate from search; for bringing in starter content.',
    to: { path: '/operator-guide', query: { tab: 'library', section: 'catalog' } },
  },
  {
    id: 'symptoms',
    title: 'Symptom guide',
    blurb: 'Crop symptom lookup — what wilting, spotting, or deficiency patterns mean for a given crop.',
    to: { path: '/operator-guide', query: { tab: 'library', section: 'symptoms' } },
  },
]
</script>
