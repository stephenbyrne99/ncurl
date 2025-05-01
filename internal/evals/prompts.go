package evals

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/anthropics/anthropic-sdk-go"
)

// RequestEvalInput represents input data for request evaluation prompts
type RequestEvalInput struct {
	NaturalLanguage    string `json:"natural_language"`
	GeneratedMethod    string `json:"generated_method"`
	GeneratedURL       string `json:"generated_url"`
	GeneratedHeaders   string `json:"generated_headers"`
	GeneratedBody      string `json:"generated_body"`
	ExpectedMethod     string `json:"expected_method"`
	ExpectedURL        string `json:"expected_url"`
	ExpectedHeaders    string `json:"expected_headers"`
	ExpectedBody       string `json:"expected_body"`
	EvaluationCriteria string `json:"evaluation_criteria,omitempty"`
}

// PromptTemplate for evaluating ncurl requests
type PromptTemplate struct {
	SystemPrompt string
	UserPrompt   string
}

// DefaultPromptTemplates contains the default templates
var DefaultPromptTemplates = struct {
	RequestEvaluation PromptTemplate
	ErrorAnalysis     PromptTemplate
	OutputValidation  PromptTemplate
}{
	RequestEvaluation: PromptTemplate{
		SystemPrompt: `You are an expert evaluator for HTTP request generation. Your task is to evaluate how well a generated HTTP request matches the expected request based on a natural language description.

Evaluation criteria:
1. Method: Is the HTTP method (GET, POST, PUT, etc.) appropriate for the described action?
2. URL: Does the URL match the expected endpoint or contain expected components?
3. Headers: Are the necessary headers included?
4. Body: Does the request body contain the expected data structure and values?

For each criterion, assign a score from 0.0 to 1.0:
- 1.0: Perfect match
- 0.8: Minor issues but functionally correct
- 0.5: Partially correct with significant issues
- 0.2: Major issues that would prevent the request from working correctly
- 0.0: Completely incorrect or missing

Provide a final overall score (average of all criteria) and detailed feedback.`,

		UserPrompt: `Evaluate this HTTP request generation:

Natural language input:
{{.NaturalLanguage}}

Generated request:
Method: {{.GeneratedMethod}}
URL: {{.GeneratedURL}}
Headers: {{.GeneratedHeaders}}
Body: {{.GeneratedBody}}

Expected values:
Method: {{.ExpectedMethod}}
URL should contain: {{.ExpectedURL}}
Headers should include: {{.ExpectedHeaders}}
Body should contain: {{.ExpectedBody}}

{{if .EvaluationCriteria}}Additional evaluation criteria:
{{.EvaluationCriteria}}{{end}}

Provide your evaluation in JSON format with these fields:
{
  "method_score": 0.0 to 1.0,
  "url_score": 0.0 to 1.0,
  "headers_score": 0.0 to 1.0,
  "body_score": 0.0 to 1.0,
  "overall_score": 0.0 to 1.0,
  "reasoning": "Your detailed assessment explaining scores",
  "suggestions": "Suggestions for improvement"
}`,
	},

	ErrorAnalysis: PromptTemplate{
		SystemPrompt: `You are an expert error analyst for HTTP request generation systems. Your task is to analyze why a natural language description failed to generate a valid HTTP request and provide insights into the failure.`,

		UserPrompt: `The following natural language input failed to generate a valid HTTP request:

Natural language input:
{{.NaturalLanguage}}

Error message:
{{.GeneratedBody}}

Please analyze this failure and provide insights in JSON format:
{
  "error_type": "ambiguity|parsing|invalid_structure|unsupported_feature|other",
  "error_severity": "low|medium|high",
  "analysis": "Your detailed analysis of why this failed",
  "suggestions": "Suggestions for fixing the input or improving the system"
}`,
	},

	OutputValidation: PromptTemplate{
		SystemPrompt: `You are an expert validator for HTTP responses. Your task is to determine if an HTTP response meets the expectations based on the original natural language request.`,

		UserPrompt: `Validate if this HTTP response satisfies the original request:

Original natural language request:
{{.NaturalLanguage}}

HTTP request that was generated and sent:
Method: {{.GeneratedMethod}}
URL: {{.GeneratedURL}}
Headers: {{.GeneratedHeaders}}
Body: {{.GeneratedBody}}

Response received:
{{.ExpectedBody}}

Provide your validation in JSON format:
{
  "is_valid": true|false,
  "satisfaction_score": 0.0 to 1.0,
  "reasoning": "Your detailed assessment",
  "missing_information": "Any information missing from the response"
}`,
	},
}

// RenderTemplate renders a prompt template with the given input
func RenderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// EvaluateWithAnthropicPrompt sends a prompt to Anthropic Claude to evaluate the request
func EvaluateWithAnthropicPrompt(ctx context.Context, model string, input RequestEvalInput) (float64, string, error) {
	// Use default model if not specified
	if model == "" {
		model = anthropic.ModelClaude3_7SonnetLatest
	}

	// Render the prompt templates
	promptTemplate := DefaultPromptTemplates.RequestEvaluation
	systemPrompt, err := RenderTemplate(promptTemplate.SystemPrompt, input)
	if err != nil {
		return 0, "", fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err := RenderTemplate(promptTemplate.UserPrompt, input)
	if err != nil {
		return 0, "", fmt.Errorf("failed to render user prompt: %w", err)
	}

	// Create Anthropic client
	client := anthropic.NewClient() // uses ANTHROPIC_API_KEY env var

	// Send the request to Anthropic
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})

	if err != nil {
		return 0, "", fmt.Errorf("Anthropic API error: %w", err)
	}

	// Extract the response content
	if len(msg.Content) == 0 {
		return 0, "", fmt.Errorf("empty response from Anthropic")
	}

	responseText := msg.Content[0].Text

	// Parse the JSON response
	type EvaluationResponse struct {
		MethodScore  float64 `json:"method_score"`
		URLScore     float64 `json:"url_score"`
		HeadersScore float64 `json:"headers_score"`
		BodyScore    float64 `json:"body_score"`
		OverallScore float64 `json:"overall_score"`
		Reasoning    string  `json:"reasoning"`
		Suggestions  string  `json:"suggestions"`
	}

	// Try to extract JSON from the response
	var jsonStr string
	if start := bytes.Index([]byte(responseText), []byte("{")); start != -1 {
		if end := bytes.LastIndex([]byte(responseText), []byte("}")); end != -1 && end > start {
			jsonStr = responseText[start : end+1]
		}
	}

	if jsonStr == "" {
		return 0, responseText, fmt.Errorf("could not extract JSON from response")
	}

	var evalResponse EvaluationResponse
	if err := json.Unmarshal([]byte(jsonStr), &evalResponse); err != nil {
		return 0, responseText, fmt.Errorf("failed to parse evaluation response: %w", err)
	}

	// Format the details including reasoning and suggestions
	details := fmt.Sprintf("Scores:\n"+
		"- Method: %.2f\n"+
		"- URL: %.2f\n"+
		"- Headers: %.2f\n"+
		"- Body: %.2f\n"+
		"- Overall: %.2f\n\n"+
		"Reasoning: %s\n\n"+
		"Suggestions: %s",
		evalResponse.MethodScore,
		evalResponse.URLScore,
		evalResponse.HeadersScore,
		evalResponse.BodyScore,
		evalResponse.OverallScore,
		evalResponse.Reasoning,
		evalResponse.Suggestions)

	return evalResponse.OverallScore, details, nil
}
