BEGIN;

CREATE TABLE requests_denormalized
(
    -- An internal unique id that identifies a crawl.
    id              INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this request started.
    timestamp       TIMESTAMPTZ NOT NULL,
    -- The message type of this request
    request_type    message_type NOT NULL,
    -- Peer ID of the ant doing the request,
    ant_id          TEXT NOT NULL,
    -- The peer related to this request
    peer_id         TEXT NOT NULL,
    -- The key ID of this request (?)
    key_id          TEXT NOT NULL,
    -- An array of all multi address IDs of the remote peer.
    multi_address_ids TEXT[],

    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_requests_dnf_timestamp ON requests (timestamp);

COMMIT;
