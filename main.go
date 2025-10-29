package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/handlers"
	"Secure-Document-Exchange-Portal/internal/middleware"
	"Secure-Document-Exchange-Portal/internal/services"
	"Secure-Document-Exchange-Portal/templates"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
)

func AccessShare(c *fiber.Ctx, db *database.Queries, storage services.StorageService, cachedRepo *services.CachedRepository) error {
	token := c.Params("token")

	// Use cached repository for share lookup
	share, err := cachedRepo.GetShareByToken(c.Context(), token)
	if err != nil {
		c.Set("Content-Type", "text/html")
		return c.Status(fiber.StatusNotFound).SendString(`
			<html><body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
			<div style="background: #fee; border: 1px solid #fcc; padding: 20px; border-radius: 8px;">
				<h2 style="color: #c00; margin: 0 0 10px 0;">❌ Share Not Found</h2>
				<p>This share link does not exist or has been deleted.</p>
			</div>
			</body></html>
		`)
	}

	// Check expiration
	if share.ExpiresAt.Before(time.Now()) {
		c.Set("Content-Type", "text/html")
		return c.Status(fiber.StatusGone).SendString(`
			<html><body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
			<div style="background: #ffe; border: 1px solid #ffa; padding: 20px; border-radius: 8px;">
				<h2 style="color: #a80; margin: 0 0 10px 0;">⏱️ Share Expired</h2>
				<p>This share link has expired and is no longer available.</p>
			</div>
			</body></html>
		`)
	}

	// Check access count
	if share.MaxAccess != -1 && share.AccessCount >= share.MaxAccess {
		c.Set("Content-Type", "text/html")
		return c.Status(fiber.StatusGone).SendString(`
			<html><body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
			<div style="background: #ffe; border: 1px solid #ffa; padding: 20px; border-radius: 8px;">
				<h2 style="color: #a80; margin: 0 0 10px 0;">🔒 Access Limit Reached</h2>
				<p>This share link has reached its maximum number of accesses.</p>
			</div>
			</body></html>
		`)
	}

	// Check password if set
	if share.PasswordHash != nil {
		password := c.FormValue("password")
		if password == "" {
			password = c.Query("password")
		}

		if password == "" {
			// Show password form
			errorMsg := ""
			if c.Method() == "POST" {
				errorMsg = `<p style="color: #c00; margin-bottom: 15px;">❌ Password is required</p>`
			}

			c.Set("Content-Type", "text/html")
			return c.SendString(fmt.Sprintf(`
				<html>
				<head>
					<title>Password Protected Share</title>
					<style>
						body { font-family: sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; background: #f5f5f5; }
						.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
						h2 { color: #333; margin: 0 0 10px 0; }
						.subtitle { color: #666; margin-bottom: 20px; font-size: 14px; }
						.file-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
						.file-info strong { color: #555; }
						input[type="password"] { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
						button { width: 100%%; padding: 12px; background: #4CAF50; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 15px; }
						button:hover { background: #45a049; }
					</style>
				</head>
				<body>
					<div class="container">
						<h2>🔒 Password Protected</h2>
						<p class="subtitle">This file is password protected</p>
						<div class="file-info">
							<strong>File:</strong> %s<br>
							<strong>Size:</strong> %.2f MB
						</div>
						%s
						<form method="POST">
							<input type="password" name="password" placeholder="Enter password" required autofocus>
							<button type="submit">Access File</button>
						</form>
					</div>
				</body>
				</html>
			`, share.Filename, float64(share.FileSize)/1024/1024, errorMsg))
		}

		// Check password hash using bcrypt
		if err := bcrypt.CompareHashAndPassword([]byte(*share.PasswordHash), []byte(password)); err != nil {
			errorMsg := `<p style="color: #c00; margin-bottom: 15px;">❌ Invalid password. Please try again.</p>`
			c.Set("Content-Type", "text/html")
			return c.SendString(fmt.Sprintf(`
				<html>
				<head>
					<title>Password Protected Share</title>
					<style>
						body { font-family: sans-serif; max-width: 500px; margin: 50px auto; padding: 20px; background: #f5f5f5; }
						.container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
						h2 { color: #333; margin: 0 0 10px 0; }
						.subtitle { color: #666; margin-bottom: 20px; font-size: 14px; }
						.file-info { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
						.file-info strong { color: #555; }
						input[type="password"] { width: 100%%; padding: 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 14px; box-sizing: border-box; }
						button { width: 100%%; padding: 12px; background: #4CAF50; color: white; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; margin-top: 15px; }
						button:hover { background: #45a049; }
					</style>
				</head>
				<body>
					<div class="container">
						<h2>🔒 Password Protected</h2>
						<p class="subtitle">This file is password protected</p>
						<div class="file-info">
							<strong>File:</strong> %s<br>
							<strong>Size:</strong> %.2f MB
						</div>
						%s
						<form method="POST">
							<input type="password" name="password" placeholder="Enter password" required autofocus>
							<button type="submit">Access File</button>
						</form>
					</div>
				</body>
				</html>
			`, share.Filename, float64(share.FileSize)/1024/1024, errorMsg))
		}
	}

	// Update access count (and invalidate cache after update)
	shareID, err := uuid.Parse(share.ID)
	if err != nil {
		// log error but continue
	} else {
		err = db.UpdateShareAccess(c.Context(), pgtype.UUID{Bytes: shareID, Valid: true})
		if err != nil {
			// log error, but continue
		}
		// Invalidate the share cache since access count changed
		cachedRepo.InvalidateShare(c.Context(), token)
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
	var storage services.StorageService

	// Get storage configuration from environment
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = "localhost:9000" // Default for local development
	}
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	if s3AccessKey == "" {
		s3AccessKey = "minioadmin" // Default for local development
	}
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	if s3SecretKey == "" {
		s3SecretKey = "minioadmin" // Default for local development
	}
	s3UseSSL := os.Getenv("S3_USE_SSL") == "true"

	minioStorage, err := services.NewMinIOService(s3Endpoint, s3AccessKey, s3SecretKey, s3UseSSL)
	minioAvailable := false
	
	if err == nil {
		// Test the connection by trying to list buckets
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err = minioStorage.ListBuckets(ctx)
		if err == nil {
			// Ensure the documents bucket exists
			err = minioStorage.EnsureBucket(context.Background(), "documents")
			if err == nil {
				minioAvailable = true
				storage = minioStorage
				log.Println("✓ Using MinIO storage")
			}
		}
	}
	
	if !minioAvailable {
		log.Println("MinIO not available, falling back to local storage")
		localStorage, err := services.NewLocalStorageService("./storage")
		if err != nil {
			log.Fatal("Failed to initialize storage:", err)
		}
		storage = localStorage
		log.Println("✓ Using local file storage at ./storage")
	}

	// Initialize cache with configuration from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local development
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	// Redis DB defaults to 0 if not specified
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		fmt.Sscanf(dbStr, "%d", &redisDB)
	}

	cache := services.NewRedisCache(redisAddr, redisPassword, redisDB)

	// Create cached repository
	cachedRepo := services.NewCachedRepository(queries, cache)

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

	// Security headers middleware
	app.Use(middleware.SecurityHeaders())

	// CORS configuration - restrict to specific origins in production
	allowOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowOrigins == "" {
		// Default to localhost for development
		allowOrigins = "http://localhost:8080,http://127.0.0.1:8080"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Global rate limiting for API endpoints
	app.Use("/api", middleware.GlobalRateLimiter())

	// Static files
	app.Use("/static", filesystem.New(filesystem.Config{
		Root: http.Dir("./static"),
	}))

	// Auth handler instance (declared later, but needed here)
	var authHandler *handlers.AuthHandler

	// Web routes (HTML responses)
	app.Get("/", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.DocumentListPage([]templates.Document{})).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/login", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.LoginPage([]string{}, "")).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.RegisterPage([]string{})).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/documents", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.DocumentListPage([]templates.Document{})).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/documents/upload", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return templates.UploadForm().Render(c.Context(), c.Response().BodyWriter())
	})

	api := app.Group("/api")

	// Auth routes (public) with strict rate limiting
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.AuthRateLimiter()) // Apply auth rate limiter
	authHandler = handlers.NewAuthHandler(queries, jwtService, cachedRepo)
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	app.Get("/logout", func(c *fiber.Ctx) error {
		return authHandler.Logout(c)
	})



	// Protected routes
	protected := api.Group("", auth.AuthMiddleware(jwtService))
	docHandler := handlers.NewDocumentHandler(queries, storage, cachedRepo)
	documents := protected.Group("/documents")
	documents.Post("", docHandler.Upload)
	documents.Get("", docHandler.List)
	documents.Get("/:id/view", docHandler.View)
	documents.Get("/:id/download", docHandler.Download)
	documents.Post("/:id/share", docHandler.CreateShare)
	documents.Delete("/:id", docHandler.Delete)
	documents.Get("/:id", docHandler.Download)

	// Web document routes
	app.Get("/documents", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.DocumentListPage([]templates.Document{})).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/documents/upload", func(c *fiber.Ctx) error {
		isAuth := auth.IsAuthenticated(c, jwtService)
		userName := auth.GetUserName(c, jwtService, queries)
		c.Set("Content-Type", "text/html")
		return templates.Base(isAuth, userName, templates.UploadForm()).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/documents/:id/share", docHandler.GetShareForm)

	// Close modal endpoint
	app.Get("/api/close-modal", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(`<div id="share-modal"></div>`)
	})

	_ = authGroup
	_ = protected

	// Public share access (GET and POST for password submission) with rate limiting
	shareGroup := app.Group("/api/share")
	shareGroup.Use(middleware.SharePasswordRateLimiter()) // Apply share password rate limiter
	shareGroup.Get("/:token", func(c *fiber.Ctx) error {
		return AccessShare(c, queries, storage, cachedRepo)
	})
	shareGroup.Post("/:token", func(c *fiber.Ctx) error {
		return AccessShare(c, queries, storage, cachedRepo)
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})


	log.Fatal(app.Listen(":8080"))
}

