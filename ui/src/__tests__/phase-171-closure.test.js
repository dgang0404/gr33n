/**
 * Phase 171 — Demo farm zone layouts seed closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 171 — demo zone layouts in seed', () => {
  const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')

  it('persists meta_data.layout for all demo farm zones', () => {
    expect(seed).toContain('Phase 171 — demo farm canvas')
    expect(seed).toContain("jsonb_build_object('layout'")
    for (const name of [
      'Veg Room',
      'Flower Room',
      'Propagation Room',
      'Herb & Greens Room',
      'Outdoor Garden',
      'Outdoor Pepper Bed',
      'Outdoor Berry Patch',
    ]) {
      expect(seed).toContain(`('${name}'`)
    }
    expect(seed).toContain('"x":0.28,"y":0.06,"w":0.20,"h":0.18')
  })

  it('documents phase in current-state', () => {
    const state = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')
    expect(state).toContain('Phase 171')
  })
})
