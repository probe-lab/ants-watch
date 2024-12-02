CREATE TABLE requests
(
    id               UUID,
    ant_multihash    String,
    remote_multihash String,
    agent_version    String,
    protocols        Array(String),
    started_at       DateTime,
    request_type     String,
    key_multihash    String,
    multi_addresses  Array(String)
) ENGINE = ReplicatedMergeTree()
    PRIMARY KEY (started_at)
TTL started_at + INTERVAL 1 DAY;
