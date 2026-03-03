<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-bold text-white">JADAM Inputs & Recipes</h2>
      <button class="text-xs bg-gr33n-700 hover:bg-gr33n-600 text-white px-3 py-1.5 rounded-lg transition-colors">
        + Add Input
      </button>
    </div>
    <div class="card overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="text-left text-gray-500 text-xs uppercase tracking-wide border-b border-gray-800">
            <th class="pb-2 pr-4">Input</th>
            <th class="pb-2 pr-4">Type</th>
            <th class="pb-2 pr-4">Dilution</th>
            <th class="pb-2 pr-4">Application</th>
            <th class="pb-2">Status</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-800">
          <tr v-for="item in inputs" :key="item.id" class="hover:bg-gray-800/50 transition-colors">
            <td class="py-2 pr-4 font-semibold text-white">{{ item.name }}</td>
            <td class="py-2 pr-4 text-gray-400">{{ item.type }}</td>
            <td class="py-2 pr-4 font-mono text-gray-300">{{ item.dilution }}</td>
            <td class="py-2 pr-4 text-gray-400">{{ item.application }}</td>
            <td class="py-2"><span class="badge-ok">Ready</span></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup>
// Placeholder — wire to /api/inputs once that endpoint is built
const inputs = [
  { id: 1, name: 'JMS',  type: 'Microbial',   dilution: '1:1000', application: 'Soil drench' },
  { id: 2, name: 'LAB',  type: 'Microbial',   dilution: '1:500',  application: 'Foliar spray' },
  { id: 3, name: 'FPJ',  type: 'Nutrient',    dilution: '1:500',  application: 'Foliar spray' },
  { id: 4, name: 'OHN',  type: 'Pest control',dilution: '1:1000', application: 'Soil & foliar' },
  { id: 5, name: 'WCA',  type: 'Mineral',     dilution: '1:1000', application: 'Soil drench' },
]
</script>
