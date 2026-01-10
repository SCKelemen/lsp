package adapter_3_17

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
	protocol316 "github.com/SCKelemen/lsp/protocol_3_16"
)

func TestCoreToProtocolPosition(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		corePos  core.Position
		wantLine uint32
		wantChar uint32
	}{
		{
			name:     "ASCII",
			content:  "hello world",
			corePos:  core.Position{Line: 0, Character: 5},
			wantLine: 0,
			wantChar: 5,
		},
		{
			name:     "emoji - after surrogate pair",
			content:  "hello 😀 world",
			corePos:  core.Position{Line: 0, Character: 10}, // space after emoji
			wantLine: 0,
			wantChar: 8, // UTF-16: h(0) e(1) l(2) l(3) o(4) sp(5) emoji(6,7) sp(8)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CoreToProtocolPosition(tt.corePos, tt.content)
			if got.Line != protocol316.UInteger(tt.wantLine) {
				t.Errorf("Line: got %v, want %v", got.Line, tt.wantLine)
			}
			if got.Character != protocol316.UInteger(tt.wantChar) {
				t.Errorf("Character: got %v, want %v", got.Character, tt.wantChar)
			}
		})
	}
}

func TestProtocolToCorePosition(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		protocolPos protocol316.Position
		wantLine    int
		wantChar    int
	}{
		{
			name:        "ASCII",
			content:     "hello world",
			protocolPos: protocol316.Position{Line: 0, Character: 5},
			wantLine:    0,
			wantChar:    5,
		},
		{
			name:        "emoji - after surrogate pair",
			content:     "hello 😀 world",
			protocolPos: protocol316.Position{Line: 0, Character: 8}, // space after emoji in UTF-16
			wantLine:    0,
			wantChar:    10, // space after emoji in UTF-8 (10 bytes)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProtocolToCorePosition(tt.protocolPos, tt.content)
			if got.Line != tt.wantLine {
				t.Errorf("Line: got %v, want %v", got.Line, tt.wantLine)
			}
			if got.Character != tt.wantChar {
				t.Errorf("Character: got %v, want %v", got.Character, tt.wantChar)
			}
		})
	}
}

func TestCoreToProtocolDiagnostic(t *testing.T) {
	content := "hello 😀 world"
	severity := core.SeverityError
	coreDiag := core.Diagnostic{
		Range: core.Range{
			Start: core.Position{Line: 0, Character: 0},
			End:   core.Position{Line: 0, Character: 6}, // start of emoji
		},
		Severity: &severity,
		Message:  "Test error",
		Source:   "test",
	}

	protocolDiag := CoreToProtocolDiagnostic(coreDiag, content)

	if protocolDiag.Message != "Test error" {
		t.Errorf("Message: got %v, want %v", protocolDiag.Message, "Test error")
	}
	if protocolDiag.Range.Start.Character != 0 {
		t.Errorf("Start character: got %v, want 0", protocolDiag.Range.Start.Character)
	}
	if protocolDiag.Range.End.Character != 6 {
		t.Errorf("End character: got %v, want 6", protocolDiag.Range.End.Character)
	}
	if protocolDiag.Severity == nil || *protocolDiag.Severity != protocol316.DiagnosticSeverity(core.SeverityError) {
		t.Errorf("Severity: got %v, want %v", protocolDiag.Severity, core.SeverityError)
	}
}

func TestCoreToProtocolDiagnostics(t *testing.T) {
	content := "hello world"
	severity := core.SeverityWarning
	coreDiags := []core.Diagnostic{
		{
			Range: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 0, Character: 5},
			},
			Severity: &severity,
			Message:  "Warning 1",
			Source:   "test",
		},
		{
			Range: core.Range{
				Start: core.Position{Line: 0, Character: 6},
				End:   core.Position{Line: 0, Character: 11},
			},
			Severity: &severity,
			Message:  "Warning 2",
			Source:   "test",
		},
	}

	protocolDiags := CoreToProtocolDiagnostics(coreDiags, content)

	if len(protocolDiags) != 2 {
		t.Fatalf("Expected 2 diagnostics, got %d", len(protocolDiags))
	}
	if protocolDiags[0].Message != "Warning 1" {
		t.Errorf("First diagnostic message: got %v, want %v", protocolDiags[0].Message, "Warning 1")
	}
	if protocolDiags[1].Message != "Warning 2" {
		t.Errorf("Second diagnostic message: got %v, want %v", protocolDiags[1].Message, "Warning 2")
	}
}

func TestProtocolToCoreDiagnostics(t *testing.T) {
	content := "hello world"
	severity := protocol316.DiagnosticSeverityWarning
	protocolDiags := []protocol316.Diagnostic{
		{
			Range: protocol316.Range{
				Start: protocol316.Position{Line: 0, Character: 0},
				End:   protocol316.Position{Line: 0, Character: 5},
			},
			Severity: &severity,
			Message:  "Warning 1",
		},
	}

	coreDiags := ProtocolToCoreDiagnostics(protocolDiags, content)

	if len(coreDiags) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(coreDiags))
	}
	if coreDiags[0].Message != "Warning 1" {
		t.Errorf("Message: got %v, want %v", coreDiags[0].Message, "Warning 1")
	}
	if coreDiags[0].Severity == nil || *coreDiags[0].Severity != core.SeverityWarning {
		t.Errorf("Severity: got %v, want %v", coreDiags[0].Severity, core.SeverityWarning)
	}
}
