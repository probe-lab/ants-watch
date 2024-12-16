CREATE TABLE node_population (
    timestamp            DateTime,
    remote_multihash     AggregateFunction(uniqExact, String),
    agent_version_type   LowCardinality(String),
    agent_version_semver Array(Int16)
) ENGINE = AggregatingMergeTree()
    PRIMARY KEY (timestamp, agent_version_type)