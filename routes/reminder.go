package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/ehanz12/BeforeYouLeave/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupReminder(api fiber.Router) {
	reminder := api.Group("/reminder")
	reminder.Get("/", middlewares.ProtectedRoute, handlers.ListReminders)
	reminder.Get("/:reminderId", middlewares.ProtectedRoute, handlers.DetailReminder)
	reminder.Patch("/:reminderId/read", middlewares.ProtectedRoute, handlers.ReadReminder)
	reminder.Patch("/:reminderId/skip", middlewares.ProtectedRoute, handlers.SkipReminder)
	reminder.Patch("/:reminderId/items/:checklistItemId", middlewares.ProtectedRoute, handlers.UpdateReminderItemCheck)
}
