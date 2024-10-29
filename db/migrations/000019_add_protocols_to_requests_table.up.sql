BEGIN;

ALTER TABLE requests
      ADD COLUMN protocols_set_id INT,
      ADD CONSTRAINT fk_requests_protocols_set_id FOREIGN KEY (protocols_set_id) REFERENCES protocols_sets (id) ON DELETE SET NULL;

COMMIT;
