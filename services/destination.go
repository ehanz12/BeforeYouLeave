package services

import (
	"errors"

	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/models"
)

func CreateDestination(req request.DestinationRequest) (models.Destination, error) {
	tx := databases.DB.Begin()
	if tx.Error != nil {
		return models.Destination{}, errors.New("terjadi kesalahan sistem !")
	}
	if req.Name == "" || req.IsActive == nil {
		return models.Destination{}, errors.New("nama dan active harus diisi !")
	}
}
