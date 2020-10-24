package task

import (
	"github.com/gofiber/fiber/v2"
)

func PostTask(c *fiber.Ctx) error {
	return c.SendString("posts a task")
}

func DeleteTask(c *fiber.Ctx) error {
	return c.SendString("deletes a taask")
}
