BEGIN;

CREATE OR REPLACE FUNCTION upsert_peer(
    new_multi_hash TEXT,
    new_agent_version_id INT DEFAULT NULL,
    new_protocols_set_id INT DEFAULT NULL,
    new_created_at TIMESTAMPTZ DEFAULT NOW(),
    new_last_seen_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_peer$
    WITH sel AS (
        -- Select the peer if it exists by multi_hash
        SELECT id, multi_hash, agent_version_id, protocols_set_id
        FROM peers
        WHERE multi_hash = new_multi_hash
    ),
    ups AS (
        -- Insert a new peer if it doesn't exist
        INSERT INTO peers AS p (multi_hash, agent_version_id, protocols_set_id, created_at, updated_at, last_seen_at)
        SELECT new_multi_hash, new_agent_version_id, new_protocols_set_id, new_created_at, new_created_at, new_last_seen_at
        WHERE NOT EXISTS (SELECT NULL FROM sel)
        ON CONFLICT ON CONSTRAINT uq_peers_multi_hash DO UPDATE
            SET multi_hash       = EXCLUDED.multi_hash,
                agent_version_id = COALESCE(EXCLUDED.agent_version_id, p.agent_version_id),
                protocols_set_id = COALESCE(EXCLUDED.protocols_set_id, p.protocols_set_id),
                updated_at       = EXCLUDED.updated_at,
                last_seen_at     = EXCLUDED.last_seen_at
        RETURNING id, multi_hash
    ),
    upd AS (
        -- Update existing peer if there are changes in agent_version_id, protocols_set_id, or last_seen_at
        UPDATE peers
        SET agent_version_id = COALESCE(new_agent_version_id, agent_version_id),
            protocols_set_id = COALESCE(new_protocols_set_id, protocols_set_id),
            updated_at       = new_created_at,
            last_seen_at     = new_last_seen_at
        WHERE id = (SELECT id FROM sel) AND (
            COALESCE(agent_version_id, -1) != COALESCE(new_agent_version_id, -1) OR
            COALESCE(protocols_set_id, -1) != COALESCE(new_protocols_set_id, -1) OR
            last_seen_at != new_last_seen_at
        )
        RETURNING peers.id
    )
    -- Return the ID of the inserted or updated peer
    SELECT id FROM sel
    UNION
    SELECT id FROM ups
    UNION
    SELECT id FROM upd;
$upsert_peer$ LANGUAGE sql;

COMMIT;
