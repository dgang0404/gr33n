<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-xl font-semibold text-white">Zones</h1>
      <button @click="showCreateForm = true"
        class="px-3 py-1.5 bg-green-700 hover:bg-green-600 text-white text-xs rounded-lg">
        + New Zone
      </button>
    </div>

    <!-- Create / Edit modal -->
    <div v-if="showCreateForm || editZone" class="mb-6 bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h2 class="text-white text-sm font-semibold mb-3">{{ editZone ? 'Edit Zone' : 'New Zone' }}</h2>
      <form @submit.prevent="submitZone" class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <input v-model="zoneForm.name" placeholder="Zone name" required
          class="input-field" />
        <input v-model="zoneForm.description" placeholder="Description"
          class="input-field" />
        <select v-model="zoneForm.zone_type" class="input-field">
          <option value="">Select type</option>
          <option v-for="t in zoneTypes" :key="t" :value="t">{{ t }}</option>
        </select>
        <input v-model.number="zoneForm.area_sqm" type="number" step="0.1" min="0"
          placeholder="Area (m²)" class="input-field" />
        <div class="flex gap-2 sm:col-span-2">
          <button type="submit" :disabled="saving"
            class="px-4 py-2 bg-green-700 hover:bg-green-600 text-white text-sm rounded-lg disabled:opacity-50">
            {{ saving ? 'Saving…' : (editZone ? 'Update' : 'Create') }}
          </button>
          <button type="button" @click="cancelForm"
            class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm rounded-lg">
            Cancel
          </button>
        </div>
      </form>
    </div>

    <div v-if="store.loading" class="text-zinc-400 text-sm">Loading zones…</div>
    <div v-else-if="!store.zones.length" class="text-zinc-500 text-sm">No zones found.</div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="zone in store.zones"
        :key="zone.id"
        class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 hover:border-green-700 transition-colors group"
      >
        <router-link :to="`/zones/${zone.id}`" class="block">
          <div class="flex items-start justify-between mb-3">
            <span class="text-white font-medium group-hover:text-green-400 transition-colors">
              {{ zone.name }}
            </span>
            <span :class="zoneBadge(zone.zone_type)" class="text-xs font-medium px-2 py-0.5 rounded-full capitalize">
              {{ zone.zone_type || 'unknown' }}
            </span>
          </div>

          <p v-if="zone.description" class="text-zinc-500 text-sm mb-3 line-clamp-2">
            {{ zone.description }}
          </p>

          <div class="flex items-center gap-4 text-xs text-zinc-400">
            <span>🌡 {{ store.sensorsByZone(zone.id).length }} sensors</span>
            <span>⚙️ {{ store.devicesByZone(zone.id).length }} devices</span>
            <span v-if="zone.area_sqm">📐 {{ zone.area_sqm }} m²</span>
          </div>
        </router-link>
        <div class="flex gap-2 mt-3 pt-3 border-t border-zinc-800">
          <button @click.prevent="startEdit(zone)"
            class="text-xs text-zinc-400 hover:text-zinc-200">Edit</button>
          <button @click.prevent="confirmDelete(zone)"
            class="text-xs text-red-500 hover:text-red-400">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'

const store = useFarmStore()
const farmContext = useFarmContextStore()

const showCreateForm = ref(false)
const editZone = ref(null)
const saving = ref(false)
const zoneForm = ref({ name: '', description: '', zone_type: '', area_sqm: null })
const zoneTypes = ['indoor', 'outdoor', 'greenhouse', 'nursery', 'seedling', 'veg', 'flower', 'storage']

onMounted(() => {
  if (!store.zones.length && farmContext.farmId) store.loadAll(farmContext.farmId)
})

function startEdit(zone) {
  editZone.value = zone
  zoneForm.value = { name: zone.name, description: zone.description || '', zone_type: zone.zone_type || '', area_sqm: zone.area_sqm }
  showCreateForm.value = false
}

function cancelForm() {
  showCreateForm.value = false
  editZone.value = null
  zoneForm.value = { name: '', description: '', zone_type: '', area_sqm: null }
}

async function submitZone() {
  saving.value = true
  try {
    if (editZone.value) {
      await store.updateZone(editZone.value.id, zoneForm.value)
    } else {
      await store.createZone(farmContext.farmId, zoneForm.value)
    }
    cancelForm()
  } finally { saving.value = false }
}

async function confirmDelete(zone) {
  if (!confirm(`Delete zone "${zone.name}"?`)) return
  await store.deleteZone(zone.id)
}

const BADGE = {
  indoor:     'bg-indigo-900/60 text-indigo-300',
  outdoor:    'bg-emerald-900/60 text-emerald-300',
  greenhouse: 'bg-green-900/60 text-green-300',
  nursery:    'bg-yellow-900/60 text-yellow-300',
  seedling:   'bg-lime-900/60 text-lime-300',
  veg:        'bg-teal-900/60 text-teal-300',
  flower:     'bg-pink-900/60 text-pink-300',
  storage:    'bg-zinc-700/60 text-zinc-300',
}
function zoneBadge(type) {
  if (!type) return 'bg-zinc-800 text-zinc-400'
  const k = type.toLowerCase()
  for (const [name, cls] of Object.entries(BADGE)) {
    if (k.includes(name)) return cls
  }
  return 'bg-zinc-800 text-zinc-400'
}
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
