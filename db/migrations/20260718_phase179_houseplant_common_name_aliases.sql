-- Phase 179 — houseplant common-name aliases (philodendron, pothos, monstera,
-- ficus). The Guardian "add my philodendron to {zone}" setup-pack demo phrase
-- (docs/operator-tour.md, ui/src/lib/guardianStarters.js,
-- internal/farmguardian/proposals_setup_pack.go) has referenced these names
-- since Phase 32, but the houseplant catalog entry never carried alias rows
-- for them, so the rule-assisted matcher could never resolve a crop_key and
-- silently produced zero proposals. See data/crop_library.yaml `houseplant`.
INSERT INTO gr33ncrops.crop_catalog_aliases (alias, crop_key)
VALUES
    ('philodendron', 'houseplant'),
    ('pothos', 'houseplant'),
    ('monstera', 'houseplant'),
    ('ficus', 'houseplant')
ON CONFLICT (alias) DO NOTHING;
