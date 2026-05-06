package handlers

import (
	"summarizer-api/globals"

	"github.com/gofiber/fiber/v3"
)

func ClearDBHandler(c fiber.Ctx) error {
	if err := globals.DB.Clear(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Database cleared successfully",
	})
}
