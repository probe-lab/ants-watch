BEGIN;

CREATE TABLE keys
(
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- Use peer ID for keys that are also peers
    peer_id          INT,
    -- The peer ID in the form of Qm... or 12D3...
    multi_hash       TEXT        NOT NULL CHECK ( TRIM(multi_hash) != '' ),

    PRIMARY KEY (id),

    CONSTRAINT fk_keys_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE SET NULL,
    CONSTRAINT chk_keys_multi_hash_or_peer_id CHECK (peer_id IS NOT NULL OR (multi_hash IS NOT NULL AND TRIM(multi_hash) != ''))
);

COMMIT;
