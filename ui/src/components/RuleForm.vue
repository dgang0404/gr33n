<template>
  <div class="space-y-6">
    <!-- Rule header fields -->
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label class="text-xs text-zinc-400 block mb-1">Name</label>
        <input v-model="form.name"
          class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white"
          placeholder="EC too low → turn pump on" />
      </div>
      <div>
        <label class="text-xs text-zinc-400 block mb-1 flex items-center">
          Cooldown (seconds)
          <HelpTip position="top">
            Minimum seconds between successive fires. While cooling down the
            worker records a <span class="font-mono">skipped</span> run with
            <span class="font-mono">message=cooldown</span>.
          </HelpTip>
        </label>
        <input type="number" min="0" v-model.number="form.cooldown_period_seconds"
          class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white"
          placeholder="0" />
      </div>
    </div>

    <div>
      <label class="text-xs text-zinc-400 block mb-1">Description</label>
      <input v-model="form.description"
        class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-1.5 text-sm text-white"
        placeholder="What this rule does, in plain language" />
    </div>

    <div class="flex items-center gap-2">
      <input type="checkbox" v-model="form.is_active" id="rule-active"
        class="rounded border-zinc-600" />
      <label for="rule-active" class="text-xs text-zinc-300">Active</label>
    </div>

    <!-- Pane 1: Trigger -->
    <section class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 space-y-3">
      <div class="flex items-center">
        <h3 class="text-sm font-semibold text-white">1 · Trigger</h3>
        <HelpTip position="bottom">
          <p class="mb-1">When should this rule wake up and evaluate?</p>
          <p class="text-zinc-400">
            <span class="font-mono">sensor_reading_threshold</span> fires on every new
            reading for the target sensor. <span class="font-mono">manual_api_trigger</span>
            only runs when invoked explicitly. Other sources are reserved for future phases.
          </p>
        </HelpTip>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-zinc-400 block mb-1">Source</label>
          <select v-model="form.trigger_source"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white">
            <option v-for="src in triggerSources" :key="src" :value="src">{{ src }}</option>
          </select>
        </div>
        <div v-if="form.trigger_source === 'sensor_reading_threshold'">
          <label class="text-xs text-zinc-400 block mb-1">Trigger sensor</label>
          <select v-model.number="triggerSensorId"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white">
            <option :value="0" disabled>Select sensor</option>
            <option v-for="s in sensors" :key="s.id" :value="s.id">
              {{ s.name }}{{ s.sensor_type ? ' (' + s.sensor_type + ')' : '' }}
            </option>
          </select>
        </div>
        <div v-if="form.trigger_source === 'animal_lifecycle_event'">
          <label class="text-xs text-zinc-400 block mb-1">Flock / group</label>
          <select v-model.number="triggerAnimalGroupId"
            class="w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-sm text-white">
            <option :value="0" disabled>Select animal group</option>
            <option v-for="g in animalGroups" :key="g.id" :value="g.id">{{ g.label }}</option>
          </select>
        </div>
      </div>
      <p v-if="form.trigger_source === 'sensor_reading_threshold'" class="text-[11px] text-zinc-600">
        The trigger just wakes the evaluator. The conditions below decide whether the rule actually fires.
      </p>
      <p v-if="form.trigger_source === 'animal_lifecycle_event'" class="text-[11px] text-zinc-600">
        Add an <span class="font-mono">animal_event</span> condition below for "this flock's latest
        event is X" — e.g. released_to_pasture opens a gate, penned_for_night closes it.
      </p>
    </section>

    <!-- Pane 2: Conditions -->
    <section class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <h3 class="text-sm font-semibold text-white">2 · Conditions</h3>
          <HelpTip position="bottom">
            Predicates evaluated against each sensor's latest reading.
            <span class="font-mono">ALL</span> = every predicate must match;
            <span class="font-mono">ANY</span> = one match is enough.
            An empty list = the rule fires on every trigger.
          </HelpTip>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-xs text-zinc-500">Logic</label>
          <select v-model="form.condition_logic"
            class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
            <option value="ALL">ALL</option>
            <option value="ANY">ANY</option>
          </select>
          <button type="button" @click="addCondition"
            class="text-[11px] text-zinc-300 border border-zinc-700 rounded px-2 py-1 hover:text-white">
            + Add predicate
          </button>
        </div>
      </div>
      <p v-if="!form.conditions.length" class="text-[11px] text-zinc-600 italic">
        No conditions — the rule fires on every trigger wake-up.
      </p>
      <div v-else class="space-y-2">
        <div v-for="(p, idx) in form.conditions" :key="idx"
          class="bg-zinc-900 border border-zinc-800 rounded p-2 space-y-1">
          <div class="flex items-center gap-2 text-[11px] text-zinc-400">
            <label class="text-zinc-500">Type</label>
            <select :value="p.type || 'hard'" @change="setPredicateType(p, $event.target.value)"
              class="bg-zinc-900 border border-zinc-700 rounded px-1.5 py-1 text-xs text-white">
              <option value="hard">sensor threshold</option>
              <option value="setpoint">setpoint (zone/cycle)</option>
              <option value="animal_event">animal lifecycle event</option>
            </select>
            <HelpTip v-if="p.type === 'setpoint'" position="right">
              When checked, the rule reads min/ideal/max from
              <code>gr33ncore.zone_setpoints</code> at every tick, using the
              trigger's zone + active crop cycle's <code>current_stage</code>.
              Configure setpoints under Operate → Setpoints.
              If no setpoint exists for the resolved scope, the rule skips with
              <code>no_setpoint_for_scope</code> — not a failure, just "nothing to
              compare against yet."
            </HelpTip>
            <HelpTip v-else-if="p.type === 'animal_event'" position="right">
              Passes iff the animal group's most recent lifecycle event
              (<code>gr33nanimals.animal_lifecycle_events</code>) matches
              <code>event_type</code> exactly. No lifecycle event yet skips with
              <code>no_animal_event_yet</code> — not a failure.
            </HelpTip>
            <div class="flex-1"></div>
            <button type="button" @click="removeCondition(idx)"
              class="text-[11px] text-red-400 hover:text-red-300 px-1">Remove</button>
          </div>
          <!-- Hard predicate fields -->
          <div v-if="!p.type || p.type === 'hard'"
            class="grid grid-cols-[minmax(0,1fr)_80px_100px] gap-2 items-center">
            <select v-model.number="p.sensor_id"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option :value="0" disabled>Select sensor</option>
              <option v-for="s in sensors" :key="s.id" :value="s.id">
                {{ s.name }}{{ s.sensor_type ? ' (' + s.sensor_type + ')' : '' }}
              </option>
            </select>
            <select v-model="p.op"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option value="lt">&lt;</option>
              <option value="lte">&le;</option>
              <option value="eq">=</option>
              <option value="gte">&ge;</option>
              <option value="gt">&gt;</option>
              <option value="ne">&ne;</option>
            </select>
            <input type="number" step="any" v-model.number="p.value"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white" />
          </div>
          <!-- Setpoint predicate fields -->
          <div v-else-if="p.type === 'setpoint'" class="grid grid-cols-[minmax(0,1fr)_140px_140px] gap-2 items-center">
            <input v-model="p.sensor_type" type="text" list="sensor-type-list"
              placeholder="sensor_type (dew_point, vpd, …)"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white" />
            <datalist id="sensor-type-list">
              <option v-for="t in uniqueSensorTypes" :key="t" :value="t" />
            </datalist>
            <select v-model="p.scope"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option value="current_stage">scope: current_stage</option>
              <option value="zone_default">scope: zone_default</option>
            </select>
            <select v-model="p.op"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option value="out_of_range">out_of_range</option>
              <option value="below_ideal">below_ideal</option>
              <option value="above_ideal">above_ideal</option>
              <option value="inside_range">inside_range</option>
            </select>
          </div>
          <!-- Animal-event predicate fields -->
          <div v-else class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-2 items-center">
            <select v-model.number="p.animal_group_id"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option :value="0" disabled>Select animal group</option>
              <option v-for="g in animalGroups" :key="g.id" :value="g.id">{{ g.label }}</option>
            </select>
            <input v-model="p.event_type" type="text" list="animal-event-type-list"
              placeholder="event_type (released_to_pasture, penned_for_night, …)"
              class="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white" />
            <datalist id="animal-event-type-list">
              <option v-for="t in commonAnimalEventTypes" :key="t" :value="t" />
            </datalist>
          </div>
        </div>
      </div>
    </section>

    <!-- Pane 3: Actions -->
    <section class="bg-zinc-950 border border-zinc-800 rounded-lg p-4 space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center">
          <h3 class="text-sm font-semibold text-white">3 · Actions</h3>
          <HelpTip position="bottom">
            Dispatched in the listed order. Supported types:
            <span class="font-mono">control_actuator</span> toggles a relay,
            <span class="font-mono">create_task</span> spawns a todo,
            <span class="font-mono">send_notification</span> renders a template
            and fans it out as a push alert.
          </HelpTip>
        </div>
        <button type="button" @click="addAction"
          class="text-[11px] text-zinc-300 border border-zinc-700 rounded px-2 py-1 hover:text-white">
          + Add action
        </button>
      </div>
      <p v-if="!form.actions.length" class="text-[11px] text-zinc-600 italic">
        No actions — this rule will evaluate and record runs but won't change anything.
      </p>
      <div v-else class="space-y-3">
        <div v-for="(a, idx) in form.actions" :key="a._key"
          class="bg-zinc-900 border border-zinc-800 rounded-lg p-3 space-y-2">
          <div class="flex items-center gap-2">
            <span class="text-[10px] text-zinc-500 font-mono shrink-0">#{{ idx + 1 }}</span>
            <select v-model="a.action_type" @change="onActionTypeChange(a)"
              class="bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
              <option value="control_actuator">control_actuator</option>
              <option value="create_task">create_task</option>
              <option value="send_notification">send_notification</option>
            </select>
            <input type="number" min="0" v-model.number="a.delay_before_execution_seconds"
              class="w-24 bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
              placeholder="delay (s)" />
            <div class="flex-1"></div>
            <button v-if="idx > 0" type="button" @click="moveAction(idx, -1)"
              class="text-[11px] text-zinc-500 hover:text-zinc-200 px-1" title="Move up">↑</button>
            <button v-if="idx < form.actions.length - 1" type="button" @click="moveAction(idx, 1)"
              class="text-[11px] text-zinc-500 hover:text-zinc-200 px-1" title="Move down">↓</button>
            <button type="button" @click="removeAction(idx)"
              class="text-[11px] text-red-400 hover:text-red-300 px-1">Remove</button>
          </div>

          <!-- control_actuator form -->
          <div v-if="a.action_type === 'control_actuator'"
            class="grid grid-cols-2 gap-2">
            <div>
              <label class="text-[10px] text-zinc-500 block mb-0.5">Target actuator</label>
              <select v-model.number="a.target_actuator_id"
                class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
                <option :value="null" disabled>Select actuator</option>
                <option v-for="act in actuators" :key="act.id" :value="act.id">
                  {{ act.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="text-[10px] text-zinc-500 block mb-0.5">Command</label>
              <input v-model="a.action_command"
                class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white font-mono"
                placeholder="on | off | set:50" />
            </div>
            <div v-if="actuatorSupportsPulse(a.target_actuator_id)" class="col-span-2 flex items-center gap-2">
              <label class="text-[10px] text-zinc-500 shrink-0">Run for (seconds, optional)</label>
              <input type="number" min="1" max="3600" v-model.number="a._params.duration_seconds"
                class="w-24 bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                placeholder="e.g. 5" />
              <span class="text-[10px] text-zinc-600">
                Timed feed/water pulse — device auto-reverts after this many seconds. Leave blank for a plain on/off.
              </span>
            </div>
          </div>

          <!-- create_task form -->
          <div v-else-if="a.action_type === 'create_task'" class="space-y-2">
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="text-[10px] text-zinc-500 block mb-0.5">Title</label>
                <input v-model="a._params.title"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                  placeholder="Check EC reservoir" />
              </div>
              <div>
                <label class="text-[10px] text-zinc-500 block mb-0.5">Zone</label>
                <select v-model.number="a._params.zone_id"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white">
                  <option :value="null">(none)</option>
                  <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
                </select>
              </div>
            </div>
            <div>
              <label class="text-[10px] text-zinc-500 block mb-0.5">Description</label>
              <input v-model="a._params.description"
                class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                placeholder="optional" />
            </div>
            <div class="grid grid-cols-3 gap-2">
              <div>
                <label class="text-[10px] text-zinc-500 block mb-0.5">Task type</label>
                <input v-model="a._params.task_type"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                  placeholder="e.g. inspection" />
              </div>
              <div>
                <label class="text-[10px] text-zinc-500 block mb-0.5">Priority (0-3)</label>
                <input type="number" min="0" max="3" v-model.number="a._params.priority"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white" />
              </div>
              <div>
                <label class="text-[10px] text-zinc-500 block mb-0.5">Due in (days)</label>
                <input type="number" min="0" v-model.number="a._params.due_in_days"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white" />
              </div>
            </div>
          </div>

          <!-- send_notification form -->
          <div v-else-if="a.action_type === 'send_notification'" class="space-y-2">
            <div class="flex items-start gap-2">
              <div class="flex-1">
                <label class="text-[10px] text-zinc-500 block mb-0.5 flex items-center">
                  Notification template
                  <HelpTip position="top">
                    Pick a farm notification template. Subject and body render with rule context and optional variables below.
                  </HelpTip>
                </label>
                <select
                  v-model.number="a.target_notification_template_id"
                  class="w-full bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                >
                  <option :value="null">— Select template —</option>
                  <option v-for="t in notificationTemplates" :key="t.id" :value="t.id">
                    {{ t.template_key }}{{ t.description ? ` — ${t.description}` : '' }}
                  </option>
                </select>
                <p v-if="!notificationTemplates.length && farmId" class="text-[10px] text-zinc-600 mt-1">
                  No templates yet — create one via API or Settings (coming soon).
                </p>
              </div>
            </div>
            <div>
              <div class="flex items-center justify-between">
                <label class="text-[10px] text-zinc-500 block mb-0.5">Variables (optional)</label>
                <button type="button" @click="addVariable(a)"
                  class="text-[10px] text-zinc-400 hover:text-zinc-200">+ Add</button>
              </div>
              <div v-for="(kv, vi) in a._variables" :key="vi"
                class="grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)_auto] gap-1 items-center mt-1">
                <input v-model="kv.key"
                  class="bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                  placeholder="key" />
                <input v-model="kv.value"
                  class="bg-zinc-950 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
                  placeholder="value" />
                <button type="button" @click="removeVariable(a, vi)"
                  class="text-[10px] text-red-400 hover:text-red-300 px-1">×</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div v-if="errorMessage" class="text-red-400 text-xs">{{ errorMessage }}</div>

    <div class="flex justify-end gap-3 pt-1">
      <button type="button" @click="$emit('cancel')"
        class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200">
        Cancel
      </button>
      <button type="button" :disabled="saving" @click="submit"
        class="px-3 py-1.5 text-xs rounded bg-gr33n-600 hover:bg-gr33n-500 text-white font-medium disabled:opacity-50">
        {{ saving ? 'Saving…' : (editing ? 'Update rule' : 'Create rule') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted } from 'vue'
import HelpTip from './HelpTip.vue'
import api from '../api'
import { supportsPulseCommand } from '../lib/plantNeeds.js'

const props = defineProps({
  rule: { type: Object, default: null },
  actions: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  zones: { type: Array, default: () => [] },
  animalGroups: { type: Array, default: () => [] },
  farmId: { type: [Number, String], default: null },
  saving: { type: Boolean, default: false },
  errorMessage: { type: String, default: '' },
})

const emit = defineEmits(['submit', 'cancel'])

const notificationTemplates = ref([])

async function loadNotificationTemplates() {
  const fid = props.farmId
  if (!fid) {
    notificationTemplates.value = []
    return
  }
  try {
    const r = await api.get(`/farms/${fid}/notification-templates`)
    notificationTemplates.value = Array.isArray(r.data) ? r.data : []
  } catch {
    notificationTemplates.value = []
  }
}

onMounted(loadNotificationTemplates)
watch(() => props.farmId, loadNotificationTemplates)

const editing = computed(() => !!props.rule?.id)

const triggerSources = [
  'sensor_reading_threshold',
  'specific_time_cron',
  'actuator_state_changed',
  'manual_api_trigger',
  'task_status_updated',
  'new_system_log_event',
  'external_webhook_received',
  'animal_lifecycle_event',
]

// Phase 210 — free-text examples for the animal_event predicate's event_type
// field. gr33nanimals.animal_lifecycle_events.event_type has no DB enum
// (vocabulary settles from real use), so this is a hint, not a whitelist.
const commonAnimalEventTypes = [
  'added', 'born', 'died', 'note',
  'released_to_pasture', 'penned_for_night', 'moved_zone',
]

let nextKey = 1
function newKey() { return `a-${nextKey++}` }

function emptyForm() {
  return {
    name: '',
    description: '',
    is_active: true,
    trigger_source: 'sensor_reading_threshold',
    cooldown_period_seconds: 0,
    condition_logic: 'ALL',
    conditions: [],
    actions: [],
  }
}

const form = ref(emptyForm())
const triggerSensorId = ref(0)
const triggerAnimalGroupId = ref(0)

// Phase 210 — mirrors internal/handler/actuator.PulseDurationAllowed via the
// UI's device taxonomy fallback (ui/src/lib/plantNeeds.js). Looking the
// actuator up by id (rather than trusting a stale actuator_type on the
// action row) keeps this correct if the target is switched mid-edit.
function actuatorSupportsPulse(actuatorId) {
  const act = props.actuators.find(a => a.id === actuatorId)
  return act ? supportsPulseCommand(act.actuator_type) : false
}

function buildActionFormState(a) {
  const ap = a.action_parameters || {}
  const params = typeof ap === 'object' && ap !== null ? ap : {}
  return {
    _key: newKey(),
    id: a.id ?? null,
    action_type: a.action_type,
    execution_order: a.execution_order ?? 0,
    target_actuator_id: a.target_actuator_id ?? null,
    target_notification_template_id: a.target_notification_template_id ?? null,
    action_command: a.action_command ?? '',
    delay_before_execution_seconds: a.delay_before_execution_seconds ?? 0,
    _params: {
      title: params.title ?? '',
      description: params.description ?? '',
      zone_id: params.zone_id ?? null,
      task_type: params.task_type ?? '',
      priority: params.priority ?? 1,
      due_in_days: params.due_in_days ?? null,
      duration_seconds: params.duration_seconds ?? null,
    },
    _variables: params.variables && typeof params.variables === 'object'
      ? Object.entries(params.variables).map(([key, value]) => ({ key, value: String(value) }))
      : [],
  }
}

function hydrateFromProps() {
  if (!props.rule) {
    form.value = emptyForm()
    triggerSensorId.value = props.sensors[0]?.id || 0
    triggerAnimalGroupId.value = props.animalGroups[0]?.id || 0
    return
  }
  const r = props.rule
  const conditions = r.conditions_jsonb && typeof r.conditions_jsonb === 'object' && Array.isArray(r.conditions_jsonb.predicates)
    ? r.conditions_jsonb.predicates.map(p => {
        if (p.type === 'setpoint') {
          return {
            type: 'setpoint',
            sensor_type: p.sensor_type || '',
            scope: p.scope || 'current_stage',
            op: p.op || 'out_of_range',
            sensor_id: 0,
            value: 0,
            animal_group_id: 0,
            event_type: '',
          }
        }
        if (p.type === 'animal_event') {
          return {
            type: 'animal_event',
            animal_group_id: Number(p.animal_group_id) || 0,
            event_type: p.event_type || '',
            sensor_id: 0,
            op: 'gte',
            value: 0,
            sensor_type: '',
            scope: 'current_stage',
          }
        }
        return {
          type: 'hard',
          sensor_id: Number(p.sensor_id) || 0,
          op: p.op || 'gte',
          value: Number(p.value) || 0,
          sensor_type: '',
          scope: 'current_stage',
          animal_group_id: 0,
          event_type: '',
        }
      })
    : []
  const trigCfg = r.trigger_configuration && typeof r.trigger_configuration === 'object'
    ? r.trigger_configuration
    : {}
  triggerSensorId.value = Number(trigCfg.sensor_id) || 0
  triggerAnimalGroupId.value = Number(trigCfg.animal_group_id) || 0
  form.value = {
    name: r.name || '',
    description: r.description || '',
    is_active: !!r.is_active,
    trigger_source: r.trigger_source || 'sensor_reading_threshold',
    cooldown_period_seconds: r.cooldown_period_seconds ?? 0,
    condition_logic: r.condition_logic || 'ALL',
    conditions,
    actions: (props.actions || []).slice().sort((a, b) => a.execution_order - b.execution_order).map(buildActionFormState),
  }
}

watch(() => [props.rule, props.actions, props.sensors, props.animalGroups], hydrateFromProps, { immediate: true })

function addCondition() {
  form.value.conditions.push({
    type: 'hard',
    sensor_id: props.sensors[0]?.id || 0,
    op: 'gte',
    value: 0,
    sensor_type: '',
    scope: 'current_stage',
    animal_group_id: 0,
    event_type: '',
  })
}
function removeCondition(idx) { form.value.conditions.splice(idx, 1) }

// Switch a predicate between "hard" (sensor_id/op/value), "setpoint"
// (sensor_type/scope/op), and "animal_event" (animal_group_id/event_type).
// Resets the other variants' fields on switch so saving the form doesn't
// leak stale values into the wire payload (the backend validator rejects
// e.g. sensor_id + value on a setpoint or animal_event predicate).
function setPredicateType(p, type) {
  p.type = type
  if (type === 'setpoint') {
    p.sensor_id = 0
    p.value = 0
    if (!p.sensor_type) p.sensor_type = props.sensors?.[0]?.sensor_type || ''
    p.scope = p.scope || 'current_stage'
    if (!['out_of_range', 'below_ideal', 'above_ideal', 'inside_range'].includes(p.op)) {
      p.op = 'out_of_range'
    }
  } else if (type === 'animal_event') {
    p.sensor_id = 0
    p.value = 0
    p.sensor_type = ''
    if (!p.animal_group_id) p.animal_group_id = props.animalGroups?.[0]?.id || 0
  } else {
    p.type = 'hard'
    p.sensor_type = ''
    p.scope = 'current_stage'
    p.animal_group_id = 0
    p.event_type = ''
    if (!p.sensor_id) p.sensor_id = props.sensors?.[0]?.id || 0
    if (!['lt', 'lte', 'eq', 'gte', 'gt', 'ne'].includes(p.op)) p.op = 'gte'
  }
}

const uniqueSensorTypes = computed(() => {
  const set = new Set()
  for (const s of props.sensors || []) {
    if (s.sensor_type) set.add(s.sensor_type)
  }
  return Array.from(set).sort()
})

function addAction() {
  form.value.actions.push({
    _key: newKey(),
    id: null,
    action_type: 'control_actuator',
    execution_order: form.value.actions.length,
    target_actuator_id: props.actuators[0]?.id ?? null,
    target_notification_template_id: null,
    action_command: 'on',
    delay_before_execution_seconds: 0,
    _params: {
      title: '', description: '', zone_id: null,
      task_type: '', priority: 1, due_in_days: null, duration_seconds: null,
    },
    _variables: [],
  })
}
function removeAction(idx) { form.value.actions.splice(idx, 1) }
function moveAction(idx, delta) {
  const next = idx + delta
  if (next < 0 || next >= form.value.actions.length) return
  const arr = form.value.actions
  ;[arr[idx], arr[next]] = [arr[next], arr[idx]]
}
function onActionTypeChange(a) {
  if (a.action_type !== 'control_actuator') {
    a.target_actuator_id = null
    a.action_command = ''
  } else if (!a.target_actuator_id) {
    a.target_actuator_id = props.actuators[0]?.id ?? null
    a.action_command = a.action_command || 'on'
  }
  if (a.action_type !== 'send_notification') {
    a.target_notification_template_id = null
  }
}

function addVariable(a) { a._variables.push({ key: '', value: '' }) }
function removeVariable(a, i) { a._variables.splice(i, 1) }

function buildActionParameters(a) {
  if (a.action_type === 'control_actuator') {
    const d = Number(a._params.duration_seconds)
    return d > 0 ? { duration_seconds: d } : null
  }
  if (a.action_type === 'create_task') {
    const p = a._params
    const out = {}
    if (p.title) out.title = p.title
    if (p.description) out.description = p.description
    if (p.zone_id != null && p.zone_id !== '') out.zone_id = Number(p.zone_id)
    if (p.task_type) out.task_type = p.task_type
    if (p.priority != null && !Number.isNaN(Number(p.priority))) out.priority = Number(p.priority)
    if (p.due_in_days != null && !Number.isNaN(Number(p.due_in_days))) out.due_in_days = Number(p.due_in_days)
    return out
  }
  if (a.action_type === 'send_notification' && a._variables.length) {
    const vars = {}
    for (const kv of a._variables) {
      if (kv.key) vars[kv.key] = kv.value
    }
    return Object.keys(vars).length ? { variables: vars } : null
  }
  return null
}

function submit() {
  if (!form.value.name) {
    emit('submit', { error: 'Name is required.' })
    return
  }
  for (const [i, p] of form.value.conditions.entries()) {
    if (p.type === 'setpoint') {
      if (!p.sensor_type?.trim()) {
        emit('submit', { error: `Condition ${i + 1}: sensor_type is required for setpoint predicates.` })
        return
      }
    } else if (p.type === 'animal_event') {
      if (!p.animal_group_id) {
        emit('submit', { error: `Condition ${i + 1}: pick an animal group.` })
        return
      }
      if (!p.event_type?.trim()) {
        emit('submit', { error: `Condition ${i + 1}: event_type is required for animal_event predicates.` })
        return
      }
    } else if (!p.sensor_id) {
      emit('submit', { error: `Condition ${i + 1}: pick a sensor.` })
      return
    }
  }
  const rulePayload = {
    name: form.value.name,
    description: form.value.description || null,
    is_active: !!form.value.is_active,
    trigger_source: form.value.trigger_source,
    trigger_configuration: (() => {
      if (form.value.trigger_source === 'sensor_reading_threshold' && triggerSensorId.value) {
        return { sensor_id: Number(triggerSensorId.value) }
      }
      if (form.value.trigger_source === 'animal_lifecycle_event' && triggerAnimalGroupId.value) {
        return { animal_group_id: Number(triggerAnimalGroupId.value) }
      }
      return {}
    })(),
    condition_logic: form.value.condition_logic,
    conditions: form.value.conditions.map(p => {
      if (p.type === 'setpoint') {
        return {
          type: 'setpoint',
          sensor_type: p.sensor_type.trim(),
          scope: p.scope || 'current_stage',
          op: p.op,
        }
      }
      if (p.type === 'animal_event') {
        return {
          type: 'animal_event',
          animal_group_id: Number(p.animal_group_id),
          event_type: p.event_type.trim(),
        }
      }
      return {
        type: 'hard',
        sensor_id: Number(p.sensor_id),
        op: p.op,
        value: Number(p.value),
      }
    }),
    cooldown_period_seconds: Number(form.value.cooldown_period_seconds) || 0,
  }
  const actionPayloads = form.value.actions.map((a, idx) => {
    const params = buildActionParameters(a)
    const payload = {
      id: a.id,
      execution_order: idx,
      action_type: a.action_type,
      target_actuator_id: a.action_type === 'control_actuator' ? (a.target_actuator_id ?? null) : null,
      target_notification_template_id: a.action_type === 'send_notification'
        ? (a.target_notification_template_id ?? null)
        : null,
      action_command: a.action_type === 'control_actuator' ? (a.action_command || null) : null,
      action_parameters: params,
      delay_before_execution_seconds: Number(a.delay_before_execution_seconds) || 0,
    }
    return payload
  })
  emit('submit', { rule: rulePayload, actions: actionPayloads })
}
</script>
