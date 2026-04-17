<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">
        Aquaponics loops
        <HelpTip position="bottom">
          A loop pairs a fish-tank zone with a grow-bed zone. The actual pumps, sensors,
          dosers, and rules still live on the zones themselves — the loop row exists so
          reporting and RAG can answer "which tank feeds which bed" in one query.
        </HelpTip>
      </h1>
      <div class="flex items-center gap-3">
        <button
          @click="openCreate"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        >+ New loop</button>
        <button @click="refresh" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
        <span class="text-xs text-zinc-500">{{ loops.length }} loops</span>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading loops…</div>

    <div
      v-else-if="!loops.length"
      class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center"
    >
      No aquaponics loops yet. Create one by linking a fish-tank zone to a grow-bed zone.
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div
        v-for="loop in loops"
        :key="loop.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4"
      >
        <div class="flex items-start justify-between gap-2 mb-3">
          <div>
            <div class="flex items-center gap-2">
              <p class="text-white text-sm font-medium">{{ loop.label }}</p>
              <span
                v-if="!loop.active"
                class="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500 border border-zinc-700"
              >inactive</span>
            </div>
          </div>
          <span class="text-lg">🐟</span>
        </div>
        <div class="grid grid-cols-2 gap-3 text-xs mb-3">
          <div>
            <p class="text-zinc-500">Fish tank zone</p>
            <p class="text-white truncate">{{ zoneName(loop.fish_tank_zone_id) || '—' }}</p>
          </div>
          <div>
            <p class="text-zinc-500">Grow bed zone</p>
            <p class="text-white truncate">{{ zoneName(loop.grow_bed_zone_id) || '—' }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3 border-t border-zinc-800 pt-2">
          <button @click="openEdit(loop)" class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
          <button @click="confirmDelete(loop)" class="text-xs text-red-500 hover:text-red-400 ml-auto">
            Delete
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit modal -->
    <div
      v-if="showForm"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="showForm = false"
    >
      <div class="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4">
        <h2 class="text-white font-semibold">{{ editing ? 'Edit loop' : 'New loop' }}</h2>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Label *</label>
          <input
            v-model="form.label"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
          />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Fish tank zone</label>
          <select
            v-model="form.fish_tank_zone_id"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          >
            <option :value="null">— none —</option>
            <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.display_name || z.internal_identifier }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Grow bed zone</label>
          <select
            v-model="form.grow_bed_zone_id"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          >
            <option :value="null">— none —</option>
            <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.display_name || z.internal_identifier }}</option>
          </select>
        </div>
        <label v-if="editing" class="flex items-center gap-2 text-xs text-zinc-400">
          <input type="checkbox" v-model="form.active" class="accent-green-600" />
          Active
        </label>
        <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-1">
          <button
            @click="showForm = false"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >Cancel</button>
          <button
            @click="submitLoop"
            :disabled="submitting || !form.label.trim()"
            class="px-4 py-1.5 text-xs rounded-lg bg-green-700 hover:bg-green-600 text-white font-medium disabled:opacity-40"
          >{{ submitting ? 'Saving…' : editing ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Delete confirm -->
    <div
      v-if="deleteTarget"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="deleteTarget = null"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-sm space-y-4">
        <h3 class="text-white font-semibold">Delete loop</h3>
        <p class="text-sm text-zinc-300">
          Delete <span class="text-white font-medium">{{ deleteTarget.label }}</span>? The zones
          themselves are not touched.
        </p>
        <div class="flex justify-end gap-3 pt-2">
          <button
            @click="deleteTarget = null"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >Cancel</button>
          <button
            @click="doDelete"
            :disabled="submitting"
            class="px-3 py-1.5 text-xs rounded bg-red-600 hover:bg-red-500 text-white font-medium disabled:opacity-50"
          >{{ submitting ? 'Deleting…' : 'Delete' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import HelpTip from '../components/HelpTip.vue'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const loops = ref([])
const loading = ref(false)
const showForm = ref(false)
const editing = ref(null)
const submitting = ref(false)
const formError = ref('')
const deleteTarget = ref(null)
const form = ref(emptyForm())

const zones = computed(() => store.zones || [])

function emptyForm() {
  return { label: '', fish_tank_zone_id: null, grow_bed_zone_id: null, active: true }
}

function zoneName(id) {
  if (!id) return ''
  const z = zones.value.find((x) => x.id === id)
  return z ? (z.display_name || z.internal_identifier || `Zone ${z.id}`) : ''
}

function openCreate() {
  editing.value = null
  form.value = emptyForm()
  formError.value = ''
  showForm.value = true
}

function openEdit(loop) {
  editing.value = loop
  form.value = {
    label: loop.label || '',
    fish_tank_zone_id: loop.fish_tank_zone_id ?? null,
    grow_bed_zone_id: loop.grow_bed_zone_id ?? null,
    active: loop.active !== false,
  }
  formError.value = ''
  showForm.value = true
}

function confirmDelete(loop) { deleteTarget.value = loop }

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    loops.value = await store.loadAquaponicsLoops(fid)
  } finally {
    loading.value = false
  }
}

async function submitLoop() {
  formError.value = ''
  const fid = farmContext.farmId
  if (!fid) { formError.value = 'No farm selected'; return }
  const label = form.value.label.trim()
  if (!label) return
  const payload = {
    label,
    fish_tank_zone_id: form.value.fish_tank_zone_id ?? null,
    grow_bed_zone_id: form.value.grow_bed_zone_id ?? null,
  }
  submitting.value = true
  try {
    if (editing.value) {
      payload.active = !!form.value.active
      await store.updateAquaponicsLoop(editing.value.id, payload)
    } else {
      await store.createAquaponicsLoop(fid, payload)
    }
    showForm.value = false
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Failed to save'
  } finally {
    submitting.value = false
  }
}

async function doDelete() {
  submitting.value = true
  try {
    await store.deleteAquaponicsLoop(deleteTarget.value.id)
    deleteTarget.value = null
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to delete'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  const fid = farmContext.farmId
  if (fid && !(store.zones && store.zones.length)) {
    await store.loadAll(fid)
  }
  await refresh()
})
watch(() => farmContext.farmId, async (fid) => {
  if (fid) await store.loadAll(fid)
  await refresh()
})
</script>
