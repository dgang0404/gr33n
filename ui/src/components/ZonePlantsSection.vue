<template>
  <div class="space-y-6" data-test="zone-plants-section">
    <ZoneCurrentGrowStrip
      :zone-id="zoneId"
      :farm-id="farmId"
      :zone="zone"
      :cycles="cropCycles"
      @start-grow="emit('start-grow', $event)"
      @harvest="emit('harvest', $event)"
    />

    <div v-if="historyCycles.length" class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <h2 class="text-sm font-semibold text-white mb-3">Past grows in this zone</h2>
      <ul class="space-y-2">
        <li
          v-for="c in historyCycles"
          :key="c.id"
          class="flex items-center justify-between gap-2 text-sm border-b border-zinc-800/60 pb-2 last:border-0 last:pb-0"
        >
          <div class="min-w-0">
            <p class="text-zinc-200 truncate">{{ c.name || c.strain_or_variety || 'Grow' }}</p>
            <p class="text-zinc-600 text-xs">{{ formatHarvest(c) }}</p>
          </div>
          <router-link
            :to="`/crop-cycles/${c.id}/summary`"
            class="text-xs text-green-600 hover:text-green-400 shrink-0"
          >
            Summary →
          </router-link>
        </li>
      </ul>
    </div>

    <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
      <div class="flex items-center justify-between gap-2 mb-3 flex-wrap">
        <h2 class="text-sm font-semibold text-white">Plants in this zone</h2>
        <button
          type="button"
          class="text-xs font-medium px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70"
          data-test="zone-plants-add-plant"
          @click="openCreate"
        >
          + Add plant
        </button>
      </div>

      <p v-if="!zonePlants.length" class="text-zinc-500 text-sm">
        No plants linked to this zone yet — add one from the knowledge base or start a grow above.
      </p>

      <ul v-else class="space-y-2">
        <li
          v-for="p in zonePlants"
          :key="p.id"
          class="flex items-center justify-between gap-2 bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2"
        >
          <div class="min-w-0">
            <p class="text-sm text-zinc-200 truncate">{{ p.display_name }}</p>
            <p v-if="p.variety_or_cultivar" class="text-zinc-600 text-xs">{{ p.variety_or_cultivar }}</p>
          </div>
          <button
            type="button"
            class="text-xs text-green-500 hover:text-green-300 shrink-0"
            @click="emit('start-grow-strain', p)"
          >
            Start grow
          </button>
        </li>
      </ul>

      <router-link
        v-nav-hint="'/zones'"
        :to="{ path: '/zones', query: { tab: 'strains' } }"
        class="inline-block mt-3 text-xs text-green-600 hover:text-green-400"
      >
        All farm plants →
      </router-link>
    </div>

    <div
      v-if="showModal"
      class="fixed inset-0 z-50 bg-black/70 p-4 flex items-center justify-center"
      @click.self="showModal = false"
    >
      <div class="w-full max-w-md bg-zinc-900 border border-zinc-700 rounded-xl p-5 space-y-4">
        <h2 class="text-white font-semibold">New plant</h2>
        <p class="text-[11px] text-zinc-500 leading-relaxed">
          Pick a crop from the knowledge base (EC, watering, light targets). Use
          <router-link to="/settings" class="text-green-500 hover:text-green-400">Settings → Crops &amp; targets</router-link>
          to tune EC for your farm.
        </p>
        <CropLibraryPicker
          v-if="farmId"
          :farm-id="farmId"
          v-model="form.crop_profile_id"
          required
          data-test="zone-plants-crop-picker"
          @select="onCropSelect"
        />
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Your label for this plant *</label>
          <input
            v-model="form.display_name"
            type="text"
            placeholder="e.g. Flower Room Romas (defaults from crop type)"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
            data-test="zone-plants-plant-name"
          />
        </div>
        <div>
          <label class="block text-xs text-zinc-500 mb-1">Variety / cultivar</label>
          <input
            v-model="form.variety_or_cultivar"
            type="text"
            class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white"
          />
        </div>
        <p v-if="formError" class="text-red-400 text-xs">{{ formError }}</p>
        <div class="flex justify-end gap-3">
          <button type="button" class="text-xs text-zinc-400" @click="showModal = false">Cancel</button>
          <button
            type="button"
            class="text-xs px-4 py-1.5 rounded-lg bg-green-700 text-white disabled:opacity-40"
            :disabled="submitting || !form.display_name.trim() || !form.crop_profile_id"
            @click="submitForm"
          >
            {{ submitting ? 'Saving…' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import ZoneCurrentGrowStrip from './ZoneCurrentGrowStrip.vue'
import CropLibraryPicker from './CropLibraryPicker.vue'

const props = defineProps({
  zoneId: { type: Number, required: true },
  farmId: { type: Number, default: null },
  zone: { type: Object, default: null },
  cropCycles: { type: Array, default: () => [] },
  plants: { type: Array, default: () => [] },
})

const emit = defineEmits(['start-grow', 'harvest', 'start-grow-strain', 'plants-updated'])

const store = useFarmStore()
const farmContext = useFarmContextStore()

const showModal = ref(false)
const submitting = ref(false)
const formError = ref('')
const form = ref({ display_name: '', variety_or_cultivar: '', crop_profile_id: null })

const historyCycles = computed(() =>
  props.cropCycles
    .filter((c) => Number(c.zone_id) === props.zoneId && !c.is_active)
    .sort((a, b) => new Date(b.updated_at || b.created_at) - new Date(a.updated_at || a.created_at)),
)

const zonePlantIds = computed(() => {
  const ids = new Set()
  for (const c of props.cropCycles) {
    if (Number(c.zone_id) === props.zoneId && c.plant_id) ids.add(Number(c.plant_id))
  }
  return ids
})

const zonePlants = computed(() =>
  props.plants.filter((p) => zonePlantIds.value.has(Number(p.id))),
)

function formatHarvest(cycle) {
  if (cycle.harvested_at) {
    return `Harvested ${new Date(cycle.harvested_at).toLocaleDateString()}`
  }
  return cycle.is_active ? 'Active' : 'Ended'
}

function openCreate() {
  form.value = { display_name: '', variety_or_cultivar: '', crop_profile_id: null }
  formError.value = ''
  showModal.value = true
}

function onCropSelect(item) {
  if (!item?.display_name) return
  if (!form.value.display_name.trim()) {
    form.value.display_name = item.display_name
  }
}

async function submitForm() {
  const fid = props.farmId || farmContext.farmId
  if (!fid) return
  if (!form.value.crop_profile_id) {
    formError.value = 'Choose a plant type from the knowledge base'
    return
  }
  submitting.value = true
  formError.value = ''
  try {
    await store.createPlant(fid, {
      display_name: form.value.display_name.trim(),
      variety_or_cultivar: form.value.variety_or_cultivar.trim() || null,
      crop_profile_id: form.value.crop_profile_id,
      meta: {},
    })
    showModal.value = false
    emit('plants-updated')
  } catch (e) {
    formError.value = e.response?.data?.error || e.message || 'Failed to save'
  } finally {
    submitting.value = false
  }
}
</script>
