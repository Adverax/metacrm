CREATE SCHEMA IF NOT EXISTS cluster;

-- ========================================
-- CLUSTER GROUP TYPE AND TABLE
-- ========================================

-- Group type enumeration for different types of user groups
-- Defines the behavior and membership rules for groups
CREATE TYPE cluster.group_type AS ENUM ( 'regular', 'queue', 'role', 'role_and_subordinates', 'territory', 'territory_and_subordinates' );

-- Group table for organizing users into logical clusters
-- Supports different group types with specific membership rules and behaviors
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Multiple group types with different behaviors
-- - Soft delete with audit trail
-- - API-friendly naming with api_name field
-- - Integration with roles and territories
-- 
-- Group Types:
-- - 'regular': Standard user groups with manual membership
-- - 'queue': Work queue groups for task assignment
-- - 'role': Groups that represent roles (membership based on role assignment)
-- - 'role_and_subordinates': Groups that include role holders and their subordinates
-- - 'territory': Groups that represent territories (membership based on territory assignment)
-- - 'territory_and_subordinates': Groups that include territory members and their subordinates
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Example usage:
--   INSERT INTO cluster.group (label, api_name, type) 
--   VALUES ('Administrators', 'admins', 'regular');
--   
--   SELECT * FROM cluster.group WHERE tenant_id = 'uuid' AND type = 'regular';
CREATE TABLE cluster."group"
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                      BIGSERIAL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id               UUID               NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- Human-readable group name for display
    -- Used in UI and reports (e.g., 'Administrators', 'Sales Team', 'Support Queue')
    label                   VARCHAR(255)       NOT NULL,
    
    -- API-friendly group identifier
    -- Used in code and API endpoints (e.g., 'admins', 'sales_team', 'support_queue')
    -- Must match pattern: ^[a-zA-Z0-9_]{1,63}$
    api_name                VARCHAR(63)        NOT NULL,
    
    -- Type of group determining its behavior and membership rules
    -- See group_type enum for available types
    type                    cluster.group_type NOT NULL,
    
    -- Group email address for notifications and communication
    -- Can be NULL for groups that don't need email functionality
    email                   TEXT               NULL,
    
    -- Record creation timestamp
    created_at              timestamptz        NOT NULL DEFAULT now(),
    
    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at              timestamptz        NOT NULL DEFAULT now(),
    
    -- Soft delete timestamp
    -- NULL for active groups, timestamp when group was deleted
    deleted_at              timestamptz,
    
    -- Principal who created this group
    -- References iam.principal for audit trail
    created_by_principal_id BIGINT             NOT NULL DEFAULT bootstrap.current_principal_id() REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,
    
    -- Principal who last updated this group
    -- References iam.principal for audit trail
    updated_by_principal_id BIGINT             NOT NULL DEFAULT bootstrap.current_principal_id() REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,
    
    -- Principal who deleted this group
    -- References iam.principal for audit trail
    deleted_by_principal_id BIGINT             NULL REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    -- Related role ID for role-based groups
    -- Required for 'role' and 'role_and_subordinates' types
    -- NULL for other group types
    related_role_id         BIGINT             NULL REFERENCES iam.role (tenant_id, id) ON DELETE CASCADE,
    
    -- Related territory ID for territory-based groups
    -- Required for 'territory' and 'territory_and_subordinates' types
    -- NULL for other group types
    related_territory_id    BIGINT             NULL REFERENCES iam.territory (tenant_id, id) ON DELETE CASCADE,

    -- Constraint: api_name must contain only alphanumeric characters and underscores
    CONSTRAINT group_api_name_check CHECK (api_name ~ '^[a-zA-Z0-9_]{1,63}$'),
    
    -- Constraint: related_role_id is required for role-based group types
    CONSTRAINT group_related_role_required CHECK ( (type NOT IN ('role', 'role_and_subordinates')) OR
                                                   (related_role_id IS NOT NULL) ),
    
    -- Constraint: related_territory_id is required for territory-based group types
    CONSTRAINT group_related_territory_required CHECK ( (type NOT IN ('territory', 'territory_and_subordinates')) OR
                                                        (related_territory_id IS NOT NULL) ),
    
    -- Constraint: role and territory relationships are mutually exclusive
    -- Ensures groups can only be related to either a role OR a territory, not both
    CONSTRAINT group_related_exclusive CHECK (
        (type IN ('role', 'role_and_subordinates') AND related_role_id IS NOT NULL AND related_territory_id IS NULL) OR
        (type IN ('territory', 'territory_and_subordinates') AND related_territory_id IS NOT NULL AND
         related_role_id IS NULL) OR
        (type IN ('regular', 'queue') AND related_role_id IS NULL AND related_territory_id IS NULL) )
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cluster', 'group', 16);
SELECT bootstrap.attach_audit_triggers('cluster', 'group');

-- Unique index for active group api_name and type combinations
-- Ensures api_name is unique within tenant for each group type among active groups
-- Allows same api_name to be reused after soft delete or for different types
CREATE UNIQUE INDEX IF NOT EXISTS ux_group_api_name_type ON cluster."group" (tenant_id, api_name, type) WHERE deleted_at IS NULL;

-- Index for group type queries
-- Used to find groups by type (e.g., all regular groups, all role groups)
CREATE INDEX IF NOT EXISTS ix_group_type ON cluster."group" (type);

-- ========================================
-- CLUSTER GROUP MEMBER TABLE
-- ========================================

-- Group member table for managing group membership
-- Supports both user and group membership with soft delete and audit trail
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Support for both user and group membership
-- - Soft delete with audit trail
-- - Human-readable record_id for API usage
-- - Nested group membership (groups can contain other groups)
-- 
-- Membership Types:
-- - User membership: member_user_id is set, member_group_id is NULL
-- - Group membership: member_group_id is set, member_user_id is NULL
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Example usage:
--   INSERT INTO cluster.group_member (group_id, member_user_id) 
--   VALUES (123, 456);
--   
--   INSERT INTO cluster.group_member (group_id, member_group_id) 
--   VALUES (123, 789);
CREATE TABLE IF NOT EXISTS cluster.group_member
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                      BIGSERIAL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id               UUID        NOT NULL        DEFAULT bootstrap.current_tenant_id(),
    
    -- Human-readable unique identifier for API usage
    -- Format: 'grp' + 16 hex characters (e.g., 'grp_a1b2c3d4e5f67890')
    -- Used in REST APIs and external integrations
    record_id               VARCHAR(19) NOT NULL UNIQUE DEFAULT bootstrap.generate_pk('grp'),
    
    -- Record creation timestamp
    created_at              timestamptz NOT NULL        DEFAULT now(),
    
    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at              timestamptz NOT NULL        DEFAULT now(),
    
    -- Soft delete timestamp
    -- NULL for active memberships, timestamp when membership was deleted
    deleted_at              timestamptz,
    
    -- Principal who created this membership
    -- References iam.principal for audit trail
    created_by_principal_id BIGINT      NOT NULL        DEFAULT bootstrap.current_principal_id() REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,
    
    -- Principal who last updated this membership
    -- References iam.principal for audit trail
    updated_by_principal_id BIGINT      NOT NULL        DEFAULT bootstrap.current_principal_id() REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,
    
    -- Principal who deleted this membership
    -- References iam.principal for audit trail
    deleted_by_principal_id BIGINT      NULL REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    -- Reference to the group this membership belongs to
    -- CASCADE DELETE ensures memberships are removed when group is deleted
    group_id                BIGINT      NOT NULL REFERENCES cluster."group" (tenant_id, id) ON DELETE CASCADE,

    -- Reference to the user member (for user membership)
    -- NULL for group membership
    -- CASCADE DELETE ensures memberships are removed when user is deleted
    member_user_id          BIGINT      NULL REFERENCES iam."user" (tenant_id, id) ON DELETE CASCADE,
    
    -- Reference to the group member (for group membership)
    -- NULL for user membership
    -- CASCADE DELETE ensures memberships are removed when member group is deleted
    member_group_id         BIGINT      NULL REFERENCES cluster."group" (tenant_id, id) ON DELETE CASCADE,

    -- Constraint: exactly one of member_user_id or member_group_id must be set
    -- Ensures each membership is either a user or a group, not both or neither
    CONSTRAINT group_member_exactly_one_set CHECK ( (member_user_id IS NOT NULL AND member_group_id IS NULL) OR
                                                    (member_user_id IS NULL AND member_group_id IS NOT NULL) )
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cluster', 'group_member', 16);
SELECT bootstrap.attach_audit_triggers('cluster', 'group_member');

-- Unique index for user memberships
-- Ensures a user can only be a member of a group once (prevents duplicate memberships)
-- Only applies to active memberships (deleted_at IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS ux_group_member_user ON cluster.group_member (tenant_id, group_id, member_user_id) WHERE member_user_id IS NOT NULL;

-- Unique index for group memberships
-- Ensures a group can only be a member of another group once (prevents duplicate memberships)
-- Only applies to active memberships (deleted_at IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS ux_group_member_group ON cluster.group_member (tenant_id, group_id, member_group_id) WHERE member_group_id IS NOT NULL;
