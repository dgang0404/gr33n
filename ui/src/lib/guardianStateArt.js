/** Hand-drawn Guardian state artwork — see ui/public/assets/guardian/druid/README.md */

export const GUARDIAN_STATE_ART_STATES = [
  'sleeping',
  'dormant',
  'stirring',
  'ready',
  'busy',
  'unavailable',
]

export const GUARDIAN_STATE_ART_BASE = '/assets/guardian/druid'
export const GUARDIAN_STATE_ART_MANIFEST_URL = `${GUARDIAN_STATE_ART_BASE}/manifest.json`

const STATE_ALT_LABELS = {
  sleeping: 'Guardian sleeping',
  dormant: 'Guardian resting to save power',
  stirring: 'Guardian awakening',
  ready: 'Guardian ready',
  busy: 'Guardian answering',
  unavailable: 'Guardian unavailable',
}

let manifestCache = null
let manifestPromise = null

export function resetGuardianStateArtManifestCache() {
  manifestCache = null
  manifestPromise = null
}

export async function fetchGuardianStateArtManifest() {
  if (manifestCache) return manifestCache
  if (!manifestPromise) {
    manifestPromise = fetch(GUARDIAN_STATE_ART_MANIFEST_URL)
      .then((r) => (r.ok ? r.json() : { version: 1, files: {} }))
      .catch(() => ({ version: 1, files: {} }))
      .then((manifest) => {
        manifestCache = normalizeManifest(manifest)
        return manifestCache
      })
  }
  return manifestPromise
}

export function normalizeManifest(raw) {
  const files = raw?.files && typeof raw.files === 'object' ? raw.files : {}
  const normalized = {}
  for (const state of GUARDIAN_STATE_ART_STATES) {
    const name = files[state]
    if (typeof name === 'string' && name.trim() && !name.includes('/') && !name.includes('..')) {
      normalized[state] = name.trim()
    }
  }
  return { version: raw?.version ?? 1, files: normalized }
}

export function guardianStateArtUrl(state, manifest) {
  if (!state || !manifest?.files) return null
  const file = manifest.files[state]
  if (!file) return null
  return `${GUARDIAN_STATE_ART_BASE}/${file}`
}

export function guardianStateArtAlt(state) {
  return STATE_ALT_LABELS[state] || `Guardian — ${state || 'unknown'}`
}

export function guardianStateHasArt(state, manifest) {
  return !!guardianStateArtUrl(state, manifest)
}
