ALTER TABLE requests
    ADD COLUMN agent_version_type   LowCardinality(String) AFTER agent_version,
    ADD COLUMN agent_version_semver Array(Int16) AFTER agent_version_type;
