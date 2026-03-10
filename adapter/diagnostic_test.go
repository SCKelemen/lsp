package adapter_3_16

import (
	"encoding/json"
	"testing"

	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

func TestProtocolToCoreDiagnosticIntCode(t *testing.T) {
	var diag protocol.Diagnostic
	payload := []byte(`{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":1}},"message":"x","code":123}`)
	if err := json.Unmarshal(payload, &diag); err != nil {
		t.Fatalf("failed to unmarshal diagnostic: %v", err)
	}

	got := ProtocolToCoreDiagnostic(diag, "a")
	if got.Code == nil {
		t.Fatal("expected integer diagnostic code to be preserved")
	}
	if !got.Code.IsInt || got.Code.IntValue != 123 {
		t.Fatalf("unexpected diagnostic code: %#v", got.Code)
	}
}

func TestCoreToProtocolDiagnosticWithResolverRelatedInfo(t *testing.T) {
	mainContent := "abcdef"
	otherURI := "file:///other.txt"
	otherContent := "😀x"

	diag := core.Diagnostic{
		Range:   core.Range{Start: core.Position{Line: 0, Character: 1}, End: core.Position{Line: 0, Character: 2}},
		Message: "main",
		RelatedInformation: []core.DiagnosticRelatedInformation{
			{
				Location: core.Location{
					URI: otherURI,
					Range: core.Range{
						Start: core.Position{Line: 0, Character: 4}, // after emoji in UTF-8
						End:   core.Position{Line: 0, Character: 5},
					},
				},
				Message: "related",
			},
		},
	}

	got := CoreToProtocolDiagnosticWithResolver(diag, mainContent, func(uri string) (string, bool) {
		if uri == otherURI {
			return otherContent, true
		}
		return "", false
	})

	if len(got.RelatedInformation) != 1 {
		t.Fatalf("expected 1 related info entry, got %d", len(got.RelatedInformation))
	}

	start := got.RelatedInformation[0].Location.Range.Start.Character
	end := got.RelatedInformation[0].Location.Range.End.Character
	if start != 2 || end != 3 {
		t.Fatalf("expected related UTF-16 chars [2,3], got [%d,%d]", start, end)
	}
}

func TestProtocolToCoreDiagnosticWithResolverRelatedInfo(t *testing.T) {
	mainContent := "abcdef"
	otherURI := "file:///other.txt"
	otherContent := "😀x"

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 1},
			End:   protocol.Position{Line: 0, Character: 2},
		},
		Message: "main",
		RelatedInformation: []protocol.DiagnosticRelatedInformation{
			{
				Location: protocol.Location{
					URI: protocol.DocumentUri(otherURI),
					Range: protocol.Range{
						Start: protocol.Position{Line: 0, Character: 2}, // after emoji in UTF-16
						End:   protocol.Position{Line: 0, Character: 3},
					},
				},
				Message: "related",
			},
		},
	}

	got := ProtocolToCoreDiagnosticWithResolver(diag, mainContent, func(uri string) (string, bool) {
		if uri == otherURI {
			return otherContent, true
		}
		return "", false
	})

	if len(got.RelatedInformation) != 1 {
		t.Fatalf("expected 1 related info entry, got %d", len(got.RelatedInformation))
	}

	start := got.RelatedInformation[0].Location.Range.Start.Character
	end := got.RelatedInformation[0].Location.Range.End.Character
	if start != 4 || end != 5 {
		t.Fatalf("expected related UTF-8 chars [4,5], got [%d,%d]", start, end)
	}
}
