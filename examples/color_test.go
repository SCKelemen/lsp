package examples

import (
	"math"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestColorProvider_ProvideDocumentColors(t *testing.T) {
	provider := &ColorProvider{}

	tests := []struct {
		name      string
		content   string
		wantCount int
		wantColor *core.Color // Check first color if provided
	}{
		{
			name:      "hex 6 digit",
			content:   "color: #FF0000;",
			wantCount: 1,
			wantColor: &core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0},
		},
		{
			name:      "hex 3 digit",
			content:   "color: #F00;",
			wantCount: 1,
			wantColor: &core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0},
		},
		{
			name:      "hex 8 digit with alpha",
			content:   "color: #FF000080;",
			wantCount: 1,
			wantColor: &core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 0.5},
		},
		{
			name:      "rgb format",
			content:   "color: rgb(255, 128, 0);",
			wantCount: 1,
			wantColor: &core.Color{Red: 1.0, Green: 0.5019607843137255, Blue: 0.0, Alpha: 1.0},
		},
		{
			name:      "rgba format",
			content:   "color: rgba(255, 0, 0, 0.5);",
			wantCount: 1,
			wantColor: &core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 0.5},
		},
		{
			name: "multiple colors",
			content: `
.red { color: #FF0000; }
.green { color: #00FF00; }
.blue { color: rgb(0, 0, 255); }
`,
			wantCount: 3,
		},
		{
			name:      "no colors",
			content:   "just some text without colors",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colors := provider.ProvideDocumentColors("test.css", tt.content)

			if len(colors) != tt.wantCount {
				t.Errorf("expected %d colors, got %d", tt.wantCount, len(colors))
			}

			if tt.wantColor != nil && len(colors) > 0 {
				got := colors[0].Color
				if !colorsEqual(got, *tt.wantColor) {
					t.Errorf("expected color %+v, got %+v", *tt.wantColor, got)
				}
			}
		})
	}
}

func TestColorProvider_ProvideColorPresentations(t *testing.T) {
	provider := &ColorProvider{}

	color := core.Color{
		Red:   1.0,
		Green: 0.5,
		Blue:  0.0,
		Alpha: 1.0,
	}

	rng := core.Range{
		Start: core.Position{Line: 0, Character: 0},
		End:   core.Position{Line: 0, Character: 7},
	}

	presentations := provider.ProvideColorPresentations("test.css", "", color, rng)

	if len(presentations) < 3 {
		t.Fatalf("expected at least 3 presentations (hex, rgb, rgba), got %d", len(presentations))
	}

	// Check that we have different formats
	labels := make(map[string]bool)
	for _, p := range presentations {
		labels[p.Label] = true

		// Verify each presentation has a text edit
		if p.TextEdit == nil {
			t.Errorf("presentation %q has no text edit", p.Label)
		}
	}

	// Should have different presentation formats
	if len(labels) < 3 {
		t.Errorf("expected at least 3 different presentation formats, got %d", len(labels))
	}
}

func TestColorProvider_ParseHexColor(t *testing.T) {
	provider := &ColorProvider{}

	tests := []struct {
		name    string
		hex     string
		want    core.Color
		wantOk  bool
	}{
		{
			name:   "3 digit",
			hex:    "#F00",
			want:   core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0},
			wantOk: true,
		},
		{
			name:   "6 digit",
			hex:    "#FF8000",
			want:   core.Color{Red: 1.0, Green: 0.5019607843137255, Blue: 0.0, Alpha: 1.0},
			wantOk: true,
		},
		{
			name:   "8 digit with alpha",
			hex:    "#FF000080",
			want:   core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 0.5019607843137255},
			wantOk: true,
		},
		{
			name:   "lowercase",
			hex:    "#ff0000",
			want:   core.Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0},
			wantOk: true,
		},
		{
			name:   "invalid length",
			hex:    "#FF00",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := provider.parseHexColor(tt.hex)

			if ok != tt.wantOk {
				t.Errorf("parseHexColor(%q) ok = %v, want %v", tt.hex, ok, tt.wantOk)
			}

			if tt.wantOk && !colorsEqual(got, tt.want) {
				t.Errorf("parseHexColor(%q) = %+v, want %+v", tt.hex, got, tt.want)
			}
		})
	}
}

func TestColorProvider_ColorToFormats(t *testing.T) {
	provider := &ColorProvider{}

	color := core.Color{
		Red:   1.0,
		Green: 0.5019607843137255,
		Blue:  0.0,
		Alpha: 0.5,
	}

	hex := provider.colorToHex(color)
	if hex != "#FF80007F" && hex != "#FF800080" {
		t.Errorf("colorToHex() = %q, want #FF80007F or #FF800080", hex)
	}

	rgb := provider.colorToRGB(color)
	if rgb != "rgb(255, 128, 0)" {
		t.Errorf("colorToRGB() = %q, want rgb(255, 128, 0)", rgb)
	}

	rgba := provider.colorToRGBA(color)
	if rgba != "rgba(255, 128, 0, 0.50)" {
		t.Errorf("colorToRGBA() = %q, want rgba(255, 128, 0, 0.50)", rgba)
	}
}

// colorsEqual compares two colors with a small epsilon for floating point comparison
func colorsEqual(a, b core.Color) bool {
	const epsilon = 0.01
	return math.Abs(a.Red-b.Red) < epsilon &&
		math.Abs(a.Green-b.Green) < epsilon &&
		math.Abs(a.Blue-b.Blue) < epsilon &&
		math.Abs(a.Alpha-b.Alpha) < epsilon
}
