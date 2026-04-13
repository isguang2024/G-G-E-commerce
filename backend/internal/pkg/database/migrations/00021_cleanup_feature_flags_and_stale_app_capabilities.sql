-- +goose Up
-- +goose StatementBegin
WITH cleaned AS (
    SELECT
        id,
        (
            capabilities
                - 'managed_pages'
                - 'runtime_navigation'
                - 'app_switchable'
                #- '{auth,session_mode}'
                #- '{auth,sessionMode}'
                #- '{routing,entry_mode}'
                #- '{routing,entryMode}'
                #- '{routing,route_prefix}'
                #- '{routing,routePrefix}'
                #- '{routing,supports_public_runtime}'
                #- '{routing,supportsPublicRuntime}'
                #- '{runtime,kind}'
                #- '{runtime,supports_worktab}'
                #- '{runtime,supportsWorktab}'
                #- '{navigation,supports_multi_space}'
                #- '{navigation,supportsMultiSpace}'
                #- '{navigation,default_landing_mode}'
                #- '{navigation,defaultLandingMode}'
                #- '{navigation,supports_space_badges}'
                #- '{navigation,supportsSpaceBadges}'
                #- '{integration,supports_broadcast_channel}'
                #- '{integration,supportsBroadcastChannel}'
        ) AS next_capabilities
    FROM apps
    WHERE deleted_at IS NULL
), normalized_routing AS (
    SELECT
        id,
        CASE
            WHEN jsonb_typeof(next_capabilities -> 'routing') = 'object'
                 AND next_capabilities -> 'routing' = '{}'::jsonb
                THEN next_capabilities - 'routing'
            ELSE next_capabilities
        END AS next_capabilities
    FROM cleaned
), normalized_runtime AS (
    SELECT
        id,
        CASE
            WHEN jsonb_typeof(next_capabilities -> 'runtime') = 'object'
                 AND next_capabilities -> 'runtime' = '{}'::jsonb
                THEN next_capabilities - 'runtime'
            ELSE next_capabilities
        END AS next_capabilities
    FROM normalized_routing
), normalized_navigation AS (
    SELECT
        id,
        CASE
            WHEN jsonb_typeof(next_capabilities -> 'navigation') = 'object'
                 AND next_capabilities -> 'navigation' = '{}'::jsonb
                THEN next_capabilities - 'navigation'
            ELSE next_capabilities
        END AS next_capabilities
    FROM normalized_runtime
), normalized_integration AS (
    SELECT
        id,
        CASE
            WHEN jsonb_typeof(next_capabilities -> 'integration') = 'object'
                 AND next_capabilities -> 'integration' = '{}'::jsonb
                THEN next_capabilities - 'integration'
            ELSE next_capabilities
        END AS next_capabilities
    FROM normalized_navigation
), normalized_auth AS (
    SELECT
        id,
        CASE
            WHEN jsonb_typeof(next_capabilities -> 'auth') = 'object'
                 AND next_capabilities -> 'auth' = '{}'::jsonb
                THEN next_capabilities - 'auth'
            ELSE next_capabilities
        END AS next_capabilities
    FROM normalized_integration
)
UPDATE apps AS target
SET
    capabilities = normalized_auth.next_capabilities,
    meta = (target.meta - 'feature_flags'),
    updated_at = NOW()
FROM normalized_auth
WHERE target.id = normalized_auth.id
  AND (
      target.capabilities IS DISTINCT FROM normalized_auth.next_capabilities
      OR target.meta ? 'feature_flags'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
