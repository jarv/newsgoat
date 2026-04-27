package opml

import (
	"testing"
)

func TestParse_FlatFeeds(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Feed A" type="rss" xmlUrl="https://a.com/feed"/>
    <outline text="Feed B" type="rss" xmlUrl="https://b.com/feed"/>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].URL != "https://a.com/feed" {
		t.Errorf("unexpected URL: %s", entries[0].URL)
	}
	if len(entries[0].Folders) != 0 {
		t.Errorf("expected no folders, got %v", entries[0].Folders)
	}
}

func TestParse_NestedFolders(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline title="Tech" text="Tech">
      <outline text="Ars" type="rss" xmlUrl="https://ars.com/feed"/>
      <outline text="Verge" type="rss" xmlUrl="https://verge.com/feed"/>
    </outline>
    <outline title="News">
      <outline text="BBC" type="rss" xmlUrl="https://bbc.com/feed"/>
    </outline>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Folders[0] != "Tech" {
		t.Errorf("expected folder Tech, got %v", entries[0].Folders)
	}
	if entries[2].Folders[0] != "News" {
		t.Errorf("expected folder News, got %v", entries[2].Folders)
	}
}

func TestParse_DeepNesting(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline title="Level1">
      <outline title="Level2">
        <outline text="Deep" type="rss" xmlUrl="https://deep.com/feed"/>
      </outline>
    </outline>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Folders[0] != "Level2" {
		t.Errorf("expected folder Level2 (immediate parent), got %v", entries[0].Folders)
	}
}

func TestParse_DuplicateURLs(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Feed A" xmlUrl="https://a.com/feed"/>
    <outline text="Feed A Copy" xmlUrl="https://a.com/feed"/>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after dedup, got %d", len(entries))
	}
}

func TestParse_MissingXMLURL(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline title="Empty Folder"/>
    <outline text="Feed" xmlUrl="https://a.com/feed"/>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestParse_PrefersTitleOverText(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline title="My Title" text="My Text">
      <outline text="Feed" xmlUrl="https://a.com/feed"/>
    </outline>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries[0].Folders[0] != "My Title" {
		t.Errorf("expected folder 'My Title', got %v", entries[0].Folders)
	}
}

func TestParse_FallsBackToText(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Only Text">
      <outline text="Feed" xmlUrl="https://a.com/feed"/>
    </outline>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries[0].Folders[0] != "Only Text" {
		t.Errorf("expected folder 'Only Text', got %v", entries[0].Folders)
	}
}

func TestParse_NoFeeds(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline title="Empty"/>
  </body>
</opml>`)
	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error for OPML with no feeds")
	}
}

func TestParse_InvalidXML(t *testing.T) {
	_, err := Parse([]byte("not xml"))
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestParse_AtomType(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Atom Feed" type="atom" xmlUrl="https://atom.com/feed"/>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestParse_MixedFolderAndRootFeeds(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.0">
  <body>
    <outline text="Root Feed" xmlUrl="https://root.com/feed"/>
    <outline title="Tech">
      <outline text="Nested" xmlUrl="https://nested.com/feed"/>
    </outline>
  </body>
</opml>`)
	entries, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if len(entries[0].Folders) != 0 {
		t.Errorf("root feed should have no folders, got %v", entries[0].Folders)
	}
	if entries[1].Folders[0] != "Tech" {
		t.Errorf("nested feed should have folder Tech, got %v", entries[1].Folders)
	}
}
