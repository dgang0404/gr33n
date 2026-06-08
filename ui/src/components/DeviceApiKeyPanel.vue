<template>
  <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3" data-test="device-api-key-panel">
    <div class="flex items-start justify-between gap-2">
      <div>
        <h3 class="text-sm font-semibold text-white">Device API key</h3>
        <p class="text-xs text-zinc-500 mt-0.5">
          Per-Pi credential — revoke one device without rotating the whole farm.
        </p>
      </div>
      <span
        v-if="meta.uses_legacy_auth"
        class="text-[10px] uppercase tracking-wide text-amber-400 shrink-0"
        data-test="device-legacy-auth-badge"
      >
        Legacy shared key
      </span>
    </div>

    <p v-if="loading" class="text-xs text-zinc-500">Loading keys…</p>
    <p v-else-if="loadError" class="text-xs text-red-400">{{ loadError }}</p>

    <ul v-else-if="keys.length" class="space-y-2 text-xs">
      <li
        v-for="k in keys"
        :key="k.id"
        class="flex items-center justify-between gap-2 bg-zinc-950/60 border border-zinc-800 rounded-lg px-3 py-2"
        :data-test="`device-key-row-${k.id}`"
      >
        <div class="min-w-0">
          <p class="text-zinc-200">
            {{ k.label || `Key #${k.id}` }}
            <span class="text-zinc-500">· {{ k.active ? 'active' : 'revoked' }}</span>
          </p>
          <p class="text-zinc-600 text-[11px]">
            Created {{ formatDate(k.created_at) }}
            <span v-if="k.last_used_at"> · last used {{ formatDate(k.last_used_at) }}</span>
          </p>
        </div>
        <button
          v-if="k.active"
          type="button"
          class="text-xs text-red-400 hover:text-red-300 shrink-0"
          :disabled="revoking === k.id"
          @click="revokeKey(k.id)"
        >
          {{ revoking === k.id ? 'Revoking…' : 'Revoke' }}
        </button>
      </li>
    </ul>

    <p v-else class="text-xs text-zinc-500">No device key yet — issue one for this Pi.</p>

    <div v-if="issuedKey" class="rounded-lg border border-emerald-800/60 bg-emerald-950/30 p-3 space-y-2" data-test="device-key-show-once">
      <p class="text-xs text-emerald-200 font-medium">Copy this key now — it won't be shown again.</p>
      <code class="block text-[11px] text-emerald-100 break-all font-mono">{{ issuedKey }}</code>
      <div class="flex flex-wrap gap-2">
        <button type="button" class="text-xs text-green-400 hover:text-green-300" @click="copyIssued">
          {{ copied ? 'Copied!' : 'Copy key' }}
        </button>
        <button type="button" class="text-xs text-zinc-400 hover:text-zinc-200" @click="issuedKey = ''">
          Dismiss
        </button>
      </div>
      <p class="text-[11px] text-zinc-500">
        On the Pi: set <code class="text-zinc-400">GR33N_DEVICE_API_KEY</code> or save to
        <code class="text-zinc-400">/etc/gr33n/device.key</code>
      </p>
    </div>

    <div class="flex flex-wrap gap-2 pt-1">
      <button
        type="button"
        class="text-xs px-3 py-1.5 rounded-lg bg-green-800 hover:bg-green-700 text-white disabled:opacity-40"
        :disabled="issuing"
        data-test="device-issue-key"
        @click="issueKey"
      >
        {{ issuing ? 'Issuing…' : meta.active_count ? 'Rotate key' : 'Issue key' }}
      </button>
    </div>
    <p v-if="actionError" class="text-xs text-red-400">{{ actionError }}</p>
  </section>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import api from '../api'

const props = defineProps({
  deviceId: { type: Number, required: true },
})

const keys = ref([])
const meta = ref({ active_count: 0, uses_legacy_auth: true })
const loading = ref(false)
const loadError = ref('')
const issuing = ref(false)
const revoking = ref(null)
const issuedKey = ref('')
const copied = ref(false)
const actionError = ref('')

async function refresh() {
  if (!props.deviceId) return
  loading.value = true
  loadError.value = ''
  try {
    const r = await api.get(`/devices/${props.deviceId}/api-keys`)
    keys.value = r.data?.keys || []
    meta.value = {
      active_count: r.data?.active_count ?? 0,
      uses_legacy_auth: Boolean(r.data?.uses_legacy_auth),
    }
  } catch (e) {
    loadError.value = e.response?.data?.error || e.message || 'Failed to load keys'
  } finally {
    loading.value = false
  }
}

async function issueKey() {
  actionError.value = ''
  issuing.value = true
  try {
    const r = await api.post(`/devices/${props.deviceId}/api-keys`, {
      label: meta.value.active_count ? 'rotated key' : 'Pi edge key',
    })
    issuedKey.value = r.data?.api_key || ''
    await refresh()
  } catch (e) {
    actionError.value = e.response?.data?.error || e.message || 'Could not issue key'
  } finally {
    issuing.value = false
  }
}

async function revokeKey(keyId) {
  actionError.value = ''
  revoking.value = keyId
  try {
    await api.post(`/devices/${props.deviceId}/api-keys/${keyId}/revoke`)
    await refresh()
  } catch (e) {
    actionError.value = e.response?.data?.error || e.message || 'Could not revoke key'
  } finally {
    revoking.value = null
  }
}

function formatDate(ts) {
  if (!ts) return '—'
  return new Date(ts).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

async function copyIssued() {
  if (!issuedKey.value) return
  try {
    await navigator.clipboard.writeText(issuedKey.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    actionError.value = 'Copy failed — select the key text manually'
  }
}

onMounted(refresh)
watch(() => props.deviceId, refresh)
</script>
