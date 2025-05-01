// Command ncurl-eval runs evaluations to test ncurl functionality
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stephenbyrne99/ncurl/internal/evals"
)

func main() {
	// Define command line flags
	testCasesFile := flag.String("tests", "", "Path to JSON file with test cases (default: use built-in test cases)")
	outputFile := flag.String("output", "", "Path to save evaluation results (default: print to stdout)")
	modelFlag := flag.String("model", anthropic.ModelClaude3_7SonnetLatest, "Anthropic model to use")
	timeoutFlag := flag.Int("timeout", 30, "Timeout in seconds for each test case")
	verboseFlag := flag.Bool("v", false, "Verbose output")
	runIDFlag := flag.String("id", "", "Run only test cases with this ID")
	countFlag := flag.Int("count", 0, "Maximum number of test cases to run (0 = all)")
	jsonFlag := flag.Bool("json", false, "Output results in JSON format")
	genTestsFlag := flag.Bool("gen-tests", false, "Generate a template test cases file")
	genTestsOutputFlag := flag.String("gen-tests-output", "testcases.json", "Path to save generated test cases")
	
	// Custom usage message
	flag.Usage = func() {
		fmt.Println("ncurl-eval - Evaluation tool for ncurl")
		fmt.Println("\nRunning evaluations to test ncurl's ability to correctly")
		fmt.Println("translate natural language to HTTP requests.")
		fmt.Println("\nUsage: ncurl-eval [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  ncurl-eval                              # Run all built-in test cases")
		fmt.Println("  ncurl-eval -tests my-tests.json         # Run tests from a file")
		fmt.Println("  ncurl-eval -id weather-test             # Run a specific test case by ID")
		fmt.Println("  ncurl-eval -output results.md           # Save results to file")
		fmt.Println("  ncurl-eval -json -output results.json   # Save results as JSON")
		fmt.Println("  ncurl-eval -gen-tests                   # Generate template test cases file")
	}

	flag.Parse()

	// Set up exit code handling
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// Check if API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintf(os.Stderr, "ANTHROPIC_API_KEY environment variable is required\n")
		exitCode = 1
		return
	}

	// Handle generate test cases flag
	if *genTestsFlag {
		if err := evals.CreateDefaultTestCasesFile(*genTestsOutputFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating test cases file: %v\n", err)
			exitCode = 1
			return
		}
		fmt.Printf("Generated test cases file at %s\n", *genTestsOutputFlag)
		return
	}

	// Create an evaluator
	evaluator := evals.NewEvaluator(*modelFlag, time.Duration(*timeoutFlag)*time.Second)

	// Load test cases
	var testCases []evals.EvalCase
	var err error

	if *testCasesFile != "" {
		// Load test cases from file
		testCases, err = evals.LoadTestCasesFromFile(*testCasesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading test cases: %v\n", err)
			exitCode = 1
			return
		}
	} else {
		// Use default test cases
		testCases = evals.DefaultTestCases()
	}

	// Filter test cases if ID is specified
	if *runIDFlag != "" {
		var filtered []evals.EvalCase
		for _, tc := range testCases {
			if tc.ID == *runIDFlag {
				filtered = append(filtered, tc)
			}
		}
		
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "No test case found with ID: %s\n", *runIDFlag)
			exitCode = 1
			return
		}
		
		testCases = filtered
	}

	// Limit the number of test cases if count is specified
	if *countFlag > 0 && *countFlag < len(testCases) {
		testCases = testCases[:*countFlag]
	}

	// Load test cases into evaluator
	if err := evaluator.LoadTestCases(testCases); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading test cases: %v\n", err)
		exitCode = 1
		return
	}

	// Run the evaluations
	ctx := context.Background()
	results, err := evaluator.RunAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running evaluations: %v\n", err)
		exitCode = 1
		return
	}

	// Generate output
	var output string
	if *jsonFlag {
		// Output as JSON
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
			exitCode = 1
			return
		}
		output = string(jsonData)
	} else {
		// Output as markdown report
		output = evals.GenerateReport(results)
	}

	// Write output to file or stdout
	if *outputFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(*outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			exitCode = 1
			return
		}
		
		if err := os.WriteFile(*outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			exitCode = 1
			return
		}
		
		fmt.Printf("Evaluation results saved to %s\n", *outputFile)
	} else {
		// Print to stdout
		fmt.Println(output)
	}

	// Print summary to stderr if verbose
	if *verboseFlag {
		totalTests := len(results)
		successCount := 0
		var totalScore float64
		
		for _, r := range results {
			if r.Success {
				successCount++
			}
			totalScore += r.Score
		}
		
		avgScore := totalScore / float64(totalTests)
		successRate := float64(successCount) / float64(totalTests) * 100
		
		fmt.Fprintf(os.Stderr, "Evaluation Summary:\n")
		fmt.Fprintf(os.Stderr, "- Total Tests: %d\n", totalTests)
		fmt.Fprintf(os.Stderr, "- Successful Tests: %d (%.1f%%)\n", successCount, successRate)
		fmt.Fprintf(os.Stderr, "- Average Score: %.2f\n", avgScore)
	}
}