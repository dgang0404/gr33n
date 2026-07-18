/**
 * Phase 189 — inline source-metadata + placeholder-citation redaction.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 189 — Guardian inline metadata redaction', () => {
  it('adds RedactInlineSourceMetadata and RedactPlaceholderCitationMarkers', () => {
    const src = readFileSync(join(repoRoot, 'internal/farmguardian/answer_inline_metadata.go'), 'utf8')
    expect(src).toContain('func RedactInlineSourceMetadata')
    expect(src).toContain('func RedactPlaceholderCitationMarkers')
    expect(src).toContain('inlineSourceIDRE')
    expect(src).toContain('inlineDocPathRE')
    expect(src).toContain('placeholderCiteLiteralNRE')
  })

  it('wires both redactors into the chat answer-finalize hygiene pipeline', () => {
    const finalize = readFileSync(join(repoRoot, 'internal/handler/chat/answer_finalize.go'), 'utf8')
    expect(finalize).toContain('farmguardian.RedactInlineSourceMetadata(answer)')
    expect(finalize).toContain('farmguardian.RedactPlaceholderCitationMarkers(answer)')
    expect(finalize).toContain('inlineMetadata')
    expect(finalize).toContain('placeholderCite')
  })

  it('turn debug surfaces inline-metadata and placeholder-citation redaction flags', () => {
    const debug = readFileSync(join(repoRoot, 'internal/farmguardian/turn_debug.go'), 'utf8')
    expect(debug).toContain('InlineMetadataRedacted')
    expect(debug).toContain('PlaceholderCitationRedacted')
  })

  it('has regression tests built from live inline-metadata leaks', () => {
    const test = readFileSync(join(repoRoot, 'internal/farmguardian/answer_inline_metadata_test.go'), 'utf8')
    expect(test).toContain('TestRedactInlineSourceMetadata_liveHighHumidityAlert')
    expect(test).toContain('TestRedactPlaceholderCitationMarkers_literalN')
  })
})
