-- ========================================
-- IAM PERMISSIONS SYNC MIGRATION
-- ========================================
-- This migration adds functions for external services to sync user permissions
-- via lazy loading when they receive IAM events

-- ========================================
-- PERMISSION SYNC UTILITY FUNCTIONS
-- ========================================

-- Get complete user permissions snapshot for caching
-- Used by external services to get all permissions for a user
CREATE OR REPLACE FUNCTION iam.get_user_permissions_snapshot(
    p_tenant_id UUID,
    p_user_id TEXT,
    p_object_api_names TEXT[] DEFAULT NULL,
    p_include_group_memberships BOOLEAN DEFAULT true,
    p_include_permission_sources BOOLEAN DEFAULT false,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_user_internal_id BIGINT;
    v_object_permissions JSONB := '[]'::jsonb;
    v_field_permissions JSONB := '[]'::jsonb;
    v_group_memberships JSONB := '[]'::jsonb;
    v_permission_sources JSONB := '[]'::jsonb;
    v_object_record RECORD;
    v_field_record RECORD;
    v_group_record RECORD;
    v_permission_source RECORD;
    v_object_ids BIGINT[];
BEGIN
    -- Get user internal ID
    SELECT id INTO v_user_internal_id
    FROM iam."user"
    WHERE tenant_id = p_tenant_id AND record_id = p_user_id;
    
    IF v_user_internal_id IS NULL THEN
        RAISE EXCEPTION 'User not found: %', p_user_id;
    END IF;
    
    -- Get object permissions
    IF p_object_api_names IS NULL THEN
        -- Get all objects
        SELECT jsonb_agg(
            jsonb_build_object(
                'object_api_name', so.api_name,
                'object_id', so.id,
                'permissions', COALESCE(cache.get_object_permissions(p_tenant_id, v_user_internal_id, so.id, p_ttl_seconds), 0),
                'computed_at', now()
            )
        ) INTO v_object_permissions
        FROM security.object so
        WHERE so.tenant_id = p_tenant_id;
    ELSE
        -- Get specific objects
        SELECT jsonb_agg(
            jsonb_build_object(
                'object_api_name', so.api_name,
                'object_id', so.id,
                'permissions', COALESCE(cache.get_object_permissions(p_tenant_id, v_user_internal_id, so.id, p_ttl_seconds), 0),
                'computed_at', now()
            )
        ) INTO v_object_permissions
        FROM security.object so
        WHERE so.tenant_id = p_tenant_id 
          AND so.api_name = ANY(p_object_api_names);
    END IF;
    
    -- Get field permissions for specified objects
    IF p_object_api_names IS NOT NULL THEN
        SELECT jsonb_agg(
            jsonb_build_object(
                'object_api_name', so.api_name,
                'field_api_name', sf.api_name,
                'object_id', so.id,
                'field_id', sf.id,
                'permissions', COALESCE(cache.get_field_permissions(p_tenant_id, v_user_internal_id, so.id, sf.id, p_ttl_seconds), 0),
                'computed_at', now()
            )
        ) INTO v_field_permissions
        FROM security.object so
        JOIN security.field sf ON sf.tenant_id = so.tenant_id AND sf.object_id = so.id
        WHERE so.tenant_id = p_tenant_id 
          AND so.api_name = ANY(p_object_api_names);
    END IF;
    
    -- Get group memberships if requested
    IF p_include_group_memberships THEN
        SELECT jsonb_agg(
            jsonb_build_object(
                'group_record_id', cg.record_id,
                'group_label', cg.label,
                'group_api_name', cg.api_name,
                'group_type', cg.group_type,
                'membership_type', 'direct',
                'related_entity_id', CASE 
                    WHEN cg.group_type = 'role_based' THEN 
                        (SELECT record_id FROM iam.role WHERE tenant_id = p_tenant_id AND id = cg.related_entity_id)
                    WHEN cg.group_type = 'territory_based' THEN 
                        (SELECT record_id FROM iam.territory WHERE tenant_id = p_tenant_id AND id = cg.related_entity_id)
                    ELSE NULL 
                END,
                'joined_at', cgm.created_at,
                'expires_at', cgm.expires_at
            )
        ) INTO v_group_memberships
        FROM cluster.group_member cgm
        JOIN cluster.group cg ON cg.tenant_id = cgm.tenant_id AND cg.id = cgm.group_id
        WHERE cgm.tenant_id = p_tenant_id 
          AND cgm.member_user_id = v_user_internal_id
          AND cgm.deleted_at IS NULL;
    END IF;
    
    -- Get permission sources if requested
    IF p_include_permission_sources THEN
        -- This would require more complex logic to trace permission sources
        -- For now, return basic group-based sources
        SELECT jsonb_agg(
            jsonb_build_object(
                'source_type', 'group',
                'source_id', cg.record_id,
                'source_name', cg.label,
                'permissions', COALESCE(cache.get_object_permissions(p_tenant_id, v_user_internal_id, so.id, p_ttl_seconds), 0),
                'priority', 50
            )
        ) INTO v_permission_sources
        FROM cluster.group_member cgm
        JOIN cluster.group cg ON cg.tenant_id = cgm.tenant_id AND cg.id = cgm.group_id
        JOIN security.permission_set sps ON sps.tenant_id = cg.tenant_id AND sps.group_id = cg.id
        JOIN security.object_permissions sop ON sop.tenant_id = sps.tenant_id AND sop.permission_set_id = sps.id
        JOIN security.object so ON so.tenant_id = sop.tenant_id AND so.id = sop.object_id
        WHERE cgm.tenant_id = p_tenant_id 
          AND cgm.member_user_id = v_user_internal_id
          AND cgm.deleted_at IS NULL;
    END IF;
    
    -- Build result
    v_result := jsonb_build_object(
        'user_id', p_user_id,
        'tenant_id', p_tenant_id::text,
        'permissions', COALESCE(v_object_permissions, '[]'::jsonb),
        'field_permissions', COALESCE(v_field_permissions, '[]'::jsonb),
        'group_memberships', COALESCE(v_group_memberships, '[]'::jsonb),
        'permission_sources', COALESCE(v_permission_sources, '[]'::jsonb),
        'snapshot_at', now(),
        'snapshot_version', '1.0'
    );
    
    RETURN v_result;
END;
$$;

-- Get user permissions changes since timestamp
-- Used for incremental sync by external services
CREATE OR REPLACE FUNCTION iam.get_user_permissions_changes(
    p_tenant_id UUID,
    p_user_id TEXT,
    p_since TIMESTAMPTZ,
    p_object_api_names TEXT[] DEFAULT NULL,
    p_include_group_memberships BOOLEAN DEFAULT true
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_user_internal_id BIGINT;
    v_changes JSONB := '[]'::jsonb;
    v_change_record RECORD;
BEGIN
    -- Get user internal ID
    SELECT id INTO v_user_internal_id
    FROM iam."user"
    WHERE tenant_id = p_tenant_id AND record_id = p_user_id;
    
    IF v_user_internal_id IS NULL THEN
        RAISE EXCEPTION 'User not found: %', p_user_id;
    END IF;
    
    -- Get group membership changes
    IF p_include_group_memberships THEN
        -- Check for group membership changes in outbox events
        SELECT jsonb_agg(
            jsonb_build_object(
                'change_id', encode(gen_random_bytes(16), 'hex'),
                'user_id', p_user_id,
                'change_type', CASE 
                    WHEN o.event_type = 'iam.group_member.added' THEN 'group_joined'
                    WHEN o.event_type = 'iam.group_member.removed' THEN 'group_left'
                    ELSE 'permission_changed'
                END,
                'group_id', o.payload->>'group_id',
                'changed_at', to_timestamp((o.headers->>'timestamp')::numeric),
                'change_reason', o.payload->>'reason',
                'event_id', o.headers->>'event_id'
            )
        ) INTO v_changes
        FROM bootstrap.outbox o
        WHERE o.tenant_id = p_tenant_id
          AND o.event_type IN ('iam.group_member.added', 'iam.group_member.removed')
          AND o.payload->>'member_id' = p_user_id
          AND o.created_at > p_since
          AND o.status = 'done';
    END IF;
    
    -- Build result
    v_result := jsonb_build_object(
        'user_id', p_user_id,
        'tenant_id', p_tenant_id::text,
        'changes', COALESCE(v_changes, '[]'::jsonb),
        'total_changes', jsonb_array_length(COALESCE(v_changes, '[]'::jsonb)),
        'last_change', CASE 
            WHEN v_changes IS NOT NULL AND jsonb_array_length(v_changes) > 0 THEN
                (SELECT MAX(to_timestamp((o.headers->>'timestamp')::numeric))
                 FROM bootstrap.outbox o
                 WHERE o.tenant_id = p_tenant_id
                   AND o.event_type IN ('iam.group_member.added', 'iam.group_member.removed')
                   AND o.payload->>'member_id' = p_user_id
                   AND o.created_at > p_since
                   AND o.status = 'done')
            ELSE p_since
        END
    );
    
    RETURN v_result;
END;
$$;

-- Bulk get permissions for multiple users
-- Used for batch processing by external services
CREATE OR REPLACE FUNCTION iam.get_bulk_user_permissions(
    p_tenant_id UUID,
    p_user_ids TEXT[],
    p_object_api_names TEXT[] DEFAULT NULL,
    p_include_group_memberships BOOLEAN DEFAULT true,
    p_include_permission_sources BOOLEAN DEFAULT false,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_snapshots JSONB := '[]'::jsonb;
    v_user_id TEXT;
    v_snapshot JSONB;
    v_failed_users TEXT[] := '{}';
BEGIN
    -- Validate input
    IF array_length(p_user_ids, 1) > 100 THEN
        RAISE EXCEPTION 'Maximum 100 users allowed in bulk request';
    END IF;
    
    -- Get permissions for each user
    FOREACH v_user_id IN ARRAY p_user_ids
    LOOP
        BEGIN
            v_snapshot := iam.get_user_permissions_snapshot(
                p_tenant_id,
                v_user_id,
                p_object_api_names,
                p_include_group_memberships,
                p_include_permission_sources,
                p_ttl_seconds
            );
            v_snapshots := v_snapshots || v_snapshot;
        EXCEPTION WHEN OTHERS THEN
            v_failed_users := v_failed_users || v_user_id;
        END;
    END LOOP;
    
    -- Build result
    v_result := jsonb_build_object(
        'snapshots', v_snapshots,
        'total_users', array_length(p_user_ids, 1),
        'failed_users', array_length(v_failed_users, 1),
        'failed_user_ids', v_failed_users,
        'generated_at', now()
    );
    
    RETURN v_result;
END;
$$;

-- Get permissions for users affected by group membership changes
-- Used when external services receive group membership events
CREATE OR REPLACE FUNCTION iam.get_group_membership_permissions(
    p_tenant_id UUID,
    p_group_record_id TEXT,
    p_object_api_names TEXT[] DEFAULT NULL,
    p_include_all_members BOOLEAN DEFAULT false,
    p_include_permission_sources BOOLEAN DEFAULT false,
    p_ttl_seconds INTEGER DEFAULT 3600
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_group_internal_id BIGINT;
    v_group_info JSONB;
    v_member_permissions JSONB := '[]'::jsonb;
    v_member_record RECORD;
    v_user_snapshot JSONB;
BEGIN
    -- Get group internal ID
    SELECT id INTO v_group_internal_id
    FROM cluster.group
    WHERE tenant_id = p_tenant_id AND record_id = p_group_record_id;
    
    IF v_group_internal_id IS NULL THEN
        RAISE EXCEPTION 'Group not found: %', p_group_record_id;
    END IF;
    
    -- Get group info
    SELECT jsonb_build_object(
        'group_record_id', record_id,
        'group_label', label,
        'group_api_name', api_name,
        'group_type', group_type,
        'member_count', (
            SELECT COUNT(*)
            FROM cluster.group_member
            WHERE tenant_id = p_tenant_id 
              AND group_id = v_group_internal_id
              AND deleted_at IS NULL
        ),
        'last_modified', updated_at
    ) INTO v_group_info
    FROM cluster.group
    WHERE tenant_id = p_tenant_id AND id = v_group_internal_id;
    
    -- Get permissions for group members if requested
    IF p_include_all_members THEN
        FOR v_member_record IN
            SELECT 
                CASE 
                    WHEN cgm.member_user_id IS NOT NULL THEN 
                        (SELECT record_id FROM iam."user" WHERE tenant_id = p_tenant_id AND id = cgm.member_user_id)
                    WHEN cgm.member_group_id IS NOT NULL THEN 
                        (SELECT record_id FROM cluster.group WHERE tenant_id = p_tenant_id AND id = cgm.member_group_id)
                END as member_record_id,
                CASE 
                    WHEN cgm.member_user_id IS NOT NULL THEN 'user'
                    ELSE 'group'
                END as member_type
            FROM cluster.group_member cgm
            WHERE cgm.tenant_id = p_tenant_id 
              AND cgm.group_id = v_group_internal_id
              AND cgm.deleted_at IS NULL
        LOOP
            IF v_member_record.member_type = 'user' THEN
                BEGIN
                    v_user_snapshot := iam.get_user_permissions_snapshot(
                        p_tenant_id,
                        v_member_record.member_record_id,
                        p_object_api_names,
                        false, -- Don't include group memberships to avoid recursion
                        p_include_permission_sources,
                        p_ttl_seconds
                    );
                    v_member_permissions := v_member_permissions || v_user_snapshot;
                EXCEPTION WHEN OTHERS THEN
                    -- Skip failed users
                    NULL;
                END;
            END IF;
        END LOOP;
    END IF;
    
    -- Build result
    v_result := jsonb_build_object(
        'group_info', v_group_info,
        'member_permissions', v_member_permissions,
        'total_members', jsonb_array_length(v_member_permissions),
        'generated_at', now()
    );
    
    RETURN v_result;
END;
$$;

-- Check if user permissions have changed since timestamp
-- Used for efficient change detection
CREATE OR REPLACE FUNCTION iam.check_user_permissions_changed(
    p_tenant_id UUID,
    p_user_id TEXT,
    p_since TIMESTAMPTZ,
    p_object_api_names TEXT[] DEFAULT NULL
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_has_changes BOOLEAN := false;
    v_last_change TIMESTAMPTZ;
    v_changed_objects TEXT[] := '{}';
    v_total_changes INTEGER := 0;
BEGIN
    -- Check for group membership changes
    SELECT 
        COUNT(*) > 0,
        MAX(o.created_at),
        array_agg(DISTINCT o.payload->>'group_id') FILTER (WHERE o.payload->>'group_id' IS NOT NULL)
    INTO v_has_changes, v_last_change, v_changed_objects
    FROM bootstrap.outbox o
    WHERE o.tenant_id = p_tenant_id
      AND o.event_type IN ('iam.group_member.added', 'iam.group_member.removed')
      AND o.payload->>'member_id' = p_user_id
      AND o.created_at > p_since
      AND o.status = 'done';
    
    -- Get total changes count
    SELECT COUNT(*)
    INTO v_total_changes
    FROM bootstrap.outbox o
    WHERE o.tenant_id = p_tenant_id
      AND o.event_type IN ('iam.group_member.added', 'iam.group_member.removed')
      AND o.payload->>'member_id' = p_user_id
      AND o.created_at > p_since
      AND o.status = 'done';
    
    -- Build result
    v_result := jsonb_build_object(
        'has_changes', v_has_changes,
        'last_change', v_last_change,
        'changed_objects', v_changed_objects,
        'total_changes', v_total_changes
    );
    
    RETURN v_result;
END;
$$;

-- Sync permissions for users in a specific group
-- Used for bulk permission updates after group changes
CREATE OR REPLACE FUNCTION iam.sync_group_permissions(
    p_tenant_id UUID,
    p_group_record_id TEXT,
    p_user_ids TEXT[] DEFAULT NULL,
    p_object_api_names TEXT[] DEFAULT NULL,
    p_force_refresh BOOLEAN DEFAULT false
)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    v_result JSONB := '{}';
    v_group_internal_id BIGINT;
    v_users_synced INTEGER := 0;
    v_permissions_updated INTEGER := 0;
    v_failed_user_ids TEXT[] := '{}';
    v_user_id TEXT;
BEGIN
    -- Get group internal ID
    SELECT id INTO v_group_internal_id
    FROM cluster.group
    WHERE tenant_id = p_tenant_id AND record_id = p_group_record_id;
    
    IF v_group_internal_id IS NULL THEN
        RAISE EXCEPTION 'Group not found: %', p_group_record_id;
    END IF;
    
    -- If no specific users provided, get all group members
    IF p_user_ids IS NULL THEN
        SELECT array_agg(
            CASE 
                WHEN cgm.member_user_id IS NOT NULL THEN 
                    (SELECT record_id FROM iam."user" WHERE tenant_id = p_tenant_id AND id = cgm.member_user_id)
                ELSE NULL
            END
        ) INTO p_user_ids
        FROM cluster.group_member cgm
        WHERE cgm.tenant_id = p_tenant_id 
          AND cgm.group_id = v_group_internal_id
          AND cgm.member_user_id IS NOT NULL
          AND cgm.deleted_at IS NULL;
    END IF;
    
    -- Sync permissions for each user
    FOREACH v_user_id IN ARRAY p_user_ids
    LOOP
        IF v_user_id IS NOT NULL THEN
            BEGIN
                -- Invalidate user cache if force refresh
                IF p_force_refresh THEN
                    PERFORM cache.invalidate_user_permissions_cache(p_tenant_id, 
                        (SELECT id FROM iam."user" WHERE tenant_id = p_tenant_id AND record_id = v_user_id));
                END IF;
                
                -- Get updated permissions (this will rebuild cache if needed)
                PERFORM iam.get_user_permissions_snapshot(
                    p_tenant_id,
                    v_user_id,
                    p_object_api_names,
                    false, -- Don't include group memberships
                    false, -- Don't include permission sources
                    3600   -- 1 hour TTL
                );
                
                v_users_synced := v_users_synced + 1;
                
                -- Count permissions updated (simplified)
                v_permissions_updated := v_permissions_updated + 1;
                
            EXCEPTION WHEN OTHERS THEN
                v_failed_user_ids := v_failed_user_ids || v_user_id;
            END;
        END IF;
    END LOOP;
    
    -- Build result
    v_result := jsonb_build_object(
        'users_synced', v_users_synced,
        'permissions_updated', v_permissions_updated,
        'failed_user_ids', v_failed_user_ids,
        'synced_at', now()
    );
    
    RETURN v_result;
END;
$$;

-- ========================================
-- EVENT-DRIVEN PERMISSION SYNC FUNCTIONS
-- ========================================

-- Process outbox events and trigger permission sync
-- This function is called by outbox workers after processing events
CREATE OR REPLACE FUNCTION iam.process_permission_sync_events()
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    v_event_record RECORD;
    v_affected_users TEXT[];
    v_group_id TEXT;
    v_user_id TEXT;
BEGIN
    -- Process group membership events
    FOR v_event_record IN
        SELECT 
            o.id,
            o.event_type,
            o.payload,
            o.headers
        FROM bootstrap.outbox o
        WHERE o.event_type IN (
            'iam.group_member.added',
            'iam.group_member.removed',
            'iam.permission_set.assigned_to_group',
            'iam.permission_set.unassigned_from_group'
        )
          AND o.status = 'processing'
    LOOP
        -- Extract affected user from event
        v_user_id := v_event_record.payload->>'member_id';
        v_group_id := v_event_record.payload->>'group_id';
        
        IF v_user_id IS NOT NULL THEN
            -- Add to affected users list
            v_affected_users := v_affected_users || v_user_id;
            
            -- Invalidate user's permission cache
            PERFORM cache.invalidate_user_permissions_cache(
                (v_event_record.headers->>'tenant_id')::uuid,
                (SELECT id FROM iam."user" 
                 WHERE tenant_id = (v_event_record.headers->>'tenant_id')::uuid 
                   AND record_id = v_user_id)
            );
        END IF;
        
        -- Create notification event for external services
        PERFORM bootstrap.create_outbox_event(
            'permission_sync',
            v_event_record.id::text,
            'iam.permissions.sync_required',
            jsonb_build_object(
                'tenant_id', v_event_record.headers->>'tenant_id',
                'affected_users', v_affected_users,
                'group_id', v_group_id,
                'event_type', v_event_record.event_type,
                'original_event_id', v_event_record.headers->>'event_id'
            ),
            v_event_record.headers
        );
    END LOOP;
END;
$$;

-- Create permission sync notification events
-- This function creates events that external services can subscribe to
CREATE OR REPLACE FUNCTION iam.create_permission_sync_event(
    p_tenant_id UUID,
    p_user_ids TEXT[],
    p_reason TEXT,
    p_original_event_id TEXT DEFAULT NULL
)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    v_event_id TEXT := encode(gen_random_bytes(16), 'hex');
BEGIN
    PERFORM bootstrap.create_outbox_event(
        'permission_sync',
        v_event_id,
        'iam.permissions.sync_required',
        jsonb_build_object(
            'tenant_id', p_tenant_id::text,
            'affected_users', p_user_ids,
            'reason', p_reason,
            'original_event_id', p_original_event_id,
            'sync_required_at', now()
        ),
        jsonb_build_object(
            'tenant_id', p_tenant_id::text,
            'event_id', v_event_id,
            'priority', 'high'
        )
    );
END;
$$;

-- ========================================
-- MONITORING AND ANALYTICS FUNCTIONS
-- ========================================

-- Get permission sync statistics
CREATE OR REPLACE FUNCTION iam.get_permission_sync_stats(
    p_tenant_id UUID DEFAULT NULL,
    p_since TIMESTAMPTZ DEFAULT now() - interval '1 day'
)
RETURNS JSONB
LANGUAGE plpgsql
STABLE
AS $$
DECLARE
    v_result JSONB := '{}';
    v_total_events INTEGER;
    v_sync_events INTEGER;
    v_failed_events INTEGER;
    v_avg_processing_time INTERVAL;
BEGIN
    -- Get total events
    SELECT COUNT(*)
    INTO v_total_events
    FROM bootstrap.outbox o
    WHERE (p_tenant_id IS NULL OR o.headers->>'tenant_id' = p_tenant_id::text)
      AND o.created_at > p_since;
    
    -- Get sync events
    SELECT COUNT(*)
    INTO v_sync_events
    FROM bootstrap.outbox o
    WHERE (p_tenant_id IS NULL OR o.headers->>'tenant_id' = p_tenant_id::text)
      AND o.event_type LIKE 'iam.permissions.%'
      AND o.created_at > p_since;
    
    -- Get failed events
    SELECT COUNT(*)
    INTO v_failed_events
    FROM bootstrap.outbox o
    WHERE (p_tenant_id IS NULL OR o.headers->>'tenant_id' = p_tenant_id::text)
      AND o.status = 'dead'
      AND o.created_at > p_since;
    
    -- Get average processing time
    SELECT AVG(o.published_at - o.created_at)
    INTO v_avg_processing_time
    FROM bootstrap.outbox o
    WHERE (p_tenant_id IS NULL OR o.headers->>'tenant_id' = p_tenant_id::text)
      AND o.status = 'done'
      AND o.published_at IS NOT NULL
      AND o.created_at > p_since;
    
    -- Build result
    v_result := jsonb_build_object(
        'total_events', v_total_events,
        'sync_events', v_sync_events,
        'failed_events', v_failed_events,
        'avg_processing_time', v_avg_processing_time,
        'period_start', p_since,
        'period_end', now()
    );
    
    RETURN v_result;
END;
$$;