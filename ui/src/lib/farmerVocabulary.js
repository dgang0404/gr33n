/**
 * Phase 47 WS5 + Phase 45 WS3 — grow-path vocabulary ban list and zone labels.
 * @see docs/farmer-vocabulary.md
 */

/** Canonical grow-path labels (Vocabulary v2 — zones not rooms). */
export const GROW_PATH_ZONE_LABELS = {
  navMyZones: 'My zones',
  navMyZonesTitle: 'Grow areas — water, light, and climate per zone',
  feedingNavTitle: 'Daily feeding plan per zone — how each zone gets water',
  mobileZones: 'Zones',
  addZone: '+ Add zone',
  addZoneAction: 'Add zone',
  addFirstZone: 'Add my first zone',
  thisZone: 'This zone',
  thisZonePhrase: 'this zone',
  allZones: 'All zones',
  showAllZones: 'Show all zones',
}

/** Phase 45 WS3 — generic "room" as grow-area noun (zone display names may still contain Room). */
export const GROW_PATH_GENERIC_ROOM_BANS = [
  { id: 'my-rooms', pattern: /\bMy rooms\b/i, hint: 'Use GROW_PATH_ZONE_LABELS.navMyZones' },
  { id: 'this-room', pattern: /\bthis room\b/i, hint: 'Use "this zone"' },
  { id: 'per-room', pattern: /\bper room\b/i, hint: 'Use "per zone"' },
  { id: 'one-card-per-room', pattern: /\bone card per room\b/i, hint: 'Use "one card per zone"' },
  { id: 'add-room-cta', pattern: /\+ Add room\b/i, hint: 'Use "+ Add zone"' },
  { id: 'add-a-grow-room', pattern: /\bAdd a grow room\b/i, hint: 'Use "Add a zone"' },
  { id: 'add-grow-room', pattern: /\bAdd grow room\b/i, hint: 'Use "Add zone"' },
  { id: 'first-grow-room-label', pattern: /\bAdd my first grow room\b/i, hint: 'Use "Add my first zone"' },
  { id: 'show-all-rooms', pattern: /\bShow all rooms\b/i, hint: 'Use "Show all zones"' },
  { id: 'all-rooms-arrow', pattern: /\bAll rooms\s*→/i, hint: 'Use "All zones →"' },
  { id: 'open-my-rooms', pattern: /\bOpen My rooms\b/i, hint: 'Use "Open My zones"' },
  { id: 'each-rooms', pattern: /\beach room's\b/i, hint: 'Use "each zone\'s"' },
  { id: 'which-rooms', pattern: /\bWhich rooms\b/i, hint: 'Use "Which zones"' },
  { id: 'for-this-room', pattern: /\bfor this room\b/i, hint: 'Use "for this zone"' },
  { id: 'in-this-room', pattern: /\bin this room\b/i, hint: 'Use "in this zone"' },
  { id: 'today-in-this-room', pattern: /\bToday in this room\b/i, hint: 'Use "Today in this zone"' },
  { id: 'how-this-room-gets', pattern: /\bHow this room gets\b/i, hint: 'Use "How this zone gets"' },
  { id: 'no-rooms-yet', pattern: /\bNo rooms yet\b/i, hint: 'Use "No zones yet"' },
  { id: 'summarize-this-room', pattern: /\bSummarize this room\b/i, hint: 'Use "Summarize this zone"' },
  { id: 'mobile-rooms-nav', pattern: /label:\s*'Rooms'/i, hint: "Use label: 'Zones'" },
  { id: 'create-room', pattern: /\bCreate room\b/i, hint: 'Use "Create zone"' },
  { id: 'room-name-label', pattern: /\bRoom name\b/i, hint: 'Use "Zone name"' },
  { id: 'room-type-label', pattern: /\bRoom type\b/i, hint: 'Use "Zone type"' },
  { id: 'open-room', pattern: /\bOpen room\b/i, hint: 'Use "Open zone"' },
  { id: 'grow-rooms-phrase', pattern: /\bgrow rooms\b/i, hint: 'Use "zones"' },
  { id: 'grow-room-phrase', pattern: /\bgrow room\b/i, hint: 'Use "zone"' },
  { id: 'feeding-setup-this-room', pattern: /\bFeeding setup for this room\b/i, hint: 'Use "Feeding setup for this zone"' },
  { id: 'no-room-linked', pattern: /\bNo room linked\b/i, hint: 'Use "No zone linked"' },
  { id: 'all-rooms-label', pattern: /\bAll rooms\b/i, hint: 'Use "All zones"' },
]

/** User-visible phrases banned on grow routes (templates + farmer copy libs). */
export const GROW_PATH_VOCABULARY_BANS = [
  {
    id: 'setpoints-arrow',
    pattern: /Setpoints?\s*→/i,
    hint: 'Use "Comfort targets" or "Farm-wide bands →"',
  },
  {
    id: 'manage-fertigation',
    pattern: /Manage\s*→.*Fertigation/i,
    hint: 'Stay on Water tab or link to Feed & water hub',
  },
  {
    id: 'fertigation-program-label',
    pattern: /Fertigation program/i,
    hint: 'Use "Feeding plan"',
  },
  {
    id: 'executable-action',
    pattern: /executable_action/,
    hint: 'Use "Step in feeding plan" (Advanced only)',
  },
  {
    id: 'cron-expression-display',
    pattern: /cron_expression/,
    hint: 'Use scheduleRunsLabel() / humanizeCron()',
  },
  {
    id: 'application-recipe-id',
    pattern: /application_recipe_id/,
    hint: 'Use "Recipe" label (Advanced only)',
  },
  {
    id: 'schedules-arrow',
    pattern: /Schedules\s*→/i,
    hint: 'Use "Farm-wide timing →" or "What runs when"',
  },
  {
    id: 'automation-rule-label',
    pattern: /Automation rule/i,
    hint: 'Use "Automation"',
  },
  {
    id: 'schedule-or-rule',
    pattern: /schedule or rule/i,
    hint: 'Use "automation or feed timing"',
  },
  {
    id: 'zone-setpoints',
    pattern: /zone_setpoints/,
    hint: 'Internal only — never in farmer UI',
  },
  {
    id: 'predicate-display',
    pattern: /\bpredicate\b/i,
    hint: 'Omit — show rule sentence',
  },
  ...GROW_PATH_GENERIC_ROOM_BANS,
]

/**
 * @param {string} vueOrJsSource
 * @returns {string}
 */
export function extractGrowPathScanText(vueOrJsSource) {
  if (!vueOrJsSource.includes('<template')) return vueOrJsSource
  const template = vueOrJsSource.match(/<template[^>]*>([\s\S]*?)<\/template>/i)
  return template ? template[1] : vueOrJsSource
}

/** Matcher-safe phrases allowed on grow paths (Guardian setup-pack intent). */
export const GROW_PATH_VOCABULARY_ALLOWLIST = [
  /Add my philodendron to .+ with a light fertigation program/gi,
]

/**
 * @param {string} text
 * @param {typeof GROW_PATH_VOCABULARY_BANS} [bans]
 * @returns {Array<{ id: string, hint: string, match: string }>}
 */
export function findGrowPathVocabularyViolations(text, bans = GROW_PATH_VOCABULARY_BANS) {
  let scrubbed = text
  for (const allow of GROW_PATH_VOCABULARY_ALLOWLIST) {
    scrubbed = scrubbed.replace(allow, '')
  }
  const violations = []
  for (const ban of bans) {
    const m = scrubbed.match(ban.pattern)
    if (m) {
      violations.push({ id: ban.id, hint: ban.hint, match: m[0] })
    }
  }
  return violations
}
