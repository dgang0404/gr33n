/**
 * Phase 208 WS2 — natural farming field guides closure.
 */
import { describe, it, expect } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')
const dir = join(repoRoot, 'docs/field-guides')

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

const REQUIRED_SECTIONS = [
  '## What it is',
  '## When to use',
  '## Ingredients',
  '## Step-by-step preparation',
  '## Ferment / wait timeline',
  '## Ready signs',
  '## Storage',
  '## Safety & water',
  '## How to apply',
  '## Dilution table',
  '## Common mistakes',
]

describe('Phase 208 WS2 — field guides', () => {
  it('ships all 18 required natural-farming guide files', () => {
    for (const f of GUIDES) {
      expect(existsSync(join(dir, f)), f).toBe(true)
    }
    expect(GUIDES.length).toBe(18)
  })

  it('each guide has required instructional sections and front matter', () => {
    for (const f of GUIDES) {
      const text = readFileSync(join(dir, f), 'utf8')
      expect(text).toMatch(/^---\n/m)
      expect(text).toContain('safety_tier:')
      expect(text).toContain('tradition:')
      expect(text).toContain('source_tier:')
      for (const sec of REQUIRED_SECTIONS) {
        expect(text, `${f} missing ${sec}`).toContain(sec)
      }
    }
  })

  it('JMS guide cites post-audit 1:10 / 1:20 dilutions not 1:500', () => {
    const jms = readFileSync(join(dir, 'natural-farming-jms.md'), 'utf8')
    expect(jms).toContain('1:10')
    expect(jms).toContain('1:20')
    expect(jms).not.toMatch(/1:500.*JMS/i)
  })

  it('goldenrod guide is extension_method not cho_named', () => {
    const g = readFileSync(join(dir, 'natural-farming-goldenrod-jlf.md'), 'utf8')
    expect(g).toContain('source_tier: extension_method')
    expect(g).toMatch(/not a named Cho/i)
    expect(g).toContain('1:100')
  })

  it('application recipes table lists all 14 seed recipe names', () => {
    const t = readFileSync(join(dir, 'natural-farming-application-recipes.md'), 'utf8')
    const names = [
      'JMS Soil Drench',
      'JLF General Soil Drench',
      'JLF Seedling Drench',
      'JLF and JMS Combined Drench',
      'LAB Soil Conditioner',
      'OHN Pest and Immunity Drench',
      'JMS Foliar Spray',
      'FPJ Vegetative Foliar',
      'FFJ and WCA Flowering Boost',
      'BRV and WCA Cell Strengthener',
      'JHS and JWA Natural Pesticide',
      'JS Fungicide Spray',
      'JLF Foliar Feed',
      'JWA Insecticide Spray',
    ]
    for (const n of names) {
      expect(t).toContain(n)
    }
  })
})
