package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ujjwal-kr/go-bucketlist-server/user"
)

func welcome(c *fiber.Ctx) error {
	return c.SendString("Helloworld")
}

func setupRoutes(app *fiber.App) {

	app.Get("/", welcome)

	app.Get("/users", user.GetUsers)
	app.Get("/users/:id", user.GetUser)
	app.Post("/auth/login", user.Login)
	app.Post("/auth/register", user.Register)
	app.Get("/users/:id/tasks", user.GetTasks)
}

func main() {
	app := fiber.New()

	setupRoutes(app)

	app.Listen(":8080")
}
