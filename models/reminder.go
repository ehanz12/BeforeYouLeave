package models

import "time"

// Status reminder: pending=antre, sent=sudah bunyi tapi belum dibaca
// (FE ulang bunyi / tampil sticky), read=ditandai baca (dismiss),
// skipped=libur / dibatalkan manual.
const (
	ReminderPending = "pending"
	ReminderSent    = "sent"
	ReminderRead    = "read"
	ReminderSkipped = "skipped"
)

// Reminder = instance yang di-generate worker dari DestinationSchedule.
// RemindForDate = tanggal keberangkatan (misal Senin 2026-10-06).
// RemindAt = kapan bunyi (misal Minggu 2026-10-05 20:00 untuk H-1).
type Reminder struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64     `gorm:"not null;index:idx_reminders_due,priority:1" json:"user_id"`
	DestinationID uint64     `gorm:"not null;index:idx_reminders_destination,priority:1" json:"destination_id"`
	ScheduleID    uint64     `gorm:"not null;uniqueIndex:uq_reminders_schedule_for_date,priority:1" json:"schedule_id"`
	RemindForDate time.Time  `gorm:"type:date;not null;uniqueIndex:uq_reminders_schedule_for_date,priority:2;index:idx_reminders_destination,priority:2" json:"remind_for_date"`
	RemindAt      time.Time  `gorm:"type:datetime;not null;index:idx_reminders_due,priority:3" json:"remind_at"`
	Status        string     `gorm:"type:enum('pending','sent','read','skipped');not null;default:'pending';index:idx_reminders_due,priority:2" json:"status"`
	SentAt        *time.Time `json:"sent_at"`
	ReadAt        *time.Time `json:"read_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Destination Destination         `gorm:"foreignKey:DestinationID" json:"destination,omitempty"`
	Schedule    DestinationSchedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	ItemChecks  []ReminderItemCheck `gorm:"foreignKey:ReminderID" json:"item_checks,omitempty"`
}

// MarkAsRead = "cara matiin reminder adalah ditandai baca".
func (r *Reminder) MarkAsRead() {
	now := time.Now()
	r.Status = ReminderRead
	r.ReadAt = &now
}

// StillRinging = reminder tetap bunyi selama belum dibaca.
func (r *Reminder) StillRinging() bool {
	return r.Status == ReminderPending || r.Status == ReminderSent
}

func (Reminder) TableName() string { return "reminders" }
