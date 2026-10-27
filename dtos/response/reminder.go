package response

import "time"

// Reminder = satu instance pengingat untuk satu tanggal keberangkatan.
// Dipakai di list. DestinationName ikut agar FE tidak perlu request terpisah.
type Reminder struct {
	ID              uint64     `json:"id"`
	UserID          uint64     `json:"user_id"`
	DestinationID   uint64     `json:"destination_id"`
	DestinationName string     `json:"destination_name"`
	ScheduleID      uint64     `json:"schedule_id"`
	RemindForDate   time.Time  `json:"remind_for_date"`
	RemindAt        time.Time  `json:"remind_at"`
	Status          string     `json:"status"`
	SentAt          *time.Time `json:"sent_at"`
	ReadAt          *time.Time `json:"read_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ReminderItemCheck = satu barang dalam satu reminder beserta status centangnya.
type ReminderItemCheck struct {
	ChecklistItemID uint64        `json:"checklist_item_id"`
	IsChecked       bool          `json:"is_checked"`
	CheckedAt       *time.Time    `json:"checked_at"`
	ChecklistItem   ChecklistItem `json:"checklist_item"`
}

// ReminderProgress = rekap centangan, dipakai FE buat progress bar.
type ReminderProgress struct {
	Total   int `json:"total"`
	Checked int `json:"checked"`
}

// ReminderDetail = reminder lengkap dengan daftar barangnya.
type ReminderDetail struct {
	Reminder
	Destination Destination         `json:"destination"`
	ItemChecks  []ReminderItemCheck `json:"item_checks"`
	Progress    ReminderProgress    `json:"progress"`
}
