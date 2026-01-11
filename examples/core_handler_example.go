package examples

import (
	"github.com/SCKelemen/lsp"
	"github.com/SCKelemen/lsp/adapter"
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

// This example shows how to use core types in your business logic
// while working with protocol handlers.

// ExampleLanguageServer demonstrates using core types with a document manager.
type ExampleLanguageServer struct {
	documents *core.DocumentManager
}

func NewExampleLanguageServer() *ExampleLanguageServer {
	return &ExampleLanguageServer{
		documents: core.NewDocumentManager(),
	}
}

// TextDocumentDidOpen handles document open notifications.
// This converts protocol types to core types at the boundary.
func (s *ExampleLanguageServer) TextDocumentDidOpen(context *lsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	// Extract protocol types
	uri := string(params.TextDocument.URI)
	content := params.TextDocument.Text
	version := int(params.TextDocument.Version)

	// Store in document manager (using core types internally)
	s.documents.Open(uri, content, version)

	// Optionally validate and send diagnostics
	diagnostics := s.validateDocument(uri, content)
	if len(diagnostics) > 0 {
		// Convert core diagnostics to protocol diagnostics
		_ = adapter_3_16.CoreToProtocolDiagnostics(diagnostics, content)

		// Send diagnostics notification (would need to be implemented in actual server)
		// context.Notify(...)
	}

	return nil
}

// validateDocument is your business logic that works with core types.
// It uses UTF-8 byte offsets naturally, without worrying about UTF-16.
func (s *ExampleLanguageServer) validateDocument(uri, content string) []core.Diagnostic {
	// Your validation logic here using core types with UTF-8 offsets
	// Example: find all occurrences of "TODO" and create diagnostics

	var diagnostics []core.Diagnostic
	severity := core.SeverityInformation

	// Simple example: scan for "TODO" comments
	offset := 0
	for i := 0; i < len(content); i++ {
		if i+4 <= len(content) && content[i:i+4] == "TODO" {
			// Create diagnostic using core types (UTF-8 offsets)
			pos := core.ByteOffsetToPosition(content, i)
			endPos := core.ByteOffsetToPosition(content, i+4)

			diag := core.Diagnostic{
				Range: core.Range{
					Start: pos,
					End:   endPos,
				},
				Severity: &severity,
				Message:  "TODO comment found",
				Source:   "example-linter",
			}
			diagnostics = append(diagnostics, diag)
		}
		offset++
	}

	return diagnostics
}

// TextDocumentDefinition handles go-to-definition requests.
// Shows how to work with positions using core types.
func (s *ExampleLanguageServer) TextDocumentDefinition(context *lsp.Context, params *protocol.DefinitionParams) (any, error) {
	uri := string(params.TextDocument.URI)
	content := s.documents.GetContent(uri)

	// Convert protocol position (UTF-16) to core position (UTF-8)
	corePos := adapter_3_16.ProtocolToCorePosition(params.Position, content)

	// Your business logic using core types
	defLocation := s.findDefinition(uri, content, corePos)
	if defLocation == nil {
		return nil, nil
	}

	// Convert core location back to protocol location
	protocolLocation := adapter_3_16.CoreToProtocolLocation(*defLocation, content)
	return protocolLocation, nil
}

// findDefinition is your business logic that works with core types.
func (s *ExampleLanguageServer) findDefinition(uri, content string, pos core.Position) *core.Location {
	// Your definition-finding logic here using UTF-8 offsets
	// Example: just return the same position for demonstration

	return &core.Location{
		URI: uri,
		Range: core.Range{
			Start: pos,
			End:   pos,
		},
	}
}

// TextDocumentDidChange handles document change notifications.
// Shows how to apply edits using core types.
func (s *ExampleLanguageServer) TextDocumentDidChange(context *lsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	_ = string(params.TextDocument.URI)

	// For full document updates, just replace the entire content
	// In a real implementation, you would handle incremental updates
	// by parsing the TextDocumentContentChangeEvent types from ContentChanges

	// Simplified: assume full document replacement
	// (In production, you'd need to handle both full and incremental changes)

	return nil
}

// Example: Using core types in a CLI tool (no LSP server needed)
func CLIToolExample() {
	// CLI tools can work directly with core types
	content := "hello world TODO: fix this"

	// Find all TODOs using UTF-8 offsets
	var diagnostics []core.Diagnostic
	severity := core.SeverityInformation

	for i := 0; i < len(content); i++ {
		if i+4 <= len(content) && content[i:i+4] == "TODO" {
			pos := core.ByteOffsetToPosition(content, i)
			endPos := core.ByteOffsetToPosition(content, i+4)

			diag := core.Diagnostic{
				Range: core.Range{
					Start: pos,
					End:   endPos,
				},
				Severity: &severity,
				Message:  "TODO comment found",
				Source:   "cli-linter",
			}
			diagnostics = append(diagnostics, diag)
		}
	}

	// Use diagnostics in your CLI tool
	// No need to convert to protocol types unless interfacing with an LSP client
	for _, diag := range diagnostics {
		println(diag.Message, "at", diag.Range.String())
	}
}
