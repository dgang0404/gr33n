/**
 * Phase 45 WS5 — farmer-empty shells for edge modules (animals, aquaponics).
 * @see docs/workflow-guide.md §10
 * @see docs/pattern-playbooks.md
 */

/** @typedef {'animals'|'aquaponics'} ModuleShellId */

/** @type {Record<ModuleShellId, object>} */
export const MODULE_EMPTY_SHELLS = {
  animals: {
    id: 'animals',
    icon: '🐔',
    title: 'Track flocks, herds, and pens',
    summary:
      'Animal groups record head count, species, and which zone a flock lives in. Feeding, water, and climate still run through the same zones, sensors, and tasks you use for plants.',
    bullets: [
      'Create a group (e.g. Layer flock) and link it to a coop or paddock zone.',
      'Log lifecycle events — added, born, died, sold — to build an audit trail.',
      'Livestock feed and bedding flow through Supplies and Money like other inputs.',
    ],
    workflowDoc: 'docs/workflow-guide.md',
    workflowSection: '§10 Animal husbandry & aquaponics',
    playbookDoc: 'docs/pattern-playbooks.md',
    playbookSection: 'Chicken coop (`chicken_coop_v1`)',
    templateKey: 'chicken_coop_v1',
    primaryAction: 'Create first group',
  },
  aquaponics: {
    id: 'aquaponics',
    icon: '🐟',
    title: 'Link fish tanks to grow beds',
    summary:
      'An aquaponics loop pairs a fish-tank zone with a grow-bed zone so reporting and Guardian can answer which tank feeds which bed. Pumps, sensors, and rules still live on the zones.',
    bullets: [
      'Create at least two zones — one for the fish tank, one for the grow bed.',
      'Register a loop here so the farm has one row tying tank ↔ bed.',
      'Wire return pump, air pump, and water chemistry sensors on each zone.',
    ],
    workflowDoc: 'docs/workflow-guide.md',
    workflowSection: '§10.4 Aquaponics loops',
    playbookDoc: 'docs/pattern-playbooks.md',
    playbookSection: 'Small aquaponics (`small_aquaponics_v1`)',
    templateKey: 'small_aquaponics_v1',
    primaryAction: 'Create first loop',
  },
}

/**
 * @param {ModuleShellId} moduleId
 */
export function moduleEmptyShellConfig(moduleId) {
  return MODULE_EMPTY_SHELLS[moduleId] || null
}

/**
 * @param {ModuleShellId} moduleId
 * @param {number} zoneCount
 */
export function moduleShellZoneHint(moduleId, zoneCount) {
  if (moduleId === 'aquaponics' && zoneCount < 2) {
    return {
      message: 'Create at least two zones (fish tank + grow bed) before linking a loop.',
      actionLabel: 'My zones',
      actionTo: '/zones',
    }
  }
  if (moduleId === 'animals' && zoneCount === 0) {
    return {
      message: 'Optional: add a zone for the coop or paddock before creating a group.',
      actionLabel: 'My zones',
      actionTo: '/zones',
    }
  }
  return null
}
