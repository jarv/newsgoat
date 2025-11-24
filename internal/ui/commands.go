package ui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/tasks"
	"github.com/jarv/newsgoat/internal/updater"
)

func loadFeedList(feedManager feeds.FeedManager) tea.Cmd {
	return func() tea.Msg {
		feeds, err := feedManager.GetFeedStats()
		if err != nil {
			logging.Error("loadFeedList failed", "error", err)
			return ErrorMsg{Err: err}
		}
		return FeedListLoadedMsg{Feeds: feeds}
	}
}

func loadItemList(feedManager feeds.FeedManager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		items, err := feedManager.GetItemsWithReadStatus(feedID)
		if err != nil {
			logging.Error("loadItemList failed", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return ItemListLoadedMsg{Items: items}
	}
}

func loadLogList(feedManager feeds.FeedManager) tea.Cmd {
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

func clearAllLogMessages(feedManager feeds.FeedManager) tea.Cmd {
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

func refreshFeedAndReload(feedManager feeds.FeedManager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.RefreshFeed(feedID)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		// Return a custom message that will trigger both refresh complete and feed list reload
		return RefreshMsg{FeedID: feedID}
	}
}

func refreshAllFeedsConcurrent(feedManager feeds.FeedManager) tea.Cmd {
	return func() tea.Msg {
		return RefreshAllStartMsg{}
	}
}

func markItemRead(feedManager feeds.FeedManager, itemID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.MarkItemRead(itemID)
		if err != nil {
			logging.Error("Error marking item as read", "itemID", itemID, "error", err)
		}
		return nil
	}
}

func markAllItemsReadInFeed(feedManager feeds.FeedManager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		err := feedManager.MarkAllItemsReadInFeed(feedID)
		if err != nil {
			logging.Error("Error marking all items as read", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return AllItemsMarkedReadMsg{FeedID: feedID}
	}
}

func markAllItemsReadInFolder(feedManager feeds.FeedManager, folderName string) tea.Cmd {
	return func() tea.Msg {
		// Get all feeds in this folder
		allFeeds, err := feedManager.GetAllFeeds()
		if err != nil {
			logging.Error("Error getting feeds for folder", "folder", folderName, "error", err)
			return ErrorMsg{Err: err}
		}

		// Mark all items in each feed as read
		for _, feed := range allFeeds {
			// Check if this feed is in the folder
			folders, err := feedManager.GetFeedFolders(feed.ID)
			if err != nil {
				continue
			}

			// Check if folder matches
			for _, f := range folders {
				if f == folderName {
					// Mark all items in this feed as read
					if err := feedManager.MarkAllItemsReadInFeed(feed.ID); err != nil {
						logging.Error("Error marking feed items as read", "feedID", feed.ID, "error", err)
					}
					break
				}
			}
		}

		// Reload feed list to show updated counts
		return loadFeedList(feedManager)()
	}
}

func toggleItemReadStatus(feedManager feeds.FeedManager, itemID int64, currentlyRead bool) tea.Cmd {
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
			logging.Debug("Task manager already stopped", "error", err)
		}
		return tea.Quit()
	}
}

func loadFeedInfo(feedManager feeds.FeedManager, feedID int64) tea.Cmd {
	return func() tea.Msg {
		feed, err := feedManager.GetFeed(feedID)
		if err != nil {
			logging.Error("loadFeedInfo failed", "feedID", feedID, "error", err)
			return ErrorMsg{Err: err}
		}
		return FeedInfoLoadedMsg{Feed: feed}
	}
}

func openURLsFileInEditor(feedManager feeds.FeedManager) tea.Cmd {
	editor := config.GetEditor()
	if editor == "" {
		return func() tea.Msg {
			return EditorErrorMsg{Err: "EDITOR environment variable is not set"}
		}
	}

	return func() tea.Msg {
		// Get all feeds from database
		allFeeds, err := feedManager.GetAllFeeds()
		if err != nil {
			logging.Error("openURLsFileInEditor: failed to get feeds", "error", err)
			return EditorErrorMsg{Err: "Failed to get feeds from database: " + err.Error()}
		}

		// Build list of feeds with folders for temp file
		var feedsWithFolders []config.FeedWithFolders
		for _, feed := range allFeeds {
			// Only export visible feeds
			if !feed.Visible {
				continue
			}

			// Get folders for this feed
			folders, err := feedManager.GetFeedFolders(feed.ID)
			if err != nil {
				logging.Warn("Failed to get folders for feed", "feedID", feed.ID, "error", err)
				folders = []string{} // Continue with empty folders
			}

			feedsWithFolders = append(feedsWithFolders, config.FeedWithFolders{
				URL:     feed.Url,
				Folders: folders,
			})
		}

		// Create temp file with feeds
		tempPath, err := config.ExportURLsToTempFile(feedsWithFolders)
		if err != nil {
			logging.Error("openURLsFileInEditor: failed to create temp file", "error", err)
			return EditorErrorMsg{Err: "Failed to create temp file: " + err.Error()}
		}

		// Open editor with temp file
		c := exec.Command(editor, tempPath)
		return tea.ExecProcess(c, func(err error) tea.Msg {
			if err != nil {
				logging.Error("openURLsFileInEditor: editor command failed", "editor", editor, "error", err)
				return EditorErrorMsg{Err: "Failed to open editor: " + err.Error()}
			}
			return EditorFinishedMsg{TempFilePath: tempPath}
		})()
	}
}

func addURLAndDiscover(feedManager feeds.FeedManager, input string) tea.Cmd {
	return func() tea.Msg {
		// Parse input: URL followed by optional folders
		// Format: <url> folder1,folder2 or <url> "folder with spaces",folder3
		parts := strings.Fields(input)
		if len(parts) == 0 {
			return URLAddErrorMsg{Err: "No URL provided"}
		}

		urlArg := parts[0]
		var folderStr string
		if len(parts) > 1 {
			// Join remaining parts as folder string
			folderStr = strings.Join(parts[1:], " ")
		}

		// Try to discover the feed URL
		feedURL, err := discovery.DiscoverFeed(urlArg)
		if err != nil {
			return URLAddErrorMsg{Err: "Failed to discover feed: " + err.Error()}
		}

		// Add feed to database without fetching
		if err := feedManager.AddFeedWithoutFetching(feedURL); err != nil {
			// If it already exists, that's okay - we'll still update folders
			logging.Warn("Feed may already exist", "url", feedURL, "error", err)
		}

		// Parse folders if provided
		if folderStr != "" {
			// Parse the folder string to handle quoted folders
			parts := strings.Fields(folderStr)
			if len(parts) > 0 {
				// Join remaining parts to get the full folder spec
				fullSpec := strings.Join(parts, " ")
				// Create a URLEntry to parse folders
				entry := config.URLEntry{
					URL:     feedURL,
					Folders: config.ParseFolders(fullSpec),
				}

				// Get the feed from database
				feed, err := feedManager.GetFeedByURL(feedURL)
				if err != nil {
					logging.Warn("Failed to get feed for folder update", "url", feedURL, "error", err)
				} else {
					// Delete existing folders
					if err := feedManager.DeleteFeedFolders(feed.ID); err != nil {
						logging.Warn("Failed to delete old folders", "feed_id", feed.ID, "error", err)
					}

					// Add new folders
					for _, folder := range entry.Folders {
						if err := feedManager.AddFeedFolder(feed.ID, folder); err != nil {
							logging.Warn("Failed to add folder", "feed_id", feed.ID, "folder", folder, "error", err)
						}
					}
				}
			}
		}

		return URLAddSuccessMsg{URL: feedURL, DiscoveredURL: feedURL != urlArg}
	}
}

func syncFeedsWithURLs(feedManager feeds.FeedManager, urlEntries []config.URLEntry) tea.Cmd {
	return func() tea.Msg {
		// Get all feeds from database
		allFeeds, err := feedManager.GetAllFeeds()
		if err != nil {
			logging.Error("syncFeedsWithURLs: failed to get all feeds", "error", err)
			return ErrorMsg{Err: err}
		}

		// Create a map of URLs from the file for quick lookup
		urlsFromFileSet := make(map[string]config.URLEntry)
		for _, entry := range urlEntries {
			urlsFromFileSet[entry.URL] = entry
		}

		// Create a set of URLs from DB for quick lookup
		urlsFromDBSet := make(map[string]bool)
		for _, feed := range allFeeds {
			urlsFromDBSet[feed.Url] = true
		}

		// Hide feeds that are in DB but not in URLs file
		for _, feed := range allFeeds {
			if _, exists := urlsFromFileSet[feed.Url]; !exists {
				if err := feedManager.HideFeedByURL(feed.Url); err != nil {
					logging.Warn("Failed to hide feed", "url", feed.Url, "error", err)
				}
			}
		}

		// Show/Add feeds that are in URLs file and update folders
		for _, entry := range urlEntries {
			var feedID int64
			if urlsFromDBSet[entry.URL] {
				// Feed exists in DB, make sure it's visible
				if err := feedManager.ShowFeedByURL(entry.URL); err != nil {
					logging.Warn("Failed to show feed", "url", entry.URL, "error", err)
					continue
				}
				// Get feed ID
				feed, err := feedManager.GetFeedByURL(entry.URL)
				if err != nil {
					logging.Warn("Failed to get feed by URL", "url", entry.URL, "error", err)
					continue
				}
				feedID = feed.ID
			} else {
				// Feed doesn't exist, add it without fetching
				if err := feedManager.AddFeedWithoutFetching(entry.URL); err != nil {
					logging.Warn("Failed to add feed", "url", entry.URL, "error", err)
					continue
				}
				// Get the newly created feed ID
				feed, err := feedManager.GetFeedByURL(entry.URL)
				if err != nil {
					logging.Warn("Failed to get newly created feed", "url", entry.URL, "error", err)
					continue
				}
				feedID = feed.ID
			}

			// Update folders for this feed
			// First, delete existing folders
			if err := feedManager.DeleteFeedFolders(feedID); err != nil {
				logging.Warn("Failed to delete old folders", "feed_id", feedID, "error", err)
			}

			// Then add new folders
			for _, folder := range entry.Folders {
				if err := feedManager.AddFeedFolder(feedID, folder); err != nil {
					logging.Warn("Failed to add folder", "feed_id", feedID, "folder", folder, "error", err)
				}
			}
		}

		// Reload feed list after syncing
		return loadFeedList(feedManager)()
	}
}

func syncFeedsFromTempFile(feedManager feeds.FeedManager, tempFilePath string) tea.Cmd {
	return func() tea.Msg {
		// Read URLs from temp file
		urlEntries, err := config.ReadURLsFileFromPath(tempFilePath)

		// Delete temp file after reading (whether successful or not)
		if tempFilePath != "" {
			_ = os.Remove(tempFilePath)
		}

		if err != nil {
			logging.Error("syncFeedsFromTempFile: failed to read temp file", "path", tempFilePath, "error", err)
			return ErrorMsg{Err: err}
		}

		// Sync feeds with the URLs from temp file
		return syncFeedsWithURLs(feedManager, urlEntries)()
	}
}

func performSearch(feedManager feeds.FeedManager, viewState ViewState, feedID int64, searchType SearchType, query string) tea.Cmd {
	return func() tea.Msg {
		// If query is empty, return empty results (will restore unfiltered list)
		if query == "" {
			return SearchResultsMsg{}
		}

		switch viewState {
		case FeedListView:
			// Search feeds
			if searchType == TitleSearch {
				results, err := feedManager.SearchFeedsByTitle(query)
				if err != nil {
					logging.Error("performSearch: SearchFeedsByTitle failed", "query", query, "error", err)
					return ErrorMsg{Err: err}
				}
				return SearchResultsMsg{FeedResults: results, IsGlobal: false}
			} else {
				// Global search
				results, err := feedManager.SearchFeedsGlobally(query)
				if err != nil {
					logging.Error("performSearch: SearchFeedsGlobally failed", "query", query, "error", err)
					return ErrorMsg{Err: err}
				}
				// Convert to []database.SearchFeedsByTitleRow for compatibility
				converted := make([]database.SearchFeedsByTitleRow, len(results))
				for i, r := range results {
					converted[i] = database.SearchFeedsByTitleRow(r)
				}
				return SearchResultsMsg{FeedResults: converted, IsGlobal: true}
			}
		case ItemListView:
			// Search items in current feed
			if searchType == TitleSearch {
				results, err := feedManager.SearchItemsByTitle(feedID, query)
				if err != nil {
					logging.Error("performSearch: SearchItemsByTitle failed", "feedID", feedID, "query", query, "error", err)
					return ErrorMsg{Err: err}
				}
				return SearchResultsMsg{ItemResults: results, IsGlobal: false}
			} else {
				// Global search
				results, err := feedManager.SearchItemsGlobally(feedID, query)
				if err != nil {
					logging.Error("performSearch: SearchItemsGlobally failed", "feedID", feedID, "query", query, "error", err)
					return ErrorMsg{Err: err}
				}
				// Convert to []database.SearchItemsByTitleRow for compatibility
				converted := make([]database.SearchItemsByTitleRow, len(results))
				for i, r := range results {
					converted[i] = database.SearchItemsByTitleRow(r)
				}
				return SearchResultsMsg{ItemResults: converted, IsGlobal: true}
			}
		}

		return SearchResultsMsg{}
	}
}

func checkForUpdate() tea.Cmd {
	return func() tea.Msg {
		updateInfo, err := updater.CheckForUpdate()
		if err != nil {
			logging.Debug("Update check failed", "error", err)
			return UpdateCheckErrorMsg{err: err}
		}
		if updateInfo == nil {
			// No update available
			return nil
		}
		return UpdateAvailableMsg{
			CurrentVersion: updateInfo.CurrentVersion,
			LatestVersion:  updateInfo.LatestVersion,
			DownloadURL:    updateInfo.DownloadURL,
		}
	}
}

func installUpdate(downloadURL string) tea.Cmd {
	return func() tea.Msg {
		err := updater.DownloadAndInstall(downloadURL)
		if err != nil {
			logging.Error("Update installation failed", "error", err)
			return UpdateInstallErrorMsg{err: err}
		}
		return UpdateInstallCompleteMsg{}
	}
}
