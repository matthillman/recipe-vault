package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"recipelab/internal/recipeimport"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "import" {
		fmt.Fprintln(os.Stderr, "usage: recipeimport import [--url URL] [--source-text-file PATH] [--request PATH] [--out recipes]")
		os.Exit(2)
	}
	flags := flag.NewFlagSet("import", flag.ExitOnError)
	urlValue := flags.String("url", "", "recipe page URL")
	textPath := flags.String("source-text-file", "", "path to pasted/shared recipe text")
	requestPath := flags.String("request", "", "path to a JSON capture request")
	outDir := flags.String("out", "recipes", "recipe output directory")
	accessedValue := flags.String("accessed", "", "source access date (YYYY-MM-DD)")
	model := flags.String("model", envOr("RECIPE_IMPORT_MODEL", recipeimport.DefaultModel), "OpenAI model")
	noAI := flags.Bool("no-ai", false, "disable OpenAI normalization")
	dryRun := flags.Bool("dry-run", false, "print result without writing a recipe")
	_ = flags.Parse(os.Args[2:])

	request := recipeimport.Request{URL: *urlValue, Client: "cli"}
	var err error
	if *requestPath != "" {
		request, err = recipeimport.DecodeRequest(*requestPath)
		if err != nil {
			fatal(err)
		}
	}
	if *urlValue != "" {
		request.URL = *urlValue
	}
	if *textPath != "" {
		body, readErr := os.ReadFile(*textPath)
		if readErr != nil {
			fatal(readErr)
		}
		request.SourceText = string(body)
	}

	var accessed time.Time
	if *accessedValue != "" {
		accessed, err = time.Parse("2006-01-02", *accessedValue)
		if err != nil {
			fatal(fmt.Errorf("invalid --accessed date: %w", err))
		}
	}
	result, err := recipeimport.Import(context.Background(), request, recipeimport.Options{
		OutDir:   *outDir,
		Accessed: accessed,
		APIKey:   os.Getenv("OPENAI_API_KEY"),
		Model:    *model,
		NoAI:     *noAI,
		DryRun:   *dryRun,
	})
	if err != nil {
		fatal(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "recipe import failed: %v\n", err)
	os.Exit(1)
}
