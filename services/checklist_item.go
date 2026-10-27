package services

import (
	"errors"

	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/models"
)

// getOwnedDestination = cek destination benar-benar milik user.
// Dipakai sebelum akses item agar user tidak bisa sentuh destination orang lain.
func getOwnedDestination(destinationID, userID uint64) (models.Destination, error) {
	var des models.Destination
	if err := databases.DB.
		Select("id", "user_id", "name").
		Where("id = ? AND user_id = ?", destinationID, userID).
		First(&des).Error; err != nil {
		return models.Destination{}, errors.New("destinasi tidak ditemukan !")
	}
	return des, nil
}

// isItemNameTaken = nama item harus unik dalam satu destination.
// excludeID dipakai saat update supaya item tidak bentrok dengan dirinya sendiri.
func isItemNameTaken(destinationID uint64, name string, excludeID uint64) bool {
	var item models.ChecklistItem
	query := databases.DB.
		Select("id").
		Where("destination_id = ? AND name = ?", destinationID, name)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	return query.First(&item).Error == nil
}

func CreateChecklistItem(req request.CreateChecklistItemRequest, destinationID, userID uint64) (models.ChecklistItem, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return models.ChecklistItem{}, err
	}
	if req.Name == "" {
		return models.ChecklistItem{}, errors.New("nama harus diisi !")
	}
	if isItemNameTaken(destinationID, req.Name, 0) {
		return models.ChecklistItem{}, errors.New("nama sudah dipakai")
	}

	item := models.ChecklistItem{
		DestinationID: destinationID,
		Name:          req.Name,
		IsRequired:    req.IsRequired,
	}
	if err := databases.DB.Create(&item).Error; err != nil {
		return models.ChecklistItem{}, errors.New("terjadi kesalahan sistem")
	}
	return item, nil
}

func ListChecklistItems(req request.ListChecklistItemRequest, destinationID, userID uint64, page, limit int) ([]models.ChecklistItem, int64, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return nil, 0, err
	}

	query := databases.DB.Model(&models.ChecklistItem{}).Where("destination_id = ?", destinationID)
	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.IsRequired != nil {
		query = query.Where("is_required = ?", *req.IsRequired)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}

	var items []models.ChecklistItem
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, errors.New("terjadi kesalahan sistem !")
	}
	return items, total, nil
}

// getOwnedItem = item harus ada di destination itu, dan destination-nya milik user.
func getOwnedItem(itemID, destinationID, userID uint64) (models.ChecklistItem, error) {
	if _, err := getOwnedDestination(destinationID, userID); err != nil {
		return models.ChecklistItem{}, err
	}
	var item models.ChecklistItem
	if err := databases.DB.
		Where("id = ? AND destination_id = ?", itemID, destinationID).
		First(&item).Error; err != nil {
		return models.ChecklistItem{}, errors.New("item tidak ditemukan !")
	}
	return item, nil
}

func DetailChecklistItem(itemID, destinationID, userID uint64) (models.ChecklistItem, error) {
	return getOwnedItem(itemID, destinationID, userID)
}

func UpdateChecklistItem(req request.UpdateChecklistItemRequest, itemID, destinationID, userID uint64) (models.ChecklistItem, error) {
	item, err := getOwnedItem(itemID, destinationID, userID)
	if err != nil {
		return models.ChecklistItem{}, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		if *req.Name == "" {
			return models.ChecklistItem{}, errors.New("nama harus diisi !")
		}
		if isItemNameTaken(destinationID, *req.Name, item.ID) {
			return models.ChecklistItem{}, errors.New("nama sudah dipakai")
		}
		updates["name"] = *req.Name
	}
	if req.IsRequired != nil {
		updates["is_required"] = *req.IsRequired
	}

	if len(updates) == 0 {
		return models.ChecklistItem{}, errors.New("tidak ada field yang diupdate !")
	}

	if err := databases.DB.Model(&item).Updates(updates).Error; err != nil {
		return models.ChecklistItem{}, errors.New("terjadi kesalahan sistem !")
	}
	if err := databases.DB.First(&item, item.ID).Error; err != nil {
		return models.ChecklistItem{}, errors.New("terjadi kesalahan sistem !")
	}
	return item, nil
}

func DeleteChecklistItem(itemID, destinationID, userID uint64) error {
	item, err := getOwnedItem(itemID, destinationID, userID)
	if err != nil {
		return err
	}
	if err := databases.DB.Delete(&item).Error; err != nil {
		return errors.New("terjadi kesalahan sistem !")
	}
	return nil
}
