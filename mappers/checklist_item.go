package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func ToChecklistItemResponse(i models.ChecklistItem) response.ChecklistItem {
	return response.ChecklistItem{
		ID:            i.ID,
		DestinationID: i.DestinationID,
		Name:          i.Name,
		IsRequired:    i.IsRequired,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}
