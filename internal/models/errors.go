package models

import "errors"

var (
	ErrInvalidScenario    = errors.New("scenario description is required")
	ErrInvalidRowCount    = errors.New("row count must be between 1 and 1000")
	ErrRequestNotFound    = errors.New("generation request not found")
	ErrDatasetNotFound    = errors.New("dataset not found")
	ErrOpenAIFailure      = errors.New("failed to generate data with OpenAI")
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrInvalidFormat      = errors.New("invalid export format")
)
