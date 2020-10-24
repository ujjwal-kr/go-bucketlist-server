package list

import (
	"github.com/gofiber/fiber/v2"
)

func GetList(c *fiber.Ctx) error {
	return c.SendString("Gets one list item")
}

func PostList(c *fiber.Ctx) error {
	return c.SendString("Posts a list item")
}

func DeleteList(c *fiber.Ctx) error {
	return c.SendString("Deletes a list")
}

func UpdateList(c *fiber.Ctx) error {
	return c.SendString("Updates a list")
}
