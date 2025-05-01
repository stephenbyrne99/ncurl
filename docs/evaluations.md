# ncurl Evaluations

This document describes how to use the evaluation framework to test and validate the ncurl tool's ability to correctly translate natural language into HTTP requests.

## Overview

The evaluation framework is designed to:

1. Test ncurl's ability to interpret natural language requests
2. Validate the generated HTTP requests against expected results
3. Provide metrics and feedback on performance
4. Support extending the test cases for specific use cases

## Getting Started

### Prerequisites

- Go 1.22 or later
- An Anthropic API key (set as the `ANTHROPIC_API_KEY` environment variable)

### Building the Evaluation Tool

```bash
# From the project root
go build -o ncurl-eval ./cmd/ncurl-eval
```

### Running Basic Evaluations

```bash
# Run all built-in test cases
./ncurl-eval

# Run with verbose output
./ncurl-eval -v

# Save results to a file
./ncurl-eval -output results.md
```

### Command Line Options

The evaluation tool supports the following command line options:

| Option | Description |
|--------|-------------|
| `-tests` | Path to a JSON file with custom test cases |
| `-output` | Path to save evaluation results (default: print to stdout) |
| `-model` | Anthropic model to use (default: claude-3-7-sonnet-latest) |
| `-timeout` | Timeout in seconds for each test case (default: 30) |
| `-v` | Enable verbose output |
| `-id` | Run only test cases with this ID |
| `-count` | Maximum number of test cases to run (0 = all) |
| `-json` | Output results in JSON format |
| `-gen-tests` | Generate a template test cases file |
| `-gen-tests-output` | Path to save generated test cases (default: testcases.json) |

## Test Cases

### Built-in Test Cases

The evaluation framework includes a set of built-in test cases covering common HTTP request scenarios:

- Basic HTTP methods (GET, POST, PUT, DELETE)
- Common API requests (weather, GitHub, JSONPlaceholder)
- Requests with authentication and custom headers
- Complex JSON bodies
- Natural language interpretation
- Error cases

### Creating Custom Test Cases

You can create your own test cases by:

1. Generating a template file:
   ```bash
   ./ncurl-eval -gen-tests -gen-tests-output my-tests.json
   ```

2. Editing the generated JSON file to add or modify test cases

3. Running evaluations with your custom test cases:
   ```bash
   ./ncurl-eval -tests my-tests.json
   ```

### Test Case Format

Test cases are defined in JSON with the following structure:

```json
{
  "id": "unique-test-id",
  "description": "Human-readable description",
  "input": "Natural language request to evaluate",
  "expected_method": "GET|POST|PUT|DELETE|etc.",
  "expected_url": "URL or URL fragment to match",
  "expected_url_regex": "Regular expression to match URL (optional)",
  "expected_headers": {
    "Header-Name": "Expected Value"
  },
  "expected_body": "Expected body string or fragment",
  "mock_response": {
    "status_code": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "{\"example\": \"response\"}"
  },
  "prompt_template": "Custom prompt template (optional)"
}
```

## Evaluation Criteria

The evaluation framework scores requests based on the following criteria:

1. **Method Match**: Does the generated HTTP method match the expected method?
2. **URL Match**: Does the generated URL contain the expected URL components?
3. **Headers Match**: Do the generated headers include the expected headers?
4. **Body Match**: Does the generated body contain the expected content?

Each criterion is scored on a scale from 0.0 to 1.0, and an overall score is calculated. A request is considered successful if the overall score is 0.8 or higher.

## Advanced Usage

### Using Anthropic Prompts for Evaluation

The evaluation framework uses Anthropic Claude models to evaluate the quality of generated requests beyond basic matching. This allows for more nuanced assessment of whether the generated request correctly captures the intent of the natural language input.

You can customize these prompts by modifying the templates in `internal/evals/prompts.go`.

### Validating Input and Output

The framework includes utilities for validating both:

1. **Input Quality**: Assessing the clarity, completeness, and specificity of natural language inputs
2. **Response Validation**: Checking whether HTTP responses satisfy the original request intent

These can be used independently for validating requests:

```go
import "github.com/stephenbyrne99/ncurl/internal/evals"

validator := evals.NewResponseValidator("")
result, err := validator.ValidateInput(ctx, "get weather for New York")
if err != nil {
    // Handle error
}
fmt.Printf("Input clarity: %.2f\n", result.Clarity)
```

### Extending the Framework

To extend the evaluation framework:

1. **Add new test cases**: Create custom test files or modify the built-in cases
2. **Customize evaluation prompts**: Edit the templates in `prompts.go`
3. **Add validation logic**: Modify or extend the validation functions in `validator.go`
4. **Create specialized evaluations**: Build on the `Evaluator` interface for specific testing needs

## Troubleshooting

### Common Issues

- **API Key Problems**: Ensure the `ANTHROPIC_API_KEY` environment variable is set correctly
- **Timeout Errors**: Increase the timeout value with the `-timeout` flag
- **Model Errors**: Try a different model with the `-model` flag

### Debugging

Use the `-v` flag to enable verbose output with detailed information about each evaluation step.

## Conclusion

The evaluation framework provides a systematic way to test and improve ncurl's natural language interpretation capabilities. Regular evaluations can help identify areas for improvement and ensure that the tool continues to generate accurate HTTP requests from natural language inputs.