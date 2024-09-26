BEGIN;

CREATE OR REPLACE FUNCTION insert_key(peer_id INT, multi_hash TEXT)
RETURNS RECORD AS $insert_key$
BEGIN
    -- Ensure either peer_id or multi_hash is provided, following the CHECK constraint
    IF peer_id IS NULL AND (multi_hash IS NULL OR TRIM(multi_hash) = '') THEN
        RAISE EXCEPTION 'Either peer_id or non-empty multi_hash must be provided';
    END IF;

    -- Insert the new row into the keys table
    INSERT INTO keys (peer_id, multi_hash)
    VALUES (peer_id, multi_hash);
END;
$insert_key$ LANGUAGE plpgsql;

COMMIT;
