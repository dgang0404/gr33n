<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center">
        <h1 class="text-xl font-semibold text-white">Automation Rules</h1>
        <HelpTip position="bottom">
          Rules react to sensor state. Every ~15s the worker re-evaluates each
          active rule: it fetches the latest reading for every predicate, runs
          the ALL/ANY logic, honours the cooldown window, and — if satisfied —
          dispatches the listed actions in order. Schedules (cron) live on the
          Schedules page; rules (state-triggered) live here.
        </HelpTip>
      </div>
      <div class="flex items-center gap-3">
        <button class="px-3 py-1.5 text-xs rounded bg-gr33n-600 hover:bg-gr33n-500 text-white font-medium"
          @click="openCreate">+ New Rule</button>
        <button class="text-xs text-zinc-400 hover:text-zinc-200" @click="refreshAll">Refresh</button>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading automation rules…</div>

    <div v-else class="space-y-6">
      <!-- Rule list -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-white text-sm font-semibold mb-3">Rules</h2>
        <p v-if="!rules.length" class="text-zinc-500 text-sm">
          No rules yet. Create one to react to sensor thresholds automatically.
        </p>
        <div v-else class="space-y-3">
          <div v-for="rule in rules" :key="rule.id"
            class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <p class="text-sm text-zinc-200 font-medium truncate">{{ rule.name }}</p>
                  <span class="text-[10px] text-zinc-600 font-mono">#{{ rule.id }}</span>
                </div>
                <p class="text-xs text-zinc-500 mt-1">{{ triggerSummary(rule) }}</p>
                <p v-if="rule.description" class="text-xs text-zinc-600 mt-0.5 italic">{{ rule.description }}</p>
                <div class="flex flex-wrap gap-3 mt-1.5 text-[11px] text-zinc-600">
                  <span>cooldown: {{ rule.cooldown_period_seconds ?? 0 }}s</span>
                  <span>last fired: {{ formatTime(rule.last_triggered_time) }}</span>
                  <span>last evaluated: {{ formatTime(rule.last_evaluated_time) }}</span>
                  <span>{{ actionCount(rule.id) }} action{{ actionCount(rule.id) === 1 ? '' : 's' }}</span>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <button @click="openEdit(rule)"
                  class="px-2 py-1 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
                  title="Edit">&#9998;</button>
                <button @click="confirmDelete(rule)"
                  class="px-2 py-1 text-xs rounded border border-red-900/50 text-red-400 hover:text-red-300"
                  title="Delete">&#128465;</button>
                <button @click="toggleRule(rule)"
                  class="px-2 py-1 text-xs rounded border"
                  :class="rule.is_active ? 'border-green-700 text-green-400' : 'border-zinc-700 text-zinc-400'">
                  {{ rule.is_active ? 'Active' : 'Inactive' }}
                </button>
              </div>
            </div>
            <!-- Inline action preview -->
            <div v-if="ruleActions[rule.id]?.length" class="mt-2 ml-0.5 flex flex-wrap gap-1">
              <span v-for="a in ruleActions[rule.id]" :key="a.id"
                class="text-[10px] px-1.5 py-0.5 rounded bg-zinc-900 border border-zinc-800 text-zinc-400 font-mono">
                {{ a.action_type }}{{ actionHint(a) }}
              </span>
            </div>
            <!-- Cooldown banner -->
            <div v-if="cooldownRemaining(rule)"
              class="mt-2 text-[11px] text-amber-400/90 border border-amber-900/40 bg-amber-900/10 rounded px-2 py-1">
              Cooling down — next evaluation can fire in {{ cooldownRemaining(rule) }}s.
            </div>
          </div>
        </div>
      </div>

      <!-- Automation runs for rules -->
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-white text-sm font-semibold">Rule runs</h2>
          <div class="text-xs" :class="worker.running ? 'text-green-400' : 'text-zinc-500'">
            Worker: {{ worker.running ? 'running' : 'stopped' }}
          </div>
        </div>
        <p v-if="!ruleRuns.length" class="text-zinc-500 text-sm">No rule runs yet.</p>
        <div v-else class="space-y-2">
          <div v-for="r in ruleRuns" :key="r.id"
            class="bg-zinc-950 border border-zinc-800 rounded-lg p-3">
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-2">
                <span class="text-xs px-2 py-0.5 rounded" :class="statusClass(r.status)">{{ r.status }}</span>
                <span class="text-[10px] text-zinc-600">
                  rule #{{ r.rule_id }} · {{ ruleName(r.rule_id) }}
                </span>
              </div>
              <span class="text-xs text-zinc-600">{{ formatTime(r.executed_at) }}</span>
            </div>
            <p v-if="r.message" class="text-zinc-300 text-xs mt-2">{{ r.message }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      @click.self="showModal = false">
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <h3 class="text-white font-semibold mb-4">{{ editingRule ? 'Edit Rule' : 'New Rule' }}</h3>
        <RuleForm
          :rule="editingRule"
          :actions="editingRule ? (ruleActions[editingRule.id] || []) : []"
          :sensors="sensors"
          :actuators="actuators"
          :zones="zones"
          :saving="saving"
          :errorMessage="formError"
          @submit="onSubmit"
          @cancel="showModal = false"
        />
      </div>
    </div>

    <!-- Delete confirmation -->
    <div v-if="deleteTarget" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="deleteTarget = null">
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-sm space-y-4">
        <h3 class="text-white font-semibold">Delete Rule</h3>
        <p class="text-sm text-zinc-300">
          Delete <span class="text-white font-medium">{{ deleteTarget.name }}</span>? This also
          removes its executable actions (cascade).
        </p>
        <div class="flex justify-end gap-3 pt-2">
          <button @click="deleteTarget = null"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200">
            Cancel
          </button>
          <button @click="doDelete" :disabled="saving"
            class="px-3 py-1.5 text-xs rounded bg-red-600 hover:bg-red-500 text-white font-medium disabled:opacity-50">
            {{ saving ? 'Deleting…' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'
import RuleForm from '../components/RuleForm.vue'
import api from '../api'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loading = ref(false)
const saving = ref(false)
const formError = ref('')
const showModal = ref(false)
const editingRule = ref(null)
const deleteTarget = ref(null)

const rules = ref([])
const ruleActions = ref({}) // { [ruleId]: ExecutableAction[] }
const runs = ref([])
const sensors = ref([])
const actuators = ref([])
const zones = ref([])
const worker = ref({ running: false, simulation_mode: false })

const ruleRuns = computed(() => runs.value.filter(r => r.rule_id != null))

function ruleName(id) {
  return rules.value.find(r => r.id === id)?.name || ''
}
function actionCount(id) {
  return (ruleActions.value[id] || []).length
}
function sensorName(id) {
  return sensors.value.find(s => s.id === id)?.name || `sensor #${id}`
}
function actuatorName(id) {
  return actuators.value.find(a => a.id === id)?.name || `actuator #${id}`
}

const OP_LABEL = { lt: '<', lte: '≤', eq: '=', gte: '≥', gt: '>', ne: '≠' }

function triggerSummary(rule) {
  const src = rule.trigger_source
  const cfg = rule.trigger_configuration || {}
  const conds = rule.conditions_jsonb && Array.isArray(rule.conditions_jsonb.predicates)
    ? rule.conditions_jsonb.predicates
    : []
  const parts = []
  if (src === 'sensor_reading_threshold' && cfg.sensor_id) {
    parts.push(`on reading from ${sensorName(cfg.sensor_id)}`)
  } else {
    parts.push(`trigger: ${src}`)
  }
  if (conds.length) {
    const joiner = rule.condition_logic === 'ANY' ? ' OR ' : ' AND '
    const rendered = conds.map(p => `${sensorName(p.sensor_id)} ${OP_LABEL[p.op] || p.op} ${p.value}`).join(joiner)
    parts.push(`when ${rendered}`)
  }
  return parts.join(' · ')
}

function actionHint(a) {
  if (a.action_type === 'control_actuator' && a.target_actuator_id) {
    return ` → ${actuatorName(a.target_actuator_id)}${a.action_command ? ' (' + a.action_command + ')' : ''}`
  }
  if (a.action_type === 'send_notification' && a.target_notification_template_id) {
    return ` → template #${a.target_notification_template_id}`
  }
  if (a.action_type === 'create_task' && a.action_parameters?.title) {
    return ` → "${a.action_parameters.title}"`
  }
  return ''
}

function cooldownRemaining(rule) {
  if (!rule.last_triggered_time || !rule.cooldown_period_seconds) return 0
  const last = new Date(rule.last_triggered_time).getTime()
  const nextReady = last + rule.cooldown_period_seconds * 1000
  const remainingMs = nextReady - Date.now()
  if (remainingMs <= 0) return 0
  return Math.ceil(remainingMs / 1000)
}

function openCreate() {
  editingRule.value = null
  formError.value = ''
  showModal.value = true
}
function openEdit(rule) {
  editingRule.value = rule
  formError.value = ''
  showModal.value = true
}

function confirmDelete(rule) { deleteTarget.value = rule }

async function doDelete() {
  saving.value = true
  try {
    await store.deleteAutomationRule(deleteTarget.value.id)
    rules.value = rules.value.filter(r => r.id !== deleteTarget.value.id)
    delete ruleActions.value[deleteTarget.value.id]
    deleteTarget.value = null
  } catch (e) {
    formError.value = e?.response?.data?.error || 'Failed to delete rule.'
  } finally {
    saving.value = false
  }
}

async function toggleRule(rule) {
  const updated = await store.updateAutomationRuleActive(rule.id, !rule.is_active)
  const idx = rules.value.findIndex(r => r.id === rule.id)
  if (idx >= 0) rules.value[idx] = updated
}

// Diff-apply a submitted action list against the current server state.
async function syncActions(ruleId, desired) {
  const existing = ruleActions.value[ruleId] || []
  const desiredIds = new Set(desired.filter(a => a.id != null).map(a => a.id))
  for (const prev of existing) {
    if (!desiredIds.has(prev.id)) {
      await store.deleteRuleAction(prev.id)
    }
  }
  const next = []
  for (const a of desired) {
    const payload = {
      execution_order: a.execution_order,
      action_type: a.action_type,
      target_actuator_id: a.target_actuator_id,
      target_notification_template_id: a.target_notification_template_id,
      action_command: a.action_command,
      action_parameters: a.action_parameters,
      delay_before_execution_seconds: a.delay_before_execution_seconds,
    }
    if (a.id != null) {
      next.push(await store.updateRuleAction(a.id, payload))
    } else {
      next.push(await store.createRuleAction(ruleId, payload))
    }
  }
  ruleActions.value[ruleId] = next
}

async function onSubmit(result) {
  if (result.error) {
    formError.value = result.error
    return
  }
  formError.value = ''
  saving.value = true
  try {
    if (editingRule.value) {
      const updated = await store.updateAutomationRule(editingRule.value.id, result.rule)
      const idx = rules.value.findIndex(r => r.id === updated.id)
      if (idx >= 0) rules.value[idx] = updated
      await syncActions(updated.id, result.actions)
    } else {
      const created = await store.createAutomationRule(farmContext.farmId, result.rule)
      rules.value = [...rules.value, created]
      await syncActions(created.id, result.actions)
    }
    showModal.value = false
  } catch (e) {
    formError.value = e?.response?.data?.error || 'Failed to save rule.'
  } finally {
    saving.value = false
  }
}

async function refreshAll() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    if (!store.zones.length || !store.sensors.length || !store.actuators.length) {
      await store.loadAll(fid)
    }
    const [rs, rr, w] = await Promise.all([
      store.loadAutomationRules(fid),
      store.loadAutomationRuns(fid),
      api.get('/automation/worker/health'),
    ])
    rules.value = rs
    runs.value = rr
    worker.value = w.data || { running: false, simulation_mode: false }
    sensors.value = Array.isArray(store.sensors) ? store.sensors : []
    actuators.value = Array.isArray(store.actuators) ? store.actuators : []
    zones.value = Array.isArray(store.zones) ? store.zones : []
    const actionLists = await Promise.all(rs.map(r => store.loadRuleActions(r.id)))
    const next = {}
    rs.forEach((r, i) => { next[r.id] = actionLists[i] })
    ruleActions.value = next
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try { await refreshAll() }
  finally { loading.value = false }
})

function formatTime(t) {
  if (!t) return 'never'
  // pgtype.Timestamptz marshals as { Time, InfinityModifier, Valid }. If the
  // wrapper object leaks through, unwrap it defensively so the card still
  // renders. On newer builds it arrives as a plain ISO string.
  const iso = typeof t === 'string' ? t : (t?.Time || t?.time || '')
  if (!iso) return 'never'
  return new Date(iso).toLocaleString(undefined, {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit'
  })
}

const STATUS_MAP = {
  success: 'bg-green-900/50 text-green-300',
  partial_success: 'bg-yellow-900/50 text-yellow-300',
  failed: 'bg-red-900/50 text-red-300',
  skipped: 'bg-zinc-800 text-zinc-300',
}
function statusClass(status) { return STATUS_MAP[status] ?? STATUS_MAP.skipped }
</script>
