<template>
  <div class="space-y-3" data-test="comfort-band-editor">
    <p v-if="targetError" class="text-red-400 text-xs">{{ targetError }}</p>

    <p v-if="!sensorTypes.length" class="text-zinc-500 text-sm">
      {{ emptyMessage || `Add a ${needLabel} sensor before setting comfort targets.` }}
    </p>

    <form
      v-for="row in displayRows"
      :key="row.key"
      class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
      :data-test="`comfort-target-${row.sensorType}`"
      @submit.prevent="saveRow(row)"
    >
      <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
        <div>
          <p class="text-sm text-zinc-200 font-medium">{{ row.label }}</p>
          <p v-if="row.liveValue != null" class="text-[11px] text-zinc-500 mt-0.5">
            Now: {{ row.liveValue.toFixed(1) }}
            <span
              v-if="row.status === 'out_of_range'"
              class="text-red-400 ml-1"
            >· out of range</span>
          </p>
        </div>
        <span v-if="row.stageLabel" class="text-[10px] text-zinc-600">
          {{ row.stageLabel }}
        </span>
      </div>
      <div class="grid grid-cols-3 gap-2 mb-2">
        <label class="text-[10px] text-zinc-500">
          Too low
          <input
            v-model.number="draftValues[row.key].min_value"
            type="number"
            step="any"
            class="mt-0.5 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
            placeholder="min"
          />
        </label>
        <label class="text-[10px] text-zinc-500">
          Just right
          <input
            v-model.number="draftValues[row.key].ideal_value"
            type="number"
            step="any"
            class="mt-0.5 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
            placeholder="ideal"
          />
        </label>
        <label class="text-[10px] text-zinc-500">
          Too high
          <input
            v-model.number="draftValues[row.key].max_value"
            type="number"
            step="any"
            class="mt-0.5 w-full bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-xs text-white"
            placeholder="max"
          />
        </label>
      </div>
      <div class="flex items-center gap-2">
        <button
          type="submit"
          class="text-xs px-2 py-1 rounded bg-green-800 hover:bg-green-700 text-white disabled:opacity-50"
          :disabled="savingKey === row.key"
        >
          {{ savingKey === row.key ? 'Saving…' : (row.setpoint?.id ? 'Save' : 'Add target') }}
        </button>
        <button
          v-if="row.setpoint?.id && showRemove"
          type="button"
          class="text-xs text-zinc-500 hover:text-red-400"
          :disabled="savingKey === row.key"
          @click="deleteRow(row)"
        >
          Remove
        </button>
      </div>
    </form>

    <div v-if="missingTypes.length" class="space-y-2">
      <EmptyStateHint
        v-for="st in missingTypes"
        :key="st"
        reason="no_setpoint"
        :message="`No comfort target for ${sensorTypeLabel(st)} yet.`"
        action-label=""
        :action-to="null"
        compact
      />
      <div class="flex flex-wrap gap-2">
        <button
          v-for="st in missingTypes"
          :key="st"
          type="button"
          class="text-xs text-green-600 hover:text-green-400 border border-green-900/50 rounded-lg px-2 py-1"
          :data-test="`add-comfort-target-${st}`"
          @click="startAdd(st)"
        >
          + Add target for {{ sensorTypeLabel(st) }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import { NEED_META, sensorPlantNeed } from '../lib/plantNeeds.js'
import {
  buildZoneComfortBands,
  parseComfortNumber,
  validateComfortBandPayload,
  zoneSetpointForType,
} from '../lib/comfortBand.js'
import { sensorTypeLabel } from '../lib/sensorTypeLabel.js'
import EmptyStateHint from './EmptyStateHint.vue'

const props = defineProps({
  need: { type: String, required: true },
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  setpoints: { type: Array, default: () => [] },
  readings: { type: Object, default: () => ({}) },
  /** Limit editor to these sensor types (hub passes climate types only). */
  sensorTypesFilter: { type: Array, default: null },
  emptyMessage: { type: String, default: '' },
  showRemove: { type: Boolean, default: true },
  activeStage: { type: String, default: '' },
})

const emit = defineEmits(['updated'])

const targetError = ref('')
const savingKey = ref('')
const pendingAdds = ref([])
const draftValues = ref({})

const needLabel = computed(() => NEED_META[props.need]?.shortLabel?.toLowerCase() || 'sensor')

const needSensors = computed(() =>
  props.sensors.filter((s) => sensorPlantNeed(s.sensor_type) === props.need),
)

const sensorTypes = computed(() => {
  if (props.sensorTypesFilter?.length) {
    return props.sensorTypesFilter.filter((t) =>
      needSensors.value.some((s) => s.sensor_type === t),
    )
  }
  const types = new Set(needSensors.value.map((s) => s.sensor_type).filter(Boolean))
  return [...types]
})

const bandMeta = computed(() => {
  const map = new Map()
  for (const band of buildZoneComfortBands({
    zoneId: props.zoneId,
    sensors: props.sensors,
    setpoints: props.setpoints,
    readings: props.readings,
  })) {
    map.set(band.sensorType, band)
  }
  return map
})

function rowKey(sensorType, setpoint) {
  return `${sensorType}:${setpoint?.id || 'new'}`
}

function stageLabelFor(setpoint) {
  if (setpoint?.stage) return `For ${setpoint.stage.replace(/_/g, ' ')}`
  if (props.activeStage) return `For ${props.activeStage.replace(/_/g, ' ')}`
  return ''
}

const displayRows = computed(() => {
  const rows = []
  for (const sensorType of sensorTypes.value) {
    const setpoint = zoneSetpointForType(props.setpoints, props.zoneId, sensorType)
    if (setpoint || pendingAdds.value.includes(sensorType)) {
      const meta = bandMeta.value.get(sensorType)
      rows.push({
        key: rowKey(sensorType, setpoint),
        sensorType,
        label: sensorTypeLabel(sensorType),
        setpoint,
        liveValue: meta?.liveValue ?? null,
        status: meta?.status ?? 'missing',
        stageLabel: stageLabelFor(setpoint),
      })
    }
  }
  return rows
})

function syncDrafts() {
  const next = { ...draftValues.value }
  for (const row of displayRows.value) {
    if (!next[row.key]) {
      next[row.key] = {
        min_value: row.setpoint?.min_value ?? null,
        ideal_value: row.setpoint?.ideal_value ?? null,
        max_value: row.setpoint?.max_value ?? null,
      }
    }
  }
  draftValues.value = next
}

const missingTypes = computed(() =>
  sensorTypes.value.filter(
    (st) => !zoneSetpointForType(props.setpoints, props.zoneId, st) && !pendingAdds.value.includes(st),
  ),
)

watch([displayRows, () => props.setpoints], syncDrafts, { immediate: true, deep: true })

watch(
  () => props.setpoints,
  () => {
    pendingAdds.value = []
    draftValues.value = {}
    syncDrafts()
  },
  { deep: true },
)

function startAdd(sensorType) {
  if (!pendingAdds.value.includes(sensorType)) {
    pendingAdds.value = [...pendingAdds.value, sensorType]
  }
}

async function saveRow(row) {
  targetError.value = ''
  const payload = {
    zone_id: props.zoneId,
    crop_cycle_id: null,
    stage: row.setpoint?.stage ?? (props.activeStage || null),
    sensor_type: row.sensorType,
    min_value: parseComfortNumber(draftValues.value[row.key]?.min_value),
    ideal_value: parseComfortNumber(draftValues.value[row.key]?.ideal_value),
    max_value: parseComfortNumber(draftValues.value[row.key]?.max_value),
  }
  const err = validateComfortBandPayload(payload)
  if (err) {
    targetError.value = err
    return
  }

  savingKey.value = row.key
  try {
    if (row.setpoint?.id) {
      await api.put(`/setpoints/${row.setpoint.id}`, payload)
    } else {
      await api.post(`/farms/${props.farmId}/setpoints`, payload)
    }
    pendingAdds.value = pendingAdds.value.filter((t) => t !== row.sensorType)
    emit('updated')
  } catch (e) {
    targetError.value = e.response?.data?.error || e.message || 'Could not save target'
  } finally {
    savingKey.value = ''
  }
}

async function deleteRow(row) {
  if (!row.setpoint?.id) return
  if (!confirm(`Remove comfort target for ${row.label}?`)) return
  savingKey.value = row.key
  targetError.value = ''
  try {
    await api.delete(`/setpoints/${row.setpoint.id}`)
    emit('updated')
  } catch (e) {
    targetError.value = e.response?.data?.error || e.message || 'Could not remove target'
  } finally {
    savingKey.value = ''
  }
}
</script>
