<template>
  <div class="photoperiod-clock-editor">
    <!-- Preset chips -->
    <div class="preset-chips" v-if="showPresets && presetChips.length">
      <span class="preset-label">Presets:</span>
      <button
        v-for="p in presetChips"
        :key="p.key"
        class="chip"
        :class="{ active: activePresetKey === p.key }"
        type="button"
        :aria-pressed="activePresetKey === p.key"
        @click="applyPreset(p)"
      >{{ p.label }}</button>
    </div>

    <!-- Three linked fields -->
    <div class="clock-fields" role="group" aria-label="Photoperiod schedule">
      <div class="field">
        <label for="photoperiod-lights-on">Lights ON</label>
        <input
          id="photoperiod-lights-on"
          type="time"
          :value="lightsOnAt"
          @change="onStartChange($event.target.value)"
          class="time-input"
        />
        <span class="hint">anchor time</span>
      </div>

      <div class="field">
        <label for="photoperiod-duration">Duration (hours)</label>
        <input
          id="photoperiod-duration"
          type="number"
          :value="onHours"
          min="1"
          max="24"
          step="1"
          :aria-describedby="errorMsg ? 'photoperiod-error' : undefined"
          @change="onDurationChange(Number($event.target.value))"
          class="num-input"
        />
        <span class="hint">on period</span>
      </div>

      <div class="field">
        <label for="photoperiod-lights-off">Lights OFF</label>
        <input
          id="photoperiod-lights-off"
          type="time"
          :value="lightsOffAt"
          @change="onEndChange($event.target.value)"
          class="time-input"
        />
        <span class="hint">derived from start + duration</span>
      </div>
    </div>

    <!-- Summary bar -->
    <div class="summary-bar" v-if="onHours > 0">
      <span class="pill on">{{ onHours }}h ON</span>
      <span class="pill off">{{ offHours }}h OFF</span>
      <span class="next-label" v-if="nextOnLabel">{{ nextOnLabel }}</span>
    </div>

    <p id="photoperiod-error" class="error-msg" v-if="errorMsg" role="alert">{{ errorMsg }}</p>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  /** "HH:MM" 24-hour string */
  modelLightsOnAt: { type: String, default: '06:00' },
  modelOnHours:    { type: Number, default: 18 },
  timezone:        { type: String, default: 'UTC' },
  showPresets:     { type: Boolean, default: true },
  /** Phase 89 — from GET /lighting-programs/presets ({ key, label, onHours }). */
  presets:         { type: Array, default: () => [] },
})

const emit = defineEmits(['update:modelLightsOnAt', 'update:modelOnHours', 'change'])

// ── internal state ───────────────────────────────────────────────────────────

const lightsOnAt = ref(props.modelLightsOnAt)
const onHours    = ref(props.modelOnHours)
const errorMsg   = ref('')

const presetChips = computed(() =>
  (props.presets || []).filter((p) => p.key && p.onHours != null),
)

// ── computed ─────────────────────────────────────────────────────────────────

const offHours = computed(() => 24 - onHours.value)

/** Derive lights-off time as HH:MM */
const lightsOffAt = computed(() => {
  const [h, m] = parseHHMM(lightsOnAt.value)
  const totalMins = (h * 60 + m + onHours.value * 60) % (24 * 60)
  return toHHMM(Math.floor(totalMins / 60), totalMins % 60)
})

const activePresetKey = computed(() => {
  const match = presetChips.value.find(p => p.onHours === onHours.value)
  return match?.key ?? null
})

/** Plain-language hint for next ON time (uses browser-local approx if timezone matches) */
const nextOnLabel = computed(() => {
  if (!lightsOnAt.value) return ''
  return `Next ON at ${lightsOnAt.value} (${props.timezone})`
})

// ── sync props → internal ────────────────────────────────────────────────────

watch(() => props.modelLightsOnAt, v => { lightsOnAt.value = v })
watch(() => props.modelOnHours,    v => { onHours.value = v })

// ── helpers ──────────────────────────────────────────────────────────────────

function parseHHMM(s) {
  const [h, m] = (s ?? '').split(':').map(Number)
  return [isNaN(h) ? 0 : h, isNaN(m) ? 0 : m]
}

function toHHMM(h, m) {
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`
}

function emit_change() {
  emit('update:modelLightsOnAt', lightsOnAt.value)
  emit('update:modelOnHours', onHours.value)
  emit('change', { lightsOnAt: lightsOnAt.value, onHours: onHours.value, offHours: offHours.value })
}

// ── user interactions ─────────────────────────────────────────────────────────

function onStartChange(val) {
  errorMsg.value = ''
  lightsOnAt.value = val
  emit_change()
}

function onDurationChange(val) {
  errorMsg.value = ''
  if (val < 1 || val > 24) { errorMsg.value = 'Duration must be 1-24 hours'; return }
  onHours.value = val
  emit_change()
}

/** End time edited → back-compute duration = end - start (mod 24h) */
function onEndChange(val) {
  errorMsg.value = ''
  const [startH, startM] = parseHHMM(lightsOnAt.value)
  const [endH, endM]     = parseHHMM(val)
  let diffMins = (endH * 60 + endM) - (startH * 60 + startM)
  if (diffMins <= 0) diffMins += 24 * 60
  const newOnHours = Math.round(diffMins / 60)
  if (newOnHours < 1 || newOnHours > 24) { errorMsg.value = 'Duration out of range'; return }
  onHours.value = newOnHours
  emit_change()
}

function applyPreset(p) {
  onHours.value = p.onHours
  emit_change()
}
</script>

<style scoped>
.photoperiod-clock-editor {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.preset-chips {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.preset-label {
  font-size: 0.75rem;
  color: var(--color-text-muted, #6b7280);
  font-weight: 500;
}

.chip {
  padding: 0.2rem 0.65rem;
  border-radius: 9999px;
  font-size: 0.78rem;
  font-weight: 500;
  border: 1px solid #d1d5db;
  background: #f9fafb;
  color: #374151;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}

.chip:hover {
  background: #e5e7eb;
}

.chip.active {
  background: #dcfce7;
  border-color: #16a34a;
  color: #15803d;
}

.clock-fields {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 0.75rem;
  align-items: start;
}

@media (max-width: 480px) {
  .clock-fields {
    grid-template-columns: 1fr;
  }
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.field label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #374151;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.field .hint {
  font-size: 0.7rem;
  color: #9ca3af;
}

.time-input,
.num-input {
  padding: 0.45rem 0.6rem;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  background: #fff;
  color: #111827;
  width: 100%;
  box-sizing: border-box;
}

.time-input:focus,
.num-input:focus {
  outline: none;
  border-color: #16a34a;
  box-shadow: 0 0 0 2px rgba(22, 163, 74, 0.2);
}

.summary-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.pill {
  padding: 0.2rem 0.6rem;
  border-radius: 9999px;
  font-size: 0.78rem;
  font-weight: 600;
}

.pill.on  { background: #dcfce7; color: #15803d; }
.pill.off { background: #fef3c7; color: #92400e; }

.next-label {
  font-size: 0.78rem;
  color: #6b7280;
  margin-left: auto;
}

.error-msg {
  font-size: 0.8rem;
  color: #dc2626;
  margin: 0;
}
</style>
