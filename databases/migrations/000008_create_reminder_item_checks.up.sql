-- Snapshot centangan per tanggal reminder.
-- checklist_items = TEMPLATE barang (tidak berubah).
-- Tabel ini = HISTORI: Senin 2026-10-06 checklist apa saja yang sudah di-packing.
-- Saran pemakaian: saat worker membuat reminders, copy semua checklist_items
-- destination itu ke tabel ini dengan is_checked=FALSE.
CREATE TABLE reminder_item_checks(
  reminder_id BIGINT NOT NULL,
  checklist_item_id BIGINT NOT NULL,
  is_checked BOOLEAN NOT NULL DEFAULT FALSE,
  checked_at TIMESTAMP NULL DEFAULT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (reminder_id, checklist_item_id),

  CONSTRAINT fk_ric_reminder
    FOREIGN KEY (reminder_id) REFERENCES reminders(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
  CONSTRAINT fk_ric_checklist_item
    FOREIGN KEY (checklist_item_id) REFERENCES checklist_items(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);

CREATE INDEX idx_ric_item ON reminder_item_checks(checklist_item_id);
