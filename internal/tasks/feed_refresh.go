package tasks

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/logging"
)

// FeedRefreshTaskData represents the data for a feed refresh task
type FeedRefreshTaskData struct {
	FeedID int64  `json:"feed_id"`
	URL    string `json:"url"`
}

// FeedRefreshHandler handles feed refresh tasks
type FeedRefreshHandler struct {
	feedManager *feeds.Manager
}

// NewFeedRefreshHandler creates a new feed refresh handler
func NewFeedRefreshHandler(feedManager *feeds.Manager) *FeedRefreshHandler {
	return &FeedRefreshHandler{
		feedManager: feedManager,
	}
}

// Execute executes a feed refresh task
func (h *FeedRefreshHandler) Execute(ctx context.Context, task *Task) error {
	// Parse task data
	feedIDValue, ok := task.Data["feed_id"]
	if !ok {
		return fmt.Errorf("missing feed_id in task data")
	}

	var feedID int64
	switch v := feedIDValue.(type) {
	case int64:
		feedID = v
	case float64:
		feedID = int64(v)
	case string:
		var err error
		feedID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid feed_id format: %v", v)
		}
	default:
		return fmt.Errorf("invalid feed_id type: %T", v)
	}

	// Perform the feed refresh
	err := h.feedManager.RefreshFeed(feedID)
	if err != nil {
		logging.Error("Feed refresh failed", "feedID", feedID, "error", err)
		return fmt.Errorf("feed refresh failed: %w", err)
	}

	return nil
}

// CanHandle returns true if this handler can handle the given task type
func (h *FeedRefreshHandler) CanHandle(taskType TaskType) bool {
	return taskType == TaskTypeFeedRefresh
}

// CreateFeedRefreshTask creates a new feed refresh task
func CreateFeedRefreshTask(feedID int64, url string) *Task {
	return &Task{
		Type: TaskTypeFeedRefresh,
		Data: map[string]interface{}{
			"feed_id": feedID,
			"url":     url,
		},
	}
}

