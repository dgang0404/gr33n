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

    <!-- Create farm (bootstrap template picker) -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3">New farm</h2>
      <p class="text-zinc-500 text-xs mb-4">
        Create a farm you own. Starter packs are idempotent — applying the same template again does not duplicate data.
        If you link an organization with a default template, you can start from that default without picking a pack here.
      </p>
      <form class="space-y-3" @submit.prevent="submitNewFarm">
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">Farm name</label>
          <input v-model="newFarm.name" type="text" required placeholder="e.g. North greenhouse"
            class="input-field text-sm" />
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div class="flex flex-col gap-1.5">
            <label class="text-zinc-400 text-xs uppercase tracking-wide">Timezone</label>
            <input v-model="newFarm.timezone" type="text" required class="input-field text-sm" />
          </div>
          <div class="flex flex-col gap-1.5">
            <label class="text-zinc-400 text-xs uppercase tracking-wide">Currency</label>
            <input v-model="newFarm.currency" type="text" required maxlength="3"
              class="input-field text-sm uppercase" />
          </div>
        </div>
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">Organization (optional)</label>
          <select v-model="newFarm.organizationId"
            class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-sm rounded-lg px-3 py-2 focus:outline-none">
            <option value="">— None —</option>
            <option v-for="o in adminOrgs" :key="o.id" :value="String(o.id)">{{ o.name }}</option>
          </select>
        </div>
        <fieldset class="space-y-2 border border-zinc-700 rounded-lg p-3 bg-zinc-900/50">
          <legend class="text-zinc-400 text-xs uppercase tracking-wide px-1">Starting content</legend>
          <label class="flex items-start gap-2 text-sm text-zinc-300 cursor-pointer">
            <input v-model="newFarm.bootstrapMode" type="radio" value="blank" class="mt-1" />
            <span>Start blank — empty zones and schedules (explicit). Use this to skip an organization default.</span>
          </label>
          <label class="flex items-start gap-2 text-sm text-zinc-300 cursor-pointer">
            <input v-model="newFarm.bootstrapMode" type="radio" value="starter" class="mt-1" />
            <span>Apply starter pack now</span>
          </label>
          <label v-if="newFarm.organizationId" class="flex items-start gap-2 text-sm text-zinc-300 cursor-pointer">
            <input v-model="newFarm.bootstrapMode" type="radio" value="org_default" class="mt-1" />
            <span>Use organization default only (omit template in the request). Falls back to blank if the org has no default.</span>
          </label>
          <div v-if="newFarm.bootstrapMode === 'starter'" class="pl-6 space-y-2">
            <select v-model="newFarm.starterKey"
              class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-sm rounded-lg px-3 py-2 w-full max-w-md focus:outline-none">
              <option v-for="opt in starterPackOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
            <details v-if="newFarm.starterKey === jadamKey" class="text-zinc-500 text-xs">
              <summary class="text-gr33n-400 cursor-pointer select-none">{{ jadamSummary.title }}</summary>
              <ul class="list-disc pl-5 mt-2 space-y-1">
                <li v-for="(b, i) in jadamSummary.bullets" :key="i">{{ b }}</li>
              </ul>
            </details>
          </div>
        </fieldset>
        <p v-if="newFarmError" class="text-red-400 text-xs">{{ newFarmError }}</p>
        <p v-if="newFarmOk" class="text-green-400 text-xs">{{ newFarmOk }}</p>
        <button type="submit" :disabled="newFarmSaving || !auth.userId"
          class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 text-white text-sm font-semibold px-5 py-2 rounded-lg">
          {{ newFarmSaving ? 'Creating…' : 'Create farm' }}
        </button>
      </form>
    </section>

    <!-- Apply starter to existing farm (farm admin) -->
    <section v-if="farmContext.farmId"
      class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3">Current farm — apply starter pack</h2>
      <p class="text-zinc-500 text-xs mb-4">
        Use this if the farm was created blank and you want demo zones, inventory lots,
        fertigation (reservoirs, programs, schedules, mixing log), and a task linked to an irrigation schedule.
        Farm admins only. If the pack was already applied, the API returns “already applied” and leaves data unchanged.
      </p>
      <div class="flex flex-wrap items-end gap-3">
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs uppercase tracking-wide">Template</label>
          <select v-model="applyStarterKey"
            class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-sm rounded-lg px-3 py-2 min-w-[14rem]">
            <option v-for="opt in starterPackOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
        </div>
        <button type="button" :disabled="applyStarterSaving"
          class="bg-amber-700 hover:bg-amber-600 disabled:bg-zinc-700 text-white text-sm font-semibold px-4 py-2 rounded-lg"
          @click="submitApplyStarter">
          {{ applyStarterSaving ? 'Applying…' : 'Apply to current farm' }}
        </button>
      </div>
      <p v-if="applyStarterError" class="text-red-400 text-xs mt-2">{{ applyStarterError }}</p>
      <p v-if="applyStarterMsg" class="text-green-400 text-xs mt-2">{{ applyStarterMsg }}</p>
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
      <p v-if="orgDefaultError" class="text-red-400 text-xs mb-2">{{ orgDefaultError }}</p>
      <div v-if="orgLoading && !orgs.length" class="text-zinc-500 text-sm">Loading…</div>
      <ul v-else-if="orgs.length" class="space-y-2 mb-4">
        <li v-for="o in orgs" :key="o.id"
          class="flex flex-col gap-2 bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-sm">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <span class="text-zinc-200 font-medium">{{ o.name }}</span>
              <span class="text-zinc-500 text-xs ml-2">{{ o.role_in_org }} · {{ o.plan_tier }}</span>
            </div>
            <button type="button" class="text-xs text-green-500 hover:text-green-400"
              @click="loadOrgUsage(o.id)">
              Usage summary
            </button>
          </div>
          <div v-if="o.role_in_org === 'owner' || o.role_in_org === 'admin'"
            class="flex flex-wrap items-center gap-2 text-xs border-t border-zinc-800 pt-2">
            <span class="text-zinc-500 shrink-0">Default template for new farms:</span>
            <select v-model="orgDefaultDraft[o.id]"
              class="bg-zinc-950 border border-zinc-600 text-zinc-300 rounded px-2 py-1 min-w-[10rem] focus:outline-none">
              <option value="">None (blank)</option>
              <option :value="jadamKey">Indoor photoperiod v1</option>
            </select>
            <button type="button" :disabled="orgDefaultSaving === o.id"
              class="text-xs bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 text-white px-3 py-1 rounded"
              @click="saveOrgBootstrapDefault(o.id)">
              {{ orgDefaultSaving === o.id ? 'Saving…' : 'Save' }}
            </button>
          </div>
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

      <div v-if="adminOrgs.length" class="border-t border-zinc-700 pt-4 mt-4">
        <h3 class="text-zinc-200 text-sm font-semibold mb-1">Organization audit</h3>
        <p class="text-zinc-500 text-xs mb-3">
          Cross-farm and org-only events (settings, membership, exports). Requires org owner or admin.
        </p>
        <div class="flex flex-wrap gap-2 items-center mb-3">
          <select v-model="auditOrgId"
            class="bg-zinc-900 border border-zinc-700 text-zinc-300 text-sm rounded-lg px-3 py-2 min-w-[12rem] focus:outline-none"
            @change="loadOrgAudit">
            <option v-for="o in adminOrgs" :key="o.id" :value="String(o.id)">{{ o.name }}</option>
          </select>
          <button type="button" :disabled="orgAuditLoading" @click="loadOrgAudit"
            class="text-xs bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 text-white px-3 py-2 rounded-lg">
            {{ orgAuditLoading ? 'Loading…' : 'Refresh' }}
          </button>
        </div>
        <p v-if="orgAuditError" class="text-red-400 text-xs mb-2">{{ orgAuditError }}</p>
        <div v-if="orgAuditLoading && !orgAuditEvents.length" class="text-zinc-500 text-sm">Loading…</div>
        <div v-else-if="!orgAuditEvents.length" class="text-zinc-600 text-sm">No audit events for this org yet.</div>
        <div v-else class="overflow-x-auto border border-zinc-700 rounded-lg max-h-80 overflow-y-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-zinc-900 text-zinc-400 sticky top-0">
              <tr>
                <th class="px-2 py-2 font-medium">Time</th>
                <th class="px-2 py-2 font-medium">Kind</th>
                <th class="px-2 py-2 font-medium">Action</th>
                <th class="px-2 py-2 font-medium">Farm</th>
                <th class="px-2 py-2 font-medium">Target</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-zinc-800">
              <tr v-for="ev in orgAuditEvents" :key="ev.id" class="text-zinc-300 hover:bg-zinc-900/80">
                <td class="px-2 py-1.5 whitespace-nowrap text-zinc-500">{{ fmtTs(ev.activity_time) }}</td>
                <td class="px-2 py-1.5 font-mono text-gr33n-400/90">{{ auditEventKind(ev) }}</td>
                <td class="px-2 py-1.5">{{ ev.action_type }}</td>
                <td class="px-2 py-1.5 text-zinc-500">{{ ev.farm_id != null ? ev.farm_id : '—' }}</td>
                <td class="px-2 py-1.5 max-w-[14rem] truncate" :title="auditTarget(ev)">{{ auditTarget(ev) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
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

      <!-- Farm-scoped audit (owner/manager) -->
      <div v-if="farmContext.farmId && isFarmAdmin" class="border-t border-zinc-700 pt-4 mt-4">
        <h3 class="text-zinc-200 text-sm font-semibold mb-1">Farm audit</h3>
        <p class="text-zinc-500 text-xs mb-3">
          Actions on this farm (membership, settings, exports, Insert Commons sync, etc.). Newest first.
        </p>
        <button type="button" :disabled="farmAuditLoading" @click="loadFarmAudit"
          class="text-xs bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 text-white px-3 py-2 rounded-lg mb-3">
          {{ farmAuditLoading ? 'Loading…' : 'Refresh' }}
        </button>
        <p v-if="farmAuditError" class="text-red-400 text-xs mb-2">{{ farmAuditError }}</p>
        <div v-if="farmAuditLoading && !farmAuditEvents.length" class="text-zinc-500 text-sm">Loading…</div>
        <div v-else-if="!farmAuditEvents.length" class="text-zinc-600 text-sm">No audit events for this farm yet.</div>
        <div v-else class="overflow-x-auto border border-zinc-700 rounded-lg max-h-80 overflow-y-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-zinc-900 text-zinc-400 sticky top-0">
              <tr>
                <th class="px-2 py-2 font-medium">Time</th>
                <th class="px-2 py-2 font-medium">Kind</th>
                <th class="px-2 py-2 font-medium">Action</th>
                <th class="px-2 py-2 font-medium">Target</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-zinc-800">
              <tr v-for="ev in farmAuditEvents" :key="ev.id" class="text-zinc-300 hover:bg-zinc-900/80">
                <td class="px-2 py-1.5 whitespace-nowrap text-zinc-500">{{ fmtTs(ev.activity_time) }}</td>
                <td class="px-2 py-1.5 font-mono text-gr33n-400/90">{{ auditEventKind(ev) }}</td>
                <td class="px-2 py-1.5">{{ ev.action_type }}</td>
                <td class="px-2 py-1.5 max-w-[18rem] truncate" :title="auditTarget(ev)">{{ auditTarget(ev) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

    <!-- Insert Commons (benchmark sharing) -->
    <section v-if="farmContext.farmId" class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3">Insert Commons</h2>
      <p class="text-zinc-400 text-sm mb-3">
        Optional community benchmarks. The API only sends <span class="text-zinc-300">coarse aggregates</span> under a stable per-farm pseudonym
        (not your farm name). If <code class="text-green-400">INSERT_COMMONS_INGEST_URL</code> is unset, the server records an attempt but does not call out.
        You can revoke anytime by turning sharing off.
      </p>
      <p v-if="!canViewInsertCommons" class="text-amber-200/90 text-xs mb-3">
        Running sync, viewing history, and exporting bundles need <strong class="font-medium">owner</strong>, <strong class="font-medium">manager</strong>, or <strong class="font-medium">finance</strong> access on this farm.
      </p>
      <label class="flex items-center gap-2 text-zinc-300 text-sm mb-3">
        <input v-model="insertOptIn" type="checkbox" class="rounded bg-zinc-800 border-zinc-700"
          @change="onInsertOptInChange" />
        Share anonymized aggregates with Insert Commons
      </label>
      <label        v-if="insertOptIn && isFarmAdmin"
        class="flex items-center gap-2 text-zinc-300 text-sm mb-3"
      >
        <input
          v-model="insertRequireApproval"
          type="checkbox"
          class="rounded bg-zinc-800 border-zinc-700"
          @change="onInsertRequireApprovalChange"
        />
        Require owner/manager approval before each payload is sent to the ingest URL
      </label>
      <p
        v-else-if="insertOptIn && insertRequireApproval && !isFarmAdmin"
        class="text-amber-200/90 text-xs mb-3"
      >
        This farm is in <span class="font-medium">approval queue</span> mode. A farm owner or manager must approve each payload before it is sent.
      </p>
      <div v-if="isFarmAdmin" class="mb-4">
        <button
          type="button"
          :disabled="insertPreviewLoading || !farmContext.farmId"
          class="bg-zinc-800 hover:bg-zinc-700 border border-zinc-600 disabled:opacity-40 text-zinc-200 text-xs font-semibold px-4 py-2 rounded-lg mr-2"
          @click="loadInsertPreview"
        >
          {{ insertPreviewLoading ? 'Building preview…' : 'Preview ingest JSON' }}
        </button>
        <p class="text-zinc-500 text-xs mt-2">
          Read-only: same validated payload shape as sync, without sending or saving history.
        </p>
        <p v-if="insertPreviewError" class="text-red-400 text-xs mt-1">{{ insertPreviewError }}</p>
        <details v-if="insertPreviewData && insertPreviewData.valid" class="mt-2">
          <summary class="text-gr33n-400 text-xs cursor-pointer select-none">Show payload JSON</summary>
          <pre class="mt-2 p-2 bg-zinc-950 border border-zinc-800 rounded text-[10px] text-zinc-400 overflow-x-auto max-h-64 whitespace-pre-wrap break-words">{{ insertPreviewJson }}</pre>
        </details>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          :disabled="!insertOptIn || insertSyncing || !canViewInsertCommons"
          class="bg-zinc-700 hover:bg-zinc-600 disabled:opacity-40 text-white text-xs font-semibold px-4 py-2 rounded-lg"
          @click="runInsertSync"
        >
          {{ insertSyncing ? 'Syncing…' : 'Run sync' }}
        </button>
        <span v-if="insertSyncMsg" class="text-zinc-500 text-xs">{{ insertSyncMsg }}</span>
      </div>
      <div v-if="insertOptIn && canViewInsertCommons" class="mt-4">
        <div class="flex items-center justify-between mb-2">
          <p class="text-zinc-500 text-xs uppercase tracking-wide">Pending &amp; recent bundles</p>
          <button
            type="button"
            class="text-zinc-500 hover:text-white text-xs"
            @click="loadInsertBundles"
            :disabled="insertBundlesLoading"
          >
            {{ insertBundlesLoading ? 'Loading…' : 'Refresh' }}
          </button>
        </div>
        <div v-if="insertBundlesLoading && insertBundles.length === 0" class="text-zinc-600 text-xs">Loading bundles…</div>
        <div v-else-if="insertBundles.length === 0" class="text-zinc-600 text-xs">No bundles yet. Run sync to create one.</div>
        <ul v-else class="space-y-2">
          <li
            v-for="b in insertBundles"
            :key="b.id"
            class="bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-xs"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <span class="text-zinc-300 font-mono">#{{ b.id }} · {{ b.status }}</span>
              <span class="text-zinc-600 shrink-0">{{ fmtTs(b.created_at) }}</span>
            </div>
            <div class="text-zinc-600 mt-1 break-all" v-if="b.idempotency_key">idem: {{ b.idempotency_key }}</div>
            <div class="text-zinc-600 mt-1" v-if="b.delivery_http_status != null">delivery http: {{ b.delivery_http_status }}</div>
            <div class="text-red-300 mt-1 break-words" v-if="b.delivery_error">{{ b.delivery_error }}</div>
            <div class="flex flex-wrap gap-2 mt-2">
              <button
                v-if="isFarmAdmin && b.status === 'pending_approval'"
                type="button"
                class="text-green-400 hover:text-green-300 text-xs font-medium disabled:opacity-40"
                :disabled="insertBundleBusy === b.id"
                @click="approveInsertBundle(b.id)"
              >
                Approve &amp; send
              </button>
              <template v-if="isFarmAdmin && b.status === 'pending_approval'">
                <button
                  v-if="rejectExpandId !== b.id"
                  type="button"
                  class="text-red-400 hover:text-red-300 text-xs font-medium"
                  @click="openRejectBundle(b.id)"
                >
                  Reject…
                </button>
              </template>
              <button
                v-if="isFarmAdmin && b.status === 'delivery_failed'"
                type="button"
                class="text-amber-400 hover:text-amber-300 text-xs font-medium disabled:opacity-40"
                :disabled="insertBundleBusy === b.id"
                @click="retryInsertBundleDeliver(b.id)"
              >
                Retry delivery
              </button>
              <button
                type="button"
                class="text-zinc-400 hover:text-white text-xs font-medium disabled:opacity-40"
                :disabled="insertBundleBusy === b.id"
                @click="farmStore.downloadInsertCommonsBundleExport(farmContext.farmId, b.id, 'ingest')"
              >
                Export ingest JSON
              </button>
              <button
                type="button"
                class="text-zinc-400 hover:text-white text-xs font-medium disabled:opacity-40"
                :disabled="insertBundleBusy === b.id"
                @click="farmStore.downloadInsertCommonsBundleExport(farmContext.farmId, b.id, 'package_v1')"
              >
                Export package v1
              </button>
            </div>
            <div v-if="rejectExpandId === b.id" class="mt-2 space-y-2 border-t border-zinc-700 pt-2">
              <textarea
                v-model="rejectNote"
                rows="2"
                placeholder="Reason for rejection (required)"
                class="w-full bg-zinc-950 border border-zinc-600 rounded px-2 py-1.5 text-zinc-200 text-xs placeholder-zinc-600"
              />
              <div class="flex gap-2">
                <button
                  type="button"
                  class="bg-red-900/80 hover:bg-red-800 text-white text-xs font-semibold px-3 py-1 rounded disabled:opacity-40"
                  :disabled="insertBundleBusy === b.id || !rejectNote.trim()"
                  @click="confirmRejectBundle(b.id)"
                >
                  Confirm reject
                </button>
                <button type="button" class="text-zinc-500 hover:text-white text-xs" @click="cancelRejectBundle">
                  Cancel
                </button>
              </div>
            </div>
          </li>
        </ul>
      </div>
      <div v-if="insertOptIn && canViewInsertCommons" class="mt-4">
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
            <div class="text-zinc-600 mt-1" v-if="e.bundle_id != null">bundle: {{ e.bundle_id }}</div>
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
    <!-- Push Notifications -->
    <section class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 mb-5">
      <h2 class="text-white font-semibold mb-3 flex items-center gap-2">
        <span>🔔</span> Push Notifications
      </h2>
      <p class="text-zinc-500 text-xs mb-4">
        Enable push notifications to receive alerts and task reminders on your device.
        On native (Android/iOS) this uses Firebase Cloud Messaging.
      </p>
      <div class="flex items-center gap-3 mb-4">
        <label class="flex items-center gap-2 text-zinc-300 text-sm">
          <input v-model="pushEnabled" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" @change="onPushToggle" />
          Push notifications enabled
        </label>
        <span v-if="pushSaving" class="text-xs text-zinc-500">Saving…</span>
        <span v-if="pushError" class="text-xs text-red-400">{{ pushError }}</span>
      </div>
      <div v-if="pushTokens.length" class="space-y-2">
        <p class="text-zinc-400 text-xs uppercase tracking-wide mb-1">Registered devices</p>
        <div v-for="tok in pushTokens" :key="tok.id"
          class="flex items-center justify-between bg-zinc-900 border border-zinc-700 rounded-lg px-3 py-2 text-xs">
          <div class="min-w-0">
            <span class="text-zinc-200 font-mono truncate block max-w-xs">{{ tok.fcm_token?.slice(0, 24) }}…</span>
            <span class="text-zinc-500">{{ tok.platform }} · {{ formatPushDate(tok.created_at) }}</span>
          </div>
          <button @click="removePushToken(tok.fcm_token)" class="text-zinc-500 hover:text-red-400 shrink-0">✕</button>
        </div>
      </div>
      <p v-else class="text-zinc-600 text-xs">No push tokens registered.</p>
    </section>

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
import {
  BOOTSTRAP_STARTER_OPTIONS,
  BOOTSTRAP_TEMPLATE_KEYS,
  JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY,
} from '../constants/bootstrapTemplates'

const router = useRouter()
const auth   = useAuthStore()
const farmStore = useFarmStore()
const farmContext = useFarmContextStore()

const starterPackOptions = BOOTSTRAP_STARTER_OPTIONS
const jadamKey = BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1
const jadamSummary = JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY

const newFarm = reactive({
  name: '',
  timezone: 'UTC',
  currency: 'USD',
  organizationId: '',
  bootstrapMode: 'blank',
  starterKey: BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1,
})
const newFarmSaving = ref(false)
const newFarmError = ref(null)
const newFarmOk = ref(null)

const applyStarterKey = ref(BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1)
const applyStarterSaving = ref(false)
const applyStarterError = ref(null)
const applyStarterMsg = ref(null)

async function submitApplyStarter() {
  applyStarterError.value = null
  applyStarterMsg.value = null
  const fid = farmContext.farmId
  if (!fid) {
    applyStarterError.value = 'No farm selected'
    return
  }
  applyStarterSaving.value = true
  try {
    const data = await farmContext.applyBootstrapTemplate(fid, applyStarterKey.value)
    const b = data?.bootstrap
    if (b?.already_applied) {
      applyStarterMsg.value = 'This template was already applied to this farm.'
    } else if (b?.applied) {
      applyStarterMsg.value = 'Starter pack applied. Open Fertigation, Inventory, and Tasks to see linked data.'
    } else if (b?.error) {
      applyStarterError.value = String(b.error)
    } else {
      applyStarterMsg.value = 'Done.'
    }
  } catch (e) {
    applyStarterError.value = e.response?.data?.error ?? e.message ?? 'Could not apply starter'
  } finally {
    applyStarterSaving.value = false
  }
}

watch(
  () => newFarm.organizationId,
  (v) => {
    if (!v && newFarm.bootstrapMode === 'org_default') {
      newFarm.bootstrapMode = 'blank'
    }
  },
)

async function submitNewFarm() {
  newFarmError.value = null
  newFarmOk.value = null
  if (!auth.userId) {
    newFarmError.value = 'Sign in required'
    return
  }
  if (newFarm.bootstrapMode === 'org_default' && !newFarm.organizationId) {
    newFarmError.value = 'Select an organization to use its default template'
    return
  }
  newFarmSaving.value = true
  try {
    const payload = {
      name: newFarm.name.trim(),
      owner_user_id: auth.userId,
      timezone: newFarm.timezone.trim(),
      currency: newFarm.currency.trim().toUpperCase(),
      operational_status: 'active',
      scale_tier: 'small',
    }
    if (newFarm.organizationId) {
      payload.organization_id = Number(newFarm.organizationId)
    }
    if (newFarm.bootstrapMode === 'blank') {
      payload.bootstrap_template = 'none'
    } else if (newFarm.bootstrapMode === 'starter') {
      payload.bootstrap_template = newFarm.starterKey
    }

    const { farm, bootstrap } = await farmContext.createFarm(payload)
    await farmContext.selectFarm(farm.id)
    let msg = `Farm “${farm.name}” created.`
    if (bootstrap && typeof bootstrap === 'object') {
      if (bootstrap.skipped) msg += ' No starter data applied.'
      else if (bootstrap.already_applied) msg += ' Template was already applied for this farm.'
      else if (bootstrap.error) msg += ` Starter: ${bootstrap.error}`
      else msg += ' Starter pack applied.'
    }
    newFarmOk.value = msg
    newFarm.name = ''
  } catch (e) {
    newFarmError.value = e.response?.data?.error ?? 'Could not create farm'
  } finally {
    newFarmSaving.value = false
  }
}

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

const orgDefaultDraft = reactive({})
const orgDefaultSaving = ref(null)
const orgDefaultError = ref(null)

const auditOrgId = ref('')
const orgAuditEvents = ref([])
const orgAuditLoading = ref(false)
const orgAuditError = ref(null)

async function loadOrgAudit() {
  if (!auditOrgId.value) return
  orgAuditLoading.value = true
  orgAuditError.value = null
  try {
    const r = await api.get(`/organizations/${auditOrgId.value}/audit-events`, { params: { limit: 50 } })
    orgAuditEvents.value = Array.isArray(r.data) ? r.data : []
  } catch (e) {
    orgAuditError.value = e.response?.data?.error ?? 'Could not load organization audit'
    orgAuditEvents.value = []
  } finally {
    orgAuditLoading.value = false
  }
}

watch(
  adminOrgs,
  (list) => {
    if (!list.length) {
      orgAuditEvents.value = []
      return
    }
    if (!auditOrgId.value || !list.some((o) => String(o.id) === auditOrgId.value)) {
      auditOrgId.value = String(list[0].id)
    }
    loadOrgAudit()
  },
  { flush: 'post' },
)

async function saveOrgBootstrapDefault(orgId) {
  orgDefaultError.value = null
  orgDefaultSaving.value = orgId
  try {
    const raw = orgDefaultDraft[orgId] ?? ''
    await api.patch(`/organizations/${orgId}`, {
      default_bootstrap_template: raw === '' ? null : raw,
    })
    await loadOrgs()
  } catch (e) {
    orgDefaultError.value = e.response?.data?.error ?? 'Could not update org default'
  } finally {
    orgDefaultSaving.value = null
  }
}

async function loadOrgs() {
  orgLoading.value = true
  orgError.value = null
  try {
    const r = await api.get('/organizations')
    orgs.value = Array.isArray(r.data) ? r.data : []
    for (const o of orgs.value) {
      orgDefaultDraft[o.id] = o.default_bootstrap_template || ''
    }
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
const insertRequireApproval = ref(false)
const insertSyncing = ref(false)
const insertSyncMsg = ref('')
const insertPreviewLoading = ref(false)
const insertPreviewError = ref(null)
const insertPreviewData = ref(null)

const insertPreviewJson = computed(() => {
  const p = insertPreviewData.value?.payload
  if (!p) return ''
  try {
    return JSON.stringify(p, null, 2)
  } catch {
    return String(p)
  }
})
const insertHistory = ref([])
const insertHistoryLoading = ref(false)
const insertBundles = ref([])
const insertBundlesLoading = ref(false)
const insertBundleBusy = ref(null)
const rejectExpandId = ref(null)
const rejectNote = ref('')

const isFarmAdmin = computed(() => {
  const uid = auth.userId
  const fid = farmContext.farmId
  if (!uid || !fid) return false
  const f = farmStore.farm
  if (
    f &&
    Number(f.id) === Number(fid) &&
    f.owner_user_id &&
    String(f.owner_user_id).toLowerCase() === String(uid).toLowerCase()
  ) {
    return true
  }
  const m = members.value.find((x) => String(x.user_id).toLowerCase() === String(uid).toLowerCase())
  return !!(m && (m.role_in_farm === 'owner' || m.role_in_farm === 'manager'))
})

const farmAuditEvents = ref([])
const farmAuditLoading = ref(false)
const farmAuditError = ref(null)

async function loadFarmAudit() {
  const fid = farmContext.farmId
  if (!fid || !isFarmAdmin.value) {
    farmAuditEvents.value = []
    return
  }
  farmAuditLoading.value = true
  farmAuditError.value = null
  try {
    const r = await api.get(`/farms/${fid}/audit-events`, { params: { limit: 50 } })
    farmAuditEvents.value = Array.isArray(r.data) ? r.data : []
  } catch (e) {
    farmAuditError.value = e.response?.data?.error ?? 'Could not load farm audit'
    farmAuditEvents.value = []
  } finally {
    farmAuditLoading.value = false
  }
}

watch(
  () => [farmContext.farmId, isFarmAdmin.value],
  () => {
    loadFarmAudit()
  },
  { flush: 'post', immediate: true },
)

const canViewInsertCommons = computed(() => {
  const uid = auth.userId
  const fid = farmContext.farmId
  if (!uid || !fid) return false
  const f = farmStore.farm
  if (
    f &&
    Number(f.id) === Number(fid) &&
    f.owner_user_id &&
    String(f.owner_user_id).toLowerCase() === String(uid).toLowerCase()
  ) {
    return true
  }
  const m = members.value.find((x) => String(x.user_id).toLowerCase() === String(uid).toLowerCase())
  return !!(m && ['owner', 'manager', 'finance'].includes(m.role_in_farm))
})

function fmtTs(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return String(iso)
  return d.toLocaleString()
}

function auditEventKind(ev) {
  const d = ev?.details
  if (d && typeof d === 'object' && d.kind != null) return String(d.kind)
  return '—'
}

function auditTarget(ev) {
  if (!ev) return '—'
  const parts = []
  if (ev.target_table_name) parts.push(ev.target_table_name)
  if (ev.target_record_id) parts.push(ev.target_record_id)
  return parts.length ? parts.join(' · ') : '—'
}

async function loadInsertHistory() {
  if (!farmContext.farmId || !insertOptIn.value || !canViewInsertCommons.value) return
  insertHistoryLoading.value = true
  try {
    insertHistory.value = await farmStore.listInsertCommonsSyncEvents(farmContext.farmId, { limit: 8, offset: 0 })
  } catch {
    insertHistory.value = []
  } finally {
    insertHistoryLoading.value = false
  }
}

async function loadInsertBundles() {
  if (!farmContext.farmId || !insertOptIn.value || !canViewInsertCommons.value) return
  insertBundlesLoading.value = true
  try {
    insertBundles.value = await farmStore.listInsertCommonsBundles(farmContext.farmId, { limit: 25, offset: 0 })
  } catch {
    insertBundles.value = []
  } finally {
    insertBundlesLoading.value = false
  }
}

async function loadFarmSharing() {
  if (!farmContext.farmId) return
  try {
    await farmStore.loadAll(farmContext.farmId)
    insertOptIn.value = !!farmStore.farm?.insert_commons_opt_in
    insertRequireApproval.value = !!farmStore.farm?.insert_commons_require_approval
    if (!insertOptIn.value) {
      insertHistory.value = []
      insertBundles.value = []
    } else if (canViewInsertCommons.value) {
      await Promise.all([loadInsertHistory(), loadInsertBundles()])
    }
  } catch { /* ignore */ }
}

async function onInsertOptInChange() {
  if (!farmContext.farmId) return
  insertSyncMsg.value = ''
  try {
    const payload = { insert_commons_opt_in: insertOptIn.value }
    if (insertOptIn.value) {
      payload.insert_commons_require_approval = insertRequireApproval.value
    }
    await farmStore.setInsertCommonsOptIn(farmContext.farmId, payload)
    insertSyncMsg.value = insertOptIn.value ? 'Sharing enabled.' : 'Sharing disabled.'
    if (!insertOptIn.value) {
      insertHistory.value = []
      insertBundles.value = []
    } else if (canViewInsertCommons.value) {
      await Promise.all([loadInsertHistory(), loadInsertBundles()])
    }
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Could not update setting'
  }
}

async function onInsertRequireApprovalChange() {
  if (!farmContext.farmId || !isFarmAdmin.value) return
  insertSyncMsg.value = ''
  try {
    await farmStore.setInsertCommonsOptIn(farmContext.farmId, {
      insert_commons_opt_in: true,
      insert_commons_require_approval: insertRequireApproval.value,
    })
    insertSyncMsg.value = insertRequireApproval.value
      ? 'Approval queue enabled.'
      : 'Payloads will send on sync when ingest is configured (no approval step).'
    await farmStore.loadAll(farmContext.farmId)
    await Promise.all([loadInsertBundles(), loadInsertHistory()])
  } catch (e) {
    await loadFarmSharing()
    insertSyncMsg.value = e.response?.data?.error ?? 'Could not update approval setting'
  }
}

function openRejectBundle(id) {
  rejectExpandId.value = id
  rejectNote.value = ''
}

function cancelRejectBundle() {
  rejectExpandId.value = null
  rejectNote.value = ''
}

async function confirmRejectBundle(bundleId) {
  if (!farmContext.farmId || !rejectNote.value.trim()) return
  insertBundleBusy.value = bundleId
  insertSyncMsg.value = ''
  try {
    await farmStore.rejectInsertCommonsBundle(farmContext.farmId, bundleId, { note: rejectNote.value.trim() })
    insertSyncMsg.value = 'Bundle rejected.'
    cancelRejectBundle()
    await Promise.all([farmStore.loadAll(farmContext.farmId), loadInsertBundles(), loadInsertHistory()])
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Reject failed'
  } finally {
    insertBundleBusy.value = null
  }
}

async function approveInsertBundle(bundleId) {
  if (!farmContext.farmId) return
  insertBundleBusy.value = bundleId
  insertSyncMsg.value = ''
  try {
    await farmStore.approveInsertCommonsBundle(farmContext.farmId, bundleId, {})
    insertSyncMsg.value = 'Bundle approved; delivery attempted.'
    await Promise.all([farmStore.loadAll(farmContext.farmId), loadInsertBundles(), loadInsertHistory()])
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Approve failed'
    await Promise.all([loadInsertBundles(), loadInsertHistory()])
  } finally {
    insertBundleBusy.value = null
  }
}

async function retryInsertBundleDeliver(bundleId) {
  if (!farmContext.farmId) return
  insertBundleBusy.value = bundleId
  insertSyncMsg.value = ''
  try {
    await farmStore.retryInsertCommonsBundleDeliver(farmContext.farmId, bundleId)
    insertSyncMsg.value = 'Retry completed.'
    await Promise.all([farmStore.loadAll(farmContext.farmId), loadInsertBundles(), loadInsertHistory()])
  } catch (e) {
    insertSyncMsg.value = e.response?.data?.error ?? 'Retry failed'
    await Promise.all([loadInsertBundles(), loadInsertHistory()])
  } finally {
    insertBundleBusy.value = null
  }
}

async function loadInsertPreview() {
  if (!farmContext.farmId || !isFarmAdmin.value) return
  insertPreviewError.value = null
  insertPreviewData.value = null
  insertPreviewLoading.value = true
  try {
    insertPreviewData.value = await farmStore.previewInsertCommons(farmContext.farmId)
    if (!insertPreviewData.value?.valid) {
      insertPreviewError.value = insertPreviewData.value?.error ?? 'Preview invalid'
      insertPreviewData.value = null
    }
  } catch (e) {
    insertPreviewError.value = e.response?.data?.error ?? 'Preview failed'
  } finally {
    insertPreviewLoading.value = false
  }
}

async function runInsertSync() {
  if (!farmContext.farmId || !insertOptIn.value || !canViewInsertCommons.value) return
  insertSyncing.value = true
  insertSyncMsg.value = ''
  try {
    const r = await farmStore.insertCommonsSync(farmContext.farmId)
    if (r.pending_approval) {
      const bid = r.bundle_id != null ? r.bundle_id : '—'
      insertSyncMsg.value = `Queued for approval (bundle #${bid}). An owner or manager can approve it below.`
    } else {
      const status = r.delivery_status ? String(r.delivery_status) : 'unknown'
      const http = r.http_status != null ? ` (HTTP ${r.http_status})` : ''
      insertSyncMsg.value = `${status}${http} — ${r.privacy_notice || 'Sync recorded.'}`
    }
    await farmStore.loadAll(farmContext.farmId)
    await Promise.all([loadInsertHistory(), loadInsertBundles()])
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
  loadPushState()
})
watch(() => farmContext.farmId, () => {
  loadOrgs()
  loadMembers()
  loadFarmSharing()
})

watch(canViewInsertCommons, (ok) => {
  if (ok && farmContext.farmId && insertOptIn.value) {
    loadInsertHistory()
    loadInsertBundles()
  }
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

// ── Push notifications ───────────────────────────────────────────────────────
const pushEnabled = ref(false)
const pushSaving = ref(false)
const pushError = ref('')
const pushTokens = ref([])

async function loadPushState() {
  try {
    const prefs = await api.get('/profile/notification-preferences')
    pushEnabled.value = prefs.data?.push_enabled ?? false
  } catch { /* no prefs yet */ }
  try {
    const r = await api.get('/profile/push-tokens')
    pushTokens.value = Array.isArray(r.data) ? r.data : []
  } catch { /* ignore */ }
}

async function onPushToggle() {
  pushSaving.value = true
  pushError.value = ''
  try {
    await api.patch('/profile/notification-preferences', { push_enabled: pushEnabled.value })
  } catch (e) {
    pushError.value = e.response?.data?.error || e.message || 'Failed to save'
    pushEnabled.value = !pushEnabled.value
  } finally {
    pushSaving.value = false
  }
}

async function removePushToken(fcmToken) {
  try {
    await api.delete('/profile/push-tokens', { data: { fcm_token: fcmToken } })
    pushTokens.value = pushTokens.value.filter(t => t.fcm_token !== fcmToken)
  } catch { /* ignore */ }
}

function formatPushDate(ts) {
  if (!ts) return ''
  return new Date(ts).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

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
