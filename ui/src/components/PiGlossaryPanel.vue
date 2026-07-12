<template>
  <div
    class="fixed bottom-24 right-4 w-80 bg-zinc-900 border border-zinc-800 rounded-xl shadow-2xl z-20"
    :class="{ hidden: !open }"
  >
    <!-- Header -->
    <div class="bg-zinc-800 border-b border-zinc-700 px-4 py-3 flex items-center justify-between rounded-t-lg">
      <h3 class="text-sm font-semibold text-white">📖 Glossary</h3>
      <button
        type="button"
        @click="open = false"
        class="text-zinc-400 hover:text-zinc-200"
      >
        ✕
      </button>
    </div>

    <!-- Search -->
    <div class="px-4 py-3 border-b border-zinc-800">
      <input
        v-model="searchTerm"
        type="text"
        placeholder="Search terms..."
        class="w-full text-xs rounded-lg bg-zinc-950 border border-zinc-700 px-2 py-1.5 text-zinc-300 placeholder-zinc-600 focus:outline-none focus:border-green-600"
      />
    </div>

    <!-- Terms list -->
    <div class="max-h-64 overflow-y-auto px-4 py-3 space-y-3">
      <div
        v-for="term in filteredTerms"
        :key="term.id"
        class="bg-zinc-950 border border-zinc-800 rounded-lg p-2"
      >
        <div class="text-xs font-semibold text-green-400">{{ term.label }}</div>
        <p class="text-xs text-zinc-400 mt-1">{{ term.definition }}</p>
        <a
          v-if="term.link"
          :href="term.link"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs text-blue-400 hover:text-blue-300 mt-1 inline-block"
        >
          Learn more →
        </a>
      </div>

      <div v-if="filteredTerms.length === 0" class="text-xs text-zinc-500 text-center py-4">
        No results
      </div>
    </div>
  </div>

  <!-- Toggle button -->
  <button
    type="button"
    @click="open = !open"
    class="fixed bottom-28 right-4 z-20 w-12 h-12 rounded-full bg-green-700 text-white hover:bg-green-600 flex items-center justify-center text-lg shadow-lg"
    :class="{ 'bg-green-600': open }"
  >
    ?
  </button>
</template>

<script setup>
import { ref, computed } from 'vue'

const open = ref(false)
const searchTerm = ref('')

const terms = [
  {
    id: 'i2c',
    label: 'I²C',
    definition: 'Inter-Integrated Circuit. A protocol for communicating between components on the same circuit board.',
    link: '/docs/pi-integration-guide.md',
  },
  {
    id: 'gpio',
    label: 'GPIO',
    definition: 'General Purpose Input/Output. Physical pins on the Raspberry Pi for controlling devices.',
    link: null,
  },
  {
    id: 'hat',
    label: 'HAT',
    definition: 'Hardware Attached on Top. A add-on board that stacks on the Raspberry Pi.',
    link: null,
  },
  {
    id: 'dip',
    label: 'DIP Switch',
    definition: 'Dual In-line Package switch. Tiny switches on the relay card for setting the device address.',
    link: '/docs/pi-sequent-hat-setup.md',
  },
  {
    id: 'relay',
    label: 'Relay',
    definition: 'An electromagnetic switch that can turn pumps, lights, and fans on/off.',
    link: null,
  },
  {
    id: 'stack',
    label: 'Stack (of relay cards)',
    definition: 'Multiple relay HATs can be stacked on one Pi. Each card must have unique DIP settings.',
    link: null,
  },
  {
    id: 'actuator',
    label: 'Actuator',
    definition: 'A device like a pump, fan, or light that the relays control.',
    link: null,
  },
  {
    id: 'channel',
    label: 'Channel',
    definition: 'One relay output on a card. Card 0 has channels 0–7, card 1 has 8–15, etc.',
    link: null,
  },
  {
    id: 'api-key',
    label: 'API Key',
    definition: 'A secret token your Pi uses to authenticate with the gr33n API. Keep it safe!',
    link: '/docs/SECURITY.md',
  },
  {
    id: 'uid',
    label: 'UID (Unique ID)',
    definition: 'A permanent identifier for this Pi device. Usually the hostname or MAC address.',
    link: null,
  },
  {
    id: 'mqtt',
    label: 'MQTT',
    definition: 'Message Queuing Telemetry Transport. A lightweight protocol for sensor data.',
    link: '/docs/mqtt-edge-operator-playbook.md',
  },
  {
    id: 'offline-queue',
    label: 'Offline Queue',
    definition: 'Local database on Pi that buffers commands if the API goes down.',
    link: null,
  },
  {
    id: 'sync',
    label: 'Platform sync',
    definition: 'The minimal bootstrap config model where Pi gets full config from API on startup.',
    link: '/docs/phase-13-operator-documentation.md',
  },
  {
    id: 'wiring',
    label: 'Wiring (Relay Contacts)',
    definition: 'The relay has NO (normally open), NC (normally closed), and COM (common) terminals for switches.',
    link: '/docs/pi-sequent-hat-setup.md',
  },
  {
    id: 'dip-address',
    label: 'DIP Address',
    definition: 'Each relay card needs DIP 0–2 set to a unique address (0–7) to avoid conflicts on I²C.',
    link: '/docs/pi-sequent-hat-setup.md',
  },
]

const filteredTerms = computed(() => {
  if (!searchTerm.value) return terms
  const q = searchTerm.value.toLowerCase()
  return terms.filter(
    t => t.label.toLowerCase().includes(q) || t.definition.toLowerCase().includes(q)
  )
})
</script>

<style scoped></style>
