package request

type CreateChecklistItemRequest struct {
	Name       string `json:"name"`
	IsRequired bool   `json:"is_required"`
}

type UpdateChecklistItemRequest struct {
	Name       *string `json:"name"`
	IsRequired *bool   `json:"is_required"`
}

type ListChecklistItemRequest struct {
	Search     string `query:"search"`
	IsRequired *bool  `query:"is_required"`
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}
