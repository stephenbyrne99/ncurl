// Command ncurl translates natural language into HTTP requests
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stephenbyrne99/ncurl/internal/history"
	"github.com/stephenbyrne99/ncurl/internal/httpx"
	"github.com/stephenbyrne99/ncurl/internal/llm"
)

// Version information set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Logger instances for different levels
var (
	errorLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
)

var (
	// Command line flags
	timeout = flag.Int("t", 30, "Timeout in seconds for the HTTP request")
	model   = flag.String(
		"m",
		anthropic.ModelClaude3_7SonnetLatest,
		"Anthropic model to use",
	)
	jsonOnly           = flag.Bool("j", false, "Output response body as JSON only")
	verbose            = flag.Bool("v", false, "Verbose output (include request details)")
	showVersion        = flag.Bool("version", false, "Show version information")
	showHelp           = flag.Bool("help", false, "Show detailed help with usage examples")
	showHistory        = flag.Bool("history", false, "Show command history")
	historyCount       = flag.Int("history-count", 50, "Maximum number of history entries to keep")
	historyRerun       = flag.Int("rerun", 0, "Rerun a command from history by index")
	historySearch      = flag.String("search", "", "Search command history for a term")
	interactiveHistory = flag.Bool("i", false, "Interactive history selection mode")
)

// printHelp displays detailed usage information and examples
func printHelp() {
	helpText := `ncurl - curl in English

DESCRIPTION
  ncurl allows you to describe HTTP requests in plain English, and it will 
  translate your natural language into proper HTTP requests.

USAGE
  ncurl [options] "<natural language request>"
  ncurl help        Show this help message

OPTIONS
  -t <seconds>       Set timeout in seconds (default: 30)
  -m <model>         Specify Anthropic model to use (default: claude-3-7-sonnet)
  -j                 Output response body as JSON only
  -v                 Verbose output (include request details)
  -version           Show version information
  -help              Show this detailed help message

HISTORY OPTIONS
  -history           Show command history
  -history-count <n> Maximum number of history entries to keep (default: 50)
  -rerun <n>         Rerun a command from history by index
  -search <term>     Search command history for a term
  -i                 Interactive history selection mode

EXAMPLES
  # Simple GET request
  ncurl "get the latest weather for London"

  # POST with JSON data
  ncurl "post a new user with name 'John' and email 'john@example.com' to jsonplaceholder"

  # Specify headers and authentication
  ncurl "get my GitHub repos with authorization token ghp_abc123"

  # Use -j flag for JSON-only output (useful for piping to jq)
  ncurl -j "get COVID data for New York" | jq '.cases'

  # View and rerun command history
  ncurl -history
  ncurl -rerun 3

ENVIRONMENT
  ANTHROPIC_API_KEY  Required API key for the Anthropic Claude API, add this to your .zshrc or .bashrc

For more information on a specific command, run 'ncurl <command> -help'
`
	fmt.Println(helpText)
}

// isContentBinary determines if content should be treated as binary based on content type
func isContentBinary(contentType string) bool {
	return !strings.Contains(contentType, "text/") &&
		!strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "application/xml") &&
		!strings.Contains(contentType, "application/javascript")
}

// outputJSONOnlyMode outputs the response body in JSON-only mode
func outputJSONOnlyMode(body []byte, isBinary bool) {
	if isBinary {
		_, _ = os.Stdout.Write(body)
	} else {
		// For text-based content, use print which does string conversion
		fmt.Print(string(body))

		// Add a newline if not already present
		if len(body) > 0 && body[len(body)-1] != '\n' {
			fmt.Println()
		}
	}
}

// getPromptString gets the prompt string from history or command line args
func getPromptString(
	historyManager *history.Manager,
	interactiveHistory bool,
	historyRerun int,
	exitCode *int,
	logger *log.Logger,
) (string, bool) {
	switch {
	case interactiveHistory && historyManager != nil:
		// Get command from interactive history selection
		cmd, promptErr := historyManager.PromptForHistorySelection()
		if promptErr != nil {
			logger.Printf("Failed to select from history: %v\n", promptErr)
			*exitCode = 1
			return "", true
		}
		return cmd, false

	case historyRerun > 0 && historyManager != nil:
		// Get command from history by index
		entry, historyErr := historyManager.GetEntryByIndex(historyRerun)
		if historyErr != nil {
			logger.Printf("Failed to retrieve history entry: %v\n", historyErr)
			*exitCode = 1
			return "", true
		}
		return entry.Command, false

	default:
		// Get command from command line args
		args := flag.Args()
		if len(args) < 1 {
			fmt.Println("usage: ncurl [options] \"<natural language request>\"")
			fmt.Println("\nExamples:")
			fmt.Println("  ncurl \"get the weather for New York\"")
			fmt.Println("  ncurl \"post a new user to jsonplaceholder\"")
			fmt.Println("\nUse -help for detailed usage information and more examples")
			fmt.Println("\nOptions:")
			flag.PrintDefaults()
			*exitCode = 1
			return "", true
		}
		return strings.Join(args, " "), false
	}
}

// outputStandardMode outputs the response in standard mode with metadata
func outputStandardMode(response *httpx.Response, verbose bool, isBinary bool) {
	// Print metadata and headers
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Content-Type: %s\n", response.Header.Get("Content-Type"))

	if verbose {
		fmt.Println("Headers:")
		for k, v := range response.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	fmt.Println()

	// Print the response body
	if isBinary {
		// For binary data, write raw bytes
		_, _ = os.Stdout.Write(response.Body)
		fmt.Println() // Add a newline after binary data
	} else {
		// For text data, convert to string
		fmt.Println(string(response.Body))
	}
}

// handleHistoryOperations handles showing and searching command history
// Returns true if a history operation was executed (indicating the caller should return)
func handleHistoryOperations(
	historyManager *history.Manager,
	logger *log.Logger,
	showHistory bool,
	searchTerm string,
	exitCode *int,
) bool {
	if showHistory {
		if err := historyManager.PrintHistory(); err != nil {
			logger.Printf("Failed to print history: %v\n", err)
			*exitCode = 1
		}
		return true
	}

	if searchTerm != "" {
		if err := historyManager.PrintSearchResults(searchTerm); err != nil {
			logger.Printf("Failed to search history: %v\n", err)
			*exitCode = 1
		}
		return true
	}

	return false
}

func main() {
	// Set up exit code handling more cleanly
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// Check for help subcommand
	if len(os.Args) > 1 && os.Args[1] == "help" {
		printHelp()
		return
	}

	flag.Parse()

	// Initialize history manager
	historyManager, err := history.NewManager(*historyCount)
	if err != nil {
		errorLogger.Printf("Warning: Could not initialize history: %v\n", err)
	}

	// Show version information if requested
	if *showVersion {
		fmt.Printf("ncurl version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		return
	}

	// Show detailed help if requested
	if *showHelp {
		printHelp()
		return
	}

	// Handle history operations
	if historyManager != nil &&
		handleHistoryOperations(
			historyManager,
			errorLogger,
			*showHistory,
			*historySearch,
			&exitCode,
		) {
		return
	}

	// Get the command to execute - either from history, interactive selection, or command line args
	prompt, shouldReturn := getPromptString(
		historyManager,
		*interactiveHistory,
		*historyRerun,
		&exitCode,
		errorLogger,
	)
	if shouldReturn {
		return
	}

	// Ensure API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		errorLogger.Println("ANTHROPIC_API_KEY environment variable is required")
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable is required")
		fmt.Println("Please set it with: export ANTHROPIC_API_KEY=\"your-key-here\"")
		fmt.Println(
			"Or for a single command: ANTHROPIC_API_KEY=\"your-key-here\" ncurl \"your query\"",
		)
		exitCode = 1
		return
	}

	// Record command in history when exiting
	defer func() {
		if historyManager != nil {
			_ = historyManager.AddEntry(prompt, exitCode == 0)
		}
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	// Generate request spec from natural language
	client := llm.NewClient(*model)
	spec, err := client.GenerateRequestSpec(ctx, prompt)
	if err != nil {
		errorLogger.Printf("Failed to generate request: %v\n", err)
		exitCode = 1
		return
	}

	if *verbose {
		fmt.Printf("Request: %s %s\n", spec.Method, spec.URL)
		if len(spec.Headers) > 0 {
			fmt.Println("Headers:")
			for k, v := range spec.Headers {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		if spec.Body != "" {
			fmt.Println("Body:", spec.Body)
		}
		fmt.Println()
	}

	// Execute the request with context for cancellation/timeout
	response, err := httpx.ExecuteWithContext(ctx, spec)

	if err != nil {
		var reqErr *httpx.RequestError

		switch {
		case errors.As(err, &reqErr):
			errorLogger.Printf("Request error: %v\n", reqErr)
		case errors.Is(err, httpx.ErrInvalidRequest):
			errorLogger.Printf("Invalid request: %v\n", err)
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			errorLogger.Printf("Request timed out after %d seconds\n", *timeout)
		default:
			errorLogger.Printf("Request failed: %v\n", err)
		}

		exitCode = 1
		return
	}

	// Determine content type and handle output appropriately
	contentType := response.Header.Get("Content-Type")
	isBinary := isContentBinary(contentType)

	if *jsonOnly {
		outputJSONOnlyMode(response.Body, isBinary)
	} else {
		outputStandardMode(response, *verbose, isBinary)
	}
}
