<template>
  <section
    class="bg-zinc-900 border border-zinc-800 rounded-xl p-6 max-w-2xl space-y-4"
    :data-test="`module-empty-shell-${shell.id}`"
  >
    <div class="flex items-start gap-3">
      <span class="text-3xl shrink-0" aria-hidden="true">{{ shell.icon }}</span>
      <div class="min-w-0">
        <h2 class="text-base font-semibold text-white">{{ shell.title }}</h2>
        <p class="text-sm text-zinc-400 mt-1 leading-relaxed">{{ shell.summary }}</p>
      </div>
    </div>

    <ul class="text-sm text-zinc-400 space-y-1.5 list-disc list-inside leading-relaxed">
      <li v-for="(line, i) in shell.bullets" :key="i">{{ line }}</li>
    </ul>

    <div class="text-xs text-zinc-500 space-y-1.5 border-t border-zinc-800 pt-3">
      <p>
        <span class="text-zinc-600">Workflow:</span>
        <code class="text-zinc-400">{{ shell.workflowDoc }}</code>
        {{ shell.workflowSection }}
      </p>
      <p>
        <span class="text-zinc-600">Starter template:</span>
        <code class="text-zinc-400">{{ shell.templateKey }}</code>
        — see <code class="text-zinc-400">{{ shell.playbookDoc }}</code>
        ({{ shell.playbookSection }}).
      </p>
    </div>

    <p
      v-if="zoneHint"
      class="text-xs text-amber-200/90 bg-amber-950/30 border border-amber-900/50 rounded-lg px-3 py-2"
      :data-test="`module-empty-shell-zone-hint-${shell.id}`"
    >
      {{ zoneHint.message }}
      <router-link
        v-if="zoneHint.actionTo"
        v-nav-hint="'/zones'"
        :to="zoneHint.actionTo"
        class="text-gr33n-500 hover:text-gr33n-400 hover:underline ml-1"
      >
        {{ zoneHint.actionLabel }} →
      </router-link>
    </p>

    <div class="flex flex-wrap items-center gap-2 pt-1">
      <button
        type="button"
        class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-700 hover:bg-green-600 text-white"
        :data-test="`module-empty-shell-primary-${shell.id}`"
        @click="$emit('primary')"
      >
        {{ shell.primaryAction }}
      </button>
      <router-link
        v-nav-hint="'/'"
        to="/"
        class="text-xs text-zinc-400 hover:text-zinc-200"
      >
        Back to Today
      </router-link>
    </div>
  </section>
</template>

<script setup>
import { computed } from 'vue'
import { moduleEmptyShellConfig, moduleShellZoneHint } from '../lib/moduleEmptyShell.js'

const props = defineProps({
  moduleId: { type: String, required: true },
  zoneCount: { type: Number, default: 0 },
})

defineEmits(['primary'])

const shell = computed(() => moduleEmptyShellConfig(props.moduleId))
const zoneHint = computed(() => moduleShellZoneHint(props.moduleId, props.zoneCount))
</script>
