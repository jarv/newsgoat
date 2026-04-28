package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jarv/newsgoat/internal/database"
	"github.com/mark3labs/mcp-go/mcp"
)

type handlers struct {
	queries *database.Queries
}

func parseSince(s string, defaultDuration time.Duration) (time.Time, error) {
	if s == "" {
		return time.Now().Add(-defaultDuration), nil
	}
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return time.Time{}, fmt.Errorf("invalid duration: %q", s)
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return time.Time{}, fmt.Errorf("invalid duration: %q", s)
	}
	switch unit {
	case 'h':
		return time.Now().Add(-time.Duration(n) * time.Hour), nil
	case 'd':
		return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
	default:
		return time.Time{}, fmt.Errorf("invalid duration unit %q, use h or d", string(unit))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatArticleRow(id int64, title, description, link, feedTitle string, published sql.NullTime) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%d] %s\n", id, title)
	fmt.Fprintf(&b, "  Feed: %s\n", feedTitle)
	if published.Valid {
		fmt.Fprintf(&b, "  Date: %s\n", published.Time.Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(&b, "  Link: %s\n", link)
	desc := truncate(strings.TrimSpace(description), 200)
	if desc != "" {
		fmt.Fprintf(&b, "  %s\n", desc)
	}
	return b.String()
}

func (h *handlers) searchArticles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		return mcp.NewToolResultError("query parameter is required"), nil
	}
	since, err := parseSince(req.GetString("since", ""), 7*24*time.Hour)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	limit := int64(req.GetInt("limit", 50))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	queryParam := sql.NullString{String: query, Valid: true}
	rows, err := h.queries.SearchArticlesSince(ctx, database.SearchArticlesSinceParams{
		Published: sql.NullTime{Time: since, Valid: true},
		Column2:   queryParam,
		Column3:   queryParam,
		Column4:   queryParam,
		Limit:     limit,
	})
	if err != nil {
		return mcp.NewToolResultError("search failed: " + err.Error()), nil
	}

	if len(rows) == 0 {
		return mcp.NewToolResultText("No articles found matching the query."), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d articles:\n\n", len(rows))
	for _, r := range rows {
		b.WriteString(formatArticleRow(r.ID, r.Title, r.Description, r.Link, r.FeedTitle, r.Published))
		b.WriteString("\n")
	}
	return mcp.NewToolResultText(b.String()), nil
}

func (h *handlers) listRecentArticles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	since, err := parseSince(req.GetString("since", ""), 24*time.Hour)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	limit := int64(req.GetInt("limit", 50))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	folder := req.GetString("folder", "")
	feed := req.GetString("feed", "")

	type row struct {
		ID          int64
		Title       string
		Description string
		Link        string
		Published   sql.NullTime
		FeedTitle   string
	}

	var rows []row

	if folder != "" {
		dbRows, err := h.queries.ListRecentArticlesByFolder(ctx, database.ListRecentArticlesByFolderParams{
			Published:  sql.NullTime{Time: since, Valid: true},
			FolderName: folder,
			Limit:      limit,
		})
		if err != nil {
			return mcp.NewToolResultError("query failed: " + err.Error()), nil
		}
		for _, r := range dbRows {
			rows = append(rows, row{r.ID, r.Title, r.Description, r.Link, r.Published, r.FeedTitle})
		}
	} else if feed != "" {
		dbRows, err := h.queries.ListRecentArticlesByFeed(ctx, database.ListRecentArticlesByFeedParams{
			Published: sql.NullTime{Time: since, Valid: true},
			Column2:   sql.NullString{String: feed, Valid: true},
			Limit:     limit,
		})
		if err != nil {
			return mcp.NewToolResultError("query failed: " + err.Error()), nil
		}
		for _, r := range dbRows {
			rows = append(rows, row{r.ID, r.Title, r.Description, r.Link, r.Published, r.FeedTitle})
		}
	} else {
		dbRows, err := h.queries.ListRecentArticles(ctx, database.ListRecentArticlesParams{
			Published: sql.NullTime{Time: since, Valid: true},
			Limit:     limit,
		})
		if err != nil {
			return mcp.NewToolResultError("query failed: " + err.Error()), nil
		}
		for _, r := range dbRows {
			rows = append(rows, row{r.ID, r.Title, r.Description, r.Link, r.Published, r.FeedTitle})
		}
	}

	if len(rows) == 0 {
		return mcp.NewToolResultText("No recent articles found."), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d articles:\n\n", len(rows))
	for _, r := range rows {
		b.WriteString(formatArticleRow(r.ID, r.Title, r.Description, r.Link, r.FeedTitle, r.Published))
		b.WriteString("\n")
	}
	return mcp.NewToolResultText(b.String()), nil
}

func (h *handlers) getArticle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireInt("id")
	if err != nil {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	article, err := h.queries.GetArticleByID(ctx, int64(id))
	if err != nil {
		return mcp.NewToolResultError("article not found"), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n", article.Title)
	fmt.Fprintf(&b, "Feed: %s\n", article.FeedTitle)
	if article.Published.Valid {
		fmt.Fprintf(&b, "Date: %s\n", article.Published.Time.Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(&b, "Link: %s\n", article.Link)
	b.WriteString("\n")
	content := article.Content
	if content == "" {
		content = article.Description
	}
	b.WriteString(content)
	return mcp.NewToolResultText(b.String()), nil
}

func (h *handlers) listFeeds(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	feeds, err := h.queries.ListVisibleFeedsWithFolders(ctx)
	if err != nil {
		return mcp.NewToolResultError("failed to list feeds: " + err.Error()), nil
	}

	if len(feeds) == 0 {
		return mcp.NewToolResultText("No feeds configured."), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d feeds:\n\n", len(feeds))
	for _, f := range feeds {
		folders, _ := h.queries.GetFeedFolders(ctx, f.ID)
		folderStr := ""
		if len(folders) > 0 {
			folderStr = " [" + strings.Join(folders, ", ") + "]"
		}
		fmt.Fprintf(&b, "- %s (%d/%d)%s\n  %s\n", f.Title, f.UnreadItems, f.TotalItems, folderStr, f.Url)
	}
	return mcp.NewToolResultText(b.String()), nil
}
