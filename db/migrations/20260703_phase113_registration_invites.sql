-- Phase 113 — registration invite codes
CREATE TABLE IF NOT EXISTS auth.registration_invites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       TEXT UNIQUE NOT NULL,
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_by    UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_registration_invites_code_active
    ON auth.registration_invites (code)
    WHERE used_at IS NULL;
