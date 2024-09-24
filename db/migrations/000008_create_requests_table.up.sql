BEGIN;


CREATE TYPE message_type AS ENUM (
    'PUT_VALUE',
    'GET_VALUE',
    'ADD_PROVIDER',
    'GET_PROVIDERS',
    'FIND_NODE',
    'PING'
);

COMMENT ON TYPE message_type IS ''
    'The different types of messages from https://github.com/libp2p/go-libp2p-kad-dht/blob/master/pb/dht.proto#L15-L21.';

CREATE TABLE requests
(
    -- An internal unique id that identifies a crawl.
    id              INT GENERATED ALWAYS AS IDENTITY,
    -- Timestamp of when this request started.
    timestamp       TIMESTAMPTZ NOT NULL,
    -- The message type of this request
    request_type    message_type NOT NULL,
    -- Peer ID of the ant doing the request,
    ant_id         INT NOT NULL,
    -- The peer related to this request
    peer_id         INT NOT NULL,
    -- The key ID of this request (?)
    key_id          INT NOT NULL,
    -- An array of all multi address IDs of the remote peer.
    multi_address_ids INT[],

    CONSTRAINT fk_requests_ant_id FOREIGN KEY (ant_id) REFERENCES keys (id) ON DELETE SET NULL,
    CONSTRAINT fk_requests_peer_id FOREIGN KEY (peer_id) REFERENCES peers (id) ON DELETE SET NULL,
    CONSTRAINT fk_requests_key_id FOREIGN KEY (key_id) REFERENCES keys (id) ON DELETE SET NULL,

    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

CREATE INDEX idx_requests_timestamp ON requests (timestamp);

COMMIT;
