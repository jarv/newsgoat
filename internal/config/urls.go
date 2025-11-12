package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// URLEntry represents a feed URL with optional folders
type URLEntry struct {
	URL     string
	Folders []string
}

// Line represents a line in the URLs file (either a URL entry or a comment/blank line)
type Line struct {
	Entry   *URLEntry
	Raw     string // For comments and blank lines
	IsEntry bool
}

// GetEditor returns the editor to use from the EDITOR environment variable
func GetEditor() string {
	return os.Getenv("EDITOR")
}

// ParseFolders parses a comma-separated list of folders, handling quoted strings
func ParseFolders(folderStr string) []string {
	if folderStr == "" {
		return nil
	}

	var folders []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(folderStr); i++ {
		ch := folderStr[i]

		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				current.WriteByte(ch)
			} else {
				// End of folder name
				folder := strings.TrimSpace(current.String())
				if folder != "" {
					folders = append(folders, folder)
				}
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	// Add last folder
	folder := strings.TrimSpace(current.String())
	if folder != "" {
		folders = append(folders, folder)
	}

	return folders
}

func ReadURLsFileFromPath(urlsPath string) ([]URLEntry, error) {
	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		return nil, err
	}

	var entries []URLEntry
	for _, line := range lines {
		if line.IsEntry {
			entries = append(entries, *line.Entry)
		}
	}

	return entries, nil
}

// ReadAllLinesFromPath reads all lines from the URLs file, preserving comments and blank lines
func ReadAllLinesFromPath(urlsPath string) ([]Line, error) {
	file, err := os.Open(urlsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Line{}, nil
		}
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var lines []Line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawLine := scanner.Text()
		trimmedLine := strings.TrimSpace(rawLine)

		// Check if it's a comment or blank line
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			lines = append(lines, Line{
				Raw:     rawLine,
				IsEntry: false,
			})
			continue
		}

		// Split on first whitespace to separate URL from folders
		parts := strings.Fields(trimmedLine)
		if len(parts) == 0 {
			lines = append(lines, Line{
				Raw:     rawLine,
				IsEntry: false,
			})
			continue
		}

		entry := URLEntry{
			URL: parts[0],
		}

		// If there are more parts, parse folders
		if len(parts) > 1 {
			// Join remaining parts and parse as folders
			folderStr := strings.Join(parts[1:], " ")
			entry.Folders = ParseFolders(folderStr)
		}

		lines = append(lines, Line{
			Entry:   &entry,
			IsEntry: true,
		})
	}

	return lines, scanner.Err()
}

// WriteAllLines writes all lines back to the URLs file, preserving comments and blank lines
func WriteAllLines(urlsPath string, lines []Line) error {
	dir := filepath.Dir(urlsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(urlsPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		var output string
		if line.IsEntry {
			output = line.Entry.URL
			if len(line.Entry.Folders) > 0 {
				output += " " + strings.Join(line.Entry.Folders, ",")
			}
		} else {
			output = line.Raw
		}
		if _, err := writer.WriteString(output + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// FeedWithFolders represents a feed URL with its folder tags
type FeedWithFolders struct {
	URL     string
	Folders []string
}

// ExportURLsToTempFile exports all visible feeds to a temporary file
// The temp file is created in the system temp directory with a unique name
// Returns the path to the temp file or an error
// The caller is responsible for deleting the temp file when done
func ExportURLsToTempFile(feeds []FeedWithFolders) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp("", "newsgoat-urls-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
	}()

	// Write header
	writer := bufio.NewWriter(tempFile)
	header := `# Edit this file to manage your RSS feeds
#
# Format: <url> [folder1,folder2,...]
# - Each line should contain a feed URL
# - Optionally, you can add one or more folder names after the URL (comma-separated)
# - Folders with spaces should be quoted: "Folder Name"
# - Lines starting with # are comments and will be ignored
# - Save and close this file to update your feeds
#
`
	if _, err := writer.WriteString(header); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// Write each feed with its folders
	for _, feed := range feeds {
		// Format the line
		line := feed.URL
		if len(feed.Folders) > 0 {
			// Quote folders that contain spaces or commas
			var quotedFolders []string
			for _, folder := range feed.Folders {
				if strings.Contains(folder, " ") || strings.Contains(folder, ",") {
					quotedFolders = append(quotedFolders, `"`+folder+`"`)
				} else {
					quotedFolders = append(quotedFolders, folder)
				}
			}
			line += " " + strings.Join(quotedFolders, ",")
		}
		line += "\n"

		if _, err := writer.WriteString(line); err != nil {
			_ = os.Remove(tempPath)
			return "", fmt.Errorf("failed to write feed: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("failed to flush writer: %w", err)
	}

	return tempPath, nil
}
