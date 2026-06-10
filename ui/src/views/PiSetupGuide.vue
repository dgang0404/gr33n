<template>
  <div class="p-4 sm:p-6 max-w-3xl mx-auto space-y-10 pb-24 md:pb-10">

    <!-- Header -->
    <header class="space-y-2">
      <div class="flex items-center gap-2">
        <router-link v-nav-hint="'/operator-guide'" to="/operator-guide" class="text-xs text-zinc-500 hover:text-zinc-300">Operator guide</router-link>
        <span class="text-zinc-700">/</span>
        <span class="text-xs text-zinc-400">Pi + HAT setup</span>
      </div>
      <h1 class="text-2xl font-bold text-green-400">Pi + Sequent Microsystems HAT Setup</h1>
      <p class="text-sm text-zinc-400 leading-relaxed max-w-2xl">
        The fastest way to wire up a grow room with off-the-shelf hardware.
        One Raspberry Pi + one 8-relay HAT controls pumps, lights, fans, and valves.
        Stack more cards as the farm grows — up to 64 relays, no rewiring.
      </p>
    </header>

    <!-- Parts list -->
    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Starter parts list</h2>
      <div class="space-y-2">
        <div v-for="part in parts" :key="part.label"
          class="flex items-start gap-3 bg-zinc-950/60 rounded-lg px-3 py-2.5 border border-zinc-800">
          <span class="text-lg leading-none mt-0.5 shrink-0">{{ part.icon }}</span>
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">{{ part.label }}</div>
            <div class="text-xs text-zinc-500 mt-0.5">{{ part.note }}</div>
            <a v-if="part.url" :href="part.url" target="_blank" rel="noopener"
              class="text-xs text-gr33n-500 hover:underline mt-0.5 inline-block">{{ part.urlLabel }} ↗</a>
          </div>
          <div class="ml-auto text-xs text-zinc-600 shrink-0 font-mono">{{ part.qty }}</div>
        </div>
      </div>
      <p class="text-xs text-zinc-600 pt-1">
        The relay card uses only I²C (2 GPIO pins). All other GPIO pins stay free for future HATs.
      </p>
    </section>

    <!-- Stack diagram -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">How stacking works</h2>
      <p class="text-sm text-zinc-400 leading-relaxed">
        Each card slides onto the Pi's 40-pin header. You can stack up to 8 relay cards
        for 64 relays on one Pi. Each card gets a unique address via 3 DIP switch bits.
      </p>
      <div class="bg-zinc-950 border border-zinc-800 rounded-xl p-4 font-mono text-xs leading-6 text-zinc-400 overflow-x-auto">
        <pre>{{ stackDiagram }}</pre>
      </div>
      <p class="text-xs text-zinc-600">
        Cards can be stacked in any order — the DIP address determines which card responds, not the physical position.
      </p>
    </section>

    <!-- DIP switch table -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">DIP switch address table</h2>
      <p class="text-sm text-zinc-400">
        Set the 3-bit DIP switch (<code class="text-zinc-300">ID0 ID1 ID2</code>) on each card before stacking.
        Every card in the stack must have a unique address.
      </p>
      <div class="overflow-x-auto rounded-xl border border-zinc-800">
        <table class="w-full text-xs">
          <thead class="bg-zinc-900">
            <tr>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Stack level</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">I²C address</th>
              <th class="px-3 py-2 text-center text-zinc-400 font-medium">ID0</th>
              <th class="px-3 py-2 text-center text-zinc-400 font-medium">ID1</th>
              <th class="px-3 py-2 text-center text-zinc-400 font-medium">ID2</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">gr33n channels</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in dipTable" :key="row.level"
              :class="row.level === 0 ? 'bg-green-950/30 border-b border-zinc-800' : 'border-b border-zinc-800/50 hover:bg-zinc-900/40'">
              <td class="px-3 py-2 text-zinc-200 font-semibold">{{ row.level }}
                <span v-if="row.level === 0" class="ml-1 text-[10px] text-gr33n-500 font-normal">start here</span>
              </td>
              <td class="px-3 py-2 font-mono text-zinc-300">{{ row.i2c }}</td>
              <td class="px-3 py-2 text-center">
                <DipBit :on="row.id0" />
              </td>
              <td class="px-3 py-2 text-center">
                <DipBit :on="row.id1" />
              </td>
              <td class="px-3 py-2 text-center">
                <DipBit :on="row.id2" />
              </td>
              <td class="px-3 py-2 font-mono text-zinc-500 text-[11px]">{{ row.channels }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Channel map -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Channel numbering in gr33n</h2>
      <p class="text-sm text-zinc-400 leading-relaxed">
        Each relay becomes a numbered <strong class="text-zinc-200">channel</strong> in the gr33n config.
        Stack 0 relay 1 = channel 0, stack 0 relay 2 = channel 1, … stack 1 relay 1 = channel 8, and so on.
      </p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div v-for="card in channelMapCards" :key="card.stack"
          class="bg-zinc-950 border border-zinc-800 rounded-xl p-3 space-y-2">
          <div class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">
            Stack {{ card.stack }} — I²C {{ card.i2c }}
          </div>
          <div class="grid grid-cols-4 gap-1">
            <template v-for="ch in card.channels" :key="ch.relay">
              <!-- channel has a wired actuator — make it a link -->
              <router-link
                v-if="slotActuator(ch.channel)"
                v-nav-hint="'/actuators'"
                to="/actuators"
                :title="slotActuator(ch.channel).name"
                class="block rounded px-1.5 py-1 text-center border bg-green-950/40 border-green-800/50 hover:bg-green-900/40 cursor-pointer transition-colors"
              >
                <div class="text-[10px] text-zinc-600">relay {{ ch.relay }}</div>
                <div class="text-xs font-mono text-gr33n-400 font-semibold">ch{{ ch.channel }}</div>
                <div class="text-[9px] text-green-300/80 truncate leading-tight mt-0.5">{{ slotActuator(ch.channel).name }}</div>
              </router-link>
              <div
                v-else
                class="rounded px-1.5 py-1 text-center border bg-zinc-900 border-zinc-800"
              >
                <div class="text-[10px] text-zinc-600">relay {{ ch.relay }}</div>
                <div class="text-xs font-mono text-gr33n-400 font-semibold">ch{{ ch.channel }}</div>
              </div>
            </template>
          </div>
        </div>
      </div>
      <div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 space-y-2">
        <div class="text-xs font-semibold text-zinc-400">config.yaml snippet (Phase 51 wiring)</div>
        <pre class="text-xs font-mono text-zinc-300 leading-5">{{ yamlExample }}</pre>
      </div>
    </section>

    <!-- ── Live "Your farm" wiring map ──────────────────────────────────── -->
    <section class="space-y-4" data-test="pi-setup-live-wiring">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Your farm channels</h2>
        <span class="text-[10px] text-zinc-600">from platform wiring — click any row to edit</span>
      </div>

      <div v-if="!wiredDevices.length" class="rounded-xl border border-zinc-800 bg-zinc-950/50 px-4 py-5 text-center space-y-2">
        <p class="text-sm text-zinc-500">No wiring set up yet.</p>
        <p class="text-xs text-zinc-600">
          Register a Pi in
          <router-link v-nav-hint="'/settings'" to="/settings" class="text-green-600 hover:text-green-400">Settings → Devices</router-link>,
          then open a sensor or actuator and set its GPIO pin / channel.
        </p>
      </div>

      <div v-else class="space-y-6">
        <div
          v-for="device in wiredDevices"
          :key="device.id"
          class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden"
        >
          <!-- device header -->
          <div class="flex items-center gap-3 px-4 py-3 border-b border-zinc-800 bg-zinc-950/40">
            <span class="text-base">🖥️</span>
            <span class="text-sm font-semibold text-white">{{ deviceName(device) }}</span>
            <span class="text-[10px] font-mono text-zinc-500">{{ device.device_uid }}</span>
            <span
              class="ml-auto text-[10px] px-2 py-0.5 rounded-full font-medium"
              :class="device.status === 'online' ? 'bg-green-900/40 text-green-400' : 'bg-zinc-800 text-zinc-500'"
            >{{ device.status || 'offline' }}</span>
          </div>

          <!-- relay channels -->
          <div v-if="deviceActuators(device.id).length" class="px-4 py-3 border-b border-zinc-800/60">
            <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Relay channels</p>
            <div class="space-y-1">
              <router-link
                v-for="row in deviceActuators(device.id)"
                :key="row.channel"
                v-nav-hint="'/actuators'"
                :to="'/actuators'"
                class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-zinc-800/60 transition-colors group"
                data-test="pi-setup-channel-row"
              >
                <span class="font-mono text-[11px] text-gr33n-400 shrink-0 w-8">ch{{ row.channel }}</span>
                <span class="text-xs text-zinc-200 truncate">{{ row.actuator.name }}</span>
                <span class="text-[10px] text-zinc-500 ml-auto capitalize">{{ row.actuator.actuator_type }}</span>
                <span class="text-[10px] text-zinc-700 group-hover:text-zinc-400 shrink-0">→</span>
              </router-link>
            </div>
          </div>

          <!-- sensor GPIO pins -->
          <div v-if="deviceSensors(device.id).length" class="px-4 py-3">
            <p class="text-[10px] uppercase tracking-wide text-zinc-600 mb-2">Sensor pins</p>
            <div class="space-y-1">
              <router-link
                v-for="row in deviceSensors(device.id)"
                :key="row.sensor.id"
                v-nav-hint="'/sensors'"
                :to="{ name: 'sensor-detail', params: { id: row.sensor.id } }"
                class="flex items-center gap-3 rounded-lg px-3 py-2 hover:bg-zinc-800/60 transition-colors group"
                data-test="pi-setup-sensor-row"
              >
                <span class="font-mono text-[11px] text-blue-400 shrink-0 w-8 truncate">{{ row.label.split('·')[1]?.trim() || '—' }}</span>
                <span class="text-xs text-zinc-200 truncate">{{ row.sensor.name }}</span>
                <span class="text-[10px] text-zinc-500 ml-auto">{{ row.label.split('·')[0]?.trim() }}</span>
                <span class="text-[10px] text-zinc-700 group-hover:text-zinc-400 shrink-0">→ edit</span>
              </router-link>
            </div>
          </div>

          <!-- nothing wired to this device yet -->
          <div
            v-if="!deviceActuators(device.id).length && !deviceSensors(device.id).length"
            class="px-4 py-4 text-xs text-zinc-600 italic"
          >
            No wiring assigned — open
            <router-link v-nav-hint="'/actuators'" to="/actuators" class="text-green-600 hover:text-green-400">Controls</router-link>
            or
            <router-link v-nav-hint="'/sensors'" to="/sensors" class="text-green-600 hover:text-green-400">Sensors</router-link>
            and assign this Pi.
          </div>
        </div>
      </div>
    </section>

    <!-- Typical wiring plan -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Typical 8-channel farm wiring plan</h2>
      <p class="text-sm text-zinc-400">
        A single 8-relay HAT covers a complete small-to-medium grow room.
        Assign channels to outputs before wiring so the gr33n dashboard matches the hardware.
      </p>
      <div class="overflow-x-auto rounded-xl border border-zinc-800">
        <table class="w-full text-xs">
          <thead class="bg-zinc-900">
            <tr>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Channel</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Relay</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Typical use</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Load type</th>
              <th class="px-3 py-2 text-left text-zinc-400 font-medium">Notes</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in wiringPlan" :key="row.ch"
              class="border-b border-zinc-800/50 hover:bg-zinc-900/40">
              <td class="px-3 py-2 font-mono font-semibold text-gr33n-400">{{ row.ch }}</td>
              <td class="px-3 py-2 text-zinc-300">{{ row.relay }}</td>
              <td class="px-3 py-2 text-zinc-200">{{ row.use }}</td>
              <td class="px-3 py-2 text-zinc-500">{{ row.load }}</td>
              <td class="px-3 py-2 text-zinc-600">{{ row.notes }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="rounded-xl border border-amber-900/40 bg-amber-950/20 px-4 py-3 text-xs text-amber-300 space-y-1">
        <div class="font-semibold">⚡ Load exceeds 4A / 120VAC?</div>
        <div>Wire the relay output to a contactor or SSR — the relay switches the coil, not the full load.
        Typical contactors: 25A HVAC contactor for HID lights, 40A SSR for high-wattage LEDs via DIN rail.</div>
      </div>
    </section>

    <!-- Step by step -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Setup steps</h2>
      <ol class="space-y-3">
        <li v-for="step in setupSteps" :key="step.n"
          class="flex gap-4 bg-zinc-900 border border-zinc-800 rounded-xl px-4 py-3">
          <span class="text-gr33n-500 font-bold text-lg leading-none shrink-0 pt-0.5">{{ step.n }}</span>
          <div class="space-y-1">
            <div class="text-sm font-semibold text-zinc-200">{{ step.title }}</div>
            <div class="text-xs text-zinc-500 leading-relaxed">{{ step.body }}</div>
            <div v-if="step.code" class="mt-1.5">
              <pre class="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-xs font-mono text-zinc-300 whitespace-pre-wrap">{{ step.code }}</pre>
            </div>
          </div>
        </li>
      </ol>
    </section>

    <!-- Scaling -->
    <section class="space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Scaling up</h2>
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div v-for="tier in scaleTiers" :key="tier.name"
          class="bg-zinc-900 border rounded-xl p-4 space-y-2"
          :class="tier.highlight ? 'border-gr33n-800' : 'border-zinc-800'">
          <div class="text-xs font-semibold uppercase tracking-wider"
            :class="tier.highlight ? 'text-gr33n-400' : 'text-zinc-500'">{{ tier.name }}</div>
          <div class="text-sm font-semibold text-zinc-200">{{ tier.relays }} relays</div>
          <div class="text-xs text-zinc-500 leading-relaxed">{{ tier.desc }}</div>
        </div>
      </div>
    </section>

    <!-- Sensor inputs -->
    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-3">
      <h2 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Adding sensor inputs</h2>
      <p class="text-sm text-zinc-400 leading-relaxed">
        The relay HAT handles outputs. For inputs (float switches, door sensors, flow meters, VPD)
        add a Sequent input card to the same stack — it shares I²C and uses the same DIP addressing.
      </p>
      <div class="space-y-2">
        <div v-for="input in inputCards" :key="input.name"
          class="flex items-start gap-3 bg-zinc-950/60 rounded-lg px-3 py-2.5 border border-zinc-800">
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">{{ input.name }}</div>
            <div class="text-xs text-zinc-500 mt-0.5">{{ input.desc }}</div>
          </div>
          <div class="text-xs text-zinc-600 shrink-0 text-right font-mono">{{ input.channels }}</div>
        </div>
      </div>
      <p class="text-xs text-zinc-600">
        I²C addresses for input cards share the same 0x20–0x27 range, but each card <em>type</em>
        has its own address space — a relay card at stack 0 and an input card at stack 0 do not conflict.
      </p>
    </section>

    <!-- Next steps -->
    <section class="rounded-xl border border-zinc-800 bg-zinc-950/50 px-4 py-4 space-y-3">
      <div class="text-xs font-semibold uppercase tracking-wider text-zinc-500">After wiring</div>
      <div class="space-y-2 text-sm">
        <div class="flex items-center gap-2 text-zinc-400">
          <span class="text-gr33n-500">1.</span>
          <router-link v-nav-hint="'/actuators'" to="/actuators" class="text-gr33n-400 hover:underline">Add actuators</router-link>
          <span>— one per relay channel, set channel_id to match the table above</span>
        </div>
        <div class="flex items-center gap-2 text-zinc-400">
          <span class="text-gr33n-500">2.</span>
          <router-link v-nav-hint="'/sensors'" to="/sensors" class="text-gr33n-400 hover:underline">Add sensors</router-link>
          <span>— connect to the zone and assign wiring if using input cards</span>
        </div>
        <div class="flex items-center gap-2 text-zinc-400">
          <span class="text-gr33n-500">3.</span>
          <router-link v-nav-hint="'/comfort-targets'" to="/comfort-targets?tab=schedules" class="text-gr33n-400 hover:underline">Set schedules</router-link>
          <span>— light cycle, irrigation windows</span>
        </div>
        <div class="flex items-center gap-2 text-zinc-400">
          <span class="text-gr33n-500">4.</span>
          <router-link v-nav-hint="'/chat'" to="/chat" class="text-gr33n-400 hover:underline">Ask Guardian</router-link>
          <span>— "Walk me through wiring a pump to channel 2"</span>
        </div>
      </div>
      <p class="text-xs text-zinc-600 border-t border-zinc-800 pt-3">
        Full reference: <code class="text-zinc-500">docs/pi-sequent-hat-setup.md</code> ·
        <code class="text-zinc-500">docs/pi-integration-guide.md</code> ·
        <a href="https://sequentmicrosystems.com" target="_blank" rel="noopener" class="text-zinc-500 hover:text-zinc-400">sequentmicrosystems.com ↗</a>
      </p>
    </section>

  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted } from 'vue'
import { useFarmStore } from '../stores/farm'
import { resolveWiring, formatWiringLabel } from '../lib/hardwareWiring.js'

const store = useFarmStore()
onMounted(() => { if (!store.actuators.length && !store.sensors.length) store.loadFarm?.() })

// ── Live wiring helpers ───────────────────────────────────────────────────────

function channelFromActuator(a) {
  const hi = a?.hardware_identifier
  if (hi == null) return null
  const n = parseInt(String(hi), 10)
  return Number.isFinite(n) && n >= 0 ? n : null
}

/** { deviceId: { channelNumber: actuator } } */
const actuatorByChannel = computed(() => {
  const map = {}
  for (const a of store.actuators) {
    if (!a.device_id) continue
    const ch = channelFromActuator(a)
    if (ch == null) continue
    if (!map[a.device_id]) map[a.device_id] = {}
    map[a.device_id][ch] = a
  }
  return map
})

/** { deviceId: [ { sensor, wiring } ] } sorted by pin */
const sensorByDevicePin = computed(() => {
  const map = {}
  for (const s of store.sensors) {
    const w = resolveWiring(s)
    if (!w?.device_id) continue
    if (!map[w.device_id]) map[w.device_id] = []
    map[w.device_id].push({ sensor: s, wiring: w })
  }
  for (const rows of Object.values(map)) {
    rows.sort((a, b) => (a.wiring.gpio_pin ?? 99) - (b.wiring.gpio_pin ?? 99))
  }
  return map
})

/** Devices that have at least one wired actuator or sensor */
const wiredDevices = computed(() => {
  const ids = new Set([
    ...Object.keys(actuatorByChannel.value).map(Number),
    ...Object.keys(sensorByDevicePin.value).map(Number),
  ])
  return store.devices.filter(d => ids.has(d.id))
})

function deviceName(d) {
  return d.name || d.device_uid || `Device ${d.id}`
}

/** Return the first actuator wired to a given channel across all devices (for the reference card overlay). */
function slotActuator(channel) {
  for (const chMap of Object.values(actuatorByChannel.value)) {
    if (chMap[channel]) return chMap[channel]
  }
  return null
}

/** Actuators assigned to a device's channels, sorted by channel */
function deviceActuators(deviceId) {
  const chMap = actuatorByChannel.value[deviceId] || {}
  return Object.entries(chMap)
    .map(([ch, a]) => ({ channel: Number(ch), actuator: a }))
    .sort((a, b) => a.channel - b.channel)
}

/** Sensors wired to a device, with formatted label */
function deviceSensors(deviceId) {
  return (sensorByDevicePin.value[deviceId] || []).map(({ sensor, wiring }) => ({
    sensor,
    label: formatWiringLabel(wiring) || 'wired',
  }))
}

// DipBit — inline sub-component that renders a single ON/OFF DIP switch indicator.
const DipBit = defineComponent({
  props: { on: Boolean },
  setup(props) {
    return () => h('span', {
      class: props.on
        ? 'inline-block w-5 h-5 rounded bg-green-700 border border-green-600 text-[9px] font-bold text-green-100 text-center leading-5'
        : 'inline-block w-5 h-5 rounded bg-zinc-800 border border-zinc-700 text-[9px] font-bold text-zinc-600 text-center leading-5',
    }, props.on ? 'ON' : '·')
  },
})

const stackDiagram = `
  ┌──────────────────────────────────┐
  │  Stack 2 — 8-Relay HAT           │  ch 16–23  DIP: OFF ON ON
  ├──────────────────────────────────┤
  │  Stack 1 — 8-Relay HAT           │  ch  8–15  DIP: ON  OFF OFF
  ├──────────────────────────────────┤
  │  Stack 0 — 8-Relay HAT           │  ch  0–7   DIP: OFF OFF OFF  ← start here
  ├──────────────────────────────────┤
  │  Raspberry Pi 4B / 5             │
  │  I²C: GPIO2 (SDA) GPIO3 (SCL)   │
  └──────────────────────────────────┘
         ↕ only 2 GPIO pins used
`.trim()

const yamlExample = `actuators:
  - name: "Irrigation pump"
    channel_id: 0          # stack 0, relay 1
  - name: "Grow lights"
    channel_id: 4          # stack 0, relay 5
  - name: "Fan"
    channel_id: 5          # stack 0, relay 6
  - name: "Nutrient pump A"
    channel_id: 8          # stack 1, relay 1`.trim()

const parts = [
  {
    icon: '🖥️',
    label: 'Raspberry Pi 4B or 5',
    note: 'Pi 5 recommended for new builds. Zero 2W works for small setups.',
    qty: '1×',
    url: null,
    urlLabel: null,
  },
  {
    icon: '🔌',
    label: 'Sequent Microsystems 8-Relay HAT',
    note: '8 relays, NO/NC contacts, 4A/120VAC each. Stackable to 8 cards = 64 relays. I²C only.',
    qty: '1× (start)',
    url: 'https://sequentmicrosystems.com/products/eight-relays-stackable-card-for-raspberry-pi',
    urlLabel: 'sequentmicrosystems.com',
  },
  {
    icon: '⚡',
    label: '5V / 8A power supply with pluggable connector',
    note: 'Powers both the Pi and the relay card from a single supply. Each relay draws ~80mA at turn-on.',
    qty: '1×',
    url: null,
    urlLabel: null,
  },
  {
    icon: '🔧',
    label: '14–16 AWG wire for load side, 22 AWG for signal',
    note: 'Use stranded wire with ferrule crimps for pluggable connectors on the HAT.',
    qty: '—',
    url: null,
    urlLabel: null,
  },
  {
    icon: '📦',
    label: 'DIN rail enclosure (optional but recommended)',
    note: 'Fits Pi stack neatly in a control cabinet alongside breakers and contactors.',
    qty: '1×',
    url: null,
    urlLabel: null,
  },
]

const dipTable = [
  { level: 0, i2c: '0x27', id0: false, id1: false, id2: false, channels: 'ch 0 – 7' },
  { level: 1, i2c: '0x26', id0: true,  id1: false, id2: false, channels: 'ch 8 – 15' },
  { level: 2, i2c: '0x25', id0: false, id1: true,  id2: false, channels: 'ch 16 – 23' },
  { level: 3, i2c: '0x24', id0: true,  id1: true,  id2: false, channels: 'ch 24 – 31' },
  { level: 4, i2c: '0x23', id0: false, id1: false, id2: true,  channels: 'ch 32 – 39' },
  { level: 5, i2c: '0x22', id0: true,  id1: false, id2: true,  channels: 'ch 40 – 47' },
  { level: 6, i2c: '0x21', id0: false, id1: true,  id2: true,  channels: 'ch 48 – 55' },
  { level: 7, i2c: '0x20', id0: true,  id1: true,  id2: true,  channels: 'ch 56 – 63' },
]

const channelMapCards = [
  {
    stack: 0, i2c: '0x27',
    channels: Array.from({ length: 8 }, (_, i) => ({ relay: i + 1, channel: i })),
  },
  {
    stack: 1, i2c: '0x26',
    channels: Array.from({ length: 8 }, (_, i) => ({ relay: i + 1, channel: i + 8 })),
  },
]

const wiringPlan = [
  { ch: 'ch0', relay: 'Relay 1', use: 'Main irrigation pump',   load: 'Pump 120VAC',  notes: 'Via contactor if >4A' },
  { ch: 'ch1', relay: 'Relay 2', use: 'Nutrient dosing A',      load: 'Peristaltic',  notes: 'Usually 24VDC' },
  { ch: 'ch2', relay: 'Relay 3', use: 'Nutrient dosing B',      load: 'Peristaltic',  notes: 'pH up/down' },
  { ch: 'ch3', relay: 'Relay 4', use: 'Drain / return pump',    load: 'Pump 120VAC',  notes: 'Or CO₂ solenoid' },
  { ch: 'ch4', relay: 'Relay 5', use: 'Grow lights',            load: 'LED / HID',    notes: 'Contactor for HID' },
  { ch: 'ch5', relay: 'Relay 6', use: 'Exhaust fan',            load: 'Fan 120VAC',   notes: 'NC contact = fail-safe on' },
  { ch: 'ch6', relay: 'Relay 7', use: 'Humidifier / dehumid',   load: 'Appliance',    notes: '' },
  { ch: 'ch7', relay: 'Relay 8', use: 'Spare / heater',         load: '—',            notes: 'Reserve for expansion' },
]

const setupSteps = [
  {
    n: 1,
    title: 'Flash Pi OS and enable I²C',
    body: 'Use Raspberry Pi Imager. In raspi-config or the imager advanced settings, enable the I²C interface. Reboot.',
    code: `sudo raspi-config  # Interface Options → I2C → Enable
sudo reboot`,
  },
  {
    n: 2,
    title: 'Set DIP switch on relay card, then stack it',
    body: 'First card: all 3 bits OFF (stack level 0, I²C 0x27). Power off the Pi, slide the HAT onto the 40-pin header. Never stack or unstack with power on.',
    code: null,
  },
  {
    n: 3,
    title: 'Install Sequent relay drivers',
    body: 'Clone the 8relind library and install it. This gives you the 8relind CLI and Python module.',
    code: `cd ~
git clone https://github.com/SequentMicrosystems/8relind-rpi.git
cd 8relind-rpi
sudo make install
# Test: turn relay 1 on then off
8relind 0 write 1 on
8relind 0 write 1 off`,
  },
  {
    n: 4,
    title: 'Verify I²C addresses',
    body: 'i2cdetect shows which cards are visible. Stack 0 at 0x27, stack 1 at 0x26, etc.',
    code: `sudo apt install -y i2c-tools
i2cdetect -y 1
# Stack 0 alone → should show 27
# Stack 0 + 1   → should show 27 and 26`,
  },
  {
    n: 5,
    title: 'Install the gr33n Pi client and configure wiring',
    body: 'Clone the repo and copy the bootstrap config. Set your API URL, API key, and device UID. Wiring (channel→actuator mapping) is pulled from the platform — no YAML editing for each channel.',
    code: `cd ~
git clone <your-gr33n-repo>
cd gr33n-platform/pi_client
cp config.bootstrap.example.yaml config.yaml
# Edit: api.base_url, api.api_key, device.uid
nano config.yaml`,
  },
  {
    n: 6,
    title: 'Add actuators in the dashboard and assign channels',
    body: 'In gr33n → Controls → New actuator. Set the actuator type (relay), channel_id (0 for relay 1 on card 0), and assign it to a zone. The Pi pulls this wiring from the API at startup.',
    code: null,
  },
  {
    n: 7,
    title: 'Wire physical loads to relay terminals',
    body: 'Use the pluggable connectors. Common → C terminal, load → NO terminal (normally open = load off when idle). For fail-safe loads (exhaust fan) use NC terminal instead.',
    code: null,
  },
  {
    n: 8,
    title: 'Run the client and verify',
    body: 'Start the Pi client and confirm the device shows online in the dashboard. Trigger a manual pulse from Controls to test each relay LED on the card.',
    code: `python3 gr33n_client.py --config config.yaml
# In another terminal or the gr33n UI:
# Controls → your actuator → Manual pulse → 1s`,
  },
]

const scaleTiers = [
  {
    name: 'Starter room',
    relays: 8,
    desc: '1 × HAT. One grow room: pump, lights, fan, dosing, spare. Single Pi.',
    highlight: false,
  },
  {
    name: 'Full farm',
    relays: '8–32',
    desc: '2–4 × HATs stacked. Multiple rooms or flower + veg + mothers on one Pi.',
    highlight: true,
  },
  {
    name: 'Max per Pi',
    relays: 64,
    desc: '8 × HATs stacked. Large multi-room install. Add a second Pi for redundancy or zone separation.',
    highlight: false,
  },
]

const inputCards = [
  {
    name: 'Eight HV Digital Inputs HAT',
    desc: '8 opto-isolated inputs, 3–32V or 32–240V AC/DC. Float switches, door sensors, flow pulse counters.',
    channels: '8 inputs / card',
  },
  {
    name: 'Building Automation HAT',
    desc: '8 universal inputs: 1K/10K thermistors, 0–10V, dry contact. Temp, humidity, EC probes via signal conditioner.',
    channels: '8 inputs + 4 TRIAC out / card',
  },
  {
    name: 'Industrial Automation HAT',
    desc: 'Digital + analog I/O, MOSFET outputs, RS485/MODBUS. Integrates VFDs, variable-speed pumps, CO₂ controllers.',
    channels: 'Mixed analog + digital',
  },
]
</script>

