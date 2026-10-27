package services

import (
	"errors"
	"time"

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

	// Destination baru aktif secara default; field tidak dikirim berarti aktif,
	// tapi is_active=false yang dikirim eksplisit harus dihormati.
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	destination := models.Destination{
		Name:        req.Name,
		UserID:      userID,
		Description: &req.Description,
		Longitude:   &req.Longitude,
		Latitude:    &req.Latitude,
		Radius:      &req.Radius,
		IsActive:    isActive,
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

func ListDestinations(req request.ListDestinationRequest, page, limit int, UserID uint64) ([]models.Destination, int64, error) {
	query := databases.DB.Model(&models.Destination{}).Where("user_id = ?", UserID)

	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}
	if req.StartDate != "" {
		start, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, 0, errors.New("format start_date tidak valid (YYYY-MM-DD) !")
		}
		query = query.Where("created_at >= ?", start)
	}
	if req.EndDate != "" {
		end, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, 0, errors.New("format end_date tidak valid (YYYY-MM-DD) !")
		}
		query = query.Where("created_at < ?", end.AddDate(0, 0, 1))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}

	var destinations []models.Destination
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&destinations).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}
	return destinations, total, nil
}

func DetailDestination(DesID uint64, UserID uint64) (models.Destination, error) {
	var des models.Destination
	if err := databases.DB.
		Preload("Schedules").
		Preload("ChecklistItems").
		Preload("Reminders").
		Where("id = ? AND user_id = ?", DesID, UserID).
		First(&des).Error; err != nil {
		return models.Destination{}, errors.New("destinasi tidak ditemukan !")
	}
	return des, nil
}

func UpdateDestination(req request.UpdateDestinationRequest, DesID uint64, UserID uint64) (models.Destination, error) {
	var des models.Destination
	if err := databases.DB.Where("id = ? AND user_id = ?", DesID, UserID).First(&des).Error; err != nil {
		return models.Destination{}, errors.New("destinasi tidak ditemukan !")
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		if *req.Name == "" {
			return models.Destination{}, errors.New("nama harus diisi !")
		}
		var exits models.Destination
		if err := databases.DB.Select("id").Where("name = ? AND user_id = ? AND id != ?", *req.Name, UserID, DesID).First(&exits).Error; err == nil {
			return models.Destination{}, errors.New("nama sudah dipakai")
		}
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Latitude != nil {
		updates["latitude"] = *req.Latitude
	}
	if req.Longitude != nil {
		updates["longitude"] = *req.Longitude
	}
	if req.Radius != nil {
		updates["radius"] = *req.Radius
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return models.Destination{}, errors.New("tidak ada field yang diupdate !")
	}

	if err := databases.DB.Model(&des).Updates(updates).Error; err != nil {
		return models.Destination{}, errors.New("terjadi kesalahan sistem !")
	}

	if err := databases.DB.Preload("Schedules").Preload("ChecklistItems").Preload("Reminders").First(&des).Error; err != nil {
		return models.Destination{}, errors.New("terjadi kesalahan sistem !")
	}
	return des, nil
}

func DeleteDestination(DesID uint64, UserID uint64) error {
	var des models.Destination
	if err := databases.DB.Where("id = ? AND user_id = ?", DesID, UserID).First(&des).Error; err != nil {
		return errors.New("destinasi tidak ditemukan !")
	}
	if err := databases.DB.Delete(&des).Error; err != nil {
		return errors.New("terjadi kesalahan sistem !")
	}
	return nil
}
