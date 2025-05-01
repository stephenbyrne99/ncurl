package evals

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// TestEvaluatorBasics tests the basic functionality of the evaluator
func TestEvaluatorBasics(t *testing.T) {
	// Create evaluator with default settings
	evaluator := NewEvaluator("", 0)
	
	if evaluator.model != anthropic.ModelClaude3_7SonnetLatest {
		t.Errorf("Expected default model to be %s, got %s", anthropic.ModelClaude3_7SonnetLatest, evaluator.model)
	}
	
	if evaluator.timeout != 30*time.Second {
		t.Errorf("Expected default timeout to be 30s, got %v", evaluator.timeout)
	}
	
	// Create evaluator with custom settings
	customModel := "claude-3-haiku-20240307"
	customTimeout := 10 * time.Second
	customEval := NewEvaluator(customModel, customTimeout)
	
	if customEval.model != customModel {
		t.Errorf("Expected model to be %s, got %s", customModel, customEval.model)
	}
	
	if customEval.timeout != customTimeout {
		t.Errorf("Expected timeout to be %v, got %v", customTimeout, customEval.timeout)
	}
}

// TestLoadTestCases tests loading test cases
func TestLoadTestCases(t *testing.T) {
	evaluator := NewEvaluator("", 0)
	
	// Test with empty cases
	err := evaluator.LoadTestCases([]EvalCase{})
	if err == nil {
		t.Error("Expected error when loading empty test cases, got nil")
	}
	
	// Test with valid cases
	cases := []EvalCase{
		{
			ID:             "test1",
			Description:    "Simple GET request",
			Input:          "get the weather in London",
			ExpectedMethod: "GET",
			ExpectedURL:    "weather",
		},
	}
	
	err = evaluator.LoadTestCases(cases)
	if err != nil {
		t.Errorf("Unexpected error when loading valid test cases: %v", err)
	}
	
	if len(evaluator.cases) != len(cases) {
		t.Errorf("Expected %d test cases, got %d", len(cases), len(evaluator.cases))
	}
}

// TestEvaluation tests the evaluation process
// This is an integration test that requires an API key
func TestEvaluation(t *testing.T) {
	// Skip test if no API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("Skipping test; ANTHROPIC_API_KEY not set")
	}
	
	evaluator := NewEvaluator("", 0)
	
	// Test with a simple case
	testCase := EvalCase{
		ID:             "weather-test",
		Description:    "Weather API test",
		Input:          "get the current weather for London",
		ExpectedMethod: "GET",
		ExpectedURL:    "weather",
	}
	
	ctx := context.Background()
	result, err := evaluator.Run(ctx, testCase)
	
	if err != nil {
		t.Fatalf("Unexpected error running evaluation: %v", err)
	}
	
	if result.TestID != testCase.ID {
		t.Errorf("Expected TestID %s, got %s", testCase.ID, result.TestID)
	}
	
	if result.Description != testCase.Description {
		t.Errorf("Expected Description %s, got %s", testCase.Description, result.Description)
	}
	
	// We can't assert the exact result as it depends on the LLM,
	// but we can check that we got some result
	if result.ActualURL == "" {
		t.Error("Expected non-empty ActualURL")
	}
}

// TestGenerateReport tests the report generation
func TestGenerateReport(t *testing.T) {
	results := []EvalResult{
		{
			TestID:      "test1",
			Description: "Test 1",
			Success:     true,
			Score:       1.0,
			Input:       "get weather",
			ExpectedURL: "weather",
			ActualURL:   "https://api.weather.com/v1/current",
		},
		{
			TestID:      "test2",
			Description: "Test 2",
			Success:     false,
			Score:       0.5,
			Input:       "post user data",
			ExpectedURL: "users",
			ActualURL:   "https://api.example.com/data",
			Error:       "URL mismatch",
		},
	}
	
	report := GenerateReport(results)
	
	// Basic checks on the report
	if report == "" {
		t.Error("Expected non-empty report")
	}
	
	if len(report) < 100 {
		t.Errorf("Report seems too short: %s", report)
	}
}