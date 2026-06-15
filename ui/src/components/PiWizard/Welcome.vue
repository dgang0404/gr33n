<template>
  <div class="space-y-6">
    <section>
      <h2 class="text-2xl font-bold text-green-400 mb-3">Welcome to the Pi Setup Wizard</h2>
      <p class="text-zinc-400 leading-relaxed">
        Let's wire up your grow room with a Raspberry Pi and Sequent relay HAT.
        This wizard will walk you through hardware assembly, network setup, and deployment.
      </p>
    </section>

    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
      <h3 class="text-sm font-semibold uppercase tracking-widest text-zinc-500">Before you start — do you have:</h3>
      
      <div class="space-y-3">
        <label class="flex items-start gap-3 cursor-pointer">
          <input
            v-model="checklist.pi"
            type="checkbox"
            class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
          />
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">Raspberry Pi 4B or 5</div>
            <div class="text-xs text-zinc-500">Pi 5 recommended; Pi Zero 2W works for small setups</div>
          </div>
        </label>

        <label class="flex items-start gap-3 cursor-pointer">
          <input
            v-model="checklist.hat"
            type="checkbox"
            class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
          />
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">Sequent Microsystems 8-Relay HAT</div>
            <div class="text-xs text-zinc-500">Available at sequentmicrosystems.com/products</div>
          </div>
        </label>

        <label class="flex items-start gap-3 cursor-pointer">
          <input
            v-model="checklist.power"
            type="checkbox"
            class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
          />
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">5V / 8A power supply</div>
            <div class="text-xs text-zinc-500">Powers both Pi and relay card</div>
          </div>
        </label>

        <label class="flex items-start gap-3 cursor-pointer">
          <input
            v-model="checklist.wiring"
            type="checkbox"
            class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
          />
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">Wire & crimps</div>
            <div class="text-xs text-zinc-500">14–16 AWG stranded wire + ferrule crimping kit</div>
          </div>
        </label>

        <label class="flex items-start gap-3 cursor-pointer">
          <input
            v-model="checklist.network"
            type="checkbox"
            class="mt-1 w-4 h-4 rounded border-zinc-600 bg-zinc-950 text-green-600"
          />
          <div class="min-w-0">
            <div class="text-sm font-medium text-zinc-200">Ethernet or WiFi</div>
            <div class="text-xs text-zinc-500">Pi must be on same network as your computer</div>
          </div>
        </label>
      </div>

      <p class="text-xs text-zinc-600 pt-2">Don't have everything? You can order parts and come back later.</p>
    </section>

    <section class="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
      <h3 class="text-sm font-semibold text-green-400 mb-3">💡 What is a Relay HAT?</h3>
      <p class="text-sm text-zinc-400 leading-relaxed">
        A "HAT" is a hardware add-on for the Raspberry Pi. The Sequent relay HAT stacks on top
        and adds 8 controllable relays (switches) for pumps, lights, and fans. You can stack up to
        8 cards for 64 relays total. Each card uses just 2 GPIO pins via I²C, so no GPIO pins are wasted.
      </p>
    </section>
  </div>
</template>

<script setup>
import { reactive } from 'vue'

const checklist = reactive({
  pi: false,
  hat: false,
  power: false,
  wiring: false,
  network: false,
})
</script>

<style scoped></style>
