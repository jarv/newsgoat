package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	"github.com/jarv/newsgoat/internal/logging"
	mcpserver "github.com/jarv/newsgoat/internal/mcp"
	"github.com/jarv/newsgoat/internal/opml"
	"github.com/jarv/newsgoat/internal/tasks"
	"github.com/jarv/newsgoat/internal/ui"
	"github.com/jarv/newsgoat/internal/version"
)

//go:embed sql/schema.sql
var schemaSQL string

var logger *slog.Logger

func setupLogging(debug bool) {
	// Use in-memory logging handler (stores last 1000 messages)
	memoryHandler := logging.NewMemoryHandlerWithDebug(1000, debug)
	logger = slog.New(memoryHandler)

	// Set the global logger and memory handler for other packages
	logging.SetLogger(logger)
	logging.SetMemoryHandler(memoryHandler)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: newsgoat [options] [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  add <url>          Add a feed URL to the database\n")
		fmt.Fprintf(os.Stderr, "  import <file|url>  Import feeds from an OPML file\n")
		fmt.Fprintf(os.Stderr, "  mcp-server         Start MCP server (stdio transport)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_FEED_TOKEN   Access token for private GitHub repository feeds\n")
		fmt.Fprintf(os.Stderr, "  GITLAB_FEED_TOKEN   Access token for private GitLab repository feeds\n")
	}

	var (
		feedTest    = flag.Bool("feedTest", false, "Run feed test harness server")
		showVersion = flag.Bool("version", false, "Show version information")
		debug       = flag.Bool("debug", false, "Enable debug logging")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version.GetVersion())
		return
	}

	if *feedTest {
		if err := runFeedTestHarness(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
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
		case "import":
			if len(args) < 2 {
				fmt.Fprintf(os.Stderr, "Error: 'import' command requires a file path or URL\n")
				fmt.Fprintf(os.Stderr, "Usage: newsgoat import <file|url>\n")
				os.Exit(1)
			}
			if err := importOPML(args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "mcp-server":
			if err := runMCPServer(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", args[0])
			os.Exit(1)
		}
	}

	if err := run(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func addURL(urlArg string) error {
	// Initialize database
	db, queries, err := database.InitDBWithSchema(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	feedManager := feeds.NewManager(db, queries)

	// Try to discover the feed URL
	fmt.Printf("Discovering feed URL from: %s\n", urlArg)
	feedURL, err := discovery.DiscoverFeed(urlArg)
	if err != nil {
		return fmt.Errorf("failed to discover feed: %w", err)
	}

	if feedURL != urlArg {
		fmt.Printf("Discovered feed URL: %s\n", feedURL)
	}

	// Add the feed to the database
	if err := feedManager.AddFeedWithoutFetching(feedURL); err != nil {
		return fmt.Errorf("failed to add feed to database: %w", err)
	}

	fmt.Printf("Successfully added feed: %s\n", feedURL)
	return nil
}

func importOPML(source string) error {
	data, err := opml.ReadSource(source)
	if err != nil {
		return err
	}

	entries, err := opml.Parse(data)
	if err != nil {
		return err
	}

	db, queries, err := database.InitDBWithSchema(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	feedManager := feeds.NewManager(db, queries)

	allFeeds, err := feedManager.GetAllFeeds()
	if err != nil {
		return fmt.Errorf("failed to get existing feeds: %w", err)
	}
	existingURLs := make(map[string]database.Feed, len(allFeeds))
	for _, f := range allFeeds {
		existingURLs[f.Url] = f
	}

	var added, skipped, unhidden int
	for _, entry := range entries {
		if existing, ok := existingURLs[entry.URL]; ok {
			if !existing.Visible {
				if err := feedManager.ShowFeedByURL(entry.URL); err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: failed to unhide %s: %v\n", entry.URL, err)
				} else {
					unhidden++
				}
			} else {
				skipped++
			}
			continue
		}
		if err := feedManager.AddFeedWithoutFetching(entry.URL); err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: failed to add %s: %v\n", entry.URL, err)
			continue
		}
		if len(entry.Folders) > 0 {
			feed, err := feedManager.GetFeedByURL(entry.URL)
			if err == nil {
				for _, folder := range entry.Folders {
					_ = feedManager.AddFeedFolder(feed.ID, folder)
				}
			}
		}
		added++
	}

	fmt.Printf("Importing from: %s\n", source)
	fmt.Printf("  Found %d feeds in OPML file\n", len(entries))
	if skipped > 0 {
		fmt.Printf("  %d already exist (skipped)\n", skipped)
	}
	if unhidden > 0 {
		fmt.Printf("  %d were hidden and are now visible\n", unhidden)
	}
	fmt.Printf("  %d new feeds added\n", added)

	return nil
}

func runMCPServer() error {
	db, queries, err := database.InitDBWithSchema(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return mcpserver.Run(queries)
}

func run(debug bool) error {
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
	setupLogging(debug)
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

	model := ui.NewModel(feedManager, taskManager, queries, cfg)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
