import { describe, it, expect } from 'vitest'
import { routeContextRefFromRoute } from '../lib/guardianRouteRef.js'

describe('Phase 32 WS1 — guardianRouteRef', () => {
  it('maps fertigation path to route context_ref', () => {
    expect(routeContextRefFromRoute({ path: '/fertigation', meta: {} })).toEqual({
      type: 'route',
      path: '/fertigation',
      name: 'Feeding (technical)',
    })
  })

  it('maps feeding hub path to route context_ref', () => {
    expect(routeContextRefFromRoute({ path: '/feeding', meta: {} })).toEqual({
      type: 'route',
      path: '/feeding',
      name: 'Feed & water',
    })
  })

  it('Phase 43 — maps operations hub paths to farmer labels', () => {
    expect(routeContextRefFromRoute({ path: '/operations/supplies', meta: {} })).toEqual({
      type: 'route',
      path: '/operations/supplies',
      name: 'Supplies',
    })
    expect(routeContextRefFromRoute({ path: '/operations/money', meta: {} })).toMatchObject({
      name: 'Money',
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
