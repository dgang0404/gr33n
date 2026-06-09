<template>
  <p
    class="text-xs text-zinc-500 bg-zinc-900/60 border border-zinc-800 rounded-lg px-3 py-2"
    data-test="zone-advanced-hint"
  >
    Farm-wide comfort &amp; automation:
    <router-link
      v-nav-hint="'/comfort-targets'"
      to="/comfort-targets"
      class="text-zinc-400 hover:text-green-400"
    >
      Open workspace →
    </router-link>
  </p>
</template>
