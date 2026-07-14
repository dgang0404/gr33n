<template>
  <section class="space-y-4" data-test="field-guide-browse">
    <div class="space-y-1">
      <h2 class="text-white font-semibold text-sm uppercase tracking-widest text-zinc-500">
        Field guides
      </h2>
      <p class="text-xs text-zinc-500 leading-relaxed">
        Curated install and crop-care docs — the same material Guardian cites in grounded answers.
      </p>
    </div>

    <div v-if="loading" class="text-sm text-zinc-500">Loading field guides…</div>
    <div v-else-if="error" class="text-sm text-red-400">{{ error }}</div>
    <template v-else>
      <div class="flex flex-wrap gap-3 items-end">
        <div>
          <label class="block text-xs text-zinc-500 mb-1" for="field-guide-crop-filter">Crop</label>
          <select
            id="field-guide-crop-filter"
            v-model="cropFilter"
            class="bg-zinc-950 border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white min-w-[10rem]"
            data-test="field-guide-crop-filter"
          >
            <option value="">All crops</option>
            <option v-for="c in cropOptions" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
        <p class="text-xs text-zinc-600 pb-1">{{ filteredGuides.length }} guide{{ filteredGuides.length === 1 ? '' : 's' }}</p>
      </div>

      <div
        class="max-h-[min(50vh,24rem)] overflow-y-auto space-y-2 pr-1"
        data-test="field-guide-list"
      >
        <button
          v-for="g in filteredGuides"
          :key="g.slug"
          type="button"
          class="w-full text-left rounded-lg border px-4 py-3 transition-colors"
          :class="selectedSlug === g.slug
            ? 'border-green-700 bg-green-950/30'
            : 'border-zinc-800 bg-zinc-950/60 hover:border-zinc-600'"
          :data-test="`field-guide-row-${g.slug}`"
          @click="selectGuide(g)"
        >
          <p class="text-sm font-medium text-zinc-100">{{ g.title }}</p>
          <p class="text-[11px] text-zinc-500 mt-0.5">
            <span v-if="g.crop_key">{{ g.crop_key }}</span>
            <span v-else>General</span>
            <span v-if="g.guide_kind"> · {{ g.guide_kind }}</span>
          </p>
        </button>
      </div>

      <article
        v-if="selectedDetail"
        class="rounded-xl border border-zinc-800 bg-zinc-900 p-4 space-y-3"
        data-test="field-guide-detail"
      >
        <div class="flex flex-wrap items-start justify-between gap-2">
          <h3 class="text-sm font-semibold text-green-300">{{ selectedDetail.title }}</h3>
          <span v-if="selectedDetail.safety_tier" class="text-[10px] uppercase text-zinc-500">
            {{ selectedDetail.safety_tier }}
          </span>
        </div>
        <pre
          v-if="selectedDetail.body_md"
          class="text-sm text-zinc-300 whitespace-pre-wrap font-sans leading-relaxed max-h-64 overflow-y-auto"
          data-test="field-guide-body"
        >{{ selectedDetail.body_md }}</pre>
        <div class="flex flex-wrap gap-2 pt-1">
          <button
            type="button"
            class="text-xs px-3 py-1.5 rounded-lg bg-green-900/50 text-green-400 border border-green-800"
            data-test="field-guide-open-doc"
            @click="openInKnowledge(selectedDetail)"
          >
            Open indexed doc
          </button>
          <router-link
            :to="guardianChatLink(selectedDetail)"
            class="text-xs px-3 py-1.5 rounded-lg border border-zinc-700 text-zinc-300 hover:text-green-300"
            data-test="field-guide-ask-guardian"
          >
            Ask Guardian about this
          </router-link>
        </div>
      </article>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const emit = defineEmits([])

const router = useRouter()
const guides = ref([])
const loading = ref(false)
const error = ref('')
const cropFilter = ref('')
const selectedSlug = ref('')
const selectedDetail = ref(null)
const detailLoading = ref(false)

const cropOptions = computed(() => {
  const out = new Set()
  for (const g of guides.value) {
    const key = String(g.crop_key || '').trim()
    if (key) out.add(key)
  }
  return [...out].sort((a, b) => a.localeCompare(b))
})

const filteredGuides = computed(() => {
  const crop = cropFilter.value.trim().toLowerCase()
  return guides.value.filter((g) => {
    if (!crop) return true
    return String(g.crop_key || '').toLowerCase() === crop
  })
})

function citedDocPath(guide) {
  const slug = String(guide?.slug || '').trim()
  if (!slug) return ''
  return slug.endsWith('.md') ? `field-guides/${slug}` : `field-guides/${slug}.md`
}

function openInKnowledge(guide) {
  const cited = citedDocPath(guide)
  if (!cited) return
  router.replace({
    path: '/operator-guide',
    query: {
      tab: 'knowledge',
      cited_doc: cited,
      cited_type: 'field_guide',
    },
  })
}

function guardianChatLink(guide) {
  const cited = citedDocPath(guide)
  return {
    path: '/chat',
    query: cited ? { cited_doc: cited, cited_type: 'field_guide' } : {},
  }
}

async function selectGuide(guide) {
  selectedSlug.value = guide.slug
  detailLoading.value = true
  try {
    const { data } = await api.get(`/commons/agronomy-field-guides/${encodeURIComponent(guide.slug)}`)
    selectedDetail.value = data
  } catch (e) {
    selectedDetail.value = { ...guide, body_md: e.response?.data?.error || e.message || 'Failed to load guide' }
  } finally {
    detailLoading.value = false
  }
}

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/commons/agronomy-field-guides')
    guides.value = Array.isArray(data) ? data : []
    guides.value.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || String(a.title).localeCompare(String(b.title)))
  } catch (e) {
    error.value = e.response?.data?.error || e.message || 'Failed to load field guides'
    guides.value = []
  } finally {
    loading.value = false
  }
})
</script>
