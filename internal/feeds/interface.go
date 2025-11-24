package feeds

import "github.com/jarv/newsgoat/internal/database"

// FeedManager is the interface that abstracts feed operations.
// Both the local Manager and the gRPC Client implement this interface.
type FeedManager interface {
	// High-level feed operations
	AddFeed(url string) error
	AddFeedWithoutFetching(url string) error
	RefreshFeed(feedID int64) error
	RefreshFeedByURL(url string) error
	RefreshAllFeeds() error

	// Feed operations
	GetFeed(feedID int64) (database.Feed, error)
	GetFeedByURL(url string) (database.Feed, error)
	GetVisibleFeeds() ([]database.Feed, error)
	GetAllFeeds() ([]database.Feed, error)
	DeleteFeed(feedID int64) error
	HideFeedByURL(url string) error
	ShowFeedByURL(url string) error
	GetFeedStats() ([]database.GetFeedStatsRow, error)
	SearchFeedsByTitle(pattern string) ([]database.SearchFeedsByTitleRow, error)
	SearchFeedsGlobally(pattern string) ([]database.SearchFeedsGloballyRow, error)

	// Item operations
	GetItemsWithReadStatus(feedID int64) ([]database.GetItemsWithReadStatusRow, error)
	SearchItemsByTitle(feedID int64, pattern string) ([]database.SearchItemsByTitleRow, error)
	SearchItemsGlobally(feedID int64, pattern string) ([]database.SearchItemsGloballyRow, error)

	// Read status operations
	MarkItemRead(itemID int64) error
	MarkItemUnread(itemID int64) error
	MarkAllItemsReadInFeed(feedID int64) error

	// Utility methods
	ConvertHTMLToMarkdown(input string) string
	ExtractLinks(content string) []string
	AddLinkMarkersToHTML(content string) (string, []string)

	// Log operations
	GetLogMessages(limit int64) ([]database.LogMessage, error)
	GetLogMessage(id int64) (database.LogMessage, error)
	DeleteAllLogMessages() error

	// Folder operations
	AddFeedFolder(feedID int64, folder string) error
	GetFeedFolders(feedID int64) ([]string, error)
	DeleteFeedFolders(feedID int64) error
	GetFolderStats() ([]database.GetFolderStatsRow, error)

	// Callbacks
	SetRefreshCallbacks(onStart, onComplete func(int64))
}
