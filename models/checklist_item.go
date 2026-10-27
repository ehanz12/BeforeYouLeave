package models

import "time"

// ChecklistItem = TEMPLATE barang bawaan (tidak berubah per tanggal).
// Status centangan harian disimpan di ReminderItemCheck.
// IsRequired tidak memakai tag `default:` (lihat catatan di destination.go),
// supaya nilai false yang dikirim user tidak digantikan DEFAULT database.
type ChecklistItem struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	DestinationID uint64    `gorm:"not null;index" json:"destination_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	IsRequired    bool      `gorm:"not null" json:"is_required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Destination Destination `gorm:"foreignKey:DestinationID" json:"destination,omitempty"`
}

func (ChecklistItem) TableName() string { return "checklist_items" }
