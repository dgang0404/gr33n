import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFarmStore } from '../stores/farm'

const OPERATE_ROLES = new Set(['owner', 'manager', 'operator', 'worker', 'agronomist'])

/**
 * Phase 29 WS4 — whether the current user may confirm Guardian write actions
 * on the selected farm (mirrors backend FarmCaps.Operate).
 */
export function useFarmOperate(farmIdSource) {
  const canOperate = ref(true)
  const loading = ref(false)
  const auth = useAuthStore()
  const farmStore = useFarmStore()

  async function refresh(fid) {
    if (!fid) {
      canOperate.value = false
      return
    }
    if (auth.isDevMode) {
      canOperate.value = true
      return
    }
    loading.value = true
    try {
      const members = await farmStore.loadFarmMembers(fid)
      const me = members.find((m) => m.user_id === auth.userId)
      canOperate.value = !!(me && OPERATE_ROLES.has(me.role_in_farm))
    } catch {
      // Fail open for UX — server still enforces Operate on confirm.
      canOperate.value = true
    } finally {
      loading.value = false
    }
  }

  watch(
    farmIdSource,
    (fid) => { refresh(fid) },
    { immediate: true },
  )

  return { canOperate, loading, refresh }
}
