BEGIN;

ALTER TABLE requests_denormalized ADD COLUMN protocols TEXT[];

COMMIT;
