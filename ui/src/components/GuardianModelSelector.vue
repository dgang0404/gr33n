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
          :disabled="!sessionModelOptions.length"
        >
          <option
            value=""
            :disabled="defaultOptionDisabled"
          >
            {{ sessionDefaultOptionLabel }}
          </option>
          <option
            v-for="m in sessionModelOptions"
            :key="m.name"
            :value="m.name"
          >
            {{ modelOptionLabel(m) }}
          </option>
        </select>
      </div>

      <p
        class="text-[10px] leading-snug rounded border px-2 py-1.5"
        :class="props.farmContextActive ? 'border-green-900/40 bg-green-950/20 text-green-200/90' : 'border-zinc-800 bg-zinc-900/40 text-zinc-300'"
        data-test="guardian-mode-banner"
      >
        <span class="font-medium text-zinc-200">{{ modeBannerTitle }}</span>
        {{ modeBannerBody }}
      </p>

      <p
        v-if="effectiveModelSummary"
        class="text-[10px] text-zinc-500"
        data-test="guardian-effective-model"
      >
        This turn:
        <span class="text-zinc-300">{{ effectiveModelSummary }}</span>
      </p>

      <details class="rounded border border-zinc-800 bg-zinc-900/50 px-2 py-1.5 text-[10px] text-zinc-400 leading-snug" data-test="guardian-model-help">
        <summary class="cursor-pointer text-zinc-300 select-none">How models, field guides, and farm context work</summary>
        <div class="mt-2 space-y-2">
          <p data-test="guardian-field-guides-help">
            <span class="text-zinc-300">Field guides</span>
            (<span class="text-zinc-500">docs/field-guides/</span>)
            are curated gr33n docs — crop care, Pi wiring, safety, procedures.
            They are <span class="text-zinc-300">only</span> injected when
            <span class="text-zinc-300">Use farm context</span> is on, a farm is selected,
            and RAG has ingested them (<span class="text-zinc-500">make rag-ingest-field-guides</span>).
            <span class="text-zinc-300">tinyllama</span> and
            <span class="text-zinc-300">phi3:mini</span> both use the same RAG chunks when grounded;
            a larger model (e.g. llama3.1:8b) reads those chunks more reliably but is slower on CPU.
            With farm context <span class="text-zinc-300">off</span>, no field guides — only the model's
            general training (often wrong on niche horticulture).
          </p>
          <ul class="space-y-1.5 list-disc pl-4">
            <li v-if="props.farmContextActive">
              <span class="text-zinc-300">Farm context on</span> — live farm snapshot + RAG
              (field guides, platform docs, farm rows) + read tools. Requires
              {{ GROUNDED_MIN_CONTEXT_WINDOW }}+ ctx; small models are hidden.
            </li>
            <li v-else>
              <span class="text-zinc-300">Quick chat</span> — no farm data, no field guides, no embed model.
              Empty <span class="text-zinc-300">This chat</span> uses
              <span class="text-zinc-300">{{ serverDefault || 'LLM_MODEL' }}</span> from
              <span class="text-zinc-500">.env</span>. Pick
              <span class="text-zinc-300">phi3:mini</span> here for better off-farm answers (slower).
            </li>
            <li>
              <span class="text-zinc-300">This chat</span> — session override for you only.
              <span class="text-zinc-300">Farm default</span> (when farm context on) — saved per farm; click
              <span class="text-zinc-300">Save</span> after changing.
            </li>
            <li>
              <span class="text-zinc-300">chat · fast · general · ctx · cold · loaded</span> — capability,
              speed class, context window, and whether the weights are already in RAM.
            </li>
            <li>
              <span class="text-zinc-300">Embedding model</span> (<span class="text-zinc-500">EMBEDDING_MODEL</span>)
              is separate — used for RAG search when farm context is on, not listed in this dropdown.
            </li>
          </ul>
        </div>
      </details>

      <div v-if="canAdmin && props.farmContextActive && effectiveFarmId" class="flex flex-wrap items-end gap-2 border-t border-zinc-800 pt-2">
        <div class="flex-1 min-w-[8rem]">
          <label class="text-[10px] text-zinc-500 block mb-1" for="guardian-pull-model">
            Pull new model into Ollama (type tag — not in list above)
            <span class="text-amber-300/80">— internet; large models often exceed 10 min (use terminal: ollama pull)</span>
          </label>
          <input
            id="guardian-pull-model"
            v-model="pullName"
            type="text"
            placeholder="e.g. llama3.1:8b (only if not already in dropdown)"
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

      <div v-if="canAdmin && props.farmContextActive && effectiveFarmId" class="flex flex-wrap items-end gap-2">
        <div class="flex-1 min-w-[8rem]">
          <label class="text-[10px] text-zinc-500 block mb-1" for="guardian-farm-model">Farm default</label>
          <select
            id="guardian-farm-model"
            v-model="farmModelDraft"
            class="w-full bg-zinc-900 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200"
            data-test="guardian-farm-model"
            :disabled="saving"
          >
            <option
              value=""
              :disabled="!serverDefaultGroundedUsable"
            >
              {{ farmServerDefaultOptionLabel }}
            </option>
            <option
              v-for="m in groundedCapableModels"
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
          :disabled="saving || !farmDirty || !farmDefaultSaveable"
          @click="saveFarmDefault"
        >
          {{ saving ? 'Saving…' : 'Save' }}
        </button>
      </div>
      <p
        v-else-if="props.farmContextActive && effectiveFarmId"
        class="text-zinc-500"
        data-test="guardian-farm-model-readonly"
      >
        Farm default (saved):
        <span class="text-zinc-300">{{ activeFarmModel || serverDefault || 'server default' }}</span>
      </p>

      <p
        v-if="farmDefaultUnsavedWarning"
        class="text-[10px] text-amber-300/90 leading-snug rounded border border-amber-900/50 bg-amber-950/30 px-2 py-1.5"
        data-test="guardian-farm-unsaved-warning"
      >
        {{ farmDefaultUnsavedWarning }}
      </p>

      <p
        v-if="noGroundedModelsHint"
        class="text-[10px] text-red-300/90 leading-snug rounded border border-red-900/50 bg-red-950/30 px-2 py-1.5"
        data-test="guardian-no-grounded-models"
      >
        {{ noGroundedModelsHint }}
      </p>

      <p
        v-else-if="props.farmContextActive && selectedTrimHint"
        class="text-[10px] text-amber-300/80 leading-snug"
        data-test="guardian-trim-hint"
      >
        {{ selectedTrimHint }}
      </p>

      <p v-if="selectedRuntimeHint" class="text-[10px] text-amber-300/80" data-test="guardian-runtime-hint">
        {{ selectedRuntimeHint }}
      </p>

      <p
        v-if="selectedEvalHint && props.farmContextActive"
        class="text-[10px] text-zinc-400 leading-snug"
        data-test="guardian-eval-hint"
      >
        {{ selectedEvalHint }}
      </p>

      <p class="text-[10px] text-zinc-600 leading-snug">
        Session model applies to your chat only and does not change the farm default.
        Models come from the server Ollama runtime (shared across farms on this host).
        Switching the dropdown does not unload other models from RAM
        (see <span class="text-zinc-500">docs/guardian-ollama-laptop-playbook.md</span>).
        Pull downloads weights once — can take many minutes on slow internet; not per chat.
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
import { useGuardianModels } from '../composables/useGuardianModels'
import {
  GROUNDED_MIN_CONTEXT_WINDOW,
  farmServerDefaultOptionLabel as buildFarmServerDefaultOptionLabel,
  filterGroundedCapableModels,
  findModelByName,
  groundedChatBlockReason,
  pickPreferredGroundedModel,
  resolveEffectiveChatModelName,
  resolvedDefaultBlocksGrounded,
  serverDefaultUsableForGrounded,
  sessionDefaultOptionLabel as buildSessionDefaultOptionLabel,
  effectiveModelSource,
} from '../lib/guardianModelGrounded'

const props = defineProps({
  farmId: { type: [Number, String], default: null },
  /** When true, show a warning if the selected model cannot use farm context. */
  farmContextActive: { type: Boolean, default: false },
})

const sessionModel = defineModel('sessionModel', { type: String, default: '' })

const auth = useAuthStore()
const farmContext = useFarmContextStore()
const farmStore = useFarmStore()

const { models, serverDefault, loading, loadError, loadModels } = useGuardianModels()
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

/** Saved farm default only — must match server and GuardianChatPanel grounded gate. */
const resolvedChatModelName = computed(() =>
  resolveEffectiveChatModelName({
    sessionModel: sessionModel.value,
    farmModel: activeFarmModel.value,
    serverDefault: serverDefault.value,
  }),
)

const sessionModelOptions = computed(() => {
  if (!props.farmContextActive) return models.value
  return filterGroundedCapableModels(models.value)
})

/** Grounded-capable subset — farm default admin and auto-pick only. */
const groundedCapableModels = computed(() => filterGroundedCapableModels(models.value))

const serverDefaultGroundedUsable = computed(() =>
  serverDefaultUsableForGrounded(serverDefault.value, models.value),
)

const defaultOptionDisabled = computed(() =>
  props.farmContextActive &&
  resolvedDefaultBlocksGrounded({
    farmModel: activeFarmModel.value,
    serverDefault: serverDefault.value,
    models: models.value,
  }),
)

const sessionDefaultOptionLabel = computed(() =>
  buildSessionDefaultOptionLabel({
    farmModel: activeFarmModel.value,
    serverDefault: serverDefault.value,
    models: models.value,
  }),
)

const farmServerDefaultOptionLabel = computed(() =>
  buildFarmServerDefaultOptionLabel({
    serverDefault: serverDefault.value,
    models: models.value,
  }),
)

const selectedModelInfo = computed(() => findModelByName(resolvedChatModelName.value, models.value))

const modeBannerTitle = computed(() =>
  props.farmContextActive ? 'Farm context on.' : 'Quick chat (farm context off).',
)

const modeBannerBody = computed(() => {
  if (props.farmContextActive) {
    return (
      ' Answers use this farm\'s live data plus RAG — including field guides ' +
      '(docs/field-guides) when ingested. Same chunks for any grounded model; larger models read them better. ' +
      'Can take many minutes on CPU.'
    )
  }
  return (
    ' General LLM only — no field guides, no farm RAG, no live snapshot. ' +
    'Off-farm horticulture (home garden, forest understory) stays guesswork unless you add your own notes elsewhere.'
  )
})

const effectiveModelSummary = computed(() => {
  const name = resolvedChatModelName.value
  if (!name) return ''
  const src = effectiveModelSource({
    sessionModel: sessionModel.value,
    farmModel: activeFarmModel.value,
    serverDefault: serverDefault.value,
  })
  const rag = props.farmContextActive ? ' · field guides + farm RAG when ingested' : ' · no field guides'
  return `${name} (${src})${rag}`
})

const noGroundedModelsHint = computed(() => {
  if (!props.farmContextActive || groundedCapableModels.value.length) return ''
  return (
    'No grounded-capable models are installed on this server. ' +
    'Pull phi3:mini or llama3.1:8b (terminal: ollama pull phi3:mini), then refresh.'
  )
})

const farmDefaultSaveable = computed(() => {
  if (!farmModelDraft.value) return serverDefaultGroundedUsable.value
  return !!findModelByName(farmModelDraft.value, groundedCapableModels.value)
})

const farmDefaultUnsavedWarning = computed(() => {
  if (!props.farmContextActive || !canAdmin.value || !farmDirty.value) return ''
  const saved = farmModelSaved.value || ''
  const draft = farmModelDraft.value || ''
  if (draft === saved) return ''
  return (
    `Farm default not saved — until you click Save, the farm default stays ${saved || serverDefault.value || 'server default'}. ` +
    `New chats without a session override will use that, not ${draft || 'your selection'}.`
  )
})

const selectedRuntimeHint = computed(() => selectedModelInfo.value?.runtime_hint || '')

const selectedEvalHint = computed(() => evalHintForModel(selectedModelInfo.value))

const selectedTrimHint = computed(() => {
  const m = selectedModelInfo.value
  if (!m || !props.farmContextActive) return ''
  const effective = m.effective_context_window || 0
  const advertised = m.context_window || 0
  if (effective > 0 && effective < 8192 && advertised > effective) {
    return `Grounded prompts trimmed to ${effective} tokens (${m.name} CPU mode).`
  }
  return ''
})

function evalHintForModel(m) {
  if (!m?.eval) {
    return 'Quality: not yet evaluated — run make guardian-eval on the server'
  }
  const e = m.eval
  if (e.eval_status === 'not_evaluated') {
    return 'Quality: not yet evaluated — run make guardian-eval'
  }
  const cite = Math.round((e.grounded_citation_rate || 0) * 100)
  const prop = Math.round((e.proposal_valid_rate || 0) * 100)
  const lat = Math.round(e.mean_latency_ms || 0)
  const repair = e.repair_attempts_avg != null ? ` · repair ${(e.repair_attempts_avg * 100).toFixed(0)}%` : ''
  return `Quality: ${cite}% grounded cite · ${prop}% proposals · ~${lat}ms${repair} (${e.total_questions || '?'} questions)`
}

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

function ensureGroundedSessionModel() {
  if (!props.farmContextActive || !groundedCapableModels.value.length) return
  const blocks = groundedChatBlockReason(selectedModelInfo.value)
  if (!blocks) return
  const pick = pickPreferredGroundedModel(models.value)
  if (pick) sessionModel.value = pick
}

watch(
  () => [
    props.farmContextActive,
    models.value.length,
    activeFarmModel.value,
    serverDefault.value,
  ],
  () => ensureGroundedSessionModel(),
  { immediate: true },
)

async function loadModelsAndSync() {
  await loadModels()
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
    await loadModelsAndSync()
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

loadModelsAndSync()
</script>
