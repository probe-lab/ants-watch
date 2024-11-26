CREATE TABLE requests
(
    id               UUID,
    ant_multihash    String,
    remote_multihash String,
    agent_version    Nullable(String),
    protocols        Nullable(Array(String)),
    started_at       DateTime64(3),
    request_type     Enum8(
                         'PUT_VALUE',
                         'GET_VALUE',
                         'ADD_PROVIDER',
                         'GET_PROVIDERS',
                         'FIND_NODE',
                         'PING'
                         ),
    key_multihash    String,
    multi_addresses  Array(String)
) ENGINE = MergeTree()
    PRIMARY KEY (started_at)
TTL started_at + INTERVAL 1 DAY;
