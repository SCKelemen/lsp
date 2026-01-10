package core

// DocumentHighlightKind defines the kind of document highlight.
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
type DocumentHighlight struct {
	// Range is the range this highlight applies to (UTF-8 offsets).
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

	// Position is where the cursor is positioned (UTF-8 offset).
	Position Position
}

// DocumentHighlightProvider provides document highlights.
type DocumentHighlightProvider interface {
	// ProvideDocumentHighlights returns highlights for the symbol at the position.
	// Returns nil or empty slice if no highlights are available.
	ProvideDocumentHighlights(ctx DocumentHighlightContext) []DocumentHighlight
}
