package evals

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DefaultTestCases returns a set of common use cases for evaluating ncurl
func DefaultTestCases() []EvalCase {
	return []EvalCase{
		// Basic HTTP methods
		{
			ID:             "basic-get",
			Description:    "Basic GET request",
			Input:          "get data from https://httpbin.org/get",
			ExpectedMethod: "GET",
			ExpectedURL:    "httpbin.org/get",
		},
		{
			ID:             "basic-post",
			Description:    "Basic POST request with JSON body",
			Input:          "post {\"name\": \"John Doe\", \"email\": \"john@example.com\"} to https://httpbin.org/post",
			ExpectedMethod: "POST",
			ExpectedURL:    "httpbin.org/post",
			ExpectedBody:   "\"name\": \"John Doe\"",
		},
		{
			ID:             "basic-put",
			Description:    "Basic PUT request",
			Input:          "put updated user data {\"id\": 1, \"name\": \"Updated Name\"} to https://httpbin.org/put",
			ExpectedMethod: "PUT",
			ExpectedURL:    "httpbin.org/put",
			ExpectedBody:   "\"name\": \"Updated Name\"",
		},
		{
			ID:             "basic-delete",
			Description:    "Basic DELETE request",
			Input:          "delete user with id 123 from https://httpbin.org/delete",
			ExpectedMethod: "DELETE",
			ExpectedURL:    "httpbin.org/delete",
		},
		// Common APIs and scenarios
		{
			ID:             "weather-api",
			Description:    "Weather API request",
			Input:          "get the current weather for London",
			ExpectedMethod: "GET",
			ExpectedURL:    "weather",
		},
		{
			ID:             "github-api",
			Description:    "GitHub API with authentication",
			Input:          "get my GitHub repositories with authorization token FAKE_TOKEN_123",
			ExpectedMethod: "GET",
			ExpectedURL:    "github.com/user/repos",
			ExpectedHeaders: map[string]string{
				"Authorization": "token FAKE_TOKEN_123",
			},
		},
		{
			ID:             "json-api",
			Description:    "JSON API with query parameters",
			Input:          "get posts from jsonplaceholder with userId=1 and completed=true",
			ExpectedMethod: "GET",
			ExpectedURL:    "jsonplaceholder",
		},
		// Headers and authentication
		{
			ID:             "custom-headers",
			Description:    "Custom headers",
			Input:          "get data from https://httpbin.org/headers with custom header X-Custom-Header: test123 and Accept: application/json",
			ExpectedMethod: "GET",
			ExpectedURL:    "httpbin.org/headers",
			ExpectedHeaders: map[string]string{
				"X-Custom-Header": "test123",
				"Accept":          "application/json",
			},
		},
		{
			ID:             "basic-auth",
			Description:    "Basic authentication",
			Input:          "get data from https://httpbin.org/basic-auth with username 'user' and password 'pass'",
			ExpectedMethod: "GET",
			ExpectedURL:    "httpbin.org/basic-auth",
			ExpectedHeaders: map[string]string{
				"Authorization": "Basic",
			},
		},
		// Complex scenarios
		{
			ID:             "complex-json",
			Description:    "Complex JSON body",
			Input:          "post a new user to https://api.example.com/users with json body that has name 'Jane Smith', email 'jane@example.com', address with street '123 Main St', city 'Boston', and zip '02101', and includes an array of phone numbers: '555-1234' (home) and '555-5678' (cell)",
			ExpectedMethod: "POST",
			ExpectedURL:    "api.example.com/users",
			ExpectedBody:   "\"name\": \"Jane Smith\"",
		},
		{
			ID:             "natural-language",
			Description:    "Natural language request",
			Input:          "show me the current Bitcoin price",
			ExpectedMethod: "GET",
			ExpectedURL:    "bitcoin",
		},
		// Error cases
		{
			ID:             "malformed-request",
			Description:    "Malformed or ambiguous request",
			Input:          "data send now",
			ExpectedMethod: "", // Any method is acceptable
			ExpectedURL:    "", // Any URL is acceptable
		},

		// 10 Additional developer-focused test cases
		{
			ID:             "api-versioning",
			Description:    "API versioning in URL",
			Input:          "get users from API v2",
			ExpectedMethod: "GET",
			ExpectedURL:    "v2",
		},
		{
			ID:             "gzip-encoding",
			Description:    "Request with gzip encoding",
			Input:          "get data from httpbin with gzip compression",
			ExpectedMethod: "GET",
			ExpectedURL:    "httpbin",
			ExpectedHeaders: map[string]string{
				"Accept-Encoding": "gzip",
			},
		},
		{
			ID:             "jwt-auth",
			Description:    "JWT authentication",
			Input:          "get user profile with JWT eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			ExpectedMethod: "GET",
			ExpectedURL:    "profile",
			ExpectedHeaders: map[string]string{
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			},
		},
		{
			ID:             "graphql-query",
			Description:    "GraphQL query",
			Input:          "post GraphQL query { user(id: \"123\") { name, email } } to https://api.example.com/graphql",
			ExpectedMethod: "POST",
			ExpectedURL:    "graphql",
			ExpectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedBody: "query",
		},
		{
			ID:             "rest-pagination",
			Description:    "REST API with pagination",
			Input:          "get page 2 with 20 items per page from users API",
			ExpectedMethod: "GET",
			ExpectedURL:    "page=2",
		},
		{
			ID:             "file-upload",
			Description:    "Multipart file upload",
			Input:          "upload a file named 'report.pdf' to https://upload.example.com/files",
			ExpectedMethod: "POST",
			ExpectedURL:    "upload",
			ExpectedHeaders: map[string]string{
				"Content-Type": "multipart/form-data",
			},
		},
		{
			ID:             "api-key-query",
			Description:    "API key in query string",
			Input:          "get weather data with API key abc123xyz",
			ExpectedMethod: "GET",
			ExpectedURL:    "key=abc123xyz",
		},
		{
			ID:             "conditional-request",
			Description:    "Conditional request with If-Modified-Since",
			Input:          "get resource only if modified since 2023-01-01",
			ExpectedMethod: "GET",
			ExpectedHeaders: map[string]string{
				"If-Modified-Since": "2023-01-01",
			},
		},
		{
			ID:             "cors-preflight",
			Description:    "CORS preflight request",
			Input:          "send preflight CORS request to https://api.example.com/data with origin https://myapp.com",
			ExpectedMethod: "OPTIONS",
			ExpectedURL:    "api.example.com/data",
			ExpectedHeaders: map[string]string{
				"Origin":                        "https://myapp.com",
				"Access-Control-Request-Method": "GET",
			},
		},
		{
			ID:             "api-rate-limiting",
			Description:    "API with rate limiting headers",
			Input:          "get GitHub API rate limit status",
			ExpectedMethod: "GET",
			ExpectedURL:    "github.com/rate_limit",
		},

		// Local development test cases
		{
			ID:             "localhost-default",
			Description:    "Default localhost request",
			Input:          "get data from localhost",
			ExpectedMethod: "GET",
			ExpectedURL:    "localhost:3000",
		},
		{
			ID:             "localhost-custom-port",
			Description:    "Localhost with custom port",
			Input:          "get data from localhost:8080/api",
			ExpectedMethod: "GET",
			ExpectedURL:    "localhost:8080",
		},
		{
			ID:             "nextjs-api-route",
			Description:    "Next.js API route on localhost",
			Input:          "get data from Next.js API route /api/users",
			ExpectedMethod: "GET",
			ExpectedURL:    "localhost:3000/api/users",
		},
		{
			ID:             "localhost-post-json",
			Description:    "POST JSON to localhost API",
			Input:          "post user data {\"name\": \"Test User\", \"email\": \"test@example.com\"} to localhost api endpoint /users",
			ExpectedMethod: "POST",
			ExpectedURL:    "localhost:3000/users",
			ExpectedBody:   "\"name\": \"Test User\"",
		},
		{
			ID:             "localhost-graphql",
			Description:    "GraphQL query to localhost",
			Input:          "send GraphQL query to localhost { users { id name } }",
			ExpectedMethod: "POST",
			ExpectedURL:    "localhost:3000/graphql",
			ExpectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedBody: "query",
		},
		{
			ID:             "localhost-express-api",
			Description:    "Express.js API on custom port",
			Input:          "get users from Express API running on port 5000",
			ExpectedMethod: "GET",
			ExpectedURL:    "localhost:5000",
		},
		{
			ID:             "localhost-auth",
			Description:    "Localhost with authentication",
			Input:          "get data from localhost:4000 with bearer token local-dev-token",
			ExpectedMethod: "GET",
			ExpectedURL:    "localhost:4000",
			ExpectedHeaders: map[string]string{
				"Authorization": "Bearer local-dev-token",
			},
		},
	}
}

// Load test cases from a JSON file
func LoadTestCasesFromFile(filePath string) ([]EvalCase, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to open test cases file: %w", err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %w", err)
	}

	var cases []EvalCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, fmt.Errorf("failed to parse test cases: %w", err)
	}

	return cases, nil
}

// SaveTestCasesToFile saves test cases to a JSON file
func SaveTestCasesToFile(cases []EvalCase, filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(cases, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test cases: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write test cases file: %w", err)
	}

	return nil
}

// CreateDefaultTestCasesFile saves the default test cases to a specified file
func CreateDefaultTestCasesFile(filePath string) error {
	return SaveTestCasesToFile(DefaultTestCases(), filePath)
}
