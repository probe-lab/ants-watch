BEGIN;

CREATE OR REPLACE FUNCTION upsert_peer(
    new_multi_hash TEXT,
    new_agent_version_id INT DEFAULT NULL,
    new_protocols_set_id INT DEFAULT NULL,
    new_created_at TIMESTAMPTZ DEFAULT NOW(),
    new_last_seen_at TIMESTAMPTZ DEFAULT NOW()
) RETURNS INT AS
$upsert_peer$
    WITH ups AS (
        INSERT INTO peers AS p (multi_hash, agent_version_id, protocols_set_id, created_at, updated_at, last_seen_at)
        VALUES (new_multi_hash, new_agent_version_id, new_protocols_set_id, new_created_at, new_created_at, new_last_seen_at)
        ON CONFLICT ON CONSTRAINT uq_peers_multi_hash DO UPDATE
            SET multi_hash       = EXCLUDED.multi_hash,
                agent_version_id = COALESCE(EXCLUDED.agent_version_id, p.agent_version_id),
                protocols_set_id = COALESCE(EXCLUDED.protocols_set_id, p.protocols_set_id),
                updated_at       = EXCLUDED.updated_at,
                last_seen_at     = EXCLUDED.last_seen_at
        RETURNING id, multi_hash
    )
    SELECT id FROM ups;

$upsert_peer$ LANGUAGE sql;

COMMIT;
