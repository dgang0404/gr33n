<template>
  <div class="p-4 sm:p-6 max-w-3xl mx-auto space-y-6 pb-24 md:pb-10" data-test="crop-profile-detail">
    <div v-if="loading" class="text-zinc-500 text-sm">Loading profile…</div>
    <div v-else-if="error" class="text-red-400 text-sm">{{ error }}</div>
    <template v-else-if="profile">
      <header class="space-y-2">
        <router-link to="/plants" class="text-xs text-zinc-500 hover:text-zinc-300">← Plants</router-link>
        <h1 class="text-2xl font-bold text-green-400">{{ profile.display_name }}</h1>
        <p class="text-sm text-zinc-500">
          <span v-if="profile.is_builtin" class="text-amber-600/90">Built-in</span>
          <span v-else>Farm copy</span>
          <span v-if="profile.category"> · {{ profile.category }}</span>
        </p>
        <p v-if="profile.source" class="text-xs text-zinc-600">{{ profile.source }}</p>
      </header>

      <div class="flex flex-wrap gap-2">
        <button
          v-if="profile.is_builtin && farmId"
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg bg-zinc-800 border border-zinc-700 text-zinc-200 hover:border-green-800"
          data-test="crop-profile-clone"
          @click="cloneProfile"
        >
          Clone to edit
        </button>
        <button
          type="button"
          class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-green-300"
          data-test="crop-profile-export"
          @click="exportProfile"
        >
          Export JSON
        </button>
      </div>

      <div class="overflow-x-auto rounded-xl border border-zinc-800">
        <table class="w-full text-xs text-left">
          <thead class="bg-zinc-900 text-zinc-500 uppercase tracking-wider">
            <tr>
              <th class="px-3 py-2">Stage</th>
              <th class="px-3 py-2">EC (mS/cm)</th>
              <th class="px-3 py-2">pH</th>
              <th class="px-3 py-2">VPD (kPa)</th>
              <th class="px-3 py-2">Photo (h)</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="st in profile.stages"
              :key="st.id"
              class="border-t border-zinc-800 text-zinc-300"
            >
              <td class="px-3 py-2 capitalize">{{ st.stage.replace(/_/g, ' ') }}</td>
              <td class="px-3 py-2 font-mono">{{ ecLine(st) }}</td>
              <td class="px-3 py-2 font-mono">{{ range(st.ph_min, st.ph_max) }}</td>
              <td class="px-3 py-2 font-mono">{{ range(st.vpd_min_kpa, st.vpd_max_kpa) }}</td>
              <td class="px-3 py-2 font-mono">{{ num(st.photoperiod_hrs) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useFarmStore } from '../stores/farm.js'

const route = useRoute()
const router = useRouter()
const store = useFarmStore()
const profile = ref(null)
const loading = ref(true)
const error = ref('')
const farmId = computed(() => store.farmId)

function num(v) {
  if (v == null || v === '') return '—'
  const n = Number(v)
  return Number.isFinite(n) ? String(n) : '—'
}

function range(min, max) {
  const a = num(min)
  const b = num(max)
  if (a === '—' && b === '—') return '—'
  if (a !== '—' && b !== '—') return `${a}–${b}`
  return a !== '—' ? a : b
}

function ecLine(st) {
  const t = num(st.ec_target)
  const r = range(st.ec_min, st.ec_max)
  if (t !== '—' && r !== '—') return `${r} (t ${t})`
  return r !== '—' ? r : t
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    profile.value = await store.getCropProfile(Number(route.params.id))
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Failed to load profile'
  } finally {
    loading.value = false
  }
}

function exportProfile() {
  if (!profile.value) return
  const blob = new Blob([JSON.stringify(profile.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${profile.value.crop_key || 'crop-profile'}.json`
  a.click()
  URL.revokeObjectURL(url)
}

async function cloneProfile() {
  if (!farmId.value || !profile.value?.is_builtin) return
  try {
    const cloned = await store.cloneCropProfile(profile.value.id, farmId.value)
    router.push(`/crop-profiles/${cloned.id}`)
  } catch (e) {
    error.value = e?.response?.data?.error || e?.message || 'Clone failed'
  }
}

onMounted(load)
</script>
