<template>
  <section
    v-if="capabilities.aiEnabled && farmId"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-guardian-model-policy"
  >
    <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
      <span>🧠</span> Guardian inference policy
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Per-farm models for <strong class="text-zinc-400">Farm counsel</strong> (grounded, RAG + snapshot)
      and <strong class="text-zinc-400">Quick chat</strong> (general horticulture, no farm data).
      Server .env <code class="text-zinc-400">LLM_MODEL</code> is the fallback when a slot is empty.
    </p>

    <div v-if="!showEditor" class="text-xs text-zinc-500 space-y-1" data-test="settings-guardian-model-readonly">
      <p>Counsel: <span class="text-zinc-300 font-mono">{{ counselLabel }}</span></p>
      <p>Quick: <span class="text-zinc-300 font-mono">{{ quickLabel }}</span></p>
      <p v-if="timeoutLabel">Grounded timeout: <span class="text-zinc-300">{{ timeoutLabel }}s</span></p>
    </div>

    <div v-else class="space-y-3 text-sm">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div>
          <label class="text-[10px] text-zinc-500 block mb-1" for="settings-counsel-model">Farm counsel model</label>
          <select
            id="settings-counsel-model"
            v-model="counselDraft"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200 text-xs"
            data-test="settings-counsel-model"
            :disabled="saving"
          >
            <option value="">{{ farmServerDefaultOptionLabel }}</option>
            <option v-for="m in groundedCapableModels" :key="'c-' + m.name" :value="m.name">
              {{ m.name }}
            </option>
          </select>
        </div>
        <div>
          <label class="text-[10px] text-zinc-500 block mb-1" for="settings-quick-model">Quick chat model</label>
          <select
            id="settings-quick-model"
            v-model="quickDraft"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200 text-xs"
            data-test="settings-quick-model"
            :disabled="saving"
          >
            <option value="">{{ quickServerDefaultOptionLabel }}</option>
            <option v-for="m in models" :key="'q-' + m.name" :value="m.name">
              {{ m.name }}
            </option>
          </select>
        </div>
      </div>
      <div class="max-w-xs">
        <label class="text-[10px] text-zinc-500 block mb-1" for="settings-grounded-timeout">
          Grounded timeout (seconds, optional)
        </label>
        <input
          id="settings-grounded-timeout"
          v-model.number="timeoutDraft"
          type="number"
          min="60"
          placeholder="env default"
          class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200 text-xs"
          data-test="settings-grounded-timeout"
          :disabled="saving"
        />
      </div>
      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40"
          data-test="settings-guardian-model-save"
          :disabled="saving || !dirty"
          @click="save"
        >
          {{ saving ? 'Saving…' : 'Save model policy' }}
        </button>
      </div>
      <p v-if="saveError" class="text-xs text-red-300/90" role="alert" aria-live="assertive">{{ saveError }}</p>
      <p v-if="saveOk" class="text-xs text-green-400/90" role="status" aria-live="polite">Model policy saved.</p>
    </div>
  </section>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import api from '../api'
import { useAuthStore } from '../stores/auth'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'
import { useGuardianModels } from '../composables/useGuardianModels'
import {
  filterGroundedCapableModels,
  serverDefaultUsableForGrounded,
} from '../lib/guardianModelGrounded'

const props = defineProps({
  isFarmAdmin: { type: Boolean, default: false },
})

const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const farmStore = useFarmStore()
const auth = useAuthStore()
const { models, serverDefault, loadModels } = useGuardianModels()

const counselDraft = ref('')
const quickDraft = ref('')
const timeoutDraft = ref(null)
const counselSaved = ref('')
const quickSaved = ref('')
const timeoutSaved = ref(null)
const saving = ref(false)
const saveError = ref('')
const saveOk = ref(false)
const members = ref([])

const farmId = computed(() => farmContext.farmId)
const groundedCapableModels = computed(() => filterGroundedCapableModels(models.value))

const canAdmin = computed(() => {
  const uid = auth.userId
  const fid = farmId.value
  if (!uid || !fid) return false
  const f = farmStore.farm
  if (f?.owner_user_id && String(f.owner_user_id).toLowerCase() === String(uid).toLowerCase()) return true
  const m = (members.value || []).find((x) => String(x.user_id).toLowerCase() === String(uid).toLowerCase())
  return !!(m && (m.role_in_farm === 'owner' || m.role_in_farm === 'manager'))
})

const showEditor = computed(() => props.isFarmAdmin || canAdmin.value)

const farmServerDefaultOptionLabel = computed(() =>
  serverDefaultUsableForGrounded(serverDefault.value, models.value)
    ? `Server default (${serverDefault.value})`
    : 'Server default (not grounded-capable)',
)

const quickServerDefaultOptionLabel = computed(() =>
  serverDefault.value ? `Server default (${serverDefault.value})` : 'Server default (LLM_MODEL)',
)

const counselLabel = computed(() => counselSaved.value || farmStore.farm?.guardian_counsel_model || farmStore.farm?.guardian_preferred_model || 'server default')
const quickLabel = computed(() => quickSaved.value || farmStore.farm?.guardian_quick_model || 'server default')
const timeoutLabel = computed(() => timeoutSaved.value ?? farmStore.farm?.guardian_grounded_timeout_seconds ?? '')

const dirty = computed(() =>
  counselDraft.value !== (counselSaved.value || '') ||
  quickDraft.value !== (quickSaved.value || '') ||
  timeoutDraft.value !== timeoutSaved.value,
)

function syncFromFarm() {
  const f = farmStore.farm
  if (!f || Number(f.id) !== Number(farmId.value)) return
  counselSaved.value = f.guardian_counsel_model || f.guardian_preferred_model || ''
  quickSaved.value = f.guardian_quick_model || ''
  timeoutSaved.value = f.guardian_grounded_timeout_seconds ?? null
  counselDraft.value = counselSaved.value
  quickDraft.value = quickSaved.value
  timeoutDraft.value = timeoutSaved.value
}

async function loadMembers() {
  if (!farmId.value) return
  try {
    const { data } = await api.get(`/farms/${farmId.value}/members`)
    members.value = data || []
  } catch {
    members.value = []
  }
}

async function save() {
  if (!farmId.value) return
  saving.value = true
  saveError.value = ''
  saveOk.value = false
  try {
    const body = {
      guardian_counsel_model: counselDraft.value || null,
      guardian_quick_model: quickDraft.value || null,
      guardian_grounded_timeout_seconds: timeoutDraft.value > 0 ? timeoutDraft.value : null,
    }
    const { data } = await api.patch(`/farms/${farmId.value}/settings`, body)
    counselSaved.value = data.guardian_counsel_model || ''
    quickSaved.value = data.guardian_quick_model || ''
    timeoutSaved.value = data.guardian_grounded_timeout_seconds ?? null
    if (farmStore.farm && Number(farmStore.farm.id) === Number(farmId.value)) {
      farmStore.farm = {
        ...farmStore.farm,
        guardian_counsel_model: data.guardian_counsel_model,
        guardian_quick_model: data.guardian_quick_model,
        guardian_preferred_model: data.guardian_preferred_model,
        guardian_grounded_timeout_seconds: data.guardian_grounded_timeout_seconds,
      }
    }
    saveOk.value = true
  } catch (e) {
    saveError.value = e.response?.data?.error || e.message || 'Save failed'
  } finally {
    saving.value = false
  }
}

watch(farmId, () => {
  void loadModels()
  void loadMembers()
  syncFromFarm()
}, { immediate: true })

watch(() => farmStore.farm, syncFromFarm)
</script>
