package mappers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func ToDoUser(u models.User) response.AuthResponse {
	return response.AuthResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Phone: *u.Phone,
	}
}
