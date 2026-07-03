-- ============================================================
-- Queries: auth.registration_invites
-- ============================================================

-- name: GetRegistrationInviteByCode :one
SELECT id, code, created_by, expires_at, used_by, used_at, created_at
FROM auth.registration_invites
WHERE code = $1;

-- name: CreateRegistrationInvite :one
INSERT INTO auth.registration_invites (code, created_by, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: MarkRegistrationInviteUsed :exec
UPDATE auth.registration_invites
SET used_by = $2, used_at = NOW()
WHERE id = $1 AND used_at IS NULL;

-- name: ListActiveRegistrationInvites :many
SELECT id, code, created_by, expires_at, used_by, used_at, created_at
FROM auth.registration_invites
WHERE used_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC
LIMIT 100;
