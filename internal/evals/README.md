# ncurl Evaluation Framework

This package provides a framework for evaluating ncurl's ability to translate natural language into HTTP requests.

## Components

- **`evals.go`**: Core evaluation structures and logic
- **`prompts.go`**: Anthropic prompt templates for evaluation
- **`testcases.go`**: Default test cases and utilities for loading/saving test cases
- **`validator.go`**: Validation utilities for requests, responses, and inputs

## Usage Examples

### Basic Evaluation

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/stephenbyrne99/ncurl/internal/evals"
)

func main() {
    // Create an evaluator
    evaluator := evals.NewEvaluator("", 30*time.Second)
    
    // Load default test cases
    testCases := evals.DefaultTestCases()
    evaluator.LoadTestCases(testCases)
    
    // Run evaluations
    ctx := context.Background()
    results, err := evaluator.RunAll(ctx)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    // Generate and print report
    report := evals.GenerateReport(results)
    fmt.Println(report)
}
```

### Custom Test Cases

```go
// Create a custom test case
testCase := evals.EvalCase{
    ID:             "custom-test",
    Description:    "Custom API test",
    Input:          "get user profile with id 123 from my-api.example.com",
    ExpectedMethod: "GET",
    ExpectedURL:    "my-api.example.com",
    ExpectedHeaders: map[string]string{
        "Accept": "application/json",
    },
}

// Run single evaluation
result, err := evaluator.Run(ctx, testCase)
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Score: %.2f, Success: %v\n", result.Score, result.Success)
```

### Input Validation

```go
validator := evals.NewResponseValidator("")
inputResult, err := validator.ValidateInput(ctx, "get weather")
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Input clarity: %.2f\n", inputResult.Clarity)
fmt.Printf("Input specificity: %.2f\n", inputResult.Specificity)
fmt.Printf("Analysis: %s\n", inputResult.Analysis)
```

## Extending the Framework

### Adding New Test Cases

Create a new test case file:

```go
cases := []evals.EvalCase{
    {
        ID:             "my-api-test",
        Description:    "Test my custom API",
        Input:          "get data from my-api",
        ExpectedMethod: "GET",
        ExpectedURL:    "my-api",
    },
    // Add more test cases...
}

// Save to file
evals.SaveTestCasesToFile(cases, "my-tests.json")
```

### Customizing Evaluation Prompts

Modify the prompt templates in `prompts.go` to adjust evaluation criteria:

```go
// Create a custom prompt template
customTemplate := evals.PromptTemplate{
    SystemPrompt: "Your custom system prompt...",
    UserPrompt:   "Your custom user prompt template...",
}

// Use in evaluation
// (would require modifying the internal code to accept custom templates)
```

## Building the Evaluation Tool

The evaluation framework includes a command-line tool for running evaluations:

```bash
# From the project root
go build -o ncurl-eval ./cmd/ncurl-eval
```

See the main documentation in `/docs/evaluations.md` for more details on using the evaluation tool.