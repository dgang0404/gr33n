import { defineStore } from 'pinia'

/**
 * Phase 37 WS9 — Farm Guardian chat stream + transcript (survives drawer/route unmount).
 */
function apiBaseURL() {
  return import.meta.env.VITE_API_URL ?? 'http://localhost:8080'
}

export const useGuardianChatStore = defineStore('guardianChat', {
  state: () => ({
    streaming: false,
    streamingText: '',
    error: '',
    transcript: [],
    lastFarmId: null,
    /** @type {AbortController | null} */
    _abort: null,
  }),

  getters: {
    isThinking: (state) => state.streaming,
  },

  actions: {
    setTranscript(turns) {
      this.transcript = Array.isArray(turns) ? [...turns] : []
    },

    clearTranscript() {
      this.transcript = []
      if (!this.streaming) {
        this.streamingText = ''
        this.error = ''
      }
    },

    clearError() {
      this.error = ''
    },

    cancelStream() {
      if (this._abort) {
        this._abort.abort()
        this._abort = null
      }
      this.streaming = false
      this.streamingText = ''
    },

    findProposalTurn(proposalId) {
      for (const t of this.transcript) {
        if (!t.proposals) continue
        const index = t.proposals.findIndex((p) => p.proposal_id === proposalId)
        if (index >= 0) return { turn: t, index }
      }
      return null
    },

    patchProposal(proposalId, patch) {
      const hit = this.findProposalTurn(proposalId)
      if (!hit) return
      hit.turn.proposals[hit.index] = { ...hit.turn.proposals[hit.index], ...patch }
    },

    appendTurn(turn) {
      this.transcript.push(turn)
      this.streamingText = ''
    },

    /**
     * POST /v1/chat SSE. Aborts only on explicit cancelStream or a new sendMessage.
     * @returns {{ finalEvent, userMessage, attachedIds, body } | null}
     */
    async sendMessage({ message, farmId, sessionId, contextRef, attachmentIds, setupMode }) {
      const trimmed = (message || '').trim()
      if (!trimmed) return null
      if (this.streaming) this.cancelStream()

      this.error = ''
      this.streamingText = ''
      this.streaming = true
      this.lastFarmId = farmId != null ? Number(farmId) : null

      const body = { message: trimmed, stream: true }
      if (sessionId) body.session_id = sessionId
      if (farmId != null) body.farm_id = Number(farmId)
      if (contextRef) body.context_ref = contextRef
      if (attachmentIds?.length) body.attachment_ids = [...attachmentIds]
      if (setupMode) body.setup_mode = true

      const ctrl = new AbortController()
      this._abort = ctrl

      const token = localStorage.getItem('gr33n_token') ?? ''
      try {
        const resp = await fetch(apiBaseURL() + '/v1/chat', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'text/event-stream',
            ...(token ? { Authorization: 'Bearer ' + token } : {}),
          },
          body: JSON.stringify(body),
          signal: ctrl.signal,
        })
        if (!resp.ok || !resp.body) {
          let text = `HTTP ${resp.status}`
          try {
            const j = await resp.json()
            text = j.error || text
          } catch { /* ignore */ }
          this.error = text
          return null
        }
        const finalEvent = await this.consumeSSE(resp.body)
        return { finalEvent, userMessage: trimmed, attachedIds: attachmentIds || [], body }
      } catch (e) {
        if (e?.name === 'AbortError') return null
        this.error = e.message || 'chat failed'
        return null
      } finally {
        if (this._abort === ctrl) this._abort = null
        this.streaming = false
      }
    },

    async consumeSSE(stream) {
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
          const result = this.handleSSEBlock(block)
          if (result) done = result
        }
      }
      return done
    },

    handleSSEBlock(block) {
      let eventType = 'message'
      let data = ''
      for (const line of block.split('\n')) {
        if (line.startsWith('event:')) eventType = line.slice(6).trim()
        else if (line.startsWith('data:')) data += (data ? '\n' : '') + line.slice(5).trim()
      }
      if (!data) return null
      if (data === '[DONE]') return null
      let parsed
      try {
        parsed = JSON.parse(data)
      } catch {
        return null
      }
      if (eventType === 'delta' && typeof parsed.text === 'string') {
        this.streamingText += parsed.text
      } else if (eventType === 'done') {
        return parsed
      } else if (eventType === 'error') {
        this.error = parsed.error || 'LLM request failed'
      }
      return null
    },
  },
})
