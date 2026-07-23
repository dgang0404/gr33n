<template>
  <div class="workspace-shell flex-1 min-h-0 flex flex-col" data-test="workspace-shell">
    <!-- Help (and future workspaces): title + tabs + section pills on one chrome row -->
    <div
      v-if="unifiedHeader"
      class="workspace-shell__chrome shrink-0 z-20 bg-zinc-950 border-b border-zinc-800/80 px-4 sm:px-6 py-2"
      data-test="workspace-shell-unified-chrome"
    >
      <div class="flex flex-wrap items-center gap-x-2 gap-y-2">
        <h1 class="text-lg sm:text-xl font-semibold text-white flex items-center gap-2 shrink-0">
          <span v-if="headerIcon" class="text-xl" aria-hidden="true">{{ headerIcon }}</span>
          {{ headerTitle }}
        </h1>
        <p
          v-if="headerSubtitle"
          class="text-xs text-zinc-500 truncate max-w-[10rem] sm:max-w-xs md:max-w-md lg:max-w-xl"
        >
          {{ headerSubtitle }}
        </p>

        <span class="hidden sm:block h-4 w-px bg-zinc-800 shrink-0" aria-hidden="true" />

        <div
          class="hidden sm:flex gap-1 overflow-x-auto shrink-0"
          role="tablist"
          :aria-label="`${headerTitle} sections`"
        >
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            role="tab"
            :aria-selected="activeTab === tab.id"
            class="px-2.5 py-1.5 text-sm rounded-lg whitespace-nowrap transition-colors"
            :class="activeTab === tab.id
              ? 'bg-zinc-800 text-white font-medium'
              : 'text-zinc-500 hover:text-zinc-200 hover:bg-zinc-900'"
            @click="selectTab(tab.id)"
          >
            {{ tab.label }}
          </button>
        </div>

        <label class="sm:hidden flex items-center gap-2 min-w-[8rem] flex-1">
          <span class="text-xs uppercase tracking-wide text-zinc-500 shrink-0">Tab</span>
          <select
            :value="activeTab"
            class="flex-1 bg-zinc-900 border border-zinc-700 text-zinc-200 text-sm rounded-lg px-2 py-1.5"
            @change="selectTab($event.target.value)"
          >
            <option v-for="tab in tabs" :key="tab.id" :value="tab.id">{{ tab.label }}</option>
          </select>
        </label>

        <template v-if="$slots['subnav-extra']">
          <span class="hidden sm:block h-4 w-px bg-zinc-800 shrink-0" aria-hidden="true" />
          <div
            class="flex items-center overflow-x-auto shrink-0"
            data-test="workspace-shell-subnav-extra"
          >
            <slot name="subnav-extra" :active-tab="activeTab" />
          </div>
        </template>

        <slot name="actions" />
      </div>
    </div>

    <template v-else>
      <header class="px-4 sm:px-6 border-b border-zinc-800/80" :class="compact ? 'pt-3 pb-2' : 'pt-6 pb-4'">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h1
              class="font-semibold text-white flex items-center gap-2"
              :class="compact ? 'text-xl' : 'text-2xl'"
            >
              <span v-if="headerIcon" class="text-xl" aria-hidden="true">{{ headerIcon }}</span>
              {{ headerTitle }}
            </h1>
            <p v-if="headerSubtitle" class="text-sm text-zinc-400 mt-1 max-w-2xl">{{ headerSubtitle }}</p>
          </div>
          <slot name="actions" />
        </div>
      </header>

      <div
        class="workspace-shell__subnav shrink-0 z-20 bg-zinc-950 border-b border-zinc-800/80 px-4 sm:px-6"
      >
        <div class="hidden sm:flex gap-1 py-2 overflow-x-auto" role="tablist" :aria-label="`${headerTitle} sections`">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            type="button"
            role="tab"
            :aria-selected="activeTab === tab.id"
            class="px-3 py-2 text-sm rounded-lg whitespace-nowrap transition-colors"
            :class="activeTab === tab.id
              ? 'bg-zinc-800 text-white font-medium'
              : 'text-zinc-500 hover:text-zinc-200 hover:bg-zinc-900'"
            @click="selectTab(tab.id)"
          >
            {{ tab.label }}
          </button>
        </div>

        <label class="sm:hidden flex items-center gap-2 py-2">
          <span class="text-xs uppercase tracking-wide text-zinc-500 shrink-0">Section</span>
          <select
            :value="activeTab"
            class="flex-1 bg-zinc-900 border border-zinc-700 text-zinc-200 text-sm rounded-lg px-2 py-2"
            @change="selectTab($event.target.value)"
          >
            <option v-for="tab in tabs" :key="tab.id" :value="tab.id">{{ tab.label }}</option>
          </select>
        </label>

        <div
          v-if="showJumpRail"
          class="flex flex-wrap items-center gap-2 pb-2 border-t border-zinc-800/60 sm:border-0"
        >
          <span class="text-[10px] uppercase tracking-widest text-zinc-600 font-semibold">Jump to</span>
          <RouterLink
            v-for="link in jumpLinks"
            :key="link.to"
            v-nav-hint="link.to"
            :to="link.to"
            class="text-xs px-2 py-1 rounded-full border border-zinc-700 text-zinc-400 hover:text-green-400 hover:border-green-700/60 transition-colors"
          >
            {{ link.label }}
          </RouterLink>
        </div>

        <div
          v-if="$slots['subnav-extra']"
          class="workspace-shell__subnav-extra border-t border-zinc-800/60 py-2"
          data-test="workspace-shell-subnav-extra"
        >
          <slot name="subnav-extra" :active-tab="activeTab" />
        </div>
      </div>
    </template>

    <div
      ref="contentEl"
      class="workspace-shell__content flex-1 min-h-0 overflow-y-auto overscroll-y-contain px-4 sm:px-6 pt-4 sm:pt-6 pb-20 md:pb-6"
    >
      <slot :active-tab="activeTab" />
    </div>
  </div>
</template>

<script setup>
import { computed, watch, ref, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { relatedWorkspaces, resolveWorkspaceTab, WORKSPACES } from '../lib/workspaces.js'

const props = defineProps({
  workspaceId: { type: String, required: true },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  icon: { type: String, default: '' },
  compact: { type: Boolean, default: false },
  /** Single-row chrome: title, subtitle, tabs, and subnav-extra inline (Help). */
  unifiedHeader: { type: Boolean, default: false },
})

const route = useRoute()
const router = useRouter()
const contentEl = ref(null)

const ws = computed(() => WORKSPACES[props.workspaceId])
const tabs = computed(() => ws.value?.tabs ?? [])
const headerTitle = computed(() => props.title || ws.value?.label || '')
const headerSubtitle = computed(() => props.subtitle || ws.value?.subtitle || '')
const headerIcon = computed(() => props.icon || ws.value?.icon || '')

const activeTab = computed(() =>
  resolveWorkspaceTab(props.workspaceId, typeof route.query.tab === 'string' ? route.query.tab : null),
)

const jumpLinks = computed(() => {
  const base = ws.value?.route
  if (!base) return []
  return relatedWorkspaces(base).map((to) => {
    const target = Object.values(WORKSPACES).find((w) => w.route === to)
    return { to, label: target?.label ?? to }
  })
})

const showJumpRail = computed(() => {
  if (props.workspaceId === 'help') return false
  return jumpLinks.value.length > 0
})

function selectTab(tabId) {
  const query = { ...route.query, tab: tabId }
  if (tabId !== 'fleet') delete query.fleet
  if (props.workspaceId === 'help' && tabId !== 'library') delete query.section
  router.replace({ path: route.path, query })
}

watch(activeTab, async () => {
  await nextTick()
  if (contentEl.value) contentEl.value.scrollTop = 0
})

watch(
  () => route.query.tab,
  (tab) => {
    const resolved = resolveWorkspaceTab(props.workspaceId, typeof tab === 'string' ? tab : null)
    if (tab !== resolved) {
      router.replace({ path: route.path, query: { ...route.query, tab: resolved } })
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.workspace-shell__content {
  scrollbar-gutter: stable;
}

@media (prefers-reduced-motion: reduce) {
  .workspace-shell__subnav :deep(.nav-related) {
    animation: none !important;
  }
}
</style>
