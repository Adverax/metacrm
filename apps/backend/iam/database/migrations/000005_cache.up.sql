CREATE SCHEMA IF NOT EXISTS cache;

-- ========================================
-- HYBRID CACHE FOR FLS/OLS (FIELD/OBJECT LEVEL SECURITY)
-- ========================================

-- User object permissions cache table for basic object-level permissions
-- Caches computed permissions for user-object combinations to avoid expensive real-time calculations
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask storage for efficient permission checking
-- - TTL-based expiration for cache freshness
-- - Automatic cleanup of expired entries
-- - Always cached for frequently accessed objects
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Permission Bitmask Values (based on security.object_permissions):
--   1 = READ permission
--   2 = UPDATE permission  
--   4 = CREATE permission
--   8 = DELETE permission
--   Combinations: 7 = READ+UPDATE+CREATE, 15 = ALL permissions
-- 
-- Example usage:
--   INSERT INTO cache.user_object_permissions (tenant_id, user_id, object_id, base_permissions, expires_at) 
--   VALUES ('uuid', 123, 456, 7, now() + interval '1 hour'); -- READ+UPDATE+CREATE
--   
--   SELECT base_permissions FROM cache.user_object_permissions 
--   WHERE tenant_id = 'uuid' AND user_id = 123 AND object_id = 456 AND expires_at > now();
CREATE TABLE cache.user_object_permissions (
    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id        UUID        NOT NULL,
    
    -- Reference to the user
    -- References iam.user.id
    user_id          BIGINT      NOT NULL,
    
    -- Reference to the object
    -- References security.object.id
    object_id        BIGINT      NOT NULL,
    
    -- Base permissions as bitmask
    -- Combines all permissions from roles, groups, and direct assignments
    -- 1=READ, 2=UPDATE, 4=CREATE, 8=DELETE
    base_permissions INTEGER     NOT NULL,
    
    -- Timestamp when the data was cached
    -- Used for cache age tracking and debugging
    cached_at        timestamptz NOT NULL DEFAULT now(),
    
    -- Timestamp when the cache entry expires
    -- After this time, the entry is considered stale and should be refreshed
    expires_at       timestamptz NOT NULL,
    
    -- Primary key combining tenant_id, user_id, and object_id for partitioning support
    PRIMARY KEY (tenant_id, user_id, object_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cache', 'user_object_permissions', 16);

-- Index for fast user permission lookups
-- Used to find all permissions for a specific user across all objects
CREATE INDEX ON cache.user_object_permissions (expires_at) WHERE expires_at < now();
CREATE INDEX ON cache.user_object_permissions (tenant_id, user_id, expires_at) WHERE expires_at > now();

-- User field restrictions cache table for field-level security (FLS)
-- Caches field-specific permission restrictions to implement fine-grained access control
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask storage for efficient restriction checking
-- - TTL-based expiration for cache freshness
-- - Lazy loading - only cached when field restrictions are needed
-- - Automatic cleanup of expired entries
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Restriction Bitmask Values:
--   1 = READ restriction (field is hidden)
--   2 = WRITE restriction (field is read-only)
--   Combinations: 3 = READ+WRITE restrictions (field is completely restricted)
-- 
-- Example usage:
--   INSERT INTO cache.user_field_restrictions (tenant_id, user_id, object_id, field_id, restriction, expires_at) 
--   VALUES ('uuid', 123, 456, 789, 2, now() + interval '1 hour'); -- WRITE restriction (read-only field)
--   
--   SELECT restriction FROM cache.user_field_restrictions 
--   WHERE tenant_id = 'uuid' AND user_id = 123 AND object_id = 456 AND field_id = 789 AND expires_at > now();
CREATE TABLE cache.user_field_restrictions (
    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id    UUID        NOT NULL,
    
    -- Reference to the user
    -- References iam.user.id
    user_id      BIGINT      NOT NULL,
    
    -- Reference to the object
    -- References security.object.id
    object_id    BIGINT      NOT NULL,
    
    -- Reference to the field
    -- References security.field.id
    field_id     BIGINT      NOT NULL,
    
    -- Field restrictions as bitmask
    -- Defines which operations are restricted on this specific field
    -- 1=READ restriction, 2=WRITE restriction
    restriction  INTEGER     NOT NULL,
    
    -- Timestamp when the data was cached
    -- Used for cache age tracking and debugging
    cached_at    timestamptz NOT NULL DEFAULT now(),
    
    -- Timestamp when the cache entry expires
    -- After this time, the entry is considered stale and should be refreshed
    expires_at   timestamptz NOT NULL,
    
    -- Primary key combining tenant_id, user_id, object_id, and field_id for partitioning support
    PRIMARY KEY (tenant_id, user_id, object_id, field_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cache', 'user_field_restrictions', 16);

-- Index for fast user and object field restriction lookups
-- Used to find all field restrictions for a specific user-object combination
CREATE INDEX ON cache.user_field_restrictions (tenant_id, user_id, object_id, expires_at) WHERE expires_at > now();

-- User row permissions cache table for object-level security (OLS)
-- Caches row-specific permissions to implement data-level access control
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask storage for efficient permission checking
-- - TTL-based expiration for cache freshness
-- - Lazy loading - only cached when row-level permissions are needed
-- - Automatic cleanup of expired entries
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Permission Bitmask Values:
--   1 = READ permission for this specific row
--   2 = CREATE permission for this specific row
--   4 = UPDATE permission for this specific row
--   8 = DELETE permission for this specific row
--   Combinations: 7 = READ+CREATE+UPDATE, 15 = ALL permissions
-- 
-- Example usage:
--   INSERT INTO cache.user_row_permissions (tenant_id, user_id, object_id, row_id, permissions, expires_at) 
--   VALUES ('uuid', 123, 456, 999, 3, now() + interval '30 minutes'); -- READ+CREATE for specific row
--   
--   SELECT permissions FROM cache.user_row_permissions 
--   WHERE tenant_id = 'uuid' AND user_id = 123 AND object_id = 456 AND row_id = 999 AND expires_at > now();
CREATE TABLE cache.user_row_permissions (
    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id    UUID        NOT NULL,
    
    -- Reference to the user
    -- References iam.user.id
    user_id      BIGINT      NOT NULL,
    
    -- Reference to the object type
    -- References security.object.id
    object_id    BIGINT      NOT NULL,
    
    -- Reference to the specific row/record
    -- References the primary key of the actual data table (e.g., customer.id, order.id)
    row_id       BIGINT      NOT NULL,
    
    -- Row-specific permissions as bitmask
    -- Defines which operations are allowed on this specific row
    -- 1=READ, 2=CREATE, 4=UPDATE, 8=DELETE
    permissions  INTEGER     NOT NULL,
    
    -- Timestamp when the data was cached
    -- Used for cache age tracking and debugging
    cached_at    timestamptz NOT NULL DEFAULT now(),
    
    -- Timestamp when the cache entry expires
    -- After this time, the entry is considered stale and should be refreshed
    expires_at   timestamptz NOT NULL,
    
    -- Primary key combining tenant_id, user_id, object_id, and row_id for partitioning support
    PRIMARY KEY (tenant_id, user_id, object_id, row_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cache', 'user_row_permissions', 16);

-- Index for fast user and object row permission lookups
-- Used to find all row permissions for a specific user-object combination
CREATE INDEX ON cache.user_row_permissions (tenant_id, user_id, object_id, expires_at) WHERE expires_at > now();

-- Group object permissions cache table for group-level permissions optimization
-- Caches computed permissions for group-object combinations to optimize user permission calculations
-- 
-- Key Features:
-- - Multi-tenant architecture with tenant_id partitioning
-- - Bitmask storage for efficient permission checking
-- - TTL-based expiration for cache freshness
-- - Used to optimize user permission calculations by pre-computing group permissions
-- - Automatic cleanup of expired entries
-- 
-- Partitioning: HASH partitioning by tenant_id for performance and isolation
-- 
-- Permission Bitmask Values (based on security.object_permissions):
--   1 = READ permission
--   2 = UPDATE permission  
--   4 = CREATE permission
--   8 = DELETE permission
--   Combinations: 7 = READ+UPDATE+CREATE, 15 = ALL permissions
-- 
-- Example usage:
--   INSERT INTO cache.group_object_permissions (tenant_id, group_id, object_id, base_permissions, expires_at) 
--   VALUES ('uuid', 789, 456, 15, now() + interval '2 hours'); -- ALL permissions for group
--   
--   SELECT base_permissions FROM cache.group_object_permissions 
--   WHERE tenant_id = 'uuid' AND group_id = 789 AND object_id = 456 AND expires_at > now();
CREATE TABLE cache.group_object_permissions (
    -- Tenant identifier for multi-tenant isolation
    -- Must be explicitly set (no default)
    tenant_id        UUID        NOT NULL,
    
    -- Reference to the group
    -- References cluster.group.id
    group_id         BIGINT      NOT NULL,
    
    -- Reference to the object
    -- References security.object.id
    object_id        BIGINT      NOT NULL,
    
    -- Base permissions as bitmask
    -- Combines all permissions assigned to this group for this object
    -- 1=READ, 2=CREATE, 4=UPDATE, 8=DELETE
    base_permissions INTEGER     NOT NULL,
    
    -- Timestamp when the data was cached
    -- Used for cache age tracking and debugging
    cached_at        timestamptz NOT NULL DEFAULT now(),
    
    -- Timestamp when the cache entry expires
    -- After this time, the entry is considered stale and should be refreshed
    expires_at       timestamptz NOT NULL,
    
    -- Primary key combining tenant_id, group_id, and object_id for partitioning support
    PRIMARY KEY (tenant_id, group_id, object_id)
) PARTITION BY HASH (tenant_id);

SELECT bootstrap.make_partitions('cache', 'group_object_permissions', 16);

-- Index for fast group permission lookups
-- Used to find all permissions for a specific group across all objects
CREATE INDEX ON cache.group_object_permissions (tenant_id, group_id, expires_at) WHERE expires_at > now();

-- ========================================
-- FAST DATA EXTRACTION FUNCTIONS
-- ========================================

-- Get object permissions with automatic caching
-- Retrieves user permissions for a specific object with automatic cache management
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_user_id: User ID to check permissions for
--   p_object_id: Object ID to check permissions for
--   p_ttl_seconds: Cache TTL in seconds (default: 3600 = 1 hour)
-- 
-- Returns: INTEGER - Permission bitmask (1=READ, 2=CREATE, 4=UPDATE, 8=DELETE)
-- 
-- Examples:
--   SELECT cache.get_object_permissions('uuid', 123, 456); -- Default 1 hour TTL
--   SELECT cache.get_object_permissions('uuid', 123, 456, 1800); -- 30 minutes TTL
--
-- Usage in applications:
--   -- Check if user can read an object
--   IF (cache.get_object_permissions(tenant_id, user_id, object_id) & 1) > 0 THEN
--     -- User has READ permission
--   END IF;
CREATE OR REPLACE FUNCTION cache.get_object_permissions(
    p_tenant_id UUID, 
    p_user_id BIGINT, 
    p_object_id BIGINT,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS INTEGER
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    cached_permissions INTEGER;
    computed_permissions INTEGER;
BEGIN
    -- Try to get from cache
    SELECT base_permissions INTO cached_permissions
    FROM cache.user_object_permissions
    WHERE tenant_id = p_tenant_id 
      AND user_id = p_user_id 
      AND object_id = p_object_id
      AND expires_at > now();
    
    -- If not in cache, compute and cache
    IF cached_permissions IS NULL THEN
        -- Compute permissions based on user groups
        SELECT COALESCE(bit_or(gop.base_permissions), 0) INTO computed_permissions
        FROM cluster.group_member gm
        JOIN cache.group_object_permissions gop ON (
            gop.tenant_id = p_tenant_id 
            AND gop.group_id = gm.group_id 
            AND gop.object_id = p_object_id
            AND gop.expires_at > now()
        )
        WHERE gm.tenant_id = p_tenant_id 
          AND gm.member_user_id = p_user_id
          AND gm.deleted_at IS NULL;
        
        -- If no group permissions, compute individual permissions
        IF computed_permissions IS NULL OR computed_permissions = 0 THEN
            RETURN 0; -- without caching
        END IF;
        
        -- Cache the result
        INSERT INTO cache.user_object_permissions (tenant_id, user_id, object_id, base_permissions, expires_at)
        VALUES (p_tenant_id, p_user_id, p_object_id, computed_permissions, now() + (p_ttl_seconds || ' seconds')::interval)
        ON CONFLICT (tenant_id, user_id, object_id) 
        DO UPDATE SET 
            base_permissions = EXCLUDED.base_permissions,
            cached_at = now(),
            expires_at = EXCLUDED.expires_at;
        
        cached_permissions := computed_permissions;
    END IF;
    
    RETURN cached_permissions;
END;
$$;

-- Get field permissions with restrictions (FLS - Field Level Security)
-- Retrieves user permissions for a specific field with field-level restrictions applied
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_user_id: User ID to check permissions for
--   p_object_id: Object ID to check permissions for
--   p_field_id: Field ID to check permissions for
--   p_ttl_seconds: Cache TTL in seconds (default: 3600 = 1 hour)
-- 
-- Returns: INTEGER - Effective permission bitmask after applying field restrictions
-- 
-- Examples:
--   SELECT cache.get_field_permissions('uuid', 123, 456, 789); -- Check field permissions
--   SELECT cache.get_field_permissions('uuid', 123, 456, 789, 1800); -- 30 minutes TTL
--
-- Usage in applications:
--   -- Check if user can read a specific field
--   IF (cache.get_field_permissions(tenant_id, user_id, object_id, field_id) & 1) > 0 THEN
--     -- User can read this field
--   END IF;
CREATE OR REPLACE FUNCTION cache.get_field_permissions(
    p_tenant_id UUID, 
    p_user_id BIGINT, 
    p_object_id BIGINT, 
    p_field_id BIGINT,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS INTEGER
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    base_permissions INTEGER;
    field_restriction INTEGER;
    result INTEGER;
BEGIN
    -- Get basic object permissions
    base_permissions := cache.get_object_permissions(p_tenant_id, p_user_id, p_object_id, p_ttl_seconds);
    
    -- Check field restrictions
    SELECT restriction INTO field_restriction
    FROM cache.user_field_restrictions
    WHERE tenant_id = p_tenant_id 
      AND user_id = p_user_id 
      AND object_id = p_object_id
      AND field_id = p_field_id
      AND expires_at > now();
    
    -- Apply restrictions
    result := base_permissions;
    IF field_restriction IS NOT NULL THEN
        result := result & field_restriction;
    END IF;
    
    RETURN result;
END;
$$;

-- Get row permissions (OLS - Object Level Security)
-- Retrieves user permissions for a specific row/record with automatic cache management
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_user_id: User ID to check permissions for
--   p_object_id: Object type ID to check permissions for
--   p_row_id: Specific row/record ID to check permissions for
--   p_ttl_seconds: Cache TTL in seconds (default: 3600 = 1 hour)
-- 
-- Returns: INTEGER - Permission bitmask for the specific row (1=READ, 2=UPDATE, 4=CREATE, 8=DELETE)
-- 
-- Examples:
--   SELECT cache.get_row_permissions('uuid', 123, 456, 999); -- Check row permissions
--   SELECT cache.get_row_permissions('uuid', 123, 456, 999, 1800); -- 30 minutes TTL
--
-- Usage in applications:
--   -- Check if user can read a specific customer record
--   IF (cache.get_row_permissions(tenant_id, user_id, customer_object_id, customer_id) & 1) > 0 THEN
--     -- User can read this customer record
--   END IF;
CREATE OR REPLACE FUNCTION cache.get_row_permissions(
    p_tenant_id UUID, 
    p_user_id BIGINT, 
    p_object_id BIGINT, 
    p_row_id BIGINT,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS INTEGER
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    cached_permissions INTEGER;
    base_permissions INTEGER;
    result INTEGER;
BEGIN
    -- Try to get from cache
    SELECT permissions INTO cached_permissions
    FROM cache.user_row_permissions
    WHERE tenant_id = p_tenant_id 
      AND user_id = p_user_id 
      AND object_id = p_object_id
      AND row_id = p_row_id
      AND expires_at > now();
    
    -- If not in cache, compute
    IF cached_permissions IS NULL THEN
        -- Get basic object permissions
        base_permissions := cache.get_object_permissions(p_tenant_id, p_user_id, p_object_id, p_ttl_seconds);
        
        -- Here should be logic for computing permissions for specific row
        -- For now return basic permissions
        result := base_permissions;
        
        -- Cache the result
        INSERT INTO cache.user_row_permissions (tenant_id, user_id, object_id, row_id, permissions, expires_at)
        VALUES (p_tenant_id, p_user_id, p_object_id, p_row_id, result, now() + (p_ttl_seconds || ' seconds')::interval)
        ON CONFLICT (tenant_id, user_id, object_id, row_id) 
        DO UPDATE SET 
            permissions = EXCLUDED.permissions,
            cached_at = now(),
            expires_at = EXCLUDED.expires_at;
        
        cached_permissions := result;
    END IF;
    
    RETURN cached_permissions;
END;
$$;

-- Fast permission check with automatic level detection
-- Performs permission check at the appropriate level (object, field, or row) with automatic cache management
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_user_id: User ID to check permissions for
--   p_object_id: Object ID to check permissions for
--   p_field_id: Field ID (optional) - enables field-level security check
--   p_row_id: Row ID (optional) - enables row-level security check
--   p_required_permission: Required permission bitmask (1=READ, 2=CREATE, 4=UPDATE, 8=DELETE)
--   p_ttl_seconds: Cache TTL in seconds (default: 3600 = 1 hour)
-- 
-- Returns: BOOLEAN - true if user has the required permission, false otherwise
-- 
-- Examples:
--   SELECT cache.has_permission('uuid', 123, 456); -- Object-level READ check
--   SELECT cache.has_permission('uuid', 123, 456, NULL, NULL, 2); -- Object-level CREATE check
--   SELECT cache.has_permission('uuid', 123, 456, 789); -- Field-level READ check
--   SELECT cache.has_permission('uuid', 123, 456, NULL, 999); -- Row-level READ check
--
-- Usage in applications:
--   -- Check if user can read a customer record
--   IF cache.has_permission(tenant_id, user_id, customer_object_id, NULL, customer_id, 1) THEN
--     -- User can read this customer record
--   END IF;
CREATE OR REPLACE FUNCTION cache.has_permission(
    p_tenant_id UUID, 
    p_user_id BIGINT, 
    p_object_id BIGINT, 
    p_field_id BIGINT DEFAULT NULL,
    p_row_id BIGINT DEFAULT NULL,
    p_required_permission INTEGER DEFAULT 1,  -- 1 = READ, 2 = UPDATE, 4 = CREATE, 8 = DELETE
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS BOOLEAN
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    actual_permissions INTEGER;
BEGIN
    -- Determine check level
    IF p_row_id IS NOT NULL THEN
        -- Row-level check (OLS)
        actual_permissions := cache.get_row_permissions(p_tenant_id, p_user_id, p_object_id, p_row_id, p_ttl_seconds);
    ELSIF p_field_id IS NOT NULL THEN
        -- Field-level check (FLS)
        actual_permissions := cache.get_field_permissions(p_tenant_id, p_user_id, p_object_id, p_field_id, p_ttl_seconds);
    ELSE
        -- Object-level check
        actual_permissions := cache.get_object_permissions(p_tenant_id, p_user_id, p_object_id, p_ttl_seconds);
    END IF;
    
    -- Check for required permission
    RETURN (actual_permissions & p_required_permission) = p_required_permission;
END;
$$;

-- ========================================
-- CACHE MANAGEMENT FUNCTIONS
-- ========================================

-- Invalidate user permissions cache
-- Removes all cached permissions for a specific user to force recalculation
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_user_id: User ID to invalidate cache for
-- 
-- Returns: void
-- 
-- Examples:
--   SELECT cache.invalidate_user_permissions_cache('uuid', 123);
--
-- Usage in applications:
--   -- After user role changes, invalidate their cache
--   SELECT cache.invalidate_user_permissions_cache(tenant_id, user_id);
CREATE OR REPLACE FUNCTION cache.invalidate_user_permissions_cache(p_tenant_id UUID, p_user_id BIGINT)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM cache.user_object_permissions WHERE tenant_id = p_tenant_id AND user_id = p_user_id;
    DELETE FROM cache.user_field_restrictions WHERE tenant_id = p_tenant_id AND user_id = p_user_id;
    DELETE FROM cache.user_row_permissions WHERE tenant_id = p_tenant_id AND user_id = p_user_id;
END;
$$;

-- Invalidate group permissions cache
-- Removes all cached permissions for a specific group and all its members
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_group_id: Group ID to invalidate cache for
-- 
-- Returns: void
-- 
-- Examples:
--   SELECT cache.invalidate_group_permissions_cache('uuid', 789);
--
-- Usage in applications:
--   -- After group permissions change, invalidate group and member caches
--   SELECT cache.invalidate_group_permissions_cache(tenant_id, group_id);
CREATE OR REPLACE FUNCTION cache.invalidate_group_permissions_cache(p_tenant_id UUID, p_group_id BIGINT)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM cache.group_object_permissions WHERE tenant_id = p_tenant_id AND group_id = p_group_id;
    
    -- Invalidate cache for all users in this group
    DELETE FROM cache.user_object_permissions 
    WHERE tenant_id = p_tenant_id 
      AND user_id IN (
          SELECT member_user_id 
          FROM cluster.group_member 
          WHERE tenant_id = p_tenant_id 
            AND group_id = p_group_id 
            AND member_user_id IS NOT NULL
            AND deleted_at IS NULL
      );
END;
$$;

-- Invalidate object permissions cache
-- Removes all cached permissions for a specific object across all users and groups
-- 
-- Parameters:
--   p_tenant_id: Tenant identifier
--   p_object_id: Object ID to invalidate cache for
-- 
-- Returns: void
-- 
-- Examples:
--   SELECT cache.invalidate_object_permissions_cache('uuid', 456);
--
-- Usage in applications:
--   -- After object permissions change, invalidate all related caches
--   SELECT cache.invalidate_object_permissions_cache(tenant_id, object_id);
CREATE OR REPLACE FUNCTION cache.invalidate_object_permissions_cache(p_tenant_id UUID, p_object_id BIGINT)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM cache.user_object_permissions WHERE tenant_id = p_tenant_id AND object_id = p_object_id;
    DELETE FROM cache.user_field_restrictions WHERE tenant_id = p_tenant_id AND object_id = p_object_id;
    DELETE FROM cache.user_row_permissions WHERE tenant_id = p_tenant_id AND object_id = p_object_id;
    DELETE FROM cache.group_object_permissions WHERE tenant_id = p_tenant_id AND object_id = p_object_id;
END;
$$;

-- Cleanup expired permissions cache
-- Removes all expired cache entries from all cache tables to free up storage space
-- 
-- Parameters: None
-- 
-- Returns: void
-- 
-- Examples:
--   SELECT cache.cleanup_expired_permissions_cache();
--
-- Usage in applications:
--   -- Run periodically (e.g., via cron job) to clean up expired cache
--   SELECT cache.cleanup_expired_permissions_cache();
CREATE OR REPLACE FUNCTION cache.cleanup_expired_permissions_cache()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM cache.user_object_permissions WHERE expires_at < now();
    DELETE FROM cache.user_field_restrictions WHERE expires_at < now();
    DELETE FROM cache.user_row_permissions WHERE expires_at < now();
    DELETE FROM cache.group_object_permissions WHERE expires_at < now();
    DELETE FROM cache.user_cache WHERE expires_at < now();
    DELETE FROM cache.role_cache WHERE expires_at < now();
    DELETE FROM cache.group_cache WHERE expires_at < now();
END;
$$;

-- ========================================
-- TRIGGERS FOR CACHE UPDATES
-- ========================================

-- Function to send event to outbox
CREATE OR REPLACE FUNCTION cache.send_cache_invalidation_event(
    p_tenant_id UUID,
    p_aggregate_type TEXT,
    p_aggregate_id TEXT,
    p_event_type TEXT
)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO bootstrap.outbox (aggregate_type, aggregate_id, event_type, payload, headers)
    VALUES (
        p_aggregate_type,
        p_aggregate_id,
        p_event_type,
        '{}'::jsonb,  -- Empty payload - only the fact of change
        jsonb_build_object('tenant_id', p_tenant_id)
    );
END;
$$;

-- Trigger for cache update when user changes
CREATE OR REPLACE FUNCTION cache.trigger_user_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Invalidate user cache
    PERFORM cache.invalidate_user_permissions_cache(NEW.tenant_id, NEW.id);
    
    -- Send event to outbox
    PERFORM cache.send_cache_invalidation_event(
        NEW.tenant_id,
        'user',
        NEW.id::text,
        'iam.user_permissions_invalidated'
    );
    
    RETURN NEW;
END;
$$;

-- Trigger for cache update when group changes
CREATE OR REPLACE FUNCTION cache.trigger_group_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Invalidate group cache
    PERFORM cache.invalidate_group_permissions_cache(NEW.tenant_id, NEW.id);
    
    -- Send event to outbox
    PERFORM cache.send_cache_invalidation_event(
        NEW.tenant_id,
        'group',
        NEW.id::text,
        'iam.group_permissions_invalidated'
    );
    
    RETURN NEW;
END;
$$;

-- Trigger for cache update when group member changes
CREATE OR REPLACE FUNCTION cache.trigger_group_member_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Invalidate user cache
    IF NEW.member_user_id IS NOT NULL THEN
        PERFORM cache.invalidate_user_permissions_cache(NEW.tenant_id, NEW.member_user_id);
        
        -- Send event to outbox
        PERFORM cache.send_cache_invalidation_event(
            NEW.tenant_id,
            'user',
            NEW.member_user_id::text,
            'iam.user_group_membership_changed'
        );
    END IF;
    
    RETURN NEW;
END;
$$;

-- Trigger for cache update when permissions change
CREATE OR REPLACE FUNCTION cache.trigger_permissions_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Invalidate object cache
    PERFORM cache.invalidate_object_permissions_cache(NEW.tenant_id, NEW.object_id);
    
    -- Send event to outbox
    PERFORM cache.send_cache_invalidation_event(
        NEW.tenant_id,
        'object',
        NEW.object_id::text,
        'iam.object_permissions_changed'
    );
    
    RETURN NEW;
END;
$$;

-- Trigger for cache update when field permissions change
CREATE OR REPLACE FUNCTION cache.trigger_field_permissions_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Invalidate object cache (since field permissions changed)
    PERFORM cache.invalidate_object_permissions_cache(NEW.tenant_id, NEW.object_id);
    
    -- Send event to outbox
    PERFORM cache.send_cache_invalidation_event(
        NEW.tenant_id,
        'object',
        NEW.object_id::text,
        'iam.field_permissions_changed'
    );
    
    RETURN NEW;
END;
$$;

-- ========================================
-- TRIGGER BINDING TO TABLES
-- ========================================

-- Triggers for users
CREATE TRIGGER trg_user_cache_invalidation
    AFTER UPDATE ON iam.user
    FOR EACH ROW
    EXECUTE FUNCTION cache.trigger_user_cache_invalidation();

-- Triggers for groups
CREATE TRIGGER trg_group_cache_invalidation
    AFTER UPDATE ON cluster.group
    FOR EACH ROW
    EXECUTE FUNCTION cache.trigger_group_cache_invalidation();

-- Triggers for group members
CREATE TRIGGER trg_group_member_cache_invalidation
    AFTER INSERT OR UPDATE OR DELETE ON cluster.group_member
    FOR EACH ROW
    EXECUTE FUNCTION cache.trigger_group_member_cache_invalidation();

-- Triggers for permissions
CREATE TRIGGER trg_object_permissions_cache_invalidation
    AFTER INSERT OR UPDATE OR DELETE ON security.object_permissions
    FOR EACH ROW
    EXECUTE FUNCTION cache.trigger_permissions_cache_invalidation();

-- Triggers for field permissions
CREATE TRIGGER trg_field_permissions_cache_invalidation
    AFTER INSERT OR UPDATE OR DELETE ON security.field_permissions
    FOR EACH ROW
    EXECUTE FUNCTION cache.trigger_field_permissions_cache_invalidation();

-- ========================================
-- ADDITIONAL FUNCTIONS
-- ========================================



-- Function to invalidate entire tenant cache
CREATE OR REPLACE FUNCTION cache.invalidate_tenant_cache(p_tenant_id UUID)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    DELETE FROM cache.user_object_permissions WHERE tenant_id = p_tenant_id;
    DELETE FROM cache.user_field_restrictions WHERE tenant_id = p_tenant_id;
    DELETE FROM cache.user_row_permissions WHERE tenant_id = p_tenant_id;
    DELETE FROM cache.group_object_permissions WHERE tenant_id = p_tenant_id;
END;
$$;

-- ========================================
-- TESTING AND MONITORING FUNCTIONS
-- ========================================

-- Function to get cache statistics
CREATE OR REPLACE FUNCTION cache.get_cache_stats(p_tenant_id UUID DEFAULT NULL)
RETURNS TABLE (
    cache_table TEXT,
    total_records BIGINT,
    expired_records BIGINT,
    active_records BIGINT,
    cache_hit_ratio NUMERIC
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'user_object_permissions'::TEXT as cache_table,
        COUNT(*) as total_records,
        COUNT(*) FILTER (WHERE expires_at < now()) as expired_records,
        COUNT(*) FILTER (WHERE expires_at > now()) as active_records,
        CASE 
            WHEN COUNT(*) > 0 THEN 
                ROUND((COUNT(*) FILTER (WHERE expires_at > now())::NUMERIC / COUNT(*)::NUMERIC) * 100, 2)
            ELSE 0 
        END as cache_hit_ratio
    FROM cache.user_object_permissions
    WHERE p_tenant_id IS NULL OR tenant_id = p_tenant_id
    
    UNION ALL
    
    SELECT 
        'user_field_restrictions'::TEXT,
        COUNT(*),
        COUNT(*) FILTER (WHERE expires_at < now()),
        COUNT(*) FILTER (WHERE expires_at > now()),
        CASE 
            WHEN COUNT(*) > 0 THEN 
                ROUND((COUNT(*) FILTER (WHERE expires_at > now())::NUMERIC / COUNT(*)::NUMERIC) * 100, 2)
            ELSE 0 
        END
    FROM cache.user_field_restrictions
    WHERE p_tenant_id IS NULL OR tenant_id = p_tenant_id
    
    UNION ALL
    
    SELECT 
        'user_row_permissions'::TEXT,
        COUNT(*),
        COUNT(*) FILTER (WHERE expires_at < now()),
        COUNT(*) FILTER (WHERE expires_at > now()),
        CASE 
            WHEN COUNT(*) > 0 THEN 
                ROUND((COUNT(*) FILTER (WHERE expires_at > now())::NUMERIC / COUNT(*)::NUMERIC) * 100, 2)
            ELSE 0 
        END
    FROM cache.user_row_permissions
    WHERE p_tenant_id IS NULL OR tenant_id = p_tenant_id
    
    UNION ALL
    
    SELECT 
        'group_object_permissions'::TEXT,
        COUNT(*),
        COUNT(*) FILTER (WHERE expires_at < now()),
        COUNT(*) FILTER (WHERE expires_at > now()),
        CASE 
            WHEN COUNT(*) > 0 THEN 
                ROUND((COUNT(*) FILTER (WHERE expires_at > now())::NUMERIC / COUNT(*)::NUMERIC) * 100, 2)
            ELSE 0 
        END
    FROM cache.group_object_permissions
    WHERE p_tenant_id IS NULL OR tenant_id = p_tenant_id;
END;
$$;

-- Function to benchmark cache performance
CREATE OR REPLACE FUNCTION cache.benchmark_permissions_check(
    p_tenant_id UUID,
    p_user_id BIGINT,
    p_object_id BIGINT,
    p_field_id BIGINT DEFAULT NULL,
    p_row_id BIGINT DEFAULT NULL,
    p_iterations INTEGER DEFAULT 1000
)
RETURNS TABLE (
    operation TEXT,
    avg_time_ms NUMERIC,
    min_time_ms NUMERIC,
    max_time_ms NUMERIC,
    total_time_ms NUMERIC
)
LANGUAGE plpgsql
AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
    i INTEGER;
    times NUMERIC[];
    current_time NUMERIC;
BEGIN
    -- Test object permissions check
    times := ARRAY[]::NUMERIC[];
    FOR i IN 1..p_iterations LOOP
        start_time := clock_timestamp();
        PERFORM cache.get_object_permissions(p_tenant_id, p_user_id, p_object_id);
        end_time := clock_timestamp();
        current_time := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
        times := array_append(times, current_time);
    END LOOP;
    
    RETURN QUERY SELECT 
        'object_permissions'::TEXT,
        ROUND(avg(t), 3),
        ROUND(min(t), 3),
        ROUND(max(t), 3),
        ROUND(sum(t), 3)
    FROM unnest(times) AS t;
    
    -- Test field permissions check (if specified)
    IF p_field_id IS NOT NULL THEN
        times := ARRAY[]::NUMERIC[];
        FOR i IN 1..p_iterations LOOP
            start_time := clock_timestamp();
            PERFORM cache.get_field_permissions(p_tenant_id, p_user_id, p_object_id, p_field_id);
            end_time := clock_timestamp();
            current_time := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
            times := array_append(times, current_time);
        END LOOP;
        
        RETURN QUERY SELECT 
            'field_permissions'::TEXT,
            ROUND(avg(t), 3),
            ROUND(min(t), 3),
            ROUND(max(t), 3),
            ROUND(sum(t), 3)
        FROM unnest(times) AS t;
    END IF;
    
    -- Test row permissions check (if specified)
    IF p_row_id IS NOT NULL THEN
        times := ARRAY[]::NUMERIC[];
        FOR i IN 1..p_iterations LOOP
            start_time := clock_timestamp();
            PERFORM cache.get_row_permissions(p_tenant_id, p_user_id, p_object_id, p_row_id);
            end_time := clock_timestamp();
            current_time := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
            times := array_append(times, current_time);
        END LOOP;
        
        RETURN QUERY SELECT 
            'row_permissions'::TEXT,
            ROUND(avg(t), 3),
            ROUND(min(t), 3),
            ROUND(max(t), 3),
            ROUND(sum(t), 3)
        FROM unnest(times) AS t;
    END IF;
END;
$$;

-- Function to pre-warm cache
CREATE OR REPLACE FUNCTION cache.warmup_permissions_cache(
    p_tenant_id UUID,
    p_user_ids BIGINT[] DEFAULT NULL,
    p_object_ids BIGINT[] DEFAULT NULL,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS TABLE (
    users_processed INTEGER,
    objects_processed INTEGER,
    cache_entries_created INTEGER
)
LANGUAGE plpgsql
AS $$
DECLARE
    user_id BIGINT;
    object_id BIGINT;
    users_count INTEGER := 0;
    objects_count INTEGER := 0;
    entries_count INTEGER := 0;
BEGIN
    -- If users not specified, take all active ones
    IF p_user_ids IS NULL THEN
        SELECT array_agg(id) INTO p_user_ids
        FROM iam.user
        WHERE tenant_id = p_tenant_id;
    END IF;
    
    -- If objects not specified, take all
    IF p_object_ids IS NULL THEN
        SELECT array_agg(id) INTO p_object_ids
        FROM security.object
        WHERE tenant_id = p_tenant_id;
    END IF;
    
    -- Pre-fill cache
    FOREACH user_id IN ARRAY p_user_ids LOOP
        users_count := users_count + 1;
        
        FOREACH object_id IN ARRAY p_object_ids LOOP
            objects_count := objects_count + 1;
            
            -- Create cache entry
            PERFORM cache.get_object_permissions(p_tenant_id, user_id, object_id, p_ttl_seconds);
            entries_count := entries_count + 1;
        END LOOP;
    END LOOP;
    
    RETURN QUERY SELECT users_count, objects_count, entries_count;
END;
$$;