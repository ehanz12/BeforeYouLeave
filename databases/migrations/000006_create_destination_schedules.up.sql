-- Pengganti departure_profiles: 1 row = 1 hari aktif.
-- Contoh Sekolah Senin-Jumat = 5 rows. Tidak ada row Minggu = tidak ada reminder Minggu.
-- Alasan DROP + CREATE fresh (bukan RENAME): tabel lama tidak punya day_of_week
-- sehingga tidak bisa di-backfill dengan benar. Project masih awal, jadi aman.
DROP TABLE IF EXISTS departure_profiles;

CREATE TABLE destination_schedules(
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  destination_id BIGINT NOT NULL,
  day_of_week TINYINT NOT NULL COMMENT '0=Minggu, 1=Senin, ..., 6=Sabtu',
  departure_time TIME NOT NULL COMMENT 'jam berangkat hari itu, boleh beda tiap hari',
  remind_d_minus TINYINT NOT NULL DEFAULT 1 COMMENT '1=H-1 malam, 0=hari-H',
  remind_time TIME NOT NULL COMMENT 'jam bunyi reminder, misal 20:00 untuk packing malam',
  is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'toggle libur per-hari: FALSE = hari itu skip',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_schedules_destination
    FOREIGN KEY (destination_id) REFERENCES destinations(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,

  CONSTRAINT uq_schedules_destination_day UNIQUE (destination_id, day_of_week),
  CONSTRAINT chk_schedules_day CHECK (day_of_week BETWEEN 0 AND 6),
  CONSTRAINT chk_schedules_d_minus CHECK (remind_d_minus IN (0, 1))
);

CREATE INDEX idx_schedules_lookup ON destination_schedules(destination_id, day_of_week, is_active);
