/**
 * Phase 203 — Handler package consolidation guardrails.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync, existsSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

function goFilesUnder(dir) {
  const out = []
  for (const name of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, name.name)
    if (name.isDirectory()) out.push(...goFilesUnder(path))
    else if (name.name.endsWith('.go') && !name.name.endsWith('_test.go')) out.push(path)
  }
  return out
}

describe('Phase 203 — handler package consolidation', () => {
  it('NumericFromFloat64 lives only in httputil', () => {
    const hits = []
    for (const file of goFilesUnder(join(repoRoot, 'internal'))) {
      const src = readFileSync(file, 'utf8')
      if (/func numericFromFloat64/.test(src)) hits.push(file.replace(repoRoot + '/', ''))
    }
    expect(hits).toEqual([])
    const httputil = readFileSync(join(repoRoot, 'internal/httputil/pgconv.go'), 'utf8')
    expect(httputil).toContain('func NumericFromFloat64')
  })

  it('ParseLimitOffset helpers ship in httputil', () => {
    const query = readFileSync(join(repoRoot, 'internal/httputil/query.go'), 'utf8')
    expect(query).toContain('func ParseLimitOffset')
    expect(query).toContain('func ParseLimitOffsetStrict')
    expect(query).toContain('func PathValueInt64')
  })

  it('recipe handler merged into naturalfarming package', () => {
    expect(existsSync(join(repoRoot, 'internal/handler/naturalfarming/recipe.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/handler/recipe'))).toBe(false)
    const recipe = readFileSync(join(repoRoot, 'internal/handler/naturalfarming/recipe.go'), 'utf8')
    expect(recipe).toContain('func (h *Handler) ListRecipes')
  })

  it('commons crop catalog merged into commonscatalog package', () => {
    expect(existsSync(join(repoRoot, 'internal/handler/commonscatalog/crop_catalog.go'))).toBe(true)
    expect(existsSync(join(repoRoot, 'internal/handler/commonscropcatalog'))).toBe(false)
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).not.toContain('commonscropcatalog')
    expect(routes).toContain('commonsCatalog.ListCropCatalog')
  })

  it('routes register commons and natural farming in unified blocks', () => {
    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('// Commons —')
    expect(routes).toContain('// Natural farming —')
    expect(routes).not.toContain('recipehandler')
  })

  it('dead devicecmd pathSegment helpers removed', () => {
    const devicecmd = readFileSync(join(repoRoot, 'internal/handler/devicecmd/handler.go'), 'utf8')
    expect(devicecmd).not.toContain('func pathSegment')
    expect(devicecmd).not.toContain('func idSegment')
  })

  it('unused commontypes validation enums removed', () => {
    const enums = readFileSync(join(repoRoot, 'internal/platform/commontypes/enums.go'), 'utf8')
    expect(enums).not.toContain('ValidationRuleTypeEnum')
    expect(enums).not.toContain('UserActionTypeEnum')
  })
})
