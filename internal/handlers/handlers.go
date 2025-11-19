package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kennyg37/wrapperX/backend/internal/database"
	"github.com/kennyg37/wrapperX/backend/internal/models"
	"github.com/kennyg37/wrapperX/backend/internal/services"
)

type Handler struct {
	db            *database.DB
	openaiService *services.OpenAIService
	exportService *services.ExportService
}

// NewHandler creates a new handler instance
func NewHandler(db *database.DB, openaiService *services.OpenAIService, exportService *services.ExportService) *Handler {
	return &Handler{
		db:            db,
		openaiService: openaiService,
		exportService: exportService,
	}
}


func (h *Handler) GenerateMockData(c *fiber.Ctx) error {
	// Parse request body
	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	log.Printf("New generation request: %s (%d rows)", req.Scenario, req.RowCount)

	// Create generation request in database
	var requestID int64
	err := h.db.QueryRow(
		`INSERT INTO generation_requests (scenario, row_count, status)
		 VALUES ($1, $2, 'pending')
		 RETURNING id`,
		req.Scenario,
		req.RowCount,
	).Scan(&requestID)

	if err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	// Update status to processing
	_, err = h.db.Exec(
		`UPDATE generation_requests SET status = 'processing' WHERE id = $1`,
		requestID,
	)
	if err != nil {
		log.Printf("Failed to update status: %v", err)
	}

	// Generate mock data using OpenAI
	data, fieldNames, err := h.openaiService.GenerateMockData(c.Context(), req.Scenario, req.RowCount)
	if err != nil {
		// Update status to failed
		_, _ = h.db.Exec(
			`UPDATE generation_requests SET status = 'failed' WHERE id = $1`,
			requestID,
		)

		log.Printf("OpenAI error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to generate data",
			Message: err.Error(),
		})
	}

	// Convert data to JSONB for PostgreSQL
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to serialize data",
			Message: err.Error(),
		})
	}

	// Save generated data to database
	_, err = h.db.Exec(
		`INSERT INTO mock_datasets (request_id, data, field_names)
		 VALUES ($1, $2, $3)`,
		requestID,
		dataJSON,
		fieldNames,
	)

	if err != nil {
		log.Printf("Failed to save dataset: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to save dataset",
			Message: err.Error(),
		})
	}

	// Update request status to completed
	_, err = h.db.Exec(
		`UPDATE generation_requests SET status = 'completed', generated_at = $1 WHERE id = $2`,
		time.Now(),
		requestID,
	)

	if err != nil {
		log.Printf("Failed to update status: %v", err)
	}

	log.Printf("Generation request %d completed successfully", requestID)

	// Return response
	return c.Status(fiber.StatusCreated).JSON(models.GenerateResponse{
		ID:        requestID,
		Status:    "completed",
		Message:   fmt.Sprintf("Successfully generated %d rows of mock data", req.RowCount),
		CreatedAt: time.Now(),
	})
}


func (h *Handler) GetGenerationRequest(c *fiber.Ctx) error {
	// Get ID from URL parameter
	id := c.Params("id")

	var request models.GenerationRequest
	err := h.db.QueryRow(
		`SELECT id, scenario, row_count, status, generated_at, created_at, updated_at
		 FROM generation_requests
		 WHERE id = $1`,
		id,
	).Scan(
		&request.ID,
		&request.Scenario,
		&request.RowCount,
		&request.Status,
		&request.GeneratedAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Request not found",
			Message: fmt.Sprintf("No generation request found with ID %s", id),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	return c.JSON(request)
}

func (h *Handler) GetMockData(c *fiber.Ctx) error {
	requestID := c.Params("id")

	// Get request details
	var scenario string
	var status string
	err := h.db.QueryRow(
		`SELECT scenario, status FROM generation_requests WHERE id = $1`,
		requestID,
	).Scan(&scenario, &status)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Request not found",
			Message: fmt.Sprintf("No generation request found with ID %s", requestID),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	if status != "completed" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Data not available",
			Message: fmt.Sprintf("Request status is '%s', data is only available for completed requests", status),
		})
	}

	// Get dataset
	var dataJSON []byte
	var fieldNames []string
	var createdAt time.Time
	var datasetID int64

	err = h.db.QueryRow(
		`SELECT id, data, field_names, created_at
		 FROM mock_datasets
		 WHERE request_id = $1`,
		requestID,
	).Scan(&datasetID, &dataJSON, &fieldNames, &createdAt)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Dataset not found",
			Message: "Generated data not found for this request",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	// Parse JSON data
	var data []map[string]interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to parse data",
			Message: err.Error(),
		})
	}

	// Build response
	response := models.DataResponse{
		ID:         datasetID,
		RequestID:  int64(mustAtoi(requestID)),
		Scenario:   scenario,
		Data:       data,
		FieldNames: fieldNames,
		RowCount:   len(data),
		CreatedAt:  createdAt,
	}

	return c.JSON(response)
}

/*
ExportMockData handles GET /api/data/:id/export?format=csv

This endpoint exports the mock data in different formats.
Supported formats: json, csv, markdown, sql

Query parameters:
- format: export format (default: json)
- table: table name for SQL export (default: mock_data)
*/
func (h *Handler) ExportMockData(c *fiber.Ctx) error {
	requestID := c.Params("id")
	format := c.Query("format", "json") // Default to JSON
	tableName := c.Query("table", "mock_data")

	// Get dataset
	var dataJSON []byte
	var fieldNames []string

	err := h.db.QueryRow(
		`SELECT data, field_names
		 FROM mock_datasets
		 WHERE request_id = $1`,
		requestID,
	).Scan(&dataJSON, &fieldNames)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Dataset not found",
			Message: fmt.Sprintf("No dataset found for request ID %s", requestID),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	// Parse JSON data
	var data []map[string]interface{}
	if err := json.Unmarshal(dataJSON, &data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to parse data",
			Message: err.Error(),
		})
	}

	// Export data in requested format
	var exportData []byte
	var contentType string
	var filename string

	switch format {
	case "json":
		exportData, err = h.exportService.ToJSON(data, fieldNames)
		contentType = "application/json"
		filename = fmt.Sprintf("mockdata-%s.json", requestID)

	case "csv":
		exportData, err = h.exportService.ToCSV(data, fieldNames)
		contentType = "text/csv"
		filename = fmt.Sprintf("mockdata-%s.csv", requestID)

	case "markdown", "md":
		exportData, err = h.exportService.ToMarkdownTable(data, fieldNames)
		contentType = "text/markdown"
		filename = fmt.Sprintf("mockdata-%s.md", requestID)

	case "sql":
		exportData, err = h.exportService.ToSQL(data, fieldNames, tableName)
		contentType = "application/sql"
		filename = fmt.Sprintf("mockdata-%s.sql", requestID)

	default:
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid format",
			Message: fmt.Sprintf("Format '%s' is not supported. Use: json, csv, markdown, or sql", format),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Export failed",
			Message: err.Error(),
		})
	}

	// Set headers for file download
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(exportData)
}


func (h *Handler) ListGenerationRequests(c *fiber.Ctx) error {
	rows, err := h.db.Query(
		`SELECT id, scenario, row_count, status, generated_at, created_at, updated_at
		 FROM generation_requests
		 ORDER BY created_at DESC
		 LIMIT 100`,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}
	defer rows.Close()

	var requests []models.GenerationRequest

	for rows.Next() {
		var req models.GenerationRequest
		err := rows.Scan(
			&req.ID,
			&req.Scenario,
			&req.RowCount,
			&req.Status,
			&req.GeneratedAt,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
				Error:   "Failed to scan row",
				Message: err.Error(),
			})
		}
		requests = append(requests, req)
	}

	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"requests": requests,
		"count":    len(requests),
	})
}

// HealthCheck handles GET /api/health
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "mock-data-generator",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// Helper function to convert string to int
func mustAtoi(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
