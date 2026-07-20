/**
 * Phase 208 WS5 — read API routes + static YAML serve closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
const openapi = readFileSync(join(repoRoot, 'openapi.yaml'), 'utf8')
const catalogHandler = readFileSync(join(repoRoot, 'internal/handler/fieldguides/catalog.go'), 'utf8')

describe('Phase 208 WS5 — read API', () => {
  it('registers process-catalog and recipe-canon routes', () => {
    expect(routes).toContain('GET /v1/field-guides/process-catalog/materials/{id}')
    expect(routes).toContain('GET /v1/field-guides/process-catalog')
    expect(routes).toContain('GET /v1/field-guides/recipe-canon')
  })

  it('documents endpoints in OpenAPI', () => {
    expect(openapi).toContain('/v1/field-guides/process-catalog:')
    expect(openapi).toContain('/v1/field-guides/process-catalog/materials/{id}:')
    expect(openapi).toContain('/v1/field-guides/recipe-canon:')
  })

  it('handlers load canonical YAML from data/', () => {
    expect(catalogHandler).toContain('LoadMaterialCatalog')
    expect(catalogHandler).toContain('LoadRecipeCanon')
    expect(catalogHandler).toContain('MaterialByID')
  })
})
