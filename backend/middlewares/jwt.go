package middlewares

import (
	"iou_tracker/util"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid token",
			})
		}

		claims, err := util.NewJWTHelper().ParseWithClaims(token, false)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid access token",
			})
		}

		c.Locals("userID", claims["id"])

		return c.Next()
	}
}
