BEGIN;
CREATE OR REPLACE FUNCTION insert_key(peer_id INT, multi_hash TEXT)
RETURNS TABLE (inserted_id INT) AS $insert_key$
DECLARE
    new_id INT;
BEGIN
    IF peer_id IS NULL AND (multi_hash IS NULL OR TRIM(multi_hash) = '') THEN
        RAISE EXCEPTION 'Either peer_id or non-empty multi_hash must be provided';
    END IF;

    INSERT INTO keys (peer_id, multi_hash)
    VALUES (peer_id, multi_hash)
    ON CONFLICT DO NOTHING
    RETURNING id INTO new_id;

    inserted_id := new_id;
    RETURN NEXT;
END;
$insert_key$ LANGUAGE plpgsql;
COMMIT;
