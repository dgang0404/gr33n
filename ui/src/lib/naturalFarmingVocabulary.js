/**
 * Operator-facing Natural farming labels.
 * Maps to gr33nnaturalfarming.* without mirroring schema layout in the UI.
 *
 * DB table              → operator term
 * input_definitions     → Input / Inputs
 * input_batches         → Batch / Batches
 * application_recipes   → Apply recipe / Apply recipes
 */

export const NF_VOCAB = {
  input: 'Input',
  inputs: 'Inputs',
  batch: 'Batch',
  batches: 'Batches',
  applyRecipe: 'Apply recipe',
  applyRecipes: 'Apply recipes',
  fieldGuide: 'Field guide',
  makeBatch: 'Make a batch',
  readyBatches: 'Ready batches',
  manageRows: 'Inputs & batches',
}

/** Workspace tab labels (tab ids unchanged for routes/deep links). */
export const NF_WORKSPACE_TAB_LABELS = {
  batch: NF_VOCAB.makeBatch,
  library: NF_VOCAB.fieldGuide,
  recipes: NF_VOCAB.applyRecipes,
  manage: NF_VOCAB.manageRows,
  stock: NF_VOCAB.readyBatches,
}

/** Read-only canon sub-tabs inside Field guide. */
export const NF_FIELD_GUIDE_TAB_LABELS = {
  inputs: NF_VOCAB.inputs,
  application: NF_VOCAB.applyRecipes,
  programs: 'Programs',
  livestock: 'Livestock',
}

/** Farm-row editor sub-tabs inside Natural farming → Inputs & batches. */
export const NF_MANAGE_TAB_LABELS = {
  definitions: NF_VOCAB.inputs,
  batches: NF_VOCAB.batches,
}
