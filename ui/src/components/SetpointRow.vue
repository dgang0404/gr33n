<!--
  SetpointRow — inline editor for a single gr33ncore.zone_setpoints row.

  Phase 20.6 WS4. Used by the dedicated Setpoints page AND by the Zone
  detail page (and later the Crop Cycle detail tab) so the operator
  never has to leave the object they're editing to tune its ideal
  environment. Emits a single `save` or `delete` event — persistence is
  the parent's job so we don't leak API paths into this primitive.
-->
<template>
  <form
    @submit.prevent="onSave"
    class="grid grid-cols-1 md:grid-cols-7 gap-2 items-center bg-zinc-900 border border-zinc-800 rounded-lg p-3"
  >
    <!-- Scope -->
    <select v-if="!fixedScope" v-model="local.scope_kind" class="input-field text-sm md:col-span-1">
      <option value="zone">Zone</option>
      <option value="cycle">Crop cycle</option>
    </select>
    <div v-else class="text-xs text-zinc-400 md:col-span-1">
      <span v-if="local.scope_kind === 'zone'">Zone-scope</span>
      <span v-else>Cycle-scope</span>
    </div>

    <select
      v-if="local.scope_kind === 'zone'"
      v-model.number="local.zone_id"
      class="input-field text-sm md:col-span-1"
      :disabled="fixedScope"
      required
    >
      <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
    </select>
    <select
      v-else
      v-model.number="local.crop_cycle_id"
      class="input-field text-sm md:col-span-1"
      :disabled="fixedScope"
      required
    >
      <option v-for="c in cropCycles" :key="c.id" :value="c.id">{{ c.name }}</option>
    </select>

    <!-- Stage -->
    <select v-model="stageModel" class="input-field text-sm md:col-span-1">
      <option value="">Any stage</option>
      <option v-for="s in stageOptions" :key="s" :value="s">{{ s }}</option>
    </select>

    <!-- Sensor type -->
    <input
      v-model="local.sensor_type"
      type="text"
      placeholder="sensor_type (e.g. dew_point)"
      required
      class="input-field text-sm md:col-span-1"
    />

    <!-- Min / Ideal / Max -->
    <div class="grid grid-cols-3 gap-1 md:col-span-2">
      <input v-model.number="local.min_value" type="number" step="any" placeholder="min" class="input-field text-xs" />
      <input v-model.number="local.ideal_value" type="number" step="any" placeholder="ideal" class="input-field text-xs" />
      <input v-model.number="local.max_value" type="number" step="any" placeholder="max" class="input-field text-xs" />
    </div>

    <!-- Actions -->
    <div class="flex gap-1 md:col-span-1 justify-end">
      <button
        type="submit"
        :disabled="busy"
        class="px-2 py-1 bg-green-700 hover:bg-green-600 text-white text-xs rounded disabled:opacity-50"
      >
        {{ isNew ? 'Add' : 'Save' }}
      </button>
      <button
        v-if="!isNew"
        type="button"
        @click="onDelete"
        :disabled="busy"
        class="px-2 py-1 bg-red-800 hover:bg-red-700 text-white text-xs rounded disabled:opacity-50"
      >
        Delete
      </button>
    </div>

    <p v-if="error" class="md:col-span-7 text-red-400 text-xs">{{ error }}</p>
  </form>
</template>

<script setup>
import { ref, watch, computed } from 'vue'

const props = defineProps({
  value:      { type: Object, required: true },
  zones:      { type: Array,  default: () => [] },
  cropCycles: { type: Array,  default: () => [] },
  stageOptions: {
    type: Array,
    default: () => ['clone', 'seedling', 'early_veg', 'late_veg', 'early_flower', 'mid_flower', 'late_flower', 'harvest', 'dry_cure'],
  },
  busy:       { type: Boolean, default: false },
  fixedScope: { type: Boolean, default: false },
})
const emit = defineEmits(['save', 'delete'])

const local = ref(initLocal(props.value))
watch(() => props.value, v => (local.value = initLocal(v)), { deep: true })

const error = ref('')

const isNew = computed(() => !local.value.id)
const stageModel = computed({
  get: () => local.value.stage ?? '',
  set: (v) => (local.value.stage = v === '' ? null : v),
})

function initLocal(raw) {
  const copy = { ...raw }
  copy.scope_kind = copy.crop_cycle_id ? 'cycle' : 'zone'
  return copy
}

function onSave() {
  error.value = ''
  const payload = {
    sensor_type: (local.value.sensor_type ?? '').trim(),
    stage: local.value.stage || null,
    min_value: num(local.value.min_value),
    max_value: num(local.value.max_value),
    ideal_value: num(local.value.ideal_value),
  }
  if (local.value.scope_kind === 'zone') {
    payload.zone_id = local.value.zone_id
    payload.crop_cycle_id = null
  } else {
    payload.crop_cycle_id = local.value.crop_cycle_id
    payload.zone_id = null
  }

  if (!payload.sensor_type) {
    error.value = 'sensor_type is required'
    return
  }
  if (payload.min_value != null && payload.max_value != null && payload.min_value > payload.max_value) {
    error.value = 'min must be <= max'
    return
  }
  if (payload.ideal_value != null) {
    if (payload.min_value != null && payload.ideal_value < payload.min_value) {
      error.value = 'ideal must be >= min'
      return
    }
    if (payload.max_value != null && payload.ideal_value > payload.max_value) {
      error.value = 'ideal must be <= max'
      return
    }
  }
  emit('save', { id: local.value.id ?? null, payload })
}

function onDelete() {
  if (local.value.id && confirm('Delete this setpoint?')) emit('delete', local.value.id)
}

function num(v) {
  if (v === '' || v == null || Number.isNaN(Number(v))) return null
  return Number(v)
}
</script>
