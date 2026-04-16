-- Phase 14 WS5: FCM device registration for farm alert push (Capacitor; web later)

CREATE TABLE IF NOT EXISTS gr33ncore.user_push_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    platform   TEXT NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
    fcm_token  TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_push_tokens_user_id
    ON gr33ncore.user_push_tokens (user_id);

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_user_push_tokens_updated_at') THEN
    CREATE TRIGGER trg_user_push_tokens_updated_at
      BEFORE UPDATE ON gr33ncore.user_push_tokens
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;
