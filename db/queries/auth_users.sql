-- ============================================================
-- Queries: auth.users
-- ============================================================

-- name: GetAuthUserByEmail :one
SELECT id, email, password_hash, created_at FROM auth.users WHERE email = $1;

-- name: CreateAuthUser :one
INSERT INTO auth.users (email, password_hash) VALUES ($1, $2) RETURNING *;

-- name: UpdateAuthUserPasswordHash :exec
UPDATE auth.users SET password_hash = $2 WHERE id = $1;
