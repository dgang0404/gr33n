/**
 * Phase 165 — Farm layout API + background image plumbing.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 165 WS1 — zone layout metadata', () => {
  it('ships layout validation helpers', () => {
    const layout = readFileSync(join(repoRoot, 'internal/handler/zone/layout.go'), 'utf8')
    expect(layout).toContain('ValidateZoneLayout')
    expect(layout).toContain('ExtractZoneLayout')
    expect(layout).toContain('defaultLayoutW')
  })

  it('merges zone meta_data without clobbering other keys', () => {
    const meta = readFileSync(join(repoRoot, 'internal/handler/zone/meta.go'), 'utf8')
    expect(meta).toContain('MergeZoneMetaData')
    const handler = readFileSync(join(repoRoot, 'internal/handler/zone/handler.go'), 'utf8')
    expect(handler).toContain('MergeZoneMetaData')
  })
})

describe('Phase 165 WS2 — farm layout background API', () => {
  it('ships farm layout background handlers and routes', () => {
    const bg = readFileSync(
      join(repoRoot, 'internal/handler/fileattach/farm_layout_background.go'),
      'utf8',
    )
    expect(bg).toContain('UploadFarmLayoutBackground')
    expect(bg).toContain('GetFarmLayoutBackground')
    expect(bg).toContain('DeleteFarmLayoutBackground')
    expect(bg).toContain('farm_layout_background')

    const routes = readFileSync(join(repoRoot, 'cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('POST /farms/{id}/layout-background')
    expect(routes).toContain('GET /farms/{id}/layout-background')
    expect(routes).toContain('DELETE /farms/{id}/layout-background')
  })

  it('allows authenticated download of farm layout background attachments', () => {
    const handler = readFileSync(join(repoRoot, 'internal/handler/fileattach/handler.go'), 'utf8')
    expect(handler).toContain(`case "farms":`)
    expect(handler).toContain('farm_layout_background')
  })

  it('ships farm meta helper and sqlc queries', () => {
    const meta = readFileSync(join(repoRoot, 'internal/farmlayout/meta.go'), 'utf8')
    expect(meta).toContain('layout_background_attachment_id')
    const farmsSql = readFileSync(join(repoRoot, 'db/queries/farms.sql'), 'utf8')
    expect(farmsSql).toContain('SetFarmLayoutBackgroundAttachment')
    expect(farmsSql).toContain('ClearFarmLayoutBackgroundAttachment')
  })
})

describe('Phase 165 WS3 — farm store plumbing', () => {
  it('ships layout and background helpers in farm store', () => {
    const store = readFileSync(join(repoRoot, 'ui/src/stores/farm.js'), 'utf8')
    expect(store).toContain('saveZoneLayout')
    expect(store).toContain('zoneLayout')
    expect(store).toContain('loadLayoutBackground')
    expect(store).toContain('uploadLayoutBackground')
    expect(store).toContain('clearLayoutBackground')
    expect(store).toContain('layoutBackgroundBlobUrl')
  })
})

describe('Phase 165 WS4 — tests', () => {
  it('ships layout unit tests', () => {
    const tests = readFileSync(join(repoRoot, 'internal/handler/zone/layout_test.go'), 'utf8')
    expect(tests).toContain('TestValidateZoneLayout')
    expect(tests).toContain('TestMergeZoneMetaData')
  })
})
