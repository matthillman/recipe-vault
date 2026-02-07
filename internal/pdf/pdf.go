package pdf

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

// This is a minimal PDF writer tailored for this repository.
// It supports pages, basic drawing, and text with the built-in PDF Type1 fonts.

type DocumentOptions struct {
	Title string
}

type Document struct {
	opts  DocumentOptions
	pages []page
}

type page struct {
	w, h   float64
	stream bytes.Buffer
}

func NewDocument(opts DocumentOptions) *Document {
	return &Document{opts: opts}
}

func (d *Document) AddPage(widthPt, heightPt float64) int {
	d.pages = append(d.pages, page{w: widthPt, h: heightPt})
	return len(d.pages) - 1
}

func (d *Document) PageWidth(i int) float64  { return d.pages[i].w }
func (d *Document) PageHeight(i int) float64 { return d.pages[i].h }

func (d *Document) DrawLine(pageIndex int, x1, y1, x2, y2 float64) {
	p := &d.pages[pageIndex]
	fmt.Fprintf(&p.stream, "%.2f %.2f m %.2f %.2f l S\n", x1, y1, x2, y2)
}

func (d *Document) DrawRect(pageIndex int, x, y, w, h float64) {
	p := &d.pages[pageIndex]
	fmt.Fprintf(&p.stream, "%.2f %.2f %.2f %.2f re S\n", x, y, w, h)
}

func (d *Document) SetLineWidth(pageIndex int, w float64) {
	p := &d.pages[pageIndex]
	fmt.Fprintf(&p.stream, "%.2f w\n", w)
}

func (d *Document) SetGrayStroke(pageIndex int, g float64) {
	p := &d.pages[pageIndex]
	fmt.Fprintf(&p.stream, "%.3f G\n", g)
}

type Font string

const (
	Helvetica     Font = "Helvetica"
	HelveticaBold Font = "Helvetica-Bold"
	Courier       Font = "Courier"
	CourierBold   Font = "Courier-Bold"
)

func (d *Document) Text(pageIndex int, font Font, size float64, x, y float64, s string) {
	p := &d.pages[pageIndex]
	s = sanitizeText(s)
	s = escapePDFString(s)
	// BT ... ET for a single line.
	fmt.Fprintf(&p.stream, "BT /%s %.2f Tf 1 0 0 1 %.2f %.2f Tm (%s) Tj ET\n", fontResourceName(font), size, x, y, s)
}

// Bytes renders the PDF as a byte slice.
func (d *Document) Bytes() []byte {
	// Object numbering:
	// 1: Catalog
	// 2: Pages
	// 3..: Page objects + content streams
	// Fonts are added once.

	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n%\xD3\xEB\xE9\xE1\n")

	type obj struct {
		id   int
		data []byte
	}
	var objs []obj

	// Fonts (fixed resources).
	fontObjIDs := map[Font]int{
		Helvetica:     3,
		HelveticaBold: 4,
		Courier:       5,
		CourierBold:   6,
	}

	objs = append(objs, obj{fontObjIDs[Helvetica], []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")})
	objs = append(objs, obj{fontObjIDs[HelveticaBold], []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>")})
	objs = append(objs, obj{fontObjIDs[Courier], []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>")})
	objs = append(objs, obj{fontObjIDs[CourierBold], []byte("<< /Type /Font /Subtype /Type1 /BaseFont /Courier-Bold >>")})

	// Pages tree and catalog reserved:
	pagesObjID := 2
	catalogObjID := 1

	// Page objects start after fonts.
	nextID := 7
	pageIDs := make([]int, 0, len(d.pages))
	contentIDs := make([]int, 0, len(d.pages))

	for range d.pages {
		pageIDs = append(pageIDs, nextID)
		nextID++
		contentIDs = append(contentIDs, nextID)
		nextID++
	}

	// Page and content objects.
	for i, p := range d.pages {
		content := p.stream.Bytes()
		contentObj := fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content)
		objs = append(objs, obj{contentIDs[i], []byte(contentObj)})

		res := fmt.Sprintf("<< /Font << /F1 %d 0 R /F2 %d 0 R /F3 %d 0 R /F4 %d 0 R >> >>",
			fontObjIDs[Helvetica], fontObjIDs[HelveticaBold], fontObjIDs[Courier], fontObjIDs[CourierBold],
		)
		pageObj := fmt.Sprintf("<< /Type /Page /Parent %d 0 R /MediaBox [0 0 %.2f %.2f] /Resources %s /Contents %d 0 R >>",
			pagesObjID, p.w, p.h, res, contentIDs[i],
		)
		objs = append(objs, obj{pageIDs[i], []byte(pageObj)})
	}

	// Pages object (Kids array).
	var kids strings.Builder
	for _, pid := range pageIDs {
		fmt.Fprintf(&kids, "%d 0 R ", pid)
	}
	pagesObj := fmt.Sprintf("<< /Type /Pages /Count %d /Kids [ %s] >>", len(pageIDs), kids.String())
	objs = append(objs, obj{pagesObjID, []byte(pagesObj)})

	// Catalog object.
	catalogObj := fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesObjID)
	objs = append(objs, obj{catalogObjID, []byte(catalogObj)})

	// Optional info dictionary.
	infoID := nextID
	nextID++
	if strings.TrimSpace(d.opts.Title) != "" {
		now := time.Now().UTC().Format("20060102150405")
		info := fmt.Sprintf("<< /Title (%s) /Producer (recipecards) /CreationDate (D:%sZ) >>",
			escapePDFString(sanitizeText(d.opts.Title)), now,
		)
		objs = append(objs, obj{infoID, []byte(info)})
	} else {
		infoID = 0
	}

	// Sort by ID for stable xref.
	sort.Slice(objs, func(i, j int) bool { return objs[i].id < objs[j].id })

	offsets := make(map[int]int, len(objs)+1)
	// Object 0 is free.
	offsets[0] = 0

	for _, o := range objs {
		offsets[o.id] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", o.id, o.data)
	}

	// xref
	xrefStart := out.Len()
	maxID := 0
	for _, o := range objs {
		if o.id > maxID {
			maxID = o.id
		}
	}

	fmt.Fprintf(&out, "xref\n0 %d\n", maxID+1)
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxID; i++ {
		off, ok := offsets[i]
		if !ok {
			// Shouldn't happen, but keep PDF readable.
			out.WriteString("0000000000 00000 f \n")
			continue
		}
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}

	// trailer
	if infoID != 0 {
		fmt.Fprintf(&out, "trailer\n<< /Size %d /Root %d 0 R /Info %d 0 R >>\n", maxID+1, catalogObjID, infoID)
	} else {
		fmt.Fprintf(&out, "trailer\n<< /Size %d /Root %d 0 R >>\n", maxID+1, catalogObjID)
	}
	fmt.Fprintf(&out, "startxref\n%d\n%%%%EOF\n", xrefStart)
	return out.Bytes()
}

func fontResourceName(f Font) string {
	// Use short resource names so content streams stay small.
	switch f {
	case Helvetica:
		return "F1"
	case HelveticaBold:
		return "F2"
	case Courier:
		return "F3"
	case CourierBold:
		return "F4"
	default:
		return "F1"
	}
}

func escapePDFString(s string) string {
	// Escape backslashes and parentheses.
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	return s
}

func sanitizeText(s string) string {
	// Best-effort ASCII normalization for our Markdown corpus.
	repl := strings.NewReplacer(
		"\u201C", `"`, // left double quote
		"\u201D", `"`, // right double quote
		"\u2018", `'`, // left single quote
		"\u2019", `'`, // right single quote
		"\u2013", "-", // en dash
		"\u2014", "-", // em dash
		"\u00D7", "x", // multiplication sign
		"\u00BD", "1/2",
		"\u00BC", "1/4",
		"\u00BE", "3/4",
		"\u00B0", " deg ",
		"\u00F1", "n", // ñ
		"\u00E9", "e",
	)
	s = repl.Replace(s)

	// Strip remaining non-ASCII.
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
		} else if r == '\n' || r == '\t' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
