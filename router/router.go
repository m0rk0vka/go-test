package router

import (
	"github.com/gofiber/fiber"
	"github.com/m0rk0vka/go-test/controllers"
)

func Router() *fiber.App{
	app := fiber.New()

	app.Get("/list", controllers.GetNews)
	app.Post("/edit/:Id", controllers.UpdateNews)

	return app
}