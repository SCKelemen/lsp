// Package core provides document highlight types for showing symbol occurrences.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_documentHighlight
package core

// DocumentHighlightKind defines the kind of document highlight.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#documentHighlightKind
// Values: Text = 1, Read = 2, Write = 3
type DocumentHighlightKind int

const (
	// DocumentHighlightKindText indicates a textual occurrence.
	DocumentHighlightKindText DocumentHighlightKind = 1
	// DocumentHighlightKindRead indicates read-access of a symbol (e.g., reading a variable).
	DocumentHighlightKindRead DocumentHighlightKind = 2
	// DocumentHighlightKindWrite indicates write-access of a symbol (e.g., writing to a variable).
	DocumentHighlightKindWrite DocumentHighlightKind = 3
)

// DocumentHighlight represents a range to highlight in a document.
// Used to show all occurrences of a symbol when the cursor is on it.
//
// LSP Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#documentHighlight
// Protocol: { range: Range, kind?: DocumentHighlightKind }
type DocumentHighlight struct {
	// Range is the range this highlight applies to (UTF-8 offsets in core).
	Range Range

	// Kind is the highlight kind (Text, Read, Write).
	// If nil, defaults to Text.
	Kind *DocumentHighlightKind
}

// DocumentHighlightContext provides context for document highlight requests.
type DocumentHighlightContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string

	// Position is where the cursor is positioned (UTF-8 offset in core).
	Position Position
}

// DocumentHighlightProvider provides document highlights.
// This interface can be implemented to provide "find all references" style highlighting.
type DocumentHighlightProvider interface {
	// ProvideDocumentHighlights returns highlights for the symbol at the position.
	// Returns nil or empty slice if no highlights are available.
	ProvideDocumentHighlights(ctx DocumentHighlightContext) []DocumentHighlight
}
