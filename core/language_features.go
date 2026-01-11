package core

// FoldingRangeKind defines the kind of folding range.
type FoldingRangeKind string

const (
	// FoldingRangeKindComment indicates a folding range for a comment.
	FoldingRangeKindComment FoldingRangeKind = "comment"
	// FoldingRangeKindImports indicates a folding range for imports or includes.
	FoldingRangeKindImports FoldingRangeKind = "imports"
	// FoldingRangeKindRegion indicates a folding range for a region (e.g., #region).
	FoldingRangeKindRegion FoldingRangeKind = "region"
)

// FoldingRange represents a range in a document that can be folded (collapsed).
// Unlike protocol FoldingRange which uses UTF-16 offsets, this uses UTF-8 byte offsets.
type FoldingRange struct {
	// StartLine is the zero-based start line of the range to fold.
	StartLine int

	// StartCharacter is the UTF-8 byte offset from where the folded range starts.
	// If nil, defaults to the length of the start line.
	StartCharacter *int

	// EndLine is the zero-based end line of the range to fold.
	EndLine int

	// EndCharacter is the UTF-8 byte offset before the folded range ends.
	// If nil, defaults to the length of the end line.
	EndCharacter *int

	// Kind describes the kind of the folding range (comment, region, imports).
	Kind *FoldingRangeKind
}

// TextEdit represents a textual edit to a document.
// The range and replacement text use UTF-8 byte offsets.
type TextEdit struct {
	// Range is the range of text to be replaced.
	Range Range

	// NewText is the string to replace the range with.
	NewText string
}

// AnnotatedTextEdit represents a text edit with an optional annotation.
type AnnotatedTextEdit struct {
	TextEdit

	// AnnotationID is an optional identifier for the edit.
	// This can be used to group related edits or provide additional context.
	AnnotationID *string
}

// TextDocumentEdit represents edits to a single text document.
type TextDocumentEdit struct {
	// TextDocument identifies the document to change.
	TextDocument VersionedTextDocumentIdentifier

	// Edits is the list of edits to apply to the document.
	Edits []TextEdit
}

// VersionedTextDocumentIdentifier identifies a specific version of a text document.
type VersionedTextDocumentIdentifier struct {
	// URI is the text document's URI.
	URI string

	// Version is the version number of the document.
	// Null indicates that the version is not known.
	Version *int
}

// CreateFile represents an operation to create a file.
type CreateFile struct {
	// URI is the resource to create.
	URI string

	// Options contains additional options for creating the file.
	Options *CreateFileOptions
}

// CreateFileOptions contains options for creating a file.
type CreateFileOptions struct {
	// Overwrite existing file if it exists. Overwrite wins over ignoreIfExists.
	Overwrite bool

	// IgnoreIfExists will not create the file if it already exists.
	IgnoreIfExists bool
}

// RenameFile represents an operation to rename a file.
type RenameFile struct {
	// OldURI is the old resource URI.
	OldURI string

	// NewURI is the new resource URI.
	NewURI string

	// Options contains additional options for renaming the file.
	Options *RenameFileOptions
}

// RenameFileOptions contains options for renaming a file.
type RenameFileOptions struct {
	// Overwrite existing file if it exists. Overwrite wins over ignoreIfExists.
	Overwrite bool

	// IgnoreIfExists will not rename the file if the new URI already exists.
	IgnoreIfExists bool
}

// DeleteFile represents an operation to delete a file.
type DeleteFile struct {
	// URI is the file to delete.
	URI string

	// Options contains additional options for deleting the file.
	Options *DeleteFileOptions
}

// DeleteFileOptions contains options for deleting a file.
type DeleteFileOptions struct {
	// Recursive deletes the directory and all its contents if the URI is a directory.
	Recursive bool

	// IgnoreIfNotExists will not fail if the file does not exist.
	IgnoreIfNotExists bool
}

// WorkspaceEdit represents changes to many resources managed in the workspace.
type WorkspaceEdit struct {
	// Changes maps document URIs to arrays of text edits.
	Changes map[string][]TextEdit

	// DocumentChanges is the preferred way to specify edits.
	// If provided, Changes should be ignored.
	DocumentChanges []interface{} // TextDocumentEdit | CreateFile | RenameFile | DeleteFile

	// ChangeAnnotations maps annotation IDs to change annotations.
	ChangeAnnotations map[string]ChangeAnnotation
}

// ChangeAnnotation represents additional information about a change.
type ChangeAnnotation struct {
	// Label is a human-readable string describing the change.
	Label string

	// NeedsConfirmation indicates whether the change needs user confirmation.
	NeedsConfirmation bool

	// Description is an optional longer description of the change.
	Description string
}

// SymbolKind defines the kind of a symbol.
type SymbolKind int

const (
	SymbolKindFile          SymbolKind = 1
	SymbolKindModule        SymbolKind = 2
	SymbolKindNamespace     SymbolKind = 3
	SymbolKindPackage       SymbolKind = 4
	SymbolKindClass         SymbolKind = 5
	SymbolKindMethod        SymbolKind = 6
	SymbolKindProperty      SymbolKind = 7
	SymbolKindField         SymbolKind = 8
	SymbolKindConstructor   SymbolKind = 9
	SymbolKindEnum          SymbolKind = 10
	SymbolKindInterface     SymbolKind = 11
	SymbolKindFunction      SymbolKind = 12
	SymbolKindVariable      SymbolKind = 13
	SymbolKindConstant      SymbolKind = 14
	SymbolKindString        SymbolKind = 15
	SymbolKindNumber        SymbolKind = 16
	SymbolKindBoolean       SymbolKind = 17
	SymbolKindArray         SymbolKind = 18
	SymbolKindObject        SymbolKind = 19
	SymbolKindKey           SymbolKind = 20
	SymbolKindNull          SymbolKind = 21
	SymbolKindEnumMember    SymbolKind = 22
	SymbolKindStruct        SymbolKind = 23
	SymbolKindEvent         SymbolKind = 24
	SymbolKindOperator      SymbolKind = 25
	SymbolKindTypeParameter SymbolKind = 26
)

// SymbolTag defines tags for symbols.
type SymbolTag int

const (
	// SymbolTagDeprecated indicates the symbol is deprecated.
	SymbolTagDeprecated SymbolTag = 1
)

// DocumentSymbol represents programming constructs like variables, classes, interfaces, etc.
// that appear in a document. Uses UTF-8 byte offsets for ranges.
type DocumentSymbol struct {
	// Name is the name of this symbol.
	Name string

	// Detail provides additional information about this symbol.
	Detail string

	// Kind is the kind of this symbol.
	Kind SymbolKind

	// Tags are tags for this symbol.
	Tags []SymbolTag

	// Deprecated indicates if this symbol is deprecated.
	Deprecated bool

	// Range is the range enclosing this symbol (including leading/trailing whitespace and comments).
	Range Range

	// SelectionRange is the range that should be selected and revealed when navigating to this symbol.
	// This is typically the range of the symbol's name.
	SelectionRange Range

	// Children are symbols that are children of this symbol (e.g., properties of a class).
	Children []DocumentSymbol
}

// CodeActionKind defines the kind of a code action.
type CodeActionKind string

const (
	// CodeActionKindEmpty is the empty kind.
	CodeActionKindEmpty CodeActionKind = ""
	// CodeActionKindQuickFix is a quick fix action.
	CodeActionKindQuickFix CodeActionKind = "quickfix"
	// CodeActionKindRefactor is a refactoring action.
	CodeActionKindRefactor CodeActionKind = "refactor"
	// CodeActionKindRefactorExtract extracts code into a new entity.
	CodeActionKindRefactorExtract CodeActionKind = "refactor.extract"
	// CodeActionKindRefactorInline inlines code.
	CodeActionKindRefactorInline CodeActionKind = "refactor.inline"
	// CodeActionKindRefactorRewrite rewrites code.
	CodeActionKindRefactorRewrite CodeActionKind = "refactor.rewrite"
	// CodeActionKindSource is a source action (e.g., organize imports).
	CodeActionKindSource CodeActionKind = "source"
	// CodeActionKindSourceOrganizeImports organizes imports.
	CodeActionKindSourceOrganizeImports CodeActionKind = "source.organizeImports"
	// CodeActionKindSourceFixAll fixes all auto-fixable problems.
	CodeActionKindSourceFixAll CodeActionKind = "source.fixAll"
)

// CodeAction represents a change that can be performed in code.
// Uses UTF-8 byte offsets for all ranges in diagnostics and edits.
type CodeAction struct {
	// Title is a short, human-readable title for this code action.
	Title string

	// Kind is the kind of this code action.
	Kind *CodeActionKind

	// Diagnostics are the diagnostics that this code action resolves.
	Diagnostics []Diagnostic

	// IsPreferred indicates that this code action is the preferred action in a group.
	IsPreferred bool

	// Disabled indicates the reason why this code action is disabled.
	Disabled *CodeActionDisabled

	// Edit is the workspace edit this code action performs.
	Edit *WorkspaceEdit

	// Command is a command to execute instead of or in addition to the edit.
	Command *Command

	// Data is arbitrary data preserved between textDocument/codeAction and codeAction/resolve.
	Data interface{}
}

// CodeActionDisabled explains why a code action is disabled.
type CodeActionDisabled struct {
	// Reason is the human-readable reason why the code action is disabled.
	Reason string
}

// Command represents a reference to a command.
type Command struct {
	// Title is the title of the command.
	Title string

	// Tooltip is an optional tooltip when hovering over the command.
	// @since 3.18.0
	Tooltip string

	// Command is the identifier of the command to execute.
	Command string

	// Arguments are optional arguments for the command.
	Arguments []interface{}
}

// DocumentLink represents a link inside a document.
// The link can point to a URL, a file, or another location in the workspace.
// Uses UTF-8 byte offsets for the range.
type DocumentLink struct {
	// Range is the range where the link appears in the document.
	Range Range

	// Target is the URI that the link points to.
	// If nil, the link should be resolved lazily via DocumentLinkResolveProvider.
	Target *string

	// Tooltip is the tooltip text when hovering over this link.
	Tooltip *string

	// Data is arbitrary data preserved between textDocument/documentLink and documentLink/resolve.
	Data interface{}
}

// WorkspaceSymbol represents information about a symbol across the entire workspace.
// Unlike DocumentSymbol which is for a single document, this can represent symbols
// in any file in the workspace.
type WorkspaceSymbol struct {
	// Name is the name of this symbol.
	Name string

	// Kind is the kind of this symbol.
	Kind SymbolKind

	// Tags are tags for this symbol.
	Tags []SymbolTag

	// ContainerName is an optional container name (e.g., the class name for a method).
	ContainerName string

	// Location is where this symbol is defined.
	Location Location

	// Data is arbitrary data preserved between workspace/symbol and workspaceSymbol/resolve.
	Data interface{}
}

// WorkspaceSymbolParams are parameters for a workspace symbol request.
type WorkspaceSymbolParams struct {
	// Query is a non-empty query string to search for.
	// The query is used to filter symbols.
	Query string
}

// ReferenceContext specifies additional information when finding references.
type ReferenceContext struct {
	// IncludeDeclaration indicates whether the declaration should be included in the results.
	IncludeDeclaration bool
}

// InlayHintKind defines the kind of an inlay hint.
type InlayHintKind int

const (
	// InlayHintKindType is a hint for a type annotation.
	InlayHintKindType InlayHintKind = 1
	// InlayHintKindParameter is a hint for a parameter name.
	InlayHintKindParameter InlayHintKind = 2
)

// InlayHint represents an inline annotation that is shown in the editor.
// Inlay hints appear inline with the code and provide additional information
// such as parameter names or inferred types.
type InlayHint struct {
	// Position is the position where the hint should be shown.
	// The hint appears directly before this position.
	Position Position

	// Label is the text to display in the hint.
	// Can be a simple string or structured with parts.
	Label string

	// Kind is the kind of hint (type or parameter).
	Kind *InlayHintKind

	// TextEdits are optional text edits to apply when accepting the hint.
	// This allows hints to be actionable.
	TextEdits []TextEdit

	// Tooltip provides additional information shown when hovering the hint.
	Tooltip string

	// PaddingLeft adds whitespace before the hint.
	PaddingLeft bool

	// PaddingRight adds whitespace after the hint.
	PaddingRight bool

	// Data is arbitrary data preserved between textDocument/inlayHint and inlayHint/resolve.
	Data interface{}
}

// SelectionRange represents a range that can be selected in a text document.
// Selection ranges allow for smart expand/shrink selection functionality.
// Uses UTF-8 byte offsets for the range.
type SelectionRange struct {
	// Range is the range of this selection range.
	Range Range

	// Parent is the parent selection range containing this range.
	// This forms a hierarchy of selection ranges from narrow to broad.
	Parent *SelectionRange
}

// Color represents an RGBA color value.
type Color struct {
	// Red component in the range [0-1].
	Red float64

	// Green component in the range [0-1].
	Green float64

	// Blue component in the range [0-1].
	Blue float64

	// Alpha component in the range [0-1].
	Alpha float64
}

// ColorInformation represents a color reference found in a document.
// Uses UTF-8 byte offsets for the range.
type ColorInformation struct {
	// Range is the range in the document where this color appears.
	Range Range

	// Color is the actual RGBA color value.
	Color Color
}

// ColorPresentation represents how a color can be represented as text.
// For example, a color could be presented as "#FF0000", "rgb(255, 0, 0)",
// or "red" depending on the context.
type ColorPresentation struct {
	// Label is the label of this color presentation.
	// This is what will be shown in the color picker.
	Label string

	// TextEdit is an optional edit to apply when selecting this presentation.
	// If not provided, the label is used as the edit text.
	TextEdit *TextEdit

	// AdditionalTextEdits are optional additional edits that are applied
	// when selecting this color presentation.
	AdditionalTextEdits []TextEdit
}
