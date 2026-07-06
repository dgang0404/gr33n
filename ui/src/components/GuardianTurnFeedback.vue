<template>
  <div
    v-if="!streaming"
    class="flex flex-wrap items-center gap-2 mt-2"
    data-test="chat-turn-feedback"
  >
    <button
      type="button"
      class="text-xs px-2 py-1 rounded border transition-colors"
      :class="rating === 'up' ? 'border-green-700 bg-green-950/50 text-green-300' : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
      data-test="chat-feedback-up"
      :disabled="saving"
      title="Helpful"
      @click="submit('up')"
    >
      👍
    </button>
    <button
      type="button"
      class="text-xs px-2 py-1 rounded border transition-colors"
      :class="rating === 'down' ? 'border-red-800 bg-red-950/40 text-red-200' : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
      data-test="chat-feedback-down"
      :disabled="saving"
      title="Not helpful"
      @click="openDown"
    >
      👎
    </button>
    <span v-if="saved" class="text-[10px] text-zinc-500">Thanks — saved for review</span>
    <span v-if="error" class="text-[10px] text-red-300/90">{{ error }}</span>

    <div
      v-if="showDownForm"
      class="w-full mt-1 rounded-lg border border-zinc-700 bg-zinc-950/80 p-2 space-y-2"
      data-test="chat-feedback-down-form"
    >
      <p class="text-[10px] text-zinc-500">What was wrong?</p>
      <div class="flex flex-wrap gap-1">
        <button
          v-for="chip in reasonChips"
          :key="chip"
          type="button"
          class="text-[10px] px-2 py-0.5 rounded-full border border-zinc-700 text-zinc-300 hover:border-amber-800 hover:bg-amber-950/40"
          @click="reason = chip"
        >
          {{ chip }}
        </button>
      </div>
      <textarea
        v-model="reason"
        rows="2"
        maxlength="500"
        placeholder="Optional details…"
        class="w-full text-xs bg-zinc-900 border border-zinc-700 rounded px-2 py-1.5 text-zinc-200"
        data-test="chat-feedback-reason"
      />
      <div class="flex gap-2">
        <button
          type="button"
          class="text-xs px-2 py-1 rounded bg-red-950/50 border border-red-900 text-red-200 hover:bg-red-900/40 disabled:opacity-40"
          data-test="chat-feedback-down-submit"
          :disabled="saving"
          @click="submit('down')"
        >
          {{ saving ? 'Saving…' : 'Submit feedback' }}
        </button>
        <button type="button" class="text-xs text-zinc-500 hover:text-zinc-300" @click="cancelDown">Cancel</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import api from '../api'

const props = defineProps({
  sessionId: { type: String, default: '' },
  turnIndex: { type: Number, required: true },
  initialRating: { type: String, default: '' },
  initialReason: { type: String, default: '' },
  streaming: { type: Boolean, default: false },
})

const emit = defineEmits(['updated'])

const reasonChips = ['Invented data', 'Missed alert', 'Too slow', 'Other']

const rating = ref(props.initialRating || '')
const reason = ref(props.initialReason || '')
const showDownForm = ref(false)
const saving = ref(false)
const saved = ref(false)
const error = ref('')

watch(
  () => [props.initialRating, props.initialReason],
  () => {
    rating.value = props.initialRating || ''
    reason.value = props.initialReason || ''
  },
)

function openDown() {
  showDownForm.value = true
  saved.value = false
  error.value = ''
}

function cancelDown() {
  showDownForm.value = false
}

async function submit(nextRating) {
  if (!props.sessionId || props.turnIndex == null) return
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    const body = { rating: nextRating }
    if (nextRating === 'down' && reason.value.trim()) {
      body.reason = reason.value.trim()
    }
    const { data } = await api.patch(
      `/v1/chat/sessions/${props.sessionId}/turns/${props.turnIndex}/feedback`,
      body,
    )
    rating.value = data.feedback_rating || nextRating
    reason.value = data.feedback_reason || reason.value
    saved.value = true
    showDownForm.value = false
    emit('updated', {
      feedback_rating: data.feedback_rating,
      feedback_reason: data.feedback_reason,
      feedback_at: data.feedback_at,
    })
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Could not save feedback'
  } finally {
    saving.value = false
  }
}
</script>
