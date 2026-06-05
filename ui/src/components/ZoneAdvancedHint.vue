<template>
  <p
    class="text-xs text-zinc-500 bg-zinc-900/60 border border-zinc-800 rounded-lg px-3 py-2"
    data-test="zone-advanced-hint"
  >
    Power settings:
    <router-link to="/automation" class="text-zinc-400 hover:text-green-400">Automations</router-link>
    <span class="text-zinc-700"> · </span>
    <router-link to="/comfort-targets" class="text-zinc-400 hover:text-green-400">Targets &amp; schedules</router-link>
    <span class="text-zinc-700"> · </span>
    <router-link to="/schedules" class="text-zinc-400 hover:text-green-400">Schedules</router-link>
  </p>
</template>
