/**
 * Phase 153 — Guardian change-request ("PR queue") smoke fetcher.
 * Script-only (no CI/GitHub involvement) — fires write-intent prompts, then
 * fetches GET /v1/chat/proposals?status=pending to confirm they landed.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 153 — Guardian change-request smoke fetcher', () => {
  it('plan documents the pending-proposal fetch, not a GitHub PR gate', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/archive/phase_153_guardian_pr_smoke_gate.plan.md'),
      'utf8',
    )
    expect(plan).toContain('FetchPendingProposals')
    expect(plan).toContain('check-pending-proposals')
    expect(plan).toContain('guardian-qa-change-requests')
    expect(plan).toContain('Not tied to GitHub in any way')
  })

  it('eval package ships change-request fixtures and a pending-proposal fetcher', () => {
    const fixtures = readFileSync(
      join(repoRoot, 'internal/farmguardian/eval/fixtures_change_requests.go'),
      'utf8',
    )
    expect(fixtures).toContain('func ChangeRequestFixtures')

    const proposals = readFileSync(join(repoRoot, 'internal/farmguardian/eval/proposals.go'), 'utf8')
    expect(proposals).toContain('func (c *APIClient) FetchPendingProposals')
    expect(proposals).toContain('/v1/chat/proposals')
    expect(proposals).toContain('status=pending')

    const smoke = readFileSync(join(repoRoot, 'internal/farmguardian/eval/fixtures_smoke.go'), 'utf8')
    expect(smoke).toContain('change-requests')
  })

  it('cmd/guardian-eval fails when fewer pending rows than expected show up', () => {
    const main = readFileSync(join(repoRoot, 'cmd/guardian-eval/main.go'), 'utf8')
    expect(main).toContain('check-pending-proposals')
    expect(main).toContain('func reportPendingProposals')
    expect(main).toContain('func passedProposalFixtures')
  })

  it('Makefile ships guardian-qa-change-requests; smoke gate is opt-in CI only', () => {
    const makefile = readFileSync(join(repoRoot, 'Makefile'), 'utf8')
    expect(makefile).toContain('guardian-qa-change-requests')
    expect(makefile).toContain('check-pending-proposals')
    expect(makefile).toContain('guardian-qa-smoke-strict')

    const ci = readFileSync(join(repoRoot, '.github/workflows/ci.yml'), 'utf8')
    expect(ci).toContain('guardian-qa-pr')
    expect(ci).toContain('guardian-smoke')
    expect(ci).toContain('guardian-qa-smoke-strict')
    expect(ci).not.toMatch(/make guardian-qa-change-requests/)
  })
})
