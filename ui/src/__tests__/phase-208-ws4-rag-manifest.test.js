/**
 * Phase 208 WS4 — RAG manifest + DB migration closure for natural farming guides.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const guidesDir = join(repoRoot, 'docs/field-guides')
const manifest = readFileSync(join(repoRoot, 'docs/rag/field-guide-manifest.yaml'), 'utf8')
const readme = readFileSync(join(repoRoot, 'docs/field-guides/README.md'), 'utf8')
const migration = readFileSync(
  join(repoRoot, 'db/migrations/20260720_phase208_ws4_natural_farming_field_guides.sql'),
  'utf8'
)
const fieldGuidesGo = readFileSync(join(repoRoot, 'internal/rag/ingest/field_guides.go'), 'utf8')

const GUIDES = [
  'natural-farming-jms.md',
  'natural-farming-jlf-general.md',
  'natural-farming-jlf-crop-specific.md',
  'natural-farming-jlf-spring-nettle-comfrey.md',
  'natural-farming-ffj.md',
  'natural-farming-fpj.md',
  'natural-farming-lab.md',
  'natural-farming-ohn.md',
  'natural-farming-wca-wcs.md',
  'natural-farming-jwa-js-jhs.md',
  'natural-farming-brv.md',
  'natural-farming-faa.md',
  'natural-farming-compost-tea-aact.md',
  'natural-farming-application-recipes.md',
  'natural-farming-indoor-photoperiod-program.md',
  'natural-farming-goldenrod-jlf.md',
  'natural-farming-forest-garden-understory.md',
  'natural-farming-livestock-plant-feed.md',
]

describe('Phase 208 WS4 — RAG manifest', () => {
  it('lists all 18 natural-farming guides in field-guide-manifest.yaml', () => {
    for (const f of GUIDES) {
      expect(manifest).toContain(`  - ${f}`)
    }
  })

  it('each guide has domain: natural_farming and tradition in front matter', () => {
    for (const f of GUIDES) {
      const text = readFileSync(join(guidesDir, f), 'utf8')
      expect(text).toContain('domain: natural_farming')
      expect(text).toMatch(/^tradition: /m)
    }
  })

  it('README manifests natural farming section', () => {
    expect(readme).toContain('Natural farming (Phase 208)')
    expect(readme).toContain('domain: natural_farming')
    expect(readme).toContain('natural-farming-jms.md')
    expect(readme).toContain('20260720_phase208_ws4_natural_farming_field_guides.sql')
  })

  it('DB migration upserts all guide slugs with natural_farming domain', () => {
    expect(existsSync(join(repoRoot, 'db/migrations/20260720_phase208_ws4_natural_farming_field_guides.sql'))).toBe(true)
    for (const f of GUIDES) {
      const slug = f.replace(/\.md$/, '')
      expect(migration).toContain(`'${slug}'`)
    }
    expect(migration).toContain("'natural_farming', 'natural_farming'")
    expect(migration).toContain("ON CONFLICT (slug) DO UPDATE")
  })

  it('ingest attaches tradition metadata for RAG chunks', () => {
    expect(fieldGuidesGo).toContain('"tradition"')
    expect(fieldGuidesGo).toContain('natural-farming-')
  })
})
