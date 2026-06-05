<template>
  <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4" data-test="zone-comfort-targets">
    <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
      <div>
        <h3 class="text-sm font-semibold text-white">Comfort targets</h3>
        <p class="text-zinc-600 text-xs mt-0.5">How comfortable this room should be — min, ideal, and max.</p>
      </div>
      <router-link to="/setpoints" class="text-xs text-zinc-500 hover:text-green-400">
        Farm-wide bands →
      </router-link>
    </div>

    <p v-if="targetError" class="text-red-400 text-xs mb-2">{{ targetError }}</p>

    <p v-if="!sensorTypes.length" class="text-zinc-500 text-sm">
      Add a {{ needLabel }} sensor to this zone before setting comfort targets.
    </p>

    <div v-else class="space-y-3">
      <form
        v-for="row in displayRows"
        :key="row.key"
        class="bg-zinc-950 border border-zinc-800 rounded-lg p-3"
        :data-test="`comfort-target-${row.sensorType}`"
        @submit.prevent="saveRow(row)"
      >
        <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
          <p class="text-sm text-zinc-200 font-medium">{{ row.label }}</p>
          <span v-if="row.setpoint?.stage" class="text-[10px] text-zinc-600">
            Stage: {{ row.setpoint.stage }}
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
            v-if="row.setpoint?.id"
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
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import { sensorPlantNeed, NEED_META } from '../lib/plantNeeds.js'
import { sensorTypeLabel } from '../lib/sensorTypeLabel.js'
import EmptyStateHint from './EmptyStateHint.vue'

const props = defineProps({
  need: { type: String, required: true },
  zoneId: { type: Number, required: true },
  farmId: { type: Number, required: true },
  sensors: { type: Array, default: () => [] },
  setpoints: { type: Array, default: () => [] },
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
  const types = new Set(needSensors.value.map((s) => s.sensor_type).filter(Boolean))
  return [...types]
})

function zoneSetpointForType(sensorType) {
  const matches = props.setpoints.filter(
    (sp) => sp.zone_id === props.zoneId
      && !sp.crop_cycle_id
      && sp.sensor_type === sensorType,
  )
  const anyStage = matches.find((sp) => !sp.stage)
  return anyStage || matches[0] || null
}

function rowKey(sensorType, setpoint) {
  return `${sensorType}:${setpoint?.id || 'new'}`
}

const displayRows = computed(() => {
  const rows = []
  for (const sensorType of sensorTypes.value) {
    const setpoint = zoneSetpointForType(sensorType)
    if (setpoint || pendingAdds.value.includes(sensorType)) {
      rows.push({
        key: rowKey(sensorType, setpoint),
        sensorType,
        label: sensorTypeLabel(sensorType),
        setpoint,
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
  sensorTypes.value.filter((st) => !zoneSetpointForType(st) && !pendingAdds.value.includes(st)),
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

function num(v) {
  if (v === '' || v == null || Number.isNaN(Number(v))) return null
  return Number(v)
}

function validatePayload(payload) {
  if (!payload.sensor_type) return 'Sensor type is required'
  if (payload.min_value != null && payload.max_value != null && payload.min_value > payload.max_value) {
    return 'Too low must be ≤ too high'
  }
  if (payload.ideal_value != null) {
    if (payload.min_value != null && payload.ideal_value < payload.min_value) return 'Just right must be ≥ too low'
    if (payload.max_value != null && payload.ideal_value > payload.max_value) return 'Just right must be ≤ too high'
  }
  return ''
}

async function saveRow(row) {
  targetError.value = ''
  const payload = {
    zone_id: props.zoneId,
    crop_cycle_id: null,
    stage: row.setpoint?.stage ?? null,
    sensor_type: row.sensorType,
    min_value: num(draftValues.value[row.key]?.min_value),
    ideal_value: num(draftValues.value[row.key]?.ideal_value),
    max_value: num(draftValues.value[row.key]?.max_value),
  }
  const err = validatePayload(payload)
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
