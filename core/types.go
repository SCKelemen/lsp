// Package core provides protocol-agnostic types for language server operations.
// These types are optimized for Go usage and CLI tools, using UTF-8 byte offsets
// instead of LSP's UTF-16 code units. Adapters handle conversion to/from protocol types.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/
package core

import "fmt"

// Position represents a position in a text document using zero-based line and UTF-8 byte offset.
//
// IMPORTANT: Unlike LSP protocol Position (which uses UTF-16 code units per the spec),
// this core type uses UTF-8 byte offsets for natural Go string indexing. Use adapter
// functions to convert between core and protocol types at API boundaries.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position
// Protocol Position uses UTF-16 code units: { line: uinteger, character: uinteger }
// Core Position uses UTF-8 byte offsets: { Line: int, Character: int }
type Position struct {
	// Line is the zero-based line number (compatible with LSP spec)
	Line int
	// Character is the zero-based UTF-8 byte offset within the line
	// (LSP spec uses UTF-16 code units; conversion happens in adapters)
	Character int
}

// String returns a human-readable representation of the position
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Character)
}

// IsValid returns true if the position has non-negative coordinates
func (p Position) IsValid() bool {
	return p.Line >= 0 && p.Character >= 0
}

// Range represents a text range in a document with start and end positions.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#range
// Protocol: { start: Position, end: Position }
// Both start and end positions use UTF-8 offsets in core (UTF-16 in protocol).
type Range struct {
	Start Position
	End   Position
}

// String returns a human-readable representation of the range
func (r Range) String() string {
	return fmt.Sprintf("%s-%s", r.Start, r.End)
}

// IsValid returns true if both positions are valid and start comes before or equals end
func (r Range) IsValid() bool {
	if !r.Start.IsValid() || !r.End.IsValid() {
		return false
	}
	if r.Start.Line > r.End.Line {
		return false
	}
	if r.Start.Line == r.End.Line && r.Start.Character > r.End.Character {
		return false
	}
	return true
}

// Contains returns true if position p is within this range
func (r Range) Contains(p Position) bool {
	if p.Line < r.Start.Line || p.Line > r.End.Line {
		return false
	}
	if p.Line == r.Start.Line && p.Character < r.Start.Character {
		return false
	}
	if p.Line == r.End.Line && p.Character > r.End.Character {
		return false
	}
	return true
}

// Location represents a location in a resource (file, document, etc.) with a URI and range.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#location
// Protocol: { uri: DocumentUri, range: Range }
type Location struct {
	// URI is the resource identifier (file path, URL, etc.)
	URI string
	// Range is the text range within the resource
	Range Range
}

// String returns a human-readable representation of the location
func (l Location) String() string {
	return fmt.Sprintf("%s:%s", l.URI, l.Range)
}

// IsValid returns true if the URI is non-empty and the range is valid
func (l Location) IsValid() bool {
	return l.URI != "" && l.Range.IsValid()
}

// LocationLink represents a link between two locations, typically for navigation features.
// The origin is where the user initiated the action, and the target is where to navigate.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#locationLink
// Protocol: { originSelectionRange?: Range, targetUri: DocumentUri, targetRange: Range, targetSelectionRange: Range }
// Used for "Go to Definition" with better UX than plain Location.
type LocationLink struct {
	// OriginSelectionRange is the range in the origin document that is linked (optional)
	OriginSelectionRange *Range
	// TargetURI is the URI of the target resource
	TargetURI string
	// TargetRange is the full range of the target symbol (including comments, etc.)
	TargetRange Range
	// TargetSelectionRange is the range that should be selected and revealed (e.g., symbol name)
	TargetSelectionRange Range
}

// IsValid returns true if the target URI is non-empty and ranges are valid
func (ll LocationLink) IsValid() bool {
	if ll.TargetURI == "" || !ll.TargetRange.IsValid() || !ll.TargetSelectionRange.IsValid() {
		return false
	}
	if ll.OriginSelectionRange != nil && !ll.OriginSelectionRange.IsValid() {
		return false
	}
	return true
}
