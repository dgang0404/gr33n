<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="guardianPanel.open"
        class="fixed inset-0 z-40 flex flex-col md:flex-row md:justify-end"
        data-test="guardian-drawer"
      >
        <!-- Backdrop (desktop: dim left; mobile: full) -->
        <div
          class="flex-1 bg-black/50 md:bg-black/40"
          data-test="guardian-drawer-backdrop"
          aria-hidden="true"
          @click="guardianPanel.close()"
        />

        <!-- Panel: bottom sheet on mobile, right rail on md+ -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="translate-y-full md:translate-y-0 md:translate-x-full"
          enter-to-class="translate-y-0 md:translate-x-0"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="translate-y-0 md:translate-x-0"
          leave-to-class="translate-y-full md:translate-y-0 md:translate-x-full"
          appear
        >
          <aside
            v-if="guardianPanel.open"
            role="dialog"
            aria-modal="true"
            aria-labelledby="guardian-drawer-title"
            class="w-full md:w-[min(28rem,100vw)] md:max-w-md flex flex-col bg-zinc-950 border-t md:border-t-0 md:border-l border-zinc-800 shadow-2xl max-h-[min(88vh,100dvh)] md:max-h-none md:h-full"
            style="padding-bottom: max(0px, env(safe-area-inset-bottom))"
            @click.stop
          >
            <header class="flex items-center justify-between gap-3 px-4 py-3 border-b border-zinc-800 shrink-0">
              <div class="min-w-0">
                <h2 id="guardian-drawer-title" class="text-sm font-semibold text-green-400 flex items-center gap-2">
                  Farm Guardian
                </h2>
                <p v-if="farmContext.farmId" class="text-[10px] text-zinc-500 truncate">
                  Farm #{{ farmContext.farmId }}
                  <span v-if="farmContext.selectedFarm?.name"> · {{ farmContext.selectedFarm.name }}</span>
                </p>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <RouterLink
                  to="/chat"
                  class="text-[10px] text-zinc-500 hover:text-zinc-300 underline"
                  data-test="guardian-drawer-full-page"
                  @click="guardianPanel.close()"
                >
                  Full page
                </RouterLink>
                <button
                  type="button"
                  class="p-1.5 rounded-md text-zinc-400 hover:text-white hover:bg-zinc-800"
                  aria-label="Close Guardian"
                  data-test="guardian-drawer-close"
                  @click="guardianPanel.close()"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
                  </svg>
                </button>
              </div>
            </header>

            <section
              v-if="capabilities.isLite"
              data-test="chat-lite-banner"
              class="mx-4 mt-3 rounded-xl border border-amber-900/60 bg-amber-950/40 px-3 py-2 text-xs text-amber-200 shrink-0"
            >
              Farm Guardian is not available on this installation (Lite mode).
            </section>

            <div v-else class="flex-1 min-h-0 overflow-y-auto px-4 py-3">
              <GuardianChatPanel layout="compact" />
            </div>

            <footer class="px-4 py-2 border-t border-zinc-800 text-[10px] text-zinc-600 shrink-0">
              Guardian proposes changes; you confirm. Confirmed actions appear in the farm audit log.
            </footer>
          </aside>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { onMounted, watch } from 'vue'
import GuardianChatPanel from './GuardianChatPanel.vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'

const guardianPanel = useGuardianPanelStore()
const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
})

watch(
  () => guardianPanel.open,
  async (isOpen) => {
    if (isOpen && !capabilities.loaded) await capabilities.fetch()
  },
)
</script>
