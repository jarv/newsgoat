package ui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/tasks"
)

func loadFeedList(feedManager *feeds.Manager) tea.Cmd {
	return func() tea.Msg {
		feeds, err := feedManager.GetFeedStats()
		if err != nil {
			logging.Error("loadFeedList failed", "error", err)
			return ErrorMsg{Err: err}
		}
		return FeedListLoadedMsg{Feeds: feeds}
	}
}

func loadItemList(feedManager *feeds.Manager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		items, err := feedManager.GetItemsWithReadStatus(feedID)
		if err != nil {
			logging.Error("loadItemList failed", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return ItemListLoadedMsg{Items: items}
	}
}

func loadLogList(feedManager *feeds.Manager) tea.Cmd {
	return func() tea.Msg {
		logs, err := feedManager.GetLogMessages(1000) // Get last 1000 log messages
		if err != nil {
			logging.Error("loadLogList failed", "error", err)
			return ErrorMsg{Err: err}
		}
		return LogListLoadedMsg{Logs: logs}
	}
}

func loadTaskList(taskManager tasks.Manager) tea.Cmd {
	return func() tea.Msg {
		// Get all non-completed tasks (pending, running, and failed)
		allTasks, err := taskManager.ListTasks(tasks.TaskFilter{})
		if err != nil {
			logging.Error("loadTaskList failed", "error", err)
			return ErrorMsg{Err: err}
		}

		// Filter out completed tasks
		var filteredTasks []*tasks.Task
		for _, task := range allTasks {
			if task.Status != tasks.TaskStatusCompleted {
				filteredTasks = append(filteredTasks, task)
			}
		}

		return TaskListLoadedMsg{Tasks: filteredTasks}
	}
}

func clearFailedTasks(taskManager tasks.Manager) tea.Cmd {
	return func() tea.Msg {
		err := taskManager.ClearFailedTasks()
		if err != nil {
			logging.Error("clearFailedTasks failed", "error", err)
			return ErrorMsg{Err: err}
		}
		return loadTaskList(taskManager)()
	}
}

func removeTask(taskManager tasks.Manager, taskID string) tea.Cmd {
	return func() tea.Msg {
		err := taskManager.RemoveTask(taskID)
		if err != nil {
			logging.Error("removeTask failed", "taskID", taskID, "error", err)
			return ErrorMsg{Err: err}
		}
		return loadTaskList(taskManager)()
	}
}

func clearAllLogMessages(feedManager *feeds.Manager) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.DeleteAllLogMessages()
		if err != nil {
			logging.Error("clearAllLogMessages failed", "error", err)
			return ErrorMsg{Err: err}
		}
		// Return empty logs list since all were deleted
		return LogListLoadedMsg{Logs: []feeds.LogMessage{}}
	}
}

func refreshFeedAndReload(feedManager *feeds.Manager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.RefreshFeed(feedID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		// Return a custom message that will trigger both refresh complete and feed list reload
		return RefreshMsg{FeedID: feedID}
	}
}

func refreshAllFeedsConcurrent(feedManager *feeds.Manager) tea.Cmd {
	return func() tea.Msg {
		return RefreshAllStartMsg{}
	}
}

func markItemRead(feedManager *feeds.Manager, itemID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.MarkItemRead(itemID)
		if err != nil {
			logging.Error("Error marking item as read", "itemID", itemID, "error", err)
		}
		return nil
	}
}

func markAllItemsReadInFeed(feedManager *feeds.Manager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.MarkAllItemsReadInFeed(feedID)
		if err != nil {
			logging.Error("Error marking all items as read", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return AllItemsMarkedReadMsg{FeedID: feedID}
	}
}

func toggleItemReadStatus(feedManager *feeds.Manager, itemID int64, currentlyRead bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if currentlyRead {
			// If currently read, mark as unread
			err = feedManager.MarkItemUnread(itemID)
		} else {
			// If currently unread, mark as read
			err = feedManager.MarkItemRead(itemID)
		}
		if err != nil {
			logging.Error("Error toggling item read status", "itemID", itemID, "error", err)
			return ErrorMsg{Err: err}
		}
		return ItemReadStatusToggledMsg{ItemID: itemID}
	}
}

func openLink(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			logging.Warn("Unsupported platform for opening links", "platform", runtime.GOOS)
			return nil
		}

		err := cmd.Start()
		if err != nil {
			logging.Error("Error opening link", "url", url, "error", err)
		}

		return nil
	}
}

func spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}

func countdownTick() tea.Cmd {
	return tea.Tick(1*time.Second, func(time.Time) tea.Msg {
		return CountdownTickMsg{}
	})
}

func listenForTaskEvents(taskManager tasks.Manager) tea.Cmd {
	return func() tea.Msg {
		events := taskManager.Subscribe()
		event := <-events
		return TaskEventMsg{Event: event}
	}
}

func waitForReloadTimer(minutes int) tea.Cmd {
	return tea.Tick(time.Duration(minutes)*time.Minute, func(time.Time) tea.Msg {
		return ReloadTimerMsg{}
	})
}

func restartReloadTimer() tea.Cmd {
	return func() tea.Msg {
		return RestartReloadTimerMsg{}
	}
}

func quitApp(taskManager tasks.Manager) tea.Cmd {
	return func() tea.Msg {
		// Stop task manager to cancel all in-progress tasks
		if err := taskManager.Stop(); err != nil {
			logging.Error("Failed to stop task manager on quit", "error", err)
		}
		return tea.Quit()
	}
}

func loadFeedInfo(queries *database.Queries, feedID int64) tea.Cmd {
	return func() tea.Msg {
		feed, err := queries.GetFeed(context.Background(), feedID)
		if err != nil {
			logging.Error("loadFeedInfo failed", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return FeedInfoLoadedMsg{Feed: feed}
	}
}

func reloadURLsFromFile(feedManager *feeds.Manager) tea.Cmd {
	return func() tea.Msg {
		urls, err := config.ReadURLsFile()
		if err != nil {
			logging.Error("reloadURLsFromFile failed", "error", err)
			return ErrorMsg{Err: err}
		}
		urlsPath, pathErr := config.GetURLsFilePath()
		if pathErr != nil {
			urlsPath = ""
		}
		return URLsReloadedMsg{URLs: urls, FilePath: urlsPath}
	}
}

func openURLsFileInEditor() tea.Cmd {
	editor := config.GetEditor()
	if editor == "" {
		return func() tea.Msg {
			return EditorErrorMsg{Err: "EDITOR environment variable is not set"}
		}
	}

	urlsPath, err := config.GetURLsFilePath()
	if err != nil {
		logging.Error("openURLsFileInEditor: failed to get URLs file path", "error", err)
		return func() tea.Msg {
			return EditorErrorMsg{Err: "Failed to get URLs file path: " + err.Error()}
		}
	}

	c := exec.Command(editor, urlsPath)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			logging.Error("openURLsFileInEditor: editor command failed", "editor", editor, "error", err)
			return EditorErrorMsg{Err: "Failed to open editor: " + err.Error()}
		}
		return EditorFinishedMsg{}
	})
}

func addURLAndDiscover(feedManager *feeds.Manager, urlArg string) tea.Cmd {
	return func() tea.Msg {
		// Try to discover the feed URL
		feedURL, err := discovery.DiscoverFeed(urlArg)
		if err != nil {
			return URLAddErrorMsg{Err: "Failed to discover feed: " + err.Error()}
		}

		// Add the URL to the URLs file
		if err := config.AddURL(feedURL); err != nil {
			return URLAddErrorMsg{Err: "Failed to add URL to file: " + err.Error()}
		}

		// Add feed to database without fetching
		if err := feedManager.AddFeedWithoutFetching(feedURL); err != nil {
			// If it already exists, that's okay
			logging.Warn("Feed may already exist", "url", feedURL, "error", err)
		}

		return URLAddSuccessMsg{URL: feedURL, DiscoveredURL: feedURL != urlArg}
	}
}

func syncFeedsWithURLs(feedManager *feeds.Manager, urls []string) tea.Cmd {
	return func() tea.Msg {
		// Get all feeds from database
		allFeeds, err := feedManager.GetAllFeeds()
		if err != nil {
			logging.Error("syncFeedsWithURLs: failed to get all feeds", "error", err)
			return ErrorMsg{Err: err}
		}

		// Create a set of URLs from the file for quick lookup
		urlsFromFileSet := make(map[string]bool)
		for _, url := range urls {
			urlsFromFileSet[url] = true
		}

		// Create a set of URLs from DB for quick lookup
		urlsFromDBSet := make(map[string]bool)
		for _, feed := range allFeeds {
			urlsFromDBSet[feed.Url] = true
		}

		// Hide feeds that are in DB but not in URLs file
		for _, feed := range allFeeds {
			if !urlsFromFileSet[feed.Url] {
				if err := feedManager.HideFeedByURL(feed.Url); err != nil {
					logging.Warn("Failed to hide feed", "url", feed.Url, "error", err)
				}
			}
		}

		// Show/Add feeds that are in URLs file
		for _, url := range urls {
			if urlsFromDBSet[url] {
				// Feed exists in DB, make sure it's visible
				if err := feedManager.ShowFeedByURL(url); err != nil {
					logging.Warn("Failed to show feed", "url", url, "error", err)
				}
			} else {
				// Feed doesn't exist, add it without fetching
				if err := feedManager.AddFeedWithoutFetching(url); err != nil {
					logging.Warn("Failed to add feed", "url", url, "error", err)
				}
			}
		}

		// Reload feed list after syncing
		return loadFeedList(feedManager)()
	}
}

