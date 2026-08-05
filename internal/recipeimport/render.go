package recipeimport

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var dashPattern = regexp.MustCompile(`-+`)

func Slugify(title string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			if b.Len() > 0 && !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(dashPattern.ReplaceAllString(b.String(), "-"), "-")
}

func Render(recipe Normalized, source Source) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", oneLine(recipe.Title))
	b.WriteString("**Yield / Target**\n\n")
	for _, item := range recipe.Yield {
		fmt.Fprintf(&b, "- %s\n", oneLine(item))
	}
	b.WriteString("\n## Ingredients\n\n")
	lastGroup := "\x00"
	for _, ingredient := range recipe.Ingredients {
		group := oneLine(ingredient.Group)
		if group != lastGroup {
			if group != "" {
				if lastGroup != "\x00" {
					b.WriteByte('\n')
				}
				fmt.Fprintf(&b, "**%s**\n\n", group)
			}
			lastGroup = group
		}
		fmt.Fprintf(&b, "- %s: %s\n", oneLine(ingredient.Name), oneLine(ingredient.Amount))
	}
	b.WriteString("\n## Process\n\n")
	for i, step := range recipe.Process {
		fmt.Fprintf(&b, "%d. %s\n", i+1, oneLine(step))
	}
	b.WriteString("\n## Notes\n\n- Status: Imported, untested\n")
	for _, note := range recipe.Notes {
		if note = oneLine(note); note != "" {
			fmt.Fprintf(&b, "- %s\n", note)
		}
	}
	for _, warning := range recipe.Warnings {
		if warning = oneLine(warning); warning != "" {
			fmt.Fprintf(&b, "- Import warning: %s\n", warning)
		}
	}
	b.WriteString("\n## Source\n\n")
	if source.URL != "" {
		fmt.Fprintf(&b, "- Original: [%s](%s)\n", escapeLinkLabel(source.Title), source.URL)
	} else {
		b.WriteString("- Original: Shared text (no URL)\n")
	}
	if source.Author != "" {
		fmt.Fprintf(&b, "- Author: %s\n", oneLine(source.Author))
	}
	if source.Published != "" {
		fmt.Fprintf(&b, "- Published: %s\n", oneLine(source.Published))
	}
	fmt.Fprintf(&b, "- Accessed: %s\n", source.Accessed)
	return b.String()
}

func oneLine(value string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(value), " "))
}

func escapeLinkLabel(value string) string {
	value = oneLine(value)
	if value == "" {
		value = "Original recipe"
	}
	return strings.NewReplacer("[", "\\[", "]", "\\]").Replace(value)
}

func normalizeForComparison(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" {
		return strings.TrimSpace(raw)
	}
	u.Fragment = ""
	query := u.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" {
			query.Del(key)
		}
	}
	u.RawQuery = query.Encode()
	u.Host = strings.ToLower(u.Host)
	return u.String()
}
