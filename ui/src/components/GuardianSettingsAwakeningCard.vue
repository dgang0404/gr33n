<template>
  <section
    v-if="capabilities.aiEnabled"
    class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5"
    data-test="settings-guardian-awakening"
  >
    <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
      <span>✨</span> Farm Guardian readiness
    </h2>
    <p class="text-xs text-zinc-500 mb-4 leading-relaxed">
      Awakening preloads the counsel model so morning checks and farm-grounded chat start without manual
      <code class="text-zinc-400">ollama stop</code> rituals. Login triggers background warmup; use
      <strong class="text-zinc-400">Awaken now</strong> after reboot or model changes. On solar or
      battery sites, use <strong class="text-zinc-400">Rest now</strong> to release the model's RAM/CPU
      draw between sessions — Guardian wakes back up the next time you ask it something.
    </p>

    <div v-if="!readiness.loaded && readiness.loading" class="text-zinc-500 text-sm">Checking…</div>
    <div v-else-if="readiness.error" class="text-red-300/90 text-xs mb-3" data-test="settings-guardian-health-error">
      {{ readiness.error }}
    </div>

    <div v-if="readiness.awakening" class="space-y-3 text-sm" data-test="settings-guardian-awakening-body">
      <div class="flex flex-wrap items-center gap-2">
        <span class="text-zinc-500 text-xs uppercase tracking-wide">State</span>
        <span
          class="px-2 py-0.5 rounded-md text-xs font-semibold border"
          :class="stateBadgeClass"
          data-test="settings-guardian-state"
        >
          {{ stateLabel }}
        </span>
        <span v-if="lastCheckedLabel" class="text-zinc-600 text-xs" data-test="settings-guardian-last-check">
          Checked {{ lastCheckedLabel }}
        </span>
      </div>

      <dl class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
        <div class="bg-zinc-900/60 border border-zinc-700 rounded-lg px-3 py-2">
          <dt class="text-zinc-500">Chat model</dt>
          <dd class="text-zinc-200 mt-0.5 font-mono">{{ readiness.awakening.chat_model || '—' }}</dd>
          <dd class="text-zinc-600 mt-0.5">{{ readiness.awakening.chat_model_loaded ? 'Loaded in Ollama' : 'Cold — needs awakening' }}</dd>
        </div>
        <div class="bg-zinc-900/60 border border-zinc-700 rounded-lg px-3 py-2">
          <dt class="text-zinc-500">Profile</dt>
          <dd class="text-zinc-200 mt-0.5">{{ readiness.awakening.profile || '—' }}</dd>
          <dd v-if="readiness.awakening.embed_blocks_chat" class="text-amber-300/80 mt-0.5">Embed using RAM — awakening frees chat</dd>
        </div>
        <div class="bg-zinc-900/60 border border-zinc-700 rounded-lg px-3 py-2">
          <dt class="text-zinc-500">Field guide chunks</dt>
          <dd class="text-zinc-200 mt-0.5" data-test="settings-guardian-field-chunks">
            {{ readiness.awakening.field_guide_chunks ?? 0 }}
          </dd>
        </div>
        <div class="bg-zinc-900/60 border border-zinc-700 rounded-lg px-3 py-2">
          <dt class="text-zinc-500">Platform doc chunks</dt>
          <dd class="text-zinc-200 mt-0.5" data-test="settings-guardian-platform-chunks">
            {{ readiness.awakening.platform_doc_chunks ?? 0 }}
          </dd>
        </div>
        <div
          v-if="readiness.awakening.vision_model"
          class="bg-zinc-900/60 border border-zinc-700 rounded-lg px-3 py-2"
          data-test="settings-guardian-vision-model"
        >
          <dt class="text-zinc-500">Vision model</dt>
          <dd class="text-zinc-200 mt-0.5 font-mono">{{ readiness.awakening.vision_model }}</dd>
          <dd class="text-zinc-600 mt-0.5">
            {{ readiness.awakening.vision_model_loaded ? 'Loaded in Ollama' : 'Cold — loads on first photo question' }}
          </dd>
        </div>
      </dl>

      <p
        v-if="!readiness.awakening.rag_corpus_ok && farmId"
        class="text-xs text-amber-300/90 rounded border border-amber-900/50 bg-amber-950/30 px-3 py-2"
        data-test="settings-guardian-rag-warn"
      >
        Field memories not ingested for this farm — run
        <code class="text-amber-200/90">make guardian-bootstrap-farm FARM_ID={{ farmId }}</code>
        from the repo root.
      </p>

      <p v-for="(msg, i) in readiness.awakening.messages" :key="i" class="text-xs text-zinc-500">{{ msg }}</p>
      <p
        v-if="readiness.awakening.auto_dormant_minutes"
        class="text-xs text-zinc-500"
        data-test="settings-guardian-auto-dormant"
      >
        Auto-rest after {{ readiness.awakening.auto_dormant_minutes }} minutes idle
        <template v-if="readiness.awakening.state === 'ready' && readiness.awakening.idle_until_dormant_sec > 0">
          — ~{{ idleMinutesLabel }} until rest
        </template>
      </p>
      <p v-if="readiness.awakening.last_warmup_error" class="text-xs text-red-300/90">
        {{ readiness.awakening.last_warmup_error }}
      </p>

      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40"
          data-test="settings-guardian-awaken-btn"
          :disabled="awakeningBusy || !farmId"
          @click="awakenNow"
        >
          {{ awakeningBusy ? 'Awakening…' : 'Awaken now' }}
        </button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-zinc-900/70 text-zinc-300 border border-zinc-700 hover:bg-zinc-800 disabled:opacity-40"
          data-test="settings-guardian-rest-btn"
          :disabled="restBusy || !farmId || !readiness.awakening?.chat_model_loaded"
          title="Unload the chat model to save power — Guardian wakes back up on the next question"
          @click="restNow"
        >
          {{ restBusy ? 'Resting…' : 'Rest now' }}
        </button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-zinc-900 border border-zinc-700 text-zinc-300 hover:bg-zinc-800"
          data-test="settings-guardian-refresh-health"
          :disabled="readiness.loading"
          @click="refreshHealth"
        >
          Refresh
        </button>
      </div>

      <p class="text-[10px] text-zinc-600 leading-relaxed">
        Laptop CPU timeouts:
        <code class="text-zinc-500">make guardian-laptop-tune ARGS="--apply"</code>
        then restart the API — see
        <a href="https://github.com/gr33n-platform/gr33n-platform/blob/main/docs/local-operator-bootstrap.md" target="_blank" rel="noopener" class="text-green-500/80 hover:underline">local-operator-bootstrap.md</a>.
      </p>

      <details v-if="isFarmAdmin && farmId" class="rounded border border-zinc-700 bg-zinc-900/50 px-3 py-2 text-xs">
        <summary class="cursor-pointer text-zinc-300 select-none">Pull model (admin)</summary>
        <p class="text-zinc-500 mt-2 mb-2">Downloads into Ollama — internet required; large models take many minutes.</p>
        <div class="flex flex-wrap gap-2">
          <input
            v-model="pullName"
            type="text"
            placeholder="phi3:mini"
            class="flex-1 min-w-[10rem] bg-zinc-950 border border-zinc-700 rounded-lg px-2 py-1.5 text-zinc-200"
            data-test="settings-guardian-pull-input"
            :disabled="pulling"
          />
          <button
            type="button"
            class="px-3 py-1.5 rounded-lg border border-zinc-600 text-zinc-200 hover:bg-zinc-800 disabled:opacity-40"
            data-test="settings-guardian-pull-btn"
            :disabled="pulling || !pullName.trim()"
            @click="pullModel"
          >
            {{ pulling ? 'Pulling…' : 'Pull' }}
          </button>
        </div>
        <p v-if="pullError" class="text-red-300/90 mt-2">{{ pullError }}</p>
        <p v-if="pullOk" class="text-green-400/90 mt-2">Model pulled — refresh health above.</p>
      </details>
    </div>

    <p v-else-if="!farmId" class="text-xs text-amber-300/80">
      Select a farm in the sidebar to check Guardian readiness and RAG corpus for that farm.
    </p>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import { useFarmContextStore } from '../stores/farmContext'
import { useGuardianReadinessStore } from '../stores/guardianReadiness'
import api from '../api'

const props = defineProps({
  isFarmAdmin: { type: Boolean, default: false },
})

const capabilities = useCapabilitiesStore()
const farmContext = useFarmContextStore()
const readiness = useGuardianReadinessStore()

const farmId = computed(() => farmContext.farmId)
const pullName = ref('phi3:mini')
const pulling = ref(false)
const pullError = ref('')
const pullOk = ref(false)
const restBusy = ref(false)

const awakeningBusy = computed(() =>
  readiness.loading || readiness.isStirring || !!readiness.awakening?.warmup_in_progress,
)

const stateLabel = computed(() => {
  const s = readiness.awakening?.state
  const map = {
    ready: 'Ready',
    stirring: 'Awakening',
    sleeping: 'Sleeping',
    dormant: 'Resting',
    busy: 'Answering',
    unavailable: 'Unavailable',
  }
  return map[s] || s || 'Unknown'
})

const stateBadgeClass = computed(() => {
  const s = readiness.awakening?.state
  if (s === 'ready') return 'bg-green-950/50 border-green-800 text-green-300'
  if (s === 'stirring' || s === 'busy') return 'bg-amber-950/50 border-amber-800 text-amber-200'
  if (s === 'unavailable') return 'bg-red-950/40 border-red-900 text-red-200'
  if (s === 'dormant') return 'bg-zinc-900/80 border-zinc-600 text-zinc-400'
  return 'bg-zinc-900 border-zinc-600 text-zinc-300'
})

const lastCheckedLabel = computed(() => {
  const t = readiness.lastCheckedAt
  if (!t) return ''
  try {
    return new Date(t).toLocaleTimeString()
  } catch {
    return ''
  }
})

const idleMinutesLabel = computed(() => {
  const sec = readiness.awakening?.idle_until_dormant_sec
  if (!sec || sec <= 0) return ''
  return String(Math.max(1, Math.ceil(sec / 60)))
})

async function refreshHealth() {
  if (!farmId.value) {
    await readiness.fetchHealth(null, 'farm_counsel')
    return
  }
  await readiness.fetchHealth(farmId.value, 'farm_counsel')
}

async function awakenNow() {
  if (!farmId.value) return
  pullOk.value = false
  await readiness.warmup(farmId.value, 'farm_counsel')
}

async function restNow() {
  if (!farmId.value) return
  restBusy.value = true
  try {
    await readiness.restNow(farmId.value, 'farm_counsel')
  } finally {
    restBusy.value = false
  }
}

async function pullModel() {
  const name = pullName.value.trim()
  if (!name || !farmId.value) return
  pulling.value = true
  pullError.value = ''
  pullOk.value = false
  try {
    await api.post('/guardian/models/pull', { name, farm_id: Number(farmId.value) })
    pullOk.value = true
    await refreshHealth()
  } catch (e) {
    pullError.value = e.response?.data?.error || e.message || 'Pull failed'
  } finally {
    pulling.value = false
  }
}

onMounted(() => {
  void refreshHealth()
})

watch(farmId, () => {
  void refreshHealth()
})
</script>
