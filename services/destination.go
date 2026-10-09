package services

import (
	"errors"

	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func CreateDestination(req request.DestinationRequest, userID uint64) (models.Destination, error) {
	tx := databases.DB.Begin()
	if tx.Error != nil {
		return models.Destination{}, errors.New("terjadi kesalahan sistem !")
	}
	if req.Name == "" || req.IsActive == nil {
		return models.Destination{}, errors.New("nama dan active harus diisi !")
	}
	if err := databases.DB.Select("id", "name", "user_id").Where("name = ? AND user_id = ?", req.Name, userID).Error; err == nil {
		return models.Destination{}, errors.New("nama sudah dipakai")
	}

	destination := models.Destination{
		Name:        req.Name,
		UserID:      userID,
		Description: &req.Description,
		Longitude:   &req.Latitude,
		Latitude:    &req.Latitude,
		Radius:      &req.Radius,
		IsActive:    *req.IsActive,
	}
	return destination, nil
}
