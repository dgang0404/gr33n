import { describe, it, expect, vi, afterEach } from 'vitest'
import {
  configSyncBadge,
  deviceLastConfigFetchAt,
  deviceUsesPlatformSync,
  CONFIG_STALE_MINUTES_DEFAULT,
} from '../lib/deviceConfigSync.js'

describe('deviceConfigSync', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('deviceUsesPlatformSync requires uid and config_version > 0', () => {
    expect(deviceUsesPlatformSync({ device_uid: 'pi-1', config_version: 1 })).toBe(true)
    expect(deviceUsesPlatformSync({ device_uid: 'pi-1', config_version: 0 })).toBe(false)
    expect(deviceUsesPlatformSync({ config_version: 2 })).toBe(false)
  })

  it('reads last_config_fetch_at from device.config object', () => {
    const device = {
      config: { last_config_fetch_at: '2026-06-08T12:00:00+00:00' },
    }
    expect(deviceLastConfigFetchAt(device)).toBe('2026-06-08T12:00:00+00:00')
  })

  it('returns null badge when platform sync not in use', () => {
    expect(configSyncBadge({ device_uid: 'x', config_version: 0 })).toBeNull()
  })

  it('shows Never fetched when sync device has no timestamp', () => {
    const badge = configSyncBadge({ device_uid: 'pi-1', config_version: 3, config: {} })
    expect(badge).toEqual({ tone: 'muted', label: 'Never fetched' })
  })

  it('shows Config synced when fetch is fresh', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-08T12:01:00Z'))
    const device = {
      device_uid: 'pi-1',
      config_version: 3,
      config: { last_config_fetch_at: '2026-06-08T12:00:30Z' },
    }
    expect(configSyncBadge(device, { pollIntervalSeconds: 30 })).toEqual({
      tone: 'ok',
      label: 'Config synced',
    })
  })

  it('shows Config stale after default stale window', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-08T12:15:00Z'))
    const device = {
      device_uid: 'pi-1',
      config_version: 3,
      config: { last_config_fetch_at: '2026-06-08T12:00:00Z' },
    }
    expect(configSyncBadge(device, { staleMinutes: CONFIG_STALE_MINUTES_DEFAULT })).toEqual({
      tone: 'warn',
      label: 'Config stale',
    })
  })
})
