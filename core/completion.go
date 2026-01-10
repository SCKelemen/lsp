package core

// CompletionItemKind defines the kind of a completion item.
type CompletionItemKind int

const (
	CompletionItemKindText          CompletionItemKind = 1
	CompletionItemKindMethod        CompletionItemKind = 2
	CompletionItemKindFunction      CompletionItemKind = 3
	CompletionItemKindConstructor   CompletionItemKind = 4
	CompletionItemKindField         CompletionItemKind = 5
	CompletionItemKindVariable      CompletionItemKind = 6
	CompletionItemKindClass         CompletionItemKind = 7
	CompletionItemKindInterface     CompletionItemKind = 8
	CompletionItemKindModule        CompletionItemKind = 9
	CompletionItemKindProperty      CompletionItemKind = 10
	CompletionItemKindUnit          CompletionItemKind = 11
	CompletionItemKindValue         CompletionItemKind = 12
	CompletionItemKindEnum          CompletionItemKind = 13
	CompletionItemKindKeyword       CompletionItemKind = 14
	CompletionItemKindSnippet       CompletionItemKind = 15
	CompletionItemKindColor         CompletionItemKind = 16
	CompletionItemKindFile          CompletionItemKind = 17
	CompletionItemKindReference     CompletionItemKind = 18
	CompletionItemKindFolder        CompletionItemKind = 19
	CompletionItemKindEnumMember    CompletionItemKind = 20
	CompletionItemKindConstant      CompletionItemKind = 21
	CompletionItemKindStruct        CompletionItemKind = 22
	CompletionItemKindEvent         CompletionItemKind = 23
	CompletionItemKindOperator      CompletionItemKind = 24
	CompletionItemKindTypeParameter CompletionItemKind = 25
)

// CompletionItemTag defines tags for completion items.
type CompletionItemTag int

const (
	// CompletionItemTagDeprecated indicates the item is deprecated.
	CompletionItemTagDeprecated CompletionItemTag = 1
)

// InsertTextFormat defines how the insert text should be interpreted.
type InsertTextFormat int

const (
	// InsertTextFormatPlainText means the insert text is plain text.
	InsertTextFormatPlainText InsertTextFormat = 1
	// InsertTextFormatSnippet means the insert text is a snippet.
	InsertTextFormatSnippet InsertTextFormat = 2
)

// CompletionTriggerKind defines how a completion was triggered.
type CompletionTriggerKind int

const (
	// CompletionTriggerKindInvoked means completion was explicitly requested.
	CompletionTriggerKindInvoked CompletionTriggerKind = 1
	// CompletionTriggerKindTriggerCharacter means completion was triggered by a character.
	CompletionTriggerKindTriggerCharacter CompletionTriggerKind = 2
	// CompletionTriggerKindTriggerForIncompleteCompletions means completion was re-triggered.
	CompletionTriggerKindTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	// Label is the text shown in the completion list.
	Label string

	// Kind is the type of completion item.
	Kind *CompletionItemKind

	// Tags are additional tags for this item.
	Tags []CompletionItemTag

	// Detail provides additional information about this item.
	Detail string

	// Documentation provides documentation for this item.
	Documentation string

	// Deprecated indicates if this item is deprecated.
	Deprecated bool

	// Preselect indicates if this item should be preselected.
	Preselect bool

	// SortText is used for sorting (if different from label).
	SortText string

	// FilterText is used for filtering (if different from label).
	FilterText string

	// InsertText is the text to insert (if different from label).
	InsertText string

	// InsertTextFormat indicates how to interpret the insert text.
	InsertTextFormat *InsertTextFormat

	// TextEdit is the edit to apply when selecting this item.
	TextEdit *TextEdit

	// AdditionalTextEdits are additional edits to apply.
	AdditionalTextEdits []TextEdit

	// CommitCharacters are characters that trigger accepting this completion.
	CommitCharacters []string

	// Command is a command to execute after inserting this completion.
	Command *Command

	// Data is arbitrary data preserved for completion resolve.
	Data interface{}
}

// CompletionList represents a list of completion items.
type CompletionList struct {
	// IsIncomplete indicates if the list is incomplete.
	// If true, the client should re-request completions when typing continues.
	IsIncomplete bool

	// Items are the completion items.
	Items []CompletionItem
}

// CompletionContext provides context for a completion request.
type CompletionContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string

	// Position is where completion was requested (UTF-8 offset).
	Position Position

	// TriggerKind indicates how completion was triggered.
	TriggerKind CompletionTriggerKind

	// TriggerCharacter is the character that triggered completion (if any).
	TriggerCharacter string
}

// CompletionProvider provides completion suggestions.
type CompletionProvider interface {
	// ProvideCompletions returns completion items for the given context.
	// Returns nil or empty list if no completions are available.
	ProvideCompletions(ctx CompletionContext) *CompletionList
}

// CompletionItemResolveProvider resolves additional details for a completion item.
type CompletionItemResolveProvider interface {
	// ResolveCompletionItem resolves additional details for a completion item.
	// This is called when the user selects an item but before inserting it.
	ResolveCompletionItem(item CompletionItem) CompletionItem
}

// SignatureInformation represents a callable signature.
type SignatureInformation struct {
	// Label is the signature label.
	Label string

	// Documentation provides documentation for this signature.
	Documentation string

	// Parameters are the parameters of this signature.
	Parameters []ParameterInformation

	// ActiveParameter is the index of the active parameter.
	ActiveParameter *int
}

// ParameterInformation represents a parameter of a callable signature.
type ParameterInformation struct {
	// Label is the parameter label (can be a substring of the signature label).
	Label string

	// Documentation provides documentation for this parameter.
	Documentation string
}

// SignatureHelp represents signature help information.
type SignatureHelp struct {
	// Signatures are the available signatures.
	Signatures []SignatureInformation

	// ActiveSignature is the index of the active signature.
	ActiveSignature *int

	// ActiveParameter is the index of the active parameter in the active signature.
	ActiveParameter *int
}

// SignatureHelpContext provides context for signature help.
type SignatureHelpContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string

	// Position is where signature help was requested (UTF-8 offset).
	Position Position

	// TriggerCharacter is the character that triggered signature help (if any).
	TriggerCharacter string

	// IsRetrigger indicates if this is a retrigger.
	IsRetrigger bool
}

// SignatureHelpProvider provides signature help.
type SignatureHelpProvider interface {
	// ProvideSignatureHelp returns signature help for the given context.
	// Returns nil if no signature help is available.
	ProvideSignatureHelp(ctx SignatureHelpContext) *SignatureHelp
}

// RenameContext provides context for a rename operation.
type RenameContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string

	// Position is where rename was requested (UTF-8 offset).
	Position Position

	// NewName is the new name for the symbol.
	NewName string
}

// RenameProvider provides rename operations.
type RenameProvider interface {
	// ProvideRename returns a workspace edit to rename a symbol.
	// Returns nil if rename is not possible.
	ProvideRename(ctx RenameContext) *WorkspaceEdit
}

// PrepareRenameProvider checks if rename is possible at a position.
type PrepareRenameProvider interface {
	// PrepareRename checks if rename is valid at the position.
	// Returns the range of the symbol to rename, or nil if not possible.
	PrepareRename(uri, content string, position Position) *Range
}

// CodeLens represents a command that should be shown inline with source code.
type CodeLens struct {
	// Range is where the code lens should appear.
	Range Range

	// Command is the command to execute when the code lens is clicked.
	Command *Command

	// Data is arbitrary data preserved for code lens resolve.
	Data interface{}
}

// CodeLensContext provides context for code lens requests.
type CodeLensContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string
}

// CodeLensProvider provides code lenses.
type CodeLensProvider interface {
	// ProvideCodeLenses returns code lenses for the document.
	ProvideCodeLenses(ctx CodeLensContext) []CodeLens
}

// CodeLensResolveProvider resolves additional details for a code lens.
type CodeLensResolveProvider interface {
	// ResolveCodeLens resolves additional details for a code lens.
	ResolveCodeLens(lens CodeLens) CodeLens
}
