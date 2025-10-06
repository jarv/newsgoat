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

func TestGetURLType(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want URLType
	}{
		{"YouTube video", "https://www.youtube.com/watch?v=abc123", URLTypeYouTube},
		{"YouTube channel", "https://www.youtube.com/@LinusTechTips", URLTypeYouTube},
		{"GitHub tree", "https://github.com/owner/repo/tree/main/path", URLTypeGitHub},
		{"GitHub blob", "https://github.com/owner/repo/blob/main/file.go", URLTypeGitHub},
		{"GitHub commits", "https://github.com/owner/repo/commits/main", URLTypeGitHub},
		{"GitLab tree", "https://gitlab.com/group/project/-/tree/main/path", URLTypeGitLab},
		{"GitLab blob", "https://gitlab.com/group/project/-/blob/main/file.go", URLTypeGitLab},
		{"GitLab commits", "https://gitlab.com/group/project/-/commits/main", URLTypeGitLab},
		{"Generic URL", "https://example.com", URLTypeGeneric},
		{"RSS feed", "https://example.com/feed.xml", URLTypeGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetURLType(tt.url); got != tt.want {
				t.Errorf("GetURLType(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestDiscoverGitHubFeed(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "tree URL",
			url:     "https://github.com/owner/repo/tree/main/path/to/dir",
			want:    "https://github.com/owner/repo/commits/main/path/to/dir.atom",
			wantErr: false,
		},
		{
			name:    "blob URL",
			url:     "https://github.com/owner/repo/blob/develop/src/file.go",
			want:    "https://github.com/owner/repo/commits/develop/src/file.go.atom",
			wantErr: false,
		},
		{
			name:    "commits URL",
			url:     "https://github.com/owner/repo/commits/main",
			want:    "https://github.com/owner/repo/commits/main.atom",
			wantErr: false,
		},
		{
			name:    "commits URL with path",
			url:     "https://github.com/owner/repo/commits/feature-branch/internal/feeds",
			want:    "https://github.com/owner/repo/commits/feature-branch/internal/feeds.atom",
			wantErr: false,
		},
		{
			name:    "tree URL with branch only",
			url:     "https://github.com/torvalds/linux/tree/master",
			want:    "https://github.com/torvalds/linux/commits/master.atom",
			wantErr: false,
		},
		{
			name:    "blob URL with query string",
			url:     "https://github.com/owner/repo/blob/main/file.go?plain=1",
			want:    "https://github.com/owner/repo/commits/main/file.go.atom",
			wantErr: false,
		},
		{
			name:    "invalid URL - no branch/path",
			url:     "https://github.com/owner/repo",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL - wrong format",
			url:     "https://github.com/owner/repo/issues",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := discoverGitHubFeed(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoverGitHubFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("discoverGitHubFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiscoverGitLabFeed(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "tree URL",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/tree/master/app/models",
			want:    "https://gitlab.com/gitlab-org/gitlab/-/commits/master/app/models?format=atom",
			wantErr: false,
		},
		{
			name:    "blob URL",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/blob/master/app/models/user.rb",
			want:    "https://gitlab.com/gitlab-org/gitlab/-/commits/master/app/models/user.rb?format=atom",
			wantErr: false,
		},
		{
			name:    "commits URL",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/commits/master",
			want:    "https://gitlab.com/gitlab-org/gitlab/-/commits/master?format=atom",
			wantErr: false,
		},
		{
			name:    "commits URL with path",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/commits/main/internal/feeds",
			want:    "https://gitlab.com/gitlab-org/gitlab/-/commits/main/internal/feeds?format=atom",
			wantErr: false,
		},
		{
			name:    "tree URL with branch only",
			url:     "https://gitlab.com/gitlab-org/omnibus-gitlab/-/tree/master",
			want:    "https://gitlab.com/gitlab-org/omnibus-gitlab/-/commits/master?format=atom",
			wantErr: false,
		},
		{
			name:    "subgroup project",
			url:     "https://gitlab.com/group/subgroup/project/-/tree/main/src",
			want:    "https://gitlab.com/group/subgroup/project/-/commits/main/src?format=atom",
			wantErr: false,
		},
		{
			name:    "blob URL with query string",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/blob/master/Gemfile?ref_type=heads",
			want:    "https://gitlab.com/gitlab-org/gitlab/-/commits/master/Gemfile?format=atom",
			wantErr: false,
		},
		{
			name:    "invalid URL - no branch/path",
			url:     "https://gitlab.com/gitlab-org/gitlab",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid URL - wrong format",
			url:     "https://gitlab.com/gitlab-org/gitlab/-/issues",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := discoverGitLabFeed(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("discoverGitLabFeed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("discoverGitLabFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestIsLikelyFeedURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"Atom feed", "https://github.com/jarv.atom", true},
		{"XML feed", "https://example.com/feed.xml", true},
		{"RSS feed", "https://example.com/rss.rss", true},
		{"Atom with query", "https://gitlab.com/jarv.atom?feed_token=abc123", true},
		{"Regular URL", "https://example.com/page", false},
		{"GitHub tree URL", "https://github.com/owner/repo/tree/main", false},
		{"GitLab blob URL", "https://gitlab.com/group/project/-/blob/main/file.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLikelyFeedURL(tt.url); got != tt.want {
				t.Errorf("isLikelyFeedURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
