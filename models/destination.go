package models

import "time"

// Destination tempat tujuan (misal Sekolah).
// IsActive = master toggle libur. FALSE = semua jadwal di bawahnya skip,
// tidak ada reminder yang di-generate.
type Destination struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"not null;index:idx_destinations_user_active" json:"user_id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	Latitude    *float64  `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude   *float64  `gorm:"type:decimal(11,8)" json:"longitude"`
	Radius      *float64  `gorm:"type:decimal(10,2)" json:"radius"`
	IsActive    bool      `gorm:"not null;default:true;index:idx_destinations_user_active" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Schedules      []DestinationSchedule `gorm:"foreignKey:DestinationID" json:"schedules,omitempty"`
	ChecklistItems []ChecklistItem       `gorm:"foreignKey:DestinationID" json:"checklist_items,omitempty"`
	Reminders      []Reminder            `gorm:"foreignKey:DestinationID" json:"reminders,omitempty"`
}

func (Destination) TableName() string { return "destinations" }
