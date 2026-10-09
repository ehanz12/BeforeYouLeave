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
	if req.Name == "" {
		return models.Destination{}, errors.New("nama harus diisi !")
	}
	var exits models.Destination
	if err := tx.Select("id", "name", "user_id").Where("name = ? AND user_id = ?", req.Name, userID).First(&exits).Error; err == nil {
		tx.Rollback()
		return models.Destination{}, errors.New("nama sudah dipakai")
	}

	destination := models.Destination{
		Name:        req.Name,
		UserID:      userID,
		Description: &req.Description,
		Longitude:   &req.Longitude,
		Latitude:    &req.Latitude,
		Radius:      &req.Radius,
		IsActive:    req.IsActive,
	}
	if err := tx.Create(&destination).Error; err != nil {
		tx.Rollback()
		return models.Destination{}, errors.New("terjadi kesalahan sistem")
	}
	if err := tx.Commit().Error; err != nil {
		return models.Destination{}, errors.New("terjadi kesalahan sistem !")
	}
	return destination, nil
}
