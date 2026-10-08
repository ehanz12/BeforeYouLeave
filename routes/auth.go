package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/ehanz12/BeforeYouLeave/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupRouteAuth(api fiber.Router) {
	auth := api.Group("/auth")

	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Get("/me", middlewares.ProtectedRoute, handlers.Me)
}
