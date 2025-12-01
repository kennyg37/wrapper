# Mock Data Generator - Go/Fiber Backend

A powerful backend API that uses OpenAI's ChatGPT to generate realistic mock data based on scenarios. Built with Go and Fiber framework.

## Features

- **AI-Powered Data Generation**: Uses OpenAI GPT to generate contextually appropriate mock data
- **Multiple Export Formats**: JSON, CSV, Markdown, and SQL
- **PostgreSQL Storage**: Persistent storage of generation requests and datasets
- **RESTful API**: Clean, well-documented API endpoints
- **Type-Safe**: Strongly typed with Go's type system
- **Tested**: Comprehensive unit tests for core functionality
- **Fast**: Built on Fiber (fasthttp) for high performance

## Prerequisites

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **PostgreSQL 12+**: [Download PostgreSQL](https://www.postgresql.org/download/)
- **OpenAI API Key**: [Get API Key](https://platform.openai.com/api-keys)

## Quick Start

### 1. Clone and Navigate

```bash
cd backend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Database

Create a PostgreSQL database:

```bash
# Using psql
createdb mockdata_generator

# Or using SQL
psql -U postgres
CREATE DATABASE mockdata_generator;
\q
```

### 4. Configure Environment

Copy the example environment file and edit it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
PORT=3000
ENVIRONMENT=development
OPENAI_API_KEY=sk-your-actual-api-key-here
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_postgres_password
DB_NAME=mockdata_generator
DB_SSL_MODE=disable
CORS_ORIGINS=http://localhost:5173,http://localhost:4173
```

### 5. Run the Server

```bash
# Development mode (with auto-reload, install air first: go install github.com/cosmtrek/air@latest)
air

# Or run directly
go run cmd/api/main.go
```

The server will start on `http://localhost:3000`

## API Documentation

### Endpoints

#### Health Check
```http
GET /api/health
```

#### Generate Mock Data
```http
POST /api/generate
Content-Type: application/json

{
  "scenario": "10 e-commerce products with names, prices, and descriptions",
  "row_count": 10
}
```

**Response:**
```json
{
  "id": 1,
  "status": "completed",
  "message": "Successfully generated 10 rows of mock data",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### List All Requests
```http
GET /api/requests
```

#### Get Request Status
```http
GET /api/requests/:id
```

#### Get Generated Data
```http
GET /api/data/:id
```

**Response:**
```json
{
  "id": 1,
  "request_id": 1,
  "scenario": "e-commerce products...",
  "data": [
    {
      "id": 1,
      "name": "Wireless Headphones",
      "price": 79.99,
      "description": "High-quality wireless headphones..."
    }
  ],
  "field_names": ["id", "name", "price", "description"],
  "row_count": 10,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Export Data
```http
GET /api/data/:id/export?format=csv
GET /api/data/:id/export?format=json
GET /api/data/:id/export?format=markdown
GET /api/data/:id/export?format=sql&table=products
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestExportService_ToCSV ./internal/services

# Run tests verbosely
go test -v ./...
```

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── database.go          # Database connection & migrations
│   ├── handlers/
│   │   └── handlers.go          # HTTP request handlers
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   ├── logger.go            # Request logging
│   │   └── recovery.go          # Panic recovery
│   ├── models/
│   │   ├── models.go            # Data models
│   │   ├── errors.go            # Custom errors
│   │   └── models_test.go       # Model tests
│   └── services/
│       ├── openai.go            # OpenAI integration
│       ├── export.go            # Export services
│       └── export_test.go       # Export tests
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── .env.example                 # Environment template
└── README.md                    # This file
```


## Deployment

### Using Docker

```dockerfile
# Create Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 3000
CMD ["./main"]
```

Build and run:
```bash
docker build -t mockdata-api .
docker run -p 3000:3000 --env-file .env mockdata-api
```

### Environment Variables for Production

Remember to set these in your production environment:
- Set `ENVIRONMENT=production`
- Use strong database credentials
- Enable SSL for database (`DB_SSL_MODE=require`)
- Set appropriate CORS origins
- Use a reverse proxy (nginx, Caddy)
- Enable HTTPS

## License

This is a learning project. Feel free to use and modify as needed.

## Contributing

This is a learning project, but improvements are welcome:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues or questions:
- Check the troubleshooting section
- Review the code comments (extensive explanations)
- Open an issue on GitHub
