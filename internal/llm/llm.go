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
		if len(rawJSON) > 100 {
			rawJSON = rawJSON[:97] + "..."
		}
		msg += fmt.Sprintf(" (raw: %s)", rawJSON)
	}
	return msg
}

// Unwrap returns the underlying error
func (e *ModelError) Unwrap() error {
	return e.Err
}

// cleanJSONResponse removes markdown formatting from a model response
// to extract the actual JSON content.
func cleanJSONResponse(input string) string {
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
	model           string
}

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// WithAnthropicClient allows setting a custom Anthropic client
func WithAnthropicClient(client anthropic.Client) ClientOption {
	return func(c *Client) {
		c.anthropicClient = &client
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
		model:           model,
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
			Model:   c.model,
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
`

	msg, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: 1024,
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
			Err:     fmt.Errorf("%w: %v", ErrModelFailure, err),
			Message: "failed to execute model request",
			Model:   c.model,
			Prompt:  naturalLanguage,
		}
	}

	// Check for empty responses
	if len(msg.Content) == 0 {
		return nil, &ModelError{
			Err:     ErrEmptyResponse,
			Message: "model returned empty content",
			Model:   c.model,
			Prompt:  naturalLanguage,
		}
	}

	// Claude streams content blocks; we expect the first to be the JSON text.
	rawJSON := msg.Content[0].Text

	// Clean up the response - sometimes Claude returns markdown-formatted JSON
	cleanJSON := cleanJSONResponse(rawJSON)

	var spec httpx.RequestSpec
	if err := json.Unmarshal([]byte(cleanJSON), &spec); err != nil {
		return nil, &ModelError{
			Err:     fmt.Errorf("%w: %v", ErrInvalidJSON, err),
			Message: "failed to parse model response as JSON",
			Model:   c.model,
			Prompt:  naturalLanguage,
			RawJSON: rawJSON,
		}
	}

	// Validate the RequestSpec
	if err := spec.Validate(); err != nil {
		return nil, &ModelError{
			Err:     err,
			Message: "model generated invalid request specification",
			Model:   c.model,
			Prompt:  naturalLanguage,
			RawJSON: rawJSON,
		}
	}

	return &spec, nil
}