package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// GlobalRateLimiter creates a rate limiter for all API endpoints
func GlobalRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,               // Max 100 requests
		Expiration: 1 * time.Minute,  // Per minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Rate limit by IP address
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		},
	})
}

// AuthRateLimiter creates a strict rate limiter for authentication endpoints
func AuthRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,                 // Max 5 login attempts
		Expiration: 15 * time.Minute,  // Per 15 minutes
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many authentication attempts. Please try again in 15 minutes.",
			})
		},
	})
}

// SharePasswordRateLimiter creates a strict rate limiter for share password attempts
func SharePasswordRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,                 // Max 3 password attempts
		Expiration: 5 * time.Minute,   // Per 5 minutes
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP + share token
			return c.IP() + ":" + c.Params("token")
		},
		LimitReached: func(c *fiber.Ctx) error {
			c.Set("Content-Type", "text/html")
			return c.Status(fiber.StatusTooManyRequests).SendString(`
				<html><body style="font-family: sans-serif; max-width: 600px; margin: 50px auto; padding: 20px;">
				<div style="background: #fee; border: 1px solid #fcc; padding: 20px; border-radius: 8px;">
					<h2 style="color: #c00; margin: 0 0 10px 0;">🚫 Too Many Attempts</h2>
					<p>Too many password attempts. Please wait 5 minutes before trying again.</p>
				</div>
				</body></html>
			`)
		},
	})
}
