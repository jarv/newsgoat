package opml

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jarv/newsgoat/internal/config"
)

type document struct {
	XMLName xml.Name `xml:"opml"`
	Body    body     `xml:"body"`
}

type body struct {
	Outlines []outline `xml:"outline"`
}

type outline struct {
	Text     string    `xml:"text,attr"`
	Title    string    `xml:"title,attr"`
	XMLURL   string    `xml:"xmlUrl,attr"`
	Outlines []outline `xml:"outline"`
}

func (o outline) displayName() string {
	if o.Title != "" {
		return o.Title
	}
	return o.Text
}

func Parse(data []byte) ([]config.URLEntry, error) {
	var doc document
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid OPML: %w", err)
	}

	var entries []config.URLEntry
	seen := make(map[string]bool)

	var walk func(outlines []outline, folder string)
	walk = func(outlines []outline, folder string) {
		for _, o := range outlines {
			if o.XMLURL != "" {
				u := strings.TrimSpace(o.XMLURL)
				if u == "" || seen[u] {
					continue
				}
				seen[u] = true
				var folders []string
				if folder != "" {
					folders = []string{folder}
				}
				entries = append(entries, config.URLEntry{
					URL:     u,
					Folders: folders,
				})
			}
			if len(o.Outlines) > 0 {
				name := o.displayName()
				if o.XMLURL == "" && name != "" {
					walk(o.Outlines, name)
				} else {
					walk(o.Outlines, folder)
				}
			}
		}
	}

	walk(doc.Body.Outlines, "")

	if len(entries) == 0 {
		return nil, fmt.Errorf("no feeds found in OPML file")
	}

	return entries, nil
}

func ReadSource(source string) ([]byte, error) {
	u, err := url.Parse(source)
	if err == nil && u.Host != "" && u.Scheme != "" {
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch OPML from URL: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch OPML: HTTP %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read OPML file: %w", err)
	}
	return data, nil
}
