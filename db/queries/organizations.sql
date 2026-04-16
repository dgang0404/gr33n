-- ============================================================
-- Queries: organizations & org membership
-- ============================================================

-- name: CreateOrganization :one
INSERT INTO gr33ncore.organizations (name, plan_tier, billing_status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM gr33ncore.organizations WHERE id = $1;

-- name: UpdateOrganization :one
UPDATE gr33ncore.organizations
SET name = $2, plan_tier = $3, billing_status = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListOrganizationsForUser :many
SELECT
    o.id,
    o.name,
    o.plan_tier,
    o.billing_status,
    o.created_at,
    o.updated_at,
    m.role_in_org
FROM gr33ncore.organizations o
JOIN gr33ncore.organization_memberships m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.name ASC;

-- name: CreateOrganizationMembership :one
INSERT INTO gr33ncore.organization_memberships (organization_id, user_id, role_in_org)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrganizationMembership :one
SELECT organization_id, user_id, role_in_org, joined_at
FROM gr33ncore.organization_memberships
WHERE organization_id = $1 AND user_id = $2;

-- name: GetOrganizationUsageSummary :one
SELECT
    (SELECT COUNT(*)::bigint FROM gr33ncore.farms f
        WHERE f.organization_id = $1 AND f.deleted_at IS NULL) AS farm_count,
    (SELECT COUNT(*)::bigint FROM gr33ncore.devices d
        INNER JOIN gr33ncore.farms f ON f.id = d.farm_id
        WHERE f.organization_id = $1 AND f.deleted_at IS NULL AND d.deleted_at IS NULL) AS device_count,
    (SELECT COUNT(*)::bigint FROM gr33ncore.sensors s
        INNER JOIN gr33ncore.farms f ON f.id = s.farm_id
        WHERE f.organization_id = $1 AND f.deleted_at IS NULL AND s.deleted_at IS NULL) AS sensor_count,
    (SELECT COUNT(*)::bigint FROM gr33ncore.tasks t
        INNER JOIN gr33ncore.farms f ON f.id = t.farm_id
        WHERE f.organization_id = $1 AND f.deleted_at IS NULL AND t.deleted_at IS NULL) AS task_count,
    (SELECT COUNT(*)::bigint FROM gr33ncore.cost_transactions c
        INNER JOIN gr33ncore.farms f ON f.id = c.farm_id
        WHERE f.organization_id = $1 AND f.deleted_at IS NULL) AS cost_transaction_count;
