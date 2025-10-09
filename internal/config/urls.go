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
	file, err := os.Open(urlsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []URLEntry{}, nil
		}
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var entries []URLEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first whitespace to separate URL from folders
		parts := strings.Fields(line)
		if len(parts) == 0 {
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

		entries = append(entries, entry)
	}

	return entries, scanner.Err()
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

func AddURL(url string) error {
	entries, err := ReadURLsFile()
	if err != nil {
		return err
	}

	// Check if URL already exists
	for _, entry := range entries {
		if entry.URL == url {
			return nil
		}
	}

	// Convert entries back to strings for writing
	urls := make([]string, len(entries)+1)
	for i, entry := range entries {
		if len(entry.Folders) > 0 {
			urls[i] = entry.URL + " " + strings.Join(entry.Folders, ",")
		} else {
			urls[i] = entry.URL
		}
	}
	urls[len(urls)-1] = url

	return WriteURLsFile(urls)
}

// AddURLLine adds a complete URL line (including folders) to the URLs file
func AddURLLine(line string) error {
	entries, err := ReadURLsFile()
	if err != nil {
		return err
	}

	// Parse the line to get the URL
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	newURL := parts[0]

	// Check if URL already exists
	for _, entry := range entries {
		if entry.URL == newURL {
			return nil
		}
	}

	// Convert entries back to strings for writing
	urls := make([]string, len(entries)+1)
	for i, entry := range entries {
		if len(entry.Folders) > 0 {
			urls[i] = entry.URL + " " + strings.Join(entry.Folders, ",")
		} else {
			urls[i] = entry.URL
		}
	}
	urls[len(urls)-1] = line

	return WriteURLsFile(urls)
}

func RemoveURL(url string) error {
	entries, err := ReadURLsFile()
	if err != nil {
		return err
	}

	// Filter out the URL to remove
	var newURLs []string
	for _, entry := range entries {
		if entry.URL != url {
			if len(entry.Folders) > 0 {
				newURLs = append(newURLs, entry.URL+" "+strings.Join(entry.Folders, ","))
			} else {
				newURLs = append(newURLs, entry.URL)
			}
		}
	}

	return WriteURLsFile(newURLs)
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
	if _, err := writer.WriteString("# Add your RSS feeds to this file\n"); err != nil {
		return err
	}

	return writer.Flush()
}
