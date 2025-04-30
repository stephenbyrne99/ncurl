// Package history provides functionality for managing command history
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryEntry represents a single command in the history
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Success   bool      `json:"success"`
}

// Manager handles the saving and loading of command history
type Manager struct {
	historyFile string
	maxEntries  int
}

// NewManager creates a new history manager
func NewManager(maxEntries int) (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Create .ncurl directory if it doesn't exist
	configDir := filepath.Join(home, ".ncurl")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Manager{
		historyFile: filepath.Join(configDir, "history.json"),
		maxEntries:  maxEntries,
	}, nil
}

// AddEntry adds a new entry to the history
func (m *Manager) AddEntry(command string, success bool) error {
	entries, err := m.GetEntries()
	if err != nil {
		// If we can't read the history, just start with an empty slice
		entries = []HistoryEntry{}
	}

	// Add new entry
	entry := HistoryEntry{
		Timestamp: time.Now(),
		Command:   command,
		Success:   success,
	}

	// Prepend the new entry (most recent first)
	entries = append([]HistoryEntry{entry}, entries...)

	// Trim to max entries
	if len(entries) > m.maxEntries {
		entries = entries[:m.maxEntries]
	}

	// Write back to file
	return m.saveEntries(entries)
}

// GetEntries retrieves all history entries
func (m *Manager) GetEntries() ([]HistoryEntry, error) {
	// Check if history file exists
	if _, err := os.Stat(m.historyFile); os.IsNotExist(err) {
		return []HistoryEntry{}, nil
	}

	file, err := os.Open(m.historyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	var entries []HistoryEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode history data: %w", err)
	}

	return entries, nil
}

// saveEntries writes entries to the history file
func (m *Manager) saveEntries(entries []HistoryEntry) error {
	file, err := os.Create(m.historyFile)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		return fmt.Errorf("failed to encode history data: %w", err)
	}

	return nil
}

// PrintHistory prints the command history to stdout
func (m *Manager) PrintHistory() error {
	entries, err := m.GetEntries()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No command history found")
		return nil
	}

	fmt.Println("Command History:")
	fmt.Println("---------------")
	for i, entry := range entries {
		status := "✓"
		if !entry.Success {
			status = "✗"
		}
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, status, entry.Command, entry.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// ErrEntryNotFound is returned when a requested history entry doesn't exist
var ErrEntryNotFound = errors.New("history entry not found")

// GetEntryByIndex retrieves a specific history entry by its index (1-based)
func (m *Manager) GetEntryByIndex(index int) (HistoryEntry, error) {
	entries, err := m.GetEntries()
	if err != nil {
		return HistoryEntry{}, err
	}

	// Convert to 0-based index for slice access
	idx := index - 1
	if idx < 0 || idx >= len(entries) {
		return HistoryEntry{}, fmt.Errorf("%w: index %d is out of range", ErrEntryNotFound, index)
	}

	return entries[idx], nil
}

// SearchHistory returns entries that contain the given search term
func (m *Manager) SearchHistory(term string) ([]HistoryEntry, error) {
	entries, err := m.GetEntries()
	if err != nil {
		return nil, err
	}

	if term == "" {
		return entries, nil
	}

	var results []HistoryEntry
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Command), strings.ToLower(term)) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// PrintSearchResults prints history entries that match the search term
func (m *Manager) PrintSearchResults(term string) error {
	results, err := m.SearchHistory(term)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Printf("No commands found matching '%s'\n", term)
		return nil
	}

	fmt.Printf("Commands matching '%s':\n", term)
	fmt.Println("---------------")
	for i, entry := range results {
		status := "✓"
		if !entry.Success {
			status = "✗"
		}
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, status, entry.Command, entry.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// PromptForHistorySelection displays an interactive history menu and returns the selected command
func (m *Manager) PromptForHistorySelection() (string, error) {
	entries, err := m.GetEntries()
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return "", fmt.Errorf("no command history available")
	}

	fmt.Println("Command History:")
	fmt.Println("---------------")
	for i, entry := range entries {
		status := "✓"
		if !entry.Success {
			status = "✗"
		}
		fmt.Printf("%d. [%s] %s (%s)\n", i+1, status, entry.Command, entry.Timestamp.Format("2006-01-02 15:04:05"))
	}

	var selectedIndex int
	fmt.Print("\nEnter number to select command (or 0 to cancel): ")
	_, err = fmt.Scanln(&selectedIndex)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	if selectedIndex == 0 {
		return "", fmt.Errorf("selection cancelled")
	}

	if selectedIndex < 1 || selectedIndex > len(entries) {
		return "", fmt.Errorf("invalid selection: %d (valid range: 1-%d)", selectedIndex, len(entries))
	}

	return entries[selectedIndex-1].Command, nil
}