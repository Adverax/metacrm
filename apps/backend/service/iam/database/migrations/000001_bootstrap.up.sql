CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS bootstrap;

-- Outbox pattern status enumeration
-- Used to track the processing state of domain events
CREATE TYPE bootstrap.outbox_status AS ENUM ('pending','processing','done','dead');

-- Outbox table for reliable event publishing
-- Implements the Outbox pattern to ensure eventual consistency between database transactions and external systems
--
-- Example usage:
-- INSERT INTO bootstrap.outbox (aggregate_type, aggregate_id, event_type, payload, headers)
-- VALUES ('user', '123', 'user_created', '{"name": "John Doe", "email": "john@example.com"}', '{"tenant_id": "uuid-here"}');
--
-- Processing flow:
-- 1. 'pending' - Event created, waiting to be processed
-- 2. 'processing' - Event is being processed by a worker
-- 3. 'done' - Event successfully published to external system
-- 4. 'dead' - Event failed after maximum retry attempts
CREATE TABLE bootstrap.outbox
(
    -- Unique identifier for the outbox record
    id              bigserial PRIMARY KEY,

    -- Type of aggregate that generated the event (e.g., 'user', 'order', 'product')
    -- Used for routing and filtering events
    aggregate_type  text                    NOT NULL,

    -- ID of the specific aggregate instance that generated the event
    -- Example: '123' for user with ID 123, 'ord-abc123' for order with ID ord-abc123
    aggregate_id    text                    NOT NULL,

    -- Type of event that occurred (e.g., 'user_created', 'order_placed', 'payment_processed')
    -- Used by consumers to understand what happened
    event_type      text                    NOT NULL,

    -- Event payload containing the actual data
    -- Example: {"name": "John Doe", "email": "john@example.com", "role": "admin"}
    payload         jsonb                   NOT NULL,

    -- Additional metadata and headers
    -- Example: {"tenant_id": "uuid-here", "correlation_id": "req-123", "user_id": "456"}
    headers         jsonb                   NOT NULL DEFAULT '{}'::jsonb,

    -- Current processing status of the event
    status          bootstrap.outbox_status NOT NULL DEFAULT 'pending'::bootstrap.outbox_status,

    -- Number of processing attempts made
    -- Incremented on each retry, used for exponential backoff
    attempt         int                     NOT NULL DEFAULT 0,

    -- When to attempt processing next (for retry logic)
    -- Example: now() + (attempt * 2) minutes for exponential backoff
    next_attempt_at timestamptz             NOT NULL DEFAULT now(),

    -- Identifier of the worker/process currently processing this event
    -- Used to prevent duplicate processing and for monitoring
    locked_by       text,

    -- When the event was locked for processing
    -- Used to detect stale locks and implement timeout logic
    locked_at       timestamptz,

    -- When the event was originally created
    created_at      timestamptz             NOT NULL DEFAULT now(),

    -- When the event was successfully published to external system
    -- NULL until event is successfully processed
    published_at    timestamptz
);

-- Index for efficient querying of events ready for processing
-- Used by outbox workers to find pending events that are due for retry
-- Example query: SELECT * FROM bootstrap.outbox WHERE status = 'pending' AND next_attempt_at <= now()
CREATE INDEX ON bootstrap.outbox (status, next_attempt_at);

-- Index for time-based queries and cleanup operations
-- Used for monitoring, analytics, and removing old processed events
-- Example query: SELECT * FROM bootstrap.outbox WHERE created_at > now() - interval '1 day'
CREATE INDEX ON bootstrap.outbox (created_at);

-- ========================================
-- OUTBOX PATTERN USAGE EXAMPLES
-- ========================================
--
-- 1. Creating an event (typically in a transaction with business logic):
-- BEGIN;
--   -- Business logic here
--   INSERT INTO iam.user (name, email) VALUES ('John Doe', 'john@example.com');
--
--   -- Create outbox event
--   INSERT INTO bootstrap.outbox (aggregate_type, aggregate_id, event_type, payload, headers)
--   VALUES (
--     'user',
--     '123',
--     'user_created',
--     '{"name": "John Doe", "email": "john@example.com", "role": "user"}',
--     '{"tenant_id": "550e8400-e29b-41d4-a716-446655440000", "correlation_id": "req-abc123"}'
--   );
-- COMMIT;
--
-- 2. Processing events (outbox worker):
-- -- Get next batch of pending events
-- SELECT * FROM bootstrap.outbox
-- WHERE status = 'pending'
--   AND next_attempt_at <= now()
-- ORDER BY created_at
-- LIMIT 100;
--
-- -- Mark event as processing
-- UPDATE bootstrap.outbox
-- SET status = 'processing',
--     locked_by = 'worker-1',
--     locked_at = now(),
--     attempt = attempt + 1
-- WHERE id = 123;
--
-- -- After successful processing
-- UPDATE bootstrap.outbox
-- SET status = 'done',
--     published_at = now(),
--     locked_by = NULL,
--     locked_at = NULL
-- WHERE id = 123;
--
-- -- After failed processing (with retry logic)
-- UPDATE bootstrap.outbox
-- SET status = CASE
--     WHEN attempt >= 5 THEN 'dead'
--     ELSE 'pending'
--   END,
--   next_attempt_at = now() + (attempt * 2 || ' minutes')::interval,
--   locked_by = NULL,
--   locked_at = NULL
-- WHERE id = 123;
--
-- 3. Monitoring queries:
-- -- Count events by status
-- SELECT status, COUNT(*) FROM bootstrap.outbox GROUP BY status;
--
-- -- Find stuck events (processing for too long)
-- SELECT * FROM bootstrap.outbox
-- WHERE status = 'processing'
--   AND locked_at < now() - interval '10 minutes';
--
-- -- Find events ready for retry
-- SELECT COUNT(*) FROM bootstrap.outbox
-- WHERE status = 'pending'
--   AND next_attempt_at <= now();


-- ========================================
-- BOOTSTRAP UTILITY FUNCTIONS
-- ========================================

-- Generate unique primary key with given prefix (3 characters) + 16 hex characters
-- Creates human-readable IDs with predictable prefixes for different entity types
--
-- Parameters:
--   v_prefix: 3-character prefix (e.g., 'usr', 'ord', 'prd')
--
-- Returns: 19-character string (3 prefix + 16 hex)
--
-- Examples:
--   SELECT bootstrap.generate_pk('usr'); -- Returns: 'usra1b2c3d4e5f67890'
--   SELECT bootstrap.generate_pk('ord'); -- Returns: 'ordf9e8d7c6b5a43210'
--   SELECT bootstrap.generate_pk('grp'); -- Returns: 'grp1234567890abcdef'
--
-- Usage in tables:
--   INSERT INTO iam.user (id, name) VALUES (bootstrap.generate_pk('usr'), 'John Doe');
CREATE OR REPLACE FUNCTION bootstrap.generate_pk(v_prefix CHAR(3))
    RETURNS TEXT VOLATILE LANGUAGE plpgsql
    SET search_path = pg_catalog, public, pg_temp AS
$$
BEGIN
    -- Validate input parameters
    IF v_prefix IS NULL THEN
        RAISE EXCEPTION 'Prefix cannot be NULL';
    END IF;

    IF length(trim(v_prefix)) = 0 THEN
        RAISE EXCEPTION 'Prefix cannot be empty';
    END IF;

    IF length(trim(v_prefix)) > 3 THEN
        RAISE EXCEPTION 'Prefix cannot be longer than 3 characters';
    END IF;

    -- Check for valid characters (only letters and numbers)
    IF v_prefix !~ '^[a-zA-Z0-9]+$' THEN
        RAISE EXCEPTION 'Prefix can only contain alphanumeric characters';
    END IF;

    -- Generate ID
    RETURN v_prefix || encode(gen_random_bytes(8), 'hex');
END;
$$;

-- Get current tenant ID from session context
-- Retrieves the tenant ID that was set for the current database session
--
-- Returns: UUID of current tenant or NULL if not set
--
-- Examples:
--   SELECT bootstrap.current_tenant_id(); -- Returns: '550e8400-e29b-41d4-a716-446655440000'
--   SELECT bootstrap.current_tenant_id(); -- Returns: NULL (if not set)
--
-- Usage in tables:
--   tenant_id UUID NOT NULL DEFAULT bootstrap.current_tenant_id()
--
-- Note: This function is STABLE, meaning it returns the same value within a transaction
CREATE OR REPLACE FUNCTION bootstrap.current_tenant_id()
    RETURNS uuid STABLE LANGUAGE plpgsql AS
$$
DECLARE
    v_tenant_id_text TEXT;
    v_tenant_id UUID;
BEGIN
    -- Get value from session settings
    BEGIN
        v_tenant_id_text := current_setting('app.tenant_id', true);
    EXCEPTION WHEN OTHERS THEN
        RETURN NULL;
    END;

    -- Check for empty value
    IF v_tenant_id_text IS NULL OR trim(v_tenant_id_text) = '' THEN
        RETURN NULL;
    END IF;

    -- Validate UUID format
    BEGIN
        v_tenant_id := v_tenant_id_text::uuid;
        RETURN v_tenant_id;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION 'Invalid tenant ID format: %', v_tenant_id_text;
    END;
END;
$$;

-- Get current principal ID from session context
-- Retrieves the principal (user/system) ID that was set for the current database session
-- Used for audit trails and tracking who performed database operations
--
-- Returns: BIGINT ID of current principal or NULL if not set
--
-- Examples:
--   SELECT bootstrap.current_principal_id(); -- Returns: 123
--   SELECT bootstrap.current_principal_id(); -- Returns: NULL (if not set)
--
-- Usage in tables:
--   created_by_principal_id BIGINT NOT NULL DEFAULT bootstrap.current_principal_id()
--   updated_by_principal_id BIGINT NOT NULL DEFAULT bootstrap.current_principal_id()
--
-- Note: This function is STABLE, meaning it returns the same value within a transaction
CREATE OR REPLACE FUNCTION bootstrap.current_principal_id()
    RETURNS bigint STABLE LANGUAGE plpgsql AS
$$
DECLARE
    v_principal_id_text TEXT;
    v_principal_id BIGINT;
BEGIN
    -- Get value from session settings
    BEGIN
        v_principal_id_text := current_setting('app.principal_id', true);
    EXCEPTION WHEN OTHERS THEN
        RETURN NULL;
    END;

    -- Check for empty value
    IF v_principal_id_text IS NULL OR trim(v_principal_id_text) = '' THEN
        RETURN NULL;
    END IF;

    -- Validate BIGINT format
    BEGIN
        v_principal_id := v_principal_id_text::bigint;
    EXCEPTION WHEN invalid_text_representation THEN
        RAISE EXCEPTION 'Invalid principal ID format: %', v_principal_id_text;
    END;

    -- Check for positive values
    IF v_principal_id <= 0 THEN
        RAISE EXCEPTION 'Principal ID must be positive: %', v_principal_id;
    END IF;

    RETURN v_principal_id;
END;
$$;

-- Set session context for tenant and principal
-- Establishes the tenant and principal context for the current database session
-- This context is used by current_tenant_id() and current_principal_id() functions
--
-- Parameters:
--   p_tenant: UUID of the tenant to set as current
--   p_principal: BIGINT ID of the principal (user/system) to set as current
--
-- Returns: void
--
-- Examples:
--   SELECT bootstrap.set_ctx('550e8400-e29b-41d4-a716-446655440000', 123);
--   -- Now bootstrap.current_tenant_id() returns the tenant UUID
--   -- and bootstrap.current_principal_id() returns 123
--
-- Usage in application:
--   -- At the beginning of each request/transaction
--   SELECT bootstrap.set_ctx(user_tenant_id, user_principal_id);
--   -- All subsequent operations will use this context
--
-- Note: This function is VOLATILE and affects session state
CREATE OR REPLACE FUNCTION bootstrap.set_ctx(p_tenant uuid, p_principal bigint)
    RETURNS void VOLATILE LANGUAGE plpgsql AS
$$
BEGIN
    -- Validate tenant_id
    IF p_tenant IS NULL THEN
        RAISE EXCEPTION 'Tenant ID cannot be NULL';
    END IF;

    -- Validate principal_id
    IF p_principal IS NULL THEN
        RAISE EXCEPTION 'Principal ID cannot be NULL';
    END IF;

    IF p_principal <= 0 THEN
        RAISE EXCEPTION 'Principal ID must be positive: %', p_principal;
    END IF;

    -- Set session context
    PERFORM set_config('app.tenant_id', p_tenant::text, true);
    PERFORM set_config('app.principal_id', p_principal::text, true);
END;
$$;

-- Convert text to database-safe slug
-- Transforms any text into a valid identifier for database objects (tables, columns, etc.)
-- Removes special characters and replaces dots with underscores
--
-- Parameters:
--   p_alias: Input text to convert to slug
--
-- Returns: Clean slug with only alphanumeric characters and underscores
--
-- Examples:
--   SELECT bootstrap.slug_from('user.profile'); -- Returns: 'user_profile'
--   SELECT bootstrap.slug_from('my-table@2024'); -- Returns: 'mytable2024'
--   SELECT bootstrap.slug_from('test.name#1'); -- Returns: 'test_name1'
--   SELECT bootstrap.slug_from('simple_name'); -- Returns: 'simple_name'
--
-- Usage in dynamic SQL:
--   SELECT bootstrap.slug_from('schema.table') -- For creating safe object names
--
-- Note: This function is IMMUTABLE and STABLE, safe for use in indexes and constraints
CREATE OR REPLACE FUNCTION bootstrap.slug_from(p_alias TEXT)
    RETURNS TEXT IMMUTABLE LANGUAGE plpgsql AS
$$
DECLARE
    v_slug TEXT;
BEGIN
    -- Validate input parameter
    IF p_alias IS NULL THEN
        RAISE EXCEPTION 'Input alias cannot be NULL';
    END IF;

    IF trim(p_alias) = '' THEN
        RAISE EXCEPTION 'Input alias cannot be empty';
    END IF;

    -- Process slug
    v_slug := regexp_replace(replace(p_alias, '.', '_'), '[^a-zA-Z0-9_]', '', 'g');

    -- Check result
    IF v_slug IS NULL OR trim(v_slug) = '' THEN
        RAISE EXCEPTION 'Resulting slug is empty after processing: %', p_alias;
    END IF;

    -- Check maximum length (for PostgreSQL identifiers)
    IF length(v_slug) > 63 THEN
        RAISE EXCEPTION 'Resulting slug is too long (max 63 characters): %', v_slug;
    END IF;

    RETURN v_slug;
END;
$$;

-- ========================================
-- AUDIT TRIGGER FUNCTIONS
-- ========================================

-- Simple timestamp update trigger function
-- Updates the updated_at field with current timestamp on every UPDATE operation
-- Used for basic audit trails without user tracking
--
-- Returns: NEW record with updated timestamp
--
-- Usage:
--   CREATE TRIGGER trg_table_update_time
--   BEFORE UPDATE ON my_table
--   FOR EACH ROW
--   EXECUTE FUNCTION bootstrap.updated_at_setter();
--
-- Requirements:
--   - Table must have an 'updated_at' column of type timestamptz
--
-- Example:
--   UPDATE my_table SET name = 'New Name' WHERE id = 1;
--   -- updated_at will be automatically set to current timestamp
CREATE OR REPLACE FUNCTION bootstrap.updated_at_setter() RETURNS trigger
    LANGUAGE plpgsql AS
$function$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$function$;

-- Advanced update trigger function with user tracking
-- Updates both updated_at timestamp and updated_by_principal_id fields
-- Provides full audit trail including who made the change
--
-- Returns: NEW record with updated timestamp and principal ID
--
-- Usage:
--   CREATE TRIGGER trg_table_update
--   BEFORE UPDATE ON my_table
--   FOR EACH ROW
--   EXECUTE FUNCTION bootstrap.updated_setter();
--
-- Requirements:
--   - Table must have 'updated_at' column of type timestamptz
--   - Table must have 'updated_by_principal_id' column of type bigint
--   - Session context must be set with bootstrap.set_ctx()
--
-- Example:
--   SELECT bootstrap.set_ctx('tenant-uuid', 123);
--   UPDATE my_table SET name = 'New Name' WHERE id = 1;
--   -- updated_at = now(), updated_by_principal_id = 123
--
-- Note: If no principal context is set, only updated_at will be set
CREATE OR REPLACE FUNCTION bootstrap.updated_setter() RETURNS trigger
    LANGUAGE plpgsql AS
$function$
DECLARE
    v_principal_id BIGINT;
BEGIN
    NEW.updated_at = now();
    v_principal_id := bootstrap.current_principal_id();
    IF v_principal_id IS NOT NULL THEN
        NEW.updated_by_principal_id = v_principal_id;
    END IF;
    RETURN NEW;
END;
$function$;

-- Complete audit trigger function with soft delete support
-- Handles both regular updates and soft deletes with full audit trail
-- Tracks who updated records and who deleted them (soft delete)
--
-- Returns: NEW record with appropriate audit fields set
--
-- Usage:
--   CREATE TRIGGER trg_table_update_with_delete
--   BEFORE UPDATE ON my_table
--   FOR EACH ROW
--   EXECUTE FUNCTION bootstrap.updated_deleted_setter();
--
-- Requirements:
--   - Table must have 'updated_at' column of type timestamptz
--   - Table must have 'updated_by_principal_id' column of type bigint
--   - Table must have 'deleted_at' column of type timestamptz
--   - Table must have 'deleted_by_principal_id' column of type bigint
--   - Session context must be set with bootstrap.set_ctx()
--
-- Behavior:
--   - On regular update: sets updated_at and updated_by_principal_id
--   - On soft delete (deleted_at changes from NULL to timestamp):
--     sets deleted_by_principal_id to current principal
--   - On restore (deleted_at changes from timestamp to NULL):
--     sets updated_at and updated_by_principal_id
--
-- Example:
--   SELECT bootstrap.set_ctx('tenant-uuid', 123);
--   -- Regular update
--   UPDATE my_table SET name = 'New Name' WHERE id = 1;
--   -- updated_at = now(), updated_by_principal_id = 123
--
--   -- Soft delete
--   UPDATE my_table SET deleted_at = now() WHERE id = 1;
--   -- deleted_by_principal_id = 123
--
-- Note: If no principal context is set, only timestamps will be set
CREATE OR REPLACE FUNCTION bootstrap.updated_deleted_setter() RETURNS trigger
    LANGUAGE plpgsql AS
$function$
DECLARE
    v_principal_id BIGINT;
BEGIN
    v_principal_id := bootstrap.current_principal_id();
    if NEW.deleted_at IS NULL then
        NEW.updated_at = now();
        IF v_principal_id IS NOT NULL THEN
            NEW.updated_by_principal_id = v_principal_id;
        END IF;
    ELSE
        IF OLD.deleted_at IS NULL THEN
            IF v_principal_id IS NOT NULL THEN
                NEW.deleted_by_principal_id = v_principal_id;
            END IF;
        END IF;
    END IF;
    RETURN NEW;
END;
$function$;

-- ========================================
-- AUTOMATED TRIGGER MANAGEMENT FUNCTIONS
-- ========================================

-- Automatically attach appropriate audit triggers to a table
-- Analyzes table structure and creates the most suitable audit trigger
-- Supports three levels of audit functionality based on available columns
--
-- Parameters:
--   p_schema_name: Schema name containing the table
--   p_table_name: Table name to attach triggers to
--
-- Returns: void
--
-- Trigger Selection Logic:
--   1. If table has updated_at + updated_by_principal_id + deleted_at + deleted_by_principal_id:
--      → Uses updated_deleted_setter (full audit with soft delete)
--   2. If table has updated_at + updated_by_principal_id:
--      → Uses updated_setter (audit with user tracking)
--   3. If table has only updated_at:
--      → Uses updated_at_setter (basic timestamp audit)
--   4. If table has none of these columns:
--      → No triggers attached
--
-- Examples:
--   SELECT bootstrap.attach_audit_triggers('iam', 'user');
--   SELECT bootstrap.attach_audit_triggers('security', 'permission');
--   SELECT bootstrap.attach_audit_triggers('cluster', 'group');
--
-- Usage in migrations:
--   -- After creating a table with audit columns
--   CREATE TABLE my_table (
--     id BIGSERIAL PRIMARY KEY,
--     name TEXT NOT NULL,
--     updated_at timestamptz NOT NULL DEFAULT now(),
--     updated_by_principal_id BIGINT NOT NULL DEFAULT bootstrap.current_principal_id(),
--     deleted_at timestamptz,
--     deleted_by_principal_id BIGINT
--   );
--   SELECT bootstrap.attach_audit_triggers('public', 'my_table');
--
-- Note: This function is idempotent - can be called multiple times safely
CREATE OR REPLACE FUNCTION bootstrap.attach_audit_triggers(p_schema_name TEXT, p_table_name TEXT) RETURNS void
    LANGUAGE plpgsql AS
$function$
DECLARE
    v_alias             TEXT;
    v_dataset           TEXT := format('%s.%s', p_schema_name, p_table_name);
    v_has_updated_at    BOOLEAN;
    v_has_updated_by_id BOOLEAN;
    v_has_deleted_at    BOOLEAN;
    v_has_deleted_by_id BOOLEAN;
    v_has_name          BOOLEAN;
BEGIN
    v_alias := bootstrap.slug_from(v_dataset);

    SELECT COALESCE(bool_or(column_name = 'updated_at'), FALSE),
           COALESCE(bool_or(column_name = 'updated_by_principal_id'), FALSE),
           COALESCE(bool_or(column_name = 'deleted_at'), FALSE),
           COALESCE(bool_or(column_name = 'deleted_by_principal_id'), FALSE),
           COALESCE(bool_or(column_name = 'name'), FALSE)
    INTO v_has_updated_at, v_has_updated_by_id, v_has_deleted_at, v_has_deleted_by_id, v_has_name
    FROM information_schema.columns
    WHERE table_schema = p_schema_name
      AND table_name = p_table_name;

    IF v_has_updated_at THEN
        IF v_has_updated_by_id THEN
            IF v_has_deleted_at AND v_has_deleted_by_id THEN
                -- has deleted_at and deleted_by_id - use update_with_delete
                EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_update_with_delete ON %s;', v_alias, v_dataset);
                EXECUTE format('
        CREATE TRIGGER trg_%s_update_with_delete
        BEFORE UPDATE ON %s
        FOR EACH ROW
        EXECUTE FUNCTION bootstrap.updated_deleted_setter();
    ', v_alias, v_dataset);
            ELSE
                -- has updated_by_id - use update_with_user
                EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_update ON %s;', v_alias, v_dataset);
                EXECUTE format('
        CREATE TRIGGER trg_%s_update
        BEFORE UPDATE ON %s
        FOR EACH ROW
        EXECUTE FUNCTION bootstrap.updated_setter();
    ', v_alias, v_dataset);
            END IF;
        ELSE
            -- has only updated_at - use update_time
            EXECUTE format('DROP TRIGGER IF EXISTS trg_%s_update_time ON %s;', v_alias, v_dataset);
            EXECUTE format('
        CREATE TRIGGER trg_%s_update_time
        BEFORE UPDATE ON %s
        FOR EACH ROW
        EXECUTE FUNCTION bootstrap.updated_at_setter();
    ', v_alias, v_dataset);
        END IF;
    END IF;
END
$function$;

-- ========================================
-- PARTITIONING UTILITY FUNCTIONS
-- ========================================

-- Create hash partitions for a table
-- Automatically creates the specified number of hash partitions for a table
-- Uses HASH partitioning with MODULUS/REMAINDER for even distribution
--
-- Parameters:
--   p_schema_name: Schema name containing the table
--   p_table_name: Table name to partition
--   p_count: Number of partitions to create
--
-- Returns: void
--
-- Requirements:
--   - Table must already exist and be defined as PARTITION BY HASH
--   - Table must have a partitioning column (usually tenant_id)
--
-- Examples:
--   -- Create 16 partitions for iam.user table
--   SELECT bootstrap.make_partitions('iam', 'user', 16);
--
--   -- Create 8 partitions for security.object table
--   SELECT bootstrap.make_partitions('security', 'object', 8);
--
-- Usage in migrations:
--   -- 1. Create partitioned table
--   CREATE TABLE iam.user (
--     id BIGSERIAL,
--     tenant_id UUID NOT NULL,
--     name TEXT NOT NULL,
--     PRIMARY KEY (id, tenant_id)
--   ) PARTITION BY HASH (tenant_id);
--
--   -- 2. Create partitions
--   SELECT bootstrap.make_partitions('iam', 'user', 16);
--
-- Generated partition names:
--   - iam.user_p0, iam.user_p1, iam.user_p2, ..., iam.user_p15
--
-- Partition distribution:
--   - Each partition gets 1/16 of the hash space
--   - Rows are distributed based on hash(tenant_id) % 16
--
-- Note: This function is idempotent - can be called multiple times safely
CREATE OR REPLACE FUNCTION bootstrap.make_partitions(p_schema_name TEXT, p_table_name TEXT, p_count INT) RETURNS void
    LANGUAGE plpgsql AS
$function$
DECLARE
    v_schema_exists BOOLEAN;
    v_table_exists BOOLEAN;
    v_is_partitioned BOOLEAN;
BEGIN
    -- Validate input parameters
    IF p_schema_name IS NULL OR trim(p_schema_name) = '' THEN
        RAISE EXCEPTION 'Schema name cannot be NULL or empty';
    END IF;

    IF p_table_name IS NULL OR trim(p_table_name) = '' THEN
        RAISE EXCEPTION 'Table name cannot be NULL or empty';
    END IF;

    IF p_count IS NULL OR p_count <= 0 THEN
        RAISE EXCEPTION 'Partition count must be positive: %', p_count;
    END IF;

    IF p_count > 1000 THEN
        RAISE EXCEPTION 'Partition count too large (max 1000): %', p_count;
    END IF;

    -- Check if schema exists
    SELECT EXISTS(
        SELECT 1 FROM information_schema.schemata
        WHERE schema_name = p_schema_name
    ) INTO v_schema_exists;

    IF NOT v_schema_exists THEN
        RAISE EXCEPTION 'Schema does not exist: %', p_schema_name;
    END IF;

    -- Check if table exists
    SELECT EXISTS(
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = p_schema_name AND table_name = p_table_name
    ) INTO v_table_exists;

    IF NOT v_table_exists THEN
        RAISE EXCEPTION 'Table does not exist: %.%', p_schema_name, p_table_name;
    END IF;

    -- Check if table is partitioned
    SELECT EXISTS(
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = p_schema_name
          AND c.relname = p_table_name
          AND c.relkind = 'p'  -- partitioned table
    ) INTO v_is_partitioned;

    IF NOT v_is_partitioned THEN
        RAISE EXCEPTION 'Table %.% is not partitioned', p_schema_name, p_table_name;
    END IF;

    -- Create partitions
    FOR r IN 0..p_count-1 LOOP
        EXECUTE format('CREATE TABLE IF NOT EXISTS %I."%s_p%3$s" PARTITION OF %I."%s"
                       FOR VALUES WITH (MODULUS %5$L, REMAINDER %3$s);',
                       p_schema_name, p_table_name, r, p_schema_name, p_table_name, p_count);
    END LOOP;
END
$function$;
