package handlers

import (
	"strings"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db         *database.Queries
	jwtService *auth.JWTService
}

func NewAuthHandler(db *database.Queries, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create user
	user, err := h.db.CreateUser(c.Context(), database.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
	})
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User already exists"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Get user by email
	user, err := h.db.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate token
	userUUID := uuid.MustParse(user.ID.String())
	token, err := h.jwtService.GenerateToken(userUUID, 24*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		},
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header format"})
	}

	tokenString := tokenParts[1]
	claims, err := h.jwtService.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	// Check if user still exists
	userID := pgtype.UUID{Bytes: claims.UserID, Valid: true}
	_, err = h.db.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	// Generate new token
	newToken, err := h.jwtService.GenerateToken(claims.UserID, 24*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"token":      newToken,
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}
