package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/handlers"
	"Secure-Document-Exchange-Portal/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
)

func AccessShare(c *fiber.Ctx, db *database.Queries, storage services.StorageService, cache services.CacheService) error {
	token := c.Params("token")

	// Query DB (cache disabled for now due to pgtype marshaling issues)
	share, err := db.GetShareByToken(c.Context(), token)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Share not found"})
	}

	// Check expiration
	if share.ExpiresAt.Time.Before(time.Now()) {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "Share expired"})
	}

	// Check access count
	if share.MaxAccess.Int32 != -1 && share.AccessCount.Int32 >= share.MaxAccess.Int32 {
		return c.Status(fiber.StatusGone).JSON(fiber.Map{"error": "Access limit exceeded"})
	}

	// Check password if set
	if share.PasswordHash.Valid {
		password := c.Query("password")
		if password == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Password required"})
		}
		// TODO: check password hash
		if password != share.PasswordHash.String { // placeholder
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid password"})
		}
	}

	// Update access count
	err = db.UpdateShareAccess(c.Context(), share.ID)
	if err != nil {
		// log error, but continue
	}

	// Download document
	obj, err := storage.Download(c.Context(), "documents", share.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to download file"})
	}
	defer obj.Close()

	// TODO: Decrypt
	data, err := io.ReadAll(obj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	decryptedData := data // placeholder

	// Send file
	c.Set("Content-Type", share.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", share.Filename))
	return c.Send(decryptedData)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	db, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	queries := database.New(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("jwt_secret not set")
	}
	jwtService := auth.NewJWTService(jwtSecret)

	// Initialize storage
	storage, err := services.NewMinIOService("localhost:9000", "minioadmin", "minioadmin", false)
	if err != nil {
		log.Fatal("Failed to connect to storage:", err)
	}

	// Initialize cache
	cache := services.NewRedisCache("localhost:6379", "", 0)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(cors.New())

	api := app.Group("/api")

	// Auth routes (public)
	authGroup := api.Group("/auth")
	authHandler := handlers.NewAuthHandler(queries, jwtService)
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)

	// Protected routes
	protected := api.Group("", auth.AuthMiddleware(jwtService))
	docHandler := handlers.NewDocumentHandler(queries, storage)
	documents := protected.Group("/documents")
	documents.Post("", docHandler.Upload)
	documents.Get("", docHandler.List)
	documents.Get("/:id", docHandler.Download)
	documents.Delete("/:id", docHandler.Delete)
	documents.Post("/:id/share", docHandler.CreateShare)

	_ = authGroup
	_ = protected

	// Public share access
	app.Get("/api/share/:token", func(c *fiber.Ctx) error {
		return AccessShare(c, queries, storage, cache)
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	log.Fatal(app.Listen(":8080"))
}

