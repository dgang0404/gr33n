-- Phase 13 WS4: Optional bookkeeping metadata on cost rows (invoice/receipt refs, counterparty)
-- Reversible: DROP COLUMN on document_type, document_reference, counterparty.

ALTER TABLE gr33ncore.cost_transactions ADD COLUMN IF NOT EXISTS document_type TEXT,
    ADD COLUMN IF NOT EXISTS document_reference TEXT,
    ADD COLUMN IF NOT EXISTS counterparty TEXT;
