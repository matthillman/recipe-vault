package recipeimport

import (
	"net/http"
	"time"
)

const DefaultModel = "gpt-5.6-terra"

type Request struct {
	URL            string `json:"url,omitempty"`
	SourceText     string `json:"sourceText,omitempty"`
	Note           string `json:"note,omitempty"`
	Client         string `json:"client,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

type Source struct {
	URL       string `json:"url"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Published string `json:"published"`
	Accessed  string `json:"accessed"`
}

type Extracted struct {
	Title        string   `json:"title"`
	Author       string   `json:"author"`
	Published    string   `json:"published"`
	Yield        []string `json:"yield"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
	SourceURL    string   `json:"source_url"`
	SourceText   string   `json:"source_text"`
}

type Ingredient struct {
	Group  string `json:"group"`
	Name   string `json:"name"`
	Amount string `json:"amount"`
}

type Normalized struct {
	Title        string       `json:"title"`
	Yield        []string     `json:"yield"`
	Ingredients  []Ingredient `json:"ingredients"`
	Process      []string     `json:"process"`
	Notes        []string     `json:"notes"`
	Warnings     []string     `json:"warnings"`
	SourceTitle  string       `json:"source_title"`
	SourceAuthor string       `json:"source_author"`
	Published    string       `json:"published"`
}

type Options struct {
	OutDir       string
	Accessed     time.Time
	APIKey       string
	Model        string
	NoAI         bool
	DryRun       bool
	HTTPClient   HTTPDoer
	OpenAIClient HTTPDoer
}

type Result struct {
	Slug     string   `json:"slug"`
	Path     string   `json:"path,omitempty"`
	Title    string   `json:"title"`
	Warnings []string `json:"warnings,omitempty"`
	Markdown string   `json:"markdown,omitempty"`
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
