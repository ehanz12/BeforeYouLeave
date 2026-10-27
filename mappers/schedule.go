package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func ToScheduleResponse(s models.DestinationSchedule) response.Schedule {
	return response.Schedule{
		ID:            s.ID,
		DestinationID: s.DestinationID,
		DayOfWeek:     s.DayOfWeek,
		DepartureTime: s.DepartureTime,
		RemindDMinus:  s.RemindDMinus,
		RemindTime:    s.RemindTime,
		IsActive:      s.IsActive,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}
