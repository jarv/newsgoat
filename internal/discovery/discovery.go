package discovery

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// URLType represents the type of URL detected
type URLType string

const (
	URLTypeYouTube URLType = "youtube"
	URLTypeGitHub  URLType = "github"
	URLTypeGitLab  URLType = "gitlab"
	URLTypeGeneric URLType = "generic"
)

// DiscoverFeed attempts to discover an RSS/Atom feed URL from a given URL.
// If the URL is already a feed, it returns it as-is.
// If it's a YouTube URL, it extracts the channel ID and returns the YouTube RSS feed.
// If it's a GitHub URL, it converts it to the appropriate Atom feed URL.
// If it's a GitLab URL, it converts it to the appropriate Atom feed URL.
// If it's an HTML page, it searches for feed links in the HTML.
func DiscoverFeed(url string) (string, error) {
	// If URL already looks like a feed (ends with .atom, .xml, .rss), treat it as generic
	if isLikelyFeedURL(url) {
		// Skip GitHub/GitLab pattern matching and go straight to content type check
		return checkGenericFeed(url)
	}

	// Check URL type and handle accordingly
	urlType := GetURLType(url)

	switch urlType {
	case URLTypeYouTube:
		return discoverYouTubeFeed(url)
	case URLTypeGitHub:
		return discoverGitHubFeed(url)
	case URLTypeGitLab:
		return discoverGitLabFeed(url)
	}

	// For generic URLs, fetch and check content type
	return checkGenericFeed(url)
}

// isLikelyFeedURL checks if a URL ends with common feed extensions
func isLikelyFeedURL(url string) bool {
	// Strip query string for checking
	baseURL := url
	if idx := strings.Index(url, "?"); idx != -1 {
		baseURL = url[:idx]
	}
	return strings.HasSuffix(baseURL, ".atom") ||
		strings.HasSuffix(baseURL, ".xml") ||
		strings.HasSuffix(baseURL, ".rss")
}

// checkGenericFeed fetches a URL and checks if it's a feed based on content type
func checkGenericFeed(url string) (string, error) {
	// For generic URLs, fetch and check content type
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

// GetURLType determines the type of URL
func GetURLType(url string) URLType {
	if isYouTubeURL(url) {
		return URLTypeYouTube
	}
	if isGitHubURL(url) {
		return URLTypeGitHub
	}
	if isGitLabURL(url) {
		return URLTypeGitLab
	}
	return URLTypeGeneric
}

// isYouTubeURL checks if a URL is a YouTube URL
func isYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be")
}

// isGitHubURL checks if a URL is a GitHub URL
func isGitHubURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/")
}

// isGitLabURL checks if a URL is a GitLab URL
func isGitLabURL(url string) bool {
	return strings.HasPrefix(url, "https://gitlab.com/")
}

// discoverGitHubFeed converts a GitHub URL to its Atom feed URL
// Handles URLs like:
// - https://github.com/<repo>/tree/<branch>/path -> https://github.com/<repo>/commits/<branch>/path.atom
// - https://github.com/<repo>/blob/<branch>/path -> https://github.com/<repo>/commits/<branch>/path.atom
// - https://github.com/<repo>/commits/<branch>/path -> https://github.com/<repo>/commits/<branch>/path.atom
func discoverGitHubFeed(url string) (string, error) {
	// Strip query string if present
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Pattern: https://github.com/<repo path>/{tree|blob|commits}/<branch>/path/to/location
	pattern := regexp.MustCompile(`^(https://github\.com/[^/]+/[^/]+)/(tree|blob|commits)/(.+)$`)
	matches := pattern.FindStringSubmatch(url)

	if len(matches) != 4 {
		return "", fmt.Errorf("URL does not match GitHub tree/blob/commits pattern")
	}

	repoBase := matches[1]      // https://github.com/owner/repo
	pathType := matches[2]      // tree, blob, or commits
	branchAndPath := matches[3] // branch/path/to/location

	// Replace tree or blob with commits (if it's already commits, this is a no-op)
	_ = pathType // We don't actually need this since we always use "commits"

	// Construct the Atom feed URL
	feedURL := fmt.Sprintf("%s/commits/%s.atom", repoBase, branchAndPath)

	return feedURL, nil
}

// discoverGitLabFeed converts a GitLab URL to its Atom feed URL
// Handles URLs like:
// - https://gitlab.com/<project>/-/tree/<branch>/path -> https://gitlab.com/<project>/-/commits/<branch>/path?format=atom
// - https://gitlab.com/<project>/-/blob/<branch>/path -> https://gitlab.com/<project>/-/commits/<branch>/path?format=atom
// - https://gitlab.com/<project>/-/commits/<branch>/path -> https://gitlab.com/<project>/-/commits/<branch>/path?format=atom
func discoverGitLabFeed(url string) (string, error) {
	// Strip query string if present
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Pattern: https://gitlab.com/<project path>/-/{tree|blob|commits}/<branch>/path/to/location
	pattern := regexp.MustCompile(`^(https://gitlab\.com/[^/]+(?:/[^/]+)*)/-/(tree|blob|commits)/(.+)$`)
	matches := pattern.FindStringSubmatch(url)

	if len(matches) != 4 {
		return "", fmt.Errorf("URL does not match GitLab tree/blob/commits pattern")
	}

	projectBase := matches[1]   // https://gitlab.com/group/project or https://gitlab.com/group/subgroup/project
	pathType := matches[2]      // tree, blob, or commits
	branchAndPath := matches[3] // branch/path/to/location

	// Replace tree or blob with commits (if it's already commits, this is a no-op)
	_ = pathType // We don't actually need this since we always use "commits"

	// Construct the Atom feed URL with ?format=atom
	feedURL := fmt.Sprintf("%s/-/commits/%s?format=atom", projectBase, branchAndPath)

	return feedURL, nil
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
