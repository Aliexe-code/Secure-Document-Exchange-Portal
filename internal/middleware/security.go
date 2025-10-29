package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds essential security headers to all responses
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		c.Set("X-Frame-Options", "DENY")

		// Enable XSS protection in older browsers
		c.Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only enable HSTS in production with HTTPS
		if os.Getenv("APP_ENV") == "production" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		// Adjust this based on your actual needs
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' https://unpkg.com; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Set("Content-Security-Policy", csp)

		// Permissions Policy (formerly Feature-Policy)
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		return c.Next()
	}
}
