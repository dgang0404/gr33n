<template>
  <div class="p-6 max-w-2xl mx-auto space-y-6" data-test="device-setup-wizard">
    <div>
      <h1 class="text-xl font-semibold text-white">Connect edge device</h1>
      <p class="text-zinc-500 text-sm mt-1">
        Register your Pi (or relay controller), copy config hints, test the connection, and add actuators.
      </p>
    </div>

    <div class="flex gap-2 text-[10px] uppercase tracking-wide text-zinc-500" aria-label="Wizard steps">
      <span :class="step === 'register' ? 'text-green-400' : ''">1 Register</span>
      <span>›</span>
      <span :class="step === 'apikey' ? 'text-green-400' : ''">2 Pi config</span>
      <span>›</span>
      <span :class="step === 'test' ? 'text-green-400' : ''">3 Test</span>
      <span>›</span>
      <span :class="step === 'actuators' ? 'text-green-400' : ''">4 Actuators</span>
      <span>›</span>
      <span :class="step === 'done' ? 'text-green-400' : ''">5 Done</span>
    </div>

    <p v-if="loadError" class="text-sm text-red-400">{{ loadError }}</p>

    <!-- Step 1 — Register -->
    <template v-if="step === 'register'">
      <form class="space-y-4" @submit.prevent="registerDevice">
        <label class="block">
          <span class="text-zinc-400 text-xs">Device name</span>
          <input v-model="form.name" type="text" required placeholder="e.g. Veg Room Pi" class="input-field mt-1 w-full" data-test="device-wizard-name" />
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Device UID (unique id for pi_client)</span>
          <div class="flex gap-2 mt-1">
            <input v-model="form.deviceUid" type="text" required class="input-field flex-1 font-mono text-xs" data-test="device-wizard-uid" />
            <button type="button" class="text-xs px-2 py-1 border border-zinc-700 rounded text-zinc-400 hover:text-white" @click="regenerateUid">
              Regenerate
            </button>
          </div>
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Device type</span>
          <select v-model="form.deviceType" class="input-field mt-1 w-full">
            <option v-for="t in deviceTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </label>
        <label class="block">
          <span class="text-zinc-400 text-xs">Grow room (optional)</span>
          <select v-model.number="form.zoneId" class="input-field mt-1 w-full">
            <option :value="null">— Farm-wide / assign later —</option>
            <option v-for="z in store.zones" :key="z.id" :value="z.id">{{ z.name }}</option>
          </select>
        </label>
        <p v-if="submitError" class="text-sm text-red-400">{{ submitError }}</p>
        <div class="flex flex-wrap gap-2">
          <button type="submit" :disabled="saving" class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40" data-test="device-wizard-register">
            {{ saving ? 'Registering…' : 'Register device' }}
          </button>
          <router-link to="/settings" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200">Cancel</router-link>
        </div>
      </form>
    </template>

    <!-- Step 2 — API key & Pi config -->
    <template v-else-if="step === 'apikey'">
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Pi client configuration</h2>
        <p class="text-xs text-zinc-500">
          This install uses a shared <strong class="text-zinc-400">PI_API_KEY</strong> on the API server
          (not per-device). Set the same value in <code class="text-zinc-400">pi_client/config.yaml</code>
          under <code class="text-zinc-400">api.api_key</code>. Your admin can find it in server env / Settings → Pi Client.
        </p>
        <pre class="text-[11px] text-zinc-300 bg-zinc-950 border border-zinc-800 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap" data-test="device-wizard-config-snippet">{{ configSnippet }}</pre>
        <button type="button" class="text-xs text-green-400 hover:text-green-300 underline" @click="copySnippet">
          {{ copied ? 'Copied!' : 'Copy config snippet' }}
        </button>
      </section>

      <section class="bg-zinc-900/60 border border-zinc-800 rounded-xl p-4">
        <h2 class="text-sm font-semibold text-zinc-300 mb-2">Field checklist</h2>
        <ul class="space-y-1.5 text-xs text-zinc-400">
          <li v-for="item in piChecklist" :key="item.id" class="flex gap-2">
            <span class="text-zinc-600 shrink-0">☐</span>
            <span>{{ item.label }}</span>
          </li>
        </ul>
        <p class="text-[11px] text-zinc-600 mt-3">
          Full guide: <router-link to="/operator-guide" class="text-green-600 hover:text-green-400">Operator guide</router-link>
          and <code class="text-zinc-500">docs/pi-integration-guide.md</code> §8.3 in the repo.
        </p>
      </section>

      <div class="flex flex-wrap gap-2">
        <button type="button" class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white" @click="step = 'test'">
          Continue to connection test
        </button>
        <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'register'">Back</button>
      </div>
    </template>

    <!-- Step 3 — Test connection -->
    <template v-else-if="step === 'test'">
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Connection test</h2>
        <p class="text-xs text-zinc-500">
          Start <code class="text-zinc-400">pi_client</code> on the Pi, then check whether the device reports online.
        </p>
        <dl class="grid grid-cols-2 gap-2 text-xs">
          <div><dt class="text-zinc-600">Device</dt><dd class="text-zinc-200">{{ createdDevice?.name }}</dd></div>
          <div><dt class="text-zinc-600">Status</dt>
            <dd :class="deviceOnline ? 'text-green-400' : 'text-amber-300'">{{ statusLabel }}</dd>
          </div>
        </dl>
        <button type="button" :disabled="polling" class="text-xs px-3 py-1.5 border border-zinc-700 rounded text-zinc-300 hover:text-white disabled:opacity-40" data-test="device-wizard-poll" @click="pollDevice">
          {{ polling ? 'Checking…' : 'Check again' }}
        </button>
        <p v-if="pollMessage" class="text-xs text-zinc-500">{{ pollMessage }}</p>
      </section>
      <div class="flex flex-wrap gap-2">
        <button type="button" class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white" @click="step = 'actuators'">
          {{ deviceOnline ? 'Continue' : 'Continue anyway' }}
        </button>
        <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'apikey'">Back</button>
      </div>
    </template>

    <!-- Step 4 — Actuators -->
    <template v-else-if="step === 'actuators'">
      <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-3">
        <h2 class="text-sm font-semibold text-white">Add actuators (optional)</h2>
        <p class="text-xs text-zinc-500">Link outputs to this device and room so automations and lighting programs can target them.</p>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="t in actuatorTemplates"
            :key="t.id"
            type="button"
            class="text-xs px-3 py-1.5 rounded-full border transition-colors"
            :class="selectedActuators.has(t.id)
              ? 'border-green-600 bg-green-950/40 text-green-300'
              : 'border-zinc-700 text-zinc-400 hover:border-zinc-500'"
            @click="toggleActuator(t.id)"
          >
            {{ t.label }}
          </button>
        </div>
        <label v-if="selectedActuators.size" class="block text-xs text-zinc-500">
          GPIO / channel hint (optional)
          <input v-model="hardwareId" type="text" placeholder="BCM17" class="input-field mt-1 w-full font-mono" />
        </label>
        <p v-if="actuatorError" class="text-sm text-red-400">{{ actuatorError }}</p>
      </section>
      <div class="flex flex-wrap gap-2">
        <button type="button" :disabled="savingActuators" class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white disabled:opacity-40" data-test="device-wizard-finish" @click="finishActuators">
          {{ savingActuators ? 'Saving…' : (selectedActuators.size ? 'Create actuators & finish' : 'Skip & finish') }}
        </button>
        <button type="button" class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200" @click="step = 'test'">Back</button>
      </div>
    </template>

    <!-- Step 5 — Done -->
    <template v-else>
      <section class="bg-zinc-900 border border-green-900/50 rounded-xl p-4 space-y-2">
        <p class="text-sm text-green-300 font-medium">{{ doneMessage }}</p>
        <p v-if="actuatorsCreated" class="text-xs text-zinc-500">{{ actuatorsCreated }} actuator(s) linked to this device.</p>
      </section>
      <div class="flex flex-wrap gap-2">
        <router-link v-if="form.zoneId" :to="`/zones/${form.zoneId}`" class="px-4 py-2 text-sm font-medium rounded-lg bg-green-700 hover:bg-green-600 text-white">
          Open room
        </router-link>
        <router-link to="/actuators" class="px-4 py-2 text-sm text-zinc-300 border border-zinc-700 rounded-lg">
          View controls
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '../api.js'
import { useFarmStore } from '../stores/farm.js'
import { useFarmContextStore } from '../stores/farmContext.js'
import { parseZoneIdQuery } from '../lib/zoneContext.js'
import {
  DEVICE_TYPE_OPTIONS,
  DEVICE_ACTUATOR_TEMPLATES,
  PI_FIELD_CHECKLIST,
  suggestDeviceUid,
  buildDeviceCreatePayload,
  buildActuatorCreatePayload,
  buildPiConfigSnippet,
  isDeviceOnline,
  formatDeviceStatusLabel,
} from '../lib/deviceSetupWizard.js'

const route = useRoute()
const store = useFarmStore()
const farmContext = useFarmContextStore()

const step = ref('register')
const saving = ref(false)
const savingActuators = ref(false)
const submitError = ref('')
const actuatorError = ref('')
const loadError = ref('')
const polling = ref(false)
const pollMessage = ref('')
const copied = ref(false)
const doneMessage = ref('')
const actuatorsCreated = ref(0)
const createdDevice = ref(null)
const selectedActuators = ref(new Set())
const hardwareId = ref('BCM17')

const form = reactive({
  name: '',
  deviceUid: '',
  deviceType: 'raspberry_pi_edge',
  zoneId: null,
})

const farmId = computed(() => {
  const raw = route.params.id
  const n = Number(Array.isArray(raw) ? raw[0] : raw)
  return Number.isFinite(n) && n > 0 ? n : null
})

const deviceTypes = DEVICE_TYPE_OPTIONS
const actuatorTemplates = DEVICE_ACTUATOR_TEMPLATES
const piChecklist = PI_FIELD_CHECKLIST

const configSnippet = computed(() => {
  if (!createdDevice.value || !farmId.value) return ''
  return buildPiConfigSnippet({
    baseUrl: typeof window !== 'undefined' ? `${window.location.origin.replace(/:\d+$/, ':8080')}` : 'http://<api-lan-ip>:8080',
    farmId: farmId.value,
    deviceId: createdDevice.value.id,
    deviceUid: createdDevice.value.device_uid || form.deviceUid,
  })
})

const deviceOnline = computed(() => isDeviceOnline(createdDevice.value))
const statusLabel = computed(() => formatDeviceStatusLabel(createdDevice.value))

async function ensureFarmContext() {
  loadError.value = ''
  if (!farmId.value) {
    loadError.value = 'Invalid farm id in URL.'
    return false
  }
  if (!farmContext.farms.length) {
    try {
      await farmContext.fetchFarms()
    } catch (e) {
      loadError.value = e.response?.data?.error || 'Could not load farms'
      return false
    }
  }
  if (!farmContext.farms.some((f) => f.id === farmId.value)) {
    loadError.value = 'Farm not found or you do not have access.'
    return false
  }
  if (farmContext.farmId !== farmId.value) {
    await farmContext.selectFarm(farmId.value)
  }
  const zoneFromQuery = parseZoneIdQuery(route.query.zone_id)
  if (zoneFromQuery != null) {
    form.zoneId = zoneFromQuery
  }
  if (!form.deviceUid) {
    form.deviceUid = suggestDeviceUid(farmId.value)
  }
  return true
}

function regenerateUid() {
  if (farmId.value) form.deviceUid = suggestDeviceUid(farmId.value)
}

async function registerDevice() {
  if (!farmId.value) return
  saving.value = true
  submitError.value = ''
  try {
    const payload = buildDeviceCreatePayload(form)
    const device = await store.createDevice(farmId.value, payload)
    createdDevice.value = device
    step.value = 'apikey'
  } catch (e) {
    submitError.value = e.response?.data?.error || e.message || 'Could not register device'
  } finally {
    saving.value = false
  }
}

async function copySnippet() {
  try {
    await navigator.clipboard.writeText(configSnippet.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    copied.value = false
  }
}

async function pollDevice() {
  if (!createdDevice.value?.id) return
  polling.value = true
  pollMessage.value = ''
  try {
    const r = await api.get(`/devices/${createdDevice.value.id}`)
    createdDevice.value = r.data
    const idx = store.devices.findIndex((d) => d.id === r.data.id)
    if (idx >= 0) store.devices[idx] = r.data
    else store.devices.push(r.data)
    pollMessage.value = deviceOnline.value
      ? 'Device is online — Pi heartbeat received.'
      : 'Still offline — confirm pi_client is running and PI_API_KEY matches the server.'
  } catch (e) {
    pollMessage.value = e.response?.data?.error || e.message || 'Could not refresh device'
  } finally {
    polling.value = false
  }
}

function toggleActuator(id) {
  const next = new Set(selectedActuators.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedActuators.value = next
}

async function finishActuators() {
  if (!farmId.value || !createdDevice.value) return
  savingActuators.value = true
  actuatorError.value = ''
  actuatorsCreated.value = 0
  try {
    for (const t of actuatorTemplates) {
      if (!selectedActuators.value.has(t.id)) continue
      const { body } = buildActuatorCreatePayload({
        farmId: farmId.value,
        deviceId: createdDevice.value.id,
        zoneId: form.zoneId || createdDevice.value.zone_id,
        template: t,
        hardwareId: hardwareId.value,
      })
      await store.createActuator(farmId.value, body)
      actuatorsCreated.value += 1
    }
    doneMessage.value = `“${createdDevice.value.name}” registered${deviceOnline.value ? ' and online' : ''}.`
    step.value = 'done'
  } catch (e) {
    actuatorError.value = e.response?.data?.error || e.message || 'Could not create actuators'
  } finally {
    savingActuators.value = false
  }
}

onMounted(() => {
  void ensureFarmContext()
})

watch(() => route.fullPath, () => {
  void ensureFarmContext()
})
</script>

<style scoped>
.input-field {
  @apply bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:outline-none focus:ring-1 focus:ring-green-600 focus:border-green-600;
}
</style>
