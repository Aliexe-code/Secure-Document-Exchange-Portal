package auth

import (
	"strings"

	"Secure-Document-Exchange-Portal/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const UserIDKey = "user_id"

func AuthMiddleware(jwtService *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := GetToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		claims, err := jwtService.ValidateToken(token)
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

func GetToken(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		token := c.Cookies("auth_token")
		if token == "" {
			return ""
		}
		return token
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return ""
	}

	return tokenParts[1]
}

func IsAuthenticated(c *fiber.Ctx, jwtService *JWTService) bool {
	token := GetToken(c)
	if token == "" {
		return false
	}

	_, err := jwtService.ValidateToken(token)
	return err == nil
}

func GetUserName(c *fiber.Ctx, jwtService *JWTService, db *database.Queries) string {
	if !IsAuthenticated(c, jwtService) {
		return ""
	}

	token := GetToken(c)
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		return ""
	}

	userID := pgtype.UUID{Bytes: claims.UserID, Valid: true}
	user, err := db.GetUserByID(c.Context(), userID)
	if err != nil {
		return ""
	}

	return user.FullName
}
