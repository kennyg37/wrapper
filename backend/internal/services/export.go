package services

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type ExportService struct{}

func NewExportService() *ExportService {
	return &ExportService{}
}

func (s *ExportService) ToJSON(data []map[string]interface{}, fieldNames []string) ([]byte, error) {
	// Create a structured response
	response := map[string]interface{}{
		"fields": fieldNames,
		"data":   data,
		"count":  len(data),
	}

	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return jsonData, nil
}

func (s *ExportService) ToCSV(data []map[string]interface{}, fieldNames []string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header row
	if err := writer.Write(fieldNames); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, row := range data {
		values := make([]string, len(fieldNames))
		for i, field := range fieldNames {
			values[i] = formatValue(row[field])
		}

		if err := writer.Write(values); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	// Flush any buffered data
	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

/*
ToMarkdownTable converts data to a Markdown table format.

CONCEPT: Additional Export Formats

Markdown tables are useful for:
- Documentation
- GitHub issues/PRs
- Human-readable format

Format:
| Field1 | Field2 |
|--------|--------|
| value1 | value2 |
*/
func (s *ExportService) ToMarkdownTable(data []map[string]interface{}, fieldNames []string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	var buf bytes.Buffer

	// Write header
	buf.WriteString("| ")
	buf.WriteString(strings.Join(fieldNames, " | "))
	buf.WriteString(" |\n")

	// Write separator
	buf.WriteString("|")
	for range fieldNames {
		buf.WriteString("--------|")
	}
	buf.WriteString("\n")

	// Write data rows
	for _, row := range data {
		buf.WriteString("| ")
		values := make([]string, len(fieldNames))
		for i, field := range fieldNames {
			values[i] = formatValue(row[field])
		}
		buf.WriteString(strings.Join(values, " | "))
		buf.WriteString(" |\n")
	}

	return buf.Bytes(), nil
}

/*
ToSQL generates INSERT statements for the data.

CONCEPT: SQL Export

This generates SQL INSERT statements that can be:
- Imported into any SQL database
- Used in migrations or seed files
- Shared with teammates

We use a generic table name that can be customized.
*/
func (s *ExportService) ToSQL(data []map[string]interface{}, fieldNames []string, tableName string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to export")
	}

	if tableName == "" {
		tableName = "mock_data"
	}

	var buf bytes.Buffer

	// Write CREATE TABLE statement
	buf.WriteString("-- Generated data\n")
	buf.WriteString(fmt.Sprintf("-- Table: %s\n\n", tableName))
	buf.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", tableName))

	// Infer column types from first row
	firstRow := data[0]
	for i, field := range fieldNames {
		colType := inferSQLType(firstRow[field])
		buf.WriteString(fmt.Sprintf("  %s %s", field, colType))
		if i < len(fieldNames)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString(");\n\n")

	// Write INSERT statements
	for _, row := range data {
		buf.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (",
			tableName,
			strings.Join(fieldNames, ", ")))

		values := make([]string, len(fieldNames))
		for i, field := range fieldNames {
			values[i] = formatSQLValue(row[field])
		}

		buf.WriteString(strings.Join(values, ", "))
		buf.WriteString(");\n")
	}

	return buf.Bytes(), nil
}

// formatValue converts any value to a string for CSV/Markdown
func formatValue(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case float64:
		// Remove unnecessary decimal places
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatSQLValue formats a value for SQL INSERT statement
func formatSQLValue(value interface{}) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// inferSQLType infers SQL column type from a value
func inferSQLType(value interface{}) string {
	if value == nil {
		return "TEXT"
	}

	switch value.(type) {
	case float64:
		return "NUMERIC"
	case bool:
		return "BOOLEAN"
	case string:
		return "TEXT"
	default:
		return "TEXT"
	}
}

/*
GetAvailableFormats returns a list of supported export formats.

CONCEPT: Enumeration Pattern

Instead of hardcoding format names everywhere,
we provide a function that returns valid formats.
This makes it easy to:
- Add new formats
- Validate user input
- Document supported formats
*/
func (s *ExportService) GetAvailableFormats() []string {
	formats := []string{"json", "csv", "markdown", "sql"}
	sort.Strings(formats)
	return formats
}
