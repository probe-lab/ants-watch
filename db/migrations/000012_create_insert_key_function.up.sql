BEGIN;

CREATE OR REPLACE FUNCTION insert_key(key_multi_hash TEXT)
RETURNS INT AS
$insert_key$
DECLARE
    new_id       INT;
    key_peer_id  INT;
    key_model_id INT;
BEGIN
    SELECT id INTO key_model_id FROM keys k WHERE k.multi_hash = key_multi_hash;

    IF key_model_id IS NULL THEN
        SELECT id INTO key_peer_id FROM peers p WHERE p.multi_hash = key_multi_hash;

        IF key_peer_id IS NOT NULL THEN
            INSERT INTO keys (peer_id, multi_hash)
            VALUES (key_peer_id, NULL)
            ON CONFLICT DO NOTHING
            RETURNING id INTO new_id;
        ELSE
            INSERT INTO keys (peer_id, multi_hash)
            VALUES (NULL, key_multi_hash)
            ON CONFLICT DO NOTHING
            RETURNING id INTO new_id;
        END IF;
        ELSE

        new_id := key_model_id;
    END IF;

    RETURN new_id;
END;
$insert_key$ LANGUAGE plpgsql;

COMMIT;
