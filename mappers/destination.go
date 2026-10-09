package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func ToDestinationResponse(d models.Destination) response.Destination {
	return response.Destination{
		ID:          d.ID,
		Name:        d.Name,
		Description: *d.Description,
		Latitude:    *d.Latitude,
		Longitude:   *d.Longitude,
		Radius:      *d.Radius,
		IsActive:    d.IsActive,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
