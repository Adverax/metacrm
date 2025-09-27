-- ========================================
-- IAM EVENTS GENERATION MIGRATION
-- ========================================
-- This migration adds event generation functions for all IAM entities
-- Events are automatically published to the outbox table for reliable delivery
-- to external systems and other microservices

-- ========================================
-- EVENT GENERATION UTILITY FUNCTIONS
-- ========================================

-- Generic function to create outbox events
-- Used by all entity-specific event generation functions
--
-- Parameters:
--   p_aggregate_type: Type of aggregate (e.g., 'user', 'role', 'group')
--   p_aggregate_id: ID of the aggregate instance
--   p_event_type: Type of event (e.g., 'user.created', 'role.updated')
--   p_payload: Event payload as JSONB
--   p_headers: Additional headers as JSONB
--
-- Returns: void
--
-- Example:
--   SELECT bootstrap.create_outbox_event(
--     'user', 'usr_123', 'user.created',
--     '{"name": "John Doe", "email": "john@example.com"}',
--     '{"tenant_id": "uuid", "correlation_id": "req-123"}'
--   );
CREATE OR REPLACE FUNCTION bootstrap.create_outbox_event(
    p_aggregate_type TEXT,
    p_aggregate_id TEXT,
    p_event_type TEXT,
    p_payload JSONB,
    p_headers JSONB DEFAULT '{}'::jsonb
)
    RETURNS void
    LANGUAGE plpgsql
AS $$
DECLARE
    v_tenant_id UUID;
    v_principal_id BIGINT;
BEGIN
    -- Get current context
    v_tenant_id := bootstrap.current_tenant_id();
    v_principal_id := bootstrap.current_principal_id();

    -- Add tenant_id and principal_id to headers if available
    IF v_tenant_id IS NOT NULL THEN
        p_headers := p_headers || jsonb_build_object('tenant_id', v_tenant_id::text);
    END IF;

    IF v_principal_id IS NOT NULL THEN
        p_headers := p_headers || jsonb_build_object('principal_id', v_principal_id::text);
    END IF;

    -- Add timestamp and event metadata
    p_headers := p_headers || jsonb_build_object(
            'timestamp', extract(epoch from now())::text,
            'version', '1.0',
            'event_id', encode(gen_random_bytes(16), 'hex')
                              );

    -- Insert into outbox
    INSERT INTO bootstrap.outbox (aggregate_type, aggregate_id, event_type, payload, headers)
    VALUES (p_aggregate_type, p_aggregate_id, p_event_type, p_payload, p_headers);
END;
$$;

-- ========================================
-- USER EVENT GENERATION FUNCTIONS
-- ========================================

-- Generate user.created event
-- Called when a new user is created
CREATE OR REPLACE FUNCTION iam.generate_user_created_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_user_record_id TEXT;
BEGIN
    -- Get user record_id
    SELECT record_id INTO v_user_record_id
    FROM iam."user"
    WHERE tenant_id = NEW.tenant_id AND id = NEW.id;

    -- Build event payload
    v_payload := jsonb_build_object(
            'tenant_id', NEW.tenant_id::text,
            'user_id', v_user_record_id,
            'name', NEW.name,
            'email', NEW.email,
            'manager_id', CASE WHEN NEW.manager_id IS NOT NULL THEN
                                   (SELECT record_id FROM iam."user" WHERE tenant_id = NEW.tenant_id AND id = NEW.manager_id)
                               ELSE NULL END,
            'external_id', NEW.external_id,
            'created_by', CASE WHEN NEW.created_by_principal_id IS NOT NULL THEN
                                   NEW.created_by_principal_id::text
                               ELSE NULL END,
            'source', 'admin_created'  -- Default source, can be overridden
                 );

    -- Create outbox event
    PERFORM bootstrap.create_outbox_event(
            'user',
            v_user_record_id,
            'iam.user.created',
            v_payload
            );

    RETURN NEW;
END;
$$;

-- Generate user.updated event
-- Called when user information is updated
CREATE OR REPLACE FUNCTION iam.generate_user_updated_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_changes JSONB := '{}';
    v_user_record_id TEXT;
BEGIN
    -- Get user record_id
    SELECT record_id INTO v_user_record_id
    FROM iam."user"
    WHERE tenant_id = NEW.tenant_id AND id = NEW.id;

    -- Build changes object
    IF OLD.name IS DISTINCT FROM NEW.name THEN
        v_changes := v_changes || jsonb_build_object('name', jsonb_build_object(
                'old_value', OLD.name,
                'new_value', NEW.name
                                                             ));
    END IF;

    IF OLD.email IS DISTINCT FROM NEW.email THEN
        v_changes := v_changes || jsonb_build_object('email', jsonb_build_object(
                'old_value', OLD.email,
                'new_value', NEW.email
                                                              ));
    END IF;

    IF OLD.manager_id IS DISTINCT FROM NEW.manager_id THEN
        v_changes := v_changes || jsonb_build_object('manager_id', jsonb_build_object(
                'old_value', CASE WHEN OLD.manager_id IS NOT NULL THEN
                                      (SELECT record_id FROM iam."user" WHERE tenant_id = OLD.tenant_id AND id = OLD.manager_id)
                                  ELSE NULL END,
                'new_value', CASE WHEN NEW.manager_id IS NOT NULL THEN
                                      (SELECT record_id FROM iam."user" WHERE tenant_id = NEW.tenant_id AND id = NEW.manager_id)
                                  ELSE NULL END
                                                                   ));
    END IF;

    -- Only create event if there are actual changes
    IF v_changes != '{}' THEN
        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'user_id', v_user_record_id,
                'changes', v_changes,
                'updated_by', CASE WHEN NEW.updated_by_principal_id IS NOT NULL THEN
                                       NEW.updated_by_principal_id::text
                                   ELSE NULL END
                     );

        PERFORM bootstrap.create_outbox_event(
                'user',
                v_user_record_id,
                'iam.user.updated',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- Generate user.deleted event
-- Called when user is soft deleted
CREATE OR REPLACE FUNCTION iam.generate_user_deleted_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_user_record_id TEXT;
BEGIN
    -- Only trigger on soft delete (deleted_at changes from NULL to timestamp)
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        -- Get user record_id
        SELECT record_id INTO v_user_record_id
        FROM iam."user"
        WHERE tenant_id = NEW.tenant_id AND id = NEW.id;

        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'user_id', v_user_record_id,
                'name', NEW.name,
                'email', NEW.email,
                'deleted_by', CASE WHEN NEW.deleted_by_principal_id IS NOT NULL THEN
                                       NEW.deleted_by_principal_id::text
                                   ELSE NULL END,
                'reason', 'User account deactivated'  -- Default reason
                     );

        PERFORM bootstrap.create_outbox_event(
                'user',
                v_user_record_id,
                'iam.user.deleted',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- Generate user.manager_changed event
-- Called when user's manager is changed
CREATE OR REPLACE FUNCTION iam.generate_user_manager_changed_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_user_record_id TEXT;
BEGIN
    -- Only trigger if manager_id actually changed
    IF OLD.manager_id IS DISTINCT FROM NEW.manager_id THEN
        -- Get user record_id
        SELECT record_id INTO v_user_record_id
        FROM iam."user"
        WHERE tenant_id = NEW.tenant_id AND id = NEW.id;

        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'user_id', v_user_record_id,
                'old_manager_id', CASE WHEN OLD.manager_id IS NOT NULL THEN
                                           (SELECT record_id FROM iam."user" WHERE tenant_id = OLD.tenant_id AND id = OLD.manager_id)
                                       ELSE NULL END,
                'new_manager_id', CASE WHEN NEW.manager_id IS NOT NULL THEN
                                           (SELECT record_id FROM iam."user" WHERE tenant_id = NEW.tenant_id AND id = NEW.manager_id)
                                       ELSE NULL END,
                'changed_by', CASE WHEN NEW.updated_by_principal_id IS NOT NULL THEN
                                       NEW.updated_by_principal_id::text
                                   ELSE NULL END
                     );

        PERFORM bootstrap.create_outbox_event(
                'user',
                v_user_record_id,
                'iam.user.manager_changed',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- ========================================
-- ROLE EVENT GENERATION FUNCTIONS
-- ========================================

-- Generate role.created event
CREATE OR REPLACE FUNCTION iam.generate_role_created_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
BEGIN
    v_payload := jsonb_build_object(
            'tenant_id', NEW.tenant_id::text,
            'role_id', NEW.id::text,
            'label', NEW.label,
            'api_name', NEW.api_name,
            'parent_id', CASE WHEN NEW.parent_id IS NOT NULL THEN NEW.parent_id::text ELSE NULL END,
            'created_by', NEW.created_by_principal_id::text
                 );

    PERFORM bootstrap.create_outbox_event(
            'role',
            NEW.id::text,
            'iam.role.created',
            v_payload
            );

    RETURN NEW;
END;
$$;

-- Generate role.updated event
CREATE OR REPLACE FUNCTION iam.generate_role_updated_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_changes JSONB := '{}';
BEGIN
    -- Build changes object
    IF OLD.label IS DISTINCT FROM NEW.label THEN
        v_changes := v_changes || jsonb_build_object('label', jsonb_build_object(
                'old_value', OLD.label,
                'new_value', NEW.label
                                                              ));
    END IF;

    IF OLD.api_name IS DISTINCT FROM NEW.api_name THEN
        v_changes := v_changes || jsonb_build_object('api_name', jsonb_build_object(
                'old_value', OLD.api_name,
                'new_value', NEW.api_name
                                                                 ));
    END IF;

    IF OLD.parent_id IS DISTINCT FROM NEW.parent_id THEN
        v_changes := v_changes || jsonb_build_object('parent_id', jsonb_build_object(
                'old_value', OLD.parent_id::text,
                'new_value', NEW.parent_id::text
                                                                  ));
    END IF;

    -- Only create event if there are actual changes
    IF v_changes != '{}' THEN
        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'role_id', NEW.id::text,
                'changes', v_changes,
                'updated_by', NEW.updated_by_principal_id::text
                     );

        PERFORM bootstrap.create_outbox_event(
                'role',
                NEW.id::text,
                'iam.role.updated',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- Generate role.deleted event
CREATE OR REPLACE FUNCTION iam.generate_role_deleted_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
BEGIN
    -- Only trigger on soft delete
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'role_id', NEW.id::text,
                'label', NEW.label,
                'api_name', NEW.api_name,
                'deleted_by', NEW.deleted_by_principal_id::text,
                'reason', 'Role deactivated'
                     );

        PERFORM bootstrap.create_outbox_event(
                'role',
                NEW.id::text,
                'iam.role.deleted',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- ========================================
-- GROUP MEMBER EVENT GENERATION FUNCTIONS
-- ========================================

-- Generate group_member.added event
CREATE OR REPLACE FUNCTION cluster.generate_group_member_added_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_group_record_id TEXT;
    v_member_record_id TEXT;
BEGIN
    -- Get group record_id
    SELECT record_id INTO v_group_record_id
    FROM cluster.group
    WHERE tenant_id = NEW.tenant_id AND id = NEW.group_id;

    -- Get member record_id (user or group)
    IF NEW.member_user_id IS NOT NULL THEN
        SELECT record_id INTO v_member_record_id
        FROM iam."user"
        WHERE tenant_id = NEW.tenant_id AND id = NEW.member_user_id;
    ELSIF NEW.member_group_id IS NOT NULL THEN
        SELECT record_id INTO v_member_record_id
        FROM cluster.group
        WHERE tenant_id = NEW.tenant_id AND id = NEW.member_group_id;
    END IF;

    v_payload := jsonb_build_object(
            'tenant_id', NEW.tenant_id::text,
            'group_id', v_group_record_id,
            'member_type', CASE WHEN NEW.member_user_id IS NOT NULL THEN 'user' ELSE 'group' END,
            'member_id', v_member_record_id,
            'added_by', CASE WHEN NEW.created_by_principal_id IS NOT NULL THEN
                                 NEW.created_by_principal_id::text
                             ELSE NULL END,
            'expires_at', NEW.expires_at
                 );

    PERFORM bootstrap.create_outbox_event(
            'group_member',
            NEW.record_id,
            'iam.group_member.added',
            v_payload
            );

    RETURN NEW;
END;
$$;

-- Generate group_member.removed event
CREATE OR REPLACE FUNCTION cluster.generate_group_member_removed_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_group_record_id TEXT;
    v_member_record_id TEXT;
BEGIN
    -- Only trigger on soft delete
    IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        -- Get group record_id
        SELECT record_id INTO v_group_record_id
        FROM cluster.group
        WHERE tenant_id = OLD.tenant_id AND id = OLD.group_id;

        -- Get member record_id (user or group)
        IF OLD.member_user_id IS NOT NULL THEN
            SELECT record_id INTO v_member_record_id
            FROM iam."user"
            WHERE tenant_id = OLD.tenant_id AND id = OLD.member_user_id;
        ELSIF OLD.member_group_id IS NOT NULL THEN
            SELECT record_id INTO v_member_record_id
            FROM cluster.group
            WHERE tenant_id = OLD.tenant_id AND id = OLD.member_group_id;
        END IF;

        v_payload := jsonb_build_object(
                'tenant_id', OLD.tenant_id::text,
                'group_id', v_group_record_id,
                'member_type', CASE WHEN OLD.member_user_id IS NOT NULL THEN 'user' ELSE 'group' END,
                'member_id', v_member_record_id,
                'removed_by', CASE WHEN NEW.deleted_by_principal_id IS NOT NULL THEN
                                       NEW.deleted_by_principal_id::text
                                   ELSE NULL END,
                'reason', 'Member removed from group'
                     );

        PERFORM bootstrap.create_outbox_event(
                'group_member',
                OLD.record_id,
                'iam.group_member.removed',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- ========================================
-- PERMISSION EVENT GENERATION FUNCTIONS
-- ========================================

-- Generate permission_set.assigned_to_group event
CREATE OR REPLACE FUNCTION security.generate_permission_set_assigned_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_group_record_id TEXT;
BEGIN
    -- Only trigger when group_id is set (assigned to group)
    IF NEW.group_id IS NOT NULL THEN
        -- Get group record_id
        SELECT record_id INTO v_group_record_id
        FROM cluster.group
        WHERE tenant_id = NEW.tenant_id AND id = NEW.group_id;

        v_payload := jsonb_build_object(
                'tenant_id', NEW.tenant_id::text,
                'permission_set_id', NEW.id::text,
                'api_name', NEW.api_name,
                'group_id', v_group_record_id,
                'assigned_by', CASE WHEN NEW.created_by_principal_id IS NOT NULL THEN
                                        NEW.created_by_principal_id::text
                                    ELSE NULL END
                     );

        PERFORM bootstrap.create_outbox_event(
                'permission_set',
                NEW.id::text,
                'iam.permission_set.assigned_to_group',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- Generate permission_set.unassigned_from_group event
CREATE OR REPLACE FUNCTION security.generate_permission_set_unassigned_event()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
DECLARE
    v_payload JSONB;
    v_group_record_id TEXT;
BEGIN
    -- Only trigger when group_id is removed (unassigned from group)
    IF OLD.group_id IS NOT NULL AND NEW.group_id IS NULL THEN
        -- Get group record_id
        SELECT record_id INTO v_group_record_id
        FROM cluster.group
        WHERE tenant_id = OLD.tenant_id AND id = OLD.group_id;

        v_payload := jsonb_build_object(
                'tenant_id', OLD.tenant_id::text,
                'permission_set_id', OLD.id::text,
                'api_name', OLD.api_name,
                'group_id', v_group_record_id,
                'unassigned_by', CASE WHEN NEW.updated_by_principal_id IS NOT NULL THEN
                                          NEW.updated_by_principal_id::text
                                      ELSE NULL END
                     );

        PERFORM bootstrap.create_outbox_event(
                'permission_set',
                OLD.id::text,
                'iam.permission_set.unassigned_from_group',
                v_payload
                );
    END IF;

    RETURN NEW;
END;
$$;

-- ========================================
-- ATTACH EVENT TRIGGERS
-- ========================================

-- User event triggers
CREATE TRIGGER trg_user_created_event
    AFTER INSERT ON iam."user"
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_user_created_event();

CREATE TRIGGER trg_user_updated_event
    AFTER UPDATE ON iam."user"
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_user_updated_event();

CREATE TRIGGER trg_user_deleted_event
    AFTER UPDATE ON iam."user"
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_user_deleted_event();

CREATE TRIGGER trg_user_manager_changed_event
    AFTER UPDATE ON iam."user"
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_user_manager_changed_event();

-- Role event triggers
CREATE TRIGGER trg_role_created_event
    AFTER INSERT ON iam.role
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_role_created_event();

CREATE TRIGGER trg_role_updated_event
    AFTER UPDATE ON iam.role
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_role_updated_event();

CREATE TRIGGER trg_role_deleted_event
    AFTER UPDATE ON iam.role
    FOR EACH ROW
EXECUTE FUNCTION iam.generate_role_deleted_event();

-- Group member event triggers
CREATE TRIGGER trg_group_member_added_event
    AFTER INSERT ON cluster.group_member
    FOR EACH ROW
EXECUTE FUNCTION cluster.generate_group_member_added_event();

CREATE TRIGGER trg_group_member_removed_event
    AFTER UPDATE ON cluster.group_member
    FOR EACH ROW
EXECUTE FUNCTION cluster.generate_group_member_removed_event();

-- Permission set event triggers
CREATE TRIGGER trg_permission_set_assigned_event
    AFTER INSERT ON security.permission_set
    FOR EACH ROW
EXECUTE FUNCTION security.generate_permission_set_assigned_event();

CREATE TRIGGER trg_permission_set_unassigned_event
    AFTER UPDATE ON security.permission_set
    FOR EACH ROW
EXECUTE FUNCTION security.generate_permission_set_unassigned_event();

-- ========================================
-- EVENT MONITORING FUNCTIONS
-- ========================================

-- Get pending events count by type
-- Useful for monitoring event processing
CREATE OR REPLACE FUNCTION bootstrap.get_pending_events_count()
    RETURNS TABLE(event_type TEXT, pending_count BIGINT)
    LANGUAGE plpgsql
    STABLE
AS $$
BEGIN
    RETURN QUERY
        SELECT
            o.event_type,
            COUNT(*) as pending_count
        FROM bootstrap.outbox o
        WHERE o.status = 'pending'
        GROUP BY o.event_type
        ORDER BY pending_count DESC;
END;
$$;

-- Get events ready for processing
-- Used by outbox workers to get next batch of events
CREATE OR REPLACE FUNCTION bootstrap.get_events_ready_for_processing(p_limit INTEGER DEFAULT 100)
    RETURNS TABLE(
                     id BIGINT,
                     aggregate_type TEXT,
                     aggregate_id TEXT,
                     event_type TEXT,
                     payload JSONB,
                     headers JSONB,
                     attempt INTEGER
                 )
    LANGUAGE plpgsql
    STABLE
AS $$
BEGIN
    RETURN QUERY
        SELECT
            o.id,
            o.aggregate_type,
            o.aggregate_id,
            o.event_type,
            o.payload,
            o.headers,
            o.attempt
        FROM bootstrap.outbox o
        WHERE o.status = 'pending'
          AND o.next_attempt_at <= now()
        ORDER BY o.created_at
        LIMIT p_limit;
END;
$$;

-- Mark event as processing
-- Used by outbox workers to claim events
CREATE OR REPLACE FUNCTION bootstrap.mark_event_processing(p_id BIGINT, p_worker_id TEXT)
    RETURNS BOOLEAN
    LANGUAGE plpgsql
AS $$
DECLARE
    v_updated_rows INTEGER;
BEGIN
    UPDATE bootstrap.outbox
    SET status = 'processing',
        locked_by = p_worker_id,
        locked_at = now(),
        attempt = attempt + 1
    WHERE id = p_id
      AND status = 'pending'
      AND next_attempt_at <= now();

    GET DIAGNOSTICS v_updated_rows = ROW_COUNT;
    RETURN v_updated_rows > 0;
END;
$$;

-- Mark event as completed
-- Used by outbox workers after successful processing
CREATE OR REPLACE FUNCTION bootstrap.mark_event_completed(p_id BIGINT)
    RETURNS void
    LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE bootstrap.outbox
    SET status = 'done',
        published_at = now(),
        locked_by = NULL,
        locked_at = NULL
    WHERE id = p_id;
END;
$$;

-- Mark event as failed (with retry logic)
-- Used by outbox workers after failed processing
CREATE OR REPLACE FUNCTION bootstrap.mark_event_failed(p_id BIGINT, p_max_attempts INTEGER DEFAULT 5)
    RETURNS void
    LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE bootstrap.outbox
    SET status = CASE
                     WHEN attempt >= p_max_attempts THEN 'dead'
                     ELSE 'pending'
        END,
        next_attempt_at = now() + (attempt * 2 || ' minutes')::interval,
        locked_by = NULL,
        locked_at = NULL
    WHERE id = p_id;
END;
$$;
