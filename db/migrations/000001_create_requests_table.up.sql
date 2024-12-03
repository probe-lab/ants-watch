CREATE TABLE requests
(
    id               UUID,
    queen_id         UUID,
    ant_multihash    String,
    remote_multihash String,
    agent_version    String,
    protocols        Array(String),
    started_at       DateTime64(3),
    request_type     String,
    key_multihash    String,
    multi_addresses  Array(String)
) ENGINE = MergeTree
    PRIMARY KEY (started_at)
TTL toDateTime(started_at) + INTERVAL 1 DAY;
