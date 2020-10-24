package user

import (
	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	return c.SendString("Gets the lists of the current user")
}

func GetUsers(c *fiber.Ctx) error {
	return c.SendString("All the users")
}

func Login(c *fiber.Ctx) error {
	return c.SendString("Login Route")
}

func Register(c *fiber.Ctx) error {
	return c.SendString("Register route")
}

func GetTasks(c *fiber.Ctx) error {
	return c.SendString("Gets the tasks of the current user")
}
