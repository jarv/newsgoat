package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"time"

	"github.com/jarv/newsgoat/internal/database"
	pb "github.com/jarv/newsgoat/internal/grpc/gen/newsgoat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	apiKeyHeader       = "x-api-key"
	defaultTimeout     = 30 * time.Second
	connectionTimeout  = 10 * time.Second
)

// Client wraps gRPC clients and implements the same interface as feeds.Manager
type Client struct {
	conn            *grpc.ClientConn
	feedClient      pb.FeedServiceClient
	settingsClient  pb.SettingsServiceClient
	apiKey          string
	serverURL       string
}

// NewClient creates a new gRPC client
// serverURL can be:
//   - host:port (e.g., "localhost:50051") - insecure connection
//   - https://host (e.g., "https://newsgoat.jarv.org") - TLS connection on port 443
//   - https://host:port - TLS connection on specified port
func NewClient(serverURL, apiKey string) (*Client, error) {
	target, creds, err := parseServerURL(serverURL)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for %s: %w", serverURL, err)
	}

	return &Client{
		conn:            conn,
		feedClient:      pb.NewFeedServiceClient(conn),
		settingsClient:  pb.NewSettingsServiceClient(conn),
		apiKey:          apiKey,
		serverURL:       serverURL,
	}, nil
}

// parseServerURL parses the server URL and returns the gRPC target and credentials
func parseServerURL(serverURL string) (string, credentials.TransportCredentials, error) {
	// Try to parse as URL
	u, err := url.Parse(serverURL)
	if err != nil || u.Scheme == "" {
		// Not a URL, treat as host:port with insecure connection
		return serverURL, insecure.NewCredentials(), nil
	}

	switch u.Scheme {
	case "https":
		host := u.Host
		if u.Port() == "" {
			host = u.Hostname() + ":443"
		}
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		return host, credentials.NewTLS(tlsConfig), nil
	case "http":
		host := u.Host
		if u.Port() == "" {
			host = u.Hostname() + ":80"
		}
		return host, insecure.NewCredentials(), nil
	default:
		return "", nil, fmt.Errorf("unsupported URL scheme: %s (use http, https, or host:port)", u.Scheme)
	}
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// context returns a context with API key metadata and timeout
func (c *Client) context() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	if c.apiKey != "" {
		md := metadata.New(map[string]string{
			apiKeyHeader: c.apiKey,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx, cancel
}

// Helper functions to convert protobuf types to database types

func pbFeedToDb(feed *pb.Feed) database.Feed {
	dbFeed := database.Feed{
		ID:          feed.Id,
		Url:         feed.Url,
		Title:       feed.Title,
		Description: feed.Description,
		Visible:     feed.Visible,
	}
	if feed.LastUpdated != nil {
		dbFeed.LastUpdated.Time = feed.LastUpdated.AsTime()
		dbFeed.LastUpdated.Valid = true
	}
	if feed.LastError != "" {
		dbFeed.LastError.String = feed.LastError
		dbFeed.LastError.Valid = true
	}
	if feed.LastErrorTime != nil {
		dbFeed.LastErrorTime.Time = feed.LastErrorTime.AsTime()
		dbFeed.LastErrorTime.Valid = true
	}
	if feed.CreatedAt != nil {
		dbFeed.CreatedAt.Time = feed.CreatedAt.AsTime()
		dbFeed.CreatedAt.Valid = true
	}
	if feed.Etag != "" {
		dbFeed.Etag.String = feed.Etag
		dbFeed.Etag.Valid = true
	}
	if feed.LastModified != "" {
		dbFeed.LastModified.String = feed.LastModified
		dbFeed.LastModified.Valid = true
	}
	if feed.CacheControlMaxAge != 0 {
		dbFeed.CacheControlMaxAge.Int64 = feed.CacheControlMaxAge
		dbFeed.CacheControlMaxAge.Valid = true
	}

	return dbFeed
}

func pbItemWithReadStatusToDb(item *pb.ItemWithReadStatus) database.GetItemsWithReadStatusRow {
	row := database.GetItemsWithReadStatusRow{
		ID:          item.Item.Id,
		FeedID:      item.Item.FeedId,
		Guid:        item.Item.Guid,
		Title:       item.Item.Title,
		Description: item.Item.Description,
		Content:     item.Item.Content,
		Link:        item.Item.Link,
		Read:        item.Read,
	}
	if item.Item.Published != nil {
		row.Published.Time = item.Item.Published.AsTime()
		row.Published.Valid = true
	}
	if item.Item.CreatedAt != nil {
		row.CreatedAt.Time = item.Item.CreatedAt.AsTime()
		row.CreatedAt.Valid = true
	}

	return row
}

func pbFeedStatsToDb(stats *pb.FeedStats) database.GetFeedStatsRow {
	row := database.GetFeedStatsRow{
		ID:          stats.Id,
		Title:       stats.Title,
		Url:         stats.Url,
		TotalItems:  stats.TotalItems,
		UnreadItems: stats.UnreadItems,
	}
	if stats.LastError != "" {
		row.LastError.String = stats.LastError
		row.LastError.Valid = true
	}
	if stats.LastErrorTime != nil {
		row.LastErrorTime.Time = stats.LastErrorTime.AsTime()
		row.LastErrorTime.Valid = true
	}

	return row
}

func pbSearchFeedsByTitleToDb(stats *pb.FeedStats) database.SearchFeedsByTitleRow {
	row := database.SearchFeedsByTitleRow{
		ID:          stats.Id,
		Title:       stats.Title,
		Url:         stats.Url,
		TotalItems:  stats.TotalItems,
		UnreadItems: stats.UnreadItems,
	}

	return row
}

func pbSearchFeedsGloballyToDb(stats *pb.FeedStats) database.SearchFeedsGloballyRow {
	row := database.SearchFeedsGloballyRow{
		ID:          stats.Id,
		Title:       stats.Title,
		Url:         stats.Url,
		TotalItems:  stats.TotalItems,
		UnreadItems: stats.UnreadItems,
	}

	return row
}

func pbSearchItemsByTitleToDb(item *pb.ItemWithReadStatus) database.SearchItemsByTitleRow {
	row := database.SearchItemsByTitleRow{
		ID:          item.Item.Id,
		FeedID:      item.Item.FeedId,
		Guid:        item.Item.Guid,
		Title:       item.Item.Title,
		Description: item.Item.Description,
		Content:     item.Item.Content,
		Link:        item.Item.Link,
		Read:        item.Read,
	}
	if item.Item.Published != nil {
		row.Published.Time = item.Item.Published.AsTime()
		row.Published.Valid = true
	}
	if item.Item.CreatedAt != nil {
		row.CreatedAt.Time = item.Item.CreatedAt.AsTime()
		row.CreatedAt.Valid = true
	}

	return row
}

func pbSearchItemsGloballyToDb(item *pb.ItemWithReadStatus) database.SearchItemsGloballyRow {
	row := database.SearchItemsGloballyRow{
		ID:          item.Item.Id,
		FeedID:      item.Item.FeedId,
		Guid:        item.Item.Guid,
		Title:       item.Item.Title,
		Description: item.Item.Description,
		Content:     item.Item.Content,
		Link:        item.Item.Link,
		Read:        item.Read,
	}
	if item.Item.Published != nil {
		row.Published.Time = item.Item.Published.AsTime()
		row.Published.Valid = true
	}
	if item.Item.CreatedAt != nil {
		row.CreatedAt.Time = item.Item.CreatedAt.AsTime()
		row.CreatedAt.Valid = true
	}

	return row
}

func pbLogMessageToDb(msg *pb.LogMessage) database.LogMessage {
	row := database.LogMessage{
		ID:      msg.Id,
		Level:   msg.Level,
		Message: msg.Message,
	}

	if msg.Timestamp != nil {
		row.Timestamp.Time = msg.Timestamp.AsTime()
		row.Timestamp.Valid = true
	}
	if msg.Details != "" {
		row.Attributes.String = msg.Details
		row.Attributes.Valid = true
	}

	return row
}

// High-level feed operations

func (c *Client) AddFeed(url string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.AddFeed(ctx, &pb.AddFeedRequest{Url: url})
	return err
}

func (c *Client) AddFeedWithoutFetching(url string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.AddFeedWithoutFetching(ctx, &pb.AddFeedWithoutFetchingRequest{Url: url})
	return err
}

func (c *Client) RefreshFeed(feedID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.RefreshFeed(ctx, &pb.RefreshFeedRequest{FeedId: feedID})
	return err
}

func (c *Client) RefreshFeedByURL(url string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.RefreshFeedByURL(ctx, &pb.RefreshFeedByURLRequest{Url: url})
	return err
}

func (c *Client) RefreshAllFeeds() error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.RefreshAllFeeds(ctx, &pb.RefreshAllFeedsRequest{})
	return err
}

// Feed operations

func (c *Client) GetFeed(feedID int64) (database.Feed, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetFeed(ctx, &pb.GetFeedRequest{Id: feedID})
	if err != nil {
		return database.Feed{}, err
	}

	return pbFeedToDb(resp.Feed), nil
}

func (c *Client) GetFeedByURL(url string) (database.Feed, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetFeedByURL(ctx, &pb.GetFeedByURLRequest{Url: url})
	if err != nil {
		return database.Feed{}, err
	}

	return pbFeedToDb(resp.Feed), nil
}

func (c *Client) GetVisibleFeeds() ([]database.Feed, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.ListFeeds(ctx, &pb.ListFeedsRequest{})
	if err != nil {
		return nil, err
	}

	feeds := make([]database.Feed, len(resp.Feeds))
	for i, feed := range resp.Feeds {
		feeds[i] = pbFeedToDb(feed)
	}

	return feeds, nil
}

func (c *Client) GetAllFeeds() ([]database.Feed, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.ListAllFeeds(ctx, &pb.ListAllFeedsRequest{})
	if err != nil {
		return nil, err
	}

	feeds := make([]database.Feed, len(resp.Feeds))
	for i, feed := range resp.Feeds {
		feeds[i] = pbFeedToDb(feed)
	}

	return feeds, nil
}

func (c *Client) DeleteFeed(feedID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.DeleteFeed(ctx, &pb.DeleteFeedRequest{Id: feedID})
	return err
}

func (c *Client) HideFeedByURL(url string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.HideFeedByURL(ctx, &pb.HideFeedByURLRequest{Url: url})
	return err
}

func (c *Client) ShowFeedByURL(url string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.ShowFeedByURL(ctx, &pb.ShowFeedByURLRequest{Url: url})
	return err
}

func (c *Client) GetFeedStats() ([]database.GetFeedStatsRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetFeedStats(ctx, &pb.GetFeedStatsRequest{})
	if err != nil {
		return nil, err
	}

	stats := make([]database.GetFeedStatsRow, len(resp.Stats))
	for i, stat := range resp.Stats {
		stats[i] = pbFeedStatsToDb(stat)
	}

	return stats, nil
}

func (c *Client) SearchFeedsByTitle(pattern string) ([]database.SearchFeedsByTitleRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.SearchFeedsByTitle(ctx, &pb.SearchFeedsByTitleRequest{Query: pattern})
	if err != nil {
		return nil, err
	}

	results := make([]database.SearchFeedsByTitleRow, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = pbSearchFeedsByTitleToDb(result)
	}

	return results, nil
}

func (c *Client) SearchFeedsGlobally(pattern string) ([]database.SearchFeedsGloballyRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.SearchFeedsGlobally(ctx, &pb.SearchFeedsGloballyRequest{Query: pattern})
	if err != nil {
		return nil, err
	}

	results := make([]database.SearchFeedsGloballyRow, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = pbSearchFeedsGloballyToDb(result)
	}

	return results, nil
}

// Item operations

func (c *Client) GetItemsWithReadStatus(feedID int64) ([]database.GetItemsWithReadStatusRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetItemsWithReadStatus(ctx, &pb.GetItemsWithReadStatusRequest{FeedId: feedID})
	if err != nil {
		return nil, err
	}

	items := make([]database.GetItemsWithReadStatusRow, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = pbItemWithReadStatusToDb(item)
	}

	return items, nil
}

func (c *Client) SearchItemsByTitle(feedID int64, pattern string) ([]database.SearchItemsByTitleRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.SearchItemsByTitle(ctx, &pb.SearchItemsByTitleRequest{
		FeedId: feedID,
		Query:  pattern,
	})
	if err != nil {
		return nil, err
	}

	results := make([]database.SearchItemsByTitleRow, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = pbSearchItemsByTitleToDb(result)
	}

	return results, nil
}

func (c *Client) SearchItemsGlobally(feedID int64, pattern string) ([]database.SearchItemsGloballyRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.SearchItemsGlobally(ctx, &pb.SearchItemsGloballyRequest{
		FeedId: feedID,
		Query:  pattern,
	})
	if err != nil {
		return nil, err
	}

	results := make([]database.SearchItemsGloballyRow, len(resp.Results))
	for i, result := range resp.Results {
		results[i] = pbSearchItemsGloballyToDb(result)
	}

	return results, nil
}

// Read status operations

func (c *Client) MarkItemRead(itemID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.MarkItemRead(ctx, &pb.MarkItemReadRequest{ItemId: itemID})
	return err
}

func (c *Client) MarkItemUnread(itemID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.MarkItemUnread(ctx, &pb.MarkItemUnreadRequest{ItemId: itemID})
	return err
}

func (c *Client) MarkAllItemsReadInFeed(feedID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.MarkAllItemsReadInFeed(ctx, &pb.MarkAllItemsReadInFeedRequest{FeedId: feedID})
	return err
}

// Utility methods (client-side, not gRPC)

func (c *Client) ConvertHTMLToMarkdown(input string) string {
	// In client mode, HTML conversion is not supported yet
	// Return input as-is - this could be enhanced later
	return input
}

func (c *Client) ExtractLinks(content string) []string {
	// This is a client-side utility that doesn't need to call the server
	// For now, return empty slice - implement if needed
	return []string{}
}

func (c *Client) AddLinkMarkersToHTML(content string) (string, []string) {
	// This is a client-side utility that doesn't need to call the server
	// For now, return input unchanged - implement if needed
	return content, []string{}
}

// Log operations

func (c *Client) GetLogMessages(limit int64) ([]database.LogMessage, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetLogMessages(ctx, &pb.GetLogMessagesRequest{Limit: limit})
	if err != nil {
		return nil, err
	}

	messages := make([]database.LogMessage, len(resp.Messages))
	for i, msg := range resp.Messages {
		messages[i] = pbLogMessageToDb(msg)
	}

	return messages, nil
}

func (c *Client) GetLogMessage(id int64) (database.LogMessage, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetLogMessage(ctx, &pb.GetLogMessageRequest{Id: id})
	if err != nil {
		return database.LogMessage{}, err
	}

	return pbLogMessageToDb(resp.Message), nil
}

func (c *Client) DeleteAllLogMessages() error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.DeleteAllLogMessages(ctx, &pb.DeleteAllLogMessagesRequest{})
	return err
}

// Folder operations

func (c *Client) AddFeedFolder(feedID int64, folder string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.AddFeedFolder(ctx, &pb.AddFeedFolderRequest{
		FeedId: feedID,
		Folder: folder,
	})
	return err
}

func (c *Client) GetFeedFolders(feedID int64) ([]string, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetFeedFolders(ctx, &pb.GetFeedFoldersRequest{FeedId: feedID})
	if err != nil {
		return nil, err
	}
	return resp.Folders, nil
}

func (c *Client) DeleteFeedFolders(feedID int64) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.feedClient.DeleteFeedFolders(ctx, &pb.DeleteFeedFoldersRequest{FeedId: feedID})
	return err
}

func (c *Client) GetFolderStats() ([]database.GetFolderStatsRow, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.feedClient.GetFolderStats(ctx, &pb.GetFolderStatsRequest{})
	if err != nil {
		return nil, err
	}

	stats := make([]database.GetFolderStatsRow, len(resp.Stats))
	for i, stat := range resp.Stats {
		stats[i] = database.GetFolderStatsRow{
			FolderName:  stat.Name,
			TotalItems:  stat.TotalItems,
			UnreadItems: stat.UnreadItems,
		}
	}
	return stats, nil
}

// SetRefreshCallbacks is not supported in client mode (callbacks are server-side)
func (c *Client) SetRefreshCallbacks(onStart, onComplete func(int64)) {
	// No-op in client mode - callbacks only work in standalone mode
}

// Settings operations (implements config.SettingsManager)

func (c *Client) GetSetting(key string) (string, error) {
	ctx, cancel := c.context()
	defer cancel()

	resp, err := c.settingsClient.GetSetting(ctx, &pb.GetSettingRequest{Key: key})
	if err != nil {
		return "", err
	}
	return resp.Setting.Value, nil
}

func (c *Client) SetSetting(key, value string) error {
	ctx, cancel := c.context()
	defer cancel()

	_, err := c.settingsClient.SetSetting(ctx, &pb.SetSettingRequest{
		Key:   key,
		Value: value,
	})
	return err
}
