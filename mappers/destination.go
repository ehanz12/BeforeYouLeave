package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func derefString(v *string, fallback string) string {
	if v == nil {
		return fallback
	}
	return *v
}

func derefFloat(v *float64, fallback float64) float64 {
	if v == nil {
		return fallback
	}
	return *v
}

func ToDestinationResponse(d models.Destination) response.Destination {
	return response.Destination{
		ID:          d.ID,
		Name:        d.Name,
		Description: derefString(d.Description, ""),
		Latitude:    derefFloat(d.Latitude, 0),
		Longitude:   derefFloat(d.Longitude, 0),
		Radius:      derefFloat(d.Radius, 0),
		IsActive:    d.IsActive,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToDestinationDetailResponse(d models.Destination) response.DestinationDetail {
	schedules := make([]response.Schedule, 0, len(d.Schedules))
	for _, s := range d.Schedules {
		schedules = append(schedules, response.Schedule{
			ID:            s.ID,
			DayOfWeek:     s.DayOfWeek,
			DepartureTime: s.DepartureTime,
			RemindDMinus:  s.RemindDMinus,
			RemindTime:    s.RemindTime,
			IsActive:      s.IsActive,
			CreatedAt:     s.CreatedAt,
			UpdatedAt:     s.UpdatedAt,
		})
	}

	items := make([]response.ChecklistItem, 0, len(d.ChecklistItems))
	for _, i := range d.ChecklistItems {
		items = append(items, response.ChecklistItem{
			ID:         i.ID,
			Name:       i.Name,
			IsRequired: i.IsRequired,
			CreatedAt:  i.CreatedAt,
			UpdatedAt:  i.UpdatedAt,
		})
	}

	reminders := make([]response.Reminder, 0, len(d.Reminders))
	for _, r := range d.Reminders {
		reminders = append(reminders, response.Reminder{
			ID:            r.ID,
			ScheduleID:    r.ScheduleID,
			RemindForDate: r.RemindForDate,
			RemindAt:      r.RemindAt,
			Status:        r.Status,
			SentAt:        r.SentAt,
			ReadAt:        r.ReadAt,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
		})
	}

	return response.DestinationDetail{
		Destination:    ToDestinationResponse(d),
		Schedules:      schedules,
		ChecklistItems: items,
		Reminders:      reminders,
	}
}
