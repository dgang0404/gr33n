<template>
  <div
    class="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 space-y-2 text-xs"
    data-test="guardian-model-selector"
  >
    <div class="flex flex-wrap items-center justify-between gap-2">
      <p class="text-zinc-400">Guardian model</p>
      <span
        v-if="loading"
        class="text-zinc-600"
      >Loading…</span>
      <span
        v-else-if="loadError"
        class="text-amber-300/80"
      >{{ loadError }}</span>
    </div>

    <div v-if="!loading && !loadError" class="space-y-2">
      <div class="flex flex-wrap items-center gap-2">
        <label class="text-[10px] text-zinc-500 shrink-0" for="guardian-session-model">This chat</label>
        <select
          id="guardian-session-model"
          v-model="sessionModel"
          class="flex-1 min-w-[8rem] bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200"
          data-test="guardian-session-model"
        >
          <option value="">Farm / server default</option>
          <option
            v-for="m in models"
            :key="m.name"
            :value="m.name"
          >
            {{ modelOptionLabel(m) }}
          </option>
        </select>
      </div>

      <div v-if="canAdmin" class="flex flex-wrap items-end gap-2 border-t border-zinc-800 pt-2">
        <div class="flex-1 min-w-[8rem]">
          <label class="text-[10px] text-zinc-500 block mb-1" for="guardian-pull-model">Pull model into Ollama</label>
          <input
            id="guardian-pull-model"
            v-model="pullName"
            type="text"
            placeholder="e.g. tinyllama"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200"
            data-test="guardian-pull-model-input"
            :disabled="pulling"
          >
        </div>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg bg-zinc-800 border border-zinc-700 text-zinc-200 hover:bg-zinc-700 disabled:opacity-40"
          data-test="guardian-pull-model-btn"
          :disabled="pulling || !pullName.trim() || !effectiveFarmId"
          @click="pullModel"
        >
          {{ pulling ? 'Pulling…' : 'Pull' }}
        </button>
      </div>

      <div v-if="canAdmin" class="flex flex-wrap items-end gap-2">
        <div class="flex-1 min-w-[8rem]">
          <label class="text-[10px] text-zinc-500 block mb-1" for="guardian-farm-model">Farm default</label>
          <select
            id="guardian-farm-model"
            v-model="farmModelDraft"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200"
            data-test="guardian-farm-model"
            :disabled="saving"
          >
            <option value="">Server default ({{ serverDefault || 'env' }})</option>
            <option
              v-for="m in models"
              :key="'farm-' + m.name"
              :value="m.name"
            >
              {{ modelOptionLabel(m) }}
            </option>
          </select>
        </div>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg bg-green-900/50 border border-green-800 text-green-200 hover:bg-green-900 disabled:opacity-40"
          data-test="guardian-farm-model-save"
          :disabled="saving || !farmDirty"
          @click="saveFarmDefault"
        >
          {{ saving ? 'Saving…' : 'Save' }}
        </button>
      </div>
      <p
        v-else
        class="text-zinc-500"
        data-test="guardian-farm-model-readonly"
      >
        Farm default:
        <span class="text-zinc-300">{{ activeFarmModel || serverDefault || 'server default' }}</span>
      </p>

      <p v-if="selectedRuntimeHint" class="text-[10px] text-amber-300/80" data-test="guardian-runtime-hint">
        {{ selectedRuntimeHint }}
      </p>

      <p class="text-[10px] text-zinc-600 leading-snug">
        Session model applies to your chat only and does not change the farm default.
        Models come from the server Ollama runtime (shared across farms on this host).
      </p>
      <p v-if="saveError" class="text-[10px] text-red-300/90">{{ saveError }}</p>
      <p v-if="saveOk" class="text-[10px] text-green-400/90">Farm default saved.</p>
      <p v-if="pullError" class="text-[10px] text-red-300/90">{{ pullError }}</p>
      <p v-if="pullOk" class="text-[10px] text-green-400/90">Model pulled — list refreshed.</p>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'

const props = defineProps({
  farmId: { type: [Number, String], default: null },
})

const sessionModel = defineModel('sessionModel', { type: String, default: '' })

const auth = useAuthStore()
const farmContext = useFarmContextStore()
const farmStore = useFarmStore()

const models = ref([])
const serverDefault = ref('')
const loading = ref(false)
const loadError = ref(null)
const farmModelDraft = ref('')
const farmModelSaved = ref('')
const members = ref([])
const saving = ref(false)
const saveError = ref('')
const saveOk = ref(false)
const pullName = ref('')
const pulling = ref(false)
const pullError = ref('')
const pullOk = ref(false)

const effectiveFarmId = computed(() => {
  const id = props.farmId ?? farmContext.farmId
  return id != null ? Number(id) : null
})

const canAdmin = computed(() => {
  const uid = auth.userId
  const fid = effectiveFarmId.value
  if (!uid || !fid) return false
  const f = farmStore.farm
  if (
    f &&
    Number(f.id) === fid &&
    f.owner_user_id &&
    String(f.owner_user_id).toLowerCase() === String(uid).toLowerCase()
  ) {
    return true
  }
  const membersList = members.value || []
  const m = membersList.find((x) => String(x.user_id).toLowerCase() === String(uid).toLowerCase())
  return !!(m && (m.role_in_farm === 'owner' || m.role_in_farm === 'manager'))
})

const activeFarmModel = computed(() => {
  const f = farmStore.farm
  if (!f || Number(f.id) !== effectiveFarmId.value) return ''
  return f.guardian_preferred_model || ''
})

const farmDirty = computed(() => farmModelDraft.value !== (farmModelSaved.value || ''))

const effectiveSessionModel = computed(() => {
  if (sessionModel.value) return sessionModel.value
  if (farmModelDraft.value) return farmModelDraft.value
  return serverDefault.value || ''
})

const selectedModelInfo = computed(() => {
  const name = effectiveSessionModel.value
  if (!name) return null
  return models.value.find((m) => m.name === name || m.name === `${name}:latest` || name === m.name.replace(/:latest$/, '')) || null
})

const selectedRuntimeHint = computed(() => selectedModelInfo.value?.runtime_hint || '')

function capabilityLabel(m) {
  const caps = m.capabilities || []
  if (caps.includes('vision')) return 'vision'
  if (caps.includes('completion')) return 'chat'
  return ''
}

function modelOptionLabel(m) {
  const bits = [m.name]
  const cap = capabilityLabel(m)
  if (cap) bits.push(cap)
  if (m.speed_class) bits.push(m.speed_class)
  if (m.context_window > 0) bits.push(`${m.context_window} ctx`)
  if (m.runtime_hint) bits.push(m.loaded ? 'loaded' : 'cold')
  return bits.join(' · ')
}

async function loadModels() {
  loading.value = true
  loadError.value = null
  try {
    const r = await api.get('/guardian/models')
    models.value = Array.isArray(r.data?.available_models) ? r.data.available_models : []
    serverDefault.value = r.data?.server_default || ''
  } catch (e) {
    loadError.value = e.response?.data?.error || 'Could not load models'
    models.value = []
  } finally {
    loading.value = false
  }
}

async function syncFarmDraft() {
  saveOk.value = false
  saveError.value = ''
  const fid = effectiveFarmId.value
  if (!fid) {
    farmModelDraft.value = ''
    farmModelSaved.value = ''
    return
  }
  try {
    await farmStore.loadFarm(fid)
    members.value = await farmStore.loadFarmMembers(fid)
  } catch { /* farm load best-effort */ }
  const saved = activeFarmModel.value || ''
  farmModelDraft.value = saved
  farmModelSaved.value = saved
}

async function pullModel() {
  const fid = effectiveFarmId.value
  const name = pullName.value.trim()
  if (!fid || !canAdmin.value || !name) return
  pulling.value = true
  pullError.value = ''
  pullOk.value = false
  try {
    await api.post('/guardian/models/pull', { name, farm_id: fid })
    pullOk.value = true
    await loadModels()
  } catch (e) {
    pullError.value = e.response?.data?.error || 'Pull failed'
  } finally {
    pulling.value = false
  }
}

async function saveFarmDefault() {
  const fid = effectiveFarmId.value
  if (!fid || !canAdmin.value) return
  saving.value = true
  saveError.value = ''
  saveOk.value = false
  try {
    const body = {
      guardian_preferred_model: farmModelDraft.value ? farmModelDraft.value : null,
    }
    await api.patch(`/farms/${fid}/settings`, body)
    farmModelSaved.value = farmModelDraft.value
    saveOk.value = true
    await farmStore.loadFarm(fid)
  } catch (e) {
    saveError.value = e.response?.data?.error || 'Save failed'
  } finally {
    saving.value = false
  }
}

watch(effectiveFarmId, () => {
  syncFarmDraft()
}, { immediate: true })

loadModels()
</script>
