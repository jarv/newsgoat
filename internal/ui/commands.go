package ui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
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

func loadURLsList() tea.Cmd {
	return func() tea.Msg {
		urls, err := config.ReadURLsFile()
		if err != nil {
			logging.Error("loadURLsList failed", "error", err)
			return ErrorMsg{Err: err}
		}
		urlsPath, pathErr := config.GetURLsFilePath()
		if pathErr != nil {
			urlsPath = ""
		}
		return URLsListLoadedMsg{URLs: urls, FilePath: urlsPath}
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

