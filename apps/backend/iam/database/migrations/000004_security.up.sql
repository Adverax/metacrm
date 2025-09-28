CREATE SCHEMA IF NOT EXISTS security;

-- ========================================
-- SECURITY OBJECT TABLE
-- ========================================

-- Object table for defining security objects in the system
-- Represents entities that can have permissions applied to them (e.g., 'user', 'order', 'product')
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - API-friendly naming for programmatic access
-- - Audit trail with creation and modification tracking
-- - Foundation for object-level security (OLS)
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Example usage:
--   INSERT INTO security.object (api_name) VALUES ('user');
--   INSERT INTO security.object (api_name) VALUES ('order');
--   
--   SELECT * FROM security.object WHERE tenant_id = 'uuid' AND api_name = 'user';
CREATE TABLE security.object
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id          BIGSERIAL   NOT NULL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id   UUID        NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- API-friendly object identifier
    -- Used in code and API endpoints (e.g., 'user', 'order', 'product', 'invoice')
    -- Must match pattern: ^[_a-zA-Z][a-zA-Z0-9_]{0,62}$
    -- Can start with underscore for system objects
    api_name    VARCHAR(63) NOT NULL,
    
    -- Record creation timestamp
    created_at  timestamptz NOT NULL DEFAULT now(),
    
    -- Constraint: api_name must start with letter or underscore and contain only alphanumeric characters and underscores
    CONSTRAINT security_object_api_name_check CHECK (api_name ~ '^[_a-zA-Z][a-zA-Z0-9_]{0,62}$'),
    
    -- Unique constraint: api_name must be unique within tenant
    -- Ensures no duplicate object names within the same tenant
    UNIQUE (tenant_id, api_name)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('security', 'object', 16);

-- ========================================
-- SECURITY FIELD TABLE
-- ========================================

-- Field table for defining security fields within objects
-- Represents specific fields of objects that can have field-level permissions (FLS)
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Belongs to a specific security object
-- - API-friendly naming for programmatic access
-- - Audit trail with creation and modification tracking
-- - Foundation for field-level security (FLS)
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Example usage:
--   INSERT INTO security.field (object_id, api_name) VALUES (1, 'email');
--   INSERT INTO security.field (object_id, api_name) VALUES (1, 'salary');
--   
--   SELECT * FROM security.field WHERE tenant_id = 'uuid' AND object_id = 1;
CREATE TABLE security.field
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id          BIGSERIAL   NOT NULL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id   UUID        NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- Reference to the security object this field belongs to
    -- References security.object.id
    -- CASCADE DELETE ensures fields are removed when object is deleted
    object_id   BIGINT      NOT NULL,
    
    -- API-friendly field identifier
    -- Used in code and API endpoints (e.g., 'email', 'salary', 'ssn', 'phone')
    -- Must match pattern: ^[_a-zA-Z][a-zA-Z0-9_]{0,62}$
    -- Can start with underscore for system fields
    api_name    VARCHAR(63) NOT NULL,
    
    -- Record creation timestamp
    created_at  timestamptz NOT NULL DEFAULT now(),
    
    -- Constraint: api_name must start with letter or underscore and contain only alphanumeric characters and underscores
    CONSTRAINT security_field_api_name_check CHECK (api_name ~ '^[_a-zA-Z][a-zA-Z0-9_]{0,62}$'),

    -- Foreign key to the security object this field belongs to
    -- CASCADE DELETE ensures fields are removed when object is deleted
    CONSTRAINT security_field_object_fk FOREIGN KEY (tenant_id, object_id) REFERENCES security.object (tenant_id, id) ON DELETE CASCADE,
    
    -- Unique constraint: api_name must be unique within object and tenant
    -- Ensures no duplicate field names within the same object
    UNIQUE (tenant_id, object_id, api_name)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('security', 'field', 16);
SELECT bootstrap.attach_audit_triggers('security', 'field');

-- ========================================
-- SECURITY PERMISSION SET TABLE
-- ========================================

-- Permission set table for grouping and managing permissions
-- Represents collections of permissions that can be assigned to groups or users
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Can be associated with groups for role-based access control
-- - Soft delete with audit trail
-- - Owner-based access control
-- - API-friendly naming for programmatic access
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Example usage:
--   INSERT INTO security.permission_set (api_name, label, description, group_id) 
--   VALUES ('admin_permissions', 'Administrator Permissions', 'Full system access', 1);
--   
--   SELECT * FROM security.permission_set WHERE tenant_id = 'uuid' AND api_name = 'admin_permissions';
CREATE TABLE security.permission_set
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id            BIGSERIAL   NOT NULL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id     UUID        NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- Reference to the group this permission set belongs to (optional)
    -- References cluster.group.id
    -- CASCADE DELETE ensures permission sets are removed when group is deleted
    -- NULL for standalone permission sets not tied to specific groups
    group_id      BIGINT,
    
    -- API-friendly permission set identifier
    -- Used in code and API endpoints (e.g., 'admin_permissions', 'user_permissions', 'read_only')
    -- Must match pattern: ^[_a-zA-Z][a-zA-Z0-9_]{0,62}$
    api_name      VARCHAR(63) NOT NULL,
    
    -- Human-readable permission set name for display
    -- Used in UI and reports (e.g., 'Administrator Permissions', 'User Permissions')
    label         TEXT        NOT NULL,
    
    -- Detailed description of what this permission set allows
    -- Used for documentation and user understanding
    description   TEXT,
    
    -- Record creation timestamp
    created_at    timestamptz NOT NULL DEFAULT now(),
    
    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at    timestamptz NOT NULL DEFAULT now(),
    
    -- Soft delete timestamp
    -- NULL for active permission sets, timestamp when permission set was deleted
    deleted_at    timestamptz,
    
    -- Principal who created this permission set
    -- References iam.user for audit trail
    created_by_principal_id BIGINT      NOT NULL DEFAULT bootstrap.current_principal_id(),
    
    -- Principal who last updated this permission set
    -- References iam.user for audit trail
    updated_by_principal_id BIGINT      NOT NULL DEFAULT bootstrap.current_principal_id(),
    
    -- Principal who deleted this permission set
    -- References iam.user for audit trail
    deleted_by_principal_id BIGINT,
    
    -- Constraint: api_name must start with letter or underscore and contain only alphanumeric characters and underscores
    CONSTRAINT security_permission_set_api_name_check CHECK (api_name ~ '^[_a-zA-Z][a-zA-Z0-9_]{0,62}$')б

    -- Foreign key to the group this permission set belongs to
    -- CASCADE DELETE ensures permission sets are removed when group is deleted
    CONSTRAINT security_permission_set_group_fk FOREIGN KEY (tenant_id, group_id) REFERENCES cluster."group" (tenant_id, id) ON DELETE CASCADE,

    -- Foreign key to the principal who created this permission set
    -- CASCADE DELETE ensures permission sets are removed when principal is deleted
    CONSTRAINT security_permission_set_created_by_principal_fk FOREIGN KEY (tenant_id, created_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    -- Foreign key to the principal who last updated this permission set
    -- CASCADE DELETE ensures permission sets are removed when principal is deleted
    CONSTRAINT security_permission_set_updated_by_principal_fk FOREIGN KEY (tenant_id, updated_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    -- Foreign key to the principal who deleted this permission set
    -- CASCADE DELETE ensures permission sets are removed when principal is deleted
    CONSTRAINT security_permission_set_deleted_by_principal_fk FOREIGN KEY (tenant_id, deleted_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('security', 'permission_set', 16);
SELECT bootstrap.attach_audit_triggers('security', 'permission_set');

-- Unique index for active permission set api_name lookups
-- Ensures api_name is unique within tenant for active permission sets only
-- Allows same api_name to be reused after soft delete
CREATE UNIQUE INDEX ux_permission_set_api_name_alive ON security.permission_set (api_name) WHERE deleted_at IS NULL;

-- ========================================
-- SECURITY OBJECT PERMISSIONS TABLE
-- ========================================

-- Object permissions table for storing object-level permissions
-- Links permission sets to security objects with specific permission bitmasks
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask-based permission storage for efficient checking
-- - Object-level security (OLS) implementation
-- - Automatic cleanup when permission sets or objects are deleted
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Permission Bitmask Values:
--   1 = READ permission
--   2 = UPDATE permission  
--   4 = CREATE permission
--   8 = DELETE permission
-- 
-- Example usage:
--   INSERT INTO security.object_permissions (permission_set_id, object_id, permissions) 
--   VALUES (1, 1, 7); -- READ + UPDATE + CREATE
--   
--   SELECT * FROM security.object_permissions WHERE permission_set_id = 1 AND object_id = 1;
CREATE TABLE security.object_permissions
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                BIGSERIAL NOT NULL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id         UUID      NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- Reference to the permission set these permissions belong to
    -- References security.permission_set.id
    -- CASCADE DELETE ensures permissions are removed when permission set is deleted
    permission_set_id BIGINT    NOT NULL REFERENCES security.permission_set (id) ON DELETE CASCADE,
    
    -- Reference to the security object these permissions apply to
    -- References security.object.id
    -- CASCADE DELETE ensures permissions are removed when object is deleted
    object_id         BIGINT    NOT NULL REFERENCES security.object (id) ON DELETE CASCADE,
    
    -- Permission bitmask for this object
    -- Bit 0 (1) = READ permission
    -- Bit 1 (2) = UPDATE permission
    -- Bit 2 (4) = CREATE permission
    -- Bit 3 (8) = DELETE permission
    -- Example: 7 = READ + UPDATE + CREATE (1 + 2 + 4)
    permissions       INTEGER   NOT NULL DEFAULT 0,
    
    -- Unique constraint: one permission record per permission set + object combination
    -- Ensures no duplicate permissions for the same permission set and object
    UNIQUE (tenant_id, permission_set_id, object_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('security', 'object_permissions', 16);
SELECT bootstrap.attach_audit_triggers('security', 'object_permissions');

-- ========================================
-- SECURITY FIELD PERMISSIONS TABLE
-- ========================================

-- Field permissions table for storing field-level permissions
-- Links permission sets to security fields with specific permission bitmasks
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask-based permission storage for efficient checking
-- - Field-level security (FLS) implementation
-- - Automatic cleanup when permission sets or fields are deleted
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Permission Bitmask Values:
--   1 = READ permission
--   2 = WRITE permission  
-- 
-- Example usage:
--   INSERT INTO security.field_permissions (permission_set_id, field_id, permissions) 
--   VALUES (1, 1, 1); -- READ only for sensitive field
--   
--   SELECT * FROM security.field_permissions WHERE permission_set_id = 1 AND field_id = 1;
CREATE TABLE security.field_permissions
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                BIGSERIAL NOT NULL PRIMARY KEY,
    
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id         UUID      NOT NULL DEFAULT bootstrap.current_tenant_id(),
    
    -- Reference to the permission set these permissions belong to
    -- References security.permission_set.id
    -- CASCADE DELETE ensures permissions are removed when permission set is deleted
    permission_set_id BIGINT    NOT NULL REFERENCES security.permission_set (id) ON DELETE CASCADE,
    
    -- Reference to the security field these permissions apply to
    -- References security.field.id
    -- CASCADE DELETE ensures permissions are removed when field is deleted
    field_id          BIGINT    NOT NULL REFERENCES security.field (id) ON DELETE CASCADE,
    
    -- Permission bitmask for this field
    -- Bit 0 (1) = READ permission
    -- Bit 1 (2) = WRITE permission
    -- Example: 1 = READ only (for sensitive fields like salary, SSN)
    -- Example: 3 = READ + WRITE (for editable fields)
    permissions       INTEGER   NOT NULL DEFAULT 0,
    
    -- Unique constraint: one permission record per permission set + field combination
    -- Ensures no duplicate permissions for the same permission set and field
    UNIQUE (tenant_id, permission_set_id, field_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('security', 'field_permissions', 16);
SELECT bootstrap.attach_audit_triggers('security', 'field_permissions');

