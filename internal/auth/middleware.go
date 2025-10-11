package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const UserIDKey = "user_id"

func AuthMiddleware(jwtService *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := tokenParts[1]
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Set user ID in context
		c.Locals(UserIDKey, claims.UserID)
		return c.Next()
	}
}

func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID := c.Locals(UserIDKey)
	if userID == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, fiber.NewError(fiber.StatusInternalServerError, "Invalid user ID type")
	}

	return id, nil
}
