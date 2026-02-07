package layout

import "testing"

func TestComputePlacements_DefaultLetterLandscape_25x4p(t *testing.T) {
	paper, err := ParsePaper("letter")
	if err != nil {
		t.Fatalf("ParsePaper: %v", err)
	}
	orient, err := ParseOrientation("landscape")
	if err != nil {
		t.Fatalf("ParseOrientation: %v", err)
	}
	card, err := ParseCardSize("2.5x4p")
	if err != nil {
		t.Fatalf("ParseCardSize: %v", err)
	}
	cfg := DefaultConfig(paper, orient, card)

	pl, perPage, err := cfg.ComputePlacements(9)
	if err != nil {
		t.Fatalf("ComputePlacements: %v", err)
	}
	if perPage != 8 {
		t.Fatalf("expected perPage=8, got %d", perPage)
	}
	if len(pl) != 9 {
		t.Fatalf("expected 9 placements, got %d", len(pl))
	}
	if pl[0].PageIndex != 0 || pl[8].PageIndex != 1 {
		t.Fatalf("expected page indices 0..1, got %d..%d", pl[0].PageIndex, pl[8].PageIndex)
	}
}
