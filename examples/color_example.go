package examples

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// ColorProvider provides color information for various file types.
// It detects color values in different formats (hex, rgb, rgba) and provides
// multiple presentation formats.
type ColorProvider struct{}

var (
	// Hex colors: #RGB, #RRGGBB, #RRGGBBAA
	hexColorRegex = regexp.MustCompile(`#([0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})\b`)

	// RGB/RGBA: rgb(255, 0, 0), rgba(255, 0, 0, 0.5)
	rgbColorRegex = regexp.MustCompile(`rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*([\d.]+))?\s*\)`)
)

// ProvideDocumentColors finds all color values in the document
func (p *ColorProvider) ProvideDocumentColors(uri, content string) []core.ColorInformation {
	var colors []core.ColorInformation

	// Find hex colors
	colors = append(colors, p.findHexColors(content)...)

	// Find RGB/RGBA colors
	colors = append(colors, p.findRGBColors(content)...)

	return colors
}

// ProvideColorPresentations provides different ways to represent a color
func (p *ColorProvider) ProvideColorPresentations(uri, content string, color core.Color, rng core.Range) []core.ColorPresentation {
	presentations := []core.ColorPresentation{
		// Hex format
		{
			Label: p.colorToHex(color),
			TextEdit: &core.TextEdit{
				Range:   rng,
				NewText: p.colorToHex(color),
			},
		},
		// RGB format
		{
			Label: p.colorToRGB(color),
			TextEdit: &core.TextEdit{
				Range:   rng,
				NewText: p.colorToRGB(color),
			},
		},
		// RGBA format (always include for completeness)
		{
			Label: p.colorToRGBA(color),
			TextEdit: &core.TextEdit{
				Range:   rng,
				NewText: p.colorToRGBA(color),
			},
		},
	}

	return presentations
}

// findHexColors finds all hexadecimal color values
func (p *ColorProvider) findHexColors(content string) []core.ColorInformation {
	var colors []core.ColorInformation

	matches := hexColorRegex.FindAllStringIndex(content, -1)
	for _, match := range matches {
		start, end := match[0], match[1]
		colorStr := content[start:end]

		color, ok := p.parseHexColor(colorStr)
		if !ok {
			continue
		}

		colors = append(colors, core.ColorInformation{
			Range: core.Range{
				Start: positionFromOffset(content, start),
				End:   positionFromOffset(content, end),
			},
			Color: color,
		})
	}

	return colors
}

// findRGBColors finds all RGB/RGBA color values
func (p *ColorProvider) findRGBColors(content string) []core.ColorInformation {
	var colors []core.ColorInformation

	matches := rgbColorRegex.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		if len(match) < 8 {
			continue
		}

		start, end := match[0], match[1]

		// Extract RGB values
		rStr := content[match[2]:match[3]]
		gStr := content[match[4]:match[5]]
		bStr := content[match[6]:match[7]]

		r, _ := strconv.Atoi(rStr)
		g, _ := strconv.Atoi(gStr)
		b, _ := strconv.Atoi(bStr)

		// Extract alpha if present
		alpha := 1.0
		if match[8] != -1 && match[9] != -1 {
			aStr := content[match[8]:match[9]]
			alpha, _ = strconv.ParseFloat(aStr, 64)
		}

		color := core.Color{
			Red:   float64(r) / 255.0,
			Green: float64(g) / 255.0,
			Blue:  float64(b) / 255.0,
			Alpha: alpha,
		}

		colors = append(colors, core.ColorInformation{
			Range: core.Range{
				Start: positionFromOffset(content, start),
				End:   positionFromOffset(content, end),
			},
			Color: color,
		})
	}

	return colors
}

// parseHexColor converts a hex color string to a Color
func (p *ColorProvider) parseHexColor(hex string) (core.Color, bool) {
	// Remove '#' prefix
	hex = strings.TrimPrefix(hex, "#")

	var r, g, b, a int

	switch len(hex) {
	case 3: // #RGB -> #RRGGBB
		fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
		a = 255
	case 6: // #RRGGBB
		fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		a = 255
	case 8: // #RRGGBBAA
		fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
	default:
		return core.Color{}, false
	}

	return core.Color{
		Red:   float64(r) / 255.0,
		Green: float64(g) / 255.0,
		Blue:  float64(b) / 255.0,
		Alpha: float64(a) / 255.0,
	}, true
}

// colorToHex converts a Color to hex format
func (p *ColorProvider) colorToHex(color core.Color) string {
	r := int(color.Red * 255)
	g := int(color.Green * 255)
	b := int(color.Blue * 255)

	if color.Alpha < 1.0 {
		a := int(color.Alpha * 255)
		return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	}

	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// colorToRGB converts a Color to rgb() format
func (p *ColorProvider) colorToRGB(color core.Color) string {
	r := int(color.Red * 255)
	g := int(color.Green * 255)
	b := int(color.Blue * 255)

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// colorToRGBA converts a Color to rgba() format
func (p *ColorProvider) colorToRGBA(color core.Color) string {
	r := int(color.Red * 255)
	g := int(color.Green * 255)
	b := int(color.Blue * 255)

	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, color.Alpha)
}
