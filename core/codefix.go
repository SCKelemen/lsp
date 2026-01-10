package core

// CodeFixContext provides context for code fix providers.
type CodeFixContext struct {
	// URI is the document URI.
	URI string

	// Content is the document content.
	Content string

	// Range is the range for which code fixes are requested.
	Range Range

	// Diagnostics are the diagnostics in the requested range.
	Diagnostics []Diagnostic

	// Only contains the kinds of code actions requested.
	// If empty, all kinds are requested.
	Only []CodeActionKind

	// TriggerKind indicates how the code action was triggered.
	TriggerKind CodeActionTriggerKind
}

// CodeActionTriggerKind defines how a code action was triggered.
type CodeActionTriggerKind int

const (
	// CodeActionTriggerKindInvoked means the code action was explicitly requested by the user.
	CodeActionTriggerKindInvoked CodeActionTriggerKind = 1

	// CodeActionTriggerKindAutomatic means the code action was automatically triggered (e.g., on save).
	CodeActionTriggerKindAutomatic CodeActionTriggerKind = 2
)

// CodeFixProvider provides code fixes for a document.
// Implementations work directly with core types (UTF-8 offsets).
type CodeFixProvider interface {
	// ProvideCodeFixes returns code actions for the given context.
	// Returns nil or empty slice if no fixes are available.
	ProvideCodeFixes(ctx CodeFixContext) []CodeAction
}

// CodeFixRegistry manages multiple code fix providers.
type CodeFixRegistry struct {
	providers []CodeFixProvider
}

// NewCodeFixRegistry creates a new code fix registry.
func NewCodeFixRegistry() *CodeFixRegistry {
	return &CodeFixRegistry{
		providers: make([]CodeFixProvider, 0),
	}
}

// Register adds a code fix provider to the registry.
func (r *CodeFixRegistry) Register(provider CodeFixProvider) {
	r.providers = append(r.providers, provider)
}

// ProvideCodeFixes collects code fixes from all registered providers.
func (r *CodeFixRegistry) ProvideCodeFixes(ctx CodeFixContext) []CodeAction {
	var actions []CodeAction
	for _, provider := range r.providers {
		if fixes := provider.ProvideCodeFixes(ctx); len(fixes) > 0 {
			actions = append(actions, fixes...)
		}
	}
	return actions
}

// DiagnosticProvider provides diagnostics for a document.
// Implementations work directly with core types (UTF-8 offsets).
type DiagnosticProvider interface {
	// ProvideDiagnostics returns diagnostics for the given document.
	// Returns nil or empty slice if no diagnostics are found.
	ProvideDiagnostics(uri, content string) []Diagnostic
}

// DiagnosticRegistry manages multiple diagnostic providers.
type DiagnosticRegistry struct {
	providers []DiagnosticProvider
}

// NewDiagnosticRegistry creates a new diagnostic registry.
func NewDiagnosticRegistry() *DiagnosticRegistry {
	return &DiagnosticRegistry{
		providers: make([]DiagnosticProvider, 0),
	}
}

// Register adds a diagnostic provider to the registry.
func (r *DiagnosticRegistry) Register(provider DiagnosticProvider) {
	r.providers = append(r.providers, provider)
}

// ProvideDiagnostics collects diagnostics from all registered providers.
func (r *DiagnosticRegistry) ProvideDiagnostics(uri, content string) []Diagnostic {
	var diagnostics []Diagnostic
	for _, provider := range r.providers {
		if diags := provider.ProvideDiagnostics(uri, content); len(diags) > 0 {
			diagnostics = append(diagnostics, diags...)
		}
	}
	return diagnostics
}

// FoldingRangeProvider provides folding ranges for a document.
type FoldingRangeProvider interface {
	// ProvideFoldingRanges returns folding ranges for the given document.
	ProvideFoldingRanges(uri, content string) []FoldingRange
}

// DocumentSymbolProvider provides document symbols.
type DocumentSymbolProvider interface {
	// ProvideDocumentSymbols returns document symbols for the given document.
	ProvideDocumentSymbols(uri, content string) []DocumentSymbol
}

// DefinitionProvider provides go-to-definition locations.
type DefinitionProvider interface {
	// ProvideDefinition returns definition locations for the position.
	// Returns nil if no definition is found.
	ProvideDefinition(uri, content string, position Position) []Location
}

// HoverInfo contains hover information for a position.
type HoverInfo struct {
	// Contents is the hover content (markdown or plain text).
	Contents string

	// Range is the range to which the hover applies.
	// If nil, the range is computed from the position.
	Range *Range
}

// HoverProvider provides hover information.
type HoverProvider interface {
	// ProvideHover returns hover information for the position.
	// Returns nil if no hover information is available.
	ProvideHover(uri, content string, position Position) *HoverInfo
}

// FormattingOptions contains options for formatting.
type FormattingOptions struct {
	// TabSize is the size of a tab in spaces.
	TabSize int

	// InsertSpaces indicates whether to insert spaces instead of tabs.
	InsertSpaces bool

	// TrimTrailingWhitespace indicates whether to trim trailing whitespace.
	TrimTrailingWhitespace bool

	// InsertFinalNewline indicates whether to insert a final newline.
	InsertFinalNewline bool

	// TrimFinalNewlines indicates whether to trim final newlines.
	TrimFinalNewlines bool
}

// FormattingProvider provides document formatting.
type FormattingProvider interface {
	// ProvideFormatting returns edits to format the entire document.
	ProvideFormatting(uri, content string, options FormattingOptions) []TextEdit
}

// RangeFormattingProvider provides range formatting.
type RangeFormattingProvider interface {
	// ProvideRangeFormatting returns edits to format a range in the document.
	ProvideRangeFormatting(uri, content string, r Range, options FormattingOptions) []TextEdit
}
