BEGIN;

-- Create the multi_address_sets table
CREATE TABLE multi_address_sets
(
    -- A unique id that identifies this set.
    id               INT GENERATED ALWAYS AS IDENTITY,
    PRIMARY KEY (id)
);

COMMENT ON TABLE multi_address_sets IS ''
        'Holds the unique identifier (id) for each set.'
        'This id is then used by the many-to-many multi_address_sets_addresses table.';
COMMENT ON COLUMN multi_address_sets.id IS 'An internal unique id that identifies this multi address set.';

-- Create the join table to associate multi_address_sets with multi_addresses
CREATE TABLE multi_address_sets_addresses
(
    -- The ID of the set
    set_id           INT NOT NULL,
    -- The ID of the multi address
    multi_address_id INT NOT NULL,

    -- Ensure that the same multi address can't be associated with the same set multiple times
    CONSTRAINT uq_multi_address_set UNIQUE (set_id, multi_address_id),

    -- Foreign key to the multi_address_sets table
    CONSTRAINT fk_multi_address_sets FOREIGN KEY (set_id)
        REFERENCES multi_address_sets (id) ON DELETE CASCADE,

    -- Foreign key to the multi_addresses table
    CONSTRAINT fk_multi_addresses FOREIGN KEY (multi_address_id)
        REFERENCES multi_addresses (id) ON DELETE CASCADE
);

COMMENT ON TABLE multi_address_sets_addresses IS ''
        'Serves as the association (join) table between multi_address_sets and multi_addresses.';
COMMENT ON COLUMN multi_address_sets_addresses.set_id IS 'The ID of the set to associate.';
COMMENT ON COLUMN multi_address_sets_addresses.multi_address_id IS 'THE ID of the multi address to associate with the set.';

COMMIT;
