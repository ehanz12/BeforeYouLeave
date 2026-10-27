package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func ToReminderResponse(r models.Reminder) response.Reminder {
	name := ""
	if r.Destination.Name != "" {
		name = r.Destination.Name
	}
	return response.Reminder{
		ID:              r.ID,
		UserID:          r.UserID,
		DestinationID:   r.DestinationID,
		DestinationName: name,
		ScheduleID:      r.ScheduleID,
		RemindForDate:   r.RemindForDate,
		RemindAt:        r.RemindAt,
		Status:          r.Status,
		SentAt:          r.SentAt,
		ReadAt:          r.ReadAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

func ToReminderDetailResponse(r models.Reminder) response.ReminderDetail {
	checks := make([]response.ReminderItemCheck, 0, len(r.ItemChecks))
	progress := response.ReminderProgress{}
	for _, c := range r.ItemChecks {
		checks = append(checks, response.ReminderItemCheck{
			ChecklistItemID: c.ChecklistItemID,
			IsChecked:       c.IsChecked,
			CheckedAt:       c.CheckedAt,
			ChecklistItem:   ToChecklistItemResponse(c.ChecklistItem),
		})
		if c.IsChecked {
			progress.Checked++
		}
		progress.Total++
	}

	return response.ReminderDetail{
		Reminder:    ToReminderResponse(r),
		Destination: ToDestinationResponse(r.Destination),
		ItemChecks:  checks,
		Progress:    progress,
	}
}
