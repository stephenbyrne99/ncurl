// Package llm provides a wrapper for interacting with language models
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stephenbyrne99/ncurl/internal/httpx"
)

// Common errors that can be returned by this package
var (
	ErrEmptyResponse  = errors.New("empty response from model")
	ErrInvalidJSON    = errors.New("invalid JSON from model")
	ErrModelFailure   = errors.New("model processing failed")
	ErrInvalidRequest = errors.New("invalid request to model")
)

// ModelError represents an error that occurred during model processing
type ModelError struct {
	Err     error
	Message string
	Model   string
	Prompt  string
	RawJSON string
}

// Error implements the error interface
func (e *ModelError) Error() string {
	msg := fmt.Sprintf("%s: model=%s: %v", e.Message, e.Model, e.Err)
	if e.RawJSON != "" {
		// Truncate long raw JSON responses
		rawJSON := e.RawJSON
		const maxJSONLength = 100
		const ellipsisLength = 3
		if len(rawJSON) > maxJSONLength {
			rawJSON = rawJSON[:maxJSONLength-ellipsisLength] + "..."
		}
		msg += fmt.Sprintf(" (raw: %s)", rawJSON)
	}
	return msg
}

// Unwrap returns the underlying error
func (e *ModelError) Unwrap() error {
	return e.Err
}

// CleanJSONResponse removes markdown formatting from a model response
// to extract the actual JSON content.
func CleanJSONResponse(input string) string {
	// Remove markdown code block backticks and language annotations
	re := regexp.MustCompile("```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		// Return just the content within the code block
		return strings.TrimSpace(matches[1])
	}

	// If no code block found, try to find json content directly
	jsonStart := strings.Index(input, "{")
	jsonEnd := strings.LastIndex(input, "}")

	if jsonStart >= 0 && jsonEnd > jsonStart {
		return strings.TrimSpace(input[jsonStart : jsonEnd+1])
	}

	// If all else fails, return the original input
	return strings.TrimSpace(input)
}

// Client provides methods for translating natural language to HTTP requests
type Client struct {
	anthropicClient *anthropic.Client
	Model           string // Exported for testing
}

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// WithAnthropicClient allows setting a custom Anthropic client
func WithAnthropicClient(client *anthropic.Client) ClientOption {
	return func(c *Client) {
		c.anthropicClient = client
	}
}

// NewClient creates a new LLM client with the specified model
func NewClient(model string, opts ...ClientOption) *Client {
	if model == "" {
		model = anthropic.ModelClaude3_7SonnetLatest
	}

	// Create default client
	client := anthropic.NewClient() // reads $ANTHROPIC_API_KEY
	c := &Client{
		anthropicClient: &client,
		Model:           model,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GenerateRequestSpec prompts the LLM to translate natural language into a RequestSpec
func (c *Client) GenerateRequestSpec(ctx context.Context, naturalLanguage string) (*httpx.RequestSpec, error) {
	// Check for empty prompt
	if naturalLanguage == "" {
		return nil, &ModelError{
			Err:     ErrInvalidRequest,
			Message: "empty natural language prompt",
			Model:   c.Model,
		}
	}

	// Check for context cancellation early
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	systemPrompt := `
You are a translator that converts natural‑language descriptions of HTTP
requests into a JSON object with the following shape:
{
  "method":   "GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS",
  "url":      "https://example.com/path",
  "headers":  {"Header-Name": "value", ...}, // optional
  "body":     "raw body as string"            // optional
}
Return **only** valid JSON with those exact keys (lower‑case) and no
explanation or additional text.

Guidelines:
1. Never use example.com for the url, always point to a real API
2. For cryptocurrency requests, use appropriate public APIs (e.g., coindesk, binance, etc.)
3. For weather data, use appropriate weather APIs (e.g., openweathermap, weatherapi, etc.)
4. When authentication credentials are provided, include them as appropriate headers or URL parameters
5. When specific IDs or query parameters are mentioned, include them in the URL or query string
6. When the user provides explicit JSON in the prompt, use it exactly as provided
7. Convert ambiguous natural language into the most likely intended HTTP request

Authentication and Headers:
8. For JWT tokens, use the entire token in the Authorization header (Bearer [token])
9. For API keys, follow the exact format mentioned in the input (key=xyz, appid=xyz, etc.)
10. For Basic auth, include basic auth in the Authorization header (Basic [base64])
11. For If-Modified-Since dates, format correctly as HTTP date (e.g., Sun, 01 Jan 2023 00:00:00 GMT)

URL and Endpoint Guidelines:
12. When a domain is explicitly provided (like api.example.com), always use it exactly as given
13. When API versioning is mentioned (like "v2"), include it in the path (/v2/endpoint)
14. For profile requests, use appropriate endpoint (/profile or /user/profile)
15. For uploads, use appropriate content type (multipart/form-data) and boundary

Local development guidelines:
16. For localhost requests without a port, use port 3000 by default (localhost:3000)
17. Always use the specific port if mentioned (e.g., localhost:8080 or localhost:5000)
18. For Next.js API routes, use localhost:3000/api/[route] unless another port is specified
19. For regular API endpoints on localhost, do NOT add /api unless specifically mentioned
20. For GraphQL queries to localhost, use POST to localhost:[port]/graphql with appropriate Content-Type
21. Include authentication tokens when mentioned for localhost requests
22. For other local frameworks (Express, Flask, Rails, etc.), use appropriate port conventions

Your goal is to accurately translate what the user wants into a proper HTTP request, including correctly handling local development scenarios.
`

	msg, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.Model,
		MaxTokens: 1024, // A standard token limit for this type of request
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(naturalLanguage)),
		},
	})

	// Handle context cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Handle API errors
	if err != nil {
		return nil, &ModelError{
			Err:     fmt.Errorf("%w: %w", ErrModelFailure, err),
			Message: "failed to execute model request",
			Model:   c.Model,
			Prompt:  naturalLanguage,
		}
	}

	// Check for empty responses
	if len(msg.Content) == 0 {
		return nil, &ModelError{
			Err:     ErrEmptyResponse,
			Message: "model returned empty content",
			Model:   c.Model,
			Prompt:  naturalLanguage,
		}
	}

	// Claude streams content blocks; we expect the first to be the JSON text.
	rawJSON := msg.Content[0].Text

	// Clean up the response - sometimes Claude returns markdown-formatted JSON
	cleanJSON := CleanJSONResponse(rawJSON)

	var spec httpx.RequestSpec
	if unmarshalErr := json.Unmarshal([]byte(cleanJSON), &spec); unmarshalErr != nil {
		return nil, &ModelError{
			Err:     fmt.Errorf("%w: %w", ErrInvalidJSON, unmarshalErr),
			Message: "failed to parse model response as JSON",
			Model:   c.Model,
			Prompt:  naturalLanguage,
			RawJSON: rawJSON,
		}
	}

	// Validate the RequestSpec
	if validateErr := spec.Validate(); validateErr != nil {
		return nil, &ModelError{
			Err:     validateErr,
			Message: "model generated invalid request specification",
			Model:   c.Model,
			Prompt:  naturalLanguage,
			RawJSON: rawJSON,
		}
	}

	return &spec, nil
}
