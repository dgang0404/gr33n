import { describe, it, expect } from 'vitest'
import { cropImageAlt } from '../lib/cropLibraryPicker.js'

describe('Phase 107 — crop catalog photos', () => {
  it('cropImageAlt uses display_name', () => {
    expect(cropImageAlt({ display_name: 'San Pedro', crop_key: 'san_pedro' })).toBe(
      'San Pedro catalog thumbnail',
    )
  })

  it('CropLibraryPicker renders list with thumbnail slot', async () => {
    const { readFileSync } = await import('node:fs')
    const { join } = await import('node:path')
    const vue = readFileSync(join(process.cwd(), 'src/components/CropLibraryPicker.vue'), 'utf8')
    expect(vue).toContain('crop-library-picker-list')
    expect(vue).toContain('item.image_url')
    expect(vue).toContain('cropImageAlt')
  })
})
