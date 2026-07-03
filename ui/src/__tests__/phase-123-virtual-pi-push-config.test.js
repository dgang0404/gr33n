import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

describe('Phase 123 — Pi config push notify', () => {
  it('API route registers push-config', () => {
    const routes = readFileSync(join(process.cwd(), '../cmd/api/routes.go'), 'utf8')
    expect(routes).toContain('POST /devices/{id}/push-config')
  })

  it('VirtualPi has notify Pi button', () => {
    const view = readFileSync(join(process.cwd(), 'src/views/VirtualPi.vue'), 'utf8')
    expect(view).toContain('virtual-pi-push-config')
    expect(view).toContain('push-config')
    expect(view).toContain('deviceUsesPlatformSync')
  })
})
