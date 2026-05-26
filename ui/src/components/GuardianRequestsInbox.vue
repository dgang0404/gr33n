<template>
  <div class="space-y-3" data-test="guardian-requests-inbox">
    <p class="text-xs text-zinc-500">
      Change requests Guardian opened for your review. Nothing applies until you Confirm.
    </p>

    <p v-if="!farmId" class="text-xs text-amber-400/90" data-test="guardian-inbox-no-farm">
      Select a farm to see pending requests.
    </p>

    <p v-else-if="store.loading" class="text-xs text-zinc-500">Loading pending requests…</p>

    <p v-else-if="store.error" class="text-xs text-red-400" data-test="guardian-inbox-error">
      {{ store.error }}
    </p>

    <p
      v-else-if="store.proposals.length === 0"
      class="text-xs text-zinc-500 italic"
      data-test="guardian-inbox-empty"
    >
      No pending requests for this farm.
    </p>

    <div v-else class="space-y-2" data-test="guardian-inbox-list">
      <GuardianActionProposal
        v-for="p in store.proposals"
        :key="p.proposal_id"
        :proposal="normalizeForCard(p)"
        :can-operate="canOperate"
        @confirmed="onConfirmed"
        @dismissed="onDismissed"
        @error="onError"
      />
    </div>

    <RouterLink
      v-if="farmId && showFullPageLink"
      to="/guardian/requests"
      class="inline-block text-[10px] text-zinc-500 hover:text-green-400 underline"
      data-test="guardian-inbox-full-page"
      @click="guardianPanel.close()"
    >
      Open full-page inbox
    </RouterLink>
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import GuardianActionProposal from './GuardianActionProposal.vue'
import { useFarmOperate } from '../composables/useFarmOperate'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'

const props = defineProps({
  showFullPageLink: { type: Boolean, default: true },
})

const farmContext = useFarmContextStore()
const guardianPanel = useGuardianPanelStore()
const store = useGuardianProposalsStore()

const farmId = computed(() => farmContext.farmId)
const farmIdRef = computed(() => farmId.value)
const { canOperate } = useFarmOperate(farmIdRef)

function normalizeForCard(p) {
  return {
    proposal_id: p.proposal_id,
    tool: p.tool,
    args: p.args,
    summary: p.summary,
    risk_tier: p.risk_tier || 'medium',
    expires_at: p.expires_at,
    status: p.status === 'pending' ? 'pending' : p.status,
  }
}

async function load() {
  if (farmId.value) await store.fetch(farmId.value)
}

function onConfirmed({ proposal }) {
  store.removeProposal(proposal.proposal_id)
  store.refreshPendingCount(farmId.value)
}

function onDismissed({ proposal }) {
  store.patchProposal(proposal.proposal_id, { status: 'dismissed' })
  store.removeProposal(proposal.proposal_id)
  store.pendingCount = Math.max(0, store.pendingCount - 1)
}

function onError({ proposal, error }) {
  store.patchProposal(proposal.proposal_id, { error })
}

watch(farmId, () => load(), { immediate: true })

watch(
  () => guardianPanel.open && guardianPanel.drawerTab === 'pending',
  (active) => {
    if (active) load()
  },
)

defineExpose({ reload: load })
</script>
