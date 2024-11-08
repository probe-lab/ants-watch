CREATE INDEX idx_normalized_at_is_null
ON requests_denormalized (normalized_at)
WHERE normalized_at IS NULL;
