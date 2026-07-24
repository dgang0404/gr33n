import { computed, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFarmStore } from '../stores/farm'
import { ALL_FARM_SCOPES } from '../lib/farmScopes.js'

/**
 * Phase 211.03 — resolved farm scopes for UI affordances (API still enforces).
 * Fails closed on load error.
 */
export function useFarmCaps(farmIdSource) {
  const scopes = ref(new Set())
  const roleInFarm = ref('')
  const loading = ref(false)
  const loadError = ref(false)
  const auth = useAuthStore()
  const farmStore = useFarmStore()

  async function refresh(fid) {
    if (!fid) {
      scopes.value = new Set()
      roleInFarm.value = ''
      loadError.value = false
      return
    }
    if (auth.isDevMode) {
      scopes.value = new Set(ALL_FARM_SCOPES)
      roleInFarm.value = 'owner'
      loadError.value = false
      return
    }
    loading.value = true
    loadError.value = false
    try {
      const caps = await farmStore.loadFarmCaps(fid)
      roleInFarm.value = caps.role_in_farm || ''
      scopes.value = new Set(Array.isArray(caps.scopes) ? caps.scopes : [])
    } catch {
      scopes.value = new Set()
      loadError.value = true
    } finally {
      loading.value = false
    }
  }

  function has(scope) {
    if (loadError.value) return false
    return scopes.value.has(scope)
  }

  watch(
    farmIdSource,
    (fid) => { refresh(fid) },
    { immediate: true },
  )

  const canOperate = computed(() => has('farm.operate'))

  return { scopes, roleInFarm, loading, loadError, has, canOperate, refresh }
}
