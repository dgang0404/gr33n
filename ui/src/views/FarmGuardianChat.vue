<template>
  <div class="p-4 sm:p-6 max-w-6xl mx-auto space-y-6">
    <header>
      <h1 class="text-2xl font-bold text-green-400 mb-2 flex items-center gap-2">
        Farm Guardian
        <HelpTip position="bottom">
          On-farm assistant grounded in this farm's data when a farm is selected.
          Replies stream in token-by-token via Server-Sent Events. Conversations
          persist server-side; load any prior session from the sidebar.
          See <code class="text-gr33n-400">docs/plans/phase_27_farm_guardian_ai_layer.md</code>.
        </HelpTip>
      </h1>
      <p class="text-sm text-zinc-500">
        Multi-turn. Tick <em>Use farm context</em> to ground answers in the
        selected farm's indexed chunks. Up to {{ maxHistoryTurns }} prior turns
        are replayed into each new question.
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

    <section v-else class="grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-4">
      <!-- Sessions sidebar -->
      <aside class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 max-h-[36rem] overflow-y-auto" data-test="chat-sessions">
        <div class="flex items-center justify-between">
          <h2 class="text-xs uppercase tracking-widest text-zinc-500">Sessions</h2>
          <button
            type="button"
            data-test="chat-new-session"
            class="text-xs px-2 py-1 rounded bg-zinc-800 text-zinc-200 hover:bg-zinc-700"
            :disabled="streaming"
            @click="newSession"
          >
            New
          </button>
        </div>
        <p v-if="!sessions.length" class="text-xs text-zinc-600 italic">
          No saved sessions yet. Send your first message to start one.
        </p>
        <ul class="space-y-1">
          <li
            v-for="s in sessions"
            :key="s.session_id"
            class="rounded p-2 text-xs space-y-1 group relative"
            :class="s.session_id === sessionId ? 'bg-green-900/40 border border-green-800 text-green-100' : 'hover:bg-zinc-800 text-zinc-300 border border-transparent'"
          >
            <div class="flex items-center justify-between gap-2 cursor-pointer" @click="loadSession(s.session_id)">
              <span class="font-medium truncate" :title="sessionLabel(s)">{{ sessionLabel(s) }}</span>
              <span class="text-[10px] text-zinc-500 shrink-0">{{ s.turn_count }} turn{{ s.turn_count === 1 ? '' : 's' }}</span>
            </div>
            <div class="text-[10px] text-zinc-500 flex items-center justify-between gap-1">
              <div class="flex items-center gap-2">
                <span v-if="s.any_grounded" class="text-gr33n-500">grounded</span>
                <span>{{ formatTime(s.last_turn_at) }}</span>
                <span
                  v-if="(s.total_prompt_tokens || 0) + (s.total_completion_tokens || 0) > 0"
                  class="text-zinc-600"
                  :title="`prompt ${s.total_prompt_tokens} · completion ${s.total_completion_tokens}`"
                >
                  {{ (s.total_prompt_tokens || 0) + (s.total_completion_tokens || 0) }} tok
                </span>
              </div>
              <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300"
                  :disabled="streaming"
                  data-test="chat-session-rename"
                  :title="'Rename session'"
                  @click.stop="renameSession(s)"
                >
                  ✎
                </button>
                <button
                  type="button"
                  class="px-1.5 py-0.5 rounded bg-zinc-800 hover:bg-red-900/60 text-zinc-300 hover:text-red-200"
                  :disabled="streaming"
                  data-test="chat-session-delete"
                  :title="'Delete session'"
                  @click.stop="deleteSession(s)"
                >
                  ✕
                </button>
              </div>
            </div>
          </li>
        </ul>
      </aside>

      <div class="space-y-4">
        <!-- Transcript -->
        <section
          v-if="transcript.length || streaming"
          class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4 max-h-[36rem] overflow-y-auto"
          data-test="chat-transcript"
        >
          <article
            v-for="(t, idx) in transcript"
            :key="t.turn_index ?? idx"
            class="space-y-3 border-b border-zinc-800 pb-3 last:border-b-0 last:pb-0"
          >
            <div class="text-zinc-300 text-sm" data-test="chat-user-turn">
              <span class="text-[10px] uppercase tracking-widest text-zinc-500 mr-2">you</span>
              <span class="whitespace-pre-wrap">{{ t.user_message }}</span>
            </div>
            <div class="text-zinc-100 text-sm" data-test="chat-assistant-turn">
              <span class="text-[10px] uppercase tracking-widest text-green-500 mr-2">guardian</span>
              <span class="whitespace-pre-wrap">{{ t.assistant_message }}</span>
              <div class="mt-1 text-[10px] text-zinc-600">
                {{ t.llm_model }}<span v-if="t.grounded"> · grounded · {{ t.context_count }} chunks</span>
                <span
                  v-if="(t.prompt_tokens || 0) + (t.completion_tokens || 0) > 0"
                  class="ml-2"
                  data-test="chat-turn-tokens"
                  :title="`prompt ${t.prompt_tokens} · completion ${t.completion_tokens}`"
                >
                  · {{ (t.prompt_tokens || 0) + (t.completion_tokens || 0) }} tok
                </span>
              </div>
            </div>
            <ul v-if="t.citations?.length" class="space-y-1 pl-6">
              <li
                v-for="c in t.citations"
                :key="c.ref + '-' + c.chunk_id"
                class="text-[11px] bg-zinc-950 border border-zinc-800 rounded p-2 text-zinc-400"
              >
                <span class="text-gr33n-500 font-mono">[{{ c.ref }}]</span>
                {{ c.source_type }} #{{ c.source_id }} · chunk {{ c.chunk_id }}
                <p class="mt-1 text-zinc-500">{{ c.excerpt }}</p>
              </li>
            </ul>
          </article>
          <div v-if="streaming" class="text-zinc-100 text-sm" data-test="chat-streaming-row">
            <span class="text-[10px] uppercase tracking-widest text-green-500 mr-2">guardian</span>
            <span class="whitespace-pre-wrap">{{ streamingText }}<span class="text-zinc-500 animate-pulse">▍</span></span>
          </div>
        </section>

        <!-- Composer -->
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
          </div>
          <p v-if="errorMessage" data-test="chat-error" class="text-sm text-red-400 bg-red-950/50 border border-red-900 rounded-lg px-3 py-2">
            {{ errorMessage }}
          </p>
          <p v-if="sessionId" class="text-[10px] text-zinc-600">
            session_id: <span class="font-mono">{{ sessionId }}</span>
          </p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import HelpTip from '../components/HelpTip.vue'
import api from '../api'
import { useFarmContextStore } from '../stores/farmContext'
import { useCapabilitiesStore } from '../stores/capabilities'

const maxHistoryTurns = 20

const farmContext = useFarmContextStore()
const capabilities = useCapabilitiesStore()

const message = ref('')
const useFarmContext = ref(false)
const streaming = ref(false)
const streamingText = ref('')
const errorMessage = ref('')

const sessionId = ref('')
const transcript = ref([]) // [{turn_index, user_message, assistant_message, llm_model, grounded, context_count, citations, farm_id}]
const sessions = ref([])

onMounted(async () => {
  if (!capabilities.loaded) await capabilities.fetch()
  if (!capabilities.isLite) {
    await refreshSessions()
  }
})

async function refreshSessions() {
  try {
    const r = await api.get('/v1/chat/sessions')
    sessions.value = Array.isArray(r.data?.sessions) ? r.data.sessions : []
  } catch (e) {
    sessions.value = []
  }
}

async function loadSession(id) {
  if (streaming.value) return
  try {
    const r = await api.get('/v1/chat/sessions/' + id)
    sessionId.value = id
    transcript.value = Array.isArray(r.data?.turns) ? r.data.turns : []
    errorMessage.value = ''
  } catch (e) {
    errorMessage.value = e.message || 'failed to load session'
  }
}

function newSession() {
  if (streaming.value) return
  sessionId.value = ''
  transcript.value = []
  streamingText.value = ''
  errorMessage.value = ''
}

function sessionLabel(s) {
  if (s.title && s.title.trim()) return s.title
  if (s.first_user_message && s.first_user_message.trim()) return s.first_user_message
  return 'Untitled'
}

async function renameSession(s) {
  if (streaming.value) return
  const next = window.prompt('Rename session:', s.title || s.first_user_message || '')
  if (next === null) return
  try {
    const r = await api.patch('/v1/chat/sessions/' + s.session_id, { title: next })
    const i = sessions.value.findIndex((x) => x.session_id === s.session_id)
    if (i !== -1) sessions.value[i] = { ...sessions.value[i], title: r.data?.title ?? null }
  } catch (e) {
    errorMessage.value = e.message || 'rename failed'
  }
}

async function deleteSession(s) {
  if (streaming.value) return
  if (!window.confirm('Delete this session? This cannot be undone.')) return
  try {
    await api.delete('/v1/chat/sessions/' + s.session_id)
    sessions.value = sessions.value.filter((x) => x.session_id !== s.session_id)
    if (sessionId.value === s.session_id) {
      sessionId.value = ''
      transcript.value = []
    }
  } catch (e) {
    errorMessage.value = e.message || 'delete failed'
  }
}

function apiBaseURL() {
  return import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  if (Number.isNaN(d.getTime())) return ts
  return d.toLocaleString()
}

async function send() {
  if (!message.value.trim()) return
  if (useFarmContext.value && !farmContext.farmId) return
  errorMessage.value = ''
  streamingText.value = ''
  streaming.value = true

  const userMessage = message.value.trim()
  const body = { message: userMessage, stream: true }
  if (sessionId.value) body.session_id = sessionId.value
  if (useFarmContext.value && farmContext.farmId) {
    body.farm_id = Number(farmContext.farmId)
  }

  const token = localStorage.getItem('gr33n_token') ?? ''
  let finalEvent = null
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
    finalEvent = await consumeSSE(resp.body)
    if (finalEvent) {
      sessionId.value = finalEvent.session_id || sessionId.value
      transcript.value.push({
        turn_index: finalEvent.turn_index,
        user_message: userMessage,
        assistant_message: finalEvent.answer || streamingText.value,
        llm_model: finalEvent.llm_model || '',
        grounded: !!finalEvent.grounded,
        context_count: finalEvent.context_count || 0,
        citations: Array.isArray(finalEvent.citations) ? finalEvent.citations : [],
        farm_id: body.farm_id ?? null,
      })
      message.value = ''
      await refreshSessions()
    }
  } catch (e) {
    errorMessage.value = e.message || 'chat failed'
  } finally {
    streaming.value = false
    streamingText.value = ''
  }
}

async function consumeSSE(stream) {
  const reader = stream.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  let done = null
  for (;;) {
    const { value, done: end } = await reader.read()
    if (end) break
    buf += decoder.decode(value, { stream: true })
    const events = buf.split('\n\n')
    buf = events.pop() ?? ''
    for (const block of events) {
      const result = handleSSEBlock(block)
      if (result) done = result
    }
  }
  return done
}

function handleSSEBlock(block) {
  let eventType = 'message'
  let data = ''
  for (const line of block.split('\n')) {
    if (line.startsWith('event:')) eventType = line.slice(6).trim()
    else if (line.startsWith('data:')) data += (data ? '\n' : '') + line.slice(5).trim()
  }
  if (!data) return null
  if (data === '[DONE]') return null
  let parsed
  try { parsed = JSON.parse(data) } catch { return null }
  if (eventType === 'delta' && typeof parsed.text === 'string') {
    streamingText.value += parsed.text
  } else if (eventType === 'done') {
    return parsed
  } else if (eventType === 'error') {
    errorMessage.value = parsed.error || 'LLM request failed'
  }
  return null
}
</script>
