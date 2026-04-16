<template>
  <div class="p-6 max-w-2xl">
    <h1 class="text-2xl font-bold text-green-400 mb-6">Settings</h1>

    <!-- Account info -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
        <span>👤</span> Account
      </h2>
      <div class="grid grid-cols-2 gap-3 text-sm">
        <div>
          <p class="text-zinc-500 text-xs uppercase tracking-wide mb-0.5">Username</p>
          <p class="text-white font-mono">{{ auth.username ?? 'admin' }}</p>
        </div>
        <div>
          <p class="text-zinc-500 text-xs uppercase tracking-wide mb-0.5">Session</p>
          <p class="text-green-400 text-xs font-semibold">Active — expires in {{ tokenExpiry }}</p>
        </div>
      </div>
    </section>

    <!-- Change password -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-4 flex items-center gap-2">
        <span>🔒</span> Change Password
      </h2>

      <form @submit.prevent="submitPassword" class="flex flex-col gap-4">
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">Current Password</label>
          <input
            v-model="pwForm.current"
            type="password"
            autocomplete="current-password"
            required
            class="input-field"
          />
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">New Password</label>
          <input
            v-model="pwForm.next"
            type="password"
            autocomplete="new-password"
            minlength="8"
            required
            class="input-field"
          />
          <p class="text-zinc-600 text-xs">Minimum 8 characters</p>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">Confirm New Password</label>
          <input
            v-model="pwForm.confirm"
            type="password"
            autocomplete="new-password"
            required
            class="input-field"
            :class="{ 'border-red-600': pwForm.confirm && pwForm.next !== pwForm.confirm }"
          />
          <p v-if="pwForm.confirm && pwForm.next !== pwForm.confirm" class="text-red-400 text-xs">
            Passwords do not match
          </p>
        </div>

        <!-- Error / success -->
        <p v-if="pwError" class="text-red-400 text-sm bg-red-950 border border-red-800 rounded-lg px-3 py-2">
          {{ pwError }}
        </p>
        <p v-if="pwSuccess" class="text-green-400 text-sm bg-green-950 border border-green-800 rounded-lg px-3 py-2">
          Password updated successfully
        </p>

        <div class="flex gap-3 pt-1">
          <button
            type="submit"
            :disabled="pwLoading || pwForm.next !== pwForm.confirm"
            class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 disabled:text-zinc-500
                   text-white text-sm font-semibold px-5 py-2 rounded-lg transition-colors"
          >
            {{ pwLoading ? 'Updating…' : 'Update Password' }}
          </button>
          <button
            type="button"
            @click="resetPwForm"
            class="text-zinc-400 hover:text-white text-sm px-4 py-2 rounded-lg transition-colors"
          >
            Cancel
          </button>
        </div>
      </form>
    </section>

    <!-- Organizations (multi-farm tenancy) -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
        <span>Org</span> Organizations
      </h2>
      <p class="text-zinc-500 text-xs mb-4">
        Optional grouping for several farms. Usage totals are a lightweight metering hook for future billing (not a live invoice).
      </p>
      <form @submit.prevent="createOrg" class="flex flex-wrap gap-2 mb-4">
        <input v-model="newOrgName" type="text" placeholder="New organization name" required
          class="input-field flex-1 min-w-[200px] text-sm" />
        <button type="submit" :disabled="orgSaving"
          class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 text-white text-xs font-semibold px-4 py-2 rounded-lg shrink-0">
          {{ orgSaving ? 'Creating…' : 'Create' }}
        </button>
        <button type="button" class="text-zinc-500 hover:text-zinc-300 text-xs px-2" :disabled="orgLoading" @click="loadOrgs">
          Refresh list
        </button>
      </form>
      <p v-if="orgError" class="text-red-400 text-xs mb-2">{{ orgError }}</p>
      <div v-if="orgLoading && !orgs.length" class="text-zinc-500 text-sm">Loading…</div>
      <ul v-else-if="orgs.length" class="space-y-2 mb-4">
        <li v-for="o in orgs" :key="o.id"
          class="flex flex-wrap items-center justify-between gap-2 bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm">
          <div>
            <span class="text-zinc-200 font-medium">{{ o.name }}</span>
            <span class="text-zinc-500 text-xs ml-2">{{ o.role_in_org }} · {{ o.plan_tier }}</span>
          </div>
          <button type="button" class="text-xs text-green-500 hover:text-green-400"
            @click="loadOrgUsage(o.id)">
            Usage summary
          </button>
        </li>
      </ul>
      <p v-else class="text-zinc-600 text-sm mb-4">You are not in any organization yet.</p>
      <div v-if="orgUsage" class="text-xs text-zinc-400 font-mono bg-zinc-950 border border-zinc-800 rounded-lg px-3 py-2 mb-4">
        <span class="text-zinc-500">Last requested: </span>
        farms {{ orgUsage.farm_count }} · devices {{ orgUsage.device_count }} · sensors {{ orgUsage.sensor_count }}
        · tasks {{ orgUsage.task_count }} · cost lines {{ orgUsage.cost_transaction_count }}
      </div>

      <div v-if="adminOrgs.length && farmContext.farmId" class="border-t border-zinc-700 pt-4">
        <p class="text-zinc-400 text-xs uppercase tracking-wide mb-2">Current farm → organization</p>
        <div class="flex flex-wrap gap-2 items-center">
          <select v-model="farmOrgLink"
            class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-sm rounded-lg px-3 py-2 min-w-[12rem] focus:outline-none">
            <option value="">— None —</option>
            <option v-for="o in adminOrgs" :key="o.id" :value="String(o.id)">{{ o.name }}</option>
          </select>
          <button type="button" :disabled="farmOrgSaving" @click="saveFarmOrgLink"
            class="bg-zinc-700 hover:bg-zinc-600 disabled:opacity-40 text-white text-xs font-semibold px-4 py-2 rounded-lg">
            {{ farmOrgSaving ? 'Saving…' : 'Save link' }}
          </button>
        </div>
        <p v-if="farmOrgMsg" class="text-zinc-500 text-xs mt-2">{{ farmOrgMsg }}</p>
      </div>

      <form v-if="adminOrgs.length" @submit.prevent="inviteOrgMember" class="border-t border-zinc-700 pt-4 mt-4 space-y-2">
        <p class="text-zinc-400 text-xs uppercase tracking-wide">Add existing user to organization</p>
        <div class="flex flex-wrap gap-2">
          <select v-model="orgInviteTargetId" required
            class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-xs rounded-lg px-2 py-2">
            <option v-for="o in adminOrgs" :key="o.id" :value="String(o.id)">{{ o.name }}</option>
          </select>
          <input v-model="orgInviteEmail" type="email" required placeholder="email@example.com"
            class="input-field flex-1 min-w-[180px] text-xs" />
          <select v-model="orgInviteRole" class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-xs rounded-lg px-2 py-2">
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
          <button type="submit" :disabled="orgInviting"
            class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 text-white text-xs font-semibold px-4 py-2 rounded-lg">
            {{ orgInviting ? 'Adding…' : 'Add' }}
          </button>
        </div>
        <p v-if="orgInviteError" class="text-red-400 text-xs">{{ orgInviteError }}</p>
        <p v-if="orgInviteOk" class="text-green-400 text-xs">Member added.</p>
      </form>
    </section>

    <!-- Farm Members -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-4 flex items-center gap-2">
        <span>👥</span> Farm Members
      </h2>

      <div v-if="membersLoading" class="text-zinc-500 text-sm">Loading members...</div>
      <div v-else-if="members.length === 0" class="text-zinc-500 text-sm">No members yet.</div>
      <div v-else class="space-y-2 mb-4">
        <div v-for="m in members" :key="m.user_id"
          class="flex items-center justify-between bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2.5">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-8 h-8 rounded-full bg-zinc-700 flex items-center justify-center text-xs text-zinc-300 font-bold shrink-0">
              {{ (m.full_name || m.email || '?')[0].toUpperCase() }}
            </div>
            <div class="min-w-0">
              <p class="text-white text-sm truncate">{{ m.full_name || m.email }}</p>
              <p class="text-zinc-500 text-xs truncate">{{ m.email }}</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <select :value="m.role_in_farm" @change="changeRole(m.user_id, $event.target.value)"
              class="bg-zinc-800 border border-zinc-700 text-zinc-300 text-xs rounded px-2 py-1 focus:outline-none">
              <option value="owner">Owner</option>
              <option value="manager">Manager</option>
              <option value="operator">Operator</option>
              <option value="finance">Finance</option>
              <option value="viewer">Viewer</option>
            </select>
            <button @click="removeMember(m.user_id)"
              class="text-zinc-500 hover:text-red-400 text-xs transition-colors" title="Remove">
              ✕
            </button>
          </div>
        </div>
      </div>

      <!-- Invite form -->
      <form @submit.prevent="inviteMember" class="flex gap-2">
        <input v-model="inviteEmail" type="email" placeholder="email@example.com" required
          class="input-field flex-1 text-xs" />
        <select v-model="inviteRole" class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-xs rounded-lg px-2 py-2 focus:outline-none">
          <option value="viewer">Viewer</option>
          <option value="operator">Operator</option>
          <option value="finance">Finance</option>
          <option value="manager">Manager</option>
          <option value="owner">Owner</option>
        </select>
        <button type="submit" :disabled="inviting"
          class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 text-white text-xs font-semibold px-4 py-2 rounded-lg transition-colors shrink-0">
          {{ inviting ? 'Inviting…' : 'Invite' }}
        </button>
      </form>
      <p v-if="inviteError" class="text-red-400 text-xs mt-2">{{ inviteError }}</p>
      <p v-if="inviteSuccess" class="text-green-400 text-xs mt-2">Member invited successfully.</p>
    </section>

    <!-- Insert Commons (benchmark sharing) -->
    <section v-if="farmContext.farmId" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3">Insert Commons</h2>
      <p class="text-zinc-400 text-sm mb-3">
        Optional community benchmarks. The API only sends <span class="text-zinc-300">coarse aggregates</span> under a stable per-farm pseudonym
        (not your farm name). If <code class="text-green-400">INSERT_COMMONS_INGEST_URL</code> is unset, the server records an attempt but does not call out.
        You can revoke anytime by turning sharing off.
      </p>
      <label class="flex items-center gap-2 text-zinc-300 text-sm mb-3">
        <input v-model="insertOptIn" type="checkbox" class="rounded bg-zinc-800 border-zinc-700"
          @change="onInsertOptInChange" />
        Share anonymized aggregates with Insert Commons
      </label>
      <div class="flex flex-wrap items-center gap-2">
        <button type="button" :disabled="!insertOptIn || insertSyncing"
          class="bg-zinc-700 hover:bg-zinc-600 disabled:opacity-40 text-white text-xs font-semibold px-4 py-2 rounded-lg"
          @click="runInsertSync">
          {{ insertSyncing ? 'Syncing…' : 'Run sync' }}
        </button>
        <span v-if="insertSyncMsg" class="text-zinc-500 text-xs">{{ insertSyncMsg }}</span>
      </div>
      <div v-if="insertOptIn" class="mt-4">
        <div class="flex items-center justify-between mb-2">
          <p class="text-zinc-500 text-xs uppercase tracking-wide">Recent sync attempts</p>
          <button type="button" class="text-zinc-500 hover:text-white text-xs" @click="loadInsertHistory" :disabled="insertHistoryLoading">
            {{ insertHistoryLoading ? 'Loading…' : 'Refresh' }}
          </button>
        </div>
        <div v-if="insertHistoryLoading && insertHistory.length === 0" class="text-zinc-600 text-xs">Loading history…</div>
        <div v-else-if="insertHistory.length === 0" class="text-zinc-600 text-xs">No attempts yet.</div>
        <ul v-else class="space-y-2">
          <li v-for="e in insertHistory" :key="e.id" class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-xs">
            <div class="flex items-center justify-between gap-2">
              <span class="text-zinc-300 font-mono">{{ e.status }}</span>
              <span class="text-zinc-600 shrink-0">{{ fmtTs(e.created_at) }}</span>
            </div>
            <div class="text-zinc-600 mt-1 break-all" v-if="e.idempotency_key">idem: {{ e.idempotency_key }}</div>
            <div class="text-zinc-600 mt-1" v-if="e.http_status != null">http: {{ e.http_status }}</div>
            <div class="text-red-300 mt-1 break-words" v-if="e.error">{{ e.error }}</div>
          </li>
        </ul>
      </div>
    </section>

    <!-- Pi connection info -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
        <span>🍓</span> Pi Client
      </h2>
      <p class="text-zinc-400 text-sm mb-3">
        Set this key in <code class="bg-zinc-700 px-1.5 py-0.5 rounded text-green-400 text-xs">pi_client/config.yaml</code>
        under <code class="bg-zinc-700 px-1.5 py-0.5 rounded text-green-400 text-xs">api.api_key</code>
      </p>
      <div class="flex items-center gap-2">
        <code class="flex-1 bg-zinc-900 border border-zinc-600 rounded px-3 py-2 text-xs font-mono text-zinc-300 break-all">
          {{ showKey ? piApiKey : '••••••••••••••••••••••••' }}
        </code>
        <button
          @click="showKey = !showKey"
          class="text-zinc-500 hover:text-white text-xs border border-zinc-600 rounded px-2.5 py-2 transition-colors"
        >
          {{ showKey ? 'Hide' : 'Show' }}
        </button>
      </div>
      <p class="text-zinc-600 text-xs mt-2">
        PI_API_KEY env var — set at API startup. Restart API to change.
      </p>
    </section>

    <!-- Sign out -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
      <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
        <span>🚪</span> Session
      </h2>
      <button
        @click="signOut"
        class="bg-zinc-700 hover:bg-red-900 border border-zinc-600 hover:border-red-700
               text-zinc-300 hover:text-red-300 text-sm font-semibold px-5 py-2 rounded-lg transition-colors"
      >
        Sign out
      </button>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import api from '../api'

const router = useRouter()
const auth   = useAuthStore()
const farmStore = useFarmStore()
const farmContext = useFarmContextStore()

// ── Organizations ─────────────────────────────────────────────────────────────
const orgs = ref([])
const orgLoading = ref(false)
const orgError = ref(null)
const newOrgName = ref('')
const orgSaving = ref(false)
const orgUsage = ref(null)
const farmOrgLink = ref('')
const farmOrgSaving = ref(false)
const farmOrgMsg = ref('')
const orgInviteEmail = ref('')
const orgInviteRole = ref('member')
const orgInviteTargetId = ref('')
const orgInviting = ref(false)
const orgInviteError = ref(null)
const orgInviteOk = ref(false)

const adminOrgs = computed(() =>
  orgs.value.filter((o) => o.role_in_org === 'owner' || o.role_in_org === 'admin'),
)

async function loadOrgs() {
  orgLoading.value = true
  orgError.value = null
  try {
    const r = await api.get('/organizations')
    orgs.value = Array.isArray(r.data) ? r.data : []
    if (adminOrgs.value.length && orgInviteTargetId.value === '') {
      orgInviteTargetId.value = String(adminOrgs.value[0].id)
    }
  } catch (e) {
    orgError.value = e.response?.data?.error ?? 'Could not load organizations'
    orgs.value = []
  } finally {
    orgLoading.value = false
  }
}

async function createOrg() {
  orgError.value = null
  orgSaving.value = true
  try {
    await api.post('/organizations', { name: newOrgName.value.trim() })
    newOrgName.value = ''
    await loadOrgs()
  } catch (e) {
    orgError.value = e.response?.data?.error ?? 'Could not create organization'
  } finally {
    orgSaving.value = false
  }
}

async function loadOrgUsage(orgId) {
  try {
    const r = await api.get(`/organizations/${orgId}/usage-summary`)
    orgUsage.value = r.data
  } catch {
    orgUsage.value = null
  }
}

async function saveFarmOrgLink() {
  farmOrgMsg.value = ''
  farmOrgSaving.value = true
  try {
    const fid = farmContext.farmId
    const v = farmOrgLink.value
    await api.patch(`/farms/${fid}/organization`, {
      organization_id: v === '' ? null : Number(v),
    })
    await farmStore.loadAll(fid)
    await farmContext.fetchFarms()
    farmOrgMsg.value = 'Saved.'
  } catch (e) {
    farmOrgMsg.value = e.response?.data?.error ?? 'Could not update farm organization'
  } finally {
    farmOrgSaving.value = false
  }
}

async function inviteOrgMember() {
  orgInviteError.value = null
  orgInviteOk.value = false
  orgInviting.value = true
  try {
    const oid = Number(orgInviteTargetId.value)
    await api.post(`/organizations/${oid}/members`, {
      email: orgInviteEmail.value.trim(),
      role_in_org: orgInviteRole.value,
    })
    orgInviteOk.value = true
    orgInviteEmail.value = ''
    await loadOrgs()
  } catch (e) {
    orgInviteError.value = e.response?.data?.error ?? 'Could not add member'
  } finally {
    orgInviting.value = false
  }
}

watch(
  () => farmStore.farm,
  (f) => {
    if (!f) {
      farmOrgLink.value = ''
      return
    }
    farmOrgLink.value = f.organization_id != null ? String(f.organization_id) : ''
  },
  { immediate: true },
)

// ── Password change ──────────────────────────────────────────────────────────
const pwForm    = reactive({ current: '', next: '', confirm: '' })
const pwLoading = ref(false)
const pwError   = ref(null)
const pwSuccess = ref(false)

const submitPassword = async () => {
  pwError.value   = null
  pwSuccess.value = false
  pwLoading.value = true
  try {
    await api.patch('/auth/password', {
      current_password: pwForm.current,
      new_password:     pwForm.next,
    })
    pwSuccess.value = true
    resetPwForm()
  } catch (e) {
    pwError.value = e.response?.data?.error ?? 'Password update failed'
  } finally {
    pwLoading.value = false
  }
}

const resetPwForm = () => {
  pwForm.current = ''
  pwForm.next    = ''
  pwForm.confirm = ''
  pwError.value  = null
}

// ── Farm Members ─────────────────────────────────────────────────────────────
const members = ref([])
const membersLoading = ref(false)
const inviteEmail = ref('')
const inviteRole = ref('viewer')
const inviting = ref(false)
const inviteError = ref(null)
const inviteSuccess = ref(false)

const insertOptIn = ref(false)
const insertSyncing = ref(false)
const insertSyncMsg = ref('')
const insertHistory = ref([])
const insertHistoryLoading = ref(false)

function fmtTs(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return String(iso)
  return d.toLocaleString()
}

async function loadInsertHistory() {
  if (!farmContext.farmId || !insertOptIn.value) return
  insertHistoryLoading.value = true
  try {
    insertHistory.value = await farmStore.listInsertCommonsSyncEvents(farmContext.farmId, { limit: 8, offset: 0 })
  } catch {
    insertHistory.value = []
  } finally {
    insertHistoryLoading.value = false
  }
}

async function loadFarmSharing() {
  if (!farmContext.farmId) return
  try {
    await farmStore.loadAll(farmContext.farmId)
    insertOptIn.value = !!farmStore.farm?.insert_commons_opt_in
    await loadInsertHistory()
  } catch { /* ignore */ }
}

async function onInsertOptInChange() {
  if (!farmContext.farmId) return
  insertSyncMsg.value = ''
  try {
    await farmStore.setInsertCommonsOptIn(farmContext.farmId, insertOptIn.value)
    insertSyncMsg.value = insertOptIn.value ? 'Sharing enabled.' : 'Sharing disabled.'
    if (!insertOptIn.value) {
      insertHistory.value = []
    } else {
      await loadInsertHistory()
    }
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Could not update setting'
  }
}

async function runInsertSync() {
  if (!farmContext.farmId || !insertOptIn.value) return
  insertSyncing.value = true
  insertSyncMsg.value = ''
  try {
    const r = await farmStore.insertCommonsSync(farmContext.farmId)
    const status = r.delivery_status ? String(r.delivery_status) : 'unknown'
    const http = r.http_status != null ? ` (HTTP ${r.http_status})` : ''
    insertSyncMsg.value = `${status}${http} — ${r.privacy_notice || 'Sync recorded.'}`
    await farmStore.loadAll(farmContext.farmId)
    await loadInsertHistory()
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Sync failed'
  } finally {
    insertSyncing.value = false
  }
}

async function loadMembers() {
  if (!farmContext.farmId) return
  membersLoading.value = true
  try {
    members.value = await farmStore.loadFarmMembers(farmContext.farmId)
  } finally {
    membersLoading.value = false
  }
}

async function inviteMember() {
  inviteError.value = null
  inviteSuccess.value = false
  inviting.value = true
  try {
    await farmStore.addFarmMember(farmContext.farmId, {
      email: inviteEmail.value,
      role_in_farm: inviteRole.value,
    })
    inviteSuccess.value = true
    inviteEmail.value = ''
    await loadMembers()
  } catch (e) {
    inviteError.value = e.response?.data?.error ?? 'Invite failed'
  } finally {
    inviting.value = false
  }
}

async function changeRole(userId, role) {
  try {
    await farmStore.updateFarmMemberRole(farmContext.farmId, userId, role)
    await loadMembers()
  } catch {}
}

async function removeMember(userId) {
  try {
    await farmStore.removeFarmMember(farmContext.farmId, userId)
    await loadMembers()
  } catch {}
}

onMounted(() => {
  loadOrgs()
  loadMembers()
  loadFarmSharing()
})
watch(() => farmContext.farmId, () => {
  loadOrgs()
  loadMembers()
  loadFarmSharing()
})

// ── Pi API key display ───────────────────────────────────────────────────────
const showKey  = ref(false)
const piApiKey = '(configured on server — check PI_API_KEY env var)'

// ── Token expiry ─────────────────────────────────────────────────────────────
const tokenExpiry = computed(() => {
  const token = auth.token
  if (!token) return 'unknown'
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    const diff    = payload.exp - Math.floor(Date.now() / 1000)
    if (diff <= 0) return 'expired'
    const h = Math.floor(diff / 3600)
    const m = Math.floor((diff % 3600) / 60)
    return h > 0 ? `${h}h ${m}m` : `${m}m`
  } catch { return 'unknown' }
})

// ── Sign out ─────────────────────────────────────────────────────────────────
const signOut = () => {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<style scoped>
.input-field {
  @apply bg-zinc-900 border border-zinc-700 rounded-lg px-4 py-2.5 text-white text-sm
         placeholder-zinc-600 focus:outline-none focus:border-green-500 transition-colors w-full;
}
</style>
