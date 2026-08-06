package recipeimport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

const recipeHTML = `<!doctype html><html><head>
<script data-extra="yes" type="application/ld+json">
{"@context":"https://schema.org","@graph":[
  {"@type":"BreadcrumbList","name":"Not a recipe"},
  {"@type":["Thing","Recipe"],"name":"Test Soup","author":{"name":"Ada Cook"},
   "datePublished":"2026-07-01","recipeYield":["4 servings"],
   "recipeIngredient":["2 cups stock","400 g tomatoes","1 tsp salt"],
   "recipeInstructions":[{"@type":"HowToSection","name":"Soup","itemListElement":[
     {"@type":"HowToStep","text":"Combine everything."},
     {"@type":"HowToStep","text":"Simmer for 20 minutes."}
   ]}]}
]}
</script></head><body><p>Visible fallback</p></body></html>`

func TestExtractHTMLRecipeGraph(t *testing.T) {
	got, err := ExtractHTML(recipeHTML, "https://example.com/soup")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Test Soup" || got.Author != "Ada Cook" {
		t.Fatalf("unexpected metadata: %#v", got)
	}
	if len(got.Ingredients) != 3 || len(got.Instructions) != 2 {
		t.Fatalf("unexpected extraction: %#v", got)
	}
}

func TestExtractHTMLReturnsVisibleFallback(t *testing.T) {
	page := `<html><head><script type="application/ld+json">{"@type":"Article","headline":"Shared Cake","datePublished":"2023-06-23T13:00:00-0400","publisher":{"@type":"Organization","name":"Ada Cook"}}</script></head><body><h1>Shared Cake</h1><p>Mix it.</p></body></html>`
	got, err := ExtractHTML(page, "https://example.com/cake")
	if err == nil {
		t.Fatal("expected missing JSON-LD error")
	}
	if !strings.Contains(got.SourceText, "Shared Cake") || !strings.Contains(got.SourceText, "Mix it.") {
		t.Fatalf("missing visible fallback: %q", got.SourceText)
	}
	if got.Title != "Shared Cake" || got.Author != "Ada Cook" || got.Published != "2023-06-23T13:00:00-0400" {
		t.Fatalf("missing Article fallback metadata: %#v", got)
	}
}

func TestNormalizePublished(t *testing.T) {
	if got := normalizePublished("2023-06-23T13:00:00-0400"); got != "2023-06-23" {
		t.Fatalf("got %q", got)
	}
	if got := normalizePublished("Jun 23"); got != "Jun 23" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeDeterministicRejectsMissingAmount(t *testing.T) {
	_, err := normalizeDeterministic(Extracted{
		Title:        "Soup",
		Yield:        []string{"4 servings"},
		Ingredients:  []string{"Salt to taste"},
		Instructions: []string{"Season the soup."},
	})
	if err == nil || !strings.Contains(err.Error(), "no ingredients") {
		t.Fatalf("expected conservative failure, got %v", err)
	}
}

func TestNormalizeDeterministicParsesMixedFraction(t *testing.T) {
	got, err := normalizeDeterministic(Extracted{
		Title:        "Bread",
		Yield:        []string{"1 loaf"},
		Ingredients:  []string{"1½ cups flour"},
		Instructions: []string{"Mix."},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ingredients[0].Name != "flour" || got.Ingredients[0].Amount != "1 1/2 cups" {
		t.Fatalf("unexpected ingredient: %#v", got.Ingredients[0])
	}
}

func TestNormalizeUsesStructuredResponsesOutput(t *testing.T) {
	client := doerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.openai.com/v1/responses" {
			t.Fatalf("unexpected URL: %s", req.URL)
		}
		if req.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing authorization")
		}
		requestBody, _ := io.ReadAll(req.Body)
		var payload map[string]any
		if err := json.Unmarshal(requestBody, &payload); err != nil {
			t.Fatal(err)
		}
		text, _ := payload["text"].(map[string]any)
		format, _ := text["format"].(map[string]any)
		if format["type"] != "json_schema" || format["strict"] != true {
			t.Fatalf("structured output missing: %#v", payload)
		}
		instructions, _ := payload["instructions"].(string)
		if !strings.Contains(instructions, "in clear English") || !strings.Contains(instructions, "Keep source_title in its original language") {
			t.Fatalf("English-output instructions missing: %q", instructions)
		}
		output := `{"title":"Soup","yield":["4 servings"],"ingredients":[{"group":"","name":"Tomatoes","amount":"400 g"}],"process":["Simmer."],"notes":[],"warnings":[],"source_title":"Soup","source_author":"Ada","published":""}`
		response, _ := json.Marshal(map[string]any{"output": []any{map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": output}}}}})
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(response))), Header: make(http.Header)}, nil
	})

	got, err := Normalize(context.Background(), Extracted{Title: "Soup"}, "secret", "test-model", client)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Soup" || got.Ingredients[0].Amount != "400 g" {
		t.Fatalf("unexpected normalized recipe: %#v", got)
	}
}

func TestFetchSourceSupportsHTMLPDFAndImages(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        []byte
		wantType    string
	}{
		{name: "html", contentType: "text/html; charset=utf-8", body: []byte(recipeHTML), wantType: "text/html"},
		{name: "pdf", contentType: "application/octet-stream", body: []byte("%PDF-1.7\nrecipe"), wantType: "application/pdf"},
		{name: "jpeg", contentType: "image/jpg", body: []byte{0xff, 0xd8, 0xff, 0xdb}, wantType: "image/jpeg"},
		{name: "png", contentType: "", body: []byte("\x89PNG\r\n\x1a\nrest"), wantType: "image/png"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := doerFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(string(test.body))),
					Header:     http.Header{"Content-Type": []string{test.contentType}},
					Request:    req,
				}, nil
			})
			got, err := FetchSource(client, "https://EXAMPLE.com/recipe?utm_source=share")
			if err != nil {
				t.Fatal(err)
			}
			if got.MediaType != test.wantType || got.URL != "https://example.com/recipe" {
				t.Fatalf("unexpected source: %#v", got)
			}
		})
	}
}

func TestFetchSourceRejectsUnsupportedContent(t *testing.T) {
	client := doerFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("plain text")),
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Request:    req,
		}, nil
	})
	_, err := FetchSource(client, "https://example.com/recipe.txt")
	if err == nil || !strings.Contains(err.Error(), "unsupported source content type") {
		t.Fatalf("expected unsupported-content failure, got %v", err)
	}
}

func TestNormalizeMediaBuildsMultimodalResponsesInput(t *testing.T) {
	for _, test := range []struct {
		name       string
		mediaType  string
		wantType   string
		wantField  string
		wantPrefix string
	}{
		{name: "pdf", mediaType: "application/pdf", wantType: "input_file", wantField: "file_data", wantPrefix: "data:application/pdf;base64,"},
		{name: "image", mediaType: "image/jpeg", wantType: "input_image", wantField: "image_url", wantPrefix: "data:image/jpeg;base64,"},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := doerFunc(func(req *http.Request) (*http.Response, error) {
				requestBody, _ := io.ReadAll(req.Body)
				var payload map[string]any
				if err := json.Unmarshal(requestBody, &payload); err != nil {
					t.Fatal(err)
				}
				messages, ok := payload["input"].([]any)
				if !ok || len(messages) != 1 {
					t.Fatalf("missing message input: %#v", payload["input"])
				}
				message := messages[0].(map[string]any)
				content := message["content"].([]any)
				media := content[1].(map[string]any)
				if media["type"] != test.wantType || media["detail"] != "high" {
					t.Fatalf("unexpected media input: %#v", media)
				}
				if test.mediaType == "application/pdf" && media["filename"] != "recipe.pdf" {
					t.Fatalf("unexpected PDF filename: %#v", media["filename"])
				}
				value, _ := media[test.wantField].(string)
				if !strings.HasPrefix(value, test.wantPrefix) {
					t.Fatalf("unexpected encoded input: %q", value)
				}
				output := `{"title":"Photo Soup","yield":["4 servings"],"ingredients":[{"group":"","name":"Tomatoes","amount":"400 g"}],"process":["Simmer."],"notes":[],"warnings":[],"source_title":"Photo Soup","source_author":"","published":""}`
				response, _ := json.Marshal(map[string]any{"output": []any{map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": output}}}}})
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(response))), Header: make(http.Header)}, nil
			})

			got, err := NormalizeMedia(context.Background(), FetchedSource{
				Body: []byte("recipe bytes"), URL: "https://example.com/recipe.pdf", MediaType: test.mediaType,
			}, "", "secret", "test-model", client)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "verify OCR") {
				t.Fatalf("missing media warning: %#v", got.Warnings)
			}
		})
	}
}

func TestNormalizeMediaRequiresAI(t *testing.T) {
	_, err := NormalizeMedia(context.Background(), FetchedSource{MediaType: "application/pdf"}, "", "", "", nil)
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("expected API key failure, got %v", err)
	}
}

func TestImportRoutesPDFToMediaNormalization(t *testing.T) {
	client := doerFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("%PDF-1.7\nrecipe")),
			Header:     http.Header{"Content-Type": []string{"application/pdf"}},
			Request:    req,
		}, nil
	})
	_, err := Import(context.Background(), Request{URL: "https://example.com/recipe.pdf"}, Options{
		DryRun: true, NoAI: true, HTTPClient: client,
	})
	if err == nil || !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("expected media normalization failure, got %v", err)
	}
}

func TestNormalizeOmitsProcessOnlyIngredientWithoutAmount(t *testing.T) {
	client := doerFunc(func(req *http.Request) (*http.Response, error) {
		output := `{"title":"Rolls","yield":["2 rolls"],"ingredients":[{"group":"","name":"Bread flour","amount":"159 g"},{"group":"","name":"Cornmeal","amount":"as needed"}],"process":["Sprinkle parchment with cornmeal."],"notes":[],"warnings":[],"source_title":"Rolls","source_author":"","published":""}`
		response, _ := json.Marshal(map[string]any{"output": []any{map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": output}}}}})
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(response))), Header: make(http.Header)}, nil
	})

	got, err := Normalize(context.Background(), Extracted{Title: "Rolls"}, "secret", "test-model", client)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Ingredients) != 1 || got.Ingredients[0].Name != "Bread flour" {
		t.Fatalf("unexpected ingredients: %#v", got.Ingredients)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "Cornmeal") {
		t.Fatalf("expected an explicit cornmeal warning, got %#v", got.Warnings)
	}
}

func TestImportWritesCanonicalRecipeAndDetectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	client := doerFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(recipeHTML)),
			Header:     http.Header{"Content-Type": []string{"text/html"}},
			Request:    req,
		}, nil
	})
	opts := Options{OutDir: dir, Accessed: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC), NoAI: true, HTTPClient: client}
	result, err := Import(context.Background(), Request{URL: "https://EXAMPLE.com/soup?utm_source=test", Note: "Family find"}, opts)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, result.Slug+".md"))
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(body)
	for _, want := range []string{"# Test Soup", "**Yield / Target**", "## Ingredients", "- stock: 2 cups", "## Process", "- Status: Imported, untested", "https://example.com/soup", "- Accessed: 2026-08-04"} {
		if !strings.Contains(markdown, want) {
			t.Errorf("Markdown missing %q:\n%s", want, markdown)
		}
	}
	_, err = Import(context.Background(), Request{URL: "https://example.com/soup"}, opts)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate failure, got %v", err)
	}
}

func TestNormalizeSourceURL(t *testing.T) {
	got, err := NormalizeSourceURL("https://EXAMPLE.com/path?utm_source=x&keep=y#part")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/path?keep=y" {
		t.Fatalf("got %q", got)
	}
	for _, raw := range []string{"http://example.com", "https://localhost/test", "not a URL"} {
		if _, err := NormalizeSourceURL(raw); err == nil {
			t.Errorf("expected %q to fail", raw)
		}
	}
}
