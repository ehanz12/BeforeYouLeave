package request

// IsActive harus pointer: kalau bertipe bool biasa, field yang tidak dikirim
// jadi false dan destination ter-create dalam keadaan libur, jadi tidak ada
// reminder yang pernah di-generate. Pointer membedakan "tidak diisi" dari false.
type DestinationRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Radius      float64  `json:"radius"`
	IsActive    *bool    `json:"is_active"`
}

type UpdateDestinationRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	Radius      *float64 `json:"radius"`
	IsActive    *bool    `json:"is_active"`
}

type ListDestinationRequest struct {
	Search    string `query:"search"`
	IsActive  *bool  `query:"is_active"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
}
