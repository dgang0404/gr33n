<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-automation-panel">
    <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
      <div>
        <h3 class="text-sm font-semibold text-white">What runs when</h3>
        <p class="text-zinc-600 text-xs mt-0.5">Automations and schedules for this zone.</p>
      </div>
      <div class="flex gap-3 text-xs">
        <router-link
          v-nav-hint="'/comfort-targets'"
          :to="{ path: '/comfort-targets', query: { zone_id: String(zoneId), tab: 'rules' } }"
          class="text-zinc-500 hover:text-green-400"
        >Automations →</router-link>
        <router-link
          v-nav-hint="'/comfort-targets'"
          :to="{ path: '/comfort-targets', query: { zone_id: String(zoneId), tab: 'schedules' } }"
          class="text-zinc-500 hover:text-green-400"
        >What runs when →</router-link>
      </div>
    </div>

    <p v-if="!automation.rules.length && !automation.schedules.length" class="text-zinc-500 text-sm">
      Nothing automated for {{ needLabel }} in this zone yet.
    </p>

    <div v-if="automation.schedules.length" class="mb-4">
      <p class="text-xs text-zinc-500 uppercase tracking-wide mb-2">Timed runs</p>
      <ul class="space-y-2">
        <li
          v-for="item in automation.schedules"
          :key="item.schedule.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2"
          :data-test="`zone-schedule-${item.schedule.id}`"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <p class="text-sm text-zinc-200 font-medium truncate">{{ item.schedule.name }}</p>
              <p class="text-xs text-green-400/90 mt-0.5">{{ item.runsLabel }}</p>
              <p v-if="item.linkedName" class="text-[11px] text-zinc-500 mt-0.5">
                Linked: {{ item.linkedName }}
              </p>
            </div>
            <div class="flex flex-col items-end gap-1 shrink-0">
              <span
                class="text-[10px] px-2 py-0.5 rounded-full"
                :class="item.schedule.is_active ? 'bg-green-900/50 text-green-300' : 'bg-zinc-800 text-zinc-500'"
              >
                {{ item.schedule.is_active ? 'On' : 'Off' }}
              </span>
              <button
                type="button"
                class="text-[10px] px-2 py-0.5 rounded border transition-colors disabled:opacity-50"
                :class="item.schedule.is_active ? 'border-yellow-800/50 text-yellow-400' : 'border-green-800/50 text-green-400'"
                :disabled="scheduleTogglingId === item.schedule.id"
                :data-test="`zone-schedule-toggle-${item.schedule.id}`"
                @click="toggleSchedule(item.schedule)"
              >
                {{ scheduleTogglingId === item.schedule.id ? '…' : (item.schedule.is_active ? 'Pause' : 'Resume') }}
              </button>
            </div>
          </div>
        </li>
      </ul>
    </div>

    <div v-if="automation.rules.length">
      <p class="text-xs text-zinc-500 uppercase tracking-wide mb-2">Automations</p>
      <ul class="space-y-2">
        <li
          v-for="rule in automation.rules"
          :key="rule.id"
          class="bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2"
          :data-test="`zone-rule-${rule.id}`"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0 flex-1">
              <p class="text-sm text-zinc-200 font-medium truncate">{{ rule.name }}</p>
              <p v-if="rule.description" class="text-[11px] text-zinc-600 line-clamp-2 mt-0.5">
                {{ rule.description }}
              </p>
              <p class="text-[11px] text-zinc-500 mt-1">
                Last fired: {{ formatRuleLastFired(rule.last_triggered_time) }}
              </p>
              <div v-if="ruleBadges(rule).length" class="flex flex-wrap gap-1 mt-1.5">
                <span
                  v-for="b in ruleBadges(rule)"
                  :key="b.id"
                  class="text-[10px] px-1.5 py-0.5 rounded"
                  :class="b.tone === 'warn' ? 'bg-amber-900/40 text-amber-300' : 'bg-zinc-800 text-zinc-400'"
                >
                  {{ b.label }}
                </span>
              </div>
            </div>
            <div class="flex flex-col items-end gap-1 shrink-0">
              <button
                type="button"
                class="text-xs px-2 py-1 rounded border transition-colors disabled:opacity-50"
                :class="rule.is_active ? 'border-green-700 text-green-400' : 'border-zinc-700 text-zinc-400'"
                :disabled="togglingId === rule.id"
                :data-test="`zone-rule-toggle-${rule.id}`"
                @click="toggleRule(rule)"
              >
                {{ togglingId === rule.id ? '…' : (rule.is_active ? 'On' : 'Off') }}
              </button>
              <router-link
                v-nav-hint="'/automation'"
                :to="`/automation?rule=${rule.id}`"
                class="text-[10px] text-zinc-500 hover:text-green-400"
                data-test="zone-rule-edit-automation"
              >
                Edit in Automations →
              </router-link>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { NEED_META } from '../lib/plantNeeds.js'
import { useFarmStore } from '../stores/farm.js'
import {
  zoneAutomationForNeed,
  greenhouseRuleBadges,
  formatRuleLastFired,
} from '../lib/zoneAutomationContext.js'

const props = defineProps({
  need: { type: String, required: true },
  zoneId: { type: Number, required: true },
  zoneName: { type: String, default: '' },
  sensors: { type: Array, default: () => [] },
  rules: { type: Array, default: () => [] },
  schedules: { type: Array, default: () => [] },
  activeProgram: { type: Object, default: null },
  lightingPrograms: { type: Array, default: () => [] },
})

const emit = defineEmits(['rules-updated', 'schedules-updated'])

const store = useFarmStore()
const togglingId = ref(null)
const scheduleTogglingId = ref(null)

const needLabel = computed(() => NEED_META[props.need]?.shortLabel?.toLowerCase() || 'this need')

const automation = computed(() => zoneAutomationForNeed(props.need, {
  zoneId: props.zoneId,
  zoneName: props.zoneName,
  sensors: props.sensors,
  rules: props.rules,
  schedules: props.schedules,
  activeProgram: props.activeProgram,
  lightingPrograms: props.lightingPrograms,
}))

function ruleBadges(rule) {
  return greenhouseRuleBadges(rule, props.sensors)
}

async function toggleRule(rule) {
  togglingId.value = rule.id
  try {
    await store.updateAutomationRuleActive(rule.id, !rule.is_active)
    emit('rules-updated')
  } finally {
    togglingId.value = null
  }
}

async function toggleSchedule(schedule) {
  scheduleTogglingId.value = schedule.id
  try {
    await store.updateScheduleActive(schedule.id, !schedule.is_active)
    emit('schedules-updated')
  } finally {
    scheduleTogglingId.value = null
  }
}
</script>
