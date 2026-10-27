package response

import "time"

type Destination struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Radius      float64   `json:"radius"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Schedule struct {
	ID            uint64    `json:"id"`
	DestinationID uint64    `json:"destination_id"`
	DayOfWeek     int8      `json:"day_of_week"`
	DepartureTime string    `json:"departure_time"`
	RemindDMinus  int8      `json:"remind_d_minus"`
	RemindTime    string    `json:"remind_time"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ChecklistItem struct {
	ID            uint64    `json:"id"`
	DestinationID uint64    `json:"destination_id"`
	Name          string    `json:"name"`
	IsRequired    bool      `json:"is_required"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DestinationDetail struct {
	Destination
	Schedules      []Schedule      `json:"schedules"`
	ChecklistItems []ChecklistItem `json:"checklist_items"`
	Reminders      []Reminder      `json:"reminders"`
}
