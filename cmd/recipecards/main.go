package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"recipelab/internal/cards"
	"recipelab/internal/layout"
	"recipelab/internal/pdf"
	"recipelab/internal/recipes"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "list":
		cmdList(os.Args[2:])
	case "build":
		cmdBuild(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(strings.TrimSpace(`
recipecards: generate printable recipe reference cards

Usage:
  recipecards list [--dir recipes]
  recipecards build --out out/cards.pdf --recipes slug1,slug2 [options]

Options (build):
  --dir recipes             recipes directory
  --out out/cards.pdf       output PDF path
  --recipes slugs           comma-separated recipe slugs (filenames without .md)
  --paper letter|a4         default: letter
  --page-orient landscape|portrait  default: landscape
  --card 2.5x4p|2x4p|4x2    default: 2.5x4p
  --cut-lines               draw card outlines (default: true)
  --crop-marks              draw crop marks (default: false)
`))
}

func cmdList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dir := fs.String("dir", "recipes", "recipes directory")
	_ = fs.Parse(args)

	all, err := recipes.LoadAll(*dir)
	if err != nil {
		fatal(err)
	}

	slugs := make([]string, 0, len(all))
	for slug := range all {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		r := all[slug]
		fmt.Printf("%s\t%s\t%s\n", slug, r.Title, r.Path)
	}
}

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	dir := fs.String("dir", "recipes", "recipes directory")
	outPath := fs.String("out", filepath.FromSlash("out/cards.pdf"), "output PDF path")
	recipeSlugs := fs.String("recipes", "", "comma-separated recipe slugs (filenames without .md)")
	paperName := fs.String("paper", "letter", "paper size: letter|a4")
	orientName := fs.String("page-orient", "landscape", "page orientation: landscape|portrait")
	cardName := fs.String("card", "2.5x4p", "card size: 2.5x4p|2x4p|4x2")
	cutLines := fs.Bool("cut-lines", true, "draw card outlines")
	cropMarks := fs.Bool("crop-marks", false, "draw crop marks")
	_ = fs.Parse(args)

	if strings.TrimSpace(*recipeSlugs) == "" {
		fatal(fmt.Errorf("--recipes is required (comma-separated slugs). Run: recipecards list"))
	}

	all, err := recipes.LoadAll(*dir)
	if err != nil {
		fatal(err)
	}

	want := splitCSV(*recipeSlugs)
	selected := make([]recipes.Recipe, 0, len(want))
	for _, slug := range want {
		r, ok := all[slug]
		if !ok {
			fatal(fmt.Errorf("unknown recipe slug %q (run: recipecards list)", slug))
		}
		selected = append(selected, r)
	}

	paper, err := layout.ParsePaper(*paperName)
	if err != nil {
		fatal(err)
	}
	orient, err := layout.ParseOrientation(*orientName)
	if err != nil {
		fatal(err)
	}
	cardSize, err := layout.ParseCardSize(*cardName)
	if err != nil {
		fatal(err)
	}

	cfg := layout.DefaultConfig(paper, orient, cardSize)
	cfg.DrawCutLines = *cutLines
	cfg.DrawCropMarks = *cropMarks

	doc := pdf.NewDocument(pdf.DocumentOptions{
		Title: "Recipe Cards",
	})

	cds := make([]cards.Card, 0, len(selected))
	for _, r := range selected {
		cds = append(cds, cards.FromRecipe(r))
	}

	if err := cards.RenderToPDF(doc, cfg, cds); err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, doc.Bytes(), 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s\n", *outPath)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
