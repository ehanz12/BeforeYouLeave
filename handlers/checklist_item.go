package handlers

import (
	"strconv"

	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/mappers"
	"github.com/ehanz12/BeforeYouLeave/services"
	"github.com/ehanz12/BeforeYouLeave/utils"
	"github.com/gofiber/fiber/v2"
)

func parseItemID(c *fiber.Ctx) (uint64, *fiber.Error) {
	id, err := strconv.ParseUint(c.Params("itemId"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "itemId harus berupa angka !")
	}
	return id, nil
}

func CreateChecklistItem(c *fiber.Ctx) error {
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.CreateChecklistItemRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	item, err := services.CreateChecklistItem(r, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil membuat checklist item !",
		"data":    mappers.ToChecklistItemResponse(item),
	})
}

func ListChecklistItems(c *fiber.Ctx) error {
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.ListChecklistItemRequest
	if err := c.QueryParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	page, limit := utils.NormalizePagination(r.Page, r.Limit)
	items, total, err := services.ListChecklistItems(r, destinationID, userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	list := make([]response.ChecklistItem, 0, len(items))
	for _, i := range items {
		list = append(list, mappers.ToChecklistItemResponse(i))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan list checklist item !",
		"data":    list,
		"meta":    utils.BuildPagination(page, limit, total),
	})
}

func DetailChecklistItem(c *fiber.Ctx) error {
	itemID, ferr := parseItemID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	item, err := services.DetailChecklistItem(itemID, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan checklist item !",
		"data":    mappers.ToChecklistItemResponse(item),
	})
}

func UpdateChecklistItem(c *fiber.Ctx) error {
	itemID, ferr := parseItemID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.UpdateChecklistItemRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	item, err := services.UpdateChecklistItem(r, itemID, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mengupdate checklist item !",
		"data":    mappers.ToChecklistItemResponse(item),
	})
}

func DeleteChecklistItem(c *fiber.Ctx) error {
	itemID, ferr := parseItemID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	if err := services.DeleteChecklistItem(itemID, destinationID, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil menghapus checklist item !",
	})
}
