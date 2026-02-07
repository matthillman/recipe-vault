package recipes

import "testing"

func TestLoadFile_StandardSourdough(t *testing.T) {
	r, err := LoadFile("../../recipes/standard-sourdough-loaf.md")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if r.Title == "" {
		t.Fatalf("expected title")
	}

	yield := r.FindYieldLines()
	if len(yield) == 0 {
		t.Fatalf("expected yield lines")
	}

	ings := r.ExtractListItems("Ingredients")
	if len(ings) < 5 {
		t.Fatalf("expected ingredients, got %d", len(ings))
	}

	proc := r.ExtractListItems("Process")
	if len(proc) < 4 {
		t.Fatalf("expected process steps, got %d", len(proc))
	}

	if !r.HasSectionPrefix("Formula") {
		t.Fatalf("expected formula section")
	}
	form := r.ExtractFormulaLines()
	if len(form) < 2 {
		t.Fatalf("expected formula lines, got %d", len(form))
	}
}

func TestLoadFile_FocacciaSplitRecipe(t *testing.T) {
	r, err := LoadFile("../../recipes/overnight-sourdough-focaccia.md")
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	ings := r.ExtractListItems("Ingredients")
	if len(ings) < 5 {
		t.Fatalf("expected ingredients from ### Ingredients, got %d", len(ings))
	}
	proc := r.ExtractListItems("Process")
	if len(proc) < 4 {
		t.Fatalf("expected process from ### Process, got %d", len(proc))
	}

	// This recipe doesn't have a dedicated formula section; we treat baker's % as part of ingredients.
	if r.HasSectionPrefix("Formula") {
		t.Fatalf("did not expect explicit formula section")
	}
}
