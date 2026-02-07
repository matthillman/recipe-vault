package cards

import (
	"fmt"
	"sort"
	"strings"

	"recipelab/internal/layout"
	"recipelab/internal/pdf"
	"recipelab/internal/recipes"
)

type Card struct {
	Slug  string
	Title string

	Lines []string // if provided, treated as "Card override" body

	Yield       []string
	Ingredients []string
	Formula     []string
	Process     []string
}

func FromRecipe(r recipes.Recipe) Card {
	c := Card{
		Slug:  r.Slug,
		Title: r.Title,
	}

	if override := r.CardOverride(); len(override) > 0 {
		c.Lines = override
		return c
	}

	c.Yield = r.FindYieldLines()
	c.Ingredients = r.ExtractListItems("Ingredients")
	c.Process = r.ExtractListItems("Process")
	if r.HasSectionPrefix("Formula") {
		c.Formula = r.ExtractFormulaLines()
	}

	return c
}

func RenderToPDF(doc *pdf.Document, cfg layout.Config, cards []Card) error {
	placements, perPage, err := cfg.ComputePlacements(len(cards))
	if err != nil {
		return err
	}

	// Ensure pages exist.
	pw, ph := cfg.PageSize()
	numPages := 0
	if len(cards) > 0 {
		numPages = (len(cards) + perPage - 1) / perPage
	}
	for i := 0; i < numPages; i++ {
		doc.AddPage(pw, ph)
	}

	// Render.
	for _, pl := range placements {
		c := cards[pl.Index]
		renderCard(doc, cfg, pl.PageIndex, pl.X, pl.Y, pl.W, pl.H, c)
	}
	return nil
}

func renderCard(doc *pdf.Document, cfg layout.Config, page int, x, y, w, h float64, c Card) {
	if cfg.DrawCutLines {
		doc.SetGrayStroke(page, 0.2)
		doc.SetLineWidth(page, 0.5)
		doc.DrawRect(page, x, y, w, h)
	}

	if cfg.DrawCropMarks {
		drawCropMarks(doc, page, x, y, w, h)
	}

	// Card padding.
	pad := 8.0
	tx := x + pad
	tyTop := y + h - pad // from bottom, so this is top area
	maxW := w - 2*pad
	maxH := h - 2*pad

	// Title.
	titleSize := 10.0
	lineGap := 2.0
	titleLines := wrapText(c.Title, pdf.HelveticaBold, titleSize, maxW)
	if len(titleLines) > 2 {
		titleLines = titleLines[:2]
	}

	curY := tyTop
	for _, line := range titleLines {
		doc.Text(page, pdf.HelveticaBold, titleSize, tx, curY-titleSize, line)
		curY -= titleSize + lineGap
	}
	curY -= 2 // small separator

	// Body.
	bodySize := 7.0
	labelSize := 7.0
	lineH := bodySize + 2

	var blocks []block
	if len(c.Lines) > 0 {
		blocks = []block{{Label: "", Lines: c.Lines}}
	} else {
		if len(c.Yield) > 0 {
			blocks = append(blocks, block{Label: "Yield", Lines: c.Yield})
		}
		if len(c.Ingredients) > 0 {
			blocks = append(blocks, block{Label: "Ingredients", Lines: c.Ingredients})
		}
		if len(c.Formula) > 0 && formulaLooksUseful(c.Formula) {
			blocks = append(blocks, block{Label: "Baker's % / Formula", Lines: c.Formula})
		}
		if len(c.Process) > 0 {
			blocks = append(blocks, block{Label: "Process", Lines: c.Process})
		}
	}

	// If there's a lot, keep things bounded so cards stay readable.
	applyBlockLimits(blocks)

	// Render blocks until we run out of vertical space.
	minY := y + pad
	startY := curY
	_ = maxH

	for bi, b := range blocks {
		if b.Label != "" {
			if curY-lineH < minY {
				break
			}
			doc.Text(page, pdf.HelveticaBold, labelSize, tx, curY-labelSize, b.Label)
			curY -= lineH
		}

		for _, raw := range b.Lines {
			lines := wrapText(raw, pdf.Helvetica, bodySize, maxW)
			for _, line := range lines {
				if curY-lineH < minY {
					// Overflow: mark truncation.
					doc.Text(page, pdf.Helvetica, bodySize, tx, minY, "...")
					curY = minY
					goto footer
				}
				doc.Text(page, pdf.Helvetica, bodySize, tx, curY-bodySize, line)
				curY -= lineH
			}
		}

		// Spacer between blocks.
		if bi != len(blocks)-1 {
			curY -= 2
		}
	}

footer:
	// Slug footer (tiny).
	_ = startY
	footerSize := 6.0
	doc.SetGrayStroke(page, 0.5)
	doc.Text(page, pdf.Helvetica, footerSize, x+pad, y+pad/2, fmt.Sprintf("[%s]", c.Slug))
}

type block struct {
	Label string
	Lines []string
}

func applyBlockLimits(blocks []block) {
	for i := range blocks {
		switch blocks[i].Label {
		case "Ingredients":
			blocks[i].Lines = limitLines(blocks[i].Lines, 10)
		case "Process":
			blocks[i].Lines = limitLines(blocks[i].Lines, 5)
		case "Baker's % / Formula":
			blocks[i].Lines = limitLines(blocks[i].Lines, 6)
		case "Yield":
			blocks[i].Lines = limitLines(blocks[i].Lines, 2)
		}
	}
}

func limitLines(lines []string, n int) []string {
	if n <= 0 || len(lines) <= n {
		return lines
	}
	return lines[:n]
}

func formulaLooksUseful(lines []string) bool {
	// Avoid duplicating formula lines when ingredients already contain % everywhere.
	// If formula has any line that looks like a summary (Total flour / hydration / salt), keep it.
	for _, l := range lines {
		low := strings.ToLower(l)
		if strings.Contains(low, "hydration") || strings.Contains(low, "total flour") || strings.Contains(low, "starter") {
			return true
		}
		if strings.Contains(l, ":") && strings.Contains(l, "%") {
			return true
		}
	}
	return false
}

func drawCropMarks(doc *pdf.Document, page int, x, y, w, h float64) {
	// Simple 6pt crop marks that extend outward from the card corners.
	m := 6.0
	doc.SetGrayStroke(page, 0.2)
	doc.SetLineWidth(page, 0.5)

	// Bottom-left
	doc.DrawLine(page, x-m, y, x, y)
	doc.DrawLine(page, x, y-m, x, y)
	// Bottom-right
	doc.DrawLine(page, x+w, y, x+w+m, y)
	doc.DrawLine(page, x+w, y-m, x+w, y)
	// Top-left
	doc.DrawLine(page, x-m, y+h, x, y+h)
	doc.DrawLine(page, x, y+h, x, y+h+m)
	// Top-right
	doc.DrawLine(page, x+w, y+h, x+w+m, y+h)
	doc.DrawLine(page, x+w, y+h, x+w, y+h+m)
}

func wrapText(s string, f pdf.Font, size float64, maxWidth float64) []string {
	// We use a conservative, dependency-free width estimate. It won't be typographically perfect,
	// but it's consistent and good enough for small reference cards.
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	words := splitWords(s)
	var lines []string
	var cur string

	flush := func() {
		if strings.TrimSpace(cur) != "" {
			lines = append(lines, strings.TrimSpace(cur))
		}
		cur = ""
	}

	for _, w := range words {
		cand := w
		if cur != "" {
			cand = cur + " " + w
		}
		if estimateTextWidthPt(cand, f, size) <= maxWidth {
			cur = cand
			continue
		}
		if cur != "" {
			flush()
		}

		// If the single word is too wide, hard-wrap it.
		if estimateTextWidthPt(w, f, size) > maxWidth {
			parts := hardWrap(w, f, size, maxWidth)
			lines = append(lines, parts...)
			continue
		}
		cur = w
	}
	flush()
	return lines
}

func splitWords(s string) []string {
	// Keep punctuation attached; just split on whitespace.
	fields := strings.Fields(s)
	return fields
}

func hardWrap(word string, f pdf.Font, size float64, maxWidth float64) []string {
	var out []string
	start := 0
	for start < len(word) {
		end := start + 1
		for end <= len(word) && estimateTextWidthPt(word[start:end], f, size) <= maxWidth {
			end++
		}
		// end is now one past the limit; step back.
		end--
		if end <= start {
			// Emergency escape.
			end = minInt(start+1, len(word))
		}
		out = append(out, word[start:end])
		start = end
	}
	return out
}

func estimateTextWidthPt(s string, f pdf.Font, size float64) float64 {
	// Approximate widths for Helvetica/Courier:
	// - Courier is monospaced (600 units per glyph).
	// - Helvetica uses an embedded width table for ASCII.
	units := 0
	if f == pdf.Courier || f == pdf.CourierBold {
		units = 600 * len(s)
	} else {
		for i := 0; i < len(s); i++ {
			c := s[i]
			units += helveticaWidth(c)
		}
	}
	return float64(units) * size / 1000.0
}

func helveticaWidth(c byte) int {
	// Widths in 1/1000 em for Helvetica, ASCII 32-126. Unknown chars fall back to 500.
	if c < 32 || c > 126 {
		return 500
	}
	return helvWidths[c-32]
}

// A small Helvetica width table for wrapping. (ASCII 32..126)
var helvWidths = []int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556,
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556,
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556,
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584,
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure imports stay used in some build tags.
var _ = sort.Strings
