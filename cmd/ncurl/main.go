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
	infoLogger  = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.LstdFlags)
)

var (
	// Command line flags
	timeout         = flag.Int("t", 30, "Timeout in seconds for the HTTP request")
	model           = flag.String("m", anthropic.ModelClaude3_7SonnetLatest, "Anthropic model to use")
	jsonOnly        = flag.Bool("j", false, "Output response body as JSON only")
	verbose         = flag.Bool("v", false, "Verbose output (include request details)")
	showVersion     = flag.Bool("version", false, "Show version information")
	debug           = flag.Bool("debug", false, "Enable debug logging")
	showHistory     = flag.Bool("history", false, "Show command history")
	historyCount    = flag.Int("history-count", 50, "Maximum number of history entries to keep")
	historyRerun    = flag.Int("rerun", 0, "Rerun a command from history by index")
	historySearch   = flag.String("search", "", "Search command history for a term")
)

func main() {
	// Set up exit code handling
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	flag.Parse()

	// Enable/disable debug logging based on flag
	if !*debug {
		debugLogger.SetOutput(nil)
	}

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

	// Show or search command history if requested
	if historyManager != nil {
		if *showHistory {
			if err := historyManager.PrintHistory(); err != nil {
				errorLogger.Printf("Failed to print history: %v\n", err)
				exitCode = 1
			}
			return
		}

		if *historySearch != "" {
			if err := historyManager.PrintSearchResults(*historySearch); err != nil {
				errorLogger.Printf("Failed to search history: %v\n", err)
				exitCode = 1
			}
			return
		}
	}

	// Get the command to execute - either from history or command line args
	var prompt string
	if *historyRerun > 0 && historyManager != nil {
		// Get command from history by index
		entry, err := historyManager.GetEntryByIndex(*historyRerun)
		if err != nil {
			errorLogger.Printf("Failed to retrieve history entry: %v\n", err)
			exitCode = 1
			return
		}
		prompt = entry.Command
		debugLogger.Printf("Rerunning command from history: %s", prompt)
	} else {
		// Get command from command line args
		args := flag.Args()
		if len(args) < 1 {
			fmt.Println("usage: ncurl [options] \"<natural language request>\"")
			fmt.Println("\nOptions:")
			flag.PrintDefaults()
			exitCode = 1
			return
		}
		prompt = strings.Join(args, " ")
	}

	// Ensure API key is set
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		errorLogger.Println("ANTHROPIC_API_KEY environment variable is required")
		exitCode = 1
		return
	}

	debugLogger.Printf("Processing natural language request: %s", prompt)
	
	// Record command in history when exiting
	defer func() {
		if historyManager != nil {
			if err := historyManager.AddEntry(prompt, exitCode == 0); err != nil {
				debugLogger.Printf("Failed to save command to history: %v", err)
			}
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

	debugLogger.Printf("Generated request spec: %+v", spec)

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
		if errors.As(err, &reqErr) {
			errorLogger.Printf("Request error: %v\n", reqErr)
		} else if errors.Is(err, httpx.ErrInvalidRequest) {
			errorLogger.Printf("Invalid request: %v\n", err)
		} else if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			errorLogger.Printf("Request timed out after %d seconds\n", *timeout)
		} else {
			errorLogger.Printf("Request failed: %v\n", err)
		}
		exitCode = 1
		return
	}

	debugLogger.Printf("Response received: Status=%s, ContentType=%s, BodyLength=%d",
		response.Status, response.Header.Get("Content-Type"), len(response.Body))

	// Output the response
	if *jsonOnly {
		fmt.Println(string(response.Body))
	} else {
		fmt.Printf("Status: %s\n", response.Status)
		fmt.Printf("Content-Type: %s\n", response.Header.Get("Content-Type"))
		if *verbose {
			fmt.Println("Headers:")
			for k, v := range response.Header {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		fmt.Println()
		fmt.Println(string(response.Body))
	}
}