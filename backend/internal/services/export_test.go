package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


// TestExportService_ToJSON tests JSON export functionality
func TestExportService_ToJSON(t *testing.T) {
	service := NewExportService()

	// Test data
	data := []map[string]interface{}{
		{"id": float64(1), "name": "John", "age": float64(30)},
		{"id": float64(2), "name": "Jane", "age": float64(25)},
	}
	fieldNames := []string{"id", "name", "age"}

	// Execute
	result, err := service.ToJSON(data, fieldNames)

	// Assert
	require.NoError(t, err, "ToJSON should not return an error")
	assert.NotEmpty(t, result, "Result should not be empty")

	// Check if it's valid JSON
	assert.Contains(t, string(result), `"fields"`, "Should contain fields")
	assert.Contains(t, string(result), `"data"`, "Should contain data")
	assert.Contains(t, string(result), `"count"`, "Should contain count")
	assert.Contains(t, string(result), "John", "Should contain test data")
}

// TestExportService_ToCSV tests CSV export functionality
func TestExportService_ToCSV(t *testing.T) {
	service := NewExportService()

	data := []map[string]interface{}{
		{"id": float64(1), "name": "John", "age": float64(30)},
		{"id": float64(2), "name": "Jane", "age": float64(25)},
	}
	fieldNames := []string{"id", "name", "age"}

	result, err := service.ToCSV(data, fieldNames)

	require.NoError(t, err, "ToCSV should not return an error")
	assert.NotEmpty(t, result, "Result should not be empty")

	// Convert to string for easier assertions
	csv := string(result)

	// Check header row
	assert.Contains(t, csv, "id,name,age", "Should contain header row")

	// Check data rows
	assert.Contains(t, csv, "1,John,30", "Should contain first row")
	assert.Contains(t, csv, "2,Jane,25", "Should contain second row")
}

// TestExportService_ToCSV_EmptyData tests CSV export with empty data
func TestExportService_ToCSV_EmptyData(t *testing.T) {
	service := NewExportService()

	data := []map[string]interface{}{}
	fieldNames := []string{"id", "name"}

	_, err := service.ToCSV(data, fieldNames)

	assert.Error(t, err, "Should return error for empty data")
	assert.Contains(t, err.Error(), "no data", "Error message should mention no data")
}

// TestExportService_ToMarkdownTable tests Markdown export
func TestExportService_ToMarkdownTable(t *testing.T) {
	service := NewExportService()

	data := []map[string]interface{}{
		{"id": float64(1), "name": "John"},
		{"id": float64(2), "name": "Jane"},
	}
	fieldNames := []string{"id", "name"}

	result, err := service.ToMarkdownTable(data, fieldNames)

	require.NoError(t, err, "ToMarkdownTable should not return an error")

	md := string(result)

	// Check header
	assert.Contains(t, md, "| id | name |", "Should contain header")

	// Check separator
	assert.Contains(t, md, "|-----", "Should contain separator")

	// Check data
	assert.Contains(t, md, "| 1 | John |", "Should contain data row")
}

// TestExportService_ToSQL tests SQL export
func TestExportService_ToSQL(t *testing.T) {
	service := NewExportService()

	data := []map[string]interface{}{
		{"id": float64(1), "name": "John", "active": true},
		{"id": float64(2), "name": "Jane", "active": false},
	}
	fieldNames := []string{"id", "name", "active"}

	result, err := service.ToSQL(data, fieldNames, "users")

	require.NoError(t, err, "ToSQL should not return an error")

	sql := string(result)

	// Check CREATE TABLE
	assert.Contains(t, sql, "CREATE TABLE IF NOT EXISTS users", "Should contain CREATE TABLE")
	assert.Contains(t, sql, "id NUMERIC", "Should infer NUMERIC type")
	assert.Contains(t, sql, "name TEXT", "Should infer TEXT type")
	assert.Contains(t, sql, "active BOOLEAN", "Should infer BOOLEAN type")

	// Check INSERT statements
	assert.Contains(t, sql, "INSERT INTO users", "Should contain INSERT")
	assert.Contains(t, sql, "'John'", "Should contain string value")
	assert.Contains(t, sql, "TRUE", "Should contain boolean value")
}

// TestExportService_GetAvailableFormats tests format enumeration
func TestExportService_GetAvailableFormats(t *testing.T) {
	service := NewExportService()

	formats := service.GetAvailableFormats()

	assert.NotEmpty(t, formats, "Should return formats")
	assert.Contains(t, formats, "json", "Should include json")
	assert.Contains(t, formats, "csv", "Should include csv")
	assert.Contains(t, formats, "markdown", "Should include markdown")
	assert.Contains(t, formats, "sql", "Should include sql")
}

// TestFormatValue tests the value formatting helper
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"Nil", nil, ""},
		{"String", "hello", "hello"},
		{"Integer Float", float64(42), "42"},
		{"Decimal Float", float64(3.14), "3.14"},
		{"Boolean", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}


// TestFormatSQLValue tests SQL value formatting
func TestFormatSQLValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"Nil", nil, "NULL"},
		{"String", "hello", "'hello'"},
		{"String with quote", "it's", "'it''s'"},
		{"Integer", float64(42), "42"},
		{"True", true, "TRUE"},
		{"False", false, "FALSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSQLValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
