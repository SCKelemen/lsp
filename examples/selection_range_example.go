package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoSelectionRangeProvider provides selection ranges for Go source files.
// Selection ranges enable smart expand/shrink selection in editors.
type GoSelectionRangeProvider struct{}

func (p *GoSelectionRangeProvider) ProvideSelectionRanges(uri, content string, positions []core.Position) []core.SelectionRange {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	result := make([]core.SelectionRange, len(positions))
	for i, pos := range positions {
		result[i] = p.buildSelectionRange(f, fset, content, pos)
	}

	return result
}

// buildSelectionRange creates a hierarchical selection range for a given position.
// The hierarchy goes from narrow to broad: identifier -> expression -> statement -> function -> file
func (p *GoSelectionRangeProvider) buildSelectionRange(f *ast.File, fset *token.FileSet, content string, pos core.Position) core.SelectionRange {
	offset := positionToOffset(content, pos)

	// Start with the smallest range at the position (the word/identifier)
	wordRange := p.getWordRange(content, offset)

	// Try to find the AST node at this position
	var currentNode ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		start := fset.Position(n.Pos())
		end := fset.Position(n.End())

		startOffset := positionToOffset(content, core.Position{Line: start.Line - 1, Character: start.Column - 1})
		endOffset := positionToOffset(content, core.Position{Line: end.Line - 1, Character: end.Column - 1})

		if startOffset <= offset && offset <= endOffset {
			if currentNode == nil || nodeSize(n, fset) < nodeSize(currentNode, fset) {
				currentNode = n
			}
		}

		return true
	})

	// Build the hierarchy from smallest to largest
	var selectionRange *core.SelectionRange

	// Start with word range
	selectionRange = &core.SelectionRange{
		Range: wordRange,
	}

	// Add progressively larger ranges from AST
	if currentNode != nil {
		selectionRange = p.buildASTHierarchy(f, fset, content, currentNode, selectionRange)
	}

	// Always have the full file as the outermost range
	fileRange := core.SelectionRange{
		Range: core.Range{
			Start: core.Position{Line: 0, Character: 0},
			End:   positionFromOffset(content, len(content)),
		},
	}

	// Link the chain
	if selectionRange != nil {
		current := selectionRange
		for current.Parent != nil {
			current = current.Parent
		}
		current.Parent = &fileRange
		return *selectionRange
	}

	return fileRange
}

// buildASTHierarchy creates a hierarchical chain of selection ranges from AST nodes
func (p *GoSelectionRangeProvider) buildASTHierarchy(f *ast.File, fset *token.FileSet, content string, node ast.Node, innerRange *core.SelectionRange) *core.SelectionRange {
	var ranges []*core.SelectionRange

	// Collect all ancestor nodes
	var ancestors []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check if n contains the original node
		if nodeContains(n, node, fset) && n != node {
			ancestors = append(ancestors, n)
		}

		return true
	})

	// Sort ancestors by size (smallest first)
	sortedAncestors := sortNodesBySize(ancestors, fset)

	// Build chain from innermost to outermost
	current := innerRange
	for _, ancestor := range sortedAncestors {
		start := fset.Position(ancestor.Pos())
		end := fset.Position(ancestor.End())

		ancestorRange := &core.SelectionRange{
			Range: core.Range{
				Start: core.Position{Line: start.Line - 1, Character: start.Column - 1},
				End:   core.Position{Line: end.Line - 1, Character: end.Column - 1},
			},
		}

		current.Parent = ancestorRange
		current = ancestorRange
		ranges = append(ranges, ancestorRange)
	}

	return innerRange
}

// getWordRange returns the range of the word at the given offset
func (p *GoSelectionRangeProvider) getWordRange(content string, offset int) core.Range {
	if offset < 0 || offset >= len(content) {
		return core.Range{
			Start: core.Position{Line: 0, Character: 0},
			End:   core.Position{Line: 0, Character: 0},
		}
	}

	// Find start of word
	start := offset
	for start > 0 && isIdentifierChar(rune(content[start-1])) {
		start--
	}

	// Find end of word
	end := offset
	for end < len(content) && isIdentifierChar(rune(content[end])) {
		end++
	}

	return core.Range{
		Start: positionFromOffset(content, start),
		End:   positionFromOffset(content, end),
	}
}

// Helper functions

func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

func nodeContains(outer, inner ast.Node, fset *token.FileSet) bool {
	outerStart := fset.Position(outer.Pos()).Offset
	outerEnd := fset.Position(outer.End()).Offset
	innerStart := fset.Position(inner.Pos()).Offset
	innerEnd := fset.Position(inner.End()).Offset

	return outerStart <= innerStart && innerEnd <= outerEnd
}

func nodeSize(node ast.Node, fset *token.FileSet) int {
	start := fset.Position(node.Pos()).Offset
	end := fset.Position(node.End()).Offset
	return end - start
}

func sortNodesBySize(nodes []ast.Node, fset *token.FileSet) []ast.Node {
	sorted := make([]ast.Node, len(nodes))
	copy(sorted, nodes)

	// Simple bubble sort by size
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if nodeSize(sorted[i], fset) > nodeSize(sorted[j], fset) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func positionToOffset(content string, pos core.Position) int {
	offset := 0
	currentLine := 0

	for i, r := range content {
		if currentLine == pos.Line {
			if offset+pos.Character <= i {
				return i
			}
		}

		if r == '\n' {
			currentLine++
			if currentLine > pos.Line {
				return i
			}
		}

		offset++
	}

	return len(content)
}

func positionFromOffset(content string, offset int) core.Position {
	line := 0
	lineStart := 0

	for i, r := range content {
		if i == offset {
			return core.Position{
				Line:      line,
				Character: offset - lineStart,
			}
		}

		if r == '\n' {
			line++
			lineStart = i + 1
		}
	}

	return core.Position{
		Line:      line,
		Character: offset - lineStart,
	}
}
