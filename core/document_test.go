package core

import "testing"

func TestDocumentApplyEditIgnoresInvalidRange(t *testing.T) {
	doc := NewDocument("file:///test.txt", "hello world", 1)

	doc.ApplyEdit(Range{
		Start: Position{Line: 1, Character: 0},
		End:   Position{Line: 0, Character: 0},
	}, "X")

	if got := doc.GetContent(); got != "hello world" {
		t.Fatalf("content changed for invalid range, got %q", got)
	}
	if doc.Version != 1 {
		t.Fatalf("version changed for invalid range, got %d", doc.Version)
	}
}

func TestDocumentManagerUpdateIncrementsVersion(t *testing.T) {
	dm := NewDocumentManager()
	doc := dm.Open("file:///test.txt", "before", 7)

	if ok := dm.Update("file:///test.txt", "after"); !ok {
		t.Fatal("expected update to succeed")
	}

	if got := doc.GetContent(); got != "after" {
		t.Fatalf("unexpected content after update: %q", got)
	}
	if doc.Version != 8 {
		t.Fatalf("expected version 8, got %d", doc.Version)
	}
}
