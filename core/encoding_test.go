package core

import (
	"testing"
)

func TestUTF8ToUTF16Offset(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		line       int
		utf8Offset int
		want       int
	}{
		{
			name:       "simple ASCII",
			content:    "hello world",
			line:       0,
			utf8Offset: 5,
			want:       5,
		},
		{
			name:       "emoji (surrogate pair)",
			content:    "hello 😀 world",
			line:       0,
			utf8Offset: 9, // byte offset within emoji (last byte)
			want:       6,  // UTF-16 offset (start of emoji - rounds down)
		},
		{
			name:       "multi-line with emoji",
			content:    "first line\nsecond 😀 line",
			line:       1,
			utf8Offset: 10, // byte offset within emoji on second line
			want:       7,   // UTF-16 offset (start of emoji - rounds down)
		},
		{
			name:       "Chinese characters",
			content:    "你好世界",
			line:       0,
			utf8Offset: 6, // 2 Chinese chars (3 bytes each)
			want:       2,  // 2 UTF-16 code units
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UTF8ToUTF16Offset(tt.content, tt.line, tt.utf8Offset)
			if got != tt.want {
				t.Errorf("UTF8ToUTF16Offset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUTF16ToUTF8Offset(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		line        int
		utf16Offset int
		want        int
	}{
		{
			name:        "simple ASCII",
			content:     "hello world",
			line:        0,
			utf16Offset: 5,
			want:        5,
		},
		{
			name:        "emoji (surrogate pair)",
			content:     "hello 😀 world",
			line:        0,
			utf16Offset: 8, // UTF-16 offset after emoji (emoji takes 2 code units at 6-7)
			want:        10,  // byte offset after emoji
		},
		{
			name:        "multi-line with emoji",
			content:     "first line\nsecond 😀 line",
			line:        1,
			utf16Offset: 9, // UTF-16 offset after emoji
			want:        11, // byte offset after emoji on second line
		},
		{
			name:        "Chinese characters",
			content:     "你好世界",
			line:        0,
			utf16Offset: 2, // 2 UTF-16 code units
			want:        6,  // 2 Chinese chars (3 bytes each)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UTF16ToUTF8Offset(tt.content, tt.line, tt.utf16Offset)
			if got != tt.want {
				t.Errorf("UTF16ToUTF8Offset() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"ASCII only", "hello world\nfoo bar"},
		{"With emoji", "hello 😀 world\n🎉 party"},
		{"Chinese", "你好世界\n再见"},
		{"Mixed", "Hello 世界 😀\nfoo bar 你好"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test round trip for every position in the content
			lines := 0
			for _, ch := range tc.content {
				if ch == '\n' {
					lines++
				}
			}

			for line := 0; line <= lines; line++ {
				// Find line start
				lineStart := 0
				currentLine := 0
				for i := 0; i < len(tc.content) && currentLine < line; i++ {
					if tc.content[i] == '\n' {
						currentLine++
						lineStart = i + 1
					}
				}

				// Find line end
				lineEnd := lineStart
				for lineEnd < len(tc.content) && tc.content[lineEnd] != '\n' {
					lineEnd++
				}

				// Test every UTF-8 byte offset in the line
				for utf8Offset := 0; utf8Offset <= (lineEnd - lineStart); utf8Offset++ {
					utf16Offset := UTF8ToUTF16Offset(tc.content, line, utf8Offset)
					backToUTF8 := UTF16ToUTF8Offset(tc.content, line, utf16Offset)

					// The round trip should get us back to the same position or earlier
					// (in case we land in the middle of a multi-byte character)
					if backToUTF8 > utf8Offset {
						t.Errorf("Round trip failed: line=%d, utf8=%d -> utf16=%d -> utf8=%d",
							line, utf8Offset, utf16Offset, backToUTF8)
					}
				}
			}
		})
	}
}

func TestByteOffsetToPosition(t *testing.T) {
	content := "hello world\nfoo bar\nbaz"

	tests := []struct {
		offset int
		want   Position
	}{
		{0, Position{0, 0}},
		{5, Position{0, 5}},
		{11, Position{0, 11}},   // end of first line
		{12, Position{1, 0}},    // start of second line
		{15, Position{1, 3}},    // "foo"
		{20, Position{2, 0}},    // start of third line
		{100, Position{2, 3}},   // past end (clamped)
		{-1, Position{0, 0}},    // negative (clamped)
	}

	for _, tt := range tests {
		got := ByteOffsetToPosition(content, tt.offset)
		if got != tt.want {
			t.Errorf("ByteOffsetToPosition(%d) = %v, want %v", tt.offset, got, tt.want)
		}
	}
}

func TestPositionToByteOffset(t *testing.T) {
	content := "hello world\nfoo bar\nbaz"

	tests := []struct {
		pos  Position
		want int
	}{
		{Position{0, 0}, 0},
		{Position{0, 5}, 5},
		{Position{0, 11}, 11},  // end of first line
		{Position{1, 0}, 12},   // start of second line
		{Position{1, 3}, 15},   // "foo"
		{Position{2, 0}, 20},   // start of third line
		{Position{2, 100}, 23}, // past end of line (clamped)
		{Position{10, 0}, 23},  // past end of content
	}

	for _, tt := range tests {
		got := PositionToByteOffset(content, tt.pos)
		if got != tt.want {
			t.Errorf("PositionToByteOffset(%v) = %d, want %d", tt.pos, got, tt.want)
		}
	}
}
