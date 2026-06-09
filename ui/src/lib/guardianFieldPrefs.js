/**
 * Phase 67 — field assistant preferences (localStorage).
 */
const STORAGE_KEY = 'gr33n_guardian_field_prefs'

const DEFAULTS = {
  readAloud: false,
  sttProvider: 'browser', // browser | local
}

export function loadGuardianFieldPrefs() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULTS }
    return { ...DEFAULTS, ...JSON.parse(raw) }
  } catch {
    return { ...DEFAULTS }
  }
}

export function saveGuardianFieldPrefs(partial) {
  const next = { ...loadGuardianFieldPrefs(), ...partial }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  return next
}
