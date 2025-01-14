CREATE TABLE ip_information
(
    timestamp          DateTime,
    remote_multihash   AggregateFunction(uniqExact, String),
    agent_version_type LowCardinality(String),
    country            LowCardinality(String),
    continent          LowCardinality(String),
    city               LowCardinality(String),
    geo_hash           String,
    is_cloud           Bool,
    is_vpn             Bool,
    is_tor             Bool,
    is_proxy           Bool,
    is_bogon           Bool,
    is_relay           Bool,
    asn                Int,
    company            String,
    is_mobile          Bool
) ENGINE = ReplicatedAggregatingMergeTree()
    PRIMARY KEY (
        timestamp,
        agent_version_type,
        country,
        continent,
        city,
        geo_hash,
        is_cloud,
        is_vpn,
        is_tor,
        is_proxy,
        is_bogon,
        is_relay,
        asn,
        company,
        is_mobile
    )