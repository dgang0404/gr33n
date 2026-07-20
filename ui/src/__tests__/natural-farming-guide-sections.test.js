/**
 * Phase 209 WS3 — field guide section parser tests.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { batchStepCards, extractGuideSections } from '../lib/naturalFarmingGuideSections.js'

const jmsGuide = readFileSync(
  join(process.cwd(), '..', 'docs/field-guides/natural-farming-jms.md'),
  'utf8',
)

describe('naturalFarmingGuideSections', () => {
  it('extracts Ingredients and Step-by-step from JMS guide', () => {
    const sections = extractGuideSections(jmsGuide)
    expect(Object.keys(sections).some((k) => k.startsWith('Ingredients'))).toBe(true)
    expect(Object.keys(sections).some((k) => k.startsWith('Step-by-step preparation'))).toBe(true)
  })

  it('batchStepCards returns five instructional cards for JMS', () => {
    const cards = batchStepCards(jmsGuide)
    expect(cards.length).toBeGreaterThanOrEqual(5)
    expect(cards.map((c) => c.key)).toEqual(
      expect.arrayContaining(['ingredients', 'steps', 'timeline', 'ready', 'safety']),
    )
    const steps = cards.find((c) => c.key === 'steps')
    expect(steps?.body).toMatch(/Boil potato/)
  })
})
