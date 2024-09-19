BEGIN;

CREATE TABLE keys
(
    id               INT GENERATED ALWAYS AS IDENTITY,
    -- The peer ID in the form of Qm... or 12D3...
    multi_hash       TEXT        NOT NULL CHECK ( TRIM(multi_hash) != '' ),

    PRIMARY KEY (id)
);

COMMIT;
