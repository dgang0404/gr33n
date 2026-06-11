<template>
  <div class="space-y-1" data-test="crop-library-picker">
    <label v-if="label" class="block text-xs text-zinc-500 mb-1">
      {{ label }}<span v-if="required" class="text-red-400"> *</span>
    </label>

    <div v-if="selectedItem" class="flex items-start gap-2 rounded-lg border border-green-900/50 bg-green-950/20 px-3 py-2">
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
      <input
        v-model="query"
        type="search"
        :placeholder="placeholder"
        autocomplete="off"
        class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
        data-test="crop-library-picker-search"
        @focus="open = true"
      />
      <p v-if="countsLabel" class="text-[10px] text-zinc-600">{{ countsLabel }}</p>

      <div
        v-if="open && (loading || filteredGroups.length)"
        class="max-h-56 overflow-y-auto rounded-lg border border-zinc-700 bg-zinc-950 shadow-lg"
        data-test="crop-library-picker-list"
      >
        <p v-if="loading" class="px-3 py-2 text-xs text-zinc-500">Loading crop library…</p>
        <template v-else>
          <div v-for="group in filteredGroups" :key="group.key" class="border-b border-zinc-800/80 last:border-0">
            <p class="px-3 py-1.5 text-[10px] uppercase tracking-widest text-zinc-500 bg-zinc-900/80 sticky top-0">
              {{ group.label }}
            </p>
            <button
              v-for="item in group.items"
              :key="item.crop_key + String(item.crop_profile_id || '')"
              type="button"
              class="w-full text-left px-3 py-2 text-sm border-b border-zinc-900/50 last:border-0 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-zinc-900/80"
              :class="item.has_targets ? 'text-zinc-100' : 'text-zinc-400'"
              :disabled="!item.has_targets"
              :data-test="'crop-library-option-' + item.crop_key"
              @click="selectItem(item)"
            >
              <span class="block truncate">{{ item.display_name }}</span>
              <span class="block text-[10px] text-zinc-500 truncate">{{ pickerItemHint(item) }}</span>
            </button>
          </div>
          <p v-if="!filteredGroups.length" class="px-3 py-2 text-xs text-zinc-500">No crops match.</p>
        </template>
      </div>
    </template>

    <p v-if="error" class="text-red-400 text-[11px]">{{ error }}</p>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import {
  filterPickerGroups,
  findPickerItemByProfileId,
  pickerItemHint,
  pickerItemLabel,
} from '../lib/cropLibraryPicker.js'

const props = defineProps({
  farmId: { type: Number, required: true },
  modelValue: { type: Number, default: null },
  label: { type: String, default: 'Crop type' },
  placeholder: { type: String, default: 'Search tomato, orchid, basil…' },
  required: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'select'])

const store = useFarmStore()
const picker = ref(null)
const loading = ref(false)
const error = ref('')
const query = ref('')
const open = ref(false)

const filteredGroups = computed(() => filterPickerGroups(picker.value, query.value))

const selectedItem = computed(() =>
  findPickerItemByProfileId(picker.value, props.modelValue),
)

const countsLabel = computed(() => {
  const c = picker.value?.counts
  if (!c) return ''
  return `${c.with_targets} with EC targets · ${c.catalog_only} catalog-only`
})

async function loadPicker() {
  if (!props.farmId) return
  loading.value = true
  error.value = ''
  try {
    picker.value = await store.loadCropLibraryPicker(props.farmId)
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load crop library'
    picker.value = null
  } finally {
    loading.value = false
  }
}

function selectItem(item) {
  if (!item?.has_targets || !item.crop_profile_id) return
  emit('update:modelValue', item.crop_profile_id)
  emit('select', item)
  query.value = ''
  open.value = false
}

function clearSelection() {
  emit('update:modelValue', null)
  emit('select', null)
  open.value = true
}

watch(
  () => props.farmId,
  () => loadPicker(),
  { immediate: true },
)

onMounted(loadPicker)
</script>
