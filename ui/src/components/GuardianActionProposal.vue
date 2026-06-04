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

    <template v-else-if="uiStatus === 'superseded'">
      <p class="text-zinc-500 text-xs italic" data-test="guardian-proposal-superseded">
        Superseded{{ local.revision ? ` — replaced by a newer revision` : '' }}
      </p>
    </template>

    <template v-else>
      <div class="flex items-start justify-between gap-2">
        <div class="min-w-0">
          <div class="flex items-center gap-2 flex-wrap">
            <p class="text-[10px] uppercase tracking-widest text-amber-500/90">Proposed action</p>
            <span
              v-if="riskTier"
              class="text-[9px] uppercase tracking-wider px-1.5 py-0.5 rounded border"
              :class="riskBadgeClass"
              data-test="guardian-proposal-risk-badge"
            >
              {{ riskTier }}
            </span>
            <span
              v-if="revisionLabelText"
              class="text-[9px] uppercase tracking-wider px-1.5 py-0.5 rounded border border-violet-700 text-violet-300 bg-violet-950/40"
              data-test="guardian-proposal-revision-badge"
            >
              {{ revisionLabelText }}
            </span>
          </div>
          <p class="text-zinc-100 font-medium mt-0.5">{{ local.summary }}</p>
          <p class="text-[10px] text-zinc-500 mt-1">
            {{ toolLabel(local.tool) }}
            <span v-if="targetHint"> · {{ targetHint }}</span>
          </p>
        </div>
      </div>

      <p
        v-if="isHighRisk"
        class="text-xs text-red-300/95 bg-red-950/40 border border-red-900/60 rounded-md px-2.5 py-2"
        data-test="guardian-proposal-high-warning"
      >
        {{ highRiskWarning }}
      </p>

      <SetupPackProposalCard
        v-if="isSetupPack"
        :args="local.args"
      />

      <p
        v-else-if="isMediumRisk && diffSummary"
        class="text-[11px] text-sky-200/90 bg-sky-950/30 border border-sky-900/40 rounded-md px-2.5 py-1.5 font-mono"
        data-test="guardian-proposal-diff"
      >
        {{ diffSummary }}
      </p>

      <div
        v-if="impactLines.length && !isSetupPack"
        class="text-[11px] text-zinc-200 bg-zinc-900/50 border border-zinc-800 rounded-md px-2.5 py-1.5 space-y-0.5"
        data-test="guardian-proposal-impact"
      >
        <p class="text-[10px] uppercase tracking-wider text-zinc-400">If you Confirm, this will…</p>
        <ul class="space-y-0.5">
          <li v-for="(line, i) in impactLines" :key="i" class="flex gap-1.5">
            <span class="text-zinc-500 shrink-0">·</span>
            <span>{{ line }}</span>
          </li>
        </ul>
      </div>

      <div
        v-if="operatorFacts.length"
        class="text-[11px] text-emerald-200/90 bg-emerald-950/20 border border-emerald-900/40 rounded-md px-2.5 py-1.5 space-y-0.5"
        data-test="guardian-proposal-operator-facts"
      >
        <p class="text-[10px] uppercase tracking-wider text-emerald-400/80">Operator-stated (not measured)</p>
        <ul class="space-y-0.5">
          <li v-for="(f, i) in operatorFacts" :key="i">{{ f.label || `${f.field}: ${f.value}` }}</li>
        </ul>
      </div>

      <div
        v-if="revisionDiff.length"
        class="text-[11px] text-violet-200/90 bg-violet-950/20 border border-violet-900/40 rounded-md px-2.5 py-1.5 font-mono"
        data-test="guardian-proposal-revision-diff"
      >
        <p class="font-sans text-[10px] uppercase tracking-wider text-violet-400/80 mb-0.5">Changed from previous revision</p>
        <p v-for="(d, i) in revisionDiff" :key="i">{{ d }}</p>
      </div>

      <p v-if="uiError" data-test="guardian-proposal-error" class="text-xs text-red-400">
        {{ uiError }}
      </p>
      <div class="flex items-center gap-2 pt-0.5">
        <button
          type="button"
          data-test="guardian-proposal-confirm"
          class="px-3 py-1.5 rounded-lg text-xs font-medium disabled:opacity-40"
          :class="confirmButtonClass"
          :disabled="confirming || !canOperate || isExpired"
          :title="confirmTitle"
          @click="onConfirm"
        >
          {{ confirming ? 'Confirming…' : 'Confirm' }}
        </button>
        <button
          type="button"
          data-test="guardian-proposal-refine"
          class="px-3 py-1.5 rounded-lg bg-violet-900/50 text-violet-200 border border-violet-800 hover:bg-violet-900/70 text-xs disabled:opacity-40"
          :disabled="confirming || !canOperate || isExpired"
          :title="confirmTitle"
          @click="onRefine"
        >
          Refine
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
import SetupPackProposalCard from './SetupPackProposalCard.vue'
import { SETUP_PACK_HIGH_RISK_COPY } from '../lib/guardianSetupPack.js'
import { impactForProposal, revisionLabel } from '../lib/guardianImpact.js'

const props = defineProps({
  proposal: { type: Object, required: true },
  canOperate: { type: Boolean, default: true },
})

const emit = defineEmits(['confirmed', 'dismissed', 'error', 'refine'])

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
  enqueue_actuator_command: 'Queue Pi actuator command',
  create_plant: 'Create plant',
  create_crop_cycle: 'Start crop cycle',
  create_fertigation_program: 'Create fertigation program',
  create_lighting_program: 'Create lighting program',
  summarize_zone_lighting: 'Summarize zone lighting',
  apply_grow_setup_pack: 'Apply grow setup pack',
}

const riskTier = computed(() => (local.risk_tier || 'medium').toLowerCase())
const isSetupPack = computed(() => local.tool === 'apply_grow_setup_pack')
const isHighRisk = computed(() => riskTier.value === 'high')
const isMediumRisk = computed(() => riskTier.value === 'medium')

const highRiskWarning = computed(() => {
  if (isSetupPack.value) return SETUP_PACK_HIGH_RISK_COPY
  return 'High-impact change — review frozen args carefully before Confirm. This can alter farm configuration, disable automation, or apply a bootstrap template.'
})

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
  if (isHighRisk.value) {
    return 'border-red-800/70 bg-red-950/25'
  }
  if (isMediumRisk.value) {
    return 'border-sky-900/50 bg-sky-950/15'
  }
  return 'border-amber-900/50 bg-amber-950/20'
})

const riskBadgeClass = computed(() => {
  if (isHighRisk.value) return 'border-red-700 text-red-300 bg-red-950/50'
  if (isMediumRisk.value) return 'border-sky-800 text-sky-300 bg-sky-950/40'
  return 'border-zinc-700 text-zinc-400 bg-zinc-900/50'
})

const confirmButtonClass = computed(() => {
  if (isHighRisk.value) {
    return 'bg-red-900/70 text-red-100 border border-red-700 hover:bg-red-900/90'
  }
  return 'bg-green-900/60 text-green-200 border border-green-800 hover:bg-green-900/80'
})

const diffSummary = computed(() => formatDiffSummary(local.tool, local.args))

const revisionLabelText = computed(() => revisionLabel(local))
const impact = computed(() => impactForProposal(local))
const impactLines = computed(() => impact.value.lines)
const operatorFacts = computed(() => impact.value.facts)
const revisionDiff = computed(() => computeArgsDiff(local.previous_args, local.args))

const targetHint = computed(() => {
  if (isSetupPack.value) {
    const zone = local.args?.zone_name
    if (zone) return String(zone)
    if (local.args?.zone_id != null) return `zone #${local.args.zone_id}`
  }
  const id = local.args?.alert_id
  if (id != null) return `alert #${id}`
  const cycleId = local.args?.crop_cycle_id ?? local.args?.cycle_id
  if (cycleId != null) return `cycle #${cycleId}`
  const scheduleId = local.args?.schedule_id
  if (scheduleId != null) return `schedule #${scheduleId}`
  const programId = local.args?.program_id
  if (programId != null) return `program #${programId}`
  const ruleId = local.args?.rule_id
  if (ruleId != null) return `rule #${ruleId}`
  const deviceId = local.args?.device_id
  if (deviceId != null) {
    const actId = local.args?.actuator_id
    if (actId != null) return `device #${deviceId} · actuator #${actId}`
    return `device #${deviceId}`
  }
  return ''
})

const followUpLink = computed(() => {
  if (local.tool === 'apply_grow_setup_pack') {
    return '/plants'
  }
  if ((local.tool === 'create_task_from_alert' || local.tool === 'create_task') && local.result?.task_id) {
    return '/tasks'
  }
  if (local.args?.alert_id) return '/alerts'
  return null
})

const followUpLabel = computed(() => {
  if (local.tool === 'apply_grow_setup_pack') {
    return 'View plants'
  }
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

function formatDiffSummary(tool, args) {
  if (!args || typeof args !== 'object') return ''
  const parts = []
  for (const [key, val] of Object.entries(args)) {
    if (val == null || val === '') continue
    let display = val
    if (typeof val === 'object') display = JSON.stringify(val)
    parts.push(`${key}: ${display}`)
  }
  if (!parts.length) return ''
  return parts.join(' · ')
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

function onRefine() {
  if (confirming.value || !props.canOperate || isExpired.value) return
  emit('refine', { proposal: props.proposal })
}

// computeArgsDiff renders "path: old → new" lines between two frozen arg objects,
// recursing one level into nested sections (setup-pack plant/cycle/program).
function computeArgsDiff(prev, next) {
  if (!prev || !next || typeof prev !== 'object' || typeof next !== 'object') return []
  const out = []
  const walk = (a, b, prefix) => {
    const keys = new Set([...Object.keys(a || {}), ...Object.keys(b || {})])
    for (const k of keys) {
      const av = a ? a[k] : undefined
      const bv = b ? b[k] : undefined
      const path = prefix ? `${prefix}.${k}` : k
      const aObj = av && typeof av === 'object'
      const bObj = bv && typeof bv === 'object'
      if (aObj || bObj) {
        walk(aObj ? av : {}, bObj ? bv : {}, path)
        continue
      }
      if (av !== bv) {
        out.push(`${path}: ${av ?? '∅'} → ${bv ?? '∅'}`)
      }
    }
  }
  walk(prev, next, '')
  return out
}
</script>
