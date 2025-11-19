package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

/*
CONCEPT: Service Layer Architecture

The service layer sits between handlers (HTTP layer) and the database.
Benefits:
1. Business logic separation from HTTP concerns
2. Easier to test (can test without HTTP server)
3. Reusable across different interfaces (HTTP, CLI, gRPC, etc.)
4. Clear separation of concerns

Services should:
- Not know about HTTP (no request/response objects)
- Focus on business logic
- Return domain errors, not HTTP status codes
*/

type OpenAIService struct {
	client *openai.Client
}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService(apiKey string) *OpenAIService {
	return &OpenAIService{
		client: openai.NewClient(apiKey),
	}
}

/*
GenerateMockData uses OpenAI's GPT model to generate mock data based on a scenario.

CONCEPT: AI-Generated Mock Data

Instead of using hardcoded templates or random generators,
we're using ChatGPT to intelligently generate realistic mock data.

The process:
1. We send a carefully crafted prompt to GPT
2. GPT generates data that matches the scenario
3. We parse the JSON response
4. Return structured data

This is powerful because GPT can generate:
- Realistic names, emails, addresses
- Contextually appropriate data
- Complex nested structures
- Domain-specific data (medical, financial, etc.)
*/
func (s *OpenAIService) GenerateMockData(ctx context.Context, scenario string, rowCount int) ([]map[string]interface{}, []string, error) {
	// Construct a prompt that instructs GPT to generate JSON data
	prompt := fmt.Sprintf(`Generate %d rows of realistic mock data based on the following scenario: "%s"

Requirements:
1. Return ONLY a valid JSON object with this structure: {"fields": ["field1", "field2", ...], "data": [{...}, {...}, ...]}
2. The "fields" array should list all field names
3. The "data" array should contain %d objects, each with the same fields
4. Make the data realistic and varied
5. Use appropriate data types (strings, numbers, booleans)
6. Do not include any explanation, only the JSON object
7. Ensure all field names are consistent across all rows

Example for "users with contact info":
{
  "fields": ["id", "name", "email", "age", "city"],
  "data": [
    {"id": 1, "name": "John Doe", "email": "john@example.com", "age": 28, "city": "New York"},
    {"id": 2, "name": "Jane Smith", "email": "jane@example.com", "age": 34, "city": "Los Angeles"}
  ]
}`, rowCount, scenario, rowCount)

	log.Printf("🤖 Requesting mock data from OpenAI for scenario: %s (%d rows)", scenario, rowCount)

	/*
	CONCEPT: ChatGPT API

	The ChatCompletion API is conversational:
	- You send a list of messages (system, user, assistant)
	- System messages set the AI's behavior
	- User messages are the actual prompts
	- The model responds with an assistant message

	Temperature controls randomness:
	- 0.0 = deterministic, repetitive
	- 1.0 = creative, varied
	- 0.7 is a good balance for mock data
	*/
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo, // Using GPT-3.5 for cost-efficiency
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that generates realistic mock data in JSON format. Always respond with valid JSON only, no additional text.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,  // Balance between creativity and consistency
			MaxTokens:   4000, // Limit response size
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenAI")
	}

	// Extract the generated content
	content := resp.Choices[0].Message.Content
	log.Printf("📥 Received response from OpenAI (%d tokens used)", resp.Usage.TotalTokens)

	// Parse the JSON response
	var result struct {
		Fields []string                 `json:"fields"`
		Data   []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w (response: %s)", err, content)
	}

	// Validate the response
	if len(result.Fields) == 0 {
		return nil, nil, fmt.Errorf("OpenAI response missing fields")
	}

	if len(result.Data) == 0 {
		return nil, nil, fmt.Errorf("OpenAI response missing data")
	}

	log.Printf("✅ Successfully generated %d rows with %d fields", len(result.Data), len(result.Fields))

	return result.Data, result.Fields, nil
}

/*
CONCEPT: Context in Go

The context.Context parameter is a standard pattern in Go for:
1. Request cancellation (user cancels, timeout)
2. Deadline enforcement
3. Request-scoped values (like request IDs)

Always accept and pass context in functions that:
- Make external calls (APIs, databases)
- Can take significant time
- Should be cancellable
*/
