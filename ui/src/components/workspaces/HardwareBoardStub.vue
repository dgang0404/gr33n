<template>
  <p class="text-sm text-zinc-400">
    Live GPIO board view is available under
    <RouterLink v-nav-hint="'/hardware'" :to="{ path: '/hardware', query: { tab: 'board' } }" class="text-green-500 hover:text-green-400">
      Hardware → GPIO board
    </RouterLink>.
    Use
    <RouterLink v-nav-hint="'/hardware'" :to="{ path: '/hardware', query: { tab: 'reference' } }" class="text-green-500 hover:text-green-400">
      Wiring guide
    </RouterLink>
    for pin maps while setting up.
  </p>
</template>

<script setup>
</script>
