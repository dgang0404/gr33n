-- ============================================================
-- Queries: gr33ncore.profiles
-- ============================================================

-- name: GetProfileByUserID :one
SELECT user_id, full_name, email, avatar_url, role, preferences,
       hourly_rate, hourly_rate_currency, created_at, updated_at
FROM gr33ncore.profiles
WHERE user_id = $1;

-- name: GetProfileByEmail :one
SELECT user_id, full_name, email, avatar_url, role, preferences,
       hourly_rate, hourly_rate_currency, created_at, updated_at
FROM gr33ncore.profiles
WHERE email = $1;

-- name: CreateProfile :one
INSERT INTO gr33ncore.profiles (user_id, full_name, email, avatar_url, role, preferences, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
RETURNING *;

-- name: UpdateProfile :one
UPDATE gr33ncore.profiles
SET full_name = $2, avatar_url = $3, role = $4, preferences = $5, updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: UpdateProfileHourlyRate :one
-- Phase 20.9 WS1 — operator-set default wage. NULL clears the rate
-- (and the autologger will skip cost rows for logs with no
-- snapshot).
UPDATE gr33ncore.profiles
SET hourly_rate = sqlc.narg('hourly_rate')::numeric,
    hourly_rate_currency = sqlc.narg('hourly_rate_currency')::char(3),
    updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: AddFarmMember :one
INSERT INTO gr33ncore.farm_memberships (farm_id, user_id, role_in_farm, permissions, joined_at)
VALUES ($1, $2, $3, $4, NOW())
RETURNING *;

-- name: GetFarmMembership :one
SELECT farm_id, user_id, role_in_farm, permissions, joined_at
FROM gr33ncore.farm_memberships
WHERE farm_id = $1 AND user_id = $2;

-- name: GetFarmMembers :many
SELECT p.user_id, p.full_name, p.email, p.avatar_url, m.role_in_farm, m.permissions, m.joined_at
FROM gr33ncore.farm_memberships m
JOIN gr33ncore.profiles p ON p.user_id = m.user_id
WHERE m.farm_id = $1
ORDER BY m.joined_at ASC;

-- name: UpdateFarmMemberRole :one
UPDATE gr33ncore.farm_memberships
SET role_in_farm = $3
WHERE farm_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveFarmMember :exec
DELETE FROM gr33ncore.farm_memberships
WHERE farm_id = $1 AND user_id = $2;
