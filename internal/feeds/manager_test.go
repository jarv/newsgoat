package feeds

import (
	"os"
	"strings"
	"testing"
)

func TestAddLinkMarkersToHTML(t *testing.T) {
	// Read the fixture file
	htmlContent, err := os.ReadFile("testdata/gitlab_mr.html")
	if err != nil {
		t.Fatalf("Failed to read fixture file: %v", err)
	}

	manager := &Manager{}

	// Add link markers to HTML
	markedHTML, links := manager.AddLinkMarkersToHTML(string(htmlContent))

	// Print for debugging
	t.Logf("Found %d links:", len(links))
	for i, link := range links {
		t.Logf("[%d] %s", i+1, link)
	}

	// Convert to markdown
	markdown := manager.ConvertHTMLToMarkdown(markedHTML)

	// Check that all links have markers in the markdown
	// The markers are in the format \[N\] due to markdown escaping
	for i := range links {
		marker := "\\[" + string(rune('0'+i+1)) + "\\]"
		if !strings.Contains(markdown, marker) {
			// Also try without escaping
			markerUnescaped := "[" + string(rune('0'+i+1)) + "]"
			if !strings.Contains(markdown, markerUnescaped) {
				t.Logf("Markdown output:\n%s", markdown)
				t.Errorf("Link marker [%d] not found in markdown for link: %s", i+1, links[i])
			}
		}
	}

	// Verify expected links are present
	expectedLinks := []string{
		"https://sources.debian.org/src/unzip/6.0-29/debian/rules",
		"https://sources.debian.org/patches/unzip/6.0-29/",
		"https://gitlab.com/gitlab-org/omnibus-gitlab/-/issues/9225",
		"https://gitlab.com/gitlab-org/omnibus-gitlab/blob/master/CONTRIBUTING.md#definition-of-done",
		"https://gitlab.com/gitlab-org/omnibus-gitlab/-/issues?label_name=workflow%3A%3Aready+for+review",
		"https://about.gitlab.com/handbook/engineering/development/enablement/systems/distribution/merge_requests.html",
		"https://dev.gitlab.org/gitlab/omnibus-gitlab",
		"https://gitlab.com/gitlab-org/gitlab-qa",
		"https://gitlab.com/gitlab-org/charts/gitlab",
	}

	if len(links) != len(expectedLinks) {
		t.Errorf("Expected %d links, got %d", len(expectedLinks), len(links))
	}

	for i, expected := range expectedLinks {
		if i >= len(links) {
			t.Errorf("Missing expected link: %s", expected)
			continue
		}
		if links[i] != expected {
			t.Errorf("Link %d: expected %s, got %s", i+1, expected, links[i])
		}
	}
}

func TestExtractLinks(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "simple anchor",
			html: `<a href="https://example.com">link</a>`,
			expected: []string{"https://example.com"},
		},
		{
			name: "anchor with attributes",
			html: `<a href="https://example.com" class="test" target="_blank">link</a>`,
			expected: []string{"https://example.com"},
		},
		{
			name: "multiple links",
			html: `<a href="https://example.com">link1</a> <a href="https://test.com">link2</a>`,
			expected: []string{"https://example.com", "https://test.com"},
		},
		{
			name: "duplicate links",
			html: `<a href="https://example.com">link1</a> <a href="https://example.com">link2</a>`,
			expected: []string{"https://example.com"},
		},
		{
			name: "plain text URL",
			html: `Check out https://example.com for more info`,
			expected: []string{"https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links := manager.ExtractLinks(tt.html)
			if len(links) != len(tt.expected) {
				t.Errorf("Expected %d links, got %d", len(tt.expected), len(links))
				return
			}
			for i, expected := range tt.expected {
				if links[i] != expected {
					t.Errorf("Link %d: expected %s, got %s", i, expected, links[i])
				}
			}
		})
	}
}
