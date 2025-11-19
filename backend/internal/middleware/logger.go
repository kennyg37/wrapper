package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)


// Logger middleware logs each HTTP request
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record start time
		start := time.Now()

		// Process request (call next middleware/handler)
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		log.Printf(
			"[%s] %s %s - %d - %v",
			c.Method(),           // HTTP method (GET, POST, etc.)
			c.Path(),             // Request path
			c.IP(),               // Client IP address
			c.Response().StatusCode(), // Response status code
			duration,             // Request duration
		)

		return err
	}
}