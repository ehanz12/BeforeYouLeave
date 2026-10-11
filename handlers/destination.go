package handlers

import (
	"strconv"
	"strings"

	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/dtos/response"
	"github.com/ehanz12/BeforeYouLeave/mappers"
	"github.com/ehanz12/BeforeYouLeave/services"
	"github.com/ehanz12/BeforeYouLeave/utils"
	"github.com/gofiber/fiber/v2"
)

func parseDestinationID(c *fiber.Ctx) (uint64, *fiber.Error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "id harus berupa angka !")
	}
	return id, nil
}

func getUserID(c *fiber.Ctx) (uint64, bool) {
	userID, ok := c.Locals("user_id").(uint64)
	return userID, ok
}

func CreateDestination(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	var r request.DestinationRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	des, err := services.CreateDestination(r, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "berhasil membuat destinasi !", "data": mappers.ToDestinationResponse(des)})
}

func ListDestinations(c *fiber.Ctx) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	var r request.ListDestinationRequest
	if err := c.QueryParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	page, limit := utils.NormalizePagination(r.Page, r.Limit)
	destinations, total, err := services.ListDestinations(r, page, limit, userID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if strings.HasPrefix(err.Error(), "format") {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	list := make([]response.Destination, 0, len(destinations))
	for _, d := range destinations {
		list = append(list, mappers.ToDestinationResponse(d))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "berhasil mendapatkan list destinasi !",
		"data":    list,
		"meta":    utils.BuildPagination(page, limit, total),
	})
}

func DetailDestination(c *fiber.Ctx) error {
	desID, ferr := parseDestinationID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	des, err := services.DetailDestination(desID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "berhasil mendapatkan destinasi !", "data": mappers.ToDestinationDetailResponse(des)})
}

func UpdateDestination(c *fiber.Ctx) error {
	desID, ferr := parseDestinationID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	var r request.UpdateDestinationRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	des, err := services.UpdateDestination(r, desID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "berhasil mengupdate destinasi !", "data": mappers.ToDestinationDetailResponse(des)})
}

func DeleteDestination(c *fiber.Ctx) error {
	desID, ferr := parseDestinationID(c)
	if ferr != nil {
		return c.Status(ferr.Code).JSON(fiber.Map{"success": false, "error": ferr.Message})
	}
	userID, ok := getUserID(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized !"})
	}
	if err := services.DeleteDestination(desID, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "berhasil menghapus destinasi !"})
}
