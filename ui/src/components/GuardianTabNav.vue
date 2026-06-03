<template>
  <nav
    class="flex shrink-0 border-b border-zinc-800 gap-1"
    :class="compact ? 'px-0' : 'px-1'"
    aria-label="Guardian panels"
    data-test="guardian-tab-nav"
  >
    <button
      type="button"
      class="px-3 py-2 text-xs font-medium border-b-2 -mb-px transition-colors"
      :class="modelValue === 'chat'
        ? 'border-green-500 text-green-300'
        : 'border-transparent text-zinc-500 hover:text-zinc-300'"
      data-test="guardian-tab-chat"
      @click="$emit('update:modelValue', 'chat')"
    >
      Chat
    </button>
    <button
      type="button"
      class="px-3 py-2 text-xs font-medium border-b-2 -mb-px transition-colors flex items-center gap-1.5"
      :class="modelValue === 'pending'
        ? 'border-green-500 text-green-300'
        : 'border-transparent text-zinc-500 hover:text-zinc-300'"
      data-test="guardian-tab-pending"
      @click="$emit('update:modelValue', 'pending')"
    >
      Pending
      <span
        v-if="pendingCount > 0"
        class="min-w-[1.125rem] h-[1.125rem] px-1 rounded-full bg-amber-600 text-[10px] font-bold text-amber-950 flex items-center justify-center"
        data-test="guardian-pending-badge"
      >
        {{ pendingCount > 9 ? '9+' : pendingCount }}
      </span>
    </button>
  </nav>
</template>

<script setup>
defineProps({
  modelValue: { type: String, required: true },
  pendingCount: { type: Number, default: 0 },
  compact: { type: Boolean, default: false },
})

defineEmits(['update:modelValue'])
</script>
