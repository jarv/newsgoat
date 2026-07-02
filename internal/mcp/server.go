package mcp

import (
	"github.com/jarv/newsgoat/internal/database"
	"github.com/jarv/newsgoat/internal/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Run(queries *database.Queries) error {
	s := server.NewMCPServer(
		"newsgoat",
		version.GetVersion(),
		server.WithToolCapabilities(true),
	)

	h := &handlers{queries: queries}

	s.AddTool(
		mcp.NewTool("search_articles",
			mcp.WithDescription("Search article titles, descriptions, and content across all feeds"),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
			mcp.WithString("since", mcp.Description("Time window, e.g. 24h, 7d, 30d. Defaults to 7d")),
			mcp.WithNumber("limit", mcp.Description("Max results to return. Defaults to 50")),
			mcp.WithString("after", mcp.Description("Cursor for pagination. Use the next_cursor value from a previous response to fetch the next page.")),
		),
		h.searchArticles,
	)

	s.AddTool(
		mcp.NewTool("list_recent_articles",
			mcp.WithDescription("List recent articles, optionally filtered by folder or feed name"),
			mcp.WithString("since", mcp.Description("Time window, e.g. 24h, 7d, 30d. Defaults to 24h")),
			mcp.WithString("folder", mcp.Description("Filter by folder name (exact match)")),
			mcp.WithString("feed", mcp.Description("Filter by feed title (substring match)")),
			mcp.WithNumber("limit", mcp.Description("Max results to return. Defaults to 50")),
			mcp.WithString("after", mcp.Description("Cursor for pagination. Use the next_cursor value from a previous response to fetch the next page.")),
		),
		h.listRecentArticles,
	)

	s.AddTool(
		mcp.NewTool("get_article",
			mcp.WithDescription("Get full content of a specific article by ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("Article ID")),
		),
		h.getArticle,
	)

	s.AddTool(
		mcp.NewTool("list_feeds",
			mcp.WithDescription("List all visible feeds with their folders and unread/total counts"),
		),
		h.listFeeds,
	)

	return server.ServeStdio(s)
}
