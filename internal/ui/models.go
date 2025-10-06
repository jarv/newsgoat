package ui

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/tasks"
	"github.com/jarv/newsgoat/internal/themes"
	"github.com/jarv/newsgoat/internal/version"
)

const globalHelp string = "h: help"

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func formatNullTime(nt sql.NullTime) string {
	if !nt.Valid {
		return "(not set)"
	}
	return nt.Time.Format("2006-01-02 15:04:05")
}

func formatNullString(ns sql.NullString) string {
	if !ns.Valid || ns.String == "" {
		return "(not set)"
	}
	return ns.String
}

func formatNullInt64(ni sql.NullInt64) string {
	if !ni.Valid {
		return "(not set)"
	}
	return fmt.Sprintf("%d", ni.Int64)
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := ""
	for _, word := range words {
		// If adding this word would exceed width, start a new line
		testLine := currentLine
		if currentLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > width && currentLine != "" {
			// Current line is full, save it and start new line
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// getDisplayTitle returns the display title for a feed, overriding for GitHub/GitLab
func getDisplayTitle(feed database.GetFeedStatsRow) string {
	switch discovery.GetURLType(feed.Url) {
	case discovery.URLTypeGitHub, discovery.URLTypeGitLab:
		if strings.Contains(feed.Url, "commits") {
			// Remove https:// and .atom from the URL for display
			displayTitle := strings.TrimPrefix(feed.Url, "https://")
			return strings.TrimSuffix(displayTitle, ".atom")
		}
		return feed.Title
	default:
		return feed.Title
	}
}

// filterFeedsBySearch filters the feed list based on the search query
func (m *Model) filterFeedsBySearch() {
	if m.searchQuery == "" {
		// Empty search shows all feeds
		m.feedList = m.unfilteredFeedList
		return
	}

	// Filter feeds by display title (case-insensitive)
	lowerQuery := strings.ToLower(m.searchQuery)
	var filtered []database.GetFeedStatsRow
	for _, feed := range m.unfilteredFeedList {
		displayTitle := getDisplayTitle(feed)
		if strings.Contains(strings.ToLower(displayTitle), lowerQuery) {
			filtered = append(filtered, feed)
		}
	}
	m.feedList = filtered
}

type ViewState int

const (
	FeedListView ViewState = iota
	ItemListView
	ArticleView
	FeedInfoView
	LogView
	LogDetailView
	TasksView
	HelpView
	SettingsView
	URLsView
)

type Model struct {
	feedManager                     *feeds.Manager
	taskManager                     tasks.Manager
	queries                         *database.Queries
	config                          config.Config
	glamourRenderer                 *glamour.TermRenderer
	state                           ViewState
	previousState                   ViewState // Store previous state when entering help view
	feedList                        []database.GetFeedStatsRow
	allFeeds                        []database.GetFeedStatsRow // Unfiltered list of all feeds (for reload operations)
	totalFeedCount                  int                        // Total number of feeds in database (before filtering)
	itemList                        []database.GetItemsWithReadStatusRow
	currentItem                     database.GetItemsWithReadStatusRow
	currentFeed                     database.Feed // For feed info view
	logList                         []database.LogMessage
	currentLog                      database.LogMessage
	taskList                        []*tasks.Task
	urlsList                        []string
	urlsFilePath                    string
	links                           []string
	cursor                          int
	savedItemCursor                 int
	savedFeedCursor                 int
	savedLogCursor                  int
	savedTasksCursor                int
	savedSettingsCursor             int
	helpViewScroll                  int // Scroll offset for help view
	articleViewScroll               int // Scroll offset for article view
	urlsViewScroll                  int // Scroll offset for URLs view
	selectedFeed                    int64
	width                           int
	height                          int
	err                             error
	refreshing                      bool
	refreshStatus                   string
	refreshingFeeds                 map[int64]bool             // Track which feeds are currently refreshing
	pendingFeeds                    []int64                    // Feeds waiting to be refreshed (for refresh-all)
	maxConcurrency                  int                        // Max concurrent refreshes allowed
	spinnerFrame                    int                        // Current spinner animation frame
	spinnerRunning                  bool                       // Track if spinner timer is already running
	firstAutoReload                 bool                       // Track if this is the first auto reload (for SuppressFirstReload)
	pendingStartupReload            bool                       // Track if we need to reload on startup after feed list loads
	nextReloadTime                  time.Time                  // Time when next auto reload is scheduled
	editingSettings                 bool                       // Track if we're editing a setting
	selectingTheme                  bool                       // Track if we're selecting a theme
	selectingHighlight              bool                       // Track if we're selecting a highlight style
	selectingSpinner                bool                       // Track if we're selecting a spinner type
	selectingShowReadFeeds          bool                       // Track if we're selecting show read feeds
	selectingAutoReload             bool                       // Track if we're selecting auto reload
	selectingSuppressFirstReload    bool                       // Track if we're selecting suppress first reload
	selectingReloadOnStartup        bool                       // Track if we're selecting reload on startup
	selectingUnreadOnTop            bool                       // Track if we're selecting unread on top
	showRawHTML                     bool                       // Track if showing raw HTML in article view
	themeSelectCursor               int                        // Cursor position in theme selector
	highlightSelectCursor           int                        // Cursor position in highlight style selector
	spinnerSelectCursor             int                        // Cursor position in spinner type selector
	showReadFeedsSelectCursor       int                        // Cursor position in show read feeds selector
	autoReloadSelectCursor          int                        // Cursor position in auto reload selector
	suppressFirstReloadSelectCursor int                        // Cursor position in suppress first reload selector
	reloadOnStartupSelectCursor     int                        // Cursor position in reload on startup selector
	unreadOnTopSelectCursor         int                        // Cursor position in unread on top selector
	settingInput                    string                     // Current input value when editing
	showSettingsHelp                bool                       // Track if we're showing settings help
	searchMode                      bool                       // Track if search mode is active
	searchQuery                     string                     // Current search query text
	searchActive                    bool                       // Track if feeds are currently filtered by search
	unfilteredFeedList              []database.GetFeedStatsRow // Feed list before search filtering
	statusMessage                   string                     // Message to display above status bar
	statusMessageType               string                     // Type of message: "error" or "info"
	quitPressed                     bool                       // Track if 'q' was pressed once (for quit confirmation)
	ctrlCPressed                    bool                       // Track if 'ctrl+c' was pressed once (for quit confirmation)
	addingURL                       bool                       // Track if in URL adding mode
	urlInput                        string                     // Current URL input text
}

type RefreshMsg struct {
	FeedID int64
}

type RefreshAllMsg struct{}

type RefreshAllStartMsg struct{}

type RefreshAllCompleteMsg struct{}

type RefreshStartMsg struct {
	Status string
}

type RefreshCompleteMsg struct{}

type FeedRefreshStartMsg struct {
	FeedID int64
}

type FeedRefreshCompleteMsg struct {
	FeedID int64
}

type SpinnerTickMsg struct{}

type TaskEventMsg struct {
	Event tasks.TaskEvent
}

type FeedListLoadedMsg struct {
	Feeds []database.GetFeedStatsRow
}

type ItemListLoadedMsg struct {
	Items []database.GetItemsWithReadStatusRow
}

type ErrorMsg struct {
	Err error
}

type LogListLoadedMsg struct {
	Logs []database.LogMessage
}

type TaskListLoadedMsg struct {
	Tasks []*tasks.Task
}

type URLsListLoadedMsg struct {
	URLs     []string
	FilePath string
}

type URLsReloadedMsg struct {
	URLs     []string
	FilePath string
}

type EditorFinishedMsg struct{}

type EditorErrorMsg struct {
	Err string
}

type FeedInfoLoadedMsg struct {
	Feed database.Feed
}

type AllItemsMarkedReadMsg struct {
	FeedID int64
}

type ItemReadStatusToggledMsg struct {
	ItemID int64
}

type URLAddSuccessMsg struct {
	URL           string
	DiscoveredURL bool
}

type URLAddErrorMsg struct {
	Err string
}

type ReloadTimerMsg struct{}

type RestartReloadTimerMsg struct{}

type CountdownTickMsg struct{}

func NewModel(feedManager *feeds.Manager, taskManager tasks.Manager, queries *database.Queries, cfg config.Config) Model {
	// Create glamour renderer based on theme
	theme := themes.GetThemeByName(cfg.ThemeName)
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(theme.GlamourStyle),
		glamour.WithWordWrap(80),
	)

	if err != nil {
		// Fallback to default renderer if creation fails
		renderer, _ = glamour.NewTermRenderer()
	}

	return Model{
		feedManager:          feedManager,
		taskManager:          taskManager,
		queries:              queries,
		config:               cfg,
		glamourRenderer:      renderer,
		state:                FeedListView,
		cursor:               0,
		savedItemCursor:      0,
		savedFeedCursor:      0,
		savedLogCursor:       0,
		savedTasksCursor:     0,
		savedSettingsCursor:  0,
		refreshingFeeds:      make(map[int64]bool),
		pendingFeeds:         []int64{},
		maxConcurrency:       cfg.ReloadConcurrency,
		spinnerFrame:         0,
		spinnerRunning:       false,
		firstAutoReload:      true,                // First reload should be suppressed if configured
		pendingStartupReload: cfg.ReloadOnStartup, // Will trigger reload after feed list loads
	}
}

func (m *Model) SetURLsFilePath(path string) {
	m.urlsFilePath = path
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds,
		loadFeedList(m.feedManager),
		tea.WindowSize(),
		listenForTaskEvents(m.taskManager),
	)

	// Start the reload timer if auto reload is enabled
	if m.config.AutoReload && m.config.ReloadTime > 0 {
		// Note: nextReloadTime will be set in Update() when ReloadTimerMsg is processed
		cmds = append(cmds, waitForReloadTimer(m.config.ReloadTime))
		cmds = append(cmds, countdownTick())
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle paste events for URL input and search
		if msg.Paste {
			if m.addingURL {
				m.urlInput += string(msg.Runes)
				return m, nil
			} else if m.searchMode {
				m.searchQuery += string(msg.Runes)
				m.filterFeedsBySearch()
				m.cursor = 0
				m.savedFeedCursor = 0
				return m, nil
			}
		}
		return m.handleKeyPress(msg)

	case FeedListLoadedMsg:
		// Store unfiltered feeds for reload operations
		m.allFeeds = msg.Feeds
		m.totalFeedCount = len(msg.Feeds)

		// Filter feeds based on ShowReadFeeds config
		if m.config.ShowReadFeeds {
			m.feedList = msg.Feeds
		} else {
			// Filter out feeds with no unread items
			var filteredFeeds []database.GetFeedStatsRow
			for _, feed := range msg.Feeds {
				if feed.UnreadItems > 0 {
					filteredFeeds = append(filteredFeeds, feed)
				}
			}
			m.feedList = filteredFeeds
		}

		// Sort feeds if UnreadOnTop is enabled
		if m.config.UnreadOnTop {
			sort.SliceStable(m.feedList, func(i, j int) bool {
				// Feeds with unread items come first
				iHasUnread := m.feedList[i].UnreadItems > 0
				jHasUnread := m.feedList[j].UnreadItems > 0
				if iHasUnread != jHasUnread {
					return iHasUnread
				}
				// Within each group, maintain original order (stable sort)
				return false
			})
		}

		if m.state == FeedListView {
			// Preserve cursor position when refreshing feed list
			m.cursor = m.savedFeedCursor
			if m.cursor >= len(m.feedList) {
				m.cursor = max(0, len(m.feedList)-1)
			}
			m.savedFeedCursor = m.cursor
		}
		// Note: if not in FeedListView, don't modify cursor or savedFeedCursor
		// They will be set appropriately when we transition back to FeedListView

		// Trigger reload on startup if configured and this is the first load
		if m.pendingStartupReload && len(m.allFeeds) > 0 {
			m.pendingStartupReload = false
			return m, func() tea.Msg { return ReloadTimerMsg{} }
		}

		return m, nil

	case ItemListLoadedMsg:
		m.itemList = msg.Items
		if m.state == ItemListView {
			// Preserve cursor position when refreshing
			m.cursor = m.savedItemCursor
			if m.cursor >= len(m.itemList) {
				m.cursor = max(0, len(m.itemList)-1)
			}
			m.savedItemCursor = m.cursor
		} else {
			// First time entering, start at top
			m.cursor = 0
			m.savedItemCursor = 0
		}
		return m, nil

	case LogListLoadedMsg:
		m.logList = msg.Logs
		if m.state == LogView {
			// Preserve cursor position when refreshing
			m.cursor = m.savedLogCursor
			if m.cursor >= len(m.logList) {
				m.cursor = max(0, len(m.logList)-1)
			}
			m.savedLogCursor = m.cursor
		} else {
			// First time entering, start at top
			m.cursor = 0
			m.savedLogCursor = 0
		}
		return m, nil

	case TaskListLoadedMsg:
		m.taskList = msg.Tasks
		if m.state == TasksView {
			// Preserve cursor position when refreshing
			m.cursor = m.savedTasksCursor
			if m.cursor >= len(m.taskList) {
				m.cursor = max(0, len(m.taskList)-1)
			}
			m.savedTasksCursor = m.cursor
		} else {
			// First time entering, start at top
			m.cursor = 0
			m.savedTasksCursor = 0
		}
		return m, nil

	case URLsListLoadedMsg:
		m.urlsList = msg.URLs
		m.urlsFilePath = msg.FilePath
		return m, nil

	case URLsReloadedMsg:
		m.urlsList = msg.URLs
		// Set info message
		m.statusMessage = "urls reloaded from " + msg.FilePath
		m.statusMessageType = "info"
		// Sync feeds with the reloaded URLs
		return m, syncFeedsWithURLs(m.feedManager, msg.URLs)

	case EditorFinishedMsg:
		// After editor closes, reload URLs and sync feeds
		return m, reloadURLsFromFile(m.feedManager)

	case EditorErrorMsg:
		// Display error message
		m.err = fmt.Errorf("%s", msg.Err)
		return m, nil

	case FeedInfoLoadedMsg:
		m.currentFeed = msg.Feed
		m.previousState = m.state
		m.state = FeedInfoView
		return m, nil

	case RefreshStartMsg:
		m.refreshing = true
		m.refreshStatus = msg.Status
		return m, nil

	case RefreshMsg:
		// This means refresh is complete and we need to reload data
		cmd := loadFeedList(m.feedManager)
		if m.state == ItemListView {
			cmd = tea.Batch(
				loadFeedList(m.feedManager),
				loadItemList(m.feedManager, m.selectedFeed),
			)
		}
		return m, tea.Batch(
			cmd,
			func() tea.Msg { return RefreshCompleteMsg{} },
			func() tea.Msg {
				return FeedRefreshCompleteMsg(msg)
			},
		)

	case RefreshAllMsg:
		return m, tea.Batch(
			refreshAllFeedsConcurrent(m.feedManager),
			loadFeedList(m.feedManager),
			func() tea.Msg { return RefreshCompleteMsg{} },
		)

	case RefreshCompleteMsg:
		m.refreshing = false
		m.refreshStatus = ""
		// Clear all refreshing feeds
		m.refreshingFeeds = make(map[int64]bool)
		// Stop spinner
		m.spinnerRunning = false
		return m, nil

	case RefreshAllCompleteMsg:
		// Send FeedRefreshCompleteMsg for all feeds that were being refreshed
		var cmds []tea.Cmd
		for feedID := range m.refreshingFeeds {
			feedID := feedID // capture loop variable
			cmds = append(cmds, func() tea.Msg {
				return FeedRefreshCompleteMsg{FeedID: feedID}
			})
		}
		cmds = append(cmds, loadFeedList(m.feedManager))
		return m, tea.Batch(cmds...)

	case RefreshAllStartMsg:
		// Get all feeds and set up for controlled concurrency refresh
		feeds, err := m.feedManager.GetFeedStats()
		if err != nil {
			return m, func() tea.Msg { return ErrorMsg{Err: err} }
		}

		// Initialize pending feeds queue
		m.pendingFeeds = make([]int64, len(feeds))
		for i, feed := range feeds {
			m.pendingFeeds[i] = feed.ID
		}

		// Start initial batch of feeds (up to maxConcurrency)
		return m, m.startNextBatchOfFeeds()

	case FeedRefreshStartMsg:
		m.refreshingFeeds[msg.FeedID] = true
		// Start spinner animation if we have refreshing feeds and spinner is not already running
		if len(m.refreshingFeeds) > 0 && !m.spinnerRunning {
			m.spinnerRunning = true
			return m, spinnerTick()
		}
		return m, nil

	case FeedRefreshCompleteMsg:
		delete(m.refreshingFeeds, msg.FeedID)

		// If we have more pending feeds, start the next one
		cmd := loadFeedList(m.feedManager)
		if len(m.pendingFeeds) > 0 {
			cmd = tea.Batch(cmd, m.startNextBatchOfFeeds())
		} else if len(m.refreshingFeeds) == 0 {
			// No more refreshing feeds and no pending feeds - refresh all is complete
			cmd = tea.Batch(cmd, func() tea.Msg { return RefreshCompleteMsg{} })
		}

		return m, cmd

	case SpinnerTickMsg:
		// Only continue spinner if we have refreshing feeds
		if len(m.refreshingFeeds) > 0 {
			spinnerFrames := themes.GetSpinnerFrames(m.config.SpinnerType)
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, spinnerTick()
		}
		// No more refreshing feeds, stop the spinner
		m.spinnerRunning = false
		return m, nil

	case ReloadTimerMsg:
		// Check if we should suppress the first reload
		if m.firstAutoReload && m.config.SuppressFirstReload {
			// Skip this reload but mark that we've passed the first one
			m.firstAutoReload = false
		} else {
			// Automatic reload triggered
			if !m.refreshing && len(m.allFeeds) > 0 {
				m.refreshing = true
				m.refreshStatus = "Auto-refreshing all feeds..."

				// Create tasks for all feeds (use allFeeds to include filtered feeds)
				for _, feed := range m.allFeeds {
					task := tasks.CreateFeedRefreshTask(feed.ID, feed.Url)
					if err := m.taskManager.AddTask(task); err != nil {
						continue
					}
				}

				m.firstAutoReload = false
			}
		}

		// Restart the timer for the next reload if auto reload is enabled
		var cmds []tea.Cmd
		if !m.refreshing || m.config.SuppressFirstReload {
			// Only send RefreshStartMsg if we actually started a refresh
			if m.refreshing {
				cmds = append(cmds, func() tea.Msg { return RefreshStartMsg{Status: "Auto-refreshing all feeds..."} })
			}
		}
		// Restart timer only if auto reload is enabled
		if m.config.AutoReload && m.config.ReloadTime > 0 {
			m.nextReloadTime = time.Now().Add(time.Duration(m.config.ReloadTime) * time.Minute)
			cmds = append(cmds, waitForReloadTimer(m.config.ReloadTime))
		}
		return m, tea.Batch(cmds...)

	case RestartReloadTimerMsg:
		// Restart the timer (triggered when config changes)
		if m.config.AutoReload && m.config.ReloadTime > 0 {
			m.nextReloadTime = time.Now().Add(time.Duration(m.config.ReloadTime) * time.Minute)
			return m, tea.Batch(waitForReloadTimer(m.config.ReloadTime), countdownTick())
		}
		// Clear next reload time if auto reload is disabled
		m.nextReloadTime = time.Time{}
		return m, nil

	case CountdownTickMsg:
		// Continue countdown ticker if auto reload is enabled
		if m.config.AutoReload && m.config.ReloadTime > 0 {
			return m, countdownTick()
		}
		return m, nil

	case TaskEventMsg:
		event := msg.Event

		// Handle task events based on type
		switch event.Type {
		case tasks.TaskEventStarted:
			// Mark feed as refreshing if it's a feed refresh task
			if event.TaskType == tasks.TaskTypeFeedRefresh {
				if feedIDValue, ok := event.Data["feed_id"]; ok {
					var feedID int64
					switch v := feedIDValue.(type) {
					case int64:
						feedID = v
					case float64:
						feedID = int64(v)
					}

					if feedID > 0 {
						m.refreshingFeeds[feedID] = true
						// Start spinner if not already running
						if !m.spinnerRunning {
							m.spinnerRunning = true
							var cmds []tea.Cmd
							cmds = append(cmds, listenForTaskEvents(m.taskManager))
							cmds = append(cmds, spinnerTick())
							// Refresh task list if we're viewing it
							if m.state == TasksView {
								cmds = append(cmds, loadTaskList(m.taskManager))
							}
							return m, tea.Batch(cmds...)
						}
					}
				}
			}

			// Refresh task list if we're viewing it
			if m.state == TasksView {
				return m, tea.Batch(
					listenForTaskEvents(m.taskManager),
					loadTaskList(m.taskManager),
				)
			}

		case tasks.TaskEventCompleted, tasks.TaskEventFailed:
			// Mark feed as no longer refreshing
			if event.TaskType == tasks.TaskTypeFeedRefresh {
				if feedIDValue, ok := event.Data["feed_id"]; ok {
					var feedID int64
					switch v := feedIDValue.(type) {
					case int64:
						feedID = v
					case float64:
						feedID = int64(v)
					}

					if feedID > 0 {
						delete(m.refreshingFeeds, feedID)

						var cmds []tea.Cmd
						cmds = append(cmds, listenForTaskEvents(m.taskManager))
						cmds = append(cmds, loadFeedList(m.feedManager))

						// Refresh task list if we're viewing it
						if m.state == TasksView {
							cmds = append(cmds, loadTaskList(m.taskManager))
						}

						// Check if all refreshes are complete
						if len(m.refreshingFeeds) == 0 && m.refreshing {
							cmds = append(cmds, func() tea.Msg { return RefreshCompleteMsg{} })
						}

						return m, tea.Batch(cmds...)
					}
				}
			}

			// Refresh task list if we're viewing it (for non-feed-refresh tasks)
			if m.state == TasksView {
				return m, tea.Batch(
					listenForTaskEvents(m.taskManager),
					loadTaskList(m.taskManager),
				)
			}
		}

		// Continue listening for task events
		return m, listenForTaskEvents(m.taskManager)

	case AllItemsMarkedReadMsg:
		// Items were marked as read, reload the appropriate lists
		var cmds []tea.Cmd
		cmds = append(cmds, loadFeedList(m.feedManager))

		// If we're in the item list view for this feed, reload it too
		if m.state == ItemListView && m.selectedFeed == msg.FeedID {
			cmds = append(cmds, loadItemList(m.feedManager, msg.FeedID))
		}

		return m, tea.Batch(cmds...)

	case ItemReadStatusToggledMsg:
		// Item read status was toggled, reload the item list and feed list
		var cmds []tea.Cmd
		cmds = append(cmds, loadFeedList(m.feedManager))
		if m.state == ItemListView {
			cmds = append(cmds, loadItemList(m.feedManager, m.selectedFeed))
		}
		return m, tea.Batch(cmds...)

	case URLAddSuccessMsg:
		// Set success message
		if msg.DiscoveredURL {
			m.statusMessage = "Added feed: " + msg.URL + " (discovered)"
		} else {
			m.statusMessage = "Added feed: " + msg.URL
		}
		m.statusMessageType = "info"
		// Reload feed list and sync feeds
		return m, tea.Batch(loadFeedList(m.feedManager), reloadURLsFromFile(m.feedManager))

	case URLAddErrorMsg:
		// Set error message
		m.statusMessage = msg.Err
		m.statusMessageType = "error"
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.refreshing = false
		m.refreshStatus = ""
		m.refreshingFeeds = make(map[int64]bool)
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case FeedListView:
		return m.handleFeedListKeys(msg)
	case ItemListView:
		return m.handleItemListKeys(msg)
	case ArticleView:
		return m.handleArticleKeys(msg)
	case FeedInfoView:
		return m.handleFeedInfoKeys(msg)
	case LogView:
		return m.handleLogListKeys(msg)
	case LogDetailView:
		return m.handleLogDetailKeys(msg)
	case TasksView:
		return m.handleTasksViewKeys(msg)
	case HelpView:
		return m.handleHelpViewKeys(msg)
	case SettingsView:
		return m.handleSettingsViewKeys(msg)
	case URLsView:
		return m.handleURLsViewKeys(msg)
	}
	return m, nil
}

func (m Model) handleFeedListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Clear status message and quit state on any keypress (except 'q' and 'ctrl+c' themselves)
	key := msg.String()
	if key != "q" && key != "ctrl+c" {
		if m.statusMessage != "" {
			m.statusMessage = ""
			m.statusMessageType = ""
		}
		m.quitPressed = false
		m.ctrlCPressed = false
	}

	// Handle URL adding mode separately
	if m.addingURL {
		switch msg.String() {
		case "esc", "ctrl+c":
			// Cancel URL adding
			m.addingURL = false
			m.urlInput = ""
			return m, nil
		case "enter":
			// Submit URL
			if m.urlInput != "" {
				url := m.urlInput
				m.addingURL = false
				m.urlInput = ""
				return m, addURLAndDiscover(m.feedManager, url)
			}
			// Empty input, just cancel
			m.addingURL = false
			return m, nil
		case "backspace":
			// Remove last character
			if len(m.urlInput) > 0 {
				m.urlInput = m.urlInput[:len(m.urlInput)-1]
			}
			return m, nil
		default:
			// Add character to URL input if it's a single character
			key := msg.String()
			if len(key) == 1 {
				m.urlInput += key
			}
			return m, nil
		}
	}

	// Handle search mode separately
	if m.searchMode {
		switch msg.String() {
		case "esc", "ctrl+c":
			// Cancel search and restore original feed list
			m.searchMode = false
			m.searchActive = false
			m.feedList = m.unfilteredFeedList
			m.searchQuery = ""
			m.cursor = 0
			m.savedFeedCursor = 0
			return m, nil
		case "enter":
			// Accept search and exit search mode (if query is empty, also clear search)
			if m.searchQuery == "" {
				m.searchMode = false
				m.searchActive = false
				m.feedList = m.unfilteredFeedList
			} else {
				m.searchMode = false
				m.searchActive = true // Mark that feeds are filtered by search
			}
			m.searchQuery = ""
			return m, nil
		case "backspace":
			// Remove last character from search query
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.filterFeedsBySearch()
				m.cursor = 0
				m.savedFeedCursor = 0
			} else {
				// If query is empty, clear search and restore original feed list
				m.searchMode = false
				m.searchActive = false
				m.feedList = m.unfilteredFeedList
				m.cursor = 0
				m.savedFeedCursor = 0
			}
			return m, nil
		default:
			// Add character to search query if it's a single character
			key := msg.String()
			if len(key) == 1 {
				m.searchQuery += key
				m.filterFeedsBySearch()
				m.cursor = 0
				m.savedFeedCursor = 0
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		// If search is active (feeds are filtered), clear the search
		if m.searchActive {
			m.searchActive = false
			m.feedList = m.unfilteredFeedList
			m.cursor = 0
			m.savedFeedCursor = 0
			return m, nil
		}
		// Otherwise, do nothing (don't quit in feed list view)
		return m, nil

	case "q":
		// Quit confirmation: show message on first press, quit on second
		if m.quitPressed {
			return m, quitApp(m.taskManager)
		}
		m.quitPressed = true
		m.statusMessage = "press q again to quit"
		m.statusMessageType = "info"
		return m, nil

	case "ctrl+c":
		// Quit confirmation: show message on first press, quit on second
		if m.ctrlCPressed {
			return m, quitApp(m.taskManager)
		}
		m.ctrlCPressed = true
		m.statusMessage = "press ctrl+c again to quit"
		m.statusMessageType = "info"
		return m, nil

	case "ctrl+r":
		// Reload URLs from file and sync with feeds
		return m, reloadURLsFromFile(m.feedManager)

	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "j", "down":
		if len(m.feedList) > 0 && m.cursor < len(m.feedList)-1 {
			m.cursor++
			m.savedFeedCursor = m.cursor
		}

	case "k", "up":
		if len(m.feedList) > 0 && m.cursor > 0 {
			m.cursor--
			m.savedFeedCursor = m.cursor
		}

	case "ctrl+d":
		if len(m.feedList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = min(m.cursor+pageSize, len(m.feedList)-1)
			m.savedFeedCursor = m.cursor
		}

	case "ctrl+u":
		if len(m.feedList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = max(m.cursor-pageSize, 0)
			m.savedFeedCursor = m.cursor
		}

	case "enter":
		if len(m.feedList) > 0 && m.cursor < len(m.feedList) {
			// Clear search mode and filter when entering item list
			m.searchMode = false
			m.searchActive = false
			m.searchQuery = ""
			m.selectedFeed = m.feedList[m.cursor].ID
			m.state = ItemListView
			m.cursor = 0
			m.savedItemCursor = 0
			return m, loadItemList(m.feedManager, m.selectedFeed)
		}

	case "R":
		if !m.refreshing {
			m.refreshing = true
			m.refreshStatus = "Refreshing all feeds..."

			// Create tasks for all feeds (use allFeeds to include filtered feeds)
			for _, feed := range m.allFeeds {
				task := tasks.CreateFeedRefreshTask(feed.ID, feed.Url)
				if err := m.taskManager.AddTask(task); err != nil {
					// If task creation fails, log it but continue with other feeds
					continue
				}
			}

			return m, func() tea.Msg { return RefreshStartMsg{Status: "Refreshing all feeds..."} }
		}

	case "r":
		if !m.refreshing && len(m.feedList) > 0 && m.cursor < len(m.feedList) {
			m.refreshing = true
			m.refreshStatus = "Refreshing feed..."

			feed := m.feedList[m.cursor]
			task := tasks.CreateFeedRefreshTask(feed.ID, feed.Url)
			if err := m.taskManager.AddTask(task); err != nil {
				// Handle error, maybe show error message
				m.refreshing = false
				m.refreshStatus = ""
				return m, nil
			}

			return m, func() tea.Msg { return RefreshStartMsg{Status: "Refreshing feed..."} }
		}

	case "l":
		m.state = LogView
		m.cursor = 0
		m.savedLogCursor = 0
		return m, loadLogList(m.feedManager)

	case "t":
		m.state = TasksView
		m.cursor = 0
		m.savedTasksCursor = 0
		return m, loadTaskList(m.taskManager)

	case "c":
		m.state = SettingsView
		m.cursor = 0
		m.savedSettingsCursor = 0
		return m, nil

	case "u":
		// Enter URL adding mode
		m.addingURL = true
		m.urlInput = ""
		return m, nil

	case "U":
		// Check if EDITOR is set
		if config.GetEditor() == "" {
			m.statusMessage = "Set EDITOR in your env to edit urls"
			m.statusMessageType = "error"
			return m, nil
		}
		// Open URLs file in editor
		return m, openURLsFileInEditor()

	case "i":
		// Show feed info
		if len(m.feedList) > 0 && m.cursor < len(m.feedList) {
			feedID := m.feedList[m.cursor].ID
			return m, loadFeedInfo(m.queries, feedID)
		}

	case "A":
		// Mark all items in the highlighted feed as read
		if len(m.feedList) > 0 && m.cursor < len(m.feedList) {
			feedID := m.feedList[m.cursor].ID
			return m, markAllItemsReadInFeed(m.feedManager, feedID)
		}

	case "/":
		// Enter search mode
		m.searchMode = true
		m.searchQuery = ""
		// Save current feed list to restore on cancel
		m.unfilteredFeedList = make([]database.GetFeedStatsRow, len(m.feedList))
		copy(m.unfilteredFeedList, m.feedList)
		return m, nil
	}

	return m, nil
}

func (m Model) handleItemListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "q", "esc", "ctrl+c":
		// Clear search mode when returning to feed list
		m.searchMode = false
		m.searchActive = false
		m.searchQuery = ""
		m.state = FeedListView
		m.cursor = m.savedFeedCursor
		return m, loadFeedList(m.feedManager)

	case "j", "down":
		if len(m.itemList) > 0 && m.cursor < len(m.itemList)-1 {
			m.cursor++
			m.savedItemCursor = m.cursor
		}

	case "k", "up":
		if len(m.itemList) > 0 && m.cursor > 0 {
			m.cursor--
			m.savedItemCursor = m.cursor
		}

	case "ctrl+d":
		if len(m.itemList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = min(m.cursor+pageSize, len(m.itemList)-1)
			m.savedItemCursor = m.cursor
		}

	case "ctrl+u":
		if len(m.itemList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = max(m.cursor-pageSize, 0)
			m.savedItemCursor = m.cursor
		}

	case "enter":
		if len(m.itemList) > 0 && m.cursor < len(m.itemList) {
			m.currentItem = m.itemList[m.cursor]
			content := m.currentItem.Content
			if content == "" {
				content = m.currentItem.Description
			}
			m.links = m.feedManager.ExtractLinks(content)
			m.state = ArticleView

			if !m.currentItem.Read {
				return m, markItemRead(m.feedManager, m.currentItem.ID)
			}
		}

	case "r":
		if !m.refreshing {
			m.refreshing = true
			m.refreshStatus = "Refreshing feed..."
			return m, tea.Batch(
				func() tea.Msg { return RefreshStartMsg{Status: "Refreshing feed..."} },
				refreshFeedAndReload(m.feedManager, m.selectedFeed),
			)
		}

	case "A":
		// Mark all items in the current feed as read
		return m, markAllItemsReadInFeed(m.feedManager, m.selectedFeed)

	case "N":
		// Toggle read status of current item
		if len(m.itemList) > 0 && m.cursor < len(m.itemList) {
			item := m.itemList[m.cursor]
			return m, toggleItemReadStatus(m.feedManager, item.ID, item.Read)
		}

	case "o":
		// Open the current item's link in the browser
		if len(m.itemList) > 0 && m.cursor < len(m.itemList) {
			item := m.itemList[m.cursor]
			if item.Link != "" {
				return m, openLink(item.Link)
			}
		}

	case "c":
		m.previousState = m.state
		m.state = SettingsView
		m.cursor = 0
		m.savedSettingsCursor = 0
		return m, nil

	case "t":
		m.previousState = m.state
		m.state = TasksView
		m.cursor = 0
		m.savedTasksCursor = 0
		return m, loadTaskList(m.taskManager)
	}

	return m, nil
}

func (m Model) handleArticleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "q", "esc", "ctrl+c":
		m.state = ItemListView
		m.cursor = m.savedItemCursor
		m.showRawHTML = false   // Reset raw HTML view when exiting
		m.articleViewScroll = 0 // Reset scroll position when exiting
		return m, loadItemList(m.feedManager, m.selectedFeed)

	case "j", "down":
		// Calculate max scroll based on content
		allLines := m.getArticleContentLines()
		availableHeight := m.height - 3
		if availableHeight < 1 {
			availableHeight = 1
		}
		maxScroll := len(allLines) - availableHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.articleViewScroll < maxScroll {
			m.articleViewScroll++
		}

	case "k", "up":
		if m.articleViewScroll > 0 {
			m.articleViewScroll--
		}

	case "ctrl+d":
		// Calculate max scroll based on content
		allLines := m.getArticleContentLines()
		availableHeight := m.height - 3
		if availableHeight < 1 {
			availableHeight = 1
		}
		maxScroll := len(allLines) - availableHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.articleViewScroll += pageSize
		if m.articleViewScroll > maxScroll {
			m.articleViewScroll = maxScroll
		}

	case "ctrl+u":
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.articleViewScroll -= pageSize
		if m.articleViewScroll < 0 {
			m.articleViewScroll = 0
		}

	case "r":
		// Toggle raw HTML view
		m.showRawHTML = !m.showRawHTML
		return m, nil

	case "o":
		// Open the current item's link in the browser
		if m.currentItem.Link != "" {
			return m, openLink(m.currentItem.Link)
		}

	case "n":
		// Advance to the next article
		if len(m.itemList) > 0 {
			nextCursor := (m.savedItemCursor + 1) % len(m.itemList)
			if nextCursor < len(m.itemList) {
				m.savedItemCursor = nextCursor
				m.cursor = nextCursor
				m.currentItem = m.itemList[nextCursor]
				content := m.currentItem.Content
				if content == "" {
					content = m.currentItem.Description
				}
				m.links = m.feedManager.ExtractLinks(content)
				m.showRawHTML = false   // Reset raw HTML view when navigating
				m.articleViewScroll = 0 // Reset scroll position when navigating

				if !m.currentItem.Read {
					return m, markItemRead(m.feedManager, m.currentItem.ID)
				}
			}
		}

	case "N":
		// Go back to the previous article
		if len(m.itemList) > 0 {
			prevCursor := m.savedItemCursor - 1
			if prevCursor < 0 {
				prevCursor = len(m.itemList) - 1
			}
			if prevCursor >= 0 && prevCursor < len(m.itemList) {
				m.savedItemCursor = prevCursor
				m.cursor = prevCursor
				m.currentItem = m.itemList[prevCursor]
				content := m.currentItem.Content
				if content == "" {
					content = m.currentItem.Description
				}
				m.links = m.feedManager.ExtractLinks(content)
				m.showRawHTML = false   // Reset raw HTML view when navigating
				m.articleViewScroll = 0 // Reset scroll position when navigating

				if !m.currentItem.Read {
					return m, markItemRead(m.feedManager, m.currentItem.ID)
				}
			}
		}

	case "c":
		m.previousState = m.state
		m.state = SettingsView
		m.cursor = 0
		m.savedSettingsCursor = 0
		return m, nil

	case "t":
		m.previousState = m.state
		m.state = TasksView
		m.cursor = 0
		m.savedTasksCursor = 0
		return m, loadTaskList(m.taskManager)

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		linkNum := int(msg.String()[0] - '1')
		if linkNum < len(m.links) {
			return m, openLink(m.links[linkNum])
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit", m.err)
	}

	switch m.state {
	case FeedListView:
		return m.renderFeedList()
	case ItemListView:
		return m.renderItemList()
	case ArticleView:
		return m.renderArticle()
	case FeedInfoView:
		return m.renderFeedInfo()
	case LogView:
		return m.renderLogList()
	case LogDetailView:
		return m.renderLogDetail()
	case TasksView:
		return m.renderTasksView()
	case HelpView:
		return m.renderHelpView()
	case SettingsView:
		return m.renderSettingsView()
	case URLsView:
		return m.renderURLsView()
	}

	return "Loading..."
}

func (m Model) getTitleStyle() lipgloss.Style {
	theme := themes.GetThemeByName(m.config.ThemeName)
	return lipgloss.NewStyle().Bold(true).Background(lipgloss.Color(theme.FilterColor)).Foreground(lipgloss.Color(theme.TitleColorFg)).Width(m.width)
}

func (m Model) getSelectedStyle() lipgloss.Style {
	theme := themes.GetThemeByName(m.config.ThemeName)

	switch m.config.HighlightStyle {
	case "underline":
		return lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color(theme.SelectedItemColor))
	case "prefix", "prefix-underline":
		// Prefix is handled separately in rendering
		if m.config.HighlightStyle == "prefix-underline" {
			return lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color(theme.SelectedItemColor))
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SelectedItemColor))
	case "background":
		fallthrough
	default:
		return lipgloss.NewStyle().Background(lipgloss.Color(theme.SelectedItemColor)).Foreground(lipgloss.Color("229"))
	}
}

// applyHighlight applies the appropriate highlight style to a line
func (m Model) applyHighlight(line string, isSelected bool) string {
	// Add prefix if needed
	if isSelected && (m.config.HighlightStyle == "prefix" || m.config.HighlightStyle == "prefix-underline") {
		line = "> " + line
	} else if m.config.HighlightStyle == "prefix" || m.config.HighlightStyle == "prefix-underline" {
		line = "  " + line
	}

	// Apply style
	if isSelected {
		return m.getSelectedStyle().Render(line)
	}

	return line
}

func (m Model) getHelpStyle() lipgloss.Style {
	theme := themes.GetThemeByName(m.config.ThemeName)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.FilterColor))
}

func (m Model) getUnreadStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
}

func (m Model) renderFeedList() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat " + version.GetVersion() + " - RSS Reader"))

	if m.refreshing {
		b.WriteString(" - ")
		b.WriteString(m.getHelpStyle().Render(m.refreshStatus))
	}

	b.WriteString("\n\n")

	// Build status bar
	viewKeys := GetViewKeys(FeedListView)
	globalHelp := "h: help | q: quit"
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBarLeft := m.getHelpStyle().Render(statusBarText)

	// Add countdown on the right if auto reload is enabled
	var statusBar string
	if m.config.AutoReload && !m.nextReloadTime.IsZero() {
		timeUntilReload := time.Until(m.nextReloadTime)
		if timeUntilReload > 0 {
			minutes := int(timeUntilReload.Minutes())
			rightText := fmt.Sprintf("next reload in %dm", minutes)
			// Calculate spacing to push right part to the right
			leftLen := len(statusBarText)
			rightLen := len(rightText)
			spacing := m.width - leftLen - rightLen - 2
			if spacing < 1 {
				spacing = 1
			}
			// Build complete status bar text then apply styling once
			completeText := statusBarText + strings.Repeat(" ", spacing) + rightText
			statusBar = m.getHelpStyle().Render(completeText)
		} else {
			statusBar = statusBarLeft
		}
	} else {
		statusBar = statusBarLeft
	}

	if len(m.feedList) == 0 {
		var content string
		var contentLines int
		// Only show "add URLs" message if there are actually no feeds in the database
		// Don't show it if feeds are just filtered out (e.g., ShowReadFeeds = no)
		if m.totalFeedCount == 0 {
			var urlPath string
			if m.urlsFilePath != "" {
				urlPath = m.urlsFilePath
			} else {
				urlPath = "~/.config/newsgoat/urls"
			}
			content = "Add RSS feeds to " + urlPath + " by\n" +
				"editing the file by pressing 'U' or press 'u' to add\n" +
				"a single feed URL.\n" +
				"\n" +
				"Hints:\n" +
				"- Press 'R' to reload all feeds\n" +
				"- Press 'c' to view the config\n" +
				"- See keyboard shortcuts with 'h'"
			contentLines = 8
		} else if m.searchMode && m.searchQuery != "" {
			content = "No feeds match search"
			contentLines = 1
		} else {
			content = ""
			contentLines = 0
		}
		// Calculate padding to push status bar to bottom
		// usedLines = title (1) + empty line (1) + empty line (1) + content lines + status bar (1) + search line (1)
		headerLines := 2 // title + 2 newlines (counts as 2 lines for display purposes)
		bottomLines := 2 // status bar + search line
		usedLines := headerLines + contentLines + bottomLines
		padding := m.height - usedLines
		if padding < 0 {
			padding = 0
		}
		if content != "" {
			b.WriteString(content)
			b.WriteString("\n")
		}
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(statusBar)
		// Show status message line or search line
		b.WriteString("\n")
		if m.statusMessage != "" {
			theme := themes.GetThemeByName(m.config.ThemeName)
			var messageStyle lipgloss.Style
			if m.statusMessageType == "error" {
				messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
			} else {
				messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SelectedItemColor))
			}
			b.WriteString(messageStyle.Render(m.statusMessage))
		} else if m.searchMode {
			searchPrompt := "/" + m.searchQuery
			b.WriteString(m.getHelpStyle().Render(searchPrompt))
		}
		return b.String()
	}

	// Calculate viewport for scrolling
	// Reserve space for:
	// - Title line (1)
	// - Empty line after header (1)
	// - Status bar at bottom (1)
	// - Scroll indicator line (1)
	// - Search prompt line (1) - always allocated
	// Total: 5 lines
	availableHeight := m.height - 5
	if availableHeight < 3 {
		availableHeight = 3 // Minimum usable height
	}

	// Calculate start and end indices for viewport
	start := 0
	end := len(m.feedList)

	if len(m.feedList) > availableHeight {
		// Center the cursor in the viewport when possible
		halfHeight := availableHeight / 2
		start = max(0, m.cursor-halfHeight)
		end = min(len(m.feedList), start+availableHeight)

		// Adjust start if we're near the end
		if end-start < availableHeight {
			start = max(0, end-availableHeight)
		}
	}

	// Render visible feeds
	feedLines := 0
	for i := start; i < end; i++ {
		feed := m.feedList[i]

		// Status emoji: error emoji if error (but not when refreshing), unread if has unread items, nothing if all read
		var statusEmoji string
		// Don't show error emoji when actively refreshing - let the spinner show instead
		if feed.LastError.Valid && feed.LastError.String != "" && !m.refreshingFeeds[feed.ID] {
			// Try to determine error type from error message
			errorMsg := feed.LastError.String
			if strings.Contains(errorMsg, "404") {
				statusEmoji = "🔍" // Not found
			} else if strings.Contains(errorMsg, "403") {
				statusEmoji = "🚫" // Forbidden
			} else if strings.Contains(errorMsg, "429") {
				statusEmoji = "⏱️" // Too many requests
			} else if strings.Contains(errorMsg, "500") || strings.Contains(errorMsg, "502") || strings.Contains(errorMsg, "503") {
				statusEmoji = "⚠️" // Server error
			} else if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "context deadline exceeded") {
				statusEmoji = "⌛" // Timeout
			} else {
				statusEmoji = "❌" // Generic error
			}
		} else if feed.UnreadItems > 0 {
			statusEmoji = "🔵" // Unread items
		} else {
			statusEmoji = "  " // All read - two spaces to align with emoji width
		}

		// Spinner - 2 character space reserved for spinner when refreshing
		var spinner string
		if m.refreshingFeeds[feed.ID] {
			spinnerFrames := themes.GetSpinnerFrames(m.config.SpinnerType)
			spinner = spinnerFrames[m.spinnerFrame%len(spinnerFrames)] + " "
		} else {
			spinner = "  " // Two spaces when not spinning
		}

		// Count string right-justified to 9 characters
		countStr := fmt.Sprintf("(%d/%d)", feed.UnreadItems, feed.TotalItems)
		paddedCount := fmt.Sprintf("%9s", countStr)

		// Get display title - override for GitHub and GitLab feeds
		displayTitle := getDisplayTitle(feed)

		// Construct the line: status (emoji or 2 spaces) + spinner (2 chars) + count (9 chars) + space + feed title
		line := statusEmoji + spinner + paddedCount + " " + displayTitle

		// Apply highlighting
		if i == m.cursor {
			line = m.applyHighlight(line, true)
		} else {
			if feed.UnreadItems > 0 {
				line = m.getUnreadStyle().Render(line)
			}
			line = m.applyHighlight(line, false)
		}

		b.WriteString(line)
		b.WriteString("\n")
		feedLines++
	}

	// Calculate padding to push status bar to bottom
	headerLines := 2    // title + empty line
	statusBarLines := 2 // (scroll info + status bar on same line) + search line
	usedLines := headerLines + feedLines + statusBarLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more feeds
	var scrollInfo string
	if len(m.feedList) > availableHeight {
		scrollInfo = fmt.Sprintf("(%d-%d of %d)  ", start+1, end, len(m.feedList))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
	}

	// Recalculate status bar if we have countdown to account for scroll indicator
	if m.config.AutoReload && !m.nextReloadTime.IsZero() {
		timeUntilReload := time.Until(m.nextReloadTime)
		if timeUntilReload > 0 {
			minutes := int(timeUntilReload.Minutes())
			rightText := fmt.Sprintf("next reload in %dm", minutes)
			// Calculate spacing accounting for scroll indicator
			leftLen := len(statusBarText)
			rightLen := len(rightText)
			scrollLen := len(scrollInfo)
			spacing := m.width - scrollLen - leftLen - rightLen - 2
			if spacing < 1 {
				spacing = 1
			}
			completeText := statusBarText + strings.Repeat(" ", spacing) + rightText
			statusBar = m.getHelpStyle().Render(completeText)
		}
	}

	b.WriteString(statusBar)

	// Show status message line above search line if present
	b.WriteString("\n")
	if m.statusMessage != "" {
		theme := themes.GetThemeByName(m.config.ThemeName)
		var messageStyle lipgloss.Style
		if m.statusMessageType == "error" {
			messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
		} else {
			messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.SelectedItemColor))
		}
		b.WriteString(messageStyle.Render(m.statusMessage))
	} else if m.addingURL {
		// Show URL input modal
		urlPrompt := "Add URL: " + m.urlInput
		b.WriteString(m.getHelpStyle().Render(urlPrompt))
	} else if m.searchMode {
		searchPrompt := "/" + m.searchQuery
		b.WriteString(m.getHelpStyle().Render(searchPrompt))
	}

	return b.String()
}

func (m Model) renderItemList() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Feed Items"))

	if m.refreshing {
		b.WriteString(" - ")
		b.WriteString(m.getHelpStyle().Render(m.refreshStatus))
	}

	b.WriteString("\n\n")

	// Build status bar
	viewKeys := GetViewKeys(ItemListView)
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)

	if len(m.itemList) == 0 {
		content := "No items found."
		// Calculate padding to push status bar to bottom
		contentLines := strings.Count(b.String()+content, "\n") + 2
		padding := m.height - contentLines - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(content)
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(statusBar)
		return b.String()
	}

	// Calculate viewport for scrolling
	// Reserve space for:
	// - Title line (1)
	// - Empty line after header (1)
	// - Status bar at bottom (1)
	// - Scroll indicator line (1)
	// Total: 4 lines
	availableHeight := m.height - 4
	if availableHeight < 3 {
		availableHeight = 3 // Minimum usable height
	}

	// Calculate start and end indices for viewport
	start := 0
	end := len(m.itemList)

	if len(m.itemList) > availableHeight {
		// Center the cursor in the viewport when possible
		halfHeight := availableHeight / 2
		start = max(0, m.cursor-halfHeight)
		end = min(len(m.itemList), start+availableHeight)

		// Adjust start if we're near the end
		if end-start < availableHeight {
			start = max(0, end-availableHeight)
		}
	}

	// Render visible items
	itemLines := 0
	for i := start; i < end; i++ {
		item := m.itemList[i]

		// Format date as MM-DD
		datePrefix := "     " // Default fallback if no date
		if item.Published.Valid {
			datePrefix = item.Published.Time.Format("01-02")
		}

		readPrefix := "  "
		if !item.Read {
			readPrefix = "🔵"
		}

		line := datePrefix + " " + readPrefix + item.Title

		// Apply highlighting
		if i == m.cursor {
			line = m.applyHighlight(line, true)
		} else {
			if !item.Read {
				line = m.getUnreadStyle().Render(line)
			}
			line = m.applyHighlight(line, false)
		}

		b.WriteString(line)
		b.WriteString("\n")
		itemLines++
	}

	// Calculate padding to push status bar to bottom
	headerLines := 2 // title + empty line
	usedLines := headerLines + itemLines
	padding := m.height - usedLines - 1 // -1 for status bar line
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more items
	if len(m.itemList) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d)", start+1, end, len(m.itemList))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
		b.WriteString("  ")
	}

	b.WriteString(statusBar)

	return b.String()
}

func (m *Model) getArticleContentLines() []string {
	// Build content
	var contentBuilder strings.Builder

	content := m.currentItem.Content
	if content == "" {
		content = m.currentItem.Description
	}

	// If showing raw HTML, apply word wrapping and skip processing
	if m.showRawHTML {
		// Split content by newlines to preserve existing line breaks
		lines := strings.Split(content, "\n")
		var wrappedLines []string

		// Apply word wrap to each line
		wrapWidth := m.width - 4 // Leave some margin
		if wrapWidth < 40 {
			wrapWidth = 40
		}

		for _, line := range lines {
			if line == "" {
				wrappedLines = append(wrappedLines, "")
			} else {
				wrapped := wrapText(line, wrapWidth)
				wrappedLines = append(wrappedLines, wrapped...)
			}
		}

		return wrappedLines
	}

	// Add link markers to HTML BEFORE converting to markdown
	// This ensures the markers are properly preserved during conversion
	content, _ = m.feedManager.AddLinkMarkersToHTML(content)

	// Convert HTML to markdown
	content = m.feedManager.ConvertHTMLToMarkdown(content)

	// Render markdown content using glamour
	if m.glamourRenderer != nil {
		renderedContent, err := m.glamourRenderer.Render(content)
		if err == nil {
			content = renderedContent
		}
	}

	contentBuilder.WriteString(content)
	contentBuilder.WriteString("\n\n")

	if len(m.links) > 0 {
		contentBuilder.WriteString(m.getHelpStyle().Render("Links:"))
		contentBuilder.WriteString("\n")
		for i, link := range m.links {
			contentBuilder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, link))
		}
	}

	// Split content into lines for scrolling
	return strings.Split(contentBuilder.String(), "\n")
}

func (m Model) renderArticle() string {
	allLines := m.getArticleContentLines()

	// Calculate available height for content (height - title - status bar)
	availableHeight := m.height - 3 // -3 for title (2 lines) and status bar (1 line)
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine which lines to show based on scroll position
	start := m.articleViewScroll
	if start >= len(allLines) {
		start = len(allLines) - 1
	}
	if start < 0 {
		start = 0
	}

	end := start + availableHeight
	if end > len(allLines) {
		end = len(allLines)
	}

	visibleLines := allLines[start:end]

	// Build final output
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render(m.currentItem.Title))
	b.WriteString("\n\n")

	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Calculate padding to push status bar to bottom
	usedLines := len(visibleLines) + 2  // +2 for title and spacing (title + 2 newlines = 2 display lines)
	padding := m.height - usedLines - 1 // -1 for status bar
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more lines
	viewKeys := GetViewKeys(ArticleView)
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)
	if len(allLines) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d) ", start+1, end, len(allLines))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
	}
	b.WriteString(statusBar)

	return b.String()
}

func (m Model) handleLogListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "q", "esc", "ctrl+c":
		// Clear search mode when returning to feed list
		m.searchMode = false
		m.searchActive = false
		m.searchQuery = ""
		m.state = FeedListView
		return m, loadFeedList(m.feedManager)

	case "j", "down":
		if len(m.logList) > 0 {
			m.cursor = (m.cursor + 1) % len(m.logList)
			m.savedLogCursor = m.cursor
		}

	case "k", "up":
		if len(m.logList) > 0 {
			m.cursor = (m.cursor - 1 + len(m.logList)) % len(m.logList)
			m.savedLogCursor = m.cursor
		}

	case "ctrl+d":
		if len(m.logList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = min(m.cursor+pageSize, len(m.logList)-1)
			m.savedLogCursor = m.cursor
		}

	case "ctrl+u":
		if len(m.logList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = max(m.cursor-pageSize, 0)
			m.savedLogCursor = m.cursor
		}

	case "enter":
		if len(m.logList) > 0 && m.cursor < len(m.logList) {
			m.currentLog = m.logList[m.cursor]
			m.state = LogDetailView
		}

	case "A":
		return m, clearAllLogMessages(m.feedManager)
	}

	return m, nil
}

func (m Model) handleLogDetailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "q", "esc", "ctrl+c":
		m.state = LogView
		m.cursor = m.savedLogCursor
		return m, nil
	}

	return m, nil
}

func (m Model) handleTasksViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "q", "esc", "ctrl+c":
		// Clear search mode when returning to feed list
		m.searchMode = false
		m.searchActive = false
		m.searchQuery = ""
		m.state = FeedListView
		return m, loadFeedList(m.feedManager)

	case "j", "down":
		if len(m.taskList) > 0 {
			m.cursor = (m.cursor + 1) % len(m.taskList)
			m.savedTasksCursor = m.cursor
		}

	case "k", "up":
		if len(m.taskList) > 0 {
			m.cursor = (m.cursor - 1 + len(m.taskList)) % len(m.taskList)
			m.savedTasksCursor = m.cursor
		}

	case "ctrl+d":
		if len(m.taskList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = min(m.cursor+pageSize, len(m.taskList)-1)
			m.savedTasksCursor = m.cursor
		}

	case "ctrl+u":
		if len(m.taskList) > 0 {
			pageSize := m.height / 2
			if pageSize < 1 {
				pageSize = 5
			}
			m.cursor = max(m.cursor-pageSize, 0)
			m.savedTasksCursor = m.cursor
		}

	case "A":
		return m, clearFailedTasks(m.taskManager)

	case "D":
		if len(m.taskList) > 0 && m.cursor < len(m.taskList) {
			taskID := m.taskList[m.cursor].ID
			return m, removeTask(m.taskManager, taskID)
		}

	case "r":
		// Refresh the task list
		return m, loadTaskList(m.taskManager)
	}

	return m, nil
}

func (m Model) renderLogList() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Log Messages"))
	b.WriteString("\n\n")

	// Build status bar
	viewKeys := GetViewKeys(LogView)
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)

	if len(m.logList) == 0 {
		content := "No log messages found."
		// Calculate padding to push status bar to bottom
		contentLines := strings.Count(b.String()+content, "\n") + 2
		padding := m.height - contentLines - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(content)
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(statusBar)
		return b.String()
	}

	// Calculate viewport for scrolling
	// Reserve space for:
	// - Title line (1)
	// - Empty line after header (1)
	// - Status bar at bottom (1)
	// - Scroll indicator line (1)
	// Total: 4 lines
	availableHeight := m.height - 4
	if availableHeight < 3 {
		availableHeight = 3
	}

	start := 0
	end := len(m.logList)

	if len(m.logList) > availableHeight {
		halfHeight := availableHeight / 2
		start = max(0, m.cursor-halfHeight)
		end = min(len(m.logList), start+availableHeight)

		if end-start < availableHeight {
			start = max(0, end-availableHeight)
		}
	}

	// Render visible log messages
	logLines := 0
	for i := start; i < end; i++ {
		log := m.logList[i]

		// Format timestamp as YYYY-MM-DD HH:MM:SS
		timestampStr := "                   " // Default fallback
		if log.Timestamp.Valid {
			timestampStr = log.Timestamp.Time.Format("2006-01-02 15:04:05")
		}

		line := timestampStr + "  " + log.Message

		// Apply highlighting
		line = m.applyHighlight(line, i == m.cursor)

		b.WriteString(line)
		b.WriteString("\n")
		logLines++
	}

	// Calculate padding to push status bar to bottom
	headerLines := 2    // title + empty line
	statusBarLines := 2 // scroll info + status bar
	usedLines := headerLines + logLines + statusBarLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more logs
	if len(m.logList) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d)", start+1, end, len(m.logList))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
		b.WriteString("  ")
	}

	b.WriteString(statusBar)

	return b.String()
}

func (m Model) renderLogDetail() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Log Message Details"))
	b.WriteString("\n\n")

	// Timestamp
	if m.currentLog.Timestamp.Valid {
		b.WriteString(fmt.Sprintf("Time: %s\n", m.currentLog.Timestamp.Time.Format("2006-01-02 15:04:05")))
	}

	// Level
	b.WriteString(fmt.Sprintf("Level: %s\n", m.currentLog.Level))

	// Message
	b.WriteString(fmt.Sprintf("Message: %s\n\n", m.currentLog.Message))

	// Attributes (if any)
	if m.currentLog.Attributes.Valid && m.currentLog.Attributes.String != "" {
		b.WriteString("Attributes:\n")

		// Try to parse and pretty print JSON
		var attrs map[string]interface{}
		if err := json.Unmarshal([]byte(m.currentLog.Attributes.String), &attrs); err == nil {
			// Successfully parsed JSON - pretty print each attribute
			// Print attributes (skip source_file and source_line as they always point to logger.go)
			for key, value := range attrs {
				if key == "source_file" || key == "source_line" {
					continue // Skip source attributes
				}

				// Format the value as a string
				valueStr := fmt.Sprintf("%v", value)

				// Word wrap the value if it's long
				// Calculate available width: total width - indent (2) - key length - ": " (2) - margin (2)
				availableWidth := m.width - 6 - len(key)
				if availableWidth < 20 {
					availableWidth = 20 // Minimum width
				}

				// Wrap the value
				wrappedLines := wrapText(valueStr, availableWidth)

				// Print first line with key
				if len(wrappedLines) > 0 {
					b.WriteString(fmt.Sprintf("  %s: %s\n", key, wrappedLines[0]))
					// Print remaining lines indented
					for i := 1; i < len(wrappedLines); i++ {
						// Indent to align with first line value
						indent := strings.Repeat(" ", len(key)+4)
						b.WriteString(fmt.Sprintf("%s%s\n", indent, wrappedLines[i]))
					}
				}
			}
		} else {
			// Failed to parse JSON, just display raw (with wrapping)
			wrappedLines := wrapText(m.currentLog.Attributes.String, m.width-4)
			for _, line := range wrappedLines {
				b.WriteString("  " + line + "\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(m.getHelpStyle().Render("h: help"))

	return b.String()
}

// startNextBatchOfFeeds starts refreshing the next batch of feeds up to maxConcurrency
func (m *Model) startNextBatchOfFeeds() tea.Cmd {
	if len(m.pendingFeeds) == 0 {
		return nil
	}

	// Calculate how many feeds we can start (respect concurrency limit)
	currentlyRefreshing := len(m.refreshingFeeds)
	availableSlots := m.maxConcurrency - currentlyRefreshing

	if availableSlots <= 0 {
		return nil // Already at capacity
	}

	// Start feeds up to the available slots
	var cmds []tea.Cmd
	feedsToStart := min(availableSlots, len(m.pendingFeeds))

	for i := 0; i < feedsToStart; i++ {
		feedID := m.pendingFeeds[0]
		m.pendingFeeds = m.pendingFeeds[1:] // Remove from queue

		// Capture feedID in closure to avoid variable capture issue
		func(id int64) {
			cmds = append(cmds, tea.Batch(
				func() tea.Msg { return FeedRefreshStartMsg{FeedID: id} },
				refreshFeedAndReload(m.feedManager, id),
			))
		}(feedID)
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m Model) handleHelpViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "h", "?", "ctrl+c":
		// Return to previous view and reset scroll
		m.state = m.previousState
		m.helpViewScroll = 0
		return m, nil

	case "j", "down":
		m.helpViewScroll++
		return m, nil

	case "k", "up":
		if m.helpViewScroll > 0 {
			m.helpViewScroll--
		}
		return m, nil

	case "ctrl+d":
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.helpViewScroll += pageSize
		return m, nil

	case "ctrl+u":
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.helpViewScroll -= pageSize
		if m.helpViewScroll < 0 {
			m.helpViewScroll = 0
		}
		return m, nil
	}

	return m, nil
}

func (m Model) renderHelpView() string {
	// Build the full content first
	var content strings.Builder

	// Global keys section
	content.WriteString("Global\n")
	for _, binding := range GlobalKeys {
		content.WriteString(fmt.Sprintf("  %-15s %s\n", binding.Key, binding.Description))
	}
	content.WriteString("\n")

	// Feed List View keys
	content.WriteString("Feed List View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "r", "Refresh selected feed"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "R", "Refresh all feeds"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "A", "Mark all items in feed as read"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "i", "Show feed info"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "/", "Search feeds"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "u", "Add URL (with discovery)"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "U", "Edit URLs in $EDITOR"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "ctrl+r", "Reload URLs from file"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "l", "View logs"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "t", "View tasks"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "c", "View settings"))
	content.WriteString("\n")

	// Item List View keys
	content.WriteString("Item List View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "r", "Refresh feed"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "R", "Refresh all feeds"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "A", "Mark all items as read"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "N", "Toggle read status of item"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "o", "Open item link in browser"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "c", "View settings"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "t", "View tasks"))
	content.WriteString("\n")

	// Article View keys
	content.WriteString("Article View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "1-9", "Open numbered link in browser"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "o", "Open article link in browser"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "n", "Next article"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "N", "Previous article"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "r", "Toggle raw HTML view"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "c", "View settings"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "t", "View tasks"))
	content.WriteString("\n")

	// Settings View keys
	content.WriteString("Settings View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "?", "Toggle settings help"))
	content.WriteString("\n")

	// Tasks View keys
	content.WriteString("Tasks View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "d", "Remove selected task"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "c", "Clear all failed tasks"))
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "l", "View logs"))
	content.WriteString("\n")

	// Log View keys
	content.WriteString("Log View\n")
	content.WriteString(fmt.Sprintf("  %-15s %s\n", "c", "Clear all log messages"))
	content.WriteString("\n")

	// Status icons legend - unified section
	content.WriteString("Status Icons\n")
	content.WriteString("  🔵              Unread items/feed\n")
	content.WriteString("  🔍              404 Not Found\n")
	content.WriteString("  🚫              403 Forbidden\n")
	content.WriteString("  ⏱️              429 Too Many Requests\n")
	content.WriteString("  ⚠️              500/502/503 Server Error\n")
	content.WriteString("  ⌛              Timeout\n")
	content.WriteString("  ❌              Other Error\n")
	content.WriteString("  🕓              Pending task\n")
	content.WriteString("  🔄              Running task\n")
	content.WriteString("  💥              Failed task\n")
	content.WriteString("\n")

	// Environment Variables section
	content.WriteString("Environment Variables\n")
	content.WriteString("  GITHUB_FEED_TOKEN   Access token for private GitHub repository feeds\n")
	content.WriteString("  GITLAB_FEED_TOKEN   Access token for private GitLab repository feeds\n")

	// Split content into lines
	allLines := strings.Split(content.String(), "\n")

	// Calculate viewport
	// Reserve space for: title (1), empty line (1), status bar (1) = 3 lines
	availableHeight := m.height - 3
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Ensure scroll doesn't go past the end
	maxScroll := len(allLines) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.helpViewScroll > maxScroll {
		m.helpViewScroll = maxScroll
	}

	// Extract visible lines
	start := m.helpViewScroll
	end := min(start+availableHeight, len(allLines))
	visibleLines := allLines[start:end]

	// Build the final output
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	// Render visible lines
	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Calculate padding to push status bar to bottom
	usedLines := 2 + len(visibleLines)  // title + empty line + visible content
	padding := m.height - usedLines - 1 // -1 for status bar
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if needed
	if len(allLines) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d) ", start+1, end, len(allLines))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
	}

	b.WriteString(m.getHelpStyle().Render("j/k: scroll | esc/h: return"))

	return b.String()
}

func (m Model) renderTasksView() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Tasks"))
	b.WriteString("\n\n")

	// Build status bar
	viewKeys := GetViewKeys(TasksView)
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)

	if len(m.taskList) == 0 {
		content := "No tasks found."
		// Calculate padding to push status bar to bottom
		contentLines := strings.Count(b.String()+content, "\n") + 2
		padding := m.height - contentLines - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(content)
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(statusBar)
		return b.String()
	}

	// Calculate viewport for scrolling
	// Reserve space for:
	// - Title line (1)
	// - Empty line after header (1)
	// - Status bar at bottom (1)
	// - Scroll indicator line (1)
	// Total: 4 lines
	availableHeight := m.height - 4
	if availableHeight < 3 {
		availableHeight = 3
	}

	start := 0
	end := len(m.taskList)

	if len(m.taskList) > availableHeight {
		halfHeight := availableHeight / 2
		start = max(0, m.cursor-halfHeight)
		end = min(len(m.taskList), start+availableHeight)

		if end-start < availableHeight {
			start = max(0, end-availableHeight)
		}
	}

	// Render visible tasks
	taskLines := 0
	for i := start; i < end; i++ {
		task := m.taskList[i]

		// Status emoji based on task status
		var statusEmoji string
		switch task.Status {
		case tasks.TaskStatusPending:
			statusEmoji = "🕓"
		case tasks.TaskStatusRunning:
			statusEmoji = "🔄"
		case tasks.TaskStatusFailed:
			statusEmoji = "💥"
		default:
			statusEmoji = " "
		}

		// Build task description
		taskDesc := string(task.Type)
		if feedURL, ok := task.Data["url"].(string); ok {
			taskDesc = feedURL
		}

		// Format timestamp
		timeStr := task.CreatedAt.Format("15:04:05")

		line := fmt.Sprintf("%s %s %s", statusEmoji, timeStr, taskDesc)

		// Apply highlighting
		line = m.applyHighlight(line, i == m.cursor)

		b.WriteString(line)
		b.WriteString("\n")
		taskLines++
	}

	// Calculate padding to push status bar to bottom
	headerLines := 2    // title + empty line
	statusBarLines := 2 // scroll info + status bar
	usedLines := headerLines + taskLines + statusBarLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more tasks
	if len(m.taskList) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d)", start+1, end, len(m.taskList))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
		b.WriteString("  ")
	}

	b.WriteString(statusBar)

	return b.String()
}

func (m Model) handleSettingsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If we're selecting a theme, handle theme selector
	if m.selectingTheme {
		switch msg.String() {
		case "esc":
			// Cancel theme selection
			m.selectingTheme = false
			return m, nil

		case "j", "down":
			themeNames := themes.GetThemeNames()
			if m.themeSelectCursor < len(themeNames)-1 {
				m.themeSelectCursor++
			}
			return m, nil

		case "k", "up":
			if m.themeSelectCursor > 0 {
				m.themeSelectCursor--
			}
			return m, nil

		case "enter":
			// Apply the selected theme
			themeNames := themes.GetThemeNames()
			m.config.ThemeName = themeNames[m.themeSelectCursor]
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}

			// Update glamour renderer
			theme := themes.GetThemeByName(m.config.ThemeName)
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle(theme.GlamourStyle),
				glamour.WithWordWrap(80),
			)
			if err == nil {
				m.glamourRenderer = renderer
			}

			m.selectingTheme = false
			return m, nil
		}
		return m, nil
	}

	// If we're selecting a highlight style, handle highlight selector
	if m.selectingHighlight {
		switch msg.String() {
		case "esc":
			// Cancel highlight selection
			m.selectingHighlight = false
			return m, nil

		case "j", "down":
			highlightStyles := themes.GetHighlightStyles()
			if m.highlightSelectCursor < len(highlightStyles)-1 {
				m.highlightSelectCursor++
			}
			return m, nil

		case "k", "up":
			if m.highlightSelectCursor > 0 {
				m.highlightSelectCursor--
			}
			return m, nil

		case "enter":
			// Apply the selected highlight style
			highlightStyles := themes.GetHighlightStyles()
			m.config.HighlightStyle = highlightStyles[m.highlightSelectCursor]
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}

			m.selectingHighlight = false
			return m, nil
		}
		return m, nil
	}

	// If we're selecting a spinner type, handle spinner selector
	if m.selectingSpinner {
		switch msg.String() {
		case "esc":
			// Cancel spinner selection
			m.selectingSpinner = false
			return m, nil

		case "j", "down":
			spinnerTypes := themes.GetSpinnerTypes()
			if m.spinnerSelectCursor < len(spinnerTypes)-1 {
				m.spinnerSelectCursor++
			}
			return m, nil

		case "k", "up":
			if m.spinnerSelectCursor > 0 {
				m.spinnerSelectCursor--
			}
			return m, nil

		case "enter":
			// Apply the selected spinner type
			spinnerTypes := themes.GetSpinnerTypes()
			m.config.SpinnerType = spinnerTypes[m.spinnerSelectCursor]
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}

			m.selectingSpinner = false
			return m, nil
		}
		return m, nil
	}

	// If we're selecting show read feeds, handle selector
	if m.selectingShowReadFeeds {
		switch msg.String() {
		case "esc":
			m.selectingShowReadFeeds = false
			return m, nil
		case "j", "down":
			if m.showReadFeedsSelectCursor < 1 {
				m.showReadFeedsSelectCursor++
			}
			return m, nil
		case "k", "up":
			if m.showReadFeedsSelectCursor > 0 {
				m.showReadFeedsSelectCursor--
			}
			return m, nil
		case "enter":
			m.config.ShowReadFeeds = (m.showReadFeedsSelectCursor == 0)
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}
			m.selectingShowReadFeeds = false
			return m, loadFeedList(m.feedManager)
		}
		return m, nil
	}

	// If we're selecting auto reload, handle selector
	if m.selectingAutoReload {
		switch msg.String() {
		case "esc":
			m.selectingAutoReload = false
			return m, nil
		case "j", "down":
			if m.autoReloadSelectCursor < 1 {
				m.autoReloadSelectCursor++
			}
			return m, nil
		case "k", "up":
			if m.autoReloadSelectCursor > 0 {
				m.autoReloadSelectCursor--
			}
			return m, nil
		case "enter":
			m.config.AutoReload = (m.autoReloadSelectCursor == 0)
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}
			m.selectingAutoReload = false
			return m, restartReloadTimer()
		}
		return m, nil
	}

	// If we're selecting suppress first reload, handle selector
	if m.selectingSuppressFirstReload {
		switch msg.String() {
		case "esc":
			m.selectingSuppressFirstReload = false
			return m, nil
		case "j", "down":
			if m.suppressFirstReloadSelectCursor < 1 {
				m.suppressFirstReloadSelectCursor++
			}
			return m, nil
		case "k", "up":
			if m.suppressFirstReloadSelectCursor > 0 {
				m.suppressFirstReloadSelectCursor--
			}
			return m, nil
		case "enter":
			m.config.SuppressFirstReload = (m.suppressFirstReloadSelectCursor == 0)
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}
			m.selectingSuppressFirstReload = false
			return m, nil
		}
		return m, nil
	}

	// If we're selecting reload on startup, handle selector
	if m.selectingReloadOnStartup {
		switch msg.String() {
		case "esc":
			m.selectingReloadOnStartup = false
			return m, nil
		case "j", "down":
			if m.reloadOnStartupSelectCursor < 1 {
				m.reloadOnStartupSelectCursor++
			}
			return m, nil
		case "k", "up":
			if m.reloadOnStartupSelectCursor > 0 {
				m.reloadOnStartupSelectCursor--
			}
			return m, nil
		case "enter":
			m.config.ReloadOnStartup = (m.reloadOnStartupSelectCursor == 0)
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}
			m.selectingReloadOnStartup = false
			return m, nil
		}
		return m, nil
	}

	// If we're selecting unread on top, handle selector navigation
	if m.selectingUnreadOnTop {
		switch msg.String() {
		case "esc":
			m.selectingUnreadOnTop = false
			return m, nil
		case "j", "down":
			if m.unreadOnTopSelectCursor < 1 {
				m.unreadOnTopSelectCursor++
			}
			return m, nil
		case "k", "up":
			if m.unreadOnTopSelectCursor > 0 {
				m.unreadOnTopSelectCursor--
			}
			return m, nil
		case "enter":
			m.config.UnreadOnTop = (m.unreadOnTopSelectCursor == 0)
			if err := config.SaveConfig(m.queries, m.config); err != nil {
				m.err = err
			}
			m.selectingUnreadOnTop = false
			return m, loadFeedList(m.feedManager) // Reload to apply sorting
		}
		return m, nil
	}

	// If we're editing reload concurrency, handle input
	if m.editingSettings {
		switch msg.Type {
		case tea.KeyEsc:
			// Cancel editing
			m.editingSettings = false
			m.settingInput = ""
			return m, nil

		case tea.KeyEnter:
			// Save the setting
			m.editingSettings = false
			oldReloadTime := m.config.ReloadTime

			switch m.cursor {
			case 0:
				// Reload concurrency
				if val, parseErr := strconv.Atoi(m.settingInput); parseErr == nil {
					if val >= 1 && val <= 10 {
						m.config.ReloadConcurrency = val
						if err := config.SaveConfig(m.queries, m.config); err != nil {
							m.err = err
						}
					}
				}
			case 1:
				// Reload time
				if val, parseErr := strconv.Atoi(m.settingInput); parseErr == nil {
					if val >= 0 {
						m.config.ReloadTime = val
						if err := config.SaveConfig(m.queries, m.config); err != nil {
							m.err = err
						}
						// If reload time changed, restart the timer
						if oldReloadTime != m.config.ReloadTime {
							return m, restartReloadTimer()
						}
					}
				}
			}

			m.settingInput = ""
			return m, nil

		case tea.KeyBackspace:
			// Delete last character
			if len(m.settingInput) > 0 {
				m.settingInput = m.settingInput[:len(m.settingInput)-1]
			}
			return m, nil

		case tea.KeyRunes:
			// Add character to input
			m.settingInput += string(msg.Runes)
			return m, nil
		}

		return m, nil
	}

	// Not editing - handle navigation
	switch msg.String() {
	case "h":
		m.previousState = m.state
		m.state = HelpView
		return m, nil

	case "?":
		m.showSettingsHelp = !m.showSettingsHelp
		return m, nil

	case "q", "esc", "ctrl+c":
		// Clear search mode when returning to feed list
		m.searchMode = false
		m.searchActive = false
		m.searchQuery = ""
		m.state = FeedListView
		return m, loadFeedList(m.feedManager)

	case "j", "down":
		// 10 total settings
		if m.cursor < 9 {
			m.cursor++
			m.savedSettingsCursor = m.cursor
		}

	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.savedSettingsCursor = m.cursor
		}

	case "enter":
		// Start editing or selecting
		if m.cursor == 0 {
			// Reload concurrency - text input
			m.editingSettings = true
			m.settingInput = fmt.Sprintf("%d", m.config.ReloadConcurrency)
		} else if m.cursor == 1 {
			// Reload time - text input
			m.editingSettings = true
			m.settingInput = fmt.Sprintf("%d", m.config.ReloadTime)
		} else if m.cursor == 2 {
			// Auto reload - open selector
			m.selectingAutoReload = true
			if m.config.AutoReload {
				m.autoReloadSelectCursor = 0
			} else {
				m.autoReloadSelectCursor = 1
			}
		} else if m.cursor == 3 {
			// Suppress first reload - open selector
			m.selectingSuppressFirstReload = true
			if m.config.SuppressFirstReload {
				m.suppressFirstReloadSelectCursor = 0
			} else {
				m.suppressFirstReloadSelectCursor = 1
			}
		} else if m.cursor == 4 {
			// Reload on startup - open selector
			m.selectingReloadOnStartup = true
			if m.config.ReloadOnStartup {
				m.reloadOnStartupSelectCursor = 0
			} else {
				m.reloadOnStartupSelectCursor = 1
			}
		} else if m.cursor == 5 {
			// Theme - open theme selector
			m.selectingTheme = true
			themeNames := themes.GetThemeNames()
			for i, name := range themeNames {
				if name == m.config.ThemeName {
					m.themeSelectCursor = i
					break
				}
			}
		} else if m.cursor == 6 {
			// Highlight style - open highlight selector
			m.selectingHighlight = true
			highlightStyles := themes.GetHighlightStyles()
			for i, style := range highlightStyles {
				if style == m.config.HighlightStyle {
					m.highlightSelectCursor = i
					break
				}
			}
		} else if m.cursor == 7 {
			// Spinner type - open spinner selector
			m.selectingSpinner = true
			spinnerTypes := themes.GetSpinnerTypes()
			for i, spinnerType := range spinnerTypes {
				if spinnerType == m.config.SpinnerType {
					m.spinnerSelectCursor = i
					break
				}
			}
		} else if m.cursor == 8 {
			// Show read feeds - open selector
			m.selectingShowReadFeeds = true
			if m.config.ShowReadFeeds {
				m.showReadFeedsSelectCursor = 0
			} else {
				m.showReadFeedsSelectCursor = 1
			}
		} else if m.cursor == 9 {
			// Unread on top - open selector
			m.selectingUnreadOnTop = true
			if m.config.UnreadOnTop {
				m.unreadOnTopSelectCursor = 0
			} else {
				m.unreadOnTopSelectCursor = 1
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) renderSettingsView() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Settings"))
	b.WriteString("\n\n")

	// If selecting theme, show theme selector
	if m.selectingTheme {
		b.WriteString("Select Theme:\n")
		b.WriteString(m.getHelpStyle().Render("Color scheme for the UI"))
		b.WriteString("\n\n")
		themeNames := themes.GetThemeNames()
		for i, name := range themeNames {
			line := name
			line = m.applyHighlight(line, i == m.themeSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// Calculate padding
		headerLines := 4 // title + empty line + "Select Theme:" + help
		themeLines := len(themeNames)
		statusBarLines := 1
		usedLines := headerLines + themeLines + statusBarLines
		padding := m.height - usedLines
		if padding < 0 {
			padding = 0
		}
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting highlight style, show highlight selector
	if m.selectingHighlight {
		b.WriteString("Select Highlight Style:\n")
		b.WriteString(m.getHelpStyle().Render("How the selected item is highlighted"))
		b.WriteString("\n\n")
		highlightStyles := themes.GetHighlightStyles()
		for i, style := range highlightStyles {
			line := style
			line = m.applyHighlight(line, i == m.highlightSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// Calculate padding
		headerLines := 4 // title + empty line + "Select Highlight Style:" + help
		styleLines := len(highlightStyles)
		statusBarLines := 1
		usedLines := headerLines + styleLines + statusBarLines
		padding := m.height - usedLines
		if padding < 0 {
			padding = 0
		}
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting spinner type, show spinner selector
	if m.selectingSpinner {
		b.WriteString("Select Spinner Type:\n")
		b.WriteString(m.getHelpStyle().Render("Animation style for the loading spinner"))
		b.WriteString("\n\n")
		spinnerTypes := themes.GetSpinnerTypes()
		for i, spinnerType := range spinnerTypes {
			line := spinnerType
			line = m.applyHighlight(line, i == m.spinnerSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		// Calculate padding
		headerLines := 4 // title + empty line + "Select Spinner Type:" + help
		spinnerLines := len(spinnerTypes)
		statusBarLines := 1
		usedLines := headerLines + spinnerLines + statusBarLines
		padding := m.height - usedLines
		if padding < 0 {
			padding = 0
		}
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting show read feeds, show selector
	if m.selectingShowReadFeeds {
		b.WriteString("Show Read Feeds:\n")
		b.WriteString(m.getHelpStyle().Render("Show feeds with no unread items in the list"))
		b.WriteString("\n\n")
		options := []string{"yes", "no"}
		for i, option := range options {
			line := option
			line = m.applyHighlight(line, i == m.showReadFeedsSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", m.height-8))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting auto reload, show selector
	if m.selectingAutoReload {
		b.WriteString("Auto Reload:\n")
		b.WriteString(m.getHelpStyle().Render("Enable continuous automatic reloads using reload time"))
		b.WriteString("\n\n")
		options := []string{"yes", "no"}
		for i, option := range options {
			line := option
			line = m.applyHighlight(line, i == m.autoReloadSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", m.height-8))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting suppress first reload, show selector
	if m.selectingSuppressFirstReload {
		b.WriteString("Suppress First Reload:\n")
		b.WriteString(m.getHelpStyle().Render("Skip the first automatic reload after startup"))
		b.WriteString("\n\n")
		options := []string{"yes", "no"}
		for i, option := range options {
			line := option
			line = m.applyHighlight(line, i == m.suppressFirstReloadSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", m.height-8))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting reload on startup, show selector
	if m.selectingReloadOnStartup {
		b.WriteString("Reload On Startup:\n")
		b.WriteString(m.getHelpStyle().Render("Reload all feeds when the app starts"))
		b.WriteString("\n\n")
		options := []string{"yes", "no"}
		for i, option := range options {
			line := option
			line = m.applyHighlight(line, i == m.reloadOnStartupSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", m.height-8))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If selecting unread on top, show selector
	if m.selectingUnreadOnTop {
		b.WriteString("Unread On Top:\n")
		b.WriteString(m.getHelpStyle().Render("Show feeds with unread items at the top of the feed list"))
		b.WriteString("\n\n")
		options := []string{"yes", "no"}
		for i, option := range options {
			line := option
			line = m.applyHighlight(line, i == m.unreadOnTopSelectCursor)
			b.WriteString(line)
			b.WriteString("\n")
		}

		b.WriteString(strings.Repeat("\n", m.height-8))
		b.WriteString(m.getHelpStyle().Render("enter: select | esc: cancel"))
		return b.String()
	}

	// If showing settings help, show help text
	if m.showSettingsHelp {
		b.WriteString("Settings Help:\n\n")
		help := []string{
			"Reload Concurrency: Number of feeds to refresh in parallel (1-10) - Requires restart",
			"Reload Time: Minutes between automatic reloads",
			"Auto Reload: Enable continuous automatic reloads using reload time",
			"Suppress First Reload: Skip the first automatic reload after startup",
			"Reload On Startup: Reload all feeds when the app starts",
			"Theme: Color scheme for the UI",
			"Highlight Style: How the selected item is highlighted",
			"Spinner Type: Animation style for the loading spinner",
			"Show Read Feeds: Show feeds with no unread items in the list",
			"Unread On Top: Show feeds with unread items at the top of the feed list",
		}
		for _, line := range help {
			wrapped := wrapText(line, m.width-4)
			for _, wrappedLine := range wrapped {
				b.WriteString("  " + wrappedLine + "\n")
			}
		}

		usedLines := 3 + len(help)*2 // Estimate for wrapped lines
		padding := m.height - usedLines - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(strings.Repeat("\n", padding))
		b.WriteString(m.getHelpStyle().Render("esc: close help"))
		return b.String()
	}

	// Build status bar
	var statusBar string
	if m.editingSettings {
		statusBar = m.getHelpStyle().Render("enter: save | esc: cancel")
	} else {
		viewKeys := GetViewKeys(SettingsView)
		viewHelp := FormatStatusBar(viewKeys.StatusBar)
		var statusBarText string
		if viewHelp != "" {
			statusBarText = globalHelp + " | " + viewHelp
		} else {
			statusBarText = globalHelp
		}
		statusBar = m.getHelpStyle().Render(statusBarText)
	}

	// Define settings items
	showReadFeedsStr := "yes"
	if !m.config.ShowReadFeeds {
		showReadFeedsStr = "no"
	}
	unreadOnTopStr := "yes"
	if !m.config.UnreadOnTop {
		unreadOnTopStr = "no"
	}
	reloadTimeStr := fmt.Sprintf("%d minutes", m.config.ReloadTime)
	if m.config.ReloadTime == 0 {
		reloadTimeStr = "disabled"
	}
	autoReloadStr := "yes"
	if !m.config.AutoReload {
		autoReloadStr = "no"
	}
	suppressFirstReloadStr := "yes"
	if !m.config.SuppressFirstReload {
		suppressFirstReloadStr = "no"
	}
	reloadOnStartupStr := "yes"
	if !m.config.ReloadOnStartup {
		reloadOnStartupStr = "no"
	}
	settings := []struct {
		label string
		value string
	}{
		{"Reload Concurrency", fmt.Sprintf("%d (restart required after changing)", m.config.ReloadConcurrency)},
		{"Reload Time", reloadTimeStr},
		{"Auto Reload", autoReloadStr},
		{"Suppress First Reload", suppressFirstReloadStr},
		{"Reload On Startup", reloadOnStartupStr},
		{"Theme", m.config.ThemeName},
		{"Highlight Style", m.config.HighlightStyle},
		{"Spinner Type", m.config.SpinnerType},
		{"Show Read Feeds", showReadFeedsStr},
		{"Unread On Top", unreadOnTopStr},
	}

	// Render settings
	settingLines := 0
	for i, setting := range settings {
		var line string

		// If editing this setting, show input prompt
		if m.editingSettings && i == m.cursor {
			line = fmt.Sprintf("%-25s > %s", setting.label+":", m.settingInput)
			line = m.applyHighlight(line, true)
		} else {
			line = fmt.Sprintf("%-25s %s", setting.label+":", setting.value)
			line = m.applyHighlight(line, i == m.cursor && !m.editingSettings && !m.selectingTheme)
		}

		b.WriteString(line)
		b.WriteString("\n")
		settingLines++
	}

	// Calculate padding to push status bar to bottom
	headerLines := 2    // title + empty line
	statusBarLines := 1 // status bar
	usedLines := headerLines + settingLines + statusBarLines
	padding := m.height - usedLines
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	b.WriteString(statusBar)

	return b.String()
}

func (m Model) handleFeedInfoKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "?":
		m.previousState = m.state
		m.state = HelpView
		m.helpViewScroll = 0
		return m, nil

	case "q", "esc", "ctrl+c":
		m.state = m.previousState
		return m, nil
	}

	return m, nil
}

func (m Model) handleURLsViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		m.state = m.previousState
		m.urlsViewScroll = 0 // Reset scroll position when exiting
		return m, nil

	case "j", "down":
		// Calculate max scroll based on content
		totalLines := len(m.urlsList) + 3 // +3 for title, empty line, and file path
		if m.urlsFilePath == "" {
			totalLines = len(m.urlsList) + 2 // +2 for title and empty line
		}
		availableHeight := m.height - 3 // -3 for title, empty line, and status bar
		if availableHeight < 1 {
			availableHeight = 1
		}
		maxScroll := totalLines - availableHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.urlsViewScroll < maxScroll {
			m.urlsViewScroll++
		}

	case "k", "up":
		if m.urlsViewScroll > 0 {
			m.urlsViewScroll--
		}

	case "ctrl+d":
		// Calculate max scroll based on content
		totalLines := len(m.urlsList) + 3
		if m.urlsFilePath == "" {
			totalLines = len(m.urlsList) + 2
		}
		availableHeight := m.height - 3
		if availableHeight < 1 {
			availableHeight = 1
		}
		maxScroll := totalLines - availableHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.urlsViewScroll += pageSize
		if m.urlsViewScroll > maxScroll {
			m.urlsViewScroll = maxScroll
		}

	case "ctrl+u":
		pageSize := m.height / 2
		if pageSize < 1 {
			pageSize = 5
		}
		m.urlsViewScroll -= pageSize
		if m.urlsViewScroll < 0 {
			m.urlsViewScroll = 0
		}
	}

	return m, nil
}

func (m Model) renderFeedInfo() string {
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - Feed Info"))
	b.WriteString("\n\n")

	// Build status bar
	viewKeys := GetViewKeys(FeedInfoView)
	globalHelp := "h: help | q: quit"
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)

	// Format feed information
	info := []struct {
		label string
		value string
	}{
		{"URL", m.currentFeed.Url},
		{"Title", m.currentFeed.Title},
		{"Description", m.currentFeed.Description},
		{"Last Updated", formatNullTime(m.currentFeed.LastUpdated)},
		{"Created At", formatNullTime(m.currentFeed.CreatedAt)},
		{"Feed Last Modified", formatNullString(m.currentFeed.LastModified)},
		{"Feed ETag", formatNullString(m.currentFeed.Etag)},
		{"Cache Control Max Age", formatNullInt64(m.currentFeed.CacheControlMaxAge)},
	}

	for _, item := range info {
		b.WriteString(fmt.Sprintf("%-23s: %s\n", item.label, item.value))
	}

	// Calculate padding to push status bar to bottom
	usedLines := len(info) + 3 // +3 for title and spacing
	padding := m.height - usedLines - 1
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))
	b.WriteString(statusBar)

	return b.String()
}

func (m Model) renderURLsView() string {
	// Build all content lines first
	var allLines []string

	// Add file path line if present
	if m.urlsFilePath != "" {
		allLines = append(allLines, m.getHelpStyle().Render("File: "+m.urlsFilePath))
		allLines = append(allLines, "") // Empty line after file path
	}

	// Add URLs or "No URLs found" message
	if len(m.urlsList) == 0 {
		allLines = append(allLines, "No URLs found.")
	} else {
		allLines = append(allLines, m.urlsList...)
	}

	// Calculate available height for content (height - title - status bar)
	availableHeight := m.height - 3 // -3 for title (2 lines) and status bar (1 line)
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine which lines to show based on scroll position
	start := m.urlsViewScroll
	if start >= len(allLines) {
		start = len(allLines) - 1
	}
	if start < 0 {
		start = 0
	}

	end := start + availableHeight
	if end > len(allLines) {
		end = len(allLines)
	}

	visibleLines := allLines[start:end]

	// Build final output
	var b strings.Builder
	b.WriteString(m.getTitleStyle().Render("🐐 NewsGoat - URLs"))
	b.WriteString("\n\n")

	for _, line := range visibleLines {
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Calculate padding to push status bar to bottom
	usedLines := len(visibleLines) + 3  // +3 for title and spacing
	padding := m.height - usedLines - 1 // -1 for status bar
	if padding < 0 {
		padding = 0
	}
	b.WriteString(strings.Repeat("\n", padding))

	// Show scroll indicator if there are more lines
	viewKeys := GetViewKeys(URLsView)
	viewHelp := FormatStatusBar(viewKeys.StatusBar)
	var statusBarText string
	if viewHelp != "" {
		statusBarText = globalHelp + " | " + viewHelp
	} else {
		statusBarText = globalHelp
	}
	statusBar := m.getHelpStyle().Render(statusBarText)
	if len(allLines) > availableHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d) ", start+1, end, len(allLines))
		b.WriteString(m.getHelpStyle().Render(scrollInfo))
	}
	b.WriteString(statusBar)

	return b.String()
}
