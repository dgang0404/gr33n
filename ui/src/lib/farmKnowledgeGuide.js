/**
 * Help glossary — farm knowledge layers and federation boundaries.
 * @see docs/workflow-guide.md §11a
 */

/** @typedef {{ layer: string, job: string, examples: string }} FarmKnowledgeLayer */

/** @type {FarmKnowledgeLayer[]} */
export const FARM_KNOWLEDGE_LAYERS = [
  {
    layer: 'Executable data',
    job: 'Things you operate on',
    examples: 'Recipes, fertigation programs, inventory, zones',
  },
  {
    layer: 'Readable corpus (RAG)',
    job: 'Guardian cites & Help searches by meaning',
    examples: 'Field guides, platform docs, task/run history',
  },
  {
    layer: 'Reference catalogs',
    job: 'Structured lookup (not semantic search)',
    examples: 'Symptom guide by crop/category',
  },
  {
    layer: 'Federation',
    job: 'Optional sharing outside this database',
    examples: 'Commons Catalog packs; Insert Commons stats',
  },
]

/** @typedef {{ want: string, use: string, crossInstall: string }} FarmKnowledgeAction */

/** @type {FarmKnowledgeAction[]} */
export const FARM_KNOWLEDGE_ACTIONS = [
  {
    want: 'Copy recipe packs onto another farm on this server',
    use: 'Help → Import → Browse Catalog → Import to Farm',
    crossInstall: 'Same server only',
  },
  {
    want: 'Copy a pack to a different gr33n install',
    use: 'Publish from Farm → export JSON → import on the other install',
    crossInstall: 'Manual hand-off',
  },
  {
    want: 'Guardian cites install / crop-care docs',
    use: 'Field guides + re-ingest in Settings → Field memories',
    crossInstall: 'No — per farm on this install',
  },
  {
    want: 'Search past tasks or mixing in plain language',
    use: 'Help → Search (operational RAG)',
    crossInstall: 'No',
  },
  {
    want: 'Diagnose yellow leaves on a crop',
    use: 'Help → Symptom guide',
    crossInstall: 'No',
  },
  {
    want: 'Share anonymized cost/task rollups with a hub',
    use: 'Settings → Insert Commons → Run sync',
    crossInstall: 'Yes — live HTTP to a receiver',
  },
]

/** @typedef {{ category: string, moves: string, how: string }} FarmKnowledgeBoundary */

/** @type {FarmKnowledgeBoundary[]} */
export const FARM_KNOWLEDGE_BOUNDARIES = [
  {
    category: 'Commons Catalog packs',
    moves: 'Yes, manually',
    how: 'Export JSON → import on other install → Import to Farm',
  },
  {
    category: 'Insert Commons aggregates',
    moves: 'Yes, live',
    how: 'Each farm syncs to the same receiver URL',
  },
  {
    category: 'Field guides, platform docs, operational RAG, symptoms',
    moves: 'No',
    how: 'Each install ingests its own copy per farm',
  },
]

/** Short operator-facing explanation of Insert Commons coarse stats. */
export const INSERT_COMMONS_SUMMARY = {
  title: 'Insert Commons — outbound coarse stats (not knowledge import)',
  lead:
    'Insert Commons does not send recipes, field guides, chat, zone names, GPS, or receipts. When you opt in and Run sync, the API rolls up numbers from your farm database and POSTs a fixed JSON shape to an optional receiver URL.',
  includes: [
    'Farm pseudonym (opaque id — not your farm name)',
    'Coarse profile: scale tier, timezone bucket, currency, status',
    'Cost totals and per-category income/expense counts (no receipt text)',
    'Task counts by status (not task titles)',
    'Device counts by status (not hostnames)',
  ],
  excludes:
    'PII, location text, sensor readings, alerts, Guardian chat, RAG chunks, and recipe text are never included.',
  notFor:
    'Use Commons Catalog import for recipes. Insert Commons is for optional community benchmarks — e.g. aggregate cost bands across farms that choose to share.',
}
