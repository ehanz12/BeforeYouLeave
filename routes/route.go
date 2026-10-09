package routes

import "github.com/gofiber/fiber/v2"

func SetupRoute(app *fiber.App) {
	api := app.Group("/api")
	SetupRouteAuth(api)
	SetupDestination(api)
}
