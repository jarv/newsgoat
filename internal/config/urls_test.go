package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCommentPreservation(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()
	urlsPath := filepath.Join(testDir, "urls")

	// Create initial urls file with comments
	initialContent := `# This is a comment
# Another comment with some context
https://example.com/feed1.xml

# Section for tech feeds
https://example.com/feed2.xml Tech,Programming
# A feed with a note
https://example.com/feed3.xml News

# End of feeds
`

	err := os.WriteFile(urlsPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Read all lines
	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read lines: %v", err)
	}

	// Verify we read the correct number of lines
	expectedLineCount := 10 // including blank lines and comments
	if len(lines) != expectedLineCount {
		t.Errorf("Expected %d lines, got %d", expectedLineCount, len(lines))
	}

	// Verify comment lines are preserved
	if !lines[0].IsEntry || lines[0].Raw != "" {
		if lines[0].Raw != "# This is a comment" {
			t.Errorf("First line should be a comment, got: %v", lines[0])
		}
	}

	// Add a new URL
	lines = append(lines, Line{
		Entry: &URLEntry{
			URL: "https://example.com/feed4.xml",
		},
		IsEntry: true,
	})

	// Write back
	err = WriteAllLines(urlsPath, lines)
	if err != nil {
		t.Fatalf("Failed to write lines: %v", err)
	}

	// Read the file back
	content, err := os.ReadFile(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	finalContent := string(content)
	expectedContent := initialContent + "https://example.com/feed4.xml\n"

	if finalContent != expectedContent {
		t.Errorf("Content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, finalContent)
	}
}

func TestAddURLPreservesComments(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()
	urlsPath := filepath.Join(testDir, "urls")

	// Create initial urls file with comments
	initialContent := `# This is a comment
https://example.com/feed1.xml
# Another comment
https://example.com/feed2.xml
`

	err := os.WriteFile(urlsPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Read lines
	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read lines: %v", err)
	}

	// Check if URL already exists
	newURL := "https://example.com/feed3.xml"
	for _, line := range lines {
		if line.IsEntry && line.Entry.URL == newURL {
			t.Fatal("URL already exists")
		}
	}

	// Add the new URL
	lines = append(lines, Line{
		Entry: &URLEntry{
			URL: newURL,
		},
		IsEntry: true,
	})

	err = WriteAllLines(urlsPath, lines)
	if err != nil {
		t.Fatalf("Failed to write lines: %v", err)
	}

	// Read the file back
	content, err := os.ReadFile(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	finalContent := string(content)
	expectedContent := initialContent + "https://example.com/feed3.xml\n"

	if finalContent != expectedContent {
		t.Errorf("Content mismatch after AddURL.\nExpected:\n%s\n\nGot:\n%s", expectedContent, finalContent)
	}
}

func TestRemoveURLPreservesComments(t *testing.T) {
	// Create a temporary directory for testing
	testDir := t.TempDir()
	urlsPath := filepath.Join(testDir, "urls")

	// Create initial urls file with comments
	initialContent := `# This is a comment
https://example.com/feed1.xml
# Another comment
https://example.com/feed2.xml
https://example.com/feed3.xml
`

	err := os.WriteFile(urlsPath, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial file: %v", err)
	}

	// Read lines
	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read lines: %v", err)
	}

	// Remove feed2
	var newLines []Line
	for _, line := range lines {
		if line.IsEntry && line.Entry.URL == "https://example.com/feed2.xml" {
			continue
		}
		newLines = append(newLines, line)
	}

	err = WriteAllLines(urlsPath, newLines)
	if err != nil {
		t.Fatalf("Failed to write lines: %v", err)
	}

	// Read the file back
	content, err := os.ReadFile(urlsPath)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	finalContent := string(content)
	expectedContent := `# This is a comment
https://example.com/feed1.xml
# Another comment
https://example.com/feed3.xml
`

	if finalContent != expectedContent {
		t.Errorf("Content mismatch after RemoveURL.\nExpected:\n%s\n\nGot:\n%s", expectedContent, finalContent)
	}
}
