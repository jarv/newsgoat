package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jarv/newsgoat/internal/config"
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/discovery"
	"github.com/jarv/newsgoat/internal/feeds"
	grpcclient "github.com/jarv/newsgoat/internal/grpc/client"
	pb "github.com/jarv/newsgoat/internal/grpc/gen/newsgoat/v1"
	grpcserver "github.com/jarv/newsgoat/internal/grpc/server"
	"github.com/jarv/newsgoat/internal/logging"
	"github.com/jarv/newsgoat/internal/tasks"
	"github.com/jarv/newsgoat/internal/ui"
	"github.com/jarv/newsgoat/internal/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
		fmt.Fprintf(os.Stderr, "  add <url>    Add a feed URL to the database\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_FEED_TOKEN   Access token for private GitHub repository feeds\n")
		fmt.Fprintf(os.Stderr, "  GITLAB_FEED_TOKEN   Access token for private GitLab repository feeds\n")
		fmt.Fprintf(os.Stderr, "  NEWSGOAT_API_KEY    API key for client/server authentication\n")
	}

	var (
		feedTest   = flag.Bool("feedTest", false, "Run feed test harness server")
		showVersion = flag.Bool("version", false, "Show version information")
		debug      = flag.Bool("debug", false, "Enable debug logging")
		mode       = flag.String("mode", "standalone", "Operating mode: standalone, client, or server")
		serverURL  = flag.String("server-url", "", "gRPC server URL for client mode (e.g., localhost:50051)")
		serverPort = flag.Int("server-port", 50051, "gRPC server port for server mode")
		dbPath     = flag.String("db", "", "Path to SQLite database file (server mode only)")
	)
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

	// Handle mode selection
	switch *mode {
	case "server":
		// Server mode - run gRPC server
		apiKey := os.Getenv("NEWSGOAT_API_KEY")
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: NEWSGOAT_API_KEY environment variable is required for server mode\n")
			os.Exit(1)
		}
		if err := runServer(*serverPort, apiKey, *dbPath, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	case "client":
		// Client mode - connect to gRPC server
		if *serverURL == "" {
			fmt.Fprintf(os.Stderr, "Error: --server-url is required for client mode\n")
			flag.Usage()
			os.Exit(1)
		}
		apiKey := os.Getenv("NEWSGOAT_API_KEY")
		if err := runClient(*serverURL, apiKey, *debug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return

	case "standalone":
		// Standalone mode - local database (default)
		// Continue to command handling below

	default:
		fmt.Fprintf(os.Stderr, "Error: invalid mode '%s'. Must be: standalone, client, or server\n", *mode)
		os.Exit(1)
	}

	// Check for subcommands (standalone mode only)
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

	if err := run(*debug); err != nil {
		fmt.Fprintf(os.Stderr, "2Error: %v\n", err)
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

func run(debug bool) error {
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
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// runServer starts the gRPC server
func runServer(port int, apiKey, dbPath string, debug bool) error {
	// Initialize database
	var db, queries, err = database.InitDBWithSchema(schemaSQL)
	if dbPath != "" {
		db, queries, err = database.InitDBWithSchemaAtPath(schemaSQL, dbPath)
	}
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Error("Error closing database", "error", closeErr)
		}
	}()

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup logging
	setupLogging(debug)
	logger.Info("Starting newsgoat gRPC server", "version", version.GetVersion(), "port", port)

	// Create feed manager
	manager := feeds.NewManager(db, queries)

	// Create TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	// Create gRPC server with auth interceptor
	authInterceptor := grpcserver.NewAuthInterceptor(apiKey)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
		grpc.StreamInterceptor(authInterceptor.Stream()),
	)

	// Register services
	feedService := grpcserver.NewFeedService(manager)
	settingsService := grpcserver.NewSettingsService(queries)

	pb.RegisterFeedServiceServer(grpcServer, feedService)
	pb.RegisterSettingsServiceServer(grpcServer, settingsService)

	// Register reflection service for debugging (grpcurl, etc.)
	reflection.Register(grpcServer)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		logger.Info("gRPC server listening", "address", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig)
		grpcServer.GracefulStop()
		logger.Info("Server shutdown complete")
		return nil
	}
}

// runClient starts the TUI in client mode connected to a gRPC server
func runClient(serverURL, apiKey string, debug bool) error {
	fmt.Printf("Connecting to gRPC server at %s...\n", serverURL)

	// Create gRPC client
	client, err := grpcclient.NewClient(serverURL, apiKey)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	fmt.Println("Connected successfully!")

	// Setup in-memory logging (no local database needed in client mode)
	setupLogging(debug)

	// Use default configuration in client mode
	cfg := config.GetDefaultConfig()

	// Create task manager (but don't use it for feed refreshes in client mode)
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

	// Note: We pass the client as the manager interface to the UI
	// The UI will call the client methods, which forward to the gRPC server
	// No local database needed in client mode - queries is nil
	model := ui.NewModel(client, taskManager, nil, cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
