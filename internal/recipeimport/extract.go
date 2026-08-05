package recipeimport

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	scriptPattern      = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script\s*>`)
	typePattern        = regexp.MustCompile(`(?i)\btype\s*=\s*["']?application/ld\+json(?:\s*;[^"'\s>]*)?["']?`)
	tagPattern         = regexp.MustCompile(`(?s)<[^>]+>`)
	spacePattern       = regexp.MustCompile(`[ \t\r\f\v]+`)
	scriptBlockPattern = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script\s*>`)
	styleBlockPattern  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style\s*>`)
)

func ExtractHTML(body, sourceURL string) (Extracted, error) {
	var candidates []any
	var articles []any
	for _, match := range scriptPattern.FindAllStringSubmatch(body, -1) {
		if len(match) != 3 || !typePattern.MatchString(match[1]) {
			continue
		}
		var value any
		if err := json.Unmarshal([]byte(strings.TrimSpace(match[2])), &value); err != nil {
			continue
		}
		collectRecipes(value, &candidates)
		collectTyped(value, "Article", &articles)
	}
	if len(candidates) == 0 {
		fallback := Extracted{SourceURL: sourceURL, SourceText: visibleText(body)}
		if len(articles) > 0 {
			article, _ := articles[0].(map[string]any)
			fallback.Title = stringValue(firstNonNil(article["headline"], article["name"]))
			fallback.Author = firstString(authorValue(article["author"]), authorValue(article["publisher"]))
			fallback.Published = stringValue(firstNonNil(article["datePublished"], article["dateCreated"]))
		}
		return fallback, fmt.Errorf("page has no usable schema.org Recipe JSON-LD")
	}

	best := candidates[0]
	bestScore := scoreRecipe(best)
	for _, candidate := range candidates[1:] {
		if score := scoreRecipe(candidate); score > bestScore {
			best, bestScore = candidate, score
		}
	}
	obj, _ := best.(map[string]any)
	out := Extracted{
		Title:        stringValue(obj["name"]),
		Author:       authorValue(obj["author"]),
		Published:    stringValue(firstNonNil(obj["datePublished"], obj["dateCreated"])),
		Yield:        stringSlice(obj["recipeYield"]),
		Ingredients:  stringSlice(obj["recipeIngredient"]),
		Instructions: instructionSlice(obj["recipeInstructions"]),
		SourceURL:    sourceURL,
	}
	if out.Title == "" {
		return out, fmt.Errorf("Recipe JSON-LD is missing a title")
	}
	return out, nil
}

func collectTyped(value any, typeName string, out *[]any) {
	switch v := value.(type) {
	case []any:
		for _, child := range v {
			collectTyped(child, typeName, out)
		}
	case map[string]any:
		if hasType(v["@type"], typeName) {
			*out = append(*out, v)
		}
		for _, child := range v {
			collectTyped(child, typeName, out)
		}
	}
}

func collectRecipes(value any, out *[]any) {
	switch v := value.(type) {
	case []any:
		for _, child := range v {
			collectRecipes(child, out)
		}
	case map[string]any:
		if hasType(v["@type"], "Recipe") {
			*out = append(*out, v)
		}
		for _, child := range v {
			collectRecipes(child, out)
		}
	}
}

func hasType(value any, want string) bool {
	switch v := value.(type) {
	case string:
		return strings.EqualFold(v, want) || strings.HasSuffix(strings.ToLower(v), "/"+strings.ToLower(want))
	case []any:
		for _, item := range v {
			if hasType(item, want) {
				return true
			}
		}
	}
	return false
}

func scoreRecipe(value any) int {
	obj, _ := value.(map[string]any)
	return len(stringSlice(obj["recipeIngredient"]))*2 + len(instructionSlice(obj["recipeInstructions"]))*2 + len(stringSlice(obj["recipeYield"]))
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return cleanText(v)
	case json.Number:
		return v.String()
	case float64:
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func stringSlice(value any) []string {
	switch v := value.(type) {
	case string:
		if cleaned := cleanText(v); cleaned != "" {
			return []string{cleaned}
		}
	case []any:
		var out []string
		for _, item := range v {
			if cleaned := stringValue(item); cleaned != "" {
				out = append(out, cleaned)
			}
		}
		return out
	}
	return nil
}

func authorValue(value any) string {
	switch v := value.(type) {
	case string:
		return cleanText(v)
	case map[string]any:
		return stringValue(v["name"])
	case []any:
		var names []string
		for _, item := range v {
			if name := authorValue(item); name != "" {
				names = append(names, name)
			}
		}
		return strings.Join(names, ", ")
	}
	return ""
}

func instructionSlice(value any) []string {
	var out []string
	var visit func(any)
	visit = func(current any) {
		switch v := current.(type) {
		case string:
			if cleaned := cleanText(v); cleaned != "" {
				out = append(out, cleaned)
			}
		case []any:
			for _, child := range v {
				visit(child)
			}
		case map[string]any:
			if hasType(v["@type"], "HowToSection") {
				visit(v["itemListElement"])
				return
			}
			if text := stringValue(firstNonNil(v["text"], v["name"])); text != "" {
				out = append(out, text)
			}
		}
	}
	visit(value)
	return out
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func cleanText(value string) string {
	value = html.UnescapeString(tagPattern.ReplaceAllString(value, " "))
	return strings.TrimSpace(spacePattern.ReplaceAllString(value, " "))
}

func visibleText(body string) string {
	body = scriptBlockPattern.ReplaceAllString(body, " ")
	body = styleBlockPattern.ReplaceAllString(body, " ")
	body = regexp.MustCompile(`(?i)<br\s*/?>|</(?:p|li|h[1-6]|div|section)>`).ReplaceAllString(body, "\n")
	body = html.UnescapeString(tagPattern.ReplaceAllString(body, " "))
	lines := strings.Split(body, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(spacePattern.ReplaceAllString(line, " "))
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return truncateUTF8(strings.Join(cleaned, "\n"), 150_000)
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for len(value) > 0 && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return strings.TrimSpace(value) + "\n[Source text truncated]"
}
