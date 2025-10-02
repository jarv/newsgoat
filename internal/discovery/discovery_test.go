package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFeedFromHTML(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		baseURL     string
		wantURL     string
		wantErr     bool
	}{
		{
			name:     "RSS feed with relative URL",
			filename: "blog_with_rss.html",
			baseURL:  "https://example.com/blog",
			wantURL:  "https://example.com/feed.xml",
			wantErr:  false,
		},
		{
			name:     "Atom feed with absolute URL",
			filename: "blog_with_atom.html",
			baseURL:  "https://example.com",
			wantURL:  "https://example.com/atom.xml",
			wantErr:  false,
		},
		{
			name:     "No feed available",
			filename: "no_feed.html",
			baseURL:  "https://example.com",
			wantURL:  "",
			wantErr:  true,
		},
		{
			name:     "Multiple feeds (returns first)",
			filename: "multiple_feeds.html",
			baseURL:  "https://example.com",
			wantURL:  "https://example.com/rss.xml",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			gotURL, err := DiscoverFeedFromHTML(string(content), tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiscoverFeedFromHTML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("DiscoverFeedFromHTML() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func TestIsFeedContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"RSS XML", "application/rss+xml", true},
		{"Atom XML", "application/atom+xml", true},
		{"Generic XML", "application/xml", true},
		{"Text XML", "text/xml", true},
		{"HTML", "text/html", false},
		{"JSON", "application/json", false},
		{"With charset", "application/rss+xml; charset=utf-8", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFeedContentType(tt.contentType); got != tt.want {
				t.Errorf("isFeedContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		relativeURL string
		want        string
	}{
		{
			name:        "Absolute path",
			baseURL:     "https://example.com/blog/post",
			relativeURL: "/feed.xml",
			want:        "https://example.com/feed.xml",
		},
		{
			name:        "Relative path",
			baseURL:     "https://example.com/blog/",
			relativeURL: "feed.xml",
			want:        "https://example.com/blog/feed.xml",
		},
		{
			name:        "Relative path from page",
			baseURL:     "https://example.com/blog/post.html",
			relativeURL: "feed.xml",
			want:        "https://example.com/blog/feed.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveURL(tt.baseURL, tt.relativeURL); got != tt.want {
				t.Errorf("resolveURL(%q, %q) = %v, want %v", tt.baseURL, tt.relativeURL, got, tt.want)
			}
		})
	}
}

func TestExtractYouTubeChannelID(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "YouTube video page",
			filename: "youtube_video.html",
			want:     "UC59ZRYCHev_IqjUhremZ8Tg",
			wantErr:  false,
		},
		{
			name:     "YouTube channel page",
			filename: "youtube_channel.html",
			want:     "UCXuqSBlHAE6Xw-yeJA0Tunw",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join("testdata", tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			got, err := ExtractYouTubeChannelID(string(content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractYouTubeChannelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractYouTubeChannelID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsYouTubeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"YouTube video", "https://www.youtube.com/watch?v=abc123", true},
		{"YouTube channel", "https://www.youtube.com/channel/UCabc123", true},
		{"YouTube short URL", "https://youtu.be/abc123", true},
		{"Not YouTube", "https://example.com", false},
		{"Not YouTube with youtube in path", "https://example.com/youtube", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isYouTubeURL(tt.url); got != tt.want {
				t.Errorf("isYouTubeURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
