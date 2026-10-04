package models

import "time"

// User pemilik destinations.
type User struct {
	ID           uint64        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string        `gorm:"type:varchar(255);not null" json:"name"`
	Email        string        `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password     string        `gorm:"type:varchar(255);not null" json:"-"`
	Phone        *string       `gorm:"type:varchar(15)" json:"phone"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Destinations []Destination `gorm:"foreignKey:UserID" json:"destinations,omitempty"`
}

func (User) TableName() string { return "users" }
