CREATE MATERIALIZED VIEW node_population_mv TO node_population AS
SELECT
    toStartOfTenMinutes(started_at)  as timestamp,
    uniqExactState(remote_multihash) as remote_multihash,
    agent_version_type,
    agent_version_semver
FROM requests
GROUP BY timestamp, agent_version_type, agent_version_semver