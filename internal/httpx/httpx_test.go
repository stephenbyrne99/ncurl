package httpx_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stephenbyrne99/ncurl/internal/httpx"
)

func TestNewRequestSpec(t *testing.T) {
	spec := httpx.NewRequestSpec()

	if spec.Method != http.MethodGet {
		t.Errorf("Expected default method to be GET, got %s", spec.Method)
	}

	if spec.Headers == nil {
		t.Error("Expected headers map to be initialized")
	}
}

func TestRequestSpec_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		spec    *httpx.RequestSpec
		wantErr bool
	}{
		{
			name: "Valid spec",
			spec: &httpx.RequestSpec{
				Method: http.MethodGet,
				URL:    "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "Missing URL",
			spec: &httpx.RequestSpec{
				Method: http.MethodGet,
			},
			wantErr: true,
		},
		{
			name: "Empty method gets default",
			spec: &httpx.RequestSpec{
				URL: "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "Nil headers gets initialized",
			spec: &httpx.RequestSpec{
				Method:  http.MethodPost,
				URL:     "https://example.com",
				Headers: nil,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if err == nil && tc.spec.Method == "" {
				if tc.spec.Method != http.MethodGet {
					t.Errorf("Expected default method to be GET after validation, got %s", tc.spec.Method)
				}
			}

			if err == nil && tc.spec.Headers == nil {
				t.Error("Expected headers to be initialized after validation, got nil")
			}
		})
	}
}

func TestExecute(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected header X-Test-Header to be test-value, got %s", r.Header.Get("X-Test-Header"))
		}

		// Verify method
		if r.Method != http.MethodPost {
			t.Errorf("Expected method to be POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"success": true}`))
		if err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create request spec
	spec := &httpx.RequestSpec{
		Method: http.MethodPost,
		URL:    server.URL,
		Headers: map[string]string{
			"X-Test-Header": "test-value",
			"Content-Type":  "application/json",
		},
		Body: `{"test": "data"}`,
	}

	// Execute request
	resp, err := httpx.Execute(spec)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	expectedBody := `{"success": true}`
	if string(resp.Body) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, string(resp.Body))
	}
}

func TestExecuteWithDefaults(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("Expected method to be GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		// For an empty response body, no need to call w.Write()
	}))
	defer server.Close()

	// Create request spec with empty values
	spec := &httpx.RequestSpec{
		URL: server.URL,
	}

	// Execute request
	resp, err := httpx.Execute(spec)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestExecuteWithContext(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate work that takes time
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create request spec
	spec := &httpx.RequestSpec{
		Method: http.MethodGet,
		URL:    server.URL,
	}

	// Test with a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := httpx.ExecuteWithContext(ctx, spec)
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = httpx.ExecuteWithContext(ctx, spec)
	if err == nil {
		t.Error("Expected error with timeout context, got nil")
	}

	// Test with valid context
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := httpx.ExecuteWithContext(ctx, spec)
	if err != nil {
		t.Fatalf("ExecuteWithContext failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestRequestErrorsUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	requestErr := &httpx.RequestError{
		Err:     originalErr,
		Message: "test message",
		URL:     "https://example.com",
		Method:  http.MethodGet,
	}

	unwrappedErr := errors.Unwrap(requestErr)
	if !errors.Is(unwrappedErr, originalErr) {
		t.Errorf("Unwrap() = %v, want %v", unwrappedErr, originalErr)
	}
}
