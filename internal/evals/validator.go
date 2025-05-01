package evals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stephenbyrne99/ncurl/internal/httpx"
)

// ResponseValidator validates HTTP responses against expectations
type ResponseValidator struct {
	model  string
	client *http.Client
}

// ValidationResult represents the result of a response validation
type ValidationResult struct {
	IsValid            bool    `json:"is_valid"`
	SatisfactionScore  float64 `json:"satisfaction_score"`
	Reasoning          string  `json:"reasoning"`
	MissingInformation string  `json:"missing_information,omitempty"`
}

// InputValidationResult represents the result of input validation
type InputValidationResult struct {
	IsValid         bool    `json:"is_valid"`
	Clarity         float64 `json:"clarity_score"` // 0-1 score for clarity
	Completeness    float64 `json:"completeness"`  // 0-1 score for completeness
	Specificity     float64 `json:"specificity"`   // 0-1 score for specificity
	ErrorType       string  `json:"error_type,omitempty"`
	ErrorSeverity   string  `json:"error_severity,omitempty"`
	Analysis        string  `json:"analysis"`
	Recommendations string  `json:"recommendations,omitempty"`
}

// NewResponseValidator creates a new response validator
func NewResponseValidator(model string) *ResponseValidator {
	if model == "" {
		model = anthropic.ModelClaude3_7SonnetLatest
	}

	return &ResponseValidator{
		model:  model,
		client: &http.Client{},
	}
}

// ValidateResponse validates a HTTP response against the expected response
func (v *ResponseValidator) ValidateResponse(
	ctx context.Context,
	input RequestEvalInput,
	response []byte,
) (*ValidationResult, error) {
	// Prepare the input for evaluation
	evalInput := RequestEvalInput{
		NaturalLanguage:  input.NaturalLanguage,
		GeneratedMethod:  input.GeneratedMethod,
		GeneratedURL:     input.GeneratedURL,
		GeneratedHeaders: input.GeneratedHeaders,
		GeneratedBody:    input.GeneratedBody,
		ExpectedBody:     string(response),
	}

	// Use the output validation prompt
	promptTemplate := DefaultPromptTemplates.OutputValidation
	systemPrompt, err := RenderTemplate(promptTemplate.SystemPrompt, evalInput)
	if err != nil {
		return nil, fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err := RenderTemplate(promptTemplate.UserPrompt, evalInput)
	if err != nil {
		return nil, fmt.Errorf("failed to render user prompt: %w", err)
	}

	// Create Anthropic client
	client := anthropic.NewClient() // uses ANTHROPIC_API_KEY env var

	// Send the request to Anthropic
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     v.model,
		MaxTokens: 1024, // Standard token limit
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract the response content
	if len(msg.Content) == 0 {
		return nil, errors.New("empty response from Anthropic")
	}

	responseText := msg.Content[0].Text

	// Try to extract JSON from the response
	var jsonStr string
	if start := strings.Index(responseText, "{"); start != -1 {
		if end := strings.LastIndex(responseText, "}"); end != -1 && end > start {
			jsonStr = responseText[start : end+1]
		}
	}

	if jsonStr == "" {
		return nil, errors.New("could not extract JSON from response")
	}

	var result ValidationResult
	if unmarshalErr := json.Unmarshal([]byte(jsonStr), &result); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", unmarshalErr)
	}

	return &result, nil
}

// ValidateInput validates a natural language input for clarity and specificity
func (v *ResponseValidator) ValidateInput(
	ctx context.Context,
	naturalLanguage string,
) (*InputValidationResult, error) {
	// Create the prompt for input validation
	systemPrompt := `You are an expert evaluator for HTTP request generation inputs. Your task is to assess
the clarity, completeness, and specificity of a natural language description that will be used to generate an HTTP request.

Evaluation criteria:
1. Clarity: Is the intent clear? Is it obvious what kind of HTTP request is intended?
2. Completeness: Does it include all necessary information (like endpoints, data to send, etc.)?
3. Specificity: How specific is the description? Does it leave room for ambiguity?

For each criterion, assign a score from 0.0 to 1.0:
- 1.0: Excellent - perfectly clear, complete, and specific
- 0.7: Good - mostly clear with minor ambiguities
- 0.4: Fair - somewhat unclear or incomplete
- 0.1: Poor - very unclear, ambiguous, or incomplete

Provide your assessment as JSON with detailed reasoning.`

	userPrompt := fmt.Sprintf(
		`Please evaluate this natural language description that will be used to generate an HTTP request:

"%s"

Provide your evaluation in JSON format with these fields:
{
  "is_valid": true|false,
  "clarity_score": 0.0 to 1.0,
  "completeness": 0.0 to 1.0,
  "specificity": 0.0 to 1.0,
  "error_type": "ambiguity|incomplete_information|invalid_syntax|none",
  "error_severity": "low|medium|high|none",
  "analysis": "Your detailed assessment",
  "recommendations": "Suggestions for improving the input"
}`,
		naturalLanguage,
	)

	// Create Anthropic client
	client := anthropic.NewClient() // uses ANTHROPIC_API_KEY env var

	// Send the request to Anthropic
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     v.model,
		MaxTokens: 1024, // Standard token limit
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	// Extract the response content
	if len(msg.Content) == 0 {
		return nil, errors.New("empty response from Anthropic")
	}

	responseText := msg.Content[0].Text

	// Try to extract JSON from the response
	var jsonStr string
	if start := strings.Index(responseText, "{"); start != -1 {
		if end := strings.LastIndex(responseText, "}"); end != -1 && end > start {
			jsonStr = responseText[start : end+1]
		}
	}

	if jsonStr == "" {
		return nil, errors.New("could not extract JSON from response")
	}

	var result InputValidationResult
	if unmarshalErr := json.Unmarshal([]byte(jsonStr), &result); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse validation response: %w", unmarshalErr)
	}

	return &result, nil
}

// ValidateURL validates if a URL is well-formed and secure
func ValidateURL(urlStr string) (bool, string) {
	if urlStr == "" {
		return false, "URL is empty"
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Sprintf("Invalid URL format: %v", err)
	}

	// Check scheme
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return false, fmt.Sprintf("Invalid scheme: %s (expected http or https)", parsedURL.Scheme)
	}

	// Prefer HTTPS
	if parsedURL.Scheme == "http" {
		return true, "Warning: Using HTTP instead of HTTPS"
	}

	// Check for localhost or private IPs
	host := parsedURL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.16.") {
		return true, "Warning: Using a local/private address"
	}

	return true, ""
}

// ValidateHeaders validates HTTP headers for common issues
func ValidateHeaders(headers map[string]string) (bool, string) {
	if len(headers) == 0 {
		return true, ""
	}

	var issues []string

	for k, v := range headers {
		// Check for sensitive information in headers
		sensitiveKeys := []string{
			"authorization",
			"api-key",
			"apikey",
			"secret",
			"password",
			"token",
		}

		for _, sensitive := range sensitiveKeys {
			if strings.Contains(strings.ToLower(k), sensitive) {
				// Don't log the actual value for security reasons
				const sensitiveMinLength = 20
				if len(v) > sensitiveMinLength {
					issues = append(
						issues,
						fmt.Sprintf("Header '%s' may contain sensitive information", k),
					)
				}
			}
		}

		// Check for content-type header when necessary
		if strings.EqualFold(k, "content-type") {
			contentType := strings.ToLower(v)
			if !strings.Contains(contentType, "application/json") &&
				!strings.Contains(contentType, "application/xml") &&
				!strings.Contains(contentType, "text/plain") &&
				!strings.Contains(contentType, "multipart/form-data") &&
				!strings.Contains(contentType, "application/x-www-form-urlencoded") {
				issues = append(issues, fmt.Sprintf("Unusual Content-Type: %s", v))
			}
		}
	}

	if len(issues) > 0 {
		return false, strings.Join(issues, "; ")
	}

	return true, ""
}

// ValidateBody validates request body content
func ValidateBody(body, contentType string) (bool, string) {
	if body == "" {
		return true, ""
	}

	contentType = strings.ToLower(contentType)

	// Validate JSON body
	if strings.Contains(contentType, "application/json") ||
		strings.HasPrefix(strings.TrimSpace(body), "{") {
		var js interface{}
		err := json.Unmarshal([]byte(body), &js)
		if err != nil {
			return false, fmt.Sprintf("Invalid JSON body: %v", err)
		}
	}

	// Check for potentially sensitive information
	sensitivePatterns := []string{
		`"password"\s*:\s*"[^"]*"`,
		`"api[-_]?key"\s*:\s*"[^"]*"`,
		`"secret"\s*:\s*"[^"]*"`,
		`"token"\s*:\s*"[^"]*"`,
	}

	for _, pattern := range sensitivePatterns {
		match, _ := regexp.MatchString(pattern, body)
		if match {
			return false, "Body contains potentially sensitive information"
		}
	}

	return true, ""
}

// ValidateRequestSpec validates an entire request specification
func ValidateRequestSpec(spec *httpx.RequestSpec) (bool, map[string]string) {
	if spec == nil {
		return false, map[string]string{"error": "Request specification is nil"}
	}

	validations := make(map[string]string)

	// Validate URL
	urlValid, urlMsg := ValidateURL(spec.URL)
	if !urlValid || urlMsg != "" {
		validations["url"] = urlMsg
	}

	// Validate headers
	headersValid, headersMsg := ValidateHeaders(spec.Headers)
	if !headersValid || headersMsg != "" {
		validations["headers"] = headersMsg
	}

	// Get content type for body validation
	contentType := ""
	for k, v := range spec.Headers {
		if strings.EqualFold(k, "content-type") {
			contentType = v
			break
		}
	}

	// Validate body
	bodyValid, bodyMsg := ValidateBody(spec.Body, contentType)
	if !bodyValid || bodyMsg != "" {
		validations["body"] = bodyMsg
	}

	// Check HTTP method
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	if !validMethods[spec.Method] {
		validations["method"] = fmt.Sprintf("Invalid or uncommon HTTP method: %s", spec.Method)
	}

	// Validate method and body combination
	if (spec.Method == http.MethodGet || spec.Method == http.MethodHead) && spec.Body != "" {
		validations["method_body"] = fmt.Sprintf("%s requests should not have a body", spec.Method)
	}

	// Overall validation
	isValid := len(validations) == 0
	return isValid, validations
}

// FetchAndValidateResponse executes a request and validates the response
func (v *ResponseValidator) FetchAndValidateResponse(
	ctx context.Context,
	spec *httpx.RequestSpec,
	evalInput RequestEvalInput,
) (*ValidationResult, error) {
	// Execute the request
	response, err := httpx.ExecuteWithContext(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Use response body (already read in the httpx.ExecuteWithContext call)
	respBody := response.Body

	// Validate the response
	result, err := v.ValidateResponse(ctx, evalInput, respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to validate response: %w", err)
	}

	return result, nil
}
