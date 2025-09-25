CREATE SCHEMA iam;

-- ========================================
-- IAM USER TABLE
-- ========================================

-- User table for identity and access management
-- Stores user information with tenant isolation and hierarchical management structure
--
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Hierarchical management (users can have managers)
-- - External system integration via external_id
-- - Human-readable record_id for API usage
-- - Email-based authentication and communication
--
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
--
-- Example usage:
--   INSERT INTO iam.user (name, email, manager_id)
--   VALUES ('John Doe', 'john@example.com', 123);
--
--   SELECT * FROM iam.user WHERE tenant_id = 'uuid' AND email = 'john@example.com';
CREATE TABLE iam."user"
(
    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id         uuid         NOT NULL DEFAULT bootstrap.current_tenant_id(),

    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                bigserial    NOT NULL,

    -- Human-readable unique identifier for API usage
    -- Format: 'usr' + 16 hex characters (e.g., 'usr_a1b2c3d4e5f67890')
    -- Used in REST APIs and external integrations
    record_id         varchar(19)  NOT NULL DEFAULT bootstrap.generate_pk('usr'),

    -- External system identifier for integration
    -- Links user to external systems (LDAP, Active Directory, etc.)
    -- Can be NULL for users created directly in the system
    external_id       varchar(255),

    -- User's display name
    -- Used in UI and reports
    "name"            varchar(255) NOT NULL,

    -- Record creation timestamp
    created_at        timestamptz  NOT NULL DEFAULT now(),

    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at        timestamptz  NOT NULL DEFAULT now(),

    -- User's email address
    -- Used for authentication, notifications, and communication
    -- Must be unique within tenant
    email             varchar(255) NOT NULL,

    -- Manager's user ID for hierarchical organization
    -- Creates organizational hierarchy (user -> manager -> director, etc.)
    -- NULL for top-level users (no manager)
    manager_id        bigint       NULL,

    -- Primary key combining tenant_id and id for partitioning support
    CONSTRAINT user_pk PRIMARY KEY (tenant_id, id),

    -- Foreign key to manager (self-reference)
    -- ON DELETE SET NULL ensures hierarchy integrity when manager is deleted
    CONSTRAINT user_manager_fk FOREIGN KEY (tenant_id, manager_id) REFERENCES iam."user" (tenant_id, id) ON DELETE SET NULL
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('iam', 'user', 16);

-- Index for searching users by name within tenant
-- Used for user search functionality and name-based lookups
CREATE INDEX ON iam."user" (tenant_id, name);

-- Unique index for external system integration
-- Ensures external_id is unique within tenant for system integration
CREATE UNIQUE INDEX ON iam."user" (tenant_id, external_id);

-- Unique index for API record_id lookups
-- Ensures record_id is unique within tenant for API operations
CREATE UNIQUE INDEX ON iam."user" (tenant_id, record_id);

-- Unique index for email-based authentication
-- Ensures email is unique within tenant for login and communication
CREATE UNIQUE INDEX ON iam."user" (tenant_id, email);

-- ========================================
-- IAM ROLE TABLE
-- ========================================

-- Role table for role-based access control (RBAC)
-- Stores role definitions with hierarchical structure and soft delete support
--
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Hierarchical roles (roles can have parent roles)
-- - Soft delete with audit trail
-- - External system integration via external_id
-- - API-friendly naming with api_name field
--
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
--
-- Example usage:
--   INSERT INTO iam.role (label, api_name, parent_id)
--   VALUES ('Administrator', 'admin', NULL);
--
--   SELECT * FROM iam.role WHERE tenant_id = 'uuid' AND api_name = 'admin';
CREATE TABLE IF NOT EXISTS iam.role
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                      BIGSERIAL PRIMARY KEY,

    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id               UUID         NOT NULL DEFAULT bootstrap.current_tenant_id(),

    -- External system identifier for integration
    -- Links role to external systems (LDAP, Active Directory, etc.)
    -- Can be NULL for roles created directly in the system
    external_id             VARCHAR(255),

    -- Human-readable role name for display
    -- Used in UI and reports
    label                   VARCHAR(255) NOT NULL,

    -- API-friendly role identifier
    -- Used in code and API endpoints (e.g., 'admin', 'user', 'manager')
    -- Must match pattern: ^[a-zA-Z][a-zA-Z0-9_]{0,59}$
    api_name                VARCHAR(60)  NOT NULL,

    -- Record creation timestamp
    created_at              timestamptz  NOT NULL DEFAULT now(),

    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at              timestamptz  NOT NULL DEFAULT now(),

    -- Soft delete timestamp
    -- NULL for active roles, timestamp when role was deleted
    deleted_at              timestamptz,

    -- Principal who created this role
    -- References iam.user for audit trail
    created_by_principal_id BIGINT       NOT NULL DEFAULT bootstrap.current_principal_id(),

    -- Principal who last updated this role
    -- References iam.user for audit trail
    updated_by_principal_id BIGINT       NOT NULL DEFAULT bootstrap.current_principal_id(),

    -- Principal who deleted this role
    -- References iam.user for audit trail
    deleted_by_principal_id BIGINT       NULL,

    -- Parent role ID for hierarchical roles
    -- Creates role hierarchy (e.g., 'admin' -> 'super_admin')
    -- NULL for top-level roles
    parent_id               BIGINT       NULL REFERENCES iam.role (id) ON DELETE CASCADE,

    -- Constraint: api_name must start with letter and contain only alphanumeric characters and underscores
    CONSTRAINT iam_user_role_api_name_check CHECK (api_name ~ '^[a-zA-Z][a-zA-Z0-9_]{0,59}$'),

    -- Constraint: deleted_by_principal_id must be set when role is deleted
    CONSTRAINT iam_role_deleted_by CHECK ((deleted_at IS NULL) = (deleted_by_principal_id IS NULL)),

    CONSTRAINT iam_role_created_by_principal_fk FOREIGN KEY (tenant_id, created_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    CONSTRAINT iam_role_updated_by_principal_fk FOREIGN KEY (tenant_id, updated_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    CONSTRAINT iam_role_deleted_by_principal_fk FOREIGN KEY (tenant_id, deleted_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT

) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('iam', 'role', 16);

-- Index for role hierarchy queries
-- Used to find child roles of a parent role
CREATE INDEX ix_user_role_parent ON iam.role (tenant_id, parent_id);

-- Unique index for active role api_name lookups
-- Ensures api_name is unique within tenant for active roles only
-- Allows same api_name to be reused after soft delete
CREATE UNIQUE INDEX user_role_api_name_alive ON iam.role (tenant_id, api_name) WHERE deleted_at IS NULL;

SELECT bootstrap.attach_audit_triggers('iam', 'role');

-- ========================================
-- IAM TERRITORY TABLE
-- ========================================

-- Territory table for geographical and organizational territory management
-- Stores territory definitions with hierarchical structure and soft delete support
--
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Hierarchical territories (territories can have parent territories)
-- - Soft delete with audit trail
-- - API-friendly naming with api_name field
-- - Geographical and organizational territory support
--
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
--
-- Example usage:
--   INSERT INTO iam.territory (label, api_name, parent_id)
--   VALUES ('North America', 'north_america', NULL);
--
--   SELECT * FROM iam.territory WHERE tenant_id = 'uuid' AND api_name = 'north_america';
CREATE TABLE IF NOT EXISTS iam.territory
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id                      BIGSERIAL PRIMARY KEY,

    -- Tenant identifier for multi-tenant isolation
    -- Automatically set from session context
    tenant_id               UUID         NOT NULL DEFAULT bootstrap.current_tenant_id(),

    -- Human-readable territory name for display
    -- Used in UI and reports (e.g., 'North America', 'Europe', 'Sales Region A')
    label                   VARCHAR(255) NOT NULL,

    -- API-friendly territory identifier
    -- Used in code and API endpoints (e.g., 'north_america', 'europe', 'sales_region_a')
    -- Must match pattern: ^[a-zA-Z][a-zA-Z0-9_]{0,59}$
    api_name                VARCHAR(60)  NOT NULL,

    -- Record creation timestamp
    created_at              timestamptz  NOT NULL DEFAULT now(),

    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at              timestamptz  NOT NULL DEFAULT now(),

    -- Soft delete timestamp
    -- NULL for active territories, timestamp when territory was deleted
    deleted_at              timestamptz,

    -- Principal who created this territory
    -- References iam.user for audit trail
    created_by_principal_id BIGINT       NOT NULL DEFAULT bootstrap.current_principal_id(),

    -- Principal who last updated this territory
    -- References iam.user for audit trail
    updated_by_principal_id BIGINT       NOT NULL DEFAULT bootstrap.current_principal_id(),

    -- Principal who deleted this territory
    -- References iam.user for audit trail
    deleted_by_principal_id BIGINT       NULL,

    -- Parent territory ID for hierarchical territories
    -- Creates territory hierarchy (e.g., 'North America' -> 'United States' -> 'California')
    -- NULL for top-level territories
    parent_id               BIGINT       NULL,

    -- Constraint: api_name must start with letter and contain only alphanumeric characters and underscores
    CONSTRAINT iam_territory_api_name_check CHECK (api_name ~ '^[a-zA-Z][a-zA-Z0-9_]{0,59}$'),

    -- Constraint: deleted_by_principal_id must be set when territory is deleted
    CONSTRAINT iam_territory_deleter_deleted_by CHECK ((deleted_at IS NULL) = (deleted_by_principal_id IS NULL)),

    -- Foreign keys
    CONSTRAINT iam_territory_created_by_principal_fk FOREIGN KEY (tenant_id, created_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    CONSTRAINT iam_territory_updated_by_principal_fk FOREIGN KEY (tenant_id, updated_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    CONSTRAINT iam_territory_deleted_by_principal_fk FOREIGN KEY (tenant_id, deleted_by_principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE RESTRICT,

    CONSTRAINT iam_territory_parent_fk FOREIGN KEY (tenant_id, parent_id) REFERENCES iam.territory (tenant_id, id) ON DELETE CASCADE

    -- TODO: Add territory type field for different territory types (geographical, organizational, etc.)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('iam', 'territory', 16);

SELECT bootstrap.attach_audit_triggers('iam', 'territory');

-- Index for territory hierarchy queries
-- Used to find child territories of a parent territory
CREATE INDEX ix_territory_parent ON iam.territory (tenant_id, parent_id);

-- Unique index for active territory api_name lookups
-- Ensures api_name is unique within tenant for active territories only
-- Allows same api_name to be reused after soft delete
CREATE UNIQUE INDEX territory_api_name_alive ON iam.territory (tenant_id, api_name) WHERE deleted_at IS NULL;

-- ========================================
-- IAM PRINCIPAL TYPE AND TABLE
-- ========================================

-- Principal kind enumeration for different types of principals
-- Defines the types of entities that can authenticate and perform actions
CREATE TYPE iam.principal_kind AS ENUM ('user','service','external','system');

-- Principal table for unified authentication and authorization
-- Stores authentication credentials and metadata for all types of principals
--
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Support for different principal types (user, service, external, system)
-- - Unified login system for all principal types
-- - Active/inactive status management
-- - Audit trail with timestamps
--
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
--
-- Example usage:
--   INSERT INTO iam.principal (kind, subject_id, login)
--   VALUES ('user', 123, 'john.doe@company.com');
--
--   SELECT * FROM iam.principal WHERE tenant_id = 'uuid' AND login = 'john.doe@company.com';
CREATE TABLE IF NOT EXISTS iam.principal
(
    -- Internal sequential ID for database operations
    -- Used for foreign key relationships and internal references
    id           bigserial          NOT NULL,

    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id    uuid               NOT NULL,

    -- Type of principal (user, service, external, system)
    -- Determines how the principal authenticates and what permissions it has
    kind         iam.principal_kind NOT NULL,

    -- Reference to the subject entity
    -- For 'user' kind: references iam.user.id
    -- For other kinds: NULL (service accounts, external systems, etc.)
    subject_id   bigint,

    -- Login identifier for authentication
    -- Can be email, username, API key, or other identifier
    -- Must be unique within tenant
    login        text               NOT NULL,

    -- Active status flag
    -- false = principal is disabled and cannot authenticate
    -- true = principal is active and can authenticate
    is_active    boolean            NOT NULL DEFAULT true,

    -- Record creation timestamp
    created_at   timestamptz        NOT NULL DEFAULT now(),

    -- Last modification timestamp
    -- Automatically updated by audit triggers
    updated_at   timestamptz        NOT NULL DEFAULT now(),

    -- Primary key combining tenant_id and id for partitioning support
    PRIMARY KEY (tenant_id, id),

    -- Foreign key to user table for user principals
    -- DEFERRABLE INITIALLY DEFERRED allows user creation in same transaction
    CONSTRAINT principal_user_fk FOREIGN KEY (tenant_id, subject_id) REFERENCES iam."user" (tenant_id, id) DEFERRABLE INITIALLY DEFERRED,

    -- Constraint: subject_id is required only for 'user' kind principals
    -- Other kinds (service, external, system) should have subject_id = NULL
    CONSTRAINT principal_subject_required_chk
        CHECK ( (kind <> 'user' AND subject_id IS NULL)
            OR (kind  = 'user' AND subject_id IS NOT NULL) )
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('iam', 'principal', 16);
SELECT bootstrap.attach_audit_triggers('iam', 'principal');

-- Index for principal-subject relationship queries
-- Used to find principals by subject_id (e.g., find all principals for a user)
CREATE INDEX IF NOT EXISTS principal_tenant_subject_idx
    ON iam.principal (tenant_id, subject_id);

-- ========================================
-- IAM IDENTITY TYPE AND TABLE
-- ========================================

-- Identity kind enumeration for different authentication methods
-- Defines the types of authentication credentials that can be used
CREATE TYPE iam.identity_kind AS ENUM ('password','api_key','oauth');

-- Identity table for storing authentication credentials
-- Links principals to their authentication methods and external identity providers
--
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Support for multiple authentication methods per principal
-- - Integration with external identity providers (IdP)
-- - Flexible subject mapping for different IdP systems
--
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
--
-- Example usage:
--   INSERT INTO iam.identity (tenant_id, principal_id, kind, idp, subject)
--   VALUES ('uuid', 123, 'password', 'local', 'john.doe@company.com');
--
--   SELECT * FROM iam.identity WHERE tenant_id = 'uuid' AND idp = 'google' AND subject = 'google_user_id';
CREATE TABLE IF NOT EXISTS iam.identity
(
    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id    uuid   NOT NULL,

    -- Reference to the principal this identity belongs to
    -- References iam.principal.id
    principal_id bigint NOT NULL,

    -- Type of authentication method
    -- 'password' = local password authentication
    -- 'api_key' = API key authentication
    -- 'oauth' = OAuth-based authentication (Google, Microsoft, etc.)
    kind         iam.identity_kind NOT NULL DEFAULT 'password'::iam.identity_kind,

    -- Identity Provider (IdP) identifier
    -- 'local' = internal authentication system
    -- 'google' = Google OAuth
    -- 'microsoft' = Microsoft OAuth
    -- 'ldap' = LDAP server
    -- 'saml' = SAML provider
    idp          text   NOT NULL,

    -- Subject identifier within the IdP
    -- For 'password' + 'local': email address or username
    -- For 'oauth' + 'google': Google user ID
    -- For 'oauth' + 'microsoft': Microsoft user ID
    -- For 'ldap': LDAP distinguished name (DN)
    subject      text   NOT NULL,

    -- Unique constraint: same IdP + subject combination can only exist once per tenant
    -- Prevents duplicate identities across principals
    UNIQUE (tenant_id, idp, subject),

    -- Foreign key to principal table
    -- CASCADE DELETE ensures identities are removed when principal is deleted
    FOREIGN KEY (tenant_id, principal_id) REFERENCES iam.principal (tenant_id, id) ON DELETE CASCADE
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('iam', 'identity', 16);
SELECT bootstrap.attach_audit_triggers('iam', 'identity');

