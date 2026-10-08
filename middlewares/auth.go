package middlewares

import (
	"strings"

	"github.com/ehanz12/BeforeYouLeave/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func ProtectedRoute(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if header == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"error": "invalid authorization format"})
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")

	token, err := utils.VerifyToken(tokenStr)
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid jwt claims"})
	}
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "invalid user_id claim"})
	}

	userID := uint64(userIDFloat)

	c.Locals("user_id", userID)

	return c.Next()
}
