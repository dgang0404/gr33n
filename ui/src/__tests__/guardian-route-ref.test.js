import { describe, it, expect } from 'vitest'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'

describe('Phase 32 WS1 — guardianRouteRef', () => {
  it('maps fertigation path to route context_ref', () => {
    expect(routeContextRefFromRoute({ path: '/fertigation', meta: {} })).toEqual({
      type: 'route',
      path: '/fertigation',
      name: 'Fertigation',
    })
  })

  it('skips public auth routes', () => {
    expect(routeContextRefFromRoute({ path: '/login', meta: { public: true } })).toBeNull()
  })

  it('labels zone detail paths', () => {
    const ref = routeContextRefFromRoute({ path: '/zones/12', meta: {} })
    expect(ref).toEqual({
      type: 'route',
      path: '/zones/12',
      name: 'Zone detail',
    })
  })
})
