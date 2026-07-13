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
        @refine="onRefine"
      />
    </div>
  </div>
</template>

<script setup>
import { computed, watch } from 'vue'
import GuardianActionProposal from './GuardianActionProposal.vue'
import { useFarmOperate } from '../composables/useFarmOperate'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { useGuardianPanelStore } from '../stores/guardianPanel'

const props = defineProps({
  /** When true, reload pending proposals (drawer or full-page Pending tab). */
  active: { type: Boolean, default: true },
})

const farmContext = useFarmContextStore()
const store = useGuardianProposalsStore()
const guardianPanel = useGuardianPanelStore()

const emit = defineEmits(['refine'])

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
    session_id: p.session_id,
    revision: p.revision,
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

function onRefine({ proposal }) {
  guardianPanel.requestRefine(proposal)
  emit('refine', { proposal })
}

watch(farmId, () => {
  if (props.active) load()
}, { immediate: true })

watch(
  () => props.active,
  (isActive) => {
    if (isActive) load()
  },
)

defineExpose({ reload: load })
</script>
