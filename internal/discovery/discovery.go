package discovery

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// DiscoverFeed attempts to discover an RSS/Atom feed URL from a given URL.
// If the URL is already a feed, it returns it as-is.
// If it's a YouTube URL, it extracts the channel ID and returns the YouTube RSS feed.
// If it's an HTML page, it searches for feed links in the HTML.
func DiscoverFeed(url string) (string, error) {
	// Check if it's a YouTube URL first
	if isYouTubeURL(url) {
		return discoverYouTubeFeed(url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")

	// If it's already a feed (XML), return the URL as-is
	if isFeedContentType(contentType) {
		return url, nil
	}

	// If it's HTML, try to discover feed links
	if isHTMLContentType(contentType) {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}
		return discoverFeedFromHTML(string(body), url)
	}

	return "", fmt.Errorf("unsupported content type: %s", contentType)
}

// DiscoverFeedFromHTML discovers feed URLs from HTML content
func DiscoverFeedFromHTML(htmlContent string, baseURL string) (string, error) {
	return discoverFeedFromHTML(htmlContent, baseURL)
}

func discoverFeedFromHTML(htmlContent string, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	feedURL := findFeedLink(doc)
	if feedURL == "" {
		return "", fmt.Errorf("no feed link found in HTML")
	}

	// If the feed URL is relative, make it absolute
	if !strings.HasPrefix(feedURL, "http://") && !strings.HasPrefix(feedURL, "https://") {
		return resolveURL(baseURL, feedURL), nil
	}

	return feedURL, nil
}

func findFeedLink(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "link" {
		var rel, href, typeAttr string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "rel":
				rel = attr.Val
			case "href":
				href = attr.Val
			case "type":
				typeAttr = attr.Val
			}
		}

		// Check if this is an RSS/Atom feed link
		if rel == "alternate" && isFeedType(typeAttr) {
			return href
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findFeedLink(c); result != "" {
			return result
		}
	}

	return ""
}

func isFeedContentType(contentType string) bool {
	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	return contentType == "application/rss+xml" ||
		contentType == "application/atom+xml" ||
		contentType == "application/xml" ||
		contentType == "text/xml"
}

func isHTMLContentType(contentType string) bool {
	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	return contentType == "text/html"
}

func isFeedType(typeAttr string) bool {
	typeAttr = strings.ToLower(typeAttr)
	return typeAttr == "application/rss+xml" || typeAttr == "application/atom+xml"
}

func resolveURL(baseURL, relativeURL string) string {
	if strings.HasPrefix(relativeURL, "/") {
		// Extract scheme and host from base URL
		parts := strings.SplitN(baseURL, "://", 2)
		if len(parts) != 2 {
			return relativeURL
		}
		scheme := parts[0]
		hostAndPath := strings.SplitN(parts[1], "/", 2)
		host := hostAndPath[0]
		return fmt.Sprintf("%s://%s%s", scheme, host, relativeURL)
	}

	// For relative paths without leading slash, append to base URL directory
	lastSlash := strings.LastIndex(baseURL, "/")
	if lastSlash == -1 {
		return relativeURL
	}
	return baseURL[:lastSlash+1] + relativeURL
}

// isYouTubeURL checks if a URL is a YouTube URL
func isYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

// discoverYouTubeFeed extracts the channel ID from a YouTube URL and returns the RSS feed URL
func discoverYouTubeFeed(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch YouTube page: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	channelID, err := extractYouTubeChannelID(string(body))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID), nil
}

// extractYouTubeChannelID extracts the channel ID from YouTube HTML
func extractYouTubeChannelID(htmlContent string) (string, error) {
	// Look for channel_id in various patterns
	patterns := []string{
		`"channelId":"([^"]+)"`,
		`"channel_id":"([^"]+)"`,
		`channel_id=([A-Za-z0-9_-]+)`,
		`/channel/([A-Za-z0-9_-]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not find YouTube channel ID in page")
}

// ExtractYouTubeChannelID is exported for testing
func ExtractYouTubeChannelID(htmlContent string) (string, error) {
	return extractYouTubeChannelID(htmlContent)
}
