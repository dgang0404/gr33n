<template>
  <div class="space-y-3">
    <article
      v-for="card in cards"
      :key="card.key"
      class="rounded-xl border border-zinc-800 bg-zinc-900/80 p-4"
      :data-test="`nf-guide-card-${card.key}`"
    >
      <h4 class="text-xs uppercase tracking-wide text-zinc-500 mb-2">{{ card.title }}</h4>
      <pre class="text-sm text-zinc-200 whitespace-pre-wrap font-sans leading-relaxed">{{ card.body }}</pre>
    </article>
  </div>
</template>

<script setup>
defineProps({
  cards: { type: Array, default: () => [] },
})
</script>
