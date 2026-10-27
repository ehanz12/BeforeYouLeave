package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/routes"
	"github.com/ehanz12/BeforeYouLeave/workers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	configs.LoadEnv()

	if err := databases.ConnectDB(); err != nil {
		log.Fatalf("⚠️ gagal connect database: %v", err)
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,https://www.reihan.biz.id",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true, //jika pake jwt
	}))

	routes.SetupRoute(app)

	// Context ini dipakai worker reminder, dan di-cancel saat server dimatikan
	// (Ctrl+C atau SIGTERM dari docker/orchestrator).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workers.StartReminderWorker(ctx)

	go func() {
		<-ctx.Done()
		log.Println("🔴 server shutting down ...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("⚠️ error saat shutdown: %v", err)
		}
	}()

	port := ":" + configs.AppConfig.Port
	log.Printf("🚀 server listen on %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("⚠️ server error: %v", err)
	}
}
