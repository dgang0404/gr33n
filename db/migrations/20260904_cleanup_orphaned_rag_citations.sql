-- Phase 218 WS0: clean up orphaned RAG citations (rows that reference deleted source entities).
-- These chunks have been accumulating since citations were introduced (Phase 152) because the
-- delete handlers never called DeleteRagChunksByFarmSource. Now that cleanup is wired in, this
-- migration removes historical orphans and ensures the citation_route_audit test passes.

BEGIN;

DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1
  AND source_type = 'alert_notification'
  AND source_id NOT IN (
    SELECT id FROM gr33ncore.alerts_notifications WHERE farm_id = 1
  );

DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1
  AND source_type = 'automation_rule'
  AND source_id NOT IN (
    SELECT id FROM gr33ncore.automation_rules WHERE farm_id = 1
  );

DELETE FROM gr33ncore.rag_embedding_chunks
WHERE farm_id = 1
  AND source_type = 'crop_cycle'
  AND source_id NOT IN (
    SELECT id FROM gr33nfertigation.crop_cycles WHERE farm_id = 1
  );

COMMIT;
