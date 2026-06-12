<template>
  <div class="space-y-3">
    <div v-if="showPresets" class="mb-1">
      <label class="block text-xs text-zinc-400 font-medium mb-1.5 uppercase tracking-wide">Start from preset</label>
      <div class="flex flex-wrap gap-1.5">
        <button
          v-for="p in presets"
          :key="p.key"
          type="button"
          class="px-2.5 py-1 text-xs rounded-full border"
          :class="form.presetKey === p.key
            ? 'border-gr33n-500 bg-gr33n-900/40 text-gr33n-300'
            : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
          @click="$emit('pick-preset', p)"
        >{{ p.label }}</button>
      </div>
    </div>

    <div>
      <label class="block text-xs text-zinc-400 font-medium mb-1">Name *</label>
      <input
        v-model="form.name"
        type="text"
        class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500"
      />
    </div>

    <div v-if="showZoneSelect">
      <label class="block text-xs text-zinc-400 font-medium mb-1">Zone *</label>
      <select
        v-model="form.zoneId"
        class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500"
      >
        <option value="">— select zone —</option>
        <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
      </select>
    </div>

    <div>
      <label class="block text-xs text-zinc-400 font-medium mb-1">Grow light actuator *</label>
      <select
        v-model="form.actuatorId"
        class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500"
      >
        <option value="">— select actuator —</option>
        <option v-for="a in actuators" :key="a.id" :value="a.id">{{ a.name }}</option>
      </select>
    </div>

    <div>
      <label class="block text-xs text-zinc-400 font-medium mb-1.5">Photoperiod</label>
      <PhotoperiodClockEditor
        v-model:model-lights-on-at="form.lightsOnAt"
        v-model:model-on-hours="form.onHours"
        :timezone="form.timezone"
        :presets="presets"
        @change="$emit('clock-change', $event)"
      />
    </div>

    <div>
      <label class="block text-xs text-zinc-400 font-medium mb-1">Timezone</label>
      <input
        v-model="form.timezone"
        type="text"
        class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500"
      />
    </div>

    <div>
      <label class="block text-xs text-zinc-400 font-medium mb-1">Description</label>
      <textarea
        v-model="form.description"
        rows="2"
        class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:border-gr33n-500 resize-none"
      />
    </div>

    <label class="flex items-center gap-2 cursor-pointer">
      <input v-model="form.isActive" type="checkbox" class="accent-gr33n-500" />
      <span class="text-sm text-zinc-300">Active (enable schedules immediately)</span>
    </label>
  </div>
</template>

<script setup>
import PhotoperiodClockEditor from './PhotoperiodClockEditor.vue'

defineProps({
  form: { type: Object, required: true },
  presets: { type: Array, default: () => [] },
  actuators: { type: Array, default: () => [] },
  zones: { type: Array, default: () => [] },
  showZoneSelect: { type: Boolean, default: true },
  showPresets: { type: Boolean, default: true },
})

defineEmits(['pick-preset', 'clock-change'])
</script>
