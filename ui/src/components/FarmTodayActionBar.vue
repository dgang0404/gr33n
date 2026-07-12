<template>
  <section
    class="flex flex-wrap gap-2"
    data-test="farm-today-action-bar"
  >
    <router-link
      v-nav-hint="feedWaterLink"
      :to="feedWaterLink"
      class="inline-flex items-center gap-2 min-h-[44px] px-4 py-2 text-sm font-medium rounded-xl bg-blue-950/50 text-blue-200 border border-blue-800/70 hover:bg-blue-950/70 transition-colors"
      data-test="farm-today-action-feed"
    >
      Feed &amp; water
      <span
        v-if="lowStockCount"
        class="text-[10px] px-1.5 py-0.5 rounded-full bg-amber-900/70 text-amber-200"
        data-test="farm-today-action-low-stock"
      >
        {{ lowStockCount }} low
      </span>
    </router-link>
    <router-link
      v-nav-hint="newTaskLink"
      :to="newTaskLink"
      class="inline-flex items-center min-h-[44px] px-4 py-2 text-sm font-medium rounded-xl bg-green-900/50 text-green-300 border border-green-800/70 hover:bg-green-900/70 transition-colors"
      data-test="farm-today-action-task"
    >
      New task
    </router-link>
    <router-link
      v-nav-hint="schedulesLink"
      :to="schedulesLink"
      class="inline-flex items-center min-h-[44px] px-4 py-2 text-sm font-medium rounded-xl bg-zinc-800 text-zinc-300 border border-zinc-700 hover:bg-zinc-700 transition-colors"
      data-test="farm-today-action-schedules"
    >
      What runs when
    </router-link>
    <router-link
      v-nav-hint="'/zones'"
      to="/zones"
      class="inline-flex items-center min-h-[44px] px-4 py-2 text-sm font-medium rounded-xl bg-zinc-800 text-zinc-300 border border-zinc-700 hover:bg-zinc-700 transition-colors"
      data-test="farm-today-action-zones"
    >
      My zones
    </router-link>
  </section>
</template>

<script setup>
defineProps({
  feedWaterLink: { type: [String, Object], required: true },
  newTaskLink: { type: [String, Object], required: true },
  schedulesLink: { type: [String, Object], required: true },
  lowStockCount: { type: Number, default: 0 },
})
</script>
