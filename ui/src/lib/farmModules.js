/** Phase 115 — farm_active_modules schema names (match backend). */
export const MODULE_SCHEMA = {
  crops: 'gr33ncrops',
  naturalFarming: 'gr33nnaturalfarming',
  animals: 'gr33nanimals',
  aquaponics: 'gr33naquaponics',
}

/** Nav paths gated by module schema. */
export const MODULE_NAV = {
  [MODULE_SCHEMA.animals]: '/animals',
  [MODULE_SCHEMA.aquaponics]: '/aquaponics',
}

export function moduleMapFromRows(rows) {
  const map = {}
  for (const row of rows || []) {
    if (row?.module_schema_name) {
      map[row.module_schema_name] = !!row.is_enabled
    }
  }
  return map
}

/** When modules are not loaded yet, keep legacy behavior (show optional modules). */
export function isModuleEnabled(modules, schema, defaultWhenUnknown = true) {
  if (!modules || Object.keys(modules).length === 0) return defaultWhenUnknown
  if (modules[schema] === undefined) return defaultWhenUnknown
  return !!modules[schema]
}
