ALTER TABLE requests
    DROP COLUMN IF EXISTS agent_version_type,
    DROP COLUMN IF EXISTS agent_version_semver;