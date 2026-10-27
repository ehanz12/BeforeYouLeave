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

func parseScheduleID(c *fiber.Ctx) (uint64, *fiber.Error) {
	id, err := strconv.ParseUint(c.Params("scheduleId"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "scheduleId harus berupa angka !")
	}
	return id, nil
}

func CreateSchedule(c *fiber.Ctx) error {
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.CreateScheduleRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	schedule, err := services.CreateSchedule(r, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil membuat jadwal !",
		"data":    mappers.ToScheduleResponse(schedule),
	})
}

func ListSchedules(c *fiber.Ctx) error {
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.ListScheduleRequest
	if err := c.QueryParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	page, limit := utils.NormalizePagination(r.Page, r.Limit)
	schedules, total, err := services.ListSchedules(r, destinationID, userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	list := make([]response.Schedule, 0, len(schedules))
	for _, s := range schedules {
		list = append(list, mappers.ToScheduleResponse(s))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan list jadwal !",
		"data":    list,
		"meta":    utils.BuildPagination(page, limit, total),
	})
}

func DetailSchedule(c *fiber.Ctx) error {
	scheduleID, ferr := parseScheduleID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	schedule, err := services.DetailSchedule(scheduleID, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan jadwal !",
		"data":    mappers.ToScheduleResponse(schedule),
	})
}

func UpdateSchedule(c *fiber.Ctx) error {
	scheduleID, ferr := parseScheduleID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.UpdateScheduleRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	schedule, err := services.UpdateSchedule(r, scheduleID, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mengupdate jadwal !",
		"data":    mappers.ToScheduleResponse(schedule),
	})
}

func DeleteSchedule(c *fiber.Ctx) error {
	scheduleID, ferr := parseScheduleID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	if err := services.DeleteSchedule(scheduleID, destinationID, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil menghapus jadwal !",
	})
}

func ReplaceSchedules(c *fiber.Ctx) error {
	destinationID, userID, ferr := destinationContext(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	var r request.ReplaceSchedulesRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	schedules, err := services.ReplaceSchedules(r, destinationID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	list := make([]response.Schedule, 0, len(schedules))
	for _, s := range schedules {
		list = append(list, mappers.ToScheduleResponse(s))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil menyimpan jadwal mingguan !",
		"data":    list,
	})
}
