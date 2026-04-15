<template>
  <div class="min-h-screen bg-zinc-950 flex items-center justify-center px-4">
    <div class="w-full max-w-sm">

      <!-- Logo / title -->
      <div class="text-center mb-8">
        <span class="text-4xl font-black text-green-400 tracking-tight">gr33n</span>
        <p class="text-zinc-500 text-sm mt-1">Farm Automation Dashboard</p>
      </div>

      <!-- Card -->
      <form
        @submit.prevent="submit"
        class="bg-zinc-900 border border-zinc-700 rounded-2xl p-8 flex flex-col gap-5"
      >
        <h2 class="text-white text-lg font-semibold">Sign in</h2>

        <div v-if="auth.isDevMode" class="bg-amber-900/30 border border-amber-700/50 rounded-lg px-3 py-2">
          <p class="text-amber-300 text-xs">Dev mode — auth is disabled. Any credentials will work.</p>
        </div>

        <!-- Username -->
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs font-medium uppercase tracking-wide">Username</label>
          <input
            v-model="form.username"
            type="text"
            autocomplete="username"
            placeholder="admin"
            required
            class="bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2.5 text-white text-sm
                   placeholder-zinc-600 focus:outline-none focus:border-green-500 transition-colors"
          />
        </div>

        <!-- Password -->
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs font-medium uppercase tracking-wide">Password</label>
          <input
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            placeholder="••••••••"
            required
            class="bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2.5 text-white text-sm
                   placeholder-zinc-600 focus:outline-none focus:border-green-500 transition-colors"
          />
        </div>

        <!-- Error -->
        <p v-if="error" class="text-red-400 text-sm bg-red-950 border border-red-800 rounded-lg px-3 py-2">
          {{ error }}
        </p>

        <!-- Submit -->
        <button
          type="submit"
          :disabled="loading"
          class="bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 disabled:text-zinc-500
                 text-white font-semibold rounded-lg py-2.5 transition-colors text-sm"
        >
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>

      <p class="text-center text-zinc-600 text-xs mt-6">
        gr33n v0.1 · local-only · no cloud
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth   = useAuthStore()

auth.fetchAuthMode()

const form    = reactive({ username: '', password: '' })
const loading = ref(false)
const error   = ref(null)

const submit = async () => {
  error.value   = null
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    router.push({ name: 'dashboard' })
  } catch (e) {
    error.value = e.response?.data?.error ?? 'Login failed — check username and password'
  } finally {
    loading.value = false
  }
}
</script>
