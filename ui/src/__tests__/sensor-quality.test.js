import { describe, it, expect } from 'vitest'
import {
  classifySensorQuality,
  resolveSensitivity,
  idealRangeFromCropStage,
  QUALITY_LABEL,
} from '../lib/sensorQuality.js'

describe('sensorQuality (Phase 217 WS4/WS5/WS6)', () => {
  describe('classifySensorQuality', () => {
    const range = { low: 10, high: 20 } // target defaults to 15

    it('returns unknown for a missing value', () => {
      expect(classifySensorQuality(null, range)).toBe('unknown')
      expect(classifySensorQuality(undefined, range)).toBe('unknown')
    })

    it('returns unknown for a missing/invalid range', () => {
      expect(classifySensorQuality(15, {})).toBe('unknown')
      expect(classifySensorQuality(15, { low: 20, high: 10 })).toBe('unknown')
    })

    it('classifies dead-center-of-range as awesome', () => {
      expect(classifySensorQuality(15, range)).toBe('awesome')
    })

    it('classifies near-target as good', () => {
      expect(classifySensorQuality(12.5, range)).toBe('good')
    })

    it('classifies near-edge-but-inside-range as mid', () => {
      expect(classifySensorQuality(10.5, range)).toBe('mid')
    })

    it('classifies just outside range as bad', () => {
      expect(classifySensorQuality(9, range)).toBe('bad')
      expect(classifySensorQuality(21, range)).toBe('bad')
    })

    it('classifies far outside range as critical', () => {
      expect(classifySensorQuality(0, range)).toBe('critical')
      expect(classifySensorQuality(100, range)).toBe('critical')
    })

    it('respects an explicit target off-center', () => {
      const skewed = { low: 10, high: 20, target: 18 }
      expect(classifySensorQuality(18, skewed)).toBe('awesome')
    })

    it('narrow sensitivity shrinks the good/awesome bands', () => {
      // 12.5 is "good" at normal sensitivity but falls to "mid" when narrowed.
      expect(classifySensorQuality(12.5, range, 'normal')).toBe('good')
      expect(classifySensorQuality(12.5, range, 'narrow')).toBe('mid')
    })

    it('wide sensitivity widens the good/awesome bands', () => {
      // distance-from-target 3 falls outside the normal good band (2.5) but
      // inside the widened one (2.5 * 1.3 = 3.25).
      expect(classifySensorQuality(12, range, 'normal')).toBe('mid')
      expect(classifySensorQuality(12, range, 'wide')).toBe('good')
    })
  })

  describe('resolveSensitivity', () => {
    it('prefers a per-sensor override', () => {
      expect(resolveSensitivity({ quality_sensitivity: 'narrow' }, { meta_data: { default_sensor_sensitivity: 'wide' } })).toBe('narrow')
    })

    it('falls back to the farm-wide default', () => {
      expect(resolveSensitivity({}, { meta_data: { default_sensor_sensitivity: 'wide' } })).toBe('wide')
    })

    it('falls back to normal when neither is set', () => {
      expect(resolveSensitivity({}, {})).toBe('normal')
      expect(resolveSensitivity(null, null)).toBe('normal')
    })
  })

  describe('idealRangeFromCropStage', () => {
    it('maps ec sensor_type to ec_min/ec_target/ec_max', () => {
      const stage = { ec_min: 1.0, ec_target: 1.4, ec_max: 1.6 }
      expect(idealRangeFromCropStage('ec', stage)).toEqual({ low: 1.0, high: 1.6, target: 1.4 })
    })

    it('maps temperature sensor_type to temp_min_c/temp_max_c with no target', () => {
      const stage = { temp_min_c: 22, temp_max_c: 26 }
      expect(idealRangeFromCropStage('temperature', stage)).toEqual({ low: 22, high: 26, target: undefined })
    })

    it('returns null for an unmapped sensor_type', () => {
      expect(idealRangeFromCropStage('par', { ec_min: 1, ec_max: 2 })).toBeNull()
    })

    it('returns null when the stage row is missing', () => {
      expect(idealRangeFromCropStage('ec', null)).toBeNull()
    })

    it('returns null when the mapped fields are absent from the stage row', () => {
      expect(idealRangeFromCropStage('ph', { ec_min: 1, ec_max: 2 })).toBeNull()
    })
  })

  describe('QUALITY_LABEL', () => {
    it('has a plain-language label for every tier', () => {
      expect(QUALITY_LABEL.critical).toBe('Critical')
      expect(QUALITY_LABEL.awesome).toBe('Awesome')
    })
  })
})
