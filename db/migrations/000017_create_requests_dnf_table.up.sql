BEGIN;

CREATE TABLE requests_denormalized
(
    -- An internal unique id that identifies a crawl.
    id                       BIGINT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this request started.
    request_started_at       TIMESTAMPTZ NOT NULL,
    -- The message type of this request
    request_type             message_type NOT NULL,
    -- Peer ID of the ant doing the request,
    ant_multihash            TEXT NOT NULL,
    -- The peer related to this request
    peer_multihash           TEXT NOT NULL,
    -- The key of this request
    key_multihash            TEXT NOT NULL,
    -- An array of all multi addresses of the remote peer.
    multi_addresses          TEXT[],

    agent_version            TEXT,

    normalized_at            TIMESTAMPTZ,

    PRIMARY KEY (id, request_started_at)
) PARTITION BY RANGE (request_started_at);

CREATE INDEX idx_requests_dnf_timestamp ON requests_denormalized (request_started_at);

COMMIT;
