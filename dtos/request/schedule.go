package request

// DayOfWeek: 0=Minggu, 1=Senin, ..., 6=Sabtu (ikut konvensi time.Weekday).
type CreateScheduleRequest struct {
	DayOfWeek     int8   `json:"day_of_week"`
	DepartureTime string `json:"departure_time"` // "07:00" atau "07:00:00"
	RemindDMinus  *int8  `json:"remind_d_minus"` // 1=H-1 malam (default), 0=hari-H
	RemindTime    string `json:"remind_time"`    // "20:00" atau "20:00:00"
	IsActive      *bool  `json:"is_active"`
}

type UpdateScheduleRequest struct {
	DayOfWeek     *int8   `json:"day_of_week"`
	DepartureTime *string `json:"departure_time"`
	RemindDMinus  *int8   `json:"remind_d_minus"`
	RemindTime    *string `json:"remind_time"`
	IsActive      *bool   `json:"is_active"`
}

type ListScheduleRequest struct {
	IsActive *bool `query:"is_active"`
	Page     int   `query:"page"`
	Limit    int   `query:"limit"`
}

// ReplaceSchedulesRequest = set jadwal sepekan sekaligus (replace all).
// Dipakai form FE yang mencentang hari aktif sekaligus, misal Senin-Jumat = 5 rows.
type ReplaceSchedulesRequest struct {
	Schedules []CreateScheduleRequest `json:"schedules"`
}
