package request

type ListReminderRequest struct {
	Status        string `query:"status"`
	DestinationID uint64 `query:"destination_id"`
	// UnreadOnly = true berarti hanya pending + sent (yang masih bunyi).
	UnreadOnly bool   `query:"unread_only"`
	StartDate  string `query:"start_date"`
	EndDate    string `query:"end_date"`
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}

// UpdateItemCheckRequest = centang / uncentang satu barang pada sebuah reminder.
type UpdateItemCheckRequest struct {
	IsChecked bool `json:"is_checked"`
}
