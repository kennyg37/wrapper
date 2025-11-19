package models

import "time"


type GenerationRequest struct {
	ID          int64     `json:"id" db:"id"`
	Scenario    string    `json:"scenario" db:"scenario"`
	RowCount    int       `json:"row_count" db:"row_count"`
	Status      string    `json:"status" db:"status"` // pending, processing, completed, failed
	GeneratedAt time.Time `json:"generated_at" db:"generated_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type MockDataset struct {
	ID          int64                    `json:"id" db:"id"`
	RequestID   int64                    `json:"request_id" db:"request_id"`
	Data        []map[string]interface{} `json:"data"` 
	FieldNames  []string                 `json:"field_names" db:"field_names"`
	CreatedAt   time.Time                `json:"created_at" db:"created_at"`
}


type GenerateRequest struct {
	Scenario string `json:"scenario"` 
	RowCount int    `json:"row_count"` 
}

// Validate checks if the request is valid
func (r *GenerateRequest) Validate() error {
	if r.Scenario == "" {
		return ErrInvalidScenario
	}
	if r.RowCount < 1 || r.RowCount > 1000 {
		return ErrInvalidRowCount
	}
	return nil
}

type GenerateResponse struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type DataResponse struct {
	ID         int64                    `json:"id"`
	RequestID  int64                    `json:"request_id"`
	Scenario   string                   `json:"scenario"`
	Data       []map[string]interface{} `json:"data"`
	FieldNames []string                 `json:"field_names"`
	RowCount   int                      `json:"row_count"`
	CreatedAt  time.Time                `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
