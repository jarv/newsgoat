package filters

import (
	"testing"
)

func TestParseJSON_Valid(t *testing.T) {
	input := `{
		"global": {"URL": ["golang"], "Title": ["!test"]},
		"folder:Tech": {"Description": ["news"]},
		"feed:42": {"URL": ["!/shorts/"]}
	}`
	fm, err := ParseJSON([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(fm))
	}
	if len(fm["global"].URL) != 1 || fm["global"].URL[0] != "golang" {
		t.Errorf("unexpected global URL: %v", fm["global"].URL)
	}
	if len(fm["global"].Title) != 1 || fm["global"].Title[0] != "!test" {
		t.Errorf("unexpected global Title: %v", fm["global"].Title)
	}
}

func TestParseJSON_InvalidKey(t *testing.T) {
	input := `{"badkey": {"URL": []}}`
	_, err := ParseJSON([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestParseJSON_InvalidField(t *testing.T) {
	input := `{"global": {"BadField": ["x"]}}`
	_, err := ParseJSON([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
}

func TestParseJSON_InvalidRegex(t *testing.T) {
	input := `{"global": {"URL": ["[invalid"]}}`
	_, err := ParseJSON([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestParseJSON_EmptyObject(t *testing.T) {
	fm, err := ParseJSON([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fm) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(fm))
	}
}

func TestMatchItem_NoRules(t *testing.T) {
	if !MatchItem("url", "title", "desc", nil) {
		t.Error("expected match with no rules")
	}
}

func TestMatchItem_SimpleMatch(t *testing.T) {
	rule := Compile(Rule{Title: []string{"golang"}})
	if !MatchItem("", "I love golang", "", []CompiledRule{rule}) {
		t.Error("expected match")
	}
	if MatchItem("", "I love python", "", []CompiledRule{rule}) {
		t.Error("expected no match")
	}
}

func TestMatchItem_Negation(t *testing.T) {
	rule := Compile(Rule{URL: []string{"!/shorts/"}})
	if !MatchItem("https://example.com/video", "", "", []CompiledRule{rule}) {
		t.Error("expected match for non-shorts URL")
	}
	if MatchItem("https://example.com/shorts/123", "", "", []CompiledRule{rule}) {
		t.Error("expected no match for shorts URL")
	}
}

func TestMatchItem_MultiplePatterns_AND(t *testing.T) {
	rule := Compile(Rule{Title: []string{"golang", "tutorial"}})
	if !MatchItem("", "golang tutorial basics", "", []CompiledRule{rule}) {
		t.Error("expected match when both patterns present")
	}
	if MatchItem("", "golang basics", "", []CompiledRule{rule}) {
		t.Error("expected no match when only one pattern present")
	}
}

func TestMatchItem_MultipleRules_AND(t *testing.T) {
	r1 := Compile(Rule{Title: []string{"golang"}})
	r2 := Compile(Rule{URL: []string{"!/shorts/"}})
	if !MatchItem("https://example.com/post", "golang news", "", []CompiledRule{r1, r2}) {
		t.Error("expected match")
	}
	if MatchItem("https://example.com/shorts/1", "golang news", "", []CompiledRule{r1, r2}) {
		t.Error("expected no match due to URL negation")
	}
}

func TestMatchItem_MultipleFields(t *testing.T) {
	rule := Compile(Rule{
		URL:   []string{"example\\.com"},
		Title: []string{"news"},
	})
	if !MatchItem("https://example.com", "breaking news", "", []CompiledRule{rule}) {
		t.Error("expected match")
	}
	if MatchItem("https://other.com", "breaking news", "", []CompiledRule{rule}) {
		t.Error("expected no match due to URL mismatch")
	}
}

func TestRulesForFeed(t *testing.T) {
	fm := FilterMap{
		"global":      {URL: []string{"x"}},
		"folder:Tech": {Title: []string{"y"}},
		"feed:42":     {Description: []string{"z"}},
	}
	rules := RulesForFeed(fm, 42, []string{"Tech", "Science"})
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	rules = RulesForFeed(fm, 99, []string{"Other"})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule (global only), got %d", len(rules))
	}
}

func TestRuleToYAML_RoundTrip(t *testing.T) {
	rule := Rule{
		URL:         []string{"!/shorts/"},
		Title:       []string{"golang", "tutorial"},
		Description: []string{},
	}
	yaml := RuleToYAML(rule, "test feed")
	parsed, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("unexpected error parsing YAML: %v", err)
	}
	if len(parsed.URL) != 1 || parsed.URL[0] != "!/shorts/" {
		t.Errorf("URL mismatch: %v", parsed.URL)
	}
	if len(parsed.Title) != 2 {
		t.Errorf("Title mismatch: %v", parsed.Title)
	}
	if len(parsed.Description) != 0 {
		t.Errorf("Description should be empty: %v", parsed.Description)
	}
}

func TestParseYAML_Empty(t *testing.T) {
	rule, err := ParseYAML("# just comments\nURL:\nTitle:\nDescription:\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !IsEmpty(rule) {
		t.Error("expected empty rule")
	}
}

func TestParseYAML_InvalidField(t *testing.T) {
	_, err := ParseYAML("BadField:\n  - \"x\"\n")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestParseYAML_InvalidRegex(t *testing.T) {
	_, err := ParseYAML("URL:\n  - \"[invalid\"\n")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestParseYAML_ListWithoutHeader(t *testing.T) {
	_, err := ParseYAML("  - \"orphan\"\n")
	if err == nil {
		t.Fatal("expected error for list item without header")
	}
}

func TestHasActiveFilter(t *testing.T) {
	fm := FilterMap{"feed:42": {Title: []string{"x"}}}
	if !HasActiveFilter(fm, 42, nil) {
		t.Error("expected active filter for feed 42")
	}
	if HasActiveFilter(fm, 99, nil) {
		t.Error("expected no active filter for feed 99")
	}
}

func TestIsEmpty(t *testing.T) {
	if !IsEmpty(Rule{}) {
		t.Error("expected empty")
	}
	if IsEmpty(Rule{Title: []string{"x"}}) {
		t.Error("expected not empty")
	}
}

func TestFeedKey(t *testing.T) {
	if FeedKey(42) != "feed:42" {
		t.Error("unexpected key")
	}
}

func TestFolderKey(t *testing.T) {
	if FolderKey("Tech") != "folder:Tech" {
		t.Error("unexpected key")
	}
}
