package recipeimport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sourceURLPattern = regexp.MustCompile(`https://[^\s)>]+`)

func Import(ctx context.Context, request Request, opts Options) (Result, error) {
	request.URL = strings.TrimSpace(request.URL)
	request.SourceText = strings.TrimSpace(request.SourceText)
	if request.URL == "" && request.SourceText == "" {
		return Result{}, fmt.Errorf("provide a URL and/or source text")
	}
	if len(request.SourceText) > 32<<10 {
		return Result{}, fmt.Errorf("source text exceeds 32 KiB")
	}
	if len(request.Note) > 1<<10 {
		return Result{}, fmt.Errorf("note exceeds 1 KiB")
	}

	var extracted Extracted
	if request.URL != "" {
		body, finalURL, err := FetchPage(opts.HTTPClient, request.URL)
		if err != nil {
			if request.SourceText == "" {
				return Result{}, err
			}
			normalizedURL, normalizeErr := NormalizeSourceURL(request.URL)
			if normalizeErr != nil {
				return Result{}, normalizeErr
			}
			extracted.SourceURL = normalizedURL
		} else {
			extracted, err = ExtractHTML(body, finalURL)
			if err != nil && strings.TrimSpace(extracted.SourceText) == "" && request.SourceText == "" {
				return Result{}, err
			}
			extracted.SourceURL = finalURL
		}
	}
	if request.SourceText != "" {
		if extracted.SourceText != "" {
			extracted.SourceText += "\n\nShared text:\n" + request.SourceText
		} else {
			extracted.SourceText = request.SourceText
		}
	}
	if opts.NoAI {
		opts.APIKey = ""
	}
	normalized, err := Normalize(ctx, extracted, opts.APIKey, opts.Model, opts.OpenAIClient)
	if err != nil {
		return Result{}, err
	}
	if note := oneLine(request.Note); note != "" {
		normalized.Notes = append(normalized.Notes, "Capture note: "+note)
	}

	accessed := opts.Accessed
	if accessed.IsZero() {
		accessed = time.Now()
	}
	source := Source{
		URL:       extracted.SourceURL,
		Title:     firstString(normalized.SourceTitle, extracted.Title, normalized.Title),
		Author:    firstString(normalized.SourceAuthor, extracted.Author),
		Published: normalizePublished(firstString(extracted.Published, normalized.Published)),
		Accessed:  accessed.Format("2006-01-02"),
	}
	slug := Slugify(normalized.Title)
	if slug == "" {
		return Result{}, fmt.Errorf("title does not produce a valid ASCII slug")
	}
	markdown := Render(normalized, source)
	result := Result{Slug: slug, Title: normalized.Title, Warnings: normalized.Warnings}
	if opts.DryRun {
		result.Markdown = markdown
		return result, nil
	}
	if opts.OutDir == "" {
		opts.OutDir = "recipes"
	}
	if err := checkDuplicate(opts.OutDir, slug, source.URL); err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return Result{}, err
	}
	path := filepath.Join(opts.OutDir, slug+".md")
	if err := os.WriteFile(path, []byte(markdown), 0o644); err != nil {
		return Result{}, err
	}
	result.Path = filepath.ToSlash(path)
	return result, nil
}

func normalizePublished(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 10 && value[4] == '-' && value[7] == '-' {
		if _, err := time.Parse("2006-01-02", value[:10]); err == nil {
			return value[:10]
		}
	}
	return value
}

func checkDuplicate(dir, slug, sourceURL string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	wantedURL := normalizeForComparison(sourceURL)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if strings.TrimSuffix(entry.Name(), ".md") == slug {
			return fmt.Errorf("recipe slug already exists: %s", slug)
		}
		if wantedURL == "" {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			return readErr
		}
		for _, found := range sourceURLPattern.FindAllString(string(body), -1) {
			if normalizeForComparison(strings.TrimRight(found, ".,")) == wantedURL {
				return fmt.Errorf("source URL already exists in %s", filepath.Join(dir, entry.Name()))
			}
		}
	}
	return nil
}

func firstString(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func DecodeRequest(path string) (Request, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return Request{}, err
	}
	var request Request
	if err := json.Unmarshal(body, &request); err != nil {
		return Request{}, fmt.Errorf("decode request: %w", err)
	}
	return request, nil
}
