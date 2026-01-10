package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
	"github.com/SCKelemen/unicode/uax29"
)

// SimpleReferencesProvider finds all references to a symbol using text matching.
// This is a basic example that works for simple identifiers across files.
type SimpleReferencesProvider struct {
	// FileProvider is a function to get content for a URI.
	// In a real implementation, this would come from a document store.
	FileProvider func(uri string) (string, error)
}

func (p *SimpleReferencesProvider) FindReferences(uri, content string, position core.Position, context core.ReferenceContext) []core.Location {
	// Get the word at the position
	word := getWordAtPositionForReferences(content, position)
	if word == "" {
		return nil
	}

	var locations []core.Location

	// Find all occurrences in the current document
	locations = append(locations, p.findOccurrencesInDocument(uri, content, word, context)...)

	return locations
}

func (p *SimpleReferencesProvider) findOccurrencesInDocument(uri, content, word string, context core.ReferenceContext) []core.Location {
	var locations []core.Location

	// Get all word boundaries using UAX29
	breaks := uax29.FindWordBreaks(content)
	if len(breaks) < 2 {
		return nil
	}

	// Find all occurrences
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		candidateWord := content[start:end]
		if candidateWord == word {
			locations = append(locations, core.Location{
				URI: uri,
				Range: core.Range{
					Start: core.ByteOffsetToPosition(content, start),
					End:   core.ByteOffsetToPosition(content, end),
				},
			})
		}
	}

	return locations
}

// GoReferencesProvider finds references to Go identifiers using AST parsing.
// This provides more accurate results than text-based matching.
type GoReferencesProvider struct {
	// FileProvider is a function to get content for a URI.
	FileProvider func(uri string) (string, error)
}

func (p *GoReferencesProvider) FindReferences(uri, content string, position core.Position, context core.ReferenceContext) []core.Location {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Find the identifier at the position
	offset := core.PositionToByteOffset(content, position)
	if offset < 0 {
		return nil
	}

	var targetIdent *ast.Ident
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			identStart := fset.Position(ident.Pos()).Offset
			identEnd := fset.Position(ident.End()).Offset

			if offset >= identStart && offset < identEnd {
				targetIdent = ident
				return false
			}
		}
		return true
	})

	if targetIdent == nil {
		return nil
	}

	// Find all references to this identifier
	return p.findAllReferences(f, fset, uri, targetIdent.Name, context)
}

func (p *GoReferencesProvider) findAllReferences(f *ast.File, fset *token.FileSet, uri, name string, context core.ReferenceContext) []core.Location {
	var locations []core.Location

	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == name {
			identStart := fset.Position(ident.Pos())
			identEnd := fset.Position(ident.End())

			locations = append(locations, core.Location{
				URI: uri,
				Range: core.Range{
					Start: core.Position{Line: identStart.Line - 1, Character: identStart.Column - 1},
					End:   core.Position{Line: identEnd.Line - 1, Character: identEnd.Column - 1},
				},
			})
		}
		return true
	})

	return locations
}

// MultiFileReferencesProvider finds references across multiple files.
type MultiFileReferencesProvider struct {
	// Files maps URIs to their content.
	Files map[string]string
}

func (p *MultiFileReferencesProvider) FindReferences(uri, content string, position core.Position, context core.ReferenceContext) []core.Location {
	// Get the word at the position
	word := getWordAtPositionForReferences(content, position)
	if word == "" {
		return nil
	}

	var locations []core.Location

	// Search all files for occurrences
	for fileURI, fileContent := range p.Files {
		breaks := uax29.FindWordBreaks(fileContent)
		if len(breaks) < 2 {
			continue
		}

		for i := 0; i < len(breaks)-1; i++ {
			start := breaks[i]
			end := breaks[i+1]

			candidateWord := fileContent[start:end]
			if candidateWord == word {
				locations = append(locations, core.Location{
					URI: fileURI,
					Range: core.Range{
						Start: core.ByteOffsetToPosition(fileContent, start),
						End:   core.ByteOffsetToPosition(fileContent, end),
					},
				})
			}
		}
	}

	return locations
}

// Helper function to get word at position for references
func getWordAtPositionForReferences(content string, pos core.Position) string {
	offset := core.PositionToByteOffset(content, pos)
	if offset < 0 || offset >= len(content) {
		return ""
	}

	// Get all word boundaries
	breaks := uax29.FindWordBreaks(content)
	if len(breaks) < 2 {
		return ""
	}

	// Helper function to check if a segment is a valid word
	isValidWord := func(word string) bool {
		if len(strings.TrimSpace(word)) == 0 {
			return false
		}

		for _, r := range word {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r > 127 {
				return true
			}
		}
		return false
	}

	// Find which word boundary the offset falls into
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		if offset >= start && offset < end {
			word := content[start:end]
			if isValidWord(word) {
				return word
			}
			// If cursor is on whitespace/punctuation, try the previous word
			if i > 0 {
				prevStart := breaks[i-1]
				prevEnd := breaks[i]
				prevWord := content[prevStart:prevEnd]
				if isValidWord(prevWord) {
					return prevWord
				}
			}
			return ""
		}
	}

	return ""
}

// Example usage in CLI tool
func CLIReferencesExample() {
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

	provider := &SimpleReferencesProvider{}

	// Find all references to "sum" at line 3, character 1
	position := core.Position{Line: 3, Character: 1}
	context := core.ReferenceContext{
		IncludeDeclaration: true,
	}

	references := provider.FindReferences("file:///main.go", content, position, context)

	println("Found", len(references), "references:")
	for _, ref := range references {
		println("  -", ref.String())
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentReferences(
// 	ctx *glsp.Context,
// 	params *protocol.ReferenceParams,
// ) ([]protocol.Location, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol position to core position
// 	corePos := adapter_3_16.ProtocolToCorePosition(params.Position, content)
//
// 	// Convert protocol context to core context
// 	coreContext := core.ReferenceContext{
// 		IncludeDeclaration: params.Context.IncludeDeclaration,
// 	}
//
// 	// Use provider with core types
// 	coreLocations := s.referencesProvider.FindReferences(uri, content, corePos, coreContext)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolLocations(coreLocations, content), nil
// }
