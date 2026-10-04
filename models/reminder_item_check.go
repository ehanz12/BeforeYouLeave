package models

import "time"

// ReminderItemCheck = snapshot centangan per tanggal reminder.
// Dipakai untuk histori / pembiasaan ("Senin lalu lupa bawa apa").
// Dibuat saat worker membuat Reminder: copy semua ChecklistItem
// destination itu dengan IsChecked=FALSE.
type ReminderItemCheck struct {
	ReminderID      uint64     `gorm:"primaryKey" json:"reminder_id"`
	ChecklistItemID uint64     `gorm:"primaryKey;index:idx_ric_item" json:"checklist_item_id"`
	IsChecked       bool       `gorm:"not null;default:false" json:"is_checked"`
	CheckedAt       *time.Time `json:"checked_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	Reminder      Reminder      `gorm:"foreignKey:ReminderID" json:"reminder,omitempty"`
	ChecklistItem ChecklistItem `gorm:"foreignKey:ChecklistItemID" json:"checklist_item,omitempty"`
}

func (ReminderItemCheck) TableName() string { return "reminder_item_checks" }
