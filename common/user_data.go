package common

import (
	"github.com/gofiber/fiber/v2"
)

func GetUserData(c *fiber.Ctx) map[string]interface{} {
	return c.Locals("user_data").(map[string]interface{})
}
