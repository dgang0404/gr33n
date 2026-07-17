import { defineStore } from 'pinia'
import api from '../api'

/**
 * Phase 30 WS1 — pending Guardian change requests (PR inbox).
 */
export const useGuardianProposalsStore = defineStore('guardianProposals', {
  state: () => ({
    proposals: [],
    total: 0,
    pendingCount: 0,
    loading: false,
    error: '',
    lastFarmId: null,
  }),

  getters: {
    /** Newest pending proposal per session_id (Phase 197). */
    pendingBySessionId(state) {
      const out = {}
      for (const p of state.proposals) {
        if (p.status !== 'pending' || !p.session_id) continue
        if (!out[p.session_id]) out[p.session_id] = p
      }
      return out
    },
  },

  actions: {
    async fetch(farmId, { status = 'pending' } = {}) {
      if (!farmId || !localStorage.getItem('gr33n_token')) {
        this.proposals = []
        this.total = 0
        this.pendingCount = 0
        this.lastFarmId = null
        return
      }
      this.loading = true
      this.error = ''
      this.lastFarmId = farmId
      try {
        const r = await api.get('/v1/chat/proposals', {
          params: { farm_id: farmId, status, limit: 50, offset: 0 },
        })
        const proposals = r.data?.proposals ?? []
        this.proposals = proposals.slice().sort((a, b) => {
          const ta = Date.parse(a.created_at) || 0
          const tb = Date.parse(b.created_at) || 0
          return tb - ta
        })
        this.total = r.data?.total ?? 0
        if (status === 'pending') {
          this.pendingCount = this.total
        }
      } catch (e) {
        this.error = e?.response?.data?.error || e.message || 'Failed to load proposals'
        this.proposals = []
        this.total = 0
        if (status === 'pending') this.pendingCount = 0
      } finally {
        this.loading = false
      }
    },

    async refreshPendingCount(farmId) {
      if (!farmId || !localStorage.getItem('gr33n_token')) {
        this.pendingCount = 0
        return
      }
      try {
        const r = await api.get('/v1/chat/proposals', {
          params: { farm_id: farmId, status: 'pending', limit: 1, offset: 0 },
        })
        this.pendingCount = r.data?.total ?? 0
        this.lastFarmId = farmId
      } catch {
        this.pendingCount = 0
      }
    },

    removeProposal(proposalId) {
      const before = this.proposals.length
      this.proposals = this.proposals.filter((p) => p.proposal_id !== proposalId)
      if (this.proposals.length < before) {
        this.total = Math.max(0, this.total - 1)
        this.pendingCount = Math.max(0, this.pendingCount - 1)
      }
    },

    async dismissProposal(proposalId) {
      if (!proposalId) return
      await api.post(`/v1/chat/proposals/${proposalId}/dismiss`)
      this.removeProposal(proposalId)
    },

    patchProposal(proposalId, patch) {
      const i = this.proposals.findIndex((p) => p.proposal_id === proposalId)
      if (i >= 0) {
        this.proposals[i] = { ...this.proposals[i], ...patch }
      }
    },
  },
})
