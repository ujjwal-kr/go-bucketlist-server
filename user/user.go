package user

import (
	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	return c.SendString("Gets all users")
}

func GetUsers(c *fiber.Ctx) error {
	return c.SendString("All the users")
}

func Login(c *fiber.Ctx) error {
	return c.SendString("Gets all users")
}

func Register(c *fiber.Ctx) error {
	return c.SendString("Gets all users")
}

func GetTasks(c *fiber.Ctx) error {
	return c.SendString("Gets all users")
}
