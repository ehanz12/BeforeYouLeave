package handlers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/ehanz12/BeforeYouLeave/mappers"
	"github.com/ehanz12/BeforeYouLeave/services"
	"github.com/gofiber/fiber/v2"
)

func CreateDestination(c *fiber.Ctx) error {
	UserID := c.Locals("user_id").(uint64)
	var r request.DestinationRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}
	des, err := services.CreateDestination(r, UserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "berhasil membuat destinasi !", "data": mappers.ToDestinationResponse(des)})

}
