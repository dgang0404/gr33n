<template>
  <div
    class="rounded-lg border px-3 py-2.5 text-sm space-y-2"
    :class="cardClass"
    data-test="guardian-proposal-card"
  >
    <template v-if="uiStatus === 'confirmed'">
      <p class="text-green-300 flex items-start gap-2" data-test="guardian-proposal-done">
        <span class="shrink-0">✓</span>
        <span>{{ uiSummary || 'Action completed.' }}</span>
      </p>
      <RouterLink
        v-if="followUpLink"
        :to="followUpLink"
        class="text-[11px] text-green-400 hover:text-green-300 underline"
        data-test="guardian-proposal-followup"
      >
        {{ followUpLabel }}
      </RouterLink>
    </template>

    <template v-else-if="uiStatus === 'dismissed'">
      <p class="text-zinc-500 text-xs italic" data-test="guardian-proposal-dismissed">Dismissed</p>
    </template>

    <template v-else>
      <div class="flex items-start justify-between gap-2">
        <div class="min-w-0">
          <p class="text-[10px] uppercase tracking-widest text-amber-500/90">Proposed action</p>
          <p class="text-zinc-100 font-medium mt-0.5">{{ local.summary }}</p>
          <p class="text-[10px] text-zinc-500 mt-1">
            {{ toolLabel(local.tool) }}
            <span v-if="targetHint"> · {{ targetHint }}</span>
          </p>
        </div>
      </div>
      <p v-if="uiError" data-test="guardian-proposal-error" class="text-xs text-red-400">
        {{ uiError }}
      </p>
      <div class="flex items-center gap-2 pt-0.5">
        <button
          type="button"
          data-test="guardian-proposal-confirm"
          class="px-3 py-1.5 rounded-lg bg-green-900/60 text-green-200 border border-green-800 hover:bg-green-900/80 text-xs font-medium disabled:opacity-40"
          :disabled="confirming || !canOperate || isExpired"
          :title="confirmTitle"
          @click="onConfirm"
        >
          {{ confirming ? 'Confirming…' : 'Confirm' }}
        </button>
        <button
          type="button"
          data-test="guardian-proposal-dismiss"
          class="px-3 py-1.5 rounded-lg bg-zinc-800 text-zinc-300 hover:bg-zinc-700 text-xs"
          :disabled="confirming"
          @click="onDismiss"
        >
          Dismiss
        </button>
        <span v-if="isExpired" class="text-[10px] text-amber-400">Expired</span>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import api from '../api'

const props = defineProps({
  proposal: { type: Object, required: true },
  canOperate: { type: Boolean, default: true },
})

const emit = defineEmits(['confirmed', 'dismissed', 'error'])

const confirming = ref(false)
const local = reactive({ ...props.proposal })
const uiStatus = ref(props.proposal.status || 'pending')
const uiSummary = ref(props.proposal.confirmSummary || '')
const uiError = ref(props.proposal.error || '')
const uiResult = ref(props.proposal.result ?? null)

watch(
  () => props.proposal,
  (p) => {
    Object.assign(local, p)
    uiStatus.value = p.status || 'pending'
    uiSummary.value = p.confirmSummary || ''
    uiError.value = p.error || ''
    uiResult.value = p.result ?? null
  },
  { deep: true },
)

const TOOL_LABELS = {
  ack_alert: 'Acknowledge alert',
  mark_alert_read: 'Mark alert read',
  create_task_from_alert: 'Create task from alert',
  create_task: 'Create task',
  update_cycle_stage: 'Update crop cycle stage',
  patch_schedule: 'Patch schedule',
  patch_fertigation_program: 'Patch fertigation program',
  patch_rule: 'Patch automation rule',
  apply_bootstrap_template: 'Apply bootstrap template',
}

const isExpired = computed(() => {
  if (!local.expires_at) return false
  return new Date(local.expires_at).getTime() < Date.now()
})

const cardClass = computed(() => {
  if (uiStatus.value === 'confirmed') {
    return 'border-green-800 bg-green-950/40'
  }
  if (uiStatus.value === 'dismissed') {
    return 'border-zinc-800 bg-zinc-950/30 opacity-70'
  }
  return 'border-amber-900/50 bg-amber-950/20'
})

const targetHint = computed(() => {
  const id = local.args?.alert_id
  if (id != null) return `alert #${id}`
  const cycleId = local.args?.crop_cycle_id ?? local.args?.cycle_id
  if (cycleId != null) return `cycle #${cycleId}`
  return ''
})

const followUpLink = computed(() => {
  if ((local.tool === 'create_task_from_alert' || local.tool === 'create_task') && local.result?.task_id) {
    return '/tasks'
  }
  if (local.args?.alert_id) return '/alerts'
  return null
})

const followUpLabel = computed(() => {
  if ((local.tool === 'create_task_from_alert' || local.tool === 'create_task') && local.result?.task_id) {
    return `View task #${local.result.task_id}`
  }
  if (local.args?.alert_id) return 'View alerts'
  return 'View'
})

const confirmTitle = computed(() => {
  if (props.canOperate) return undefined
  return 'Operators only — your role cannot confirm farm actions'
})

function toolLabel(id) {
  return TOOL_LABELS[id] || id
}

async function onConfirm() {
  if (confirming.value || !props.canOperate || isExpired.value) return
  confirming.value = true
  try {
    const r = await api.post('/v1/chat/confirm', { proposal_id: local.proposal_id })
    uiStatus.value = 'confirmed'
    uiSummary.value = r.data?.summary || 'Action completed.'
    uiResult.value = r.data?.result ?? null
    uiError.value = ''
    emit('confirmed', {
      proposal: props.proposal,
      summary: uiSummary.value,
      result: uiResult.value,
    })
  } catch (e) {
    const msg = e?.response?.data?.error || e.message || 'Confirm failed'
    uiError.value = msg
    emit('error', { proposal: props.proposal, error: msg })
  } finally {
    confirming.value = false
  }
}

function onDismiss() {
  if (confirming.value) return
  uiStatus.value = 'dismissed'
  uiError.value = ''
  emit('dismissed', { proposal: props.proposal })
}
</script>
