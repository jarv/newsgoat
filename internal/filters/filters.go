package filters

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jarv/newsgoat/internal/database"
)

const settingsKey = "filters"

var validFields = map[string]bool{"URL": true, "Title": true, "Description": true}

type Rule struct {
	URL         []string `json:"URL,omitempty"`
	Title       []string `json:"Title,omitempty"`
	Description []string `json:"Description,omitempty"`
}

type FilterMap map[string]Rule

type compiledPattern struct {
	re     *regexp.Regexp
	negate bool
}

type CompiledRule struct {
	URL         []compiledPattern
	Title       []compiledPattern
	Description []compiledPattern
}

func Load(queries *database.Queries) (FilterMap, error) {
	row, err := queries.GetSetting(context.Background(), settingsKey)
	if err != nil {
		return make(FilterMap), nil
	}
	fm, err := ParseJSON([]byte(row.Value))
	if err != nil {
		return make(FilterMap), err
	}
	return fm, nil
}

func Save(queries *database.Queries, fm FilterMap) error {
	for k, r := range fm {
		if len(r.URL) == 0 && len(r.Title) == 0 && len(r.Description) == 0 {
			delete(fm, k)
		}
	}
	if len(fm) == 0 {
		return queries.DeleteSetting(context.Background(), settingsKey)
	}
	data, err := json.Marshal(fm)
	if err != nil {
		return err
	}
	return queries.SetSetting(context.Background(), database.SetSettingParams{
		Key:   settingsKey,
		Value: string(data),
	})
}

func ParseJSON(data []byte) (FilterMap, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid filter JSON: %w", err)
	}
	fm := make(FilterMap, len(raw))
	for key, val := range raw {
		if !isValidKey(key) {
			return nil, fmt.Errorf("invalid filter key %q: must be \"global\", \"folder:<name>\", or \"feed:<id>\"", key)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(val, &fields); err != nil {
			return nil, fmt.Errorf("invalid filter rule for %q: %w", key, err)
		}
		for field := range fields {
			if !validFields[field] {
				return nil, fmt.Errorf("invalid field %q in filter %q: allowed fields are URL, Title, Description", field, key)
			}
		}
		var rule Rule
		if err := json.Unmarshal(val, &rule); err != nil {
			return nil, fmt.Errorf("invalid filter rule for %q: %w", key, err)
		}
		for _, patterns := range [][]string{rule.URL, rule.Title, rule.Description} {
			for _, p := range patterns {
				raw := strings.TrimPrefix(p, "!")
				if _, err := regexp.Compile(raw); err != nil {
					return nil, fmt.Errorf("invalid regex %q: %w", p, err)
				}
			}
		}
		fm[key] = rule
	}
	return fm, nil
}

func isValidKey(key string) bool {
	if key == "global" {
		return true
	}
	if strings.HasPrefix(key, "folder:") && len(key) > 7 {
		return true
	}
	if strings.HasPrefix(key, "feed:") && len(key) > 5 {
		return true
	}
	return false
}

func Compile(rule Rule) CompiledRule {
	return CompiledRule{
		URL:         compilePatterns(rule.URL),
		Title:       compilePatterns(rule.Title),
		Description: compilePatterns(rule.Description),
	}
}

func compilePatterns(patterns []string) []compiledPattern {
	out := make([]compiledPattern, 0, len(patterns))
	for _, p := range patterns {
		negate := false
		raw := p
		if strings.HasPrefix(raw, "!") {
			negate = true
			raw = raw[1:]
		}
		re, err := regexp.Compile(raw)
		if err != nil {
			continue
		}
		out = append(out, compiledPattern{re: re, negate: negate})
	}
	return out
}

func MatchItem(url, title, description string, rules []CompiledRule) bool {
	for _, r := range rules {
		if !matchField(url, r.URL) {
			return false
		}
		if !matchField(title, r.Title) {
			return false
		}
		if !matchField(description, r.Description) {
			return false
		}
	}
	return true
}

func matchField(value string, patterns []compiledPattern) bool {
	for _, p := range patterns {
		matched := p.re.MatchString(value)
		if p.negate && matched {
			return false
		}
		if !p.negate && !matched {
			return false
		}
	}
	return true
}

func RulesForFeed(fm FilterMap, feedID int64, folderNames []string) []Rule {
	var rules []Rule
	if r, ok := fm["global"]; ok {
		rules = append(rules, r)
	}
	for _, folder := range folderNames {
		if r, ok := fm["folder:"+folder]; ok {
			rules = append(rules, r)
		}
	}
	key := fmt.Sprintf("feed:%d", feedID)
	if r, ok := fm[key]; ok {
		rules = append(rules, r)
	}
	return rules
}

func CompileRules(rules []Rule) []CompiledRule {
	out := make([]CompiledRule, len(rules))
	for i, r := range rules {
		out[i] = Compile(r)
	}
	return out
}

func HasActiveFilter(fm FilterMap, feedID int64, folderNames []string) bool {
	return len(RulesForFeed(fm, feedID, folderNames)) > 0
}

func FeedKey(feedID int64) string {
	return fmt.Sprintf("feed:%d", feedID)
}

func FolderKey(folderName string) string {
	return "folder:" + folderName
}

func RuleToYAML(rule Rule, scopeLabel string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Filter for %s\n", scopeLabel)
	b.WriteString("# Regexes are ANDed together. Prefix with ! to negate.\n")
	b.WriteString("# Examples:\n")
	b.WriteString("#   - \"golang\"        # items must match \"golang\"\n")
	b.WriteString("#   - \"!/shorts/\"     # items must NOT match \"/shorts/\"\n")
	b.WriteString("#\n")
	writeYAMLField(&b, "URL", rule.URL)
	writeYAMLField(&b, "Title", rule.Title)
	writeYAMLField(&b, "Description", rule.Description)
	return b.String()
}

func writeYAMLField(b *strings.Builder, name string, values []string) {
	b.WriteString(name + ":\n")
	for _, v := range values {
		fmt.Fprintf(b, "  - %q\n", v)
	}
}

func ParseYAML(content string) (Rule, error) {
	var rule Rule
	lines := strings.Split(content, "\n")
	var currentField *[]string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasSuffix(trimmed, ":") {
			field := strings.TrimSuffix(trimmed, ":")
			switch field {
			case "URL":
				currentField = &rule.URL
			case "Title":
				currentField = &rule.Title
			case "Description":
				currentField = &rule.Description
			default:
				return Rule{}, fmt.Errorf("unknown field %q: allowed fields are URL, Title, Description", field)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			if currentField == nil {
				return Rule{}, fmt.Errorf("list item without a field header: %q", trimmed)
			}
			val := strings.TrimPrefix(trimmed, "- ")
			val = strings.TrimSpace(val)
			if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
				(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
				val = val[1 : len(val)-1]
			}
			raw := strings.TrimPrefix(val, "!")
			if _, err := regexp.Compile(raw); err != nil {
				return Rule{}, fmt.Errorf("invalid regex %q: %w", val, err)
			}
			*currentField = append(*currentField, val)
			continue
		}
		return Rule{}, fmt.Errorf("unexpected line: %q", trimmed)
	}
	return rule, nil
}

func IsEmpty(rule Rule) bool {
	return len(rule.URL) == 0 && len(rule.Title) == 0 && len(rule.Description) == 0
}
