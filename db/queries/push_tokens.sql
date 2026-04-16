-- ============================================================
-- Queries: gr33ncore.user_push_tokens
-- ============================================================

-- name: UpsertUserPushToken :one
INSERT INTO gr33ncore.user_push_tokens (user_id, platform, fcm_token)
VALUES ($1, $2, $3)
ON CONFLICT (fcm_token) DO UPDATE SET user_id = EXCLUDED.user_id,
  platform = EXCLUDED.platform,
  updated_at = NOW()
RETURNING *;

-- name: DeleteUserPushToken :exec
DELETE FROM gr33ncore.user_push_tokens
WHERE user_id = $1 AND fcm_token = $2;

-- name: ListPushTokensByUserID :many
SELECT * FROM gr33ncore.user_push_tokens
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: DeletePushTokenByFCMToken :exec
DELETE FROM gr33ncore.user_push_tokens WHERE fcm_token = $1;

-- name: ListFarmPushNotifyMemberUserIDs :many
SELECT m.user_id
FROM gr33ncore.farm_memberships m
WHERE m.farm_id = $1
  AND m.role_in_farm IN ('owner', 'manager', 'operator')
ORDER BY m.user_id;
