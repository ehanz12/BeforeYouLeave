package main

import (

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
  //init database
  configs.LoadEnv()
  //init database
  databases.ConnectDB()

  //init route
  app := fiber.New()
	app.Use(cors.New(cors.Config{AllowOrigins: "http://localhost:5173,https://www.reihan.biz.id",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true, //jika pake jwt
	}))
  
	//setup routes
	routes.SetupRoute(app)
 app.Listen(":" + configs.AppConfig.Port)
}
