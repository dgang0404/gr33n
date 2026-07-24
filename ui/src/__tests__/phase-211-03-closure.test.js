import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { FARM_SCOPES, FARM_SCOPE_OPTIONS } from '../lib/farmScopes.js'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

describe('Phase 211.03 farm permissions closure', () => {
  it('exports scope catalog aligned with backend ids', () => {
    expect(FARM_SCOPE_OPTIONS.length).toBeGreaterThan(8)
    expect(FARM_SCOPES.nfRecipesDelete).toBe('nf.recipes.delete')
    expect(FARM_SCOPES.moneyWrite).toBe('money.costs.write')
  })

  it('useFarmCaps composable fails closed on load error', () => {
    const src = readFileSync(join(root, 'composables/useFarmCaps.js'), 'utf8')
    expect(src).toContain('loadError.value = true')
    expect(src).toContain('if (loadError.value) return false')
  })

  it('useFarmOperate delegates to useFarmCaps farm.operate', () => {
    const src = readFileSync(join(root, 'composables/useFarmOperate.js'), 'utf8')
    expect(src).toContain('useFarmCaps')
    expect(src).not.toContain('Fail open')
  })

  it('SuppliesHub gates restock and unit cost by scope', () => {
    const src = readFileSync(join(root, 'views/SuppliesHub.vue'), 'utf8')
    expect(src).toContain('FARM_SCOPES.nfBatchesWrite')
    expect(src).toContain('FARM_SCOPES.moneyWrite')
  })

  it('RecipesApplyPanel gates delete by nf.recipes.delete', () => {
    const src = readFileSync(join(root, 'components/naturalfarming/RecipesApplyPanel.vue'), 'utf8')
    expect(src).toContain('FARM_SCOPES.nfRecipesDelete')
  })

  it('farm store loads me/caps endpoint', () => {
    const src = readFileSync(join(root, 'stores/farm.js'), 'utf8')
    expect(src).toContain('/me/caps')
  })
})
