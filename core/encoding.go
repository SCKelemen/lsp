// Package core provides encoding conversion utilities for UTF-8 and UTF-16.
//
// LSP Specification requires UTF-16 code units for positions:
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocuments
//
// "Character offset in a line is measured in UTF-16 code units. So a string of the form
// 'a𐐀b' has length 4, not 3, since 𐐀 is represented by a surrogate pair in UTF-16."
//
// This package handles conversion between:
// - UTF-8 byte offsets (used in core types for natural Go string handling)
// - UTF-16 code unit offsets (used in LSP protocol per specification)
package core

import (
	"strings"
	"unicode/utf8"
)

// UTF8ToUTF16Offset converts a UTF-8 byte offset to a UTF-16 code unit offset.
// This is used when converting core types to protocol types.
//
// The line parameter is zero-based, and offset is the UTF-8 byte offset within the line.
// Returns the equivalent UTF-16 code unit offset within the same line.
//
// Character Boundary Handling: If the offset lands in the middle of a multi-byte
// UTF-8 character, it rounds DOWN to the start of that character. This ensures
// positions always point to valid character boundaries.
//
// LSP Spec Reference: UTF-16 code units are required by the LSP specification
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position
func UTF8ToUTF16Offset(content string, line int, utf8Offset int) int {
	// Find the byte offset for the line
	lineStart := 0
	for currentLine := 0; currentLine < line; currentLine++ {
		nextLineIdx := strings.Index(content[lineStart:], "\n")
		if nextLineIdx == -1 {
			// Line doesn't exist, return 0
			return 0
		}
		lineStart += nextLineIdx + 1
	}

	// Get the line content
	lineEnd := strings.Index(content[lineStart:], "\n")
	var lineContent string
	if lineEnd == -1 {
		lineContent = content[lineStart:]
	} else {
		lineContent = content[lineStart : lineStart+lineEnd]
	}

	// Ensure offset is within the line
	if utf8Offset > len(lineContent) {
		utf8Offset = len(lineContent)
	}

	// Convert UTF-8 byte offset to UTF-16 code unit offset
	utf16Offset := 0
	utf8Count := 0
	for utf8Count < utf8Offset && utf8Count < len(lineContent) {
		r, size := utf8.DecodeRuneInString(lineContent[utf8Count:])
		if r == utf8.RuneError {
			// Invalid UTF-8, skip
			utf8Count++
			utf16Offset++
			continue
		}

		// Check if consuming this character would overshoot the target
		if utf8Count+size > utf8Offset {
			// Would land in the middle of this character, stop before it
			break
		}

		utf8Count += size

		// Runes >= 0x10000 are encoded as surrogate pairs in UTF-16 (2 code units)
		if r >= 0x10000 {
			utf16Offset += 2
		} else {
			utf16Offset++
		}
	}

	return utf16Offset
}

// UTF16ToUTF8Offset converts a UTF-16 code unit offset to a UTF-8 byte offset.
// This is used when converting protocol types to core types.
//
// The line parameter is zero-based, and utf16Offset is the UTF-16 code unit offset within the line.
// Returns the equivalent UTF-8 byte offset within the same line.
//
// Character Boundary Handling: If the offset lands in the middle of a surrogate pair
// (e.g., emoji or characters >= U+10000), it rounds DOWN to the start of that character.
//
// Surrogate Pairs: Characters outside the BMP (U+10000 to U+10FFFF) are encoded as
// surrogate pairs in UTF-16, using 2 code units. For example, '𐐀' (U+10400) requires
// 2 UTF-16 code units but 4 UTF-8 bytes.
//
// LSP Spec Reference: UTF-16 code units are required by the LSP specification
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position
func UTF16ToUTF8Offset(content string, line int, utf16Offset int) int {
	// Find the byte offset for the line
	lineStart := 0
	for currentLine := 0; currentLine < line; currentLine++ {
		nextLineIdx := strings.Index(content[lineStart:], "\n")
		if nextLineIdx == -1 {
			// Line doesn't exist, return 0
			return 0
		}
		lineStart += nextLineIdx + 1
	}

	// Get the line content
	lineEnd := strings.Index(content[lineStart:], "\n")
	var lineContent string
	if lineEnd == -1 {
		lineContent = content[lineStart:]
	} else {
		lineContent = content[lineStart : lineStart+lineEnd]
	}

	// Convert UTF-16 code unit offset to UTF-8 byte offset
	utf8Offset := 0
	utf16Count := 0
	for utf16Count < utf16Offset && utf8Offset < len(lineContent) {
		r, size := utf8.DecodeRuneInString(lineContent[utf8Offset:])
		if r == utf8.RuneError {
			// Invalid UTF-8, skip
			utf8Offset++
			utf16Count++
			continue
		}

		// Check if consuming this character would overshoot the target
		runeUTF16Size := 1
		if r >= 0x10000 {
			runeUTF16Size = 2
		}

		if utf16Count+runeUTF16Size > utf16Offset {
			// Would land in the middle of this character, stop before it
			break
		}

		utf8Offset += size
		utf16Count += runeUTF16Size
	}

	return utf8Offset
}

// ByteOffsetToPosition converts an absolute byte offset in a document to a Position.
// This is useful for converting absolute byte positions (e.g., from parsers) to
// line/character positions compatible with LSP.
//
// The returned Position uses UTF-8 byte offsets (core format). Use adapters to
// convert to UTF-16 code units for LSP protocol messages.
//
// Note: Positions are zero-based per LSP specification.
func ByteOffsetToPosition(content string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}

	line := 0
	lineStart := 0

	// Find which line the offset is on
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	// Character offset is the UTF-8 byte offset from the start of the line
	character := offset - lineStart

	return Position{
		Line:      line,
		Character: character,
	}
}

// PositionToByteOffset converts a Position to an absolute byte offset in the document.
// This is the inverse of ByteOffsetToPosition.
//
// The input Position should use UTF-8 byte offsets (core format). If you have a
// protocol Position with UTF-16 code units, convert it first using adapter functions.
//
// If the position is beyond the end of the document, returns len(content).
// If the character offset is beyond the end of the line, clamps to end of line.
func PositionToByteOffset(content string, pos Position) int {
	offset := 0
	currentLine := 0

	// Skip to the target line
	for i := 0; i < len(content) && currentLine < pos.Line; i++ {
		if content[i] == '\n' {
			currentLine++
			offset = i + 1
		}
	}

	// If we didn't reach the target line, clamp to end of content
	if currentLine < pos.Line {
		return len(content)
	}

	// Add the character offset (but don't go past the end of the line)
	lineEnd := strings.Index(content[offset:], "\n")
	if lineEnd == -1 {
		lineEnd = len(content) - offset
	}

	characterOffset := pos.Character
	if characterOffset > lineEnd {
		characterOffset = lineEnd
	}

	return offset + characterOffset
}
