<template>
  <div class="p-4 sm:p-6 max-w-3xl mx-auto space-y-6">
    <header>
      <h1 class="text-2xl font-bold text-green-400 mb-2 flex items-center gap-2">
        Farm Guardian
        <HelpTip position="bottom">
          On-farm assistant grounded in this farm's data when a farm is selected.
          Replies stream in token-by-token via Server-Sent Events. When
          <code class="text-zinc-500 text-[10px]">AI_ENABLED=false</code> on the API
          this panel is read-only.
          See <code class="text-gr33n-400">docs/plans/phase_27_farm_guardian_ai_layer.md</code>.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500">
        Single-turn for now. Tick <em>Use farm context</em> to ground answers in the
        selected farm's indexed chunks (citations show below the reply).
      </p>
    </header>

    <section
      v-if="capabilities.isLite"
      data-test="chat-lite-banner"
      class="rounded-xl border border-amber-900/60 bg-amber-950/40 px-4 py-3 text-sm text-amber-200"
    >
      Farm Guardian is not available on this installation.
      Your farm is running in <strong>Lite mode</strong> — all operational features
      remain fully active. Set <code class="text-gr33n-400">AI_ENABLED=true</code>
      on the API and restart to enable chat.
    </section>

    <section v-else class="space-y-4">
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
        <div class="flex flex-col gap-2">
          <label class="text-xs text-zinc-400">Your message</label>
          <textarea
            v-model="message"
            rows="3"
            placeholder="e.g. What should I check on the morning walkthrough?"
            class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-gr33n-600"
          />
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="useFarmContext" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" />
            Use farm context
          </label>
          <span v-if="useFarmContext && !farmContext.farmId" class="text-amber-300/80 text-xs">
            Select a farm in the sidebar first to ground answers.
          </span>
          <button
            type="button"
            data-test="chat-send-button"
            :disabled="streaming || !message.trim() || (useFarmContext && !farmContext.farmId)"
            class="ml-auto px-4 py-2 rounded-lg bg-green-900/50 text-green-400 border border-green-800 hover:bg-green-900/70 disabled:opacity-40 text-sm font-medium"
            @click="send"
          >
            {{ streaming ? 'Streaming…' : 'Send' }}
          </button>
          <button
            type="button"
            :disabled="streaming"
            class="px-3 py-2 rounded-lg bg-zinc-800 text-zinc-300 border border-zinc-600 hover:bg-zinc-700 disabled:opacity-40 text-xs"
            @click="reset"
          >
            Clear
          </button>
        </div>
        <p v-if="errorMessage" data-test="chat-error" class="text-sm text-red-400 bg-red-950/50 border border-red-900 rounded-lg px-3 py-2">
          {{ errorMessage }}
        </p>
      </div>

      <section v-if="answer || streaming" class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
        <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">
          Answer
          <span v-if="llmModel" class="text-zinc-600 font-normal normal-case">· {{ llmModel }}</span>
          <span v-if="grounded" class="text-zinc-600 font-normal normal-case">· grounded</span>
        </h2>
        <div class="text-zinc-100 whitespace-pre-wrap text-sm leading-relaxed" data-test="chat-answer">{{ answer }}<span v-if="streaming" class="text-zinc-500 animate-pulse">▍</span></div>
        <div v-if="citations.length" class="border-t border-zinc-800 pt-4 space-y-2">
          <p class="text-[11px] uppercase tracking-widest text-zinc-500">Citations</p>
          <ul class="space-y-2">
            <li
              v-for="c in citations"
              :key="c.ref + '-' + c.chunk_id"
              class="text-xs bg-zinc-950 border border-zinc-800 rounded-lg p-3 text-zinc-300"
            >
              <span class="text-gr33n-500 font-mono">[{{ c.ref }}]</span>
              {{ c.source_type }} #{{ c.source_id }} · chunk {{ c.chunk_id }}
              <p class="mt-1 text-zinc-500">{{ c.excerpt }}</p>
            </li>
          </ul>
        </div>
      </section>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import HelpTip from '../components/HelpTip.vue'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'

const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()
onMounted(() => {
  if (!capabilities.loaded) capabilities.fetch()
})

const message = ref('')
const useFarmContext = ref(false)
const streaming = ref(false)
const answer = ref('')
const llmModel = ref('')
const grounded = ref(false)
const citations = ref([])
const errorMessage = ref('')

function reset() {
  message.value = ''
  answer.value = ''
  llmModel.value = ''
  grounded.value = false
  citations.value = []
  errorMessage.value = ''
}

function apiBaseURL() {
  return import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
}

async function send() {
  if (!message.value.trim()) return
  if (useFarmContext.value && !farmContext.farmId) return
  errorMessage.value = ''
  answer.value = ''
  llmModel.value = ''
  grounded.value = false
  citations.value = []
  streaming.value = true

  const body = { message: message.value.trim(), stream: true }
  if (useFarmContext.value && farmContext.farmId) {
    body.farm_id = Number(farmContext.farmId)
  }

  const token = localStorage.getItem('gr33n_token') ?? ''
  try {
    const resp = await fetch(apiBaseURL() + '/v1/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
        ...(token ? { Authorization: 'Bearer ' + token } : {}),
      },
      body: JSON.stringify(body),
    })
    if (!resp.ok || !resp.body) {
      let text = `HTTP ${resp.status}`
      try { text = (await resp.json()).error || text } catch {}
      errorMessage.value = text
      return
    }
    await consumeSSE(resp.body)
  } catch (e) {
    errorMessage.value = e.message || 'chat failed'
  } finally {
    streaming.value = false
  }
}

async function consumeSSE(stream) {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    const events = buf.split('\n\n')
    buf = events.pop() ?? ''
    for (const block of events) {
      handleSSEBlock(block)
    }
  }
}

function handleSSEBlock(block) {
  let eventType = 'message'
  let data = ''
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) eventType = line.slice(6).trim()
    else if (line.startsWith('data:')) data += (data ? '\n' : '') + line.slice(5).trim()
  }
  if (!data) return
  if (data === '[DONE]') return
  let parsed
  try { parsed = JSON.parse(data) } catch { return }
  if (eventType === 'delta' && typeof parsed.text === 'string') {
    answer.value += parsed.text
  } else if (eventType === 'done') {
    llmModel.value = parsed.llm_model || ''
    grounded.value = !!parsed.grounded
    citations.value = Array.isArray(parsed.citations) ? parsed.citations : []
    if (!answer.value && parsed.answer) answer.value = parsed.answer
  } else if (eventType === 'error') {
    errorMessage.value = parsed.error || 'LLM request failed'
  }
}
</script>
