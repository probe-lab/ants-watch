BEGIN;

-- The `peers` table keeps track of all peers ever found in the DHT
CREATE TABLE peers
(
    -- The peer ID as a database-friendly integer
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The current agent version of the peer (updated if changed).
    agent_version_id INT,
    -- The peer ID in the form of Qm... or 12D3...
    multi_hash       TEXT        NOT NULL CHECK ( TRIM(multi_hash) != '' ),

    -- When were the multi addresses updated the last time.
    updated_at       TIMESTAMPTZ NOT NULL CHECK ( updated_at >= created_at ),
    -- When was this peer instance created.
    -- This gives a pretty accurate idea of
    -- when this peer was seen the first time.
    created_at       TIMESTAMPTZ NOT NULL,

    -- When was the peer seen for the last time
    last_seen_at       TIMESTAMPTZ NOT NULL CHECK ( last_seen_at >= updated_at ),

    CONSTRAINT fk_peers_agent_version_id FOREIGN KEY (agent_version_id) REFERENCES agent_versions (id) ON DELETE SET NULL,

    -- There should only ever be distinct peer multi hash here
    CONSTRAINT uq_peers_multi_hash UNIQUE (multi_hash),

    PRIMARY KEY (id)
);

COMMIT;