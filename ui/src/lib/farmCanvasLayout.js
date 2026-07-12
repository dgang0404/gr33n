/**
 * Phase 166 WS2 — canvas layout math (normalized 0–1 coordinates).
 */

import { DEFAULT_TILE_H, DEFAULT_TILE_W } from './farmVisualStatus.js'

const NUDGE_STEP = 0.01
const MIN_SIZE = 0.08

export function clamp01(v) {
  if (!Number.isFinite(v)) return 0
  return Math.min(1, Math.max(0, v))
}

/**
 * @param {object} layout
 * @param {'left'|'right'|'up'|'down'} direction
 * @param {number} [step]
 */
export function nudgeLayout(layout, direction, step = NUDGE_STEP) {
  const next = { ...layout }
  if (direction === 'left') next.x = clamp01(next.x - step)
  if (direction === 'right') next.x = clamp01(next.x + step)
  if (direction === 'up') next.y = clamp01(next.y - step)
  if (direction === 'down') next.y = clamp01(next.y + step)
  return constrainLayout(next)
}

/**
 * Keep tile inside canvas bounds.
 * @param {object} layout
 */
export function constrainLayout(layout) {
  const w = clamp01(Number(layout.w ?? DEFAULT_TILE_W))
  const h = clamp01(Number(layout.h ?? DEFAULT_TILE_H))
  let x = clamp01(Number(layout.x ?? 0))
  let y = clamp01(Number(layout.y ?? 0))
  if (x + w > 1) x = Math.max(0, 1 - w)
  if (y + h > 1) y = Math.max(0, 1 - h)
  return { x, y, w, h }
}

/**
 * @param {object} layout normalized
 * @param {{ width: number, height: number }} canvasSize px
 */
export function layoutToStyle(layout, canvasSize) {
  const c = constrainLayout(layout)
  const w = canvasSize?.width || 1
  const h = canvasSize?.height || 1
  return {
    left: `${c.x * 100}%`,
    top: `${c.y * 100}%`,
    width: `${c.w * 100}%`,
    height: `${c.h * 100}%`,
  }
}

/**
 * @param {number} clientX
 * @param {number} clientY
 * @param {DOMRect} rect
 * @param {object} layout
 */
export function pointerDeltaToLayout(clientX, clientY, startX, startY, rect, startLayout) {
  const dx = (clientX - startX) / rect.width
  const dy = (clientY - startY) / rect.height
  return constrainLayout({
    x: Number(startLayout.x) + dx,
    y: Number(startLayout.y) + dy,
    w: startLayout.w,
    h: startLayout.h,
  })
}

/**
 * @param {object} layout
 * @param {'se'|'e'|'s'} handle
 * @param {number} dx normalized delta
 * @param {number} dy normalized delta
 */
export function resizeLayout(layout, handle, dx, dy) {
  const base = constrainLayout(layout)
  let { x, y, w, h } = base
  if (handle === 'e' || handle === 'se') w = Math.max(MIN_SIZE, w + dx)
  if (handle === 's' || handle === 'se') h = Math.max(MIN_SIZE, h + dy)
  return constrainLayout({ x, y, w, h })
}

/**
 * @param {object} solar from site weather
 * @param {Date} [now]
 */
export function sunDialProgress(solar, now = new Date()) {
  if (!solar?.sunrise_at || !solar?.sunset_at) return null
  const sunrise = new Date(solar.sunrise_at).getTime()
  const sunset = new Date(solar.sunset_at).getTime()
  const t = now.getTime()
  if (t <= sunrise) return 0
  if (t >= sunset) return 1
  return (t - sunrise) / (sunset - sunrise)
}

/**
 * @param {object} solar
 */
export function formatSunTimes(solar) {
  if (!solar?.sunrise_at || !solar?.sunset_at) return null
  const fmt = (iso) => new Date(iso).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
  return {
    sunrise: fmt(solar.sunrise_at),
    sunset: fmt(solar.sunset_at),
    daylength: solar.daylength_hours != null ? `${solar.daylength_hours} h` : null,
  }
}
