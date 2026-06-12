<template>
  <section
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-crop-targets"
  >
    <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
      <span>🌱</span> Crops &amp; targets
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Platform builtins define default EC, VPD, and DLI targets. Farm admins can
      override a crop&rsquo;s stages for this site — Guardian uses the override
      on the next chat turn (no re-ingest). Units: EC in mS/cm, VPD in kPa, DLI in mol/m²/day.
    </p>

    <div v-if="!canEdit" class="text-amber-200/90 text-xs mb-3">
      View only — farm owner or manager required to edit overrides.
    </div>

    <div class="flex flex-wrap gap-2 mb-3">
      <input
        v-model="filter"
        type="search"
        placeholder="Filter crops…"
        class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white min-w-[12rem] flex-1 max-w-md"
        data-test="crop-targets-filter"
      />
      <button
        type="button"
        class="text-xs px-3 py-1.5 rounded-lg bg-zinc-900 border border-zinc-700 text-zinc-300 hover:bg-zinc-800 disabled:opacity-40"
        :disabled="loading"
        @click="load"
      >
        {{ loading ? 'Loading…' : 'Refresh' }}
      </button>
    </div>

    <div v-if="loading && !rows.length" class="text-zinc-500 text-sm">Loading crop profiles…</div>
    <div v-else-if="error" class="text-red-400 text-xs">{{ error }}</div>
    <div v-else-if="!filteredRows.length" class="text-zinc-600 text-sm">No crops match your filter.</div>

    <div v-else class="overflow-x-auto border border-zinc-700 rounded-lg max-h-96 overflow-y-auto">
      <table class="w-full text-left text-xs">
        <thead class="bg-zinc-900 text-zinc-500 sticky top-0">
          <tr>
            <th class="px-3 py-2 font-medium">Crop</th>
            <th class="px-3 py-2 font-medium">Key</th>
            <th class="px-3 py-2 font-medium">Source</th>
            <th class="px-3 py-2 font-medium text-right">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-800">
          <tr
            v-for="row in filteredRows"
            :key="row.crop_key"
            class="text-zinc-300 hover:bg-zinc-900/60"
            :data-test="'crop-target-row-' + row.crop_key"
          >
            <td class="px-3 py-2">{{ row.display_name }}</td>
            <td class="px-3 py-2 font-mono text-zinc-500">{{ row.crop_key }}</td>
            <td class="px-3 py-2">
              <span
                v-if="row.isOverride"
                class="px-1.5 py-0.5 rounded bg-green-950/60 border border-green-900 text-green-300"
              >Farm override</span>
              <span
                v-else
                class="px-1.5 py-0.5 rounded bg-zinc-900 border border-zinc-700 text-zinc-400"
              >Built-in</span>
            </td>
            <td class="px-3 py-2 text-right whitespace-nowrap">
              <button
                type="button"
                class="text-green-400 hover:text-green-300 mr-2"
                @click="openEditor(row.crop_key)"
              >
                {{ canEdit ? (row.isOverride ? 'Edit' : 'Customize') : 'View' }}
              </button>
              <button
                v-if="canEdit && row.isOverride"
                type="button"
                class="text-red-400 hover:text-red-300 disabled:opacity-40"
                :disabled="busyKey === row.crop_key"
                @click="resetOverride(row.crop_key)"
              >
                Reset
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <p v-if="message" class="mt-2 text-xs text-emerald-400">{{ message }}</p>

    <div
      v-if="editorOpen"
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/70 p-4"
      data-test="crop-targets-editor"
      @click.self="closeEditor"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col shadow-xl">
        <header class="px-4 py-3 border-b border-zinc-800 flex items-start justify-between gap-3">
          <div>
            <h3 class="text-white font-semibold">{{ editorTitle }}</h3>
            <p class="text-zinc-500 text-xs font-mono">{{ editorCropKey }}</p>
          </div>
          <button type="button" class="text-zinc-500 hover:text-white text-lg leading-none" @click="closeEditor">×</button>
        </header>

        <div class="overflow-auto flex-1 p-4">
          <div v-if="editorLoading" class="text-zinc-500 text-sm">Loading stages…</div>
          <div v-else-if="editorError" class="text-red-400 text-xs">{{ editorError }}</div>
          <template v-else>
            <div v-if="canEdit" class="mb-3 max-w-md">
              <label class="text-zinc-500 text-[11px] uppercase tracking-wide">Display name</label>
              <input
                v-model="editorDisplayName"
                type="text"
                class="mt-1 w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white"
              />
            </div>
            <div class="overflow-x-auto rounded-lg border border-zinc-800">
              <table class="w-full text-xs text-left">
                <thead class="bg-zinc-950 text-zinc-500 uppercase tracking-wider">
                  <tr>
                    <th class="px-2 py-2">Stage</th>
                    <th class="px-2 py-2">EC min</th>
                    <th class="px-2 py-2">EC target</th>
                    <th class="px-2 py-2">EC max</th>
                    <th class="px-2 py-2">VPD min</th>
                    <th class="px-2 py-2">VPD max</th>
                    <th class="px-2 py-2">DLI</th>
                    <th class="px-2 py-2">Notes</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(st, idx) in editorStages"
                    :key="st.stage"
                    class="border-t border-zinc-800 text-zinc-300"
                  >
                    <td class="px-2 py-1.5 capitalize whitespace-nowrap">{{ stageLabel(st.stage) }}</td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].ec_min"
                        type="number"
                        step="0.01"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white disabled:opacity-70"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].ec_target"
                        type="number"
                        step="0.01"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].ec_max"
                        type="number"
                        step="0.01"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].vpd_min_kpa"
                        type="number"
                        step="0.01"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].vpd_max_kpa"
                        type="number"
                        step="0.01"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model.number="editorStages[idx].dli_target"
                        type="number"
                        step="0.1"
                        :readonly="!canEdit"
                        class="w-16 bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 font-mono text-white"
                      />
                    </td>
                    <td class="px-2 py-1">
                      <input
                        v-model="editorStages[idx].notes"
                        type="text"
                        :readonly="!canEdit"
                        class="min-w-[8rem] bg-zinc-950 border border-zinc-700 rounded px-1 py-0.5 text-white"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p class="text-zinc-600 text-[10px] mt-2">EC and VPD columns use mS/cm and kPa. DLI is mol/m²/day.</p>
          </template>
        </div>

        <footer class="px-4 py-3 border-t border-zinc-800 flex flex-wrap gap-2 justify-end">
          <button
            type="button"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:bg-zinc-800"
            @click="closeEditor"
          >
            Close
          </button>
          <button
            v-if="canEdit"
            type="button"
            class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40"
            :disabled="editorSaving || editorLoading"
            data-test="crop-targets-save"
            @click="saveEditor"
          >
            {{ editorSaving ? 'Saving…' : 'Save override' }}
          </button>
        </footer>
      </div>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useFarmStore } from '../stores/farm.js'

const props = defineProps({
  farmId: { type: [Number, String], required: true },
  canEdit: { type: Boolean, default: false },
})

const store = useFarmStore()
const profiles = ref([])
const loading = ref(false)
const error = ref('')
const message = ref('')
const filter = ref('')
const busyKey = ref('')

const editorOpen = ref(false)
const editorCropKey = ref('')
const editorTitle = ref('')
const editorDisplayName = ref('')
const editorStages = ref([])
const editorLoading = ref(false)
const editorSaving = ref(false)
const editorError = ref('')

function toNum(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function stageLabel(stage) {
  return String(stage || '').replace(/_/g, ' ')
}

function normalizeStage(st) {
  return {
    stage: st.stage,
    ec_min: toNum(st.ec_min),
    ec_target: toNum(st.ec_target),
    ec_max: toNum(st.ec_max),
    ph_min: toNum(st.ph_min),
    ph_max: toNum(st.ph_max),
    vpd_min_kpa: toNum(st.vpd_min_kpa),
    vpd_max_kpa: toNum(st.vpd_max_kpa),
    temp_min_c: toNum(st.temp_min_c),
    temp_max_c: toNum(st.temp_max_c),
    rh_min_pct: toNum(st.rh_min_pct),
    rh_max_pct: toNum(st.rh_max_pct),
    dli_target: toNum(st.dli_target),
    photoperiod_hrs: toNum(st.photoperiod_hrs),
    notes: st.notes || '',
  }
}

const rows = computed(() => {
  const byKey = new Map()
  for (const p of profiles.value) {
    const key = p.crop_key
    const isOverride = !p.is_builtin && p.farm_id != null
    const existing = byKey.get(key)
    if (!existing || isOverride) {
      byKey.set(key, {
        crop_key: key,
        display_name: p.display_name,
        isOverride,
      })
    }
  }
  return [...byKey.values()].sort((a, b) => a.display_name.localeCompare(b.display_name))
})

const filteredRows = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return rows.value
  return rows.value.filter(
    (r) => r.display_name.toLowerCase().includes(q) || r.crop_key.toLowerCase().includes(q),
  )
})

async function load() {
  if (!props.farmId) return
  loading.value = true
  error.value = ''
  message.value = ''
  try {
    profiles.value = await store.loadCropProfiles(props.farmId)
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Could not load crop profiles'
    profiles.value = []
  } finally {
    loading.value = false
  }
}

async function openEditor(cropKey) {
  editorCropKey.value = cropKey
  editorOpen.value = true
  editorLoading.value = true
  editorError.value = ''
  editorStages.value = []
  try {
    const profile = await store.getCropProfileByKey(props.farmId, cropKey)
    editorTitle.value = profile.display_name
    editorDisplayName.value = profile.display_name
    editorStages.value = (profile.stages || []).map(normalizeStage)
  } catch (e) {
    editorError.value = e?.response?.data?.error || e?.message || 'Could not load profile'
  } finally {
    editorLoading.value = false
  }
}

function closeEditor() {
  editorOpen.value = false
  editorCropKey.value = ''
  editorError.value = ''
}

async function saveEditor() {
  if (!props.canEdit || !editorCropKey.value) return
  editorSaving.value = true
  editorError.value = ''
  try {
    const stages = editorStages.value.map((st) => ({
      stage: st.stage,
      ec_min: toNum(st.ec_min),
      ec_target: toNum(st.ec_target),
      ec_max: toNum(st.ec_max),
      ph_min: toNum(st.ph_min),
      ph_max: toNum(st.ph_max),
      vpd_min_kpa: toNum(st.vpd_min_kpa),
      vpd_max_kpa: toNum(st.vpd_max_kpa),
      temp_min_c: toNum(st.temp_min_c),
      temp_max_c: toNum(st.temp_max_c),
      rh_min_pct: toNum(st.rh_min_pct),
      rh_max_pct: toNum(st.rh_max_pct),
      dli_target: toNum(st.dli_target),
      photoperiod_hrs: toNum(st.photoperiod_hrs),
      notes: st.notes?.trim() ? st.notes.trim() : null,
    }))
    await store.upsertCropProfileOverride(props.farmId, editorCropKey.value, {
      display_name: editorDisplayName.value.trim() || editorTitle.value,
      stages,
    })
    message.value = `Saved override for ${editorCropKey.value}.`
    closeEditor()
    await load()
  } catch (e) {
    editorError.value = e?.response?.data?.error || e?.message || 'Save failed'
  } finally {
    editorSaving.value = false
  }
}

async function resetOverride(cropKey) {
  if (!props.canEdit) return
  if (!window.confirm(`Remove farm override for ${cropKey}? Builtin targets will apply again.`)) return
  busyKey.value = cropKey
  message.value = ''
  error.value = ''
  try {
    await store.deleteCropProfileOverride(props.farmId, cropKey)
    message.value = `Reset ${cropKey} to built-in targets.`
    await load()
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Reset failed'
  } finally {
    busyKey.value = ''
  }
}

onMounted(load)
watch(() => props.farmId, load)
</script>
