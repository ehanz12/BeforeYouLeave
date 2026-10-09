package handlers

import (
	"github.com/ehanz12/BeforeYouLeave/dtos/request"
	"github.com/gofiber/fiber/v2"
)

func CreateDestination(c *fiber.Ctx) error {
	var r request.DestinationRequest
	if err := c.BodyParser(&r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "terjadi kesalahan request !"})
	}

}
