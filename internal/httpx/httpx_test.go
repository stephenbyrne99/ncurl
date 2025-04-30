package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRequestSpec(t *testing.T) {
	spec := NewRequestSpec()

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
		spec    *RequestSpec
		wantErr bool
	}{
		{
			name: "Valid spec",
			spec: &RequestSpec{
				Method: http.MethodGet,
				URL:    "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "Missing URL",
			spec: &RequestSpec{
				Method: http.MethodGet,
			},
			wantErr: true,
		},
		{
			name: "Empty method gets default",
			spec: &RequestSpec{
				URL: "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "Nil headers gets initialized",
			spec: &RequestSpec{
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
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// Create request spec
	spec := &RequestSpec{
		Method: http.MethodPost,
		URL:    server.URL,
		Headers: map[string]string{
			"X-Test-Header": "test-value",
			"Content-Type":  "application/json",
		},
		Body: `{"test": "data"}`,
	}

	// Execute request
	resp, err := Execute(spec)
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
	}))
	defer server.Close()

	// Create request spec with empty values
	spec := &RequestSpec{
		URL: server.URL,
	}

	// Execute request
	resp, err := Execute(spec)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate work that takes time
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create request spec
	spec := &RequestSpec{
		Method: http.MethodGet,
		URL:    server.URL,
	}

	// Test with a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := ExecuteWithContext(ctx, spec)
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}

	// Test with timeout context
	ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = ExecuteWithContext(ctx, spec)
	if err == nil {
		t.Error("Expected error with timeout context, got nil")
	}

	// Test with valid context
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ExecuteWithContext(ctx, spec)
	if err != nil {
		t.Fatalf("ExecuteWithContext failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestRequestErrorsUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	requestErr := &RequestError{
		Err:     originalErr,
		Message: "test message",
		URL:     "https://example.com",
		Method:  http.MethodGet,
	}

	unwrappedErr := errors.Unwrap(requestErr)
	if unwrappedErr != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrappedErr, originalErr)
	}
}
