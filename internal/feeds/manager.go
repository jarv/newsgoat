package feeds

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/version"
	"github.com/mmcdole/gofeed"
)

const FeedTimeout = 30 * time.Second

// Type aliases for convenience
type LogMessage = database.LogMessage

// conditionalRequestTransport wraps http.RoundTripper to add conditional request headers and User-Agent
type conditionalRequestTransport struct {
	Transport http.RoundTripper
	UserAgent string
	Manager   *Manager
	FeedURL   string
}

func (t *conditionalRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Set User-Agent
	req.Header.Set("User-Agent", t.UserAgent)

	// Add conditional request headers if we have them
	if t.Manager != nil && t.FeedURL != "" {
		t.Manager.dbMutex.RLock()
		feed, err := t.Manager.queries.GetFeedByURL(context.Background(), t.FeedURL)
		t.Manager.dbMutex.RUnlock()

		if err == nil {
			// Add If-None-Match header if we have an ETag
			if feed.Etag.Valid && feed.Etag.String != "" {
				req.Header.Set("If-None-Match", feed.Etag.String)
			}

			// Add If-Modified-Since header if we have a Last-Modified
			if feed.LastModified.Valid && feed.LastModified.String != "" {
				req.Header.Set("If-Modified-Since", feed.LastModified.String)
			}
		}
	}

	return t.Transport.RoundTrip(req)
}

type Manager struct {
	db               *sql.DB
	queries          *database.Queries
	parser           *gofeed.Parser
	refreshCallbacks map[int64]func(int64) // Callbacks for refresh events
	dbMutex          sync.RWMutex          // Global RWMutex for database operations
}

// createHTTPClientForFeed creates an HTTP client with conditional request support for a specific feed URL
func (m *Manager) createHTTPClientForFeed(feedURL string) *http.Client {
	return &http.Client{
		Timeout: FeedTimeout,
		Transport: &conditionalRequestTransport{
			Transport: http.DefaultTransport,
			UserAgent: version.GetUserAgent(),
			Manager:   m,
			FeedURL:   feedURL,
		},
	}
}

// parseCacheControl extracts max-age from Cache-Control header
func parseCacheControl(cacheControl string) (maxAge int64, hasMaxAge bool) {
	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			val := strings.TrimPrefix(part, "max-age=")
			if age, err := strconv.ParseInt(val, 10, 64); err == nil {
				return age, true
			}
		}
	}
	return 0, false
}

func (m *Manager) ConvertHTMLToMarkdown(input string) string {
	if input == "" {
		return ""
	}

	// Convert HTML to markdown using html-to-markdown v2
	markdown, err := md.ConvertString(input)
	if err != nil {
		logging.Warn("Failed to convert HTML to markdown", "error", err)
		// Fallback to original text if conversion fails
		return input
	}

	// Clean up excessive whitespace
	markdown = strings.TrimSpace(markdown)
	lines := strings.Split(markdown, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

func NewManager(db *sql.DB, queries *database.Queries) *Manager {
	// Create parser - we'll set the client per-request
	parser := gofeed.NewParser()

	return &Manager{
		db:               db,
		queries:          queries,
		parser:           parser,
		refreshCallbacks: make(map[int64]func(int64)),
	}
}

func (m *Manager) SetRefreshCallbacks(onStart, onComplete func(int64)) {
	// For now, we'll use a simple approach - in a full implementation
	// you might want to have separate start/complete callbacks
	for feedID := range m.refreshCallbacks {
		delete(m.refreshCallbacks, feedID)
	}
}

func (m *Manager) AddFeed(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), FeedTimeout)
	defer cancel()

	feed, err := m.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		logging.Error("Error parsing feed during add", "url", url, "error", err)
		return err
	}

	now := sql.NullTime{Time: time.Now(), Valid: true}

	m.dbMutex.Lock()
	_, err = m.queries.CreateFeed(context.Background(), database.CreateFeedParams{
		Url:         url,
		Title:       feed.Title,
		Description: feed.Description,
		LastUpdated: now,
		Visible:     true,
	})
	m.dbMutex.Unlock()

	if err != nil {
		return err
	}

	return m.RefreshFeedByURL(url)
}

// AddFeedWithoutFetching adds a feed to the database without fetching its content
// The feed title will be the URL until it's manually refreshed
func (m *Manager) AddFeedWithoutFetching(url string) error {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	_, err := m.queries.CreateFeed(context.Background(), database.CreateFeedParams{
		Url:         url,
		Title:       url, // Use URL as title until fetched
		Description: "",
		LastUpdated: sql.NullTime{Valid: false}, // Not yet fetched
		Visible:     true,
	})

	return err
}

// HideFeedByURL hides a feed by setting visible = false
func (m *Manager) HideFeedByURL(url string) error {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	return m.queries.HideFeedByURL(context.Background(), url)
}

// ShowFeedByURL shows a feed by setting visible = true
func (m *Manager) ShowFeedByURL(url string) error {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()

	return m.queries.ShowFeedByURL(context.Background(), url)
}

// GetAllFeeds returns all feeds (both visible and hidden)
func (m *Manager) GetAllFeeds() ([]database.Feed, error) {
	m.dbMutex.RLock()
	defer m.dbMutex.RUnlock()

	return m.queries.ListAllFeeds(context.Background())
}

func (m *Manager) RefreshFeedByURL(url string) error {
	m.dbMutex.RLock()
	feed, err := m.queries.GetFeedByURL(context.Background(), url)
	m.dbMutex.RUnlock()
	if err != nil {
		return err
	}

	return m.RefreshFeed(feed.ID)
}

func (m *Manager) RefreshFeed(feedID int64) error {
	var feed database.Feed

	// Get feed with read lock
	m.dbMutex.RLock()
	feed, err := m.queries.GetFeed(context.Background(), feedID)
	m.dbMutex.RUnlock()
	if err != nil {
		return err
	}

	// Check if feed is still within cache control max age period
	if feed.CacheControlMaxAge.Valid && feed.LastUpdated.Valid {
		cacheExpiry := feed.LastUpdated.Time.Add(time.Duration(feed.CacheControlMaxAge.Int64) * time.Second)
		if time.Now().Before(cacheExpiry) {
			logging.Debug("Feed still within cache control period, skipping fetch",
				"url", feed.Url,
				"lastUpdated", feed.LastUpdated.Time,
				"maxAge", feed.CacheControlMaxAge.Int64,
				"expiresAt", cacheExpiry)
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), FeedTimeout)
	defer cancel()

	// Create HTTP client with conditional request support
	client := m.createHTTPClientForFeed(feed.Url)

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", feed.Url, nil)
	if err != nil {
		logging.Error("Error creating request", "url", feed.Url, "error", err)
		m.recordFeedError(feedID, err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.Error("Error fetching feed", "url", feed.Url, "error", err)
		m.recordFeedError(feedID, err)
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Handle 304 Not Modified - feed hasn't changed
	if resp.StatusCode == http.StatusNotModified {
		logging.Debug("Feed not modified", "url", feed.Url, "status", resp.StatusCode)
		// Update last_updated to track that we checked
		now := sql.NullTime{Time: time.Now(), Valid: true}
		m.dbMutex.Lock()
		err = m.queries.UpdateFeed(context.Background(), database.UpdateFeedParams{
			ID:                 feedID,
			Title:              feed.Title,
			Description:        feed.Description,
			LastUpdated:        now,
			Etag:               feed.Etag,
			LastModified:       feed.LastModified,
			CacheControlMaxAge: feed.CacheControlMaxAge,
		})
		m.dbMutex.Unlock()
		return err
	}

	// Check for HTTP error status codes (anything not 2xx)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		logging.Error("HTTP error fetching feed", "url", feed.Url, "status", resp.StatusCode, "error", err)
		m.recordFeedError(feedID, err)
		return err
	}

	// Parse response headers
	etag := sql.NullString{String: resp.Header.Get("ETag"), Valid: resp.Header.Get("ETag") != ""}
	lastModified := sql.NullString{String: resp.Header.Get("Last-Modified"), Valid: resp.Header.Get("Last-Modified") != ""}

	var cacheControlMaxAge sql.NullInt64
	if cacheControl := resp.Header.Get("Cache-Control"); cacheControl != "" {
		if maxAge, hasMaxAge := parseCacheControl(cacheControl); hasMaxAge {
			cacheControlMaxAge = sql.NullInt64{Int64: maxAge, Valid: true}
		}
	}

	// Parse the feed
	parsedFeed, err := m.parser.Parse(resp.Body)
	if err != nil {
		logging.Error("Error parsing feed", "url", feed.Url, "error", err)
		m.recordFeedError(feedID, err)
		return err
	}

	// Clear any previous error since this fetch was successful
	m.recordFeedError(feedID, nil)

	// Update feed with headers
	now := sql.NullTime{Time: time.Now(), Valid: true}
	m.dbMutex.Lock()
	err = m.queries.UpdateFeed(context.Background(), database.UpdateFeedParams{
		ID:                 feedID,
		Title:              parsedFeed.Title,
		Description:        parsedFeed.Description,
		LastUpdated:        now,
		Etag:               etag,
		LastModified:       lastModified,
		CacheControlMaxAge: cacheControlMaxAge,
	})
	m.dbMutex.Unlock()
	if err != nil {
		return err
	}

	for _, item := range parsedFeed.Items {
		var published sql.NullTime
		if item.PublishedParsed != nil {
			published = sql.NullTime{Time: *item.PublishedParsed, Valid: true}
		}

		content := item.Content
		if content == "" && item.Description != "" {
			content = item.Description
		}

		description := item.Description

		// For YouTube feeds, extract media:description from extensions
		if content == "" && description == "" {
			if mediaExt, ok := item.Extensions["media"]; ok {
				if groupList, ok := mediaExt["group"]; ok && len(groupList) > 0 {
					if descList, ok := groupList[0].Children["description"]; ok && len(descList) > 0 {
						mediaDesc := descList[0].Value
						content = mediaDesc
						description = mediaDesc
					}
				}
			}
		}

		// Upsert item
		m.dbMutex.Lock()
		_, err := m.queries.UpsertItem(context.Background(), database.UpsertItemParams{
			FeedID:      feedID,
			Guid:        item.GUID,
			Title:       item.Title,
			Description: description,
			Content:     content,
			Link:        item.Link,
			Published:   published,
		})
		m.dbMutex.Unlock()
		if err != nil {
			logging.Error("Error upserting item", "guid", item.GUID, "error", err)
		}
	}

	return nil
}

func (m *Manager) RefreshAllFeeds() error {
	m.dbMutex.RLock()
	feeds, err := m.queries.ListFeeds(context.Background())
	m.dbMutex.RUnlock()
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		if err := m.RefreshFeed(feed.ID); err != nil {
			logging.Error("Error refreshing feed", "url", feed.Url, "error", err)
		}
	}

	return nil
}

func (m *Manager) GetFeedStats() ([]database.GetFeedStatsRow, error) {
	m.dbMutex.RLock()
	result, err := m.queries.GetFeedStats(context.Background())
	m.dbMutex.RUnlock()
	return result, err
}

func (m *Manager) GetItemsWithReadStatus(feedID int64) ([]database.GetItemsWithReadStatusRow, error) {
	m.dbMutex.RLock()
	result, err := m.queries.GetItemsWithReadStatus(context.Background(), feedID)
	m.dbMutex.RUnlock()
	return result, err
}

func (m *Manager) MarkItemRead(itemID int64) error {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()
	return m.queries.MarkItemRead(context.Background(), itemID)
}

func (m *Manager) MarkItemUnread(itemID int64) error {
	m.dbMutex.Lock()
	err := m.queries.MarkItemUnread(context.Background(), itemID)
	m.dbMutex.Unlock()
	return err
}

func (m *Manager) MarkAllItemsReadInFeed(feedID int64) error {
	m.dbMutex.Lock()
	err := m.queries.MarkAllItemsReadInFeed(context.Background(), feedID)
	m.dbMutex.Unlock()
	return err
}

func (m *Manager) DeleteFeed(feedID int64) error {
	m.dbMutex.Lock()
	err := m.queries.DeleteFeed(context.Background(), feedID)
	m.dbMutex.Unlock()
	return err
}

func (m *Manager) ExtractLinks(content string) []string {
	var links []string
	seen := make(map[string]bool)

	// Extract from markdown [text](url) syntax
	mdLinkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := mdLinkPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			link := match[2]
			if (strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://")) && !seen[link] {
				links = append(links, link)
				seen[link] = true
			}
		}
	}

	// Extract from HTML <a href="..."> tags (for content that wasn't converted)
	hrefPattern := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["']`)
	matches = hrefPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			link := match[1]
			if (strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://")) && !seen[link] {
				links = append(links, link)
				seen[link] = true
			}
		}
	}

	// Also extract plain URLs from text
	if strings.Contains(content, "http") {
		words := strings.Fields(content)
		for _, word := range words {
			if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
				link := strings.TrimRight(word, ".,!?;)")
				if !seen[link] {
					links = append(links, link)
					seen[link] = true
				}
			}
		}
	}

	return links
}

func (m *Manager) GetLogMessages(limit int64) ([]LogMessage, error) {
	m.dbMutex.RLock()
	result, err := m.queries.GetLogMessages(context.Background(), limit)
	m.dbMutex.RUnlock()
	return result, err
}

func (m *Manager) GetLogMessage(id int64) (LogMessage, error) {
	m.dbMutex.RLock()
	result, err := m.queries.GetLogMessage(context.Background(), id)
	m.dbMutex.RUnlock()
	return result, err
}

func (m *Manager) DeleteAllLogMessages() error {
	m.dbMutex.Lock()
	defer m.dbMutex.Unlock()
	return m.queries.DeleteAllLogMessages(context.Background())
}

func (m *Manager) recordFeedError(feedID int64, err error) {
	if err == nil {
		// Clear any previous error
		m.dbMutex.Lock()
		retryErr := m.queries.ClearFeedError(context.Background(), feedID)
		m.dbMutex.Unlock()
		if retryErr != nil {
			logging.Error("Failed to clear feed error", "feedID", feedID, "error", retryErr)
		}
		return
	}

	// Record the error
	now := sql.NullTime{Time: time.Now(), Valid: true}
	errorText := sql.NullString{String: err.Error(), Valid: true}

	m.dbMutex.Lock()
	retryErr := m.queries.UpdateFeedError(context.Background(), database.UpdateFeedErrorParams{
		ID:            feedID,
		LastError:     errorText,
		LastErrorTime: now,
	})
	m.dbMutex.Unlock()
	if retryErr != nil {
		logging.Error("Failed to update feed error", "feedID", feedID, "error", retryErr)
	}
}
