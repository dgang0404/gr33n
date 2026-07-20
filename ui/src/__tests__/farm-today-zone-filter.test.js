import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import {
  TODAY_ZONE_FILTERS,
  countZonesPerFilter,
  filterZonesForToday,
  paginateZones,
  readTodayDesktopView,
  readTodayZoneFilter,
  shouldOfferDesktopListView,
  shouldPageZoneStack,
  shouldShowTodayZoneFilterBar,
  totalZonePages,
  writeTodayDesktopView,
  writeTodayZoneFilter,
} from '../lib/farmTodayZoneFilter.js'
import { LARGE_FARM_ZONES, largeFarmStatusFor } from './fixtures/largeFarmZones.js'

describe('Phase 173 WS1 — farmTodayZoneFilter', () => {
  it('exposes the expected filter set', () => {
    expect(TODAY_ZONE_FILTERS.map((f) => f.id)).toEqual([
      'all', 'attention', 'indoor', 'outdoor', 'greenhouse',
    ])
  })

  it('filters by zone type', () => {
    const outdoor = filterZonesForToday(LARGE_FARM_ZONES, 'outdoor', largeFarmStatusFor)
    expect(outdoor.length).toBeGreaterThan(0)
    expect(outdoor.every((z) => z.zone_type === 'outdoor')).toBe(true)
  })

  it('filters by attention using getStatus', () => {
    const attention = filterZonesForToday(LARGE_FARM_ZONES, 'attention', largeFarmStatusFor)
    expect(attention.map((z) => z.id).sort((a, b) => a - b)).toEqual([1, 6, 11])
  })

  it('"all" returns the full list unfiltered', () => {
    expect(filterZonesForToday(LARGE_FARM_ZONES, 'all', largeFarmStatusFor)).toHaveLength(24)
  })

  it('counts zones per filter', () => {
    const counts = countZonesPerFilter(LARGE_FARM_ZONES, largeFarmStatusFor)
    expect(counts.all).toBe(24)
    expect(counts.attention).toBe(3)
    expect(counts.indoor + counts.outdoor + counts.greenhouse).toBe(24)
  })

  it('hides the filter bar below threshold, shows it above', () => {
    expect(shouldShowTodayZoneFilterBar(7)).toBe(false)
    expect(shouldShowTodayZoneFilterBar(9)).toBe(true)
    expect(shouldShowTodayZoneFilterBar(24)).toBe(true)
  })

  it('pages the mobile stack only above page size', () => {
    expect(shouldPageZoneStack(7)).toBe(false)
    expect(shouldPageZoneStack(9)).toBe(true)
  })

  it('paginates zones into fixed-size pages', () => {
    const page0 = paginateZones(LARGE_FARM_ZONES, 0, 8)
    const page1 = paginateZones(LARGE_FARM_ZONES, 1, 8)
    const page2 = paginateZones(LARGE_FARM_ZONES, 2, 8)
    expect(page0).toHaveLength(8)
    expect(page1).toHaveLength(8)
    expect(page2).toHaveLength(8)
    expect(page0[0].id).toBe(1)
    expect(page1[0].id).toBe(9)
    expect(totalZonePages(24, 8)).toBe(3)
  })

  it('offers desktop list view only above threshold', () => {
    expect(shouldOfferDesktopListView(7)).toBe(false)
    expect(shouldOfferDesktopListView(9)).toBe(true)
    expect(shouldOfferDesktopListView(24)).toBe(true)
  })

  describe('session persistence', () => {
    beforeEach(() => {
      sessionStorage.clear()
    })
    afterEach(() => {
      sessionStorage.clear()
    })

    it('defaults to "all" filter and round-trips writes', () => {
      expect(readTodayZoneFilter()).toBe('all')
      writeTodayZoneFilter('outdoor')
      expect(readTodayZoneFilter()).toBe('outdoor')
    })

    it('ignores invalid stored filter ids', () => {
      sessionStorage.setItem('gr33n_today_zone_filter', 'bogus')
      expect(readTodayZoneFilter()).toBe('all')
    })

    it('defaults to "map" view and round-trips writes', () => {
      expect(readTodayDesktopView()).toBe('map')
      writeTodayDesktopView('list')
      expect(readTodayDesktopView()).toBe('list')
    })
  })
})
