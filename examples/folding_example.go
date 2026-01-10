package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoFoldingProvider provides folding ranges for Go source files.
type GoFoldingProvider struct{}

func (p *GoFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	var ranges []core.FoldingRange

	// Add import folding
	if importRange := p.getImportFolding(f, fset); importRange != nil {
		ranges = append(ranges, *importRange)
	}

	// Add comment folding
	ranges = append(ranges, p.getCommentFolding(f, fset)...)

	// Add function folding
	ranges = append(ranges, p.getFunctionFolding(f, fset)...)

	return ranges
}

func (p *GoFoldingProvider) getImportFolding(f *ast.File, fset *token.FileSet) *core.FoldingRange {
	if len(f.Imports) < 2 {
		return nil
	}

	first := fset.Position(f.Imports[0].Pos())
	last := fset.Position(f.Imports[len(f.Imports)-1].End())

	if first.Line == last.Line {
		return nil
	}

	kind := core.FoldingRangeKindImports
	return &core.FoldingRange{
		StartLine: first.Line - 1,
		EndLine:   last.Line - 1,
		Kind:      &kind,
	}
}

func (p *GoFoldingProvider) getCommentFolding(f *ast.File, fset *token.FileSet) []core.FoldingRange {
	var ranges []core.FoldingRange

	kind := core.FoldingRangeKindComment

	for _, cg := range f.Comments {
		if len(cg.List) < 2 {
			continue
		}

		start := fset.Position(cg.Pos())
		end := fset.Position(cg.End())

		if start.Line == end.Line {
			continue
		}

		ranges = append(ranges, core.FoldingRange{
			StartLine: start.Line - 1,
			EndLine:   end.Line - 1,
			Kind:      &kind,
		})
	}

	return ranges
}

func (p *GoFoldingProvider) getFunctionFolding(f *ast.File, fset *token.FileSet) []core.FoldingRange {
	var ranges []core.FoldingRange

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fn.Body == nil {
			return true
		}

		start := fset.Position(fn.Body.Lbrace)
		end := fset.Position(fn.Body.Rbrace)

		if start.Line == end.Line {
			return true
		}

		startChar := start.Column - 1
		ranges = append(ranges, core.FoldingRange{
			StartLine:      start.Line - 1,
			StartCharacter: &startChar,
			EndLine:        end.Line - 1,
		})

		return true
	})

	return ranges
}

// BraceFoldingProvider provides brace-based folding for any language.
type BraceFoldingProvider struct{}

func (p *BraceFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
	var ranges []core.FoldingRange

	lines := strings.Split(content, "\n")
	stack := []int{} // Stack of opening brace line numbers

	for lineNum, line := range lines {
		for _, ch := range line {
			if ch == '{' {
				stack = append(stack, lineNum)
			} else if ch == '}' {
				if len(stack) > 0 {
					startLine := stack[len(stack)-1]
					stack = stack[:len(stack)-1]

					if lineNum > startLine {
						ranges = append(ranges, core.FoldingRange{
							StartLine: startLine,
							EndLine:   lineNum,
						})
					}
				}
			}
		}
	}

	return ranges
}

// IndentFoldingProvider provides indentation-based folding.
type IndentFoldingProvider struct {
	TabSize int
}

func NewIndentFoldingProvider() *IndentFoldingProvider {
	return &IndentFoldingProvider{TabSize: 4}
}

func (p *IndentFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
	var ranges []core.FoldingRange

	lines := strings.Split(content, "\n")
	stack := []indentBlock{}

	for lineNum, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := p.getIndentLevel(line)

		// Pop blocks with greater or equal indentation
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			block := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if lineNum-1 > block.startLine {
				ranges = append(ranges, core.FoldingRange{
					StartLine: block.startLine,
					EndLine:   lineNum - 1,
				})
			}
		}

		stack = append(stack, indentBlock{
			startLine: lineNum,
			indent:    indent,
		})
	}

	// Handle remaining blocks
	endLine := len(lines) - 1
	for len(stack) > 0 {
		block := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if endLine > block.startLine {
			ranges = append(ranges, core.FoldingRange{
				StartLine: block.startLine,
				EndLine:   endLine,
			})
		}
	}

	return ranges
}

type indentBlock struct {
	startLine int
	indent    int
}

func (p *IndentFoldingProvider) getIndentLevel(line string) int {
	spaces := 0
	for _, ch := range line {
		if ch == ' ' {
			spaces++
		} else if ch == '\t' {
			spaces += p.TabSize
		} else {
			break
		}
	}
	return spaces / p.TabSize
}

// RegionFoldingProvider provides region-based folding.
type RegionFoldingProvider struct {
	StartMarker string
	EndMarker   string
}

func NewRegionFoldingProvider(startMarker, endMarker string) *RegionFoldingProvider {
	return &RegionFoldingProvider{
		StartMarker: startMarker,
		EndMarker:   endMarker,
	}
}

func (p *RegionFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
	var ranges []core.FoldingRange

	lines := strings.Split(content, "\n")
	stack := []int{}

	kind := core.FoldingRangeKindRegion

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, p.StartMarker) {
			stack = append(stack, lineNum)
		} else if strings.Contains(trimmed, p.EndMarker) {
			if len(stack) > 0 {
				startLine := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				ranges = append(ranges, core.FoldingRange{
					StartLine: startLine,
					EndLine:   lineNum,
					Kind:      &kind,
				})
			}
		}
	}

	return ranges
}

// CompositeFoldingProvider combines multiple folding providers.
type CompositeFoldingProvider struct {
	providers []core.FoldingRangeProvider
}

func NewCompositeFoldingProvider(providers ...core.FoldingRangeProvider) *CompositeFoldingProvider {
	return &CompositeFoldingProvider{
		providers: providers,
	}
}

func (p *CompositeFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
	var allRanges []core.FoldingRange

	for _, provider := range p.providers {
		if ranges := provider.ProvideFoldingRanges(uri, content); len(ranges) > 0 {
			allRanges = append(allRanges, ranges...)
		}
	}

	return p.deduplicateRanges(allRanges)
}

func (p *CompositeFoldingProvider) deduplicateRanges(ranges []core.FoldingRange) []core.FoldingRange {
	if len(ranges) == 0 {
		return ranges
	}

	// Sort by start line
	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].StartLine == ranges[j].StartLine {
			return ranges[i].EndLine < ranges[j].EndLine
		}
		return ranges[i].StartLine < ranges[j].StartLine
	})

	// Remove exact duplicates
	result := []core.FoldingRange{ranges[0]}

	for i := 1; i < len(ranges); i++ {
		curr := ranges[i]
		prev := result[len(result)-1]

		// Skip if identical (same start and end line)
		if curr.StartLine == prev.StartLine && curr.EndLine == prev.EndLine {
			// Prefer one with a kind
			if curr.Kind != nil && prev.Kind == nil {
				result[len(result)-1] = curr
			}
			continue
		}

		result = append(result, curr)
	}

	return result
}
