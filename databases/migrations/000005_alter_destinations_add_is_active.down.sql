-- Rollback 000005: kembalikan skema destinations seperti semula.
ALTER TABLE destinations DROP INDEX idx_destinations_user_active;

ALTER TABLE destinations
  CHANGE COLUMN latitude latidude FLOAT(10, 8) NULL,
  MODIFY COLUMN longitude FLOAT(10, 8) NULL,
  MODIFY COLUMN radius FLOAT(10, 8) NULL;

ALTER TABLE destinations DROP COLUMN is_active;
