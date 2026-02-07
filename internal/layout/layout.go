package layout

import (
	"fmt"
	"strings"
)

const (
	PointsPerInch = 72.0
)

type Paper struct {
	Name   string
	Width  float64 // points
	Height float64 // points
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

type CardSize struct {
	Name   string
	Width  float64 // points
	Height float64 // points
}

type Config struct {
	Paper        Paper
	Orientation  Orientation
	Card         CardSize
	PerPage      int
	MarginLeft   float64
	MarginRight  float64
	MarginTop    float64
	MarginBottom float64
	GutterX      float64
	GutterY      float64

	DrawCutLines  bool
	DrawCropMarks bool
}

func ParsePaper(s string) (Paper, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "letter":
		return Paper{Name: "letter", Width: 8.5 * PointsPerInch, Height: 11 * PointsPerInch}, nil
	case "a4":
		return Paper{Name: "a4", Width: 210.0 / 25.4 * PointsPerInch, Height: 297.0 / 25.4 * PointsPerInch}, nil
	default:
		return Paper{}, fmt.Errorf("unknown paper: %q (expected letter|a4)", s)
	}
}

func ParseOrientation(s string) (Orientation, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "portrait":
		return Portrait, nil
	case "landscape":
		return Landscape, nil
	default:
		return Portrait, fmt.Errorf("unknown orientation: %q (expected portrait|landscape)", s)
	}
}

func ParseCardSize(s string) (CardSize, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "2.5x4p", "2.5x4":
		return CardSize{Name: "2.5x4p", Width: 2.5 * PointsPerInch, Height: 4 * PointsPerInch}, nil
	case "2x4p", "2x4":
		return CardSize{Name: "2x4p", Width: 2 * PointsPerInch, Height: 4 * PointsPerInch}, nil
	case "4x2":
		return CardSize{Name: "4x2", Width: 4 * PointsPerInch, Height: 2 * PointsPerInch}, nil
	default:
		return CardSize{}, fmt.Errorf("unknown card size: %q (expected 2.5x4p|2x4p|4x2)", s)
	}
}

func DefaultConfig(paper Paper, orient Orientation, card CardSize) Config {
	cfg := Config{
		Paper:        paper,
		Orientation:  orient,
		Card:         card,
		PerPage:      8,
		MarginLeft:   0.25 * PointsPerInch,
		MarginRight:  0.25 * PointsPerInch,
		MarginTop:    0.25 * PointsPerInch,
		MarginBottom: 0.25 * PointsPerInch,
		GutterX:      (1.0 / 6.0) * PointsPerInch, // ~0.167"
		GutterY:      0,
		DrawCutLines: true,
	}
	return cfg
}

func (c Config) PageSize() (w, h float64) {
	w, h = c.Paper.Width, c.Paper.Height
	if c.Orientation == Landscape {
		return h, w
	}
	return w, h
}

type Placement struct {
	PageIndex int
	X         float64 // points, from left
	Y         float64 // points, from bottom
	W         float64
	H         float64
	Index     int // card index in overall list
}

func (c Config) ComputePlacements(numCards int) ([]Placement, int, error) {
	pw, ph := c.PageSize()
	cw, ch := c.Card.Width, c.Card.Height

	usableW := pw - c.MarginLeft - c.MarginRight
	usableH := ph - c.MarginTop - c.MarginBottom
	if usableW <= 0 || usableH <= 0 {
		return nil, 0, fmt.Errorf("margins too large for page")
	}

	cols := maxInt(1, int((usableW+c.GutterX)/(cw+c.GutterX)))
	rows := maxInt(1, int((usableH+c.GutterY)/(ch+c.GutterY)))
	perPage := cols * rows
	if perPage <= 0 {
		return nil, 0, fmt.Errorf("card geometry does not fit on page")
	}

	placements := make([]Placement, 0, numCards)
	for i := 0; i < numCards; i++ {
		page := i / perPage
		pos := i % perPage
		r := pos / cols
		col := pos % cols

		x := c.MarginLeft + float64(col)*(cw+c.GutterX)
		// y is from bottom; place from top down.
		top := c.MarginTop + float64(r)*(ch+c.GutterY)
		y := ph - top - ch

		placements = append(placements, Placement{
			PageIndex: page,
			X:         x,
			Y:         y,
			W:         cw,
			H:         ch,
			Index:     i,
		})
	}
	return placements, perPage, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
