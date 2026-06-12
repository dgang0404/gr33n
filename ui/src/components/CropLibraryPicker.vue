<template>
  <div class="space-y-2" data-test="crop-library-picker">
    <label v-if="label" class="block text-xs text-zinc-500 mb-1">
      {{ label }}<span v-if="required" class="text-red-400"> *</span>
    </label>

    <div
      v-if="degradedBanner"
      class="rounded-lg border border-amber-800/60 bg-amber-950/30 px-3 py-2 text-[11px] text-amber-200/90"
      data-test="crop-library-picker-degraded-banner"
    >
      Knowledge base API outdated — run <code class="text-amber-100">make migrate</code> and restart the API.
      Showing a limited crop list from profiles only (no full catalog).
    </div>

    <p v-if="!loading && !error && countsLabel && !degradedBanner" class="text-[10px] text-zinc-600">
      {{ countsLabel }} — from farm knowledge base (Postgres). Adjust targets in
      <router-link to="/settings" class="text-green-500 hover:text-green-400">Settings → Crops &amp; targets</router-link>.
    </p>

    <div v-if="selectedItem && !useSelectOnly" class="flex items-start gap-2 rounded-lg border border-green-900/50 bg-green-950/20 px-3 py-2">
      <div class="min-w-0 flex-1">
        <p class="text-sm text-zinc-100 truncate">{{ pickerItemLabel(selectedItem) }}</p>
        <p v-if="pickerItemHint(selectedItem)" class="text-[10px] text-zinc-500 mt-0.5">{{ pickerItemHint(selectedItem) }}</p>
      </div>
      <button
        type="button"
        class="text-xs text-zinc-500 hover:text-zinc-300 shrink-0"
        data-test="crop-library-picker-clear"
        @click="clearSelection"
      >
        Change
      </button>
    </div>

    <template v-else>
      <select
        :value="modelValue ?? ''"
        :disabled="loading || !selectOptions.length"
        class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600 disabled:opacity-50"
        data-test="crop-library-picker-select"
        @change="onSelectChange"
      >
        <option value="" disabled>{{ loading ? 'Loading crops…' : 'Choose a crop from the knowledge base…' }}</option>
        <optgroup v-for="group in selectGroups" :key="group.key" :label="group.label">
          <option
            v-for="item in group.items"
            :key="item.crop_key + String(item.crop_profile_id || '')"
            :value="item.crop_profile_id || ''"
            :disabled="!item.has_targets || !item.crop_profile_id"
          >
            {{ optionLabel(item) }}
          </option>
        </optgroup>
      </select>
    </template>

    <div
      v-if="selectedItem && targetLines.length"
      class="rounded-lg border border-zinc-800 bg-zinc-950/80 px-3 py-2 text-[11px] text-zinc-400 space-y-0.5"
      data-test="crop-library-target-preview"
    >
      <p class="text-zinc-500 text-[10px] uppercase tracking-wide mb-1">Feeding &amp; light targets (by stage)</p>
      <p v-for="(line, i) in targetLines" :key="i" class="font-mono text-zinc-300">{{ line }}</p>
      <p v-if="targetTruncated" class="text-zinc-600">…and {{ targetTruncated }} more stages</p>
    </div>

    <p v-if="loading" class="text-zinc-500 text-xs">Loading crop knowledge base…</p>
    <p v-if="error" class="text-red-400 text-[11px]">
      {{ error }}
      <span v-if="error.includes('404')" class="block text-zinc-500 mt-1">
        Restart the API after <code class="text-zinc-400">make migrate</code> so <code class="text-zinc-400">/crop-library/picker</code> is registered.
      </span>
    </p>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import {
  findPickerItemByProfileId,
  formatStageTargetLine,
  pickerItemHint,
  pickerItemLabel,
} from '../lib/cropLibraryPicker.js'

const props = defineProps({
  farmId: { type: Number, required: true },
  modelValue: { type: Number, default: null },
  label: { type: String, default: 'Plant type (knowledge base)' },
  required: { type: Boolean, default: false },
  /** Always show &lt;select&gt; even after selection (modal forms). */
  useSelectOnly: { type: Boolean, default: true },
})

const emit = defineEmits(['update:modelValue', 'select'])

const store = useFarmStore()
const picker = ref(null)
const loading = ref(false)
const error = ref('')
const targetLines = ref([])
const targetTruncated = ref(0)
const profileLoading = ref(false)

const selectGroups = computed(() => {
  const groups = picker.value?.groups || []
  return groups
    .map((g) => ({
      ...g,
      items: (g.items || []).filter((item) => item.has_targets && item.crop_profile_id),
    }))
    .filter((g) => g.items.length > 0)
})

const selectOptions = computed(() =>
  selectGroups.value.flatMap((g) => g.items),
)

const selectedItem = computed(() =>
  findPickerItemByProfileId(picker.value, props.modelValue),
)

const countsLabel = computed(() => {
  const c = picker.value?.counts
  if (!c) return ''
  return `${c.with_targets} crops with EC / DLI / photoperiod targets`
})

const degradedBanner = computed(() => Boolean(picker.value?._degraded))

function optionLabel(item) {
  let s = item.display_name || item.crop_key
  if (item.substrate) s += ` — ${item.substrate}`
  return s
}

async function loadPicker() {
  if (!props.farmId) return
  loading.value = true
  error.value = ''
  try {
    picker.value = await store.loadCropLibraryPicker(props.farmId)
  } catch (e) {
    const msg = e.response?.data?.error || e.message || 'Failed to load crop library'
    error.value = e.response?.status === 404 ? `Knowledge base API not found (404). ${msg}` : msg
    picker.value = null
  } finally {
    loading.value = false
  }
}

async function loadTargetPreview(profileId) {
  targetLines.value = []
  targetTruncated.value = 0
  if (!profileId) return
  profileLoading.value = true
  try {
    const profile = await store.getCropProfile(profileId)
    const stages = profile?.stages || []
    const max = 4
    targetLines.value = stages.slice(0, max).map(formatStageTargetLine).filter(Boolean)
    if (stages.length > max) targetTruncated.value = stages.length - max
  } catch {
    targetLines.value = []
  } finally {
    profileLoading.value = false
  }
}

function onSelectChange(ev) {
  const raw = ev.target.value
  const id = raw ? Number(raw) : null
  if (!id) return
  const item = findPickerItemByProfileId(picker.value, id)
  if (!item?.has_targets) return
  emit('update:modelValue', id)
  emit('select', item)
}

function clearSelection() {
  emit('update:modelValue', null)
  emit('select', null)
  targetLines.value = []
  targetTruncated.value = 0
}

watch(
  () => props.farmId,
  () => loadPicker(),
  { immediate: true },
)

watch(
  () => props.modelValue,
  (id) => {
    if (id) loadTargetPreview(id)
    else {
      targetLines.value = []
      targetTruncated.value = 0
    }
  },
  { immediate: true },
)
</script>
