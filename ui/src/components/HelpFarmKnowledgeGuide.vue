<template>
  <section
    class="space-y-5 pt-2 border-t border-zinc-800"
    data-test="help-farm-knowledge-guide"
  >
    <header class="space-y-1">
      <h3 class="text-white font-semibold text-sm">Farm knowledge — how the pieces connect</h3>
      <p class="text-xs text-zinc-500 leading-relaxed">
        Several Help surfaces look like “knowledge” but do different jobs. Farm Imports is a history of catalog packs applied — not a separate store.
      </p>
    </header>

    <div class="overflow-x-auto rounded-xl border border-zinc-800">
      <table class="min-w-full text-xs text-left">
        <thead class="bg-zinc-950/80 text-zinc-500 uppercase tracking-wide">
          <tr>
            <th class="px-3 py-2 font-semibold">Layer</th>
            <th class="px-3 py-2 font-semibold">Job</th>
            <th class="px-3 py-2 font-semibold">Examples</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-800 text-zinc-400">
          <tr v-for="row in layers" :key="row.layer" class="bg-zinc-900/40">
            <td class="px-3 py-2 font-medium text-gr33n-400 whitespace-nowrap">{{ row.layer }}</td>
            <td class="px-3 py-2">{{ row.job }}</td>
            <td class="px-3 py-2">{{ row.examples }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="space-y-2">
      <h4 class="text-xs font-semibold uppercase tracking-widest text-zinc-500">If you want X, do Y</h4>
      <div class="overflow-x-auto rounded-xl border border-zinc-800">
        <table class="min-w-full text-xs text-left">
          <thead class="bg-zinc-950/80 text-zinc-500 uppercase tracking-wide">
            <tr>
              <th class="px-3 py-2 font-semibold">I want to…</th>
              <th class="px-3 py-2 font-semibold">Use…</th>
              <th class="px-3 py-2 font-semibold">Other install?</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800 text-zinc-400">
            <tr v-for="(row, i) in actions" :key="i" class="bg-zinc-900/40">
              <td class="px-3 py-2">{{ row.want }}</td>
              <td class="px-3 py-2 text-zinc-300">{{ row.use }}</td>
              <td class="px-3 py-2 whitespace-nowrap">{{ row.crossInstall }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="space-y-2">
      <h4 class="text-xs font-semibold uppercase tracking-widest text-zinc-500">What crosses an install boundary?</h4>
      <div class="overflow-x-auto rounded-xl border border-zinc-800">
        <table class="min-w-full text-xs text-left">
          <thead class="bg-zinc-950/80 text-zinc-500 uppercase tracking-wide">
            <tr>
              <th class="px-3 py-2 font-semibold">Category</th>
              <th class="px-3 py-2 font-semibold">Moves?</th>
              <th class="px-3 py-2 font-semibold">How</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800 text-zinc-400">
            <tr v-for="(row, i) in boundaries" :key="i" class="bg-zinc-900/40">
              <td class="px-3 py-2 text-zinc-300">{{ row.category }}</td>
              <td class="px-3 py-2 whitespace-nowrap">{{ row.moves }}</td>
              <td class="px-3 py-2">{{ row.how }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3 space-y-2 text-xs text-zinc-400 leading-relaxed">
      <p class="text-gr33n-400 font-semibold text-sm">{{ insertCommons.title }}</p>
      <p>{{ insertCommons.lead }}</p>
      <p><strong class="text-zinc-300">Included:</strong> {{ insertCommons.includes.join('; ') }}.</p>
      <p><strong class="text-zinc-300">Never included:</strong> {{ insertCommons.excludes }}</p>
      <p class="text-zinc-500">{{ insertCommons.notFor }}</p>
      <p class="text-zinc-600 pt-1">
        <router-link
          v-nav-hint="'/settings'"
          to="/settings"
          class="text-gr33n-500 hover:underline"
        >Settings → Insert Commons</router-link>
        · Preview ingest JSON before sync · Opt out anytime stops future sends
      </p>
    </div>
  </section>
</template>

<script setup>
import {
  FARM_KNOWLEDGE_ACTIONS,
  FARM_KNOWLEDGE_BOUNDARIES,
  FARM_KNOWLEDGE_LAYERS,
  INSERT_COMMONS_SUMMARY,
} from '../lib/farmKnowledgeGuide.js'

const layers = FARM_KNOWLEDGE_LAYERS
const actions = FARM_KNOWLEDGE_ACTIONS
const boundaries = FARM_KNOWLEDGE_BOUNDARIES
const insertCommons = INSERT_COMMONS_SUMMARY
</script>
