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

func parseReminderID(c *fiber.Ctx) (uint64, *fiber.Error) {
	id, err := strconv.ParseUint(c.Params("reminderId"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "reminderId harus berupa angka !")
	}
	return id, nil
}

func parseChecklistItemParam(c *fiber.Ctx) (uint64, *fiber.Error) {
	id, err := strconv.ParseUint(c.Params("checklistItemId"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "checklistItemId harus berupa angka !")
	}
	return id, nil
}

// reminderContext = reminderId + userID sekaligus dari request.
func reminderContext(c *fiber.Ctx) (reminderID, userID uint64, ferr *fiber.Error) {
	reminderID, ferr = parseReminderID(c)
	if ferr != nil {
		return 0, 0, ferr
	}
	var ok bool
	userID, ok = getUserID(c)
	if !ok {
		return 0, 0, fiber.NewError(fiber.StatusUnauthorized, "unauthorized !")
	}
	return reminderID, userID, nil
}

func ListReminders(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	var r request.ListReminderRequest
	if err := c.QueryParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	page, limit := utils.NormalizePagination(r.Page, r.Limit)
	reminders, total, err := services.ListReminders(r, userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	list := make([]response.Reminder, 0, len(reminders))
	for _, rm := range reminders {
		list = append(list, mappers.ToReminderResponse(rm))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan list reminder !",
		"data":    list,
		"meta":    utils.BuildPagination(page, limit, total),
	})
}

func DetailReminder(c *fiber.Ctx) error {
	reminderID, userID, ferr := reminderContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	reminder, err := services.DetailReminder(reminderID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan reminder !",
		"data":    mappers.ToReminderDetailResponse(reminder),
	})
}

func ReadReminder(c *fiber.Ctx) error {
	reminderID, userID, ferr := reminderContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	reminder, err := services.MarkReminderRead(reminderID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reminder ditandai sudah dibaca !",
		"data":    mappers.ToReminderDetailResponse(reminder),
	})
}

func SkipReminder(c *fiber.Ctx) error {
	reminderID, userID, ferr := reminderContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	reminder, err := services.SkipReminder(reminderID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "reminder dibatalkan !",
		"data":    mappers.ToReminderDetailResponse(reminder),
	})
}

func UpdateReminderItemCheck(c *fiber.Ctx) error {
	reminderID, userID, ferr := reminderContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	itemID, ferr := parseChecklistItemParam(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.UpdateItemCheckRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	reminder, err := services.UpdateItemCheck(r, reminderID, itemID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil update centangan !",
		"data":    mappers.ToReminderDetailResponse(reminder),
	})
}
