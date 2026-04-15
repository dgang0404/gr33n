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
        <!-- Tab toggle -->
        <div class="flex rounded-lg bg-zinc-800 p-0.5">
          <button type="button" @click="mode = 'login'"
            :class="mode === 'login' ? 'bg-zinc-700 text-white' : 'text-zinc-500 hover:text-zinc-300'"
            class="flex-1 text-sm font-semibold py-1.5 rounded-md transition-colors">
            Sign in
          </button>
          <button type="button" @click="mode = 'register'"
            :class="mode === 'register' ? 'bg-zinc-700 text-white' : 'text-zinc-500 hover:text-zinc-300'"
            class="flex-1 text-sm font-semibold py-1.5 rounded-md transition-colors">
            Register
          </button>
        </div>

        <div v-if="auth.isDevMode" class="bg-amber-900/30 border border-amber-700/50 rounded-lg px-3 py-2">
          <p class="text-amber-300 text-xs">Dev mode — auth is disabled. Any credentials will work.</p>
        </div>
        <div v-else-if="auth.isAuthTestMode" class="bg-violet-950/40 border border-violet-600/40 rounded-lg px-3 py-2">
          <p class="text-violet-200 text-xs">Auth test mode — use real admin credentials; JWT and Pi routes behave like production.</p>
        </div>

        <!-- Full name (register only) -->
        <div v-if="mode === 'register'" class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs font-medium uppercase tracking-wide">Full Name</label>
          <input
            v-model="form.fullName"
            type="text"
            autocomplete="name"
            placeholder="Jane Farmer"
            class="bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2.5 text-white text-sm
                   placeholder-zinc-600 focus:outline-none focus:border-green-500 transition-colors"
          />
        </div>

        <!-- Email / Username -->
        <div class="flex flex-col gap-1.5">
          <label class="text-zinc-400 text-xs font-medium uppercase tracking-wide">
            {{ mode === 'register' ? 'Email' : 'Username' }}
          </label>
          <input
            v-model="form.username"
            :type="mode === 'register' ? 'email' : 'text'"
            :autocomplete="mode === 'register' ? 'email' : 'username'"
            :placeholder="mode === 'register' ? 'you@example.com' : 'admin'"
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
            :autocomplete="mode === 'register' ? 'new-password' : 'current-password'"
            placeholder="••••••••"
            required
            :minlength="mode === 'register' ? 8 : undefined"
            class="bg-zinc-800 border border-zinc-700 rounded-lg px-4 py-2.5 text-white text-sm
                   placeholder-zinc-600 focus:outline-none focus:border-green-500 transition-colors"
          />
          <p v-if="mode === 'register'" class="text-zinc-600 text-xs">Minimum 8 characters</p>
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
          {{ loading ? (mode === 'register' ? 'Creating account…' : 'Signing in…') : (mode === 'register' ? 'Create account' : 'Sign in') }}
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
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../api'

const router = useRouter()
const route  = useRoute()
const auth   = useAuthStore()

auth.fetchAuthMode()

const mode    = ref(route.name === 'register' ? 'register' : 'login')
const form    = reactive({ username: '', password: '', fullName: '' })
const loading = ref(false)
const error   = ref(null)

const submit = async () => {
  error.value   = null
  loading.value = true
  try {
    if (mode.value === 'register') {
      const res = await api.post('/auth/register', {
        email: form.username,
        password: form.password,
        full_name: form.fullName,
      })
      auth.token = res.data.token
      auth.username = form.username
      localStorage.setItem('gr33n_token', res.data.token)
      localStorage.setItem('gr33n_user', form.username)
      if (res.data.user_id) localStorage.setItem('gr33n_user_id', res.data.user_id)
      router.push({ name: 'dashboard' })
    } else {
      await auth.login(form.username, form.password)
      router.push({ name: 'dashboard' })
    }
  } catch (e) {
    error.value = e.response?.data?.error ?? (mode.value === 'register' ? 'Registration failed' : 'Login failed — check username and password')
  } finally {
    loading.value = false
  }
}
</script>
