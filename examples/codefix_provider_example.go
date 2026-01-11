package examples

import (
	"fmt"
	"strings"

	"github.com/SCKelemen/lsp"
	"github.com/SCKelemen/lsp/adapter"
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

// Example: Codefix provider that works with core types
// Can be used in both CLI tools and LSP servers without modification

// TabToSpacesProvider is a codefix provider that converts tabs to spaces.
type TabToSpacesProvider struct{}

func (p *TabToSpacesProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	var actions []core.CodeAction

	// Find all tabs in the range
	content := ctx.Content
	for i := 0; i < len(content); i++ {
		if content[i] == '\t' {
			pos := core.ByteOffsetToPosition(content, i)

			// Check if this tab is in the requested range
			if !ctx.Range.Contains(pos) {
				continue
			}

			// Create a code action to replace the tab
			edit := core.TextEdit{
				Range: core.Range{
					Start: pos,
					End:   core.Position{Line: pos.Line, Character: pos.Character + 1},
				},
				NewText: "    ", // 4 spaces
			}

			action := core.CodeAction{
				Title: "Replace tab with spaces",
				Kind:  ptrCodeActionKind(core.CodeActionKindQuickFix),
				Edit: &core.WorkspaceEdit{
					Changes: map[string][]core.TextEdit{
						ctx.URI: {edit},
					},
				},
				IsPreferred: true,
			}

			actions = append(actions, action)
		}
	}

	return actions
}

func ptrCodeActionKind(kind core.CodeActionKind) *core.CodeActionKind {
	return &kind
}

// TODODiagnosticProvider finds TODO comments and creates diagnostics.
type TODODiagnosticProvider struct{}

func (p *TODODiagnosticProvider) ProvideDiagnostics(uri, content string) []core.Diagnostic {
	var diagnostics []core.Diagnostic
	severity := core.SeverityInformation

	// Find all TODO comments
	for i := 0; i < len(content)-4; i++ {
		if strings.HasPrefix(content[i:], "TODO") {
			pos := core.ByteOffsetToPosition(content, i)
			endPos := core.ByteOffsetToPosition(content, i+4)

			diag := core.Diagnostic{
				Range: core.Range{
					Start: pos,
					End:   endPos,
				},
				Severity: &severity,
				Message:  "TODO comment found",
				Source:   "todo-checker",
				Code:     &core.DiagnosticCode{IsInt: false, StringValue: "TODO"},
			}
			diagnostics = append(diagnostics, diag)
		}
	}

	return diagnostics
}

// TODOCodeFixProvider provides fixes for TODO comments.
type TODOCodeFixProvider struct{}

func (p *TODOCodeFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	var actions []core.CodeAction

	// Only provide fixes for TODO diagnostics
	for _, diag := range ctx.Diagnostics {
		if diag.Source != "todo-checker" {
			continue
		}

		// Create action to remove the TODO
		removeAction := core.CodeAction{
			Title:       "Remove TODO comment",
			Kind:        ptrCodeActionKind(core.CodeActionKindQuickFix),
			Diagnostics: []core.Diagnostic{diag},
			Edit: &core.WorkspaceEdit{
				Changes: map[string][]core.TextEdit{
					ctx.URI: {
						{
							Range:   diag.Range,
							NewText: "DONE",
						},
					},
				},
			},
		}
		actions = append(actions, removeAction)

		// Create action to convert to FIXME
		fixmeAction := core.CodeAction{
			Title:       "Convert to FIXME",
			Kind:        ptrCodeActionKind(core.CodeActionKindQuickFix),
			Diagnostics: []core.Diagnostic{diag},
			Edit: &core.WorkspaceEdit{
				Changes: map[string][]core.TextEdit{
					ctx.URI: {
						{
							Range:   diag.Range,
							NewText: "FIXME",
						},
					},
				},
			},
		}
		actions = append(actions, fixmeAction)
	}

	return actions
}

// Example 1: Using providers in a CLI tool
func CLIToolWithProviders() {
	// Create registries
	diagRegistry := core.NewDiagnosticRegistry()
	codeFixRegistry := core.NewCodeFixRegistry()

	// Register providers
	diagRegistry.Register(&TODODiagnosticProvider{})
	codeFixRegistry.Register(&TODOCodeFixProvider{})
	codeFixRegistry.Register(&TabToSpacesProvider{})

	// Read file
	uri := "file:///example.txt"
	content := "hello world\nTODO: fix this\n\tindented line"

	// Get diagnostics
	diagnostics := diagRegistry.ProvideDiagnostics(uri, content)

	fmt.Printf("Found %d diagnostics:\n", len(diagnostics))
	for _, diag := range diagnostics {
		fmt.Printf("  %s: %s\n", diag.Range, diag.Message)
	}

	// Get code fixes for a range
	ctx := core.CodeFixContext{
		URI:         uri,
		Content:     content,
		Range:       core.Range{Start: core.Position{0, 0}, End: core.Position{2, 100}},
		Diagnostics: diagnostics,
	}
	codeFixes := codeFixRegistry.ProvideCodeFixes(ctx)

	fmt.Printf("\nAvailable code fixes: %d\n", len(codeFixes))
	for _, fix := range codeFixes {
		fmt.Printf("  - %s\n", fix.Title)

		// Apply the fix (if it has an edit)
		if fix.Edit != nil && len(fix.Edit.Changes) > 0 {
			for uri, edits := range fix.Edit.Changes {
				fmt.Printf("    Edits for %s:\n", uri)
				for _, edit := range edits {
					fmt.Printf("      %s: '%s' -> '%s'\n",
						edit.Range,
						content[core.PositionToByteOffset(content, edit.Range.Start):core.PositionToByteOffset(content, edit.Range.End)],
						edit.NewText)
				}
			}
		}
	}
}

// Example 2: Using the same providers in an LSP server
type ProviderBasedServer struct {
	documents       *core.DocumentManager
	diagRegistry    *core.DiagnosticRegistry
	codeFixRegistry *core.CodeFixRegistry
}

func NewProviderBasedServer() *ProviderBasedServer {
	diagRegistry := core.NewDiagnosticRegistry()
	codeFixRegistry := core.NewCodeFixRegistry()

	// Register the same providers as the CLI tool
	diagRegistry.Register(&TODODiagnosticProvider{})
	codeFixRegistry.Register(&TODOCodeFixProvider{})
	codeFixRegistry.Register(&TabToSpacesProvider{})

	return &ProviderBasedServer{
		documents:       core.NewDocumentManager(),
		diagRegistry:    diagRegistry,
		codeFixRegistry: codeFixRegistry,
	}
}

// TextDocumentDidOpen handler - publishes diagnostics
func (s *ProviderBasedServer) TextDocumentDidOpen(
	context *glsp.Context,
	params *protocol.DidOpenTextDocumentParams,
) error {
	uri := string(params.TextDocument.URI)
	content := params.TextDocument.Text
	version := int(params.TextDocument.Version)

	// Store document
	s.documents.Open(uri, content, version)

	// Get diagnostics using providers (core types)
	coreDiagnostics := s.diagRegistry.ProvideDiagnostics(uri, content)

	// Convert to protocol types at the boundary
	_ = adapter_3_16.CoreToProtocolDiagnostics(coreDiagnostics, content)

	// Would send to client here
	// context.Notify(...)

	return nil
}

// TextDocumentCodeAction handler - provides code actions
func (s *ProviderBasedServer) TextDocumentCodeAction(
	context *glsp.Context,
	params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
	uri := string(params.TextDocument.URI)
	content := s.documents.GetContent(uri)

	// Convert protocol range to core range
	coreRange := adapter_3_16.ProtocolToCoreRange(params.Range, content)

	// Convert protocol diagnostics to core diagnostics
	coreDiagnostics := adapter_3_16.ProtocolToCoreDiagnostics(params.Context.Diagnostics, content)

	// Create context for providers
	ctx := core.CodeFixContext{
		URI:         uri,
		Content:     content,
		Range:       coreRange,
		Diagnostics: coreDiagnostics,
		Only:        convertCodeActionKinds(params.Context.Only),
	}

	// Get code actions using providers (core types)
	coreActions := s.codeFixRegistry.ProvideCodeFixes(ctx)

	// Convert to protocol types at the boundary
	protocolActions := make([]protocol.CodeAction, len(coreActions))
	for i, action := range coreActions {
		protocolActions[i] = coreToProtocolCodeAction(action, content)
	}

	return protocolActions, nil
}

func convertCodeActionKinds(kinds []protocol.CodeActionKind) []core.CodeActionKind {
	result := make([]core.CodeActionKind, len(kinds))
	for i, k := range kinds {
		result[i] = core.CodeActionKind(k)
	}
	return result
}

func coreToProtocolCodeAction(action core.CodeAction, content string) protocol.CodeAction {
	result := protocol.CodeAction{
		Title: action.Title,
	}

	if action.Kind != nil {
		kind := protocol.CodeActionKind(*action.Kind)
		result.Kind = &kind
	}

	// Convert diagnostics
	if len(action.Diagnostics) > 0 {
		result.Diagnostics = adapter_3_16.CoreToProtocolDiagnostics(action.Diagnostics, content)
	}

	// Convert edit
	if action.Edit != nil {
		result.Edit = coreToProtocolWorkspaceEdit(*action.Edit, content)
	}

	result.IsPreferred = &action.IsPreferred

	return result
}

func coreToProtocolWorkspaceEdit(edit core.WorkspaceEdit, content string) *protocol.WorkspaceEdit {
	result := &protocol.WorkspaceEdit{}

	if len(edit.Changes) > 0 {
		result.Changes = make(map[protocol.DocumentUri][]protocol.TextEdit)
		for uri, edits := range edit.Changes {
			result.Changes[protocol.DocumentUri(uri)] = adapter_3_16.CoreToProtocolTextEdits(edits, content)
		}
	}

	return result
}

// Key benefits of this approach:
// 1. Write provider logic once, use in CLI and LSP server
// 2. Work with natural UTF-8 offsets in provider implementations
// 3. Protocol conversion only at boundaries
// 4. Easy to test providers without LSP server
// 5. Can compose multiple providers together
