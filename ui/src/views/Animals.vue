<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">
        Animals
        <HelpTip position="bottom">
          Animal groups are the head-count + lifecycle anchor for a flock, herd, or pen.
          Feeding, watering, and climate still go through the same primitives you use
          for plants — zones, sensors, actuators, rules, schedules, tasks. The group
          row is what lets you answer "how many, where, and what happened to them"
          without stuffing it into a notes field.
        </HelpTip>
      </h1>
      <div class="flex items-center gap-3">
        <label class="text-xs text-zinc-400 flex items-center gap-2">
          <input type="checkbox" v-model="showArchived" class="accent-green-600" />
          Include archived
        </label>
        <button
          @click="openCreate"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
        >
          + New group
        </button>
        <button @click="refresh" class="text-xs text-zinc-400 hover:text-zinc-200">Refresh</button>
        <span class="text-xs text-zinc-500">{{ visibleGroups.length }} visible</span>
      </div>
    </div>

    <div v-if="loading" class="text-zinc-400 text-sm">Loading animal groups…</div>

    <div
      v-else-if="!visibleGroups.length"
      class="text-zinc-500 text-sm bg-zinc-800 border border-zinc-700 rounded-xl p-8 text-center"
    >
      No animal groups yet. Create one — e.g. "Layer flock" — then log lifecycle events
      (added, born, died, sold, health check) to build a timeline.
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div
        v-for="g in visibleGroups"
        :key="g.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-zinc-700 transition-colors"
      >
        <div class="flex items-start justify-between gap-2 mb-3">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <p class="text-white text-sm font-medium truncate">{{ g.label }}</p>
              <span
                v-if="!g.active"
                class="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-500 border border-zinc-700"
              >archived</span>
            </div>
            <p v-if="g.species" class="text-zinc-500 text-xs mt-0.5">{{ g.species }}</p>
          </div>
          <span class="text-lg shrink-0">🐾</span>
        </div>
        <div class="grid grid-cols-2 gap-3 mb-3 text-xs">
          <div>
            <p class="text-zinc-500">Count</p>
            <p class="text-white font-medium">{{ g.count ?? '—' }}</p>
          </div>
          <div>
            <p class="text-zinc-500">Primary zone</p>
            <p class="text-white truncate">{{ zoneName(g.primary_zone_id) || '—' }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3 border-t border-zinc-800 pt-2">
          <button @click="openDetail(g)" class="text-xs text-green-400 hover:text-green-300">Timeline</button>
          <button @click="openEdit(g)" class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
          <button
            v-if="g.active"
            @click="confirmArchive(g)"
            class="text-xs text-amber-400 hover:text-amber-300"
          >Archive</button>
          <button @click="confirmDelete(g)" class="text-xs text-red-500 hover:text-red-400 ml-auto">
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
        <h2 class="text-white font-semibold">{{ editing ? 'Edit group' : 'New group' }}</h2>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Label *</label>
          <input
            v-model="form.label"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
            placeholder="e.g. Layer flock"
          />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Species</label>
          <input
            v-model="form.species"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
            placeholder="chicken, goat, tilapia, …"
          />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Count</label>
            <input
              v-model.number="form.count"
              type="number"
              min="0"
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
            />
          </div>
          <div>
            <label class="block text-xs text-zinc-500 mb-1">Primary zone</label>
            <select
              v-model="form.primary_zone_id"
              class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-green-600"
            >
              <option :value="null">— none —</option>
              <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.display_name || z.internal_identifier }}</option>
            </select>
          </div>
        </div>
        <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>
        <div class="flex justify-end gap-3 pt-1">
          <button
            @click="showForm = false"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >Cancel</button>
          <button
            @click="submitGroup"
            :disabled="submitting || !form.label.trim()"
            class="px-4 py-1.5 text-xs rounded-lg bg-green-700 hover:bg-green-600 text-white font-medium disabled:opacity-40"
          >{{ submitting ? 'Saving…' : editing ? 'Update' : 'Create' }}</button>
        </div>
      </div>
    </div>

    <!-- Detail / timeline drawer -->
    <div
      v-if="detail"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="closeDetail"
    >
      <div class="w-full max-w-2xl bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-start justify-between">
          <div>
            <h2 class="text-white font-semibold">{{ detail.label }}</h2>
            <p class="text-zinc-500 text-xs">
              {{ detail.species || 'unspecified species' }} ·
              {{ detail.count ?? '—' }} head ·
              {{ zoneName(detail.primary_zone_id) || 'no primary zone' }}
            </p>
            <p v-if="deltaTotal !== null" class="text-zinc-600 text-[11px] mt-1">
              Lifecycle delta sum: {{ deltaTotal }} ·
              stored count {{ detail.count ?? '—' }}
              <span v-if="deltaTotal !== null && detail.count !== null && deltaTotal !== detail.count"
                    class="text-amber-400"> · reconcile?</span>
            </p>
          </div>
          <button @click="closeDetail" class="text-zinc-400 hover:text-zinc-200 text-lg">×</button>
        </div>
        <div class="border border-zinc-800 rounded-lg p-3 bg-zinc-950 space-y-3">
          <p class="text-xs text-zinc-500 uppercase tracking-wide">Log event</p>
          <div class="grid grid-cols-2 gap-3">
            <select
              v-model="eventForm.event_type"
              class="col-span-1 bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            >
              <option v-for="t in eventTypeChoices" :key="t" :value="t">{{ t }}</option>
            </select>
            <input
              v-model.number="eventForm.delta_count"
              type="number"
              placeholder="Δ count (signed)"
              class="col-span-1 bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            />
          </div>
          <textarea
            v-model="eventForm.notes"
            rows="2"
            placeholder="notes (optional)"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          />
          <p v-if="eventError" class="text-red-400 text-xs">{{ eventError }}</p>
          <div class="flex justify-end">
            <button
              @click="submitEvent"
              :disabled="submitting || !eventForm.event_type"
              class="px-3 py-1.5 text-xs rounded-lg bg-green-700 hover:bg-green-600 text-white font-medium disabled:opacity-40"
            >{{ submitting ? 'Saving…' : 'Add event' }}</button>
          </div>
        </div>
        <div class="space-y-2">
          <p class="text-xs text-zinc-500 uppercase tracking-wide">Timeline</p>
          <div v-if="eventsLoading" class="text-xs text-zinc-400">Loading…</div>
          <div v-else-if="!events.length" class="text-xs text-zinc-500">No lifecycle events yet.</div>
          <ul v-else class="divide-y divide-zinc-800 border border-zinc-800 rounded-lg overflow-hidden">
            <li
              v-for="ev in events"
              :key="ev.id"
              class="px-3 py-2 flex items-start gap-3 bg-zinc-950"
            >
              <span
                class="text-[10px] uppercase tracking-wide px-1.5 py-0.5 rounded mt-0.5 shrink-0"
                :class="eventBadgeClass(ev.event_type)"
              >{{ ev.event_type }}</span>
              <div class="flex-1 min-w-0">
                <p class="text-xs text-white">
                  {{ formatEventTime(ev.event_time) }}
                  <span v-if="ev.delta_count !== null" :class="ev.delta_count < 0 ? 'text-red-400' : 'text-green-400'">
                    · {{ ev.delta_count > 0 ? '+' : '' }}{{ ev.delta_count }}
                  </span>
                </p>
                <p v-if="ev.notes" class="text-xs text-zinc-400 mt-0.5">{{ ev.notes }}</p>
              </div>
              <button
                @click="deleteEvent(ev)"
                class="text-[11px] text-red-500 hover:text-red-400"
              >delete</button>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Archive confirm -->
    <div
      v-if="archiveTarget"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="archiveTarget = null"
    >
      <div class="bg-zinc-900 border border-zinc-700 rounded-xl p-6 w-full max-w-sm space-y-4">
        <h3 class="text-white font-semibold">Archive group</h3>
        <p class="text-sm text-zinc-300">
          Archive <span class="text-white font-medium">{{ archiveTarget.label }}</span>?
          Lifecycle history is preserved; the group is hidden by default.
        </p>
        <input
          v-model="archiveReason"
          type="text"
          placeholder="reason (optional)"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
        />
        <div class="flex justify-end gap-3 pt-2">
          <button
            @click="archiveTarget = null"
            class="px-3 py-1.5 text-xs rounded border border-zinc-700 text-zinc-400 hover:text-zinc-200"
          >Cancel</button>
          <button
            @click="doArchive"
            :disabled="submitting"
            class="px-3 py-1.5 text-xs rounded bg-amber-600 hover:bg-amber-500 text-white font-medium disabled:opacity-50"
          >{{ submitting ? 'Archiving…' : 'Archive' }}</button>
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
        <h3 class="text-white font-semibold">Delete group</h3>
        <p class="text-sm text-zinc-300">
          Delete <span class="text-white font-medium">{{ deleteTarget.label }}</span>?
          For normal end-of-flock use <span class="text-amber-400">Archive</span> — delete is
          for mistake-entry cleanup.
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

const groups = ref([])
const loading = ref(false)
const showArchived = ref(false)
const showForm = ref(false)
const editing = ref(null)
const submitting = ref(false)
const formError = ref('')
const deleteTarget = ref(null)
const archiveTarget = ref(null)
const archiveReason = ref('')
const form = ref(emptyGroupForm())

const detail = ref(null)
const events = ref([])
const eventsLoading = ref(false)
const deltaTotal = ref(null)
const eventForm = ref(emptyEventForm())
const eventError = ref('')

const eventTypeChoices = ['added', 'born', 'died', 'sold', 'culled', 'health_event', 'weight_check', 'note']

const visibleGroups = computed(() =>
  showArchived.value ? groups.value : groups.value.filter((g) => g.active),
)

const zones = computed(() => store.zones || [])

function emptyGroupForm() {
  return { label: '', species: '', count: null, primary_zone_id: null }
}

function emptyEventForm() {
  return { event_type: 'note', delta_count: null, notes: '' }
}

function zoneName(id) {
  if (!id) return ''
  const z = zones.value.find((x) => x.id === id)
  return z ? (z.display_name || z.internal_identifier || `Zone ${z.id}`) : ''
}

function openCreate() {
  editing.value = null
  form.value = emptyGroupForm()
  formError.value = ''
  showForm.value = true
}

function openEdit(g) {
  editing.value = g
  form.value = {
    label: g.label || '',
    species: g.species || '',
    count: g.count ?? null,
    primary_zone_id: g.primary_zone_id ?? null,
  }
  formError.value = ''
  showForm.value = true
}

function confirmDelete(g) { deleteTarget.value = g }
function confirmArchive(g) { archiveTarget.value = g; archiveReason.value = '' }

async function refresh() {
  const fid = farmContext.farmId
  if (!fid) return
  loading.value = true
  try {
    groups.value = await store.loadAnimalGroups(fid)
  } finally {
    loading.value = false
  }
}

async function submitGroup() {
  formError.value = ''
  const fid = farmContext.farmId
  if (!fid) { formError.value = 'No farm selected'; return }
  const label = form.value.label.trim()
  if (!label) return
  const payload = {
    label,
    species: form.value.species.trim() || null,
    count: form.value.count == null || form.value.count === '' ? null : Number(form.value.count),
    primary_zone_id: form.value.primary_zone_id ?? null,
  }
  submitting.value = true
  try {
    if (editing.value) {
      await store.updateAnimalGroup(editing.value.id, payload)
    } else {
      await store.createAnimalGroup(fid, payload)
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
    await store.deleteAnimalGroup(deleteTarget.value.id)
    deleteTarget.value = null
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to delete'
  } finally {
    submitting.value = false
  }
}

async function doArchive() {
  submitting.value = true
  try {
    await store.archiveAnimalGroup(archiveTarget.value.id, archiveReason.value.trim() || null)
    archiveTarget.value = null
    archiveReason.value = ''
    await refresh()
  } catch (e) {
    formError.value = e.response?.data?.error || 'Failed to archive'
  } finally {
    submitting.value = false
  }
}

async function openDetail(g) {
  detail.value = g
  events.value = []
  deltaTotal.value = null
  eventForm.value = emptyEventForm()
  eventError.value = ''
  eventsLoading.value = true
  try {
    const [evs, meta] = await Promise.all([
      store.loadLifecycleEvents(g.id),
      store.getAnimalGroup(g.id),
    ])
    events.value = evs
    deltaTotal.value = meta?.delta_total ?? 0
  } finally {
    eventsLoading.value = false
  }
}

function closeDetail() { detail.value = null }

async function submitEvent() {
  eventError.value = ''
  if (!detail.value || !eventForm.value.event_type) return
  submitting.value = true
  try {
    const body = {
      event_type: eventForm.value.event_type,
      delta_count: eventForm.value.delta_count == null || eventForm.value.delta_count === ''
        ? null
        : Number(eventForm.value.delta_count),
      notes: eventForm.value.notes.trim() || null,
    }
    await store.createLifecycleEvent(detail.value.id, body)
    await openDetail(detail.value)
  } catch (e) {
    eventError.value = e.response?.data?.error || e.message || 'Failed to log event'
  } finally {
    submitting.value = false
  }
}

async function deleteEvent(ev) {
  if (!confirm(`Delete event "${ev.event_type}"? Prefer appending a compensating event.`)) return
  submitting.value = true
  try {
    await store.deleteLifecycleEvent(ev.id)
    await openDetail(detail.value)
  } finally {
    submitting.value = false
  }
}

function eventBadgeClass(type) {
  switch (type) {
    case 'added':
    case 'born':
      return 'bg-green-900/40 text-green-300 border border-green-800'
    case 'died':
    case 'culled':
      return 'bg-red-900/30 text-red-400 border border-red-800'
    case 'sold':
      return 'bg-amber-900/30 text-amber-300 border border-amber-800'
    case 'health_event':
    case 'weight_check':
      return 'bg-sky-900/30 text-sky-300 border border-sky-800'
    default:
      return 'bg-zinc-800 text-zinc-400 border border-zinc-700'
  }
}

function formatEventTime(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleString(undefined, {
    month: 'short', day: 'numeric', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
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
