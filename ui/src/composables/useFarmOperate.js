import { useFarmCaps } from './useFarmCaps'

/**
 * Phase 29 WS4 / 211.03 — Guardian Confirm gate (farm.operate scope).
 */
export function useFarmOperate(farmIdSource) {
  const { canOperate, loading, refresh } = useFarmCaps(farmIdSource)
  return { canOperate, loading, refresh }
}
