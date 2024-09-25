BEGIN;

CREATE OR REPLACE FUNCTION insert_request(
    -- new_request_id INT,
    new_timestamp TIMESTAMPTZ,
    new_request_type message_type,
    new_ant_id INT,
    new_peer_id INT,
    new_key_id INT,
    new_multi_addresses TEXT[],
    new_protocols_set_id INT,
    new_agent_version_id INT
) RETURNS RECORD AS
$insert_request$
DECLARE
    new_peer_id             INT;
    new_multi_addresses_ids INT[];
    new_ant_id          INT;
    new_request_id            INT;
BEGIN

    SELECT upsert_peer(new_peer_multi_hash, new_agent_version_id, new_protocols_set_id, new_timestamp)
    INTO new_peer_id;

    SELECT array_agg(id) FROM upsert_multi_addresses(new_multi_addresses) INTO new_multi_addresses_ids;

    DELETE
    FROM peers_x_multi_addresses pxma
    WHERE peer_id = new_peer_id;

    INSERT INTO peers_x_multi_addresses (peer_id, multi_address_id)
    SELECT new_peer_id, new_multi_address_id
    FROM unnest(new_multi_addresses_ids) new_multi_address_id
    ON CONFLICT DO NOTHING;

    -- Now we're able to create the normalized request instance
    INSERT INTO requests (timestamp, request_type, ant_id, peer_id, key_id, multi_address_ids)
    SELECT new_timestamp,
           new_request_type,
           new_ant_id,
           new_peer_id,
           new_multi_addresses_ids
    RETURNING id INTO new_request_id;

    RETURN ROW(new_peer_id, new_request_id, new_key_id);
END;
$insert_request$ LANGUAGE plpgsql;

COMMIT;
