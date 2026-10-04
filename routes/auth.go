package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupRouteAuth(api fiber.Router) {
  auth:= api.Group("/auth")

  auth.Post("/register", handlers.Register)
}