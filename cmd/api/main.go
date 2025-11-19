package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/kennyg37/wrapperX/backend/internal/config"
	"github.com/kennyg37/wrapperX/backend/internal/database"
	"github.com/kennyg37/wrapperX/backend/internal/handlers"
	"github.com/kennyg37/wrapperX/backend/internal/middleware"
	"github.com/kennyg37/wrapperX/backend/internal/services"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Data Generator API")
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Port: %s", cfg.Port)

	// Connect to PostgreSQL database
	db, err := database.New(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services and handlers
	openaiService := services.NewOpenAIService(cfg.OpenAIAPIKey)
	exportService := services.NewExportService()

	handler := handlers.NewHandler(db, openaiService, exportService)

	app := fiber.New(fiber.Config{
		AppName: "Mock Data Generator API v1.0",
		ErrorHandler: customErrorHandler,

		// BodyLimit: maximum request body size (10MB)
		BodyLimit: 10 * 1024 * 1024,
	})


	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS(cfg.CORSOrigins))

	// API routes
	api := app.Group("/api")

	api.Get("/health", handler.HealthCheck)

	api.Post("/generate", handler.GenerateMockData)
	api.Get("/requests", handler.ListGenerationRequests)
	api.Get("/requests/:id", handler.GetGenerationRequest)

	api.Get("/data/:id", handler.GetMockData)
	api.Get("/data/:id/export", handler.ExportMockData)


	// Channel to listen for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := ":" + cfg.Port
		log.Printf("Server listening on http://localhost%s", addr)

		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	// Gracefully shutdown the server
	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}


func customErrorHandler(c *fiber.Ctx, err error) error {
	// Default to 500 Internal Server Error
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error and has a status code
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Printf("Error: %v", err)

	return c.Status(code).JSON(fiber.Map{
		"error":   message,
		"message": err.Error(),
	})
}

