<template>
  <div :class="layout === 'full' ? 'grid grid-cols-1 lg:grid-cols-[260px_1fr] gap-4' : 'flex flex-col gap-3 min-h-0'">
    <!-- Sessions: full sidebar or compact picker -->
    <aside
      v-if="layout === 'full'"
      class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3 max-h-[36rem] overflow-y-auto"
      data-test="chat-sessions"
    >
      <div class="flex items-center justify-between">
        <h2 class="text-xs uppercase tracking-widest text-zinc-500">Sessions</h2>
        <div class="flex items-center gap-1">
          <button
            v-if="!selectMode"
            type="button"
            data-test="chat-bulk-select"
            class="text-xs px-2 py-1 rounded bg-zinc-800 text-zinc-200 hover:bg-zinc-700 disabled:opacity-40"
            :disabled="streaming || !sessions.length"
            @click="enterSelectMode"
          >
            Select
          </button>
          <button
            type="button"
            data-test="chat-new-session"
            class="text-xs px-2 py-1 rounded bg-zinc-800 text-zinc-200 hover:bg-zinc-700"
            :disabled="streaming || selectMode"
            @click="newSession"
          >
            New
          </button>
        </div>
      </div>
      <div
        v-if="selectMode"
        data-test="chat-bulk-toolbar"
        class="flex items-center justify-between gap-2 rounded bg-zinc-950 border border-zinc-800 px-2 py-1.5 text-[11px] text-zinc-300"
      >
        <span data-test="chat-bulk-count">
          <strong>{{ selectedIds.length }}</strong> of {{ sessions.length }} selected
        </span>
        <div class="flex items-center gap-1">
          <button type="button" class="px-2 py-0.5 rounded bg-zinc-800 hover:bg-zinc-700" data-test="chat-bulk-select-all" :disabled="bulkSubmitting || !sessions.length" @click="selectAll">Select all</button>
          <button type="button" class="px-2 py-0.5 rounded bg-zinc-800 hover:bg-zinc-700" data-test="chat-bulk-cancel" :disabled="bulkSubmitting" @click="exitSelectMode">Cancel</button>
          <button type="button" class="px-2 py-0.5 rounded bg-red-950/60 border border-red-900 hover:bg-red-900/60 text-red-200 disabled:opacity-40" data-test="chat-bulk-delete" :disabled="bulkSubmitting || !selectedIds.length" @click="openBulkConfirm">Delete {{ selectedIds.length || '' }}</button>
        </div>
      </div>
      <p v-if="!sessions.length" class="text-xs text-zinc-600 italic">No saved sessions yet. Send your first message to start one.</p>
      <ul class="space-y-1">
        <li
          v-for="s in sessions"
          :key="s.session_id"
          class="rounded p-2 text-xs space-y-1 group relative"
          :class="[
            s.session_id === sessionId && !selectMode ? 'bg-green-900/40 border border-green-800 text-green-100' : 'hover:bg-zinc-800 text-zinc-300 border border-transparent',
            isSelected(s.session_id) ? 'ring-1 ring-red-800' : '',
          ]"
        >
          <div
            class="flex items-center justify-between gap-2"
            :class="selectMode ? 'cursor-default' : 'cursor-pointer'"
            @click="selectMode ? toggleSelection(s.session_id) : loadSession(s.session_id)"
          >
            <div class="flex items-center gap-2 min-w-0">
              <input v-if="selectMode" type="checkbox" class="rounded bg-zinc-800 border-zinc-700 shrink-0" data-test="chat-session-checkbox" :checked="isSelected(s.session_id)" :disabled="bulkSubmitting" @click.stop @change="toggleSelection(s.session_id)" />
              <span class="font-medium truncate" :title="sessionLabel(s)">{{ sessionLabel(s) }}</span>
            </div>
            <span class="text-[10px] text-zinc-500 shrink-0">{{ s.turn_count }} turn{{ s.turn_count === 1 ? '' : 's' }}</span>
          </div>
          <div class="text-[10px] text-zinc-500 flex items-center justify-between gap-1">
            <div class="flex items-center gap-2">
              <span v-if="s.any_grounded" class="text-gr33n-500">grounded</span>
              <span>{{ formatTime(s.last_turn_at) }}</span>
              <span v-if="(s.total_prompt_tokens || 0) + (s.total_completion_tokens || 0) > 0" class="text-zinc-600" :title="`prompt ${s.total_prompt_tokens} · completion ${s.total_completion_tokens}`">{{ (s.total_prompt_tokens || 0) + (s.total_completion_tokens || 0) }} tok</span>
            </div>
            <div v-if="!selectMode" class="flex items-center gap-1 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
              <button type="button" class="px-1.5 py-0.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300" :disabled="streaming" data-test="chat-session-rename" title="Rename session" @click.stop="renameSession(s)">✎</button>
              <button type="button" class="px-1.5 py-0.5 rounded bg-zinc-800 hover:bg-red-900/60 text-zinc-300 hover:text-red-200" :disabled="streaming" data-test="chat-session-delete" title="Delete session" @click.stop="deleteSession(s)">✕</button>
            </div>
          </div>
        </li>
      </ul>
    </aside>

    <div
      v-else
      class="flex flex-wrap items-center gap-2 shrink-0"
      data-test="chat-sessions-compact"
    >
      <label class="sr-only" for="guardian-session-select">Session</label>
      <select
        id="guardian-session-select"
        :value="sessionId"
        :disabled="streaming"
        class="flex-1 min-w-0 bg-zinc-950 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-zinc-200"
        data-test="chat-session-select"
        @change="onCompactSessionChange"
      >
        <option value="">New conversation</option>
        <option
          v-for="s in sessions"
          :key="s.session_id"
          :value="s.session_id"
        >
          {{ sessionLabel(s) }} ({{ s.turn_count }} turn{{ s.turn_count === 1 ? '' : 's' }})
        </option>
      </select>
      <button
        type="button"
        data-test="chat-new-session"
        class="text-xs px-2 py-1.5 rounded bg-zinc-800 text-zinc-200 hover:bg-zinc-700 shrink-0"
        :disabled="streaming"
        @click="newSession"
      >
        New
      </button>
    </div>

    <div :class="layout === 'compact' ? 'flex flex-col gap-3 min-h-0 flex-1' : 'space-y-4'">
      <!-- Transcript -->
      <section
        v-if="transcript.length || streaming"
        :class="[
          'bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-4 overflow-y-auto',
          layout === 'compact' ? 'flex-1 min-h-[8rem] max-h-[50vh]' : 'p-5 max-h-[36rem]',
        ]"
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
            <span v-if="t.attachment_ids?.length" class="text-zinc-500 text-[10px] ml-1">
              · {{ t.attachment_ids.length }} photo{{ t.attachment_ids.length === 1 ? '' : 's' }}
            </span>
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
          <p
            v-if="t.field_degraded"
            class="text-[10px] text-amber-300/90 px-3"
            data-test="chat-field-degraded-banner"
          >
            LLM offline — showing authored procedure steps only.
          </p>
          <GuardianProcedureCard v-if="t.procedure" :procedure="t.procedure" class="pl-6" />
          <div v-if="t.proposals?.length" class="pl-6 space-y-2" data-test="chat-turn-proposals">
            <GuardianActionProposal
              v-for="p in t.proposals"
              :key="p.proposal_id"
              :proposal="p"
              :can-operate="canOperate"
              @confirmed="onProposalConfirmed"
              @dismissed="onProposalDismissed"
              @error="onProposalError"
            />
          </div>
        </article>
        <div v-if="streaming" class="text-zinc-100 text-sm space-y-2" data-test="chat-streaming-row">
          <span class="text-[10px] uppercase tracking-widest text-green-500 mr-2">guardian</span>
          <span class="whitespace-pre-wrap">{{ streamingText }}<span class="text-zinc-500 animate-pulse">▍</span></span>
        </div>
      </section>

      <!-- Composer -->
      <div
        :class="[
          'bg-zinc-900 border border-zinc-800 rounded-xl space-y-3 shrink-0',
          layout === 'compact' ? 'p-3' : 'p-5 space-y-4',
        ]"
      >
        <div
          v-if="useFarmContext && farmContext.farmId && capabilities.visionChatEnabled && zoneContextId"
          class="rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 space-y-2"
          data-test="chat-vision-attach"
        >
          <div class="flex items-center justify-between gap-2">
            <p class="text-xs text-zinc-400">Zone photos (vision)</p>
            <label class="text-[10px] text-green-600 hover:text-green-400 cursor-pointer">
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                class="hidden"
                :disabled="photoUploading || streaming"
                @change="onChatPhotoSelected"
              />
              {{ photoUploading ? 'Uploading…' : '+ Upload' }}
            </label>
          </div>
          <p v-if="!zonePhotos.length" class="text-[10px] text-zinc-600">
            Attach reference photos from this zone, or upload one, then ask about leaves or layout.
          </p>
          <div v-else class="flex flex-wrap gap-2">
            <button
              v-for="p in zonePhotos"
              :key="p.id"
              type="button"
              class="relative w-14 h-14 rounded border overflow-hidden transition-colors"
              :class="isAttachmentSelected(p.id) ? 'border-green-600 ring-1 ring-green-700' : 'border-zinc-700 hover:border-zinc-500'"
              :title="p.file_name"
              @click="toggleAttachment(p.id)"
            >
              <img
                v-if="photoThumbUrls[p.id]"
                :src="photoThumbUrls[p.id]"
                :alt="p.file_name || 'Zone photo'"
                class="w-full h-full object-cover"
              />
              <span v-else class="text-[9px] text-zinc-500 p-1">#{{ p.id }}</span>
            </button>
          </div>
          <p v-if="selectedAttachmentIds.length" class="text-[10px] text-zinc-500">
            {{ selectedAttachmentIds.length }} selected (max 3 per message)
          </p>
        </div>
        <p
          v-if="capabilities.visionChatEnabled && useFarmContext"
          class="text-[10px] text-zinc-600 leading-relaxed"
          data-test="chat-vision-disclaimer"
        >
          Image analysis is advisory only — hypotheses, not certified diagnosis. Any change still needs Confirm.
        </p>
        <div v-if="setupStarters.length" class="space-y-1.5" data-test="chat-setup-starters">
          <p class="text-[10px] uppercase tracking-widest text-zinc-500">Try asking</p>
          <GuardianStarterChips :starters="setupStarters" />
        </div>
        <div class="flex flex-col gap-2">
          <label class="text-xs text-zinc-400">Your message</label>
          <textarea
            ref="messageInputRef"
            v-model="message"
            :rows="layout === 'compact' ? 2 : 3"
            placeholder="e.g. What should I check on the morning walkthrough?"
            class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-gr33n-600"
            data-test="chat-message-input"
            @keydown.enter.exact.prevent="send"
          />
        </div>
        <div class="flex flex-wrap items-center gap-3">
          <label class="flex items-center gap-2 text-zinc-300 text-sm">
            <input v-model="useFarmContext" type="checkbox" class="rounded bg-zinc-800 border-zinc-700" data-test="chat-use-farm-context" />
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
        <p v-if="sessionId && layout === 'full'" class="text-[10px] text-zinc-600">
          session_id: <span class="font-mono">{{ sessionId }}</span>
        </p>
      </div>
    </div>

    <!-- Bulk-delete confirmation modal -->
    <div
      v-if="bulkConfirming"
      role="dialog"
      aria-modal="true"
      aria-labelledby="bulk-delete-title"
      data-test="chat-bulk-confirm"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-black/60 px-4"
      @click.self="cancelBulkConfirm"
      @keydown.esc="cancelBulkConfirm"
    >
      <form
        class="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl p-5 space-y-4"
        @submit.prevent="submitBulkDelete"
      >
        <h2 id="bulk-delete-title" class="text-sm font-semibold text-zinc-100">
          Delete {{ selectedIds.length }} session{{ selectedIds.length === 1 ? '' : 's' }}?
        </h2>
        <p class="text-xs text-zinc-400">
          Every turn in the selected conversation{{ selectedIds.length === 1 ? '' : 's' }} will be
          permanently removed. This cannot be undone.
        </p>
        <p
          v-if="bulkError"
          data-test="chat-bulk-error"
          class="text-xs text-red-400 bg-red-950/50 border border-red-900 rounded px-2 py-1"
        >
          {{ bulkError }}
        </p>
        <div class="flex justify-end gap-2 pt-1">
          <button
            type="button"
            data-test="chat-bulk-confirm-cancel"
            class="px-3 py-1.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-sm"
            :disabled="bulkSubmitting"
            @click="cancelBulkConfirm"
          >
            Cancel
          </button>
          <button
            type="submit"
            data-test="chat-bulk-confirm-delete"
            class="px-3 py-1.5 rounded bg-red-950/70 hover:bg-red-900/80 border border-red-900 text-red-100 text-sm disabled:opacity-40"
            :disabled="bulkSubmitting"
          >
            {{ bulkSubmitting ? 'Deleting…' : `Delete ${selectedIds.length}` }}
          </button>
        </div>
      </form>
    </div>

    <!-- Inline rename modal -->
    <div
      v-if="renameTarget"
      role="dialog"
      aria-modal="true"
      aria-labelledby="rename-modal-title"
      data-test="chat-rename-modal"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-black/60 px-4"
      @click.self="closeRename"
      @keydown.esc="closeRename"
    >
      <form
        class="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl p-5 space-y-4"
        @submit.prevent="submitRename"
      >
        <h2 id="rename-modal-title" class="text-sm font-semibold text-zinc-100">
          Rename session
        </h2>
        <p class="text-xs text-zinc-500 truncate" :title="renameTarget.first_user_message || ''">
          {{ renameTarget.first_user_message || 'New conversation' }}
        </p>
        <input
          ref="renameInputRef"
          v-model="renameDraft"
          type="text"
          maxlength="120"
          placeholder="Custom title (leave blank to clear)"
          data-test="chat-rename-input"
          class="w-full bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-gr33n-600"
          @keydown.esc.prevent="closeRename"
        />
        <p class="text-[10px] text-zinc-600">
          Empty title falls back to the first message. Max 120 characters.
        </p>
        <p
          v-if="renameError"
          data-test="chat-rename-error"
          class="text-xs text-red-400 bg-red-950/50 border border-red-900 rounded px-2 py-1"
        >
          {{ renameError }}
        </p>
        <div class="flex justify-end gap-2 pt-1">
          <button
            type="button"
            data-test="chat-rename-cancel"
            class="px-3 py-1.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-sm"
            :disabled="renameSubmitting"
            @click="closeRename"
          >
            Cancel
          </button>
          <button
            type="submit"
            data-test="chat-rename-save"
            class="px-3 py-1.5 rounded bg-green-900/60 hover:bg-green-900/80 border border-green-800 text-green-200 text-sm disabled:opacity-40"
            :disabled="renameSubmitting"
          >
            {{ renameSubmitting ? 'Saving…' : 'Save' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import api from '../api'
import GuardianActionProposal from './GuardianActionProposal.vue'
import GuardianProcedureCard from './GuardianProcedureCard.vue'
import GuardianStarterChips from './GuardianStarterChips.vue'
import { computeFirstRunChecklist, isFirstRunIncomplete } from '../lib/firstRunChecklist.js'
import { buildSetupStarters } from '../lib/guardianStarters.js'
import { useFarmOperate } from '../composables/useFarmOperate'
import { useFarmContextStore } from '../stores/farmContext'
import { useFarmStore } from '../stores/farm'
import { useGuardianChatStore } from '../stores/guardianChat'
import { useGuardianPanelStore } from '../stores/guardianPanel'
import { useGuardianProposalsStore } from '../stores/guardianProposals'
import { useCapabilitiesStore } from '../stores/capabilities'
const props = defineProps({
  /** `full` — sidebar session list (page). `compact` — dropdown (drawer). */
  layout: {
    type: String,
    default: 'full',
    validator: (v) => v === 'full' || v === 'compact',
  },
})

const maxHistoryTurns = 20
const route = useRoute()

const farmContext = useFarmContextStore()
const farmStore = useFarmStore()
const guardianPanel = useGuardianPanelStore()
const guardianChat = useGuardianChatStore()
const { streaming, streamingText, error: errorMessage, transcript } = storeToRefs(guardianChat)
const guardianProposals = useGuardianProposalsStore()
const capabilities = useCapabilitiesStore()
const farmIdRef = computed(() => farmContext.farmId)
const zoneContextId = computed(() => {
  const ref = guardianPanel.contextRef
  if (!ref || ref.type !== 'zone' || !ref.id) return null
  return Number(ref.id)
})
const zonePhotos = ref([])
const photoThumbUrls = ref({})
const photoUploading = ref(false)
const selectedAttachmentIds = ref([])
const { canOperate } = useFarmOperate(farmIdRef)

const firstRunChecklistItems = computed(() => computeFirstRunChecklist({
  zones: farmStore.zones || [],
  devices: farmStore.devices || [],
  farmId: farmContext.farmId,
}))

const setupModeActive = computed(() => {
  if (!capabilities.aiEnabled || !farmContext.farmId) return false
  if (guardianPanel.setupMode) return true
  if (route?.path === '/chat' && route?.query?.setup === '1') return true
  if ((farmStore.zones?.length ?? 0) === 0) return true
  return isFirstRunIncomplete(firstRunChecklistItems.value)
})

const setupStarters = computed(() => {
  if (!setupModeActive.value) return []
  const devices = farmStore.devices || []
  const deviceOffline = devices.length > 0 && devices.some((d) => d.status !== 'online')
  const unreadAlerts = (farmStore.alerts || []).filter((a) => !a.is_read)
  return buildSetupStarters({
    surface: 'setup_mode_chat',
    farmId: farmContext.farmId,
    zoneCount: farmStore.zones?.length ?? 0,
    zones: farmStore.zones || [],
    unreadAlerts,
    deviceOffline,
  })
})

const message = ref('')
const useFarmContext = ref(!!farmContext.farmId)
const messageInputRef = ref(null)

const sessionId = computed({
  get: () => guardianPanel.activeSessionId,
  set: (v) => guardianPanel.setActiveSessionId(v),
})

const sessions = ref([])

const renameTarget = ref(null)
const renameDraft = ref('')
const renameSubmitting = ref(false)
const renameError = ref('')
const renameInputRef = ref(null)

const selectMode = ref(false)
const selectedIds = ref([])
const bulkConfirming = ref(false)
const bulkSubmitting = ref(false)
const bulkError = ref('')

watch(
  () => farmContext.farmId,
  (id) => {
    if (id) useFarmContext.value = true
    void loadZonePhotosForChat()
  },
)

watch(zoneContextId, () => {
  selectedAttachmentIds.value = []
  void loadZonePhotosForChat()
})

watch(
  () => guardianPanel.prefilledMessage,
  (msg) => {
    if (msg) message.value = msg
  },
)

watch(
  () => guardianPanel.open,
  async (isOpen) => {
    if (isOpen && guardianPanel.prefilledMessage) {
      message.value = guardianPanel.prefilledMessage
      await nextTick()
      messageInputRef.value?.focus?.()
    }
  },
)

function isSelected(id) {
  return selectedIds.value.includes(id)
}

function toggleSelection(id) {
  if (isSelected(id)) {
    selectedIds.value = selectedIds.value.filter((x) => x !== id)
  } else {
    selectedIds.value = [...selectedIds.value, id]
  }
}

function enterSelectMode() {
  if (streaming.value) return
  selectMode.value = true
  selectedIds.value = []
}

function exitSelectMode() {
  selectMode.value = false
  selectedIds.value = []
  bulkConfirming.value = false
  bulkError.value = ''
}

function selectAll() {
  selectedIds.value = sessions.value.map((s) => s.session_id)
}

function openBulkConfirm() {
  if (!selectedIds.value.length || bulkSubmitting.value) return
  bulkError.value = ''
  bulkConfirming.value = true
}

function cancelBulkConfirm() {
  if (bulkSubmitting.value) return
  bulkConfirming.value = false
  bulkError.value = ''
}

async function submitBulkDelete() {
  if (!selectedIds.value.length || bulkSubmitting.value) return
  bulkSubmitting.value = true
  bulkError.value = ''
  const ids = [...selectedIds.value]
  try {
    const results = await Promise.allSettled(
      ids.map((id) => api.delete('/v1/chat/sessions/' + id)),
    )
    const succeeded = []
    const failed = []
    results.forEach((r, i) => {
      if (r.status === 'fulfilled') succeeded.push(ids[i])
      else failed.push(ids[i])
    })
    if (succeeded.length) {
      sessions.value = sessions.value.filter((s) => !succeeded.includes(s.session_id))
      if (succeeded.includes(sessionId.value)) {
        sessionId.value = ''
        guardianChat.clearTranscript()
      }
    }
    if (failed.length) {
      selectedIds.value = failed
      bulkError.value = `Failed to delete ${failed.length} of ${ids.length} session${ids.length === 1 ? '' : 's'}.`
      return
    }
    exitSelectMode()
  } catch (e) {
    bulkError.value = (e && (e.response?.data?.error || e.message)) || 'bulk delete failed'
  } finally {
    bulkSubmitting.value = false
  }
}

async function refreshSessions() {
  try {
    const r = await api.get('/v1/chat/sessions')
    sessions.value = Array.isArray(r.data?.sessions) ? r.data.sessions : []
  } catch {
    sessions.value = []
  }
}

async function loadSession(id) {
  if (streaming.value || !id) return
  try {
    const r = await api.get('/v1/chat/sessions/' + id)
    sessionId.value = id
    guardianChat.setTranscript(r.data?.turns)
    guardianChat.clearError()
  } catch (e) {
    errorMessage.value = e.message || 'failed to load session'
  }
}

function onCompactSessionChange(ev) {
  const id = ev.target.value
  if (!id) {
    newSession()
    return
  }
  loadSession(id)
}

function newSession() {
  if (streaming.value) return
  sessionId.value = ''
  guardianChat.clearTranscript()
}

function sessionLabel(s) {
  if (s.title && s.title.trim()) return s.title
  if (s.first_user_message && s.first_user_message.trim()) return s.first_user_message
  return 'Untitled'
}

function renameSession(s) {
  if (streaming.value) return
  renameTarget.value = s
  renameDraft.value = s.title || ''
  renameError.value = ''
  renameSubmitting.value = false
  nextTick(() => {
    const el = renameInputRef.value
    if (el && typeof el.focus === 'function') {
      el.focus()
      if (typeof el.select === 'function') el.select()
    }
  })
}

function closeRename() {
  if (renameSubmitting.value) return
  renameTarget.value = null
  renameDraft.value = ''
  renameError.value = ''
}

async function submitRename() {
  if (!renameTarget.value || renameSubmitting.value) return
  const target = renameTarget.value
  const payload = { title: renameDraft.value }
  renameSubmitting.value = true
  renameError.value = ''
  try {
    const r = await api.patch('/v1/chat/sessions/' + target.session_id, payload)
    const i = sessions.value.findIndex((x) => x.session_id === target.session_id)
    if (i !== -1) {
      sessions.value[i] = { ...sessions.value[i], title: r.data?.title ?? null }
    }
    renameTarget.value = null
    renameDraft.value = ''
  } catch (e) {
    renameError.value = (e && (e.response?.data?.error || e.message)) || 'rename failed'
  } finally {
    renameSubmitting.value = false
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
      guardianChat.clearTranscript()
    }
  } catch (e) {
    errorMessage.value = e.message || 'delete failed'
  }
}

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  if (Number.isNaN(d.getTime())) return ts
  return d.toLocaleString()
}

function normalizeProposals(raw) {
  if (!Array.isArray(raw)) return []
  return raw.map((p) => ({
    proposal_id: p.proposal_id,
    tool: p.tool,
    args: p.args || {},
    summary: p.summary || '',
    risk_tier: p.risk_tier || 'medium',
    expires_at: p.expires_at,
    status: 'pending',
    confirmSummary: '',
    error: '',
    result: null,
  }))
}

function patchProposal(proposalId, patch) {
  guardianChat.patchProposal(proposalId, patch)
}

async function refreshAfterAlertAction(proposal, result) {
  const fid = farmContext.farmId
  if (!fid) return
  const alertId = proposal.args?.alert_id
  try {
    if (proposal.tool === 'ack_alert' && alertId) {
      await farmStore.markAlertAcknowledged(alertId)
    } else if (proposal.tool === 'mark_alert_read' && alertId) {
      await farmStore.markAlertRead(alertId)
    } else if (proposal.tool === 'create_task_from_alert' && (result?.task_id ?? result?.id)) {
      await farmStore.loadTasks(fid)
    }
    await farmStore.countUnreadAlerts(fid)
    if (farmStore.alerts.length) {
      await farmStore.loadAlerts(fid, { limit: 50, offset: 0 })
    }
  } catch {
    /* best-effort — confirm already succeeded server-side */
  }
}

function onProposalConfirmed({ proposal, summary, result }) {
  patchProposal(proposal.proposal_id, {
    status: 'confirmed',
    confirmSummary: summary,
    result,
    error: '',
  })
  guardianProposals.removeProposal(proposal.proposal_id)
  if (farmContext.farmId) guardianProposals.refreshPendingCount(farmContext.farmId)
  refreshAfterAlertAction(proposal, result)
}

function onProposalDismissed({ proposal }) {
  patchProposal(proposal.proposal_id, { status: 'dismissed', error: '' })
  guardianProposals.removeProposal(proposal.proposal_id)
  if (farmContext.farmId) guardianProposals.refreshPendingCount(farmContext.farmId)
}

function onProposalError({ proposal, error }) {
  patchProposal(proposal.proposal_id, { error: error || 'Confirm failed' })
}

function isAttachmentSelected(id) {
  return selectedAttachmentIds.value.includes(id)
}

function toggleAttachment(id) {
  const i = selectedAttachmentIds.value.indexOf(id)
  if (i >= 0) {
    selectedAttachmentIds.value = selectedAttachmentIds.value.filter((x) => x !== id)
    return
  }
  if (selectedAttachmentIds.value.length >= 3) return
  selectedAttachmentIds.value = [...selectedAttachmentIds.value, id]
}

function revokeChatPhotoThumbs() {
  for (const url of Object.values(photoThumbUrls.value)) {
    if (url) URL.revokeObjectURL(url)
  }
  photoThumbUrls.value = {}
}

async function loadZonePhotosForChat() {
  revokeChatPhotoThumbs()
  zonePhotos.value = []
  const zid = zoneContextId.value
  if (!zid || !capabilities.visionChatEnabled) return
  try {
    const r = await api.get(`/zones/${zid}/photos`)
    zonePhotos.value = r.data?.photos ?? []
    const thumbs = {}
    await Promise.all(zonePhotos.value.map(async (p) => {
      try {
        const img = await api.get(`/file-attachments/${p.id}/content`, { responseType: 'blob' })
        thumbs[p.id] = URL.createObjectURL(img.data)
      } catch { /* optional thumb */ }
    }))
    photoThumbUrls.value = thumbs
  } catch {
    zonePhotos.value = []
  }
}

async function onChatPhotoSelected(ev) {
  const file = ev.target?.files?.[0]
  ev.target.value = ''
  const zid = zoneContextId.value
  if (!file || !zid || photoUploading.value) return
  photoUploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', file)
    const r = await api.post(`/zones/${zid}/photos`, fd)
    const att = r.data?.file_attachment
    if (att?.id) {
      await loadZonePhotosForChat()
      if (selectedAttachmentIds.value.length < 3) {
        selectedAttachmentIds.value = [...selectedAttachmentIds.value, att.id]
      }
    }
  } catch (e) {
    errorMessage.value = e.response?.data?.error || e.message || 'Photo upload failed'
  } finally {
    photoUploading.value = false
  }
}

async function send() {
  if (!message.value.trim()) return
  if (useFarmContext.value && !farmContext.farmId) return
  guardianChat.clearError()

  const attachedIds = [...selectedAttachmentIds.value]
  const farmId = useFarmContext.value && farmContext.farmId ? Number(farmContext.farmId) : null
  const result = await guardianChat.sendMessage({
    message: message.value,
    farmId,
    sessionId: sessionId.value || undefined,
    contextRef: guardianPanel.chatContextRef(),
    navHistory: guardianPanel.navHistory,
    attachmentIds: attachedIds,
    setupMode: setupModeActive.value,
  })
  if (!result?.finalEvent) return

  const { finalEvent, userMessage, attachedIds: sentIds, body } = result
  sessionId.value = finalEvent.session_id || sessionId.value
  guardianChat.appendTurn({
    turn_index: finalEvent.turn_index,
    user_message: userMessage,
    assistant_message: finalEvent.answer || streamingText.value,
    llm_model: finalEvent.llm_model || '',
    grounded: !!finalEvent.grounded,
    context_count: finalEvent.context_count || 0,
    citations: Array.isArray(finalEvent.citations) ? finalEvent.citations : [],
    proposals: normalizeProposals(finalEvent.proposals),
    procedure: finalEvent.procedure ?? null,
    field_degraded: !!finalEvent.field_degraded,
    farm_id: body.farm_id ?? null,
    attachment_ids: sentIds,
    vision_used: !!finalEvent.vision_used,
  })
  message.value = ''
  selectedAttachmentIds.value = []
  guardianPanel.clearPrefill()
  if (finalEvent.proposals?.length && farmContext.farmId) {
    await guardianProposals.refreshPendingCount(farmContext.farmId)
  }
  await refreshSessions()
}

onUnmounted(revokeChatPhotoThumbs)

onMounted(async () => {
  await refreshSessions()
  if (sessionId.value) await loadSession(sessionId.value)
  if (guardianPanel.prefilledMessage) message.value = guardianPanel.prefilledMessage
})

defineExpose({
  refreshSessions,
  maxHistoryTurns,
})
</script>
