package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/tasks"
	"github.com/jarv/newsgoat/internal/ui"
	"github.com/jarv/newsgoat/internal/version"
)

//go:embed sql/schema.sql
var schemaSQL string

var logger *slog.Logger

func setupLogging(queries *database.Queries, debug bool) {
	slogHandler := logging.NewDatabaseHandlerWithDebug(queries, debug)
	logger = slog.New(slogHandler)

	// Set the global logger for other packages
	logging.SetLogger(logger)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: newsgoat [options] [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  add <url>    Add a feed URL to the URLs file\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_FEED_TOKEN   Access token for private GitHub repository feeds\n")
		fmt.Fprintf(os.Stderr, "  GITLAB_FEED_TOKEN   Access token for private GitLab repository feeds\n")
	}

	var feedTest = flag.Bool("feedTest", false, "Run feed test harness server")
	var showVersion = flag.Bool("version", false, "Show version information")
	var debug = flag.Bool("debug", false, "Enable debug logging")
	var urlFile = flag.String("u", "", "Path to URL file (overrides default location)")
	flag.StringVar(urlFile, "urlFile", "", "Path to URL file (overrides default location)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion())
		return
	}

	if *feedTest {
		if err := runFeedTestHarness(); err != nil {
			fmt.Fprintf(os.Stderr, "1Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check for subcommands
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "add":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "Error: 'add' command requires a URL argument\n")
				fmt.Fprintf(os.Stderr, "Usage: newsgoat add <url>\n")
				os.Exit(1)
			}
			if err := addURL(args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", args[0])
			os.Exit(1)
		}
	}

	if err := run(*urlFile, *debug); err != nil {
		fmt.Fprintf(os.Stderr, "2Error: %v\n", err)
		os.Exit(1)
	}
}

func addURL(urlArg string) error {
	// Try to discover the feed URL
	fmt.Printf("Discovering feed URL from: %s\n", urlArg)
	feedURL, err := discovery.DiscoverFeed(urlArg)
	if err != nil {
		return fmt.Errorf("failed to discover feed: %w", err)
	}

	if feedURL != urlArg {
		fmt.Printf("Discovered feed URL: %s\n", feedURL)
	}

	// Add the URL to the URLs file
	if err := config.AddURL(feedURL); err != nil {
		return fmt.Errorf("failed to add URL to file: %w", err)
	}

	fmt.Printf("Successfully added feed: %s\n", feedURL)
	return nil
}

func run(urlFile string, debug bool) error {
	// Initialize database first
	db, queries, err := database.InitDBWithSchema(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Load configuration from database
	cfg, err := config.LoadConfig(queries)
	if err != nil {
		fmt.Printf("Failed to load config, using defaults: %v\n", err)
		cfg = config.GetDefaultConfig()
	}

	// Setup logging after database is initialized
	setupLogging(queries, debug)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("Error closing database", "error", closeErr)
		}
	}()

	feedManager := feeds.NewManager(db, queries)

	// Create and start task manager
	taskManager := tasks.NewManager(cfg.ReloadConcurrency)
	ctx := context.Background()
	if err := taskManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task manager: %w", err)
	}
	defer func() {
		if stopErr := taskManager.Stop(); stopErr != nil {
			logger.Debug("Task manager already stopped", "error", stopErr)
		}
	}()

	// Register feed refresh handler
	feedRefreshHandler := tasks.NewFeedRefreshHandler(feedManager)
	if err := taskManager.RegisterHandler(feedRefreshHandler); err != nil {
		return fmt.Errorf("failed to register feed refresh handler: %w", err)
	}

	if err := config.CreateSampleURLsFile(); err != nil {
		logger.Warn("Failed to create sample URLs file", "error", err)
	}

	// Get URLs file path
	urlsPath, err := config.GetURLsFilePath()
	if err != nil {
		logger.Warn("Failed to get URLs file path", "error", err)
		urlsPath = ""
	}

	var urlEntries []config.URLEntry
	if urlFile != "" {
		var readErr error
		urlEntries, readErr = config.ReadURLsFileFromPath(urlFile)
		if readErr != nil {
			return fmt.Errorf("failed to read URLs file: %w", readErr)
		}
		urlsPath = urlFile
	} else {
		var readErr error
		urlEntries, readErr = config.ReadURLsFile()
		if readErr != nil {
			return fmt.Errorf("failed to read URLs file: %w", readErr)
		}
	}

	if err := syncFeedsWithURLsFile(feedManager, queries, urlEntries); err != nil {
		logger.Warn("Failed to sync feeds with URLs file", "error", err)
	}

	model := ui.NewModel(feedManager, taskManager, queries, cfg)
	model.SetURLsFilePath(urlsPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func syncFeedsWithURLsFile(feedManager *feeds.Manager, queries *database.Queries, urlEntries []config.URLEntry) error {
	// Get all feeds from database (including hidden ones)
	allFeeds, err := feedManager.GetAllFeeds()
	if err != nil {
		return fmt.Errorf("failed to get all feeds: %w", err)
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
				logger.Warn("Failed to hide feed", "url", feed.Url, "error", err)
			}
		}
	}

	// Show/Add feeds that are in URLs file and update folders
	ctx := context.Background()
	for _, entry := range urlEntries {
		var feedID int64
		if urlsFromDBSet[entry.URL] {
			// Feed exists in DB, make sure it's visible
			if err := feedManager.ShowFeedByURL(entry.URL); err != nil {
				logger.Warn("Failed to show feed", "url", entry.URL, "error", err)
				continue
			}
			// Get feed ID
			feed, err := queries.GetFeedByURL(ctx, entry.URL)
			if err != nil {
				logger.Warn("Failed to get feed by URL", "url", entry.URL, "error", err)
				continue
			}
			feedID = feed.ID
		} else {
			// Feed doesn't exist, add it without fetching
			if err := feedManager.AddFeedWithoutFetching(entry.URL); err != nil {
				logger.Warn("Failed to add feed", "url", entry.URL, "error", err)
				continue
			}
			// Get the newly created feed ID
			feed, err := queries.GetFeedByURL(ctx, entry.URL)
			if err != nil {
				logger.Warn("Failed to get newly created feed", "url", entry.URL, "error", err)
				continue
			}
			feedID = feed.ID
		}

		// Update folders for this feed
		// First, delete existing folders
		if err := queries.DeleteFeedFolders(ctx, feedID); err != nil {
			logger.Warn("Failed to delete old folders", "feed_id", feedID, "error", err)
		}

		// Then add new folders
		for _, folder := range entry.Folders {
			if err := queries.AddFeedFolder(ctx, database.AddFeedFolderParams{
				FeedID:     feedID,
				FolderName: folder,
			}); err != nil {
				logger.Warn("Failed to add folder", "feed_id", feedID, "folder", folder, "error", err)
			}
		}
	}

	return nil
}
