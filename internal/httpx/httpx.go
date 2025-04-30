// Package httpx provides HTTP request execution functionality
package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Common errors that can be returned by this package
var (
	ErrInvalidRequest = errors.New("invalid request specification")
	ErrRequestFailed  = errors.New("request execution failed")
	ErrReadResponse   = errors.New("failed to read response")
)

// RequestError represents an error that occurred during request execution
type RequestError struct {
	Err     error
	Message string
	URL     string
	Method  string
}

// Error implements the error interface
func (e *RequestError) Error() string {
	return fmt.Sprintf("%s: %s %s: %v", e.Message, e.Method, e.URL, e.Err)
}

// Unwrap returns the underlying error
func (e *RequestError) Unwrap() error {
	return e.Err
}

// RequestSpec represents a structured HTTP request specification
type RequestSpec struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// Response bundles the http.Response with the fully-read body so the
// caller can safely access it after the Response is closed.
type Response struct {
	*http.Response
	Body []byte
}

// NewRequestSpec creates a new RequestSpec with default values
func NewRequestSpec() *RequestSpec {
	return &RequestSpec{
		Method:  http.MethodGet,
		Headers: make(map[string]string),
	}
}

// Validate checks if the RequestSpec has all required fields
func (rs *RequestSpec) Validate() error {
	if rs.URL == "" {
		return fmt.Errorf("%w: missing URL", ErrInvalidRequest)
	}

	// Apply minimal defaults if not set
	if rs.Method == "" {
		rs.Method = http.MethodGet
	}
	if rs.Headers == nil {
		rs.Headers = make(map[string]string)
	}

	return nil
}

// Execute sends the HTTP request defined by the RequestSpec and returns a Response
func Execute(spec *RequestSpec) (*Response, error) {
	return ExecuteWithContext(context.Background(), spec)
}

// ExecuteWithContext sends the HTTP request with context and returns a Response
func ExecuteWithContext(ctx context.Context, spec *RequestSpec) (*Response, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	reqBody := strings.NewReader(spec.Body)
	req, err := http.NewRequestWithContext(ctx, spec.Method, spec.URL, reqBody)
	if err != nil {
		return nil, &RequestError{
			Err:     err,
			Message: "failed to create request",
			URL:     spec.URL,
			Method:  spec.Method,
		}
	}

	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // Default timeout
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, &RequestError{
			Err:     fmt.Errorf("%w: %v", ErrRequestFailed, doErr),
			Message: "request failed",
			URL:     spec.URL,
			Method:  spec.Method,
		}
	}

	var closeErr error
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && closeErr == nil {
			closeErr = cerr
		}
	}()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &RequestError{
			Err:     fmt.Errorf("%w: %v", ErrReadResponse, readErr),
			Message: "failed to read response body",
			URL:     spec.URL,
			Method:  spec.Method,
		}
	}

	result := &Response{Response: resp, Body: body}

	if closeErr != nil {
		// Return response but with an error
		err := &RequestError{
			Err:     closeErr,
			Message: "failed to close response body",
			URL:     spec.URL,
			Method:  spec.Method,
		}
		// Log the error but still return the response
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	return result, nil
}
