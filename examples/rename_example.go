package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/SCKelemen/lsp/core"
	"github.com/SCKelemen/unicode/uax29"
)

// SimpleRenameProvider provides basic rename functionality for simple identifiers.
// This is useful for simple text-based renaming in configuration files or simple languages.
type SimpleRenameProvider struct{}

func (p *SimpleRenameProvider) PrepareRename(uri, content string, position core.Position) *core.Range {
	// Find the word at the position using Unicode word boundaries
	offset := core.PositionToByteOffset(content, position)
	if offset < 0 || offset >= len(content) {
		return nil
	}

	// Get all word boundaries in the content
	breaks := uax29.FindWordBreaks(content)
	if len(breaks) < 2 {
		return nil
	}

	// Find which word boundary the offset falls into
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		if offset >= start && offset < end {
			// Check if this segment is actually a word (not whitespace or punctuation)
			word := content[start:end]
			if len(strings.TrimSpace(word)) == 0 {
				return nil // Just whitespace
			}

			// Check if it contains at least one alphanumeric character
			hasAlphaNum := false
			for _, r := range word {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r > 127 {
					hasAlphaNum = true
					break
				}
			}
			if !hasAlphaNum {
				return nil // Just punctuation
			}

			return &core.Range{
				Start: core.ByteOffsetToPosition(content, start),
				End:   core.ByteOffsetToPosition(content, end),
			}
		}
	}

	return nil
}

func (p *SimpleRenameProvider) ProvideRename(ctx core.RenameContext) *core.WorkspaceEdit {
	// Prepare rename to get the range
	renameRange := p.PrepareRename(ctx.URI, ctx.Content, ctx.Position)
	if renameRange == nil {
		return nil
	}

	// Get the current word
	startOffset := core.PositionToByteOffset(ctx.Content, renameRange.Start)
	endOffset := core.PositionToByteOffset(ctx.Content, renameRange.End)

	if startOffset < 0 || endOffset > len(ctx.Content) {
		return nil
	}

	oldName := ctx.Content[startOffset:endOffset]

	// Find all occurrences of the word (simple case-sensitive match)
	var edits []core.TextEdit

	// Use regex to find whole word matches
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
	matches := pattern.FindAllStringIndex(ctx.Content, -1)

	for _, match := range matches {
		start := match[0]
		end := match[1]

		startPos := core.ByteOffsetToPosition(ctx.Content, start)
		endPos := core.ByteOffsetToPosition(ctx.Content, end)

		edits = append(edits, core.TextEdit{
			Range: core.Range{
				Start: startPos,
				End:   endPos,
			},
			NewText: ctx.NewName,
		})
	}

	if len(edits) == 0 {
		return nil
	}

	return &core.WorkspaceEdit{
		Changes: map[string][]core.TextEdit{
			ctx.URI: edits,
		},
	}
}

// GoRenameProvider provides rename functionality for Go identifiers.
// This uses AST parsing to ensure we only rename actual identifiers.
type GoRenameProvider struct {
	// FileProvider is a function to get content for a URI
	// In a real implementation, this would come from a document store
	FileProvider func(uri string) (string, error)
}

func (p *GoRenameProvider) PrepareRename(uri, content string, position core.Position) *core.Range {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	offset := core.PositionToByteOffset(content, position)
	if offset < 0 {
		return nil
	}

	// Find the identifier at the position
	var foundRange *core.Range
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			identStart := fset.Position(ident.Pos())
			identEnd := fset.Position(ident.End())

			startOffset := core.PositionToByteOffset(content, core.Position{
				Line:      identStart.Line - 1,
				Character: identStart.Column - 1,
			})
			endOffset := core.PositionToByteOffset(content, core.Position{
				Line:      identEnd.Line - 1,
				Character: identEnd.Column - 1,
			})

			if offset >= startOffset && offset < endOffset {
				// Don't allow renaming language keywords or built-ins
				if isGoKeyword(ident.Name) || ident.Name == "_" {
					return false
				}

				foundRange = &core.Range{
					Start: core.Position{Line: identStart.Line - 1, Character: identStart.Column - 1},
					End:   core.Position{Line: identEnd.Line - 1, Character: identEnd.Column - 1},
				}
				return false
			}
		}
		return true
	})

	return foundRange
}

func (p *GoRenameProvider) ProvideRename(ctx core.RenameContext) *core.WorkspaceEdit {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	// Prepare rename to get the range and validate
	renameRange := p.PrepareRename(ctx.URI, ctx.Content, ctx.Position)
	if renameRange == nil {
		return nil
	}

	// Get the old name
	startOffset := core.PositionToByteOffset(ctx.Content, renameRange.Start)
	endOffset := core.PositionToByteOffset(ctx.Content, renameRange.End)

	if startOffset < 0 || endOffset > len(ctx.Content) {
		return nil
	}

	oldName := ctx.Content[startOffset:endOffset]

	// Parse the file to find all occurrences
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		return nil
	}

	var edits []core.TextEdit

	// Find all identifiers with the same name
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == oldName {
			identStart := fset.Position(ident.Pos())
			identEnd := fset.Position(ident.End())

			edits = append(edits, core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: identStart.Line - 1, Character: identStart.Column - 1},
					End:   core.Position{Line: identEnd.Line - 1, Character: identEnd.Column - 1},
				},
				NewText: ctx.NewName,
			})
		}
		return true
	})

	if len(edits) == 0 {
		return nil
	}

	return &core.WorkspaceEdit{
		Changes: map[string][]core.TextEdit{
			ctx.URI: edits,
		},
	}
}

// MultiFileRenameProvider demonstrates renaming across multiple files.
// This is a simplified example showing the concept.
type MultiFileRenameProvider struct {
	// Files maps URIs to their content
	Files map[string]string
}

func (p *MultiFileRenameProvider) PrepareRename(uri, content string, position core.Position) *core.Range {
	// Use Unicode word boundary detection
	offset := core.PositionToByteOffset(content, position)
	if offset < 0 || offset >= len(content) {
		return nil
	}

	// Get all word boundaries in the content
	breaks := uax29.FindWordBreaks(content)
	if len(breaks) < 2 {
		return nil
	}

	// Find which word boundary the offset falls into
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		if offset >= start && offset < end {
			// Check if this segment is actually a word
			word := content[start:end]
			if len(strings.TrimSpace(word)) == 0 {
				return nil
			}

			// Check if it contains at least one alphanumeric character
			hasAlphaNum := false
			for _, r := range word {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r > 127 {
					hasAlphaNum = true
					break
				}
			}
			if !hasAlphaNum {
				return nil
			}

			return &core.Range{
				Start: core.ByteOffsetToPosition(content, start),
				End:   core.ByteOffsetToPosition(content, end),
			}
		}
	}

	return nil
}

func (p *MultiFileRenameProvider) ProvideRename(ctx core.RenameContext) *core.WorkspaceEdit {
	renameRange := p.PrepareRename(ctx.URI, ctx.Content, ctx.Position)
	if renameRange == nil {
		return nil
	}

	startOffset := core.PositionToByteOffset(ctx.Content, renameRange.Start)
	endOffset := core.PositionToByteOffset(ctx.Content, renameRange.End)

	if startOffset < 0 || endOffset > len(ctx.Content) {
		return nil
	}

	oldName := ctx.Content[startOffset:endOffset]

	// Create a workspace edit with changes across multiple files
	changes := make(map[string][]core.TextEdit)

	// Search all files for occurrences
	for uri, content := range p.Files {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(oldName) + `\b`)
		matches := pattern.FindAllStringIndex(content, -1)

		if len(matches) > 0 {
			var fileEdits []core.TextEdit
			for _, match := range matches {
				start := match[0]
				end := match[1]

				fileEdits = append(fileEdits, core.TextEdit{
					Range: core.Range{
						Start: core.ByteOffsetToPosition(content, start),
						End:   core.ByteOffsetToPosition(content, end),
					},
					NewText: ctx.NewName,
				})
			}
			changes[uri] = fileEdits
		}
	}

	if len(changes) == 0 {
		return nil
	}

	return &core.WorkspaceEdit{
		Changes: changes,
	}
}

// isGoKeyword checks if a string is a Go keyword
func isGoKeyword(s string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return keywords[s]
}

// Example usage in CLI tool
func CLIRenameExample() {
	content := `package main

func calculateSum(a, b int) int {
	sum := a + b
	return sum
}

func main() {
	result := calculateSum(5, 10)
	println(result)
}
`

	provider := &SimpleRenameProvider{}

	// Try to rename "sum" at position (line 3, char 1)
	position := core.Position{Line: 3, Character: 1}

	// First, check if rename is possible
	prepareResult := provider.PrepareRename("file:///main.go", content, position)
	if prepareResult == nil {
		println("Cannot rename at this position")
		return
	}

	println("Can rename symbol at range:", prepareResult.String())

	// Perform the rename
	ctx := core.RenameContext{
		URI:      "file:///main.go",
		Content:  content,
		Position: position,
		NewName:  "total",
	}

	edit := provider.ProvideRename(ctx)
	if edit == nil {
		println("Rename failed")
		return
	}

	println("Rename successful!")
	println("Changes:")
	for uri, edits := range edit.Changes {
		println("  File:", uri)
		for i, e := range edits {
			println("    Edit", i+1, ":", e.Range.String(), "->", e.NewText)
		}
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentPrepareRename(
// 	ctx *lsp.Context,
// 	params *protocol.PrepareRenameParams,
// ) (*protocol.Range, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol position to core position
// 	corePos := adapter_3_16.ProtocolToCore Position(params.Position, content)
//
// 	// Use provider with core types
// 	coreRange := s.renameProvider.PrepareRename(uri, content, corePos)
// 	if coreRange == nil {
// 		return nil, nil
// 	}
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolRange(*coreRange, content), nil
// }
//
// func (s *Server) TextDocumentRename(
// 	ctx *lsp.Context,
// 	params *protocol.RenameParams,
// ) (*protocol.WorkspaceEdit, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol position to core position
// 	corePos := adapter_3_16.ProtocolToCore Position(params.Position, content)
//
// 	// Use provider with core types
// 	coreCtx := core.RenameContext{
// 		URI:      uri,
// 		Content:  content,
// 		Position: corePos,
// 		NewName:  params.NewName,
// 	}
//
// 	coreEdit := s.renameProvider.ProvideRename(coreCtx)
// 	if coreEdit == nil {
// 		return nil, nil
// 	}
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolWorkspaceEdit(*coreEdit), nil
// }
