/**
 * Paste-from-Maps coordinate parser.
 */
import { describe, it, expect } from 'vitest'
import { parseMapsCoordinates } from '../lib/siteWeather.js'

describe('parseMapsCoordinates', () => {
  it('parses decimal with N and W', () => {
    const r = parseMapsCoordinates('40.8938° N, 81.4055° W')
    expect(r.ok).toBe(true)
    expect(r.latitude).toBeCloseTo(40.8938, 4)
    expect(r.longitude).toBeCloseTo(-81.4055, 4)
  })

  it('parses labeled Maps copy with extra prose', () => {
    const text = [
      'Latitude: 40.8938° N (north of equator)',
      'Longitude: 81.4055° W (west of prime meridian)',
    ].join(' ')
    const r = parseMapsCoordinates(text)
    expect(r.ok).toBe(true)
    expect(r.latitude).toBeCloseTo(40.8938, 4)
    expect(r.longitude).toBeCloseTo(-81.4055, 4)
  })

  it('parses signed decimal pair without directions', () => {
    const r = parseMapsCoordinates('40.8938, -81.4055')
    expect(r.ok).toBe(true)
    expect(r.latitude).toBeCloseTo(40.8938, 4)
    expect(r.longitude).toBeCloseTo(-81.4055, 4)
  })

  it('keeps eastern hemisphere positive', () => {
    const r = parseMapsCoordinates('51.5074° N, 0.1278° E')
    expect(r.ok).toBe(true)
    expect(r.latitude).toBeCloseTo(51.5074, 4)
    expect(r.longitude).toBeCloseTo(0.1278, 4)
  })

  it('applies south as negative latitude', () => {
    const r = parseMapsCoordinates('32.0° S, 115.9° E')
    expect(r.ok).toBe(true)
    expect(r.latitude).toBeCloseTo(-32, 1)
    expect(r.longitude).toBeCloseTo(115.9, 1)
  })

  it('rejects empty paste', () => {
    const r = parseMapsCoordinates('   ')
    expect(r.ok).toBe(false)
    expect(r.error).toBeTruthy()
  })

  it('rejects prose without numbers', () => {
    const r = parseMapsCoordinates('near Akron Ohio')
    expect(r.ok).toBe(false)
  })
})
