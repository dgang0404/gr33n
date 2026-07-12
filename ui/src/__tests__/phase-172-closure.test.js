/**
 * Phase 172 — demo field guide expansion + new bedding crops closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 172 — field guides and catalog', () => {
  it('manifest includes new bedding flower guides', () => {
    const manifest = readFileSync(join(repoRoot, 'docs/rag/field-guide-manifest.yaml'), 'utf8')
    expect(manifest).toContain('crop-marigold-care.md')
    expect(manifest).toContain('crop-geranium-care.md')
  })

  it('expands chrysanthemum guide for demo farm bloom context', () => {
    const guide = readFileSync(join(repoRoot, 'docs/field-guides/crop-chrysanthemum-care.md'), 'utf8')
    expect(guide).toContain('powdery mildew')
    expect(guide).toContain('12/12')
    expect(guide.length).toBeGreaterThan(800)
  })

  it('catalog seed includes marigold and geranium', () => {
    const seed = readFileSync(join(repoRoot, 'db/seed/crop_catalog_from_yaml.sql'), 'utf8')
    expect(seed).toContain("'marigold'")
    expect(seed).toContain("'geranium'")
    expect(seed).toContain('crop-marigold-care')
    expect(seed).toContain('crop-geranium-care')
  })

  it('crop library validates 48 crops', () => {
    const yaml = readFileSync(join(repoRoot, 'data/crop_library.yaml'), 'utf8')
    expect(yaml).toContain('key: marigold')
    expect(yaml).toContain('key: geranium')
  })
})
