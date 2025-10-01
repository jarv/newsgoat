package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

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

func ReadURLsFile() ([]string, error) {
	urlsPath, err := GetURLsFilePath()
	if err != nil {
		return nil, err
	}

	return ReadURLsFileFromPath(urlsPath)
}

func ReadURLsFileFromPath(urlsPath string) ([]string, error) {
	file, err := os.Open(urlsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	return urls, scanner.Err()
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
	urls, err := ReadURLsFile()
	if err != nil {
		return err
	}

	for _, existingURL := range urls {
		if existingURL == url {
			return nil
		}
	}

	urls = append(urls, url)
	return WriteURLsFile(urls)
}

func RemoveURL(url string) error {
	urls, err := ReadURLsFile()
	if err != nil {
		return err
	}

	var newURLs []string
	for _, existingURL := range urls {
		if existingURL != url {
			newURLs = append(newURLs, existingURL)
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

	sampleURLs := []string{
		"# NewsGoat RSS Feeds",
		"# Add your RSS feed URLs below, one per line",
		"# Lines starting with # are comments",
		"",
		"https://feeds.feedburner.com/techcrunch",
		"https://rss.cnn.com/rss/edition.rss",
		"https://feeds.bbci.co.uk/news/rss.xml",
	}

	file, err := os.Create(urlsPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	writer := bufio.NewWriter(file)
	for _, line := range sampleURLs {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}