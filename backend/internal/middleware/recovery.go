package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)


// Recovery middleware recovers from panics and returns a 500 error
func Recovery() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,

		// StackTraceHandler: custom function to handle stack traces
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.Printf("PANIC RECOVERED: %v", e)
		},
	})
}