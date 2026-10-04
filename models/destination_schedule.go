package models

import "time"

// DayOfWeek: 0=Minggu, 1=Senin, ..., 6=Sabtu (ikut konvensi time.Weekday).
const (
	Sunday    = 0
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
	Saturday  = 6
)

// RemindDMinus: 1 = H-1 malam (default), 0 = hari-H.
const (
	RemindHMinus1 = 1
	RemindSameDay = 0
)

// DestinationSchedule = 1 row untuk 1 hari aktif.
// Contoh: Sekolah Senin-Jumat = 5 rows, tanpa row Minggu = Minggu libur, tanpa reminder.
// Jam boleh beda tiap hari (Senin 07:00, Jumat 06:30).
type DestinationSchedule struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	DestinationID uint64    `gorm:"not null;uniqueIndex:uq_schedules_destination_day,priority:1;index:idx_schedules_lookup,priority:1" json:"destination_id"`
	DayOfWeek     int8      `gorm:"type:tinyint;not null;uniqueIndex:uq_schedules_destination_day,priority:2;index:idx_schedules_lookup,priority:2;check:day_of_week BETWEEN 0 AND 6" json:"day_of_week"`
	DepartureTime string    `gorm:"type:time;not null" json:"departure_time"` // "07:00:00"
	RemindDMinus  int8      `gorm:"type:tinyint;not null;default:1" json:"remind_d_minus"`
	RemindTime    string    `gorm:"type:time;not null" json:"remind_time"` // "20:00:00"
	IsActive      bool      `gorm:"not null;default:true;index:idx_schedules_lookup,priority:3" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Destination Destination `gorm:"foreignKey:DestinationID" json:"destination,omitempty"`
	Reminders   []Reminder  `gorm:"foreignKey:ScheduleID" json:"reminders,omitempty"`
}

func (DestinationSchedule) TableName() string { return "destination_schedules" }
