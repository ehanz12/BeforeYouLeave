-- Tambah master toggle libur: FALSE = tidak diingatkan sama sekali.
ALTER TABLE destinations
  ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE AFTER radius;

-- Fix typo latidude -> latitude + tipe GPS yang benar (DECIMAL, bukan FLOAT).
-- FLOAT(10,8) tidak cocok untuk koordinat, pakai DECIMAL(10,8).
ALTER TABLE destinations
  CHANGE COLUMN latidude latitude DECIMAL(10, 8) NULL,
  MODIFY COLUMN longitude DECIMAL(11, 8) NULL,
  MODIFY COLUMN radius DECIMAL(10, 2) NULL COMMENT 'radius geofence dalam meter';

CREATE INDEX idx_destinations_user_active ON destinations(user_id, is_active);
