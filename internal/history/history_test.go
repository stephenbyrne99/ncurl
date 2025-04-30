package history

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryOperations(t *testing.T) {
	// Create a temporary directory for test history
	tempDir, err := os.MkdirTemp("", "ncurl-history-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test manager with a history file in the temp directory
	manager := &Manager{
		historyFile: filepath.Join(tempDir, "history.json"),
		maxEntries:  10,
	}

	// Test adding entries
	err = manager.AddEntry("test command 1", true)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	err = manager.AddEntry("test command 2", false)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Verify entries were saved
	entries, err := manager.GetEntries()
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Verify order (newest first)
	if entries[0].Command != "test command 2" {
		t.Errorf("Expected first entry to be newest command, got %s", entries[0].Command)
	}

	if entries[1].Command != "test command 1" {
		t.Errorf("Expected second entry to be oldest command, got %s", entries[1].Command)
	}

	// Verify success status
	if entries[0].Success {
		t.Errorf("Expected second command to have success=false")
	}

	if !entries[1].Success {
		t.Errorf("Expected first command to have success=true")
	}

	// Test max entries limit
	for i := 0; i < 15; i++ {
		err = manager.AddEntry(f("command %d", i), true)
		if err != nil {
			t.Fatalf("Failed to add entry %d: %v", i, err)
		}
	}

	entries, err = manager.GetEntries()
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 10 {
		t.Fatalf("Expected entries to be trimmed to max 10, got %d", len(entries))
	}
}

func TestGetEntryByIndex(t *testing.T) {
	// Create a temporary directory for test history
	tempDir, err := os.MkdirTemp("", "ncurl-history-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test manager with a history file in the temp directory
	manager := &Manager{
		historyFile: filepath.Join(tempDir, "history.json"),
		maxEntries:  10,
	}

	// Add some test entries
	for i := 1; i <= 5; i++ {
		err = manager.AddEntry(f("command %d", i), true)
		if err != nil {
			t.Fatalf("Failed to add entry %d: %v", i, err)
		}
	}

	// Test getting a valid entry
	entry, err := manager.GetEntryByIndex(3)
	if err != nil {
		t.Fatalf("Failed to get entry by index: %v", err)
	}

	// Index 3 should be "command 3" (entries are stored newest first)
	expected := "command 3"
	if entry.Command != expected {
		t.Errorf("Expected command '%s', got '%s'", expected, entry.Command)
	}

	// Test getting an invalid entry
	_, err = manager.GetEntryByIndex(20)
	if err == nil {
		t.Errorf("Expected error when retrieving out-of-range index")
	}
	if !errors.Is(err, ErrEntryNotFound) {
		t.Errorf("Expected ErrEntryNotFound, got %v", err)
	}
}

func TestSearchHistory(t *testing.T) {
	// Create a temporary directory for test history
	tempDir, err := os.MkdirTemp("", "ncurl-history-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test manager with a history file in the temp directory
	manager := &Manager{
		historyFile: filepath.Join(tempDir, "history.json"),
		maxEntries:  10,
	}

	// Add some test entries
	if err := manager.AddEntry("GET github API", true); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := manager.AddEntry("POST user data", false); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := manager.AddEntry("GET weather data", true); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}
	if err := manager.AddEntry("DELETE user", true); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Test searching for "GET"
	results, err := manager.SearchHistory("GET")
	if err != nil {
		t.Fatalf("Failed to search history: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'GET', got %d", len(results))
	}

	// Test case-insensitive search
	results, err = manager.SearchHistory("get")
	if err != nil {
		t.Fatalf("Failed to search history: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for case-insensitive 'get', got %d", len(results))
	}

	// Test search for "data"
	results, err = manager.SearchHistory("data")
	if err != nil {
		t.Fatalf("Failed to search history: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'data', got %d", len(results))
	}

	// Test search with no matches
	results, err = manager.SearchHistory("nonexistent")
	if err != nil {
		t.Fatalf("Failed to search history: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

// TestPromptForHistorySelectionStructure just verifies the function signature - not actual functionality
// since we can't easily test interactive prompts in unit tests
func TestPromptForHistorySelectionStructure(t *testing.T) {
	// Declare a local variable to hold the type signature we expect
	var promptFunc func() (string, error)
	
	// Get pointer to a valid manager
	manager := &Manager{
		historyFile: "test-file.json",
		maxEntries:  10,
	}
	
	// Type assertion to verify function has expected signature
	promptFunc = manager.PromptForHistorySelection
	
	// Simple call to ensure the function is properly assigned
	if promptFunc == nil {
		t.Fatalf("PromptForHistorySelection should not be nil")
	}
}

// Fix for test compilation
func f(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
