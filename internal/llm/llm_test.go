package llm

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestNewClient(t *testing.T) {
	// Test with default model
	client := NewClient("")
	if client.model != anthropic.ModelClaude3_7SonnetLatest {
		t.Errorf("Expected default model to be %s, got %s", anthropic.ModelClaude3_7SonnetLatest, client.model)
	}

	// Test with specified model
	testModel := "claude-3-opus-20240229"
	client = NewClient(testModel)
	if client.model != testModel {
		t.Errorf("Expected model to be %s, got %s", testModel, client.model)
	}

	// Test with options
	mockClient := anthropic.NewClient()
	client = NewClient(testModel, WithAnthropicClient(&mockClient))
	if client.model != testModel {
		t.Errorf("Expected model to be %s, got %s", testModel, client.model)
	}
}

func TestGenerateRequestSpec(t *testing.T) {
	// Skip test if no API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test; ANTHROPIC_API_KEY not set")
	}

	client := NewClient("")
	ctx := context.Background()

	testCases := []struct {
		name        string
		prompt      string
		wantMethod  string
		wantURLPart string
		wantErr     bool
	}{
		{
			name:        "Simple GET request",
			prompt:      "get https://httpbin.org/get",
			wantMethod:  "GET",
			wantURLPart: "httpbin.org/get",
			wantErr:     false,
		},
		{
			name:        "POST request with JSON",
			prompt:      "POST to httpbin.org/post with JSON body {\"name\": \"test\"}",
			wantMethod:  "POST",
			wantURLPart: "httpbin.org/post",
			wantErr:     false,
		},
		{
			name:        "Empty prompt",
			prompt:      "",
			wantMethod:  "",
			wantURLPart: "",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := client.GenerateRequestSpec(ctx, tc.prompt)

			if (err != nil) != tc.wantErr {
				t.Fatalf("GenerateRequestSpec() error = %v, wantErr %v", err, tc.wantErr)
			}

			if tc.wantErr {
				return
			}

			if spec.Method != tc.wantMethod {
				t.Errorf("Expected method %s, got %s", tc.wantMethod, spec.Method)
			}

			if spec.URL == "" || !contains(spec.URL, tc.wantURLPart) {
				t.Errorf("URL should contain %s, got %s", tc.wantURLPart, spec.URL)
			}
		})
	}
}

func TestModelError(t *testing.T) {
	// Test error message formatting
	origErr := errors.New("original error")
	modelErr := &ModelError{
		Err:     origErr,
		Message: "test message",
		Model:   "claude-3",
		Prompt:  "test prompt",
		RawJSON: "test JSON",
	}

	// Test Error() method
	errMsg := modelErr.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Test Unwrap() method
	unwrappedErr := errors.Unwrap(modelErr)
	if unwrappedErr != origErr {
		t.Errorf("Unwrap() = %v, want %v", unwrappedErr, origErr)
	}

	// Test long JSON truncation
	longJSON := ""
	for i := 0; i < 200; i++ {
		longJSON += "x"
	}

	modelErrLong := &ModelError{
		Err:     origErr,
		Message: "test message",
		Model:   "claude-3",
		RawJSON: longJSON,
	}

	errMsgLong := modelErrLong.Error()
	if len(errMsgLong) >= len(longJSON) {
		t.Error("Expected truncated JSON in error message")
	}
}

func TestContextCancellation(t *testing.T) {
	// Skip test if no API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test; ANTHROPIC_API_KEY not set")
	}

	client := NewClient("")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait a bit to ensure timeout happens
	time.Sleep(5 * time.Millisecond)

	_, err := client.GenerateRequestSpec(ctx, "get https://example.com")
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected DeadlineExceeded error, got %v", err)
	}
}

func TestCleanJSONResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain JSON",
			input:    `{"method": "GET", "url": "https://example.com"}`,
			expected: `{"method": "GET", "url": "https://example.com"}`,
		},
		{
			name: "Markdown code block",
			input: "```json\n" +
				`{"method": "GET", "url": "https://example.com"}` +
				"\n```",
			expected: `{"method": "GET", "url": "https://example.com"}`,
		},
		{
			name: "Markdown code block without language",
			input: "```\n" +
				`{"method": "GET", "url": "https://example.com"}` +
				"\n```",
			expected: `{"method": "GET", "url": "https://example.com"}`,
		},
		{
			name: "Text with JSON",
			input: "Here is the JSON object:\n\n" +
				`{"method": "GET", "url": "https://example.com"}`,
			expected: `{"method": "GET", "url": "https://example.com"}`,
		},
		{
			name: "Multiline JSON with whitespace",
			input: "```json\n" +
				`{
					"method": "GET",
					"url": "https://example.com"
				}` +
				"\n```",
			expected: `{
					"method": "GET",
					"url": "https://example.com"
				}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanJSONResponse(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s == substr || len(s) >= len(substr) && s[:len(substr)] == substr || len(s) >= len(substr) && s[len(s)-len(substr):] == substr || len(s) >= len(substr) && s[1:len(s)-1] != s && contains(s[1:], substr)
}
