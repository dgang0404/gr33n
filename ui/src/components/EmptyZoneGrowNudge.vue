<template>
  <section
    v-if="visible"
    class="rounded-xl border border-violet-800/60 bg-violet-950/30 px-4 py-4 space-y-3"
    data-test="empty-zone-grow-nudge"
  >
    <div>
      <p class="text-sm font-medium text-violet-100">
        {{ zoneName }} has no grow yet
      </p>
      <p class="text-xs text-violet-200/80 mt-1">
        Guardian can draft a setup change request — plant, cycle, and light feed program — for you to Confirm.
      </p>
    </div>
    <div v-if="proposal" class="border-t border-violet-800/40 pt-3">
      <GuardianActionProposal
        :proposal="proposal"
        :can-operate="canOperate"
        @dismissed="onProposalDismissed"
        @confirmed="onProposalConfirmed"
      />
    </div>
    <div v-else class="flex flex-wrap items-center gap-2">
      <button
        type="button"
        class="px-3 py-2 rounded-lg bg-violet-800/70 text-violet-100 text-xs font-medium hover:bg-violet-700/80 disabled:opacity-50"
        data-test="empty-zone-grow-nudge-offer"
        :disabled="loading"
        @click="offerSetup"
      >
        {{ loading ? 'Drafting…' : 'Set up with Guardian' }}
      </button>
      <button
        type="button"
        class="text-xs text-violet-300/80 hover:text-violet-100"
        @click="visible = false"
      >
        Not now
      </button>
      <p v-if="error" class="text-xs text-red-300 w-full">{{ error }}</p>
    </div>
  </section>
</template>

<script setup>
import { ref, computed } from 'vue'
import api from '../api'
import { useFarmOperate } from '../composables/useFarmOperate'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import GuardianActionProposal from './GuardianActionProposal.vue'

const props = defineProps({
  farmId: { type: Number, required: true },
  zoneId: { type: Number, required: true },
  zoneName: { type: String, default: 'This zone' },
})

const farmIdRef = computed(() => props.farmId)
const { canOperate } = useFarmOperate(farmIdRef)
const proposalsStore = useGuardianProposalsStore()
const guardianPanel = useGuardianPanelStore()

const visible = ref(true)
const loading = ref(false)
const error = ref('')
const proposal = ref(null)

async function offerSetup() {
  if (!props.farmId || !props.zoneId) return
  loading.value = true
  error.value = ''
  try {
    const r = await api.post('/v1/chat/proposals/suggest-empty-zone', {
      farm_id: props.farmId,
      zone_id: props.zoneId,
    })
    proposal.value = r.data
    await proposalsStore.refreshPendingCount(props.farmId)
  } catch (e) {
    error.value = e?.response?.data?.error || e.message || 'Could not draft setup request'
  } finally {
    loading.value = false
  }
}

function onProposalDismissed() {
  proposal.value = null
  visible.value = false
}

function onProposalConfirmed() {
  proposal.value = null
  visible.value = false
  proposalsStore.refreshPendingCount(props.farmId)
  guardianPanel.openPendingTab()
}
</script>
