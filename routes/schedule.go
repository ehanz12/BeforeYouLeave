package routes

import (
	"github.com/ehanz12/BeforeYouLeave/handlers"
	"github.com/ehanz12/BeforeYouLeave/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupSchedule(api fiber.Router) {
	schedules := api.Group("/destination/:destinationId/schedules")
	schedules.Post("/", middlewares.ProtectedRoute, handlers.CreateSchedule)
	schedules.Get("/", middlewares.ProtectedRoute, handlers.ListSchedules)
	schedules.Put("/", middlewares.ProtectedRoute, handlers.ReplaceSchedules)
	schedules.Get("/:scheduleId", middlewares.ProtectedRoute, handlers.DetailSchedule)
	schedules.Patch("/:scheduleId", middlewares.ProtectedRoute, handlers.UpdateSchedule)
	schedules.Delete("/:scheduleId", middlewares.ProtectedRoute, handlers.DeleteSchedule)
}
