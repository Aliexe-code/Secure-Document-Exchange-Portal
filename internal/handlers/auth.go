package handlers

import (
	"os"
	"strings"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/models"
	"Secure-Document-Exchange-Portal/internal/services"
	"Secure-Document-Exchange-Portal/internal/validation"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db         *database.Queries
	jwtService *auth.JWTService
	cache      *services.CachedRepository
}

func NewAuthHandler(db *database.Queries, jwtService *auth.JWTService, cache *services.CachedRepository) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtService: jwtService,
		cache:      cache,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest

	// Try parsing as JSON first (API calls), fallback to form data (HTMX)
	if err := c.BodyParser(&req); err != nil {
		// Fallback to form data for HTMX requests
		req = models.RegisterRequest{
			Email:    c.FormValue("email"),
			Password: c.FormValue("password"),
			FullName: c.FormValue("full_name"),
		}
	}

	// Validate email
	if err := validation.ValidateEmail(req.Email); err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusBadRequest).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>` + err.Error() + `</p></div>`)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate password
	if err := validation.ValidatePassword(req.Password); err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusBadRequest).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>` + err.Error() + `</p></div>`)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate full name
	if err := validation.ValidateFullName(req.FullName); err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusBadRequest).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>` + err.Error() + `</p></div>`)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusInternalServerError).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>Failed to hash password</p></div>`)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create user
	user, err := h.db.CreateUser(c.Context(), database.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
	})
	if err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusConflict).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>User already exists</p></div>`)
		}
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "User already exists"})
	}

	if c.Get("HX-Request") == "true" {
		// For HTMX, show success message
		return c.Status(fiber.StatusOK).SendString(`<div class="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
			<p>Registration successful! Please <a href="/login" class="font-bold underline">login here</a> to continue.</p>
		</div>`)
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

	// Try parsing as JSON first (API calls), fallback to form data (HTMX)
	if err := c.BodyParser(&req); err != nil {
		// Fallback to form data for HTMX requests
		req = models.LoginRequest{
			Email:    c.FormValue("email"),
			Password: c.FormValue("password"),
		}
	}

	if req.Email == "" || req.Password == "" {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusBadRequest).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>Email and password are required</p></div>`)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email and password are required"})
	}

	// Get user by email - with caching
	user, err := h.cache.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusUnauthorized).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>Invalid credentials</p></div>`)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusUnauthorized).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>Invalid credentials</p></div>`)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate token
	userUUID := uuid.MustParse(user.ID)
	token, err := h.jwtService.GenerateToken(userUUID, 24*time.Hour)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusInternalServerError).SendString(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded"><p>Failed to generate token</p></div>`)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
	}

	if c.Get("HX-Request") == "true" {
		// Set auth cookie for web requests
		isProduction := os.Getenv("APP_ENV") == "production"
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    token,
			HTTPOnly: true,
			Secure:   isProduction, // Secure only in production with HTTPS
			SameSite: "Lax",
			MaxAge:   86400, // 24 hours
		})
		// For HTMX, redirect to documents page on success
		c.Set("HX-Redirect", "/documents")
		return c.SendString("")
	}

	return c.JSON(models.LoginResponse{
		Token: token,
		User: models.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
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

	// Check if user still exists - with caching
	_, err = h.cache.GetUserByID(c.Context(), claims.UserID)
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

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the auth cookie
	c.ClearCookie("auth_token")

	// Always redirect to home page for web requests
	return c.Redirect("/")
}
