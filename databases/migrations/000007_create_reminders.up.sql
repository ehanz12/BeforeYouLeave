-- Instance reminder yang dibangkitkan worker dari destination_schedules.
-- Mendukung: reminder H-1 malam + "tetap bunyi sampai ditandai baca".
-- Cara matikan saat libur: jangan generate row (karena is_active=FALSE di-skip worker),
-- atau UPDATE status='skipped'.
CREATE TABLE reminders(
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  destination_id BIGINT NOT NULL,
  schedule_id BIGINT NOT NULL,
  remind_for_date DATE NOT NULL COMMENT 'tanggal keberangkatan, misal Senin 2026-10-06',
  remind_at DATETIME NOT NULL COMMENT 'kapan bunyi, misal Minggu 2026-10-05 20:00',
  status ENUM('pending','sent','read','skipped') NOT NULL DEFAULT 'pending'
    COMMENT 'pending=antre, sent=sudah bunyi tapi belum dibaca, read=ditandai baca (dismiss), skipped=libur',
  sent_at TIMESTAMP NULL DEFAULT NULL,
  read_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  CONSTRAINT fk_reminders_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
  CONSTRAINT fk_reminders_destination
    FOREIGN KEY (destination_id) REFERENCES destinations(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
  CONSTRAINT fk_reminders_schedule
    FOREIGN KEY (schedule_id) REFERENCES destination_schedules(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,

  -- Cegah dobel reminder untuk jadwal + tanggal berangkat yang sama.
  CONSTRAINT uq_reminders_schedule_for_date UNIQUE (schedule_id, remind_for_date)
);

-- Worker query: ambil yang jatuh tempo dan belum dibaca.
CREATE INDEX idx_reminders_due ON reminders(user_id, status, remind_at);
-- Riwayat per destination.
CREATE INDEX idx_reminders_destination ON reminders(destination_id, remind_for_date);
