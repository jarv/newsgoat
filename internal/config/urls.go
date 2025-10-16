package config

import (
	"bufio"
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

func GetURLsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Try new location first: ~/.config/newsgoat/urls
	newPath := filepath.Join(homeDir, ".config", "newsgoat", "urls")
	if _, err := os.Stat(newPath); err == nil {
		return newPath, nil
	}

	// Fall back to old location: ~/.newsgoat/urls
	oldPath := filepath.Join(homeDir, ".newsgoat", "urls")
	if _, err := os.Stat(oldPath); err == nil {
		return oldPath, nil
	}

	// If neither exists, return new location (will be created there)
	return newPath, nil
}

func ReadURLsFile() ([]URLEntry, error) {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return nil, err
	}

	return ReadURLsFileFromPath(urlsPath)
}

// parseFolders parses a comma-separated list of folders, handling quoted strings
func parseFolders(folderStr string) []string {
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
			entry.Folders = parseFolders(folderStr)
		}

		lines = append(lines, Line{
			Entry:   &entry,
			IsEntry: true,
		})
	}

	return lines, scanner.Err()
}

func WriteURLsFile(urls []string) error {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return err
	}

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
	for _, url := range urls {
		if _, err := writer.WriteString(url + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
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

func AddURL(url string) error {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return err
	}

	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		return err
	}

	// Check if URL already exists
	for _, line := range lines {
		if line.IsEntry && line.Entry.URL == url {
			return nil
		}
	}

	// Add the new URL as a line
	lines = append(lines, Line{
		Entry: &URLEntry{
			URL: url,
		},
		IsEntry: true,
	})

	return WriteAllLines(urlsPath, lines)
}

// AddURLLine adds a complete URL line (including folders) to the URLs file
func AddURLLine(lineStr string) error {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return err
	}

	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		return err
	}

	// Parse the line to get the URL
	parts := strings.Fields(lineStr)
	if len(parts) == 0 {
		return nil
	}
	newURL := parts[0]

	// Check if URL already exists
	for _, line := range lines {
		if line.IsEntry && line.Entry.URL == newURL {
			return nil
		}
	}

	// Parse the new entry
	entry := URLEntry{
		URL: newURL,
	}
	if len(parts) > 1 {
		folderStr := strings.Join(parts[1:], " ")
		entry.Folders = parseFolders(folderStr)
	}

	// Add the new line
	lines = append(lines, Line{
		Entry:   &entry,
		IsEntry: true,
	})

	return WriteAllLines(urlsPath, lines)
}

func RemoveURL(url string) error {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return err
	}

	lines, err := ReadAllLinesFromPath(urlsPath)
	if err != nil {
		return err
	}

	// Filter out the URL to remove
	var newLines []Line
	for _, line := range lines {
		if line.IsEntry && line.Entry.URL == url {
			continue
		}
		newLines = append(newLines, line)
	}

	return WriteAllLines(urlsPath, newLines)
}

func CreateSampleURLsFile() error {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(urlsPath); err == nil {
		return nil
	}

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

	// Write header with instructions and examples
	header := `# Add your RSS feeds to this file
#
# Format: <url> [folder1,folder2,...]
# - Each line should contain a feed URL
# - Optionally, you can add one or more folder names after the URL (comma-separated)
# - Folders with spaces should be quoted: "Folder Name"
# - Lines starting with # are comments and will be ignored
#
# For example:
# https://www.newscientist.com/feed/home/
# https://arstechnica.com/feed/ "Tech News"
#
`

	if _, err := writer.WriteString(header); err != nil {
		return err
	}

	return writer.Flush()
}
