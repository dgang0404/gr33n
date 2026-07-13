/**
 * Phase 34 WS5 — prefill chat input when operator taps Refine on a proposal card.
 */
export function refinePrefillForProposal(proposal) {
  const summary = String(proposal?.summary || 'this change').trim()
  return `Please revise this change request — ${summary}. Correction: `
}
