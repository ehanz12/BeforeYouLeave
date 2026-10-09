package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/ehanz12/BeforeYouLeave/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupDestination(api fiber.Router) {
	des := api.Group("/destination")
	des.Post("/", middlewares.ProtectedRoute, handlers.CreateDestination)
}
