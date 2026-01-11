package adapter_3_16

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
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
		{
			name:     "Chinese characters",
			content:  "你好世界",
			corePos:  core.Position{Line: 0, Character: 6}, // after 2 chars (6 bytes)
			wantLine: 0,
			wantChar: 2, // 2 UTF-16 code units
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CoreToProtocolPosition(tt.corePos, tt.content)
			if got.Line != protocol.UInteger(tt.wantLine) {
				t.Errorf("Line: got %v, want %v", got.Line, tt.wantLine)
			}
			if got.Character != protocol.UInteger(tt.wantChar) {
				t.Errorf("Character: got %v, want %v", got.Character, tt.wantChar)
			}
		})
	}
}

func TestProtocolToCorePosition(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		protocolPos protocol.Position
		wantLine    int
		wantChar    int
	}{
		{
			name:        "ASCII",
			content:     "hello world",
			protocolPos: protocol.Position{Line: 0, Character: 5},
			wantLine:    0,
			wantChar:    5,
		},
		{
			name:        "emoji - after surrogate pair",
			content:     "hello 😀 world",
			protocolPos: protocol.Position{Line: 0, Character: 8}, // space after emoji in UTF-16
			wantLine:    0,
			wantChar:    10, // space after emoji in UTF-8 (10 bytes)
		},
		{
			name:        "Chinese characters",
			content:     "你好世界",
			protocolPos: protocol.Position{Line: 0, Character: 2}, // 2 UTF-16 code units
			wantLine:    0,
			wantChar:    6, // 6 UTF-8 bytes
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

func TestRoundTripPosition(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		pos     core.Position
	}{
		{"ASCII", "hello world", core.Position{Line: 0, Character: 5}},
		{"Emoji", "hello 😀 world", core.Position{Line: 0, Character: 6}}, // start of emoji
		{"Chinese", "你好世界", core.Position{Line: 0, Character: 3}},     // start of second char
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Core -> Protocol -> Core
			protocolPos := CoreToProtocolPosition(tc.pos, tc.content)
			roundTrip := ProtocolToCorePosition(protocolPos, tc.content)

			if roundTrip != tc.pos {
				t.Errorf("Round trip failed: %v -> %v -> %v",
					tc.pos, protocolPos, roundTrip)
			}
		})
	}
}

func TestCoreToProtocolRange(t *testing.T) {
	content := "hello 😀 world"
	coreRange := core.Range{
		Start: core.Position{Line: 0, Character: 0},
		End:   core.Position{Line: 0, Character: 6}, // start of emoji
	}

	protocolRange := CoreToProtocolRange(coreRange, content)

	if protocolRange.Start.Character != 0 {
		t.Errorf("Start character: got %v, want 0", protocolRange.Start.Character)
	}
	if protocolRange.End.Character != 6 {
		t.Errorf("End character: got %v, want 6", protocolRange.End.Character)
	}
}
