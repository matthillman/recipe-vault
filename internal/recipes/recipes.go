package recipes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Recipe struct {
	Slug  string
	Title string
	Path  string

	RawLines []string

	// Raw sections keyed by the exact "## Heading" text without the leading "## ".
	Sections map[string][]string
	// Section keys in the order encountered (first occurrence wins).
	SectionOrder []string
}

func LoadAll(dir string) (map[string]Recipe, error) {
	out := map[string]Recipe{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var mdFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".md") {
			mdFiles = append(mdFiles, filepath.Join(dir, name))
		}
	}
	sort.Strings(mdFiles)

	for _, p := range mdFiles {
		r, err := LoadFile(p)
		if err != nil {
			return nil, err
		}
		if _, exists := out[r.Slug]; exists {
			return nil, fmt.Errorf("duplicate recipe slug: %s", r.Slug)
		}
		out[r.Slug] = r
	}
	return out, nil
}

func LoadFile(path string) (Recipe, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, err
	}
	txt := string(b)
	lines := splitLines(txt)

	r := Recipe{
		Slug:     strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		Path:     filepath.ToSlash(path),
		RawLines: lines,
		Sections: map[string][]string{},
	}

	var current string
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(line, "# ") && r.Title == "" {
			r.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			current = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "## "), "### "))
			if _, exists := r.Sections[current]; !exists {
				r.SectionOrder = append(r.SectionOrder, current)
			}
			r.Sections[current] = nil
			continue
		}
		if current != "" {
			r.Sections[current] = append(r.Sections[current], line)
		}
	}

	if strings.TrimSpace(r.Title) == "" {
		// Fall back to slug.
		r.Title = r.Slug
	}
	return r, nil
}

func splitLines(s string) []string {
	// Preserve empty lines for block parsing.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

func (r Recipe) Section(name string) []string {
	if r.Sections == nil {
		return nil
	}
	return r.Sections[name]
}

func (r Recipe) HasSection(name string) bool {
	_, ok := r.Sections[name]
	return ok
}

func (r Recipe) HasSectionPrefix(prefix string) bool {
	_, ok := r.findSectionKeyByPrefix(prefix)
	return ok
}

func (r Recipe) findSectionKeyByPrefix(prefix string) (string, bool) {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		return "", false
	}
	for _, k := range r.SectionOrder {
		low := strings.ToLower(k)
		if low == prefix || strings.HasPrefix(low, prefix+" ") || strings.HasPrefix(low, prefix+"(") || strings.HasPrefix(low, prefix+":") {
			return k, true
		}
	}
	for k := range r.Sections {
		low := strings.ToLower(k)
		if low == prefix || strings.HasPrefix(low, prefix+" ") || strings.HasPrefix(low, prefix+"(") || strings.HasPrefix(low, prefix+":") {
			return k, true
		}
	}
	return "", false
}

// FindYieldLines looks for a bold Yield header and captures the next 1-3 bullet lines.
func (r Recipe) FindYieldLines() []string {
	markers := []string{"Yield / Target", "Yield / Pan Target"}

	lines := r.RawLines

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		for _, m := range markers {
			if line == "**"+m+"**" {
				var out []string
				for j := i + 1; j < len(lines) && len(out) < 3; j++ {
					t := strings.TrimSpace(lines[j])
					if t == "" {
						if len(out) > 0 {
							return out
						}
						continue
					}
					if isBulletLine(t) {
						out = append(out, strings.TrimSpace(t[1:]))
						continue
					}
					// Stop if we hit non-bullet content after getting something.
					if len(out) > 0 {
						return out
					}
				}
				return out
			}
		}
	}
	return nil
}

func isBulletLine(s string) bool {
	return strings.HasPrefix(s, "- ") || strings.HasPrefix(s, "* ")
}

// ExtractListItems returns cleaned bullet/numbered list items from a section.
func (r Recipe) ExtractListItems(section string) []string {
	key, ok := r.findSectionKeyByPrefix(section)
	if !ok {
		return nil
	}
	raw := r.Section(key)
	if raw == nil {
		return nil
	}
	var out []string
	for _, line := range raw {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") {
			out = append(out, strings.TrimSpace(t[2:]))
			continue
		}
		// Numbered lists: "1. ..."
		if n, ok := cutNumberedPrefix(t); ok {
			out = append(out, n)
			continue
		}
	}
	return out
}

func cutNumberedPrefix(s string) (string, bool) {
	// Very small parser: one or more digits, then ". ".
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 {
		return "", false
	}
	if i+2 > len(s) || s[i] != '.' || s[i+1] != ' ' {
		return "", false
	}
	return strings.TrimSpace(s), true
}

// ExtractFormulaLines tries common headings, then falls back to scanning for percent signs.
func (r Recipe) ExtractFormulaLines() []string {
	// Common section names we use.
	for _, name := range []string{
		"Formula",
		"Formula (Baker's %)",
		"Formula (Baker’s %)",
	} {
		key, ok := r.findSectionKeyByPrefix(name)
		if ok {
			lines := r.ExtractListItems(name)
			if len(lines) > 0 {
				return lines
			}
			// If it wasn't a list, keep non-empty lines.
			var out []string
			for _, l := range r.Section(key) {
				t := strings.TrimSpace(l)
				if t != "" {
					out = append(out, t)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}

	// Fallback: scan ingredients for "(xx%)" style lines.
	ings := r.ExtractListItems("Ingredients")
	var out []string
	for _, l := range ings {
		if strings.Contains(l, "%") {
			out = append(out, l)
		}
	}
	return out
}

// CardOverride returns the raw lines from a "## Card" section (if present).
func (r Recipe) CardOverride() []string {
	key, ok := r.findSectionKeyByPrefix("Card")
	if !ok {
		return nil
	}
	var out []string
	for _, l := range r.Section(key) {
		t := strings.TrimRight(l, " \t")
		if strings.TrimSpace(t) == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

// Ensure we don't accidentally import fs unused in some environments.
var _ fs.DirEntry
