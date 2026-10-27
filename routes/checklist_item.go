package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/ehanz12/BeforeYouLeave/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupChecklistItem(api fiber.Router) {
	items := api.Group("/destination/:destinationId/items")
	items.Post("/", middlewares.ProtectedRoute, handlers.CreateChecklistItem)
	items.Get("/", middlewares.ProtectedRoute, handlers.ListChecklistItems)
	items.Get("/:itemId", middlewares.ProtectedRoute, handlers.DetailChecklistItem)
	items.Patch("/:itemId", middlewares.ProtectedRoute, handlers.UpdateChecklistItem)
	items.Delete("/:itemId", middlewares.ProtectedRoute, handlers.DeleteChecklistItem)
}
