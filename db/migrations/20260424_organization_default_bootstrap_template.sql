-- Phase 15: org-level default farm bootstrap template (optional).
-- When POST /farms omits bootstrap_template but sets organization_id, the API may apply this default.

ALTER TABLE gr33ncore.organizations
    ADD COLUMN IF NOT EXISTS default_bootstrap_template TEXT;

COMMENT ON COLUMN gr33ncore.organizations.default_bootstrap_template IS 'If set, new farms created with this organization_id and without explicit bootstrap_template use this key (e.g. jadam_indoor_photoperiod_v1).';
