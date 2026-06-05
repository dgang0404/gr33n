/**
 * Phase 47 WS5 — grow-path vocabulary ban list.
 * @see docs/farmer-vocabulary.md
 */

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

/**
 * @param {string} text
 * @param {typeof GROW_PATH_VOCABULARY_BANS} [bans]
 * @returns {Array<{ id: string, hint: string, match: string }>}
 */
export function findGrowPathVocabularyViolations(text, bans = GROW_PATH_VOCABULARY_BANS) {
  const violations = []
  for (const ban of bans) {
    const m = text.match(ban.pattern)
    if (m) {
      violations.push({ id: ban.id, hint: ban.hint, match: m[0] })
    }
  }
  return violations
}
