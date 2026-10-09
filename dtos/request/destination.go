package request

type DestinationRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Radius      float64 `json:"radius"`
	IsActive    bool    `json:"is_active"`
}
