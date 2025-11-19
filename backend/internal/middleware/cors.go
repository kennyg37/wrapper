package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS middleware configures Cross-Origin Resource Sharing
func CORS(allowedOrigins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: getAllowOriginsString(allowedOrigins),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders: "Content-Length,Content-Type",
		MaxAge: 86400, // 24 hours
	})
}

// getAllowOriginsString converts a slice of origins to a comma-separated string
func getAllowOriginsString(origins []string) string {
	result := ""
	for i, origin := range origins {
		result += origin
		if i < len(origins)-1 {
			result += ","
		}
	}
	return result
}
