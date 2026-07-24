/** Phase 211.03 — farm-scoped capability ids (mirror backend farmauthz). */
export const FARM_SCOPES = {
  admin: 'farm.admin',
  operate: 'farm.operate',
  moneyRead: 'money.costs.read',
  moneyWrite: 'money.costs.write',
  nfRead: 'nf.read',
  nfInputsWrite: 'nf.inputs.write',
  nfInputsDelete: 'nf.inputs.delete',
  nfBatchesWrite: 'nf.batches.write',
  nfBatchesDelete: 'nf.batches.delete',
  nfRecipesWrite: 'nf.recipes.write',
  nfRecipesDelete: 'nf.recipes.delete',
  nfPackApply: 'nf.pack.apply',
}

export const FARM_SCOPE_OPTIONS = [
  { id: FARM_SCOPES.admin, label: 'Farm admin' },
  { id: FARM_SCOPES.operate, label: 'Operate (zones, fertigation, tasks)' },
  { id: FARM_SCOPES.moneyRead, label: 'View costs & ledger' },
  { id: FARM_SCOPES.moneyWrite, label: 'Edit costs & unit prices' },
  { id: FARM_SCOPES.nfRead, label: 'View natural farming lists' },
  { id: FARM_SCOPES.nfInputsWrite, label: 'Create/edit inputs' },
  { id: FARM_SCOPES.nfInputsDelete, label: 'Delete inputs' },
  { id: FARM_SCOPES.nfBatchesWrite, label: 'Create/edit batches & restock' },
  { id: FARM_SCOPES.nfBatchesDelete, label: 'Delete batches' },
  { id: FARM_SCOPES.nfRecipesWrite, label: 'Create/edit recipes' },
  { id: FARM_SCOPES.nfRecipesDelete, label: 'Delete recipes' },
  { id: FARM_SCOPES.nfPackApply, label: 'Apply NF commons packs' },
]

export const ALL_FARM_SCOPES = FARM_SCOPE_OPTIONS.map((o) => o.id)
