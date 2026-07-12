<template>
  <Teleport to="body">
    <div
      v-if="open && zone"
      class="fixed inset-0 z-[60] flex items-end md:items-center justify-center p-0 md:p-4"
      data-test="zone-quick-actions"
      role="presentation"
      @click.self="close"
    >
      <div
        ref="panelRef"
        class="w-full md:max-w-md bg-zinc-900 border border-zinc-700 md:rounded-xl rounded-t-2xl shadow-2xl max-h-[85vh] overflow-y-auto"
        role="dialog"
        :aria-labelledby="titleId"
        aria-modal="true"
        data-test="zone-quick-actions-panel"
      >
        <header class="sticky top-0 bg-zinc-900/95 backdrop-blur border-b border-zinc-800 px-4 py-3 flex items-start justify-between gap-3">
          <div class="min-w-0">
            <h2 :id="titleId" class="text-base font-semibold text-white truncate">{{ zone.name }}</h2>
            <p class="text-[11px] text-zinc-500">{{ zoneTypeLabel }} · quick actions</p>
          </div>
          <button
            type="button"
            class="shrink-0 min-h-[44px] min-w-[44px] rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800"
            aria-label="Close"
            data-test="zone-quick-actions-close"
            @click="close"
          >
            ✕
          </button>
        </header>

        <div class="p-4 space-y-4">
          <p v-if="actionFeedback" class="text-xs" :class="actionOk ? 'text-green-400' : 'text-amber-400'">
            {{ actionFeedback }}
          </p>

          <!-- Water -->
          <section v-if="waterAction" data-test="zone-quick-water">
            <router-link
              v-if="waterAction.mode === 'setup'"
              :to="waterAction.link"
              class="flex items-center min-h-[44px] w-full px-4 py-3 rounded-xl bg-zinc-950 border border-zinc-700 text-sm text-zinc-300 hover:text-white"
              @click="close"
            >
              💧 {{ waterAction.label }}
            </router-link>
            <button
              v-else
              type="button"
              class="w-full min-h-[44px] px-4 py-3 rounded-xl bg-blue-950/40 border border-blue-800/60 text-left text-sm text-blue-200 hover:bg-blue-950/60"
              :disabled="waterBusy"
              @click="onWaterNow"
            >
              💧 {{ waterAction.label }}
            </button>
          </section>

          <!-- Light -->
          <section v-if="lights.length" data-test="zone-quick-light">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-2">Light</p>
            <div v-for="a in lights" :key="a.id" class="flex items-center gap-2 mb-2">
              <span class="text-xs text-zinc-300 flex-1 truncate">{{ a.name }}</span>
              <button
                type="button"
                class="min-h-[44px] min-w-[44px] px-3 rounded-lg bg-gr33n-700 text-white text-xs font-semibold disabled:opacity-40"
                :disabled="busyId === a.id"
                @click="sendCommand(a, 'on', `${zone.name}: light on`)"
              >
                On
              </button>
              <button
                type="button"
                class="min-h-[44px] min-w-[44px] px-3 rounded-lg bg-zinc-800 text-zinc-300 text-xs font-semibold disabled:opacity-40"
                :disabled="busyId === a.id"
                @click="sendCommand(a, 'off', `${zone.name}: light off`)"
              >
                Off
              </button>
            </div>
          </section>

          <!-- Greenhouse -->
          <section v-if="ghControls.length" data-test="zone-quick-greenhouse">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-2">Greenhouse</p>
            <div v-for="row in ghControls" :key="row.actuator.id" class="mb-2">
              <p class="text-xs text-zinc-400 mb-1 capitalize">{{ row.role }} · {{ row.actuator.name }}</p>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="cmd in row.commands"
                  :key="cmd"
                  type="button"
                  class="min-h-[44px] px-3 rounded-lg bg-zinc-800 text-zinc-200 text-xs capitalize disabled:opacity-40"
                  :disabled="busyId === row.actuator.id"
                  @click="sendCommand(row.actuator, cmd, `${zone.name}: ${cmd}`)"
                >
                  {{ cmd }}
                </button>
              </div>
            </div>
          </section>

          <!-- Tasks -->
          <section v-if="sheetTasks.length" data-test="zone-quick-tasks">
            <div class="flex items-center justify-between mb-2">
              <p class="text-[10px] uppercase tracking-wide text-zinc-500">Today's tasks</p>
              <router-link
                :to="{ path: `/zones/${zone.id}`, query: { tab: 'ops', ops: 'tasks' } }"
                class="text-[10px] text-gr33n-500"
              >
                View all →
              </router-link>
            </div>
            <label
              v-for="t in sheetTasks"
              :key="t.id"
              class="flex items-center gap-3 min-h-[44px] px-3 rounded-lg bg-zinc-950 border border-zinc-800 mb-2 cursor-pointer"
            >
              <input
                type="checkbox"
                class="w-4 h-4"
                :checked="t.status === 'completed'"
                :disabled="taskBusy === t.id"
                @change="completeTask(t)"
              />
              <span class="text-sm text-zinc-200 truncate">{{ t.title }}</span>
            </label>
          </section>

          <!-- Alerts -->
          <section v-if="sheetAlerts.length" data-test="zone-quick-alerts">
            <div class="flex items-center justify-between mb-2">
              <p class="text-[10px] uppercase tracking-wide text-zinc-500">Alerts</p>
              <router-link
                :to="{ path: `/zones/${zone.id}`, query: { tab: 'alerts' } }"
                class="text-[10px] text-gr33n-500"
              >
                View all →
              </router-link>
            </div>
            <div
              v-for="a in sheetAlerts"
              :key="a.id"
              class="flex items-center justify-between gap-2 min-h-[44px] px-3 rounded-lg bg-zinc-950 border border-zinc-800 mb-2"
            >
              <span class="text-sm text-zinc-200 truncate">{{ a.title || a.subject_rendered || 'Alert' }}</span>
              <button
                type="button"
                class="shrink-0 min-h-[44px] px-3 rounded-lg text-xs bg-amber-900/40 text-amber-200 border border-amber-800/50 disabled:opacity-40"
                :disabled="alertBusy === a.id"
                @click="ackAlert(a.id)"
              >
                Ack
              </button>
            </div>
          </section>

          <!-- Guardian -->
          <section v-if="guardianStarters.length" data-test="zone-quick-guardian">
            <p class="text-[10px] uppercase tracking-wide text-zinc-500 mb-2">Guardian</p>
            <button
              v-for="s in guardianStarters"
              :key="s.id"
              type="button"
              class="w-full min-h-[44px] mb-2 px-4 py-2 rounded-xl border border-green-800/60 bg-green-950/30 text-sm text-green-200 text-left hover:bg-green-950/50"
              @click="askGuardian(s)"
            >
              🧙 {{ s.label }}
            </button>
          </section>

          <!-- Open zone -->
          <router-link
            :to="`/zones/${zone.id}`"
            class="flex items-center justify-center min-h-[44px] w-full rounded-xl border border-zinc-700 text-sm text-zinc-300 hover:text-white hover:border-zinc-600"
            data-test="zone-quick-open-detail"
            @click="close"
          >
            ⚙️ Open {{ zone.name }}
          </router-link>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useActuatorCommands } from '../composables/useActuatorCommands'
import { useDialogFocusTrap } from '../composables/useDialogFocusTrap'
import { buildZoneQuickStarters } from '../lib/guardianStarters.js'
import {
  greenhouseActuatorsForZone,
  lightActuatorsForZone,
  resolveWaterNowAction,
  zoneAlertsForSheet,
  zoneTasksForSheet,
} from '../lib/zoneQuickActions.js'
import { formatZoneTypeLabel } from '../lib/farmVisualStatus.js'

const props = defineProps({
  open: { type: Boolean, default: false },
  zone: { type: Object, default: null },
  status: { type: Object, default: null },
  farmId: { type: Number, default: null },
  programs: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  tasks: { type: Array, default: () => [] },
  alerts: { type: Array, default: () => [] },
  sensors: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'refresh'])

const zoneTypeLabel = computed(() => formatZoneTypeLabel(props.zone?.zone_type))

const store = useFarmStore()
const guardianPanel = useGuardianPanelStore()
const { busyId, feedback, sendCommand, runPulse } = useActuatorCommands()

const panelRef = ref(null)
const openRef = computed(() => props.open)
const titleId = `zone-quick-title-${props.zone?.id ?? 'x'}`
const waterBusy = ref(false)
const taskBusy = ref(null)
const alertBusy = ref(null)
const actionFeedback = ref('')
const actionOk = ref(true)

useDialogFocusTrap(openRef, panelRef, {
  onEscape: () => close(),
})

const waterAction = computed(() =>
  props.zone ? resolveWaterNowAction({
    zone: props.zone,
    programs: props.programs,
    actuators: props.actuators,
  }) : null,
)

const lights = computed(() =>
  props.zone ? lightActuatorsForZone(props.actuators, props.zone.id) : [],
)

const ghControls = computed(() =>
  props.zone ? greenhouseActuatorsForZone(props.zone, props.actuators) : [],
)

const sheetTasks = computed(() =>
  props.zone ? zoneTasksForSheet(props.tasks, props.zone.id) : [],
)

const sheetAlerts = computed(() =>
  props.zone ? zoneAlertsForSheet(props.alerts, props.sensors, props.zone) : [],
)

const guardianStarters = computed(() =>
  buildZoneQuickStarters({
    zone: props.zone,
    status: props.status,
    farmId: props.farmId,
  }),
)

watch(() => props.open, (v) => {
  if (v) {
    actionFeedback.value = ''
    feedback.value = ''
  }
})

watch(feedback, (v) => {
  if (v) {
    actionFeedback.value = v
    actionOk.value = !v.toLowerCase().includes('fail')
  }
})

function close() {
  emit('close')
}

async function onWaterNow() {
  const wa = waterAction.value
  if (!wa || !props.zone || !props.farmId) return

  if (wa.mode === 'setup') return

  const ok = window.confirm(wa.confirm || wa.label)
  if (!ok) return

  waterBusy.value = true
  actionFeedback.value = ''
  try {
    if (wa.mode === 'program') {
      const res = await store.runFertigationProgramNow(props.farmId, wa.program.id)
      actionOk.value = true
      actionFeedback.value = res?.duplicate
        ? 'Already queued — check recent feeds after refresh'
        : `Started ${wa.program.name}`
      emit('refresh')
    } else if (wa.mode === 'pulse') {
      const pulseOk = await runPulse(wa.actuator, wa.defaultSeconds || 60)
      actionOk.value = pulseOk
      if (pulseOk) emit('refresh')
    } else if (wa.mode === 'toggle') {
      const cmdOk = await sendCommand(wa.actuator, 'on')
      actionOk.value = cmdOk
      if (cmdOk) emit('refresh')
    }
  } catch (e) {
    actionOk.value = false
    actionFeedback.value = e?.response?.data?.error || e.message || 'Water now failed'
  } finally {
    waterBusy.value = false
  }
}

async function completeTask(task) {
  taskBusy.value = task.id
  try {
    await store.updateTaskStatus(task.id, 'completed')
    actionOk.value = true
    actionFeedback.value = `Completed: ${task.title}`
    emit('refresh')
  } catch (e) {
    actionOk.value = false
    actionFeedback.value = e?.response?.data?.error || e.message || 'Could not complete task'
  } finally {
    taskBusy.value = null
  }
}

async function ackAlert(id) {
  alertBusy.value = id
  try {
    await store.markAlertAcknowledged(id)
    actionOk.value = true
    actionFeedback.value = 'Alert acknowledged'
    emit('refresh')
  } catch (e) {
    actionOk.value = false
    actionFeedback.value = e?.response?.data?.error || e.message || 'Acknowledge failed'
  } finally {
    alertBusy.value = null
  }
}

function askGuardian(starter) {
  guardianPanel.openDrawer({
    prefilledMessage: starter.message,
    contextRef: starter.contextRef ?? { type: 'zone', id: props.zone.id, name: props.zone.name },
    farmCounsel: true,
    autoSend: true,
  })
  close()
}
</script>
