package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoWorkspaceSymbolProvider searches for Go symbols across a workspace.
// This is useful for "Go to Symbol in Workspace" functionality.
type GoWorkspaceSymbolProvider struct {
	// WorkspaceRoot is the root directory of the workspace
	WorkspaceRoot string

	// Cache of symbols indexed by file
	// In a real implementation, this would be updated on file changes
	symbolCache map[string][]core.WorkspaceSymbol
}

func NewGoWorkspaceSymbolProvider(workspaceRoot string) *GoWorkspaceSymbolProvider {
	return &GoWorkspaceSymbolProvider{
		WorkspaceRoot: workspaceRoot,
		symbolCache:   make(map[string][]core.WorkspaceSymbol),
	}
}

// IndexFile indexes symbols in a single Go file.
// This should be called when files are opened or changed.
func (p *GoWorkspaceSymbolProvider) IndexFile(uri, content string) {
	if !strings.HasSuffix(uri, ".go") {
		return
	}

	var symbols []core.WorkspaceSymbol

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		// Invalid syntax - clear symbols for this file
		p.symbolCache[uri] = nil
		return
	}

	// Extract package-level symbols
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			symbol := p.funcDeclToSymbol(d, fset, uri, f.Name.Name)
			if symbol != nil {
				symbols = append(symbols, *symbol)
			}

		case *ast.GenDecl:
			// Handle type, const, var declarations
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					symbol := p.typeSpecToSymbol(s, fset, uri, f.Name.Name)
					if symbol != nil {
						symbols = append(symbols, *symbol)
					}

				case *ast.ValueSpec:
					// Constants and variables
					kind := core.SymbolKindVariable
					if d.Tok == token.CONST {
						kind = core.SymbolKindConstant
					}

					for _, name := range s.Names {
						if name.Name == "_" {
							continue
						}

						pos := fset.Position(name.Pos())
						endPos := fset.Position(name.End())

						symbols = append(symbols, core.WorkspaceSymbol{
							Name:          name.Name,
							Kind:          kind,
							ContainerName: f.Name.Name,
							Location: core.Location{
								URI: uri,
								Range: core.Range{
									Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
									End:   core.Position{Line: endPos.Line - 1, Character: endPos.Column - 1},
								},
							},
						})
					}
				}
			}
		}
	}

	p.symbolCache[uri] = symbols
}

func (p *GoWorkspaceSymbolProvider) funcDeclToSymbol(fn *ast.FuncDecl, fset *token.FileSet, uri, packageName string) *core.WorkspaceSymbol {
	if fn.Name.Name == "_" {
		return nil
	}

	pos := fset.Position(fn.Name.Pos())
	nameEnd := fset.Position(fn.Name.End())

	kind := core.SymbolKindFunction
	containerName := packageName

	// If it has a receiver, it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		kind = core.SymbolKindMethod

		// Extract receiver type name as container
		recv := fn.Recv.List[0]
		if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				containerName = ident.Name
			}
		} else if ident, ok := recv.Type.(*ast.Ident); ok {
			containerName = ident.Name
		}
	}

	return &core.WorkspaceSymbol{
		Name:          fn.Name.Name,
		Kind:          kind,
		ContainerName: containerName,
		Location: core.Location{
			URI: uri,
			Range: core.Range{
				Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
				End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
			},
		},
	}
}

func (p *GoWorkspaceSymbolProvider) typeSpecToSymbol(ts *ast.TypeSpec, fset *token.FileSet, uri, packageName string) *core.WorkspaceSymbol {
	if ts.Name.Name == "_" {
		return nil
	}

	pos := fset.Position(ts.Name.Pos())
	nameEnd := fset.Position(ts.Name.End())

	kind := core.SymbolKindStruct
	switch ts.Type.(type) {
	case *ast.InterfaceType:
		kind = core.SymbolKindInterface
	case *ast.StructType:
		kind = core.SymbolKindStruct
	default:
		// Other types (aliases, etc.)
		kind = core.SymbolKindClass
	}

	return &core.WorkspaceSymbol{
		Name:          ts.Name.Name,
		Kind:          kind,
		ContainerName: packageName,
		Location: core.Location{
			URI: uri,
			Range: core.Range{
				Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
				End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
			},
		},
	}
}

// ProvideWorkspaceSymbols returns symbols matching the query.
// The query is matched against symbol names (case-insensitive substring match).
func (p *GoWorkspaceSymbolProvider) ProvideWorkspaceSymbols(query string) []core.WorkspaceSymbol {
	var results []core.WorkspaceSymbol

	// Normalize query for case-insensitive matching
	queryLower := strings.ToLower(query)

	// Search through all cached symbols
	for _, symbols := range p.symbolCache {
		for _, symbol := range symbols {
			// Match query against symbol name (case-insensitive)
			if query == "" || strings.Contains(strings.ToLower(symbol.Name), queryLower) {
				results = append(results, symbol)
			}
		}
	}

	return results
}

// SimpleWorkspaceSymbolProvider provides workspace symbols with a simple in-memory index.
// This is useful for small workspaces or testing.
type SimpleWorkspaceSymbolProvider struct {
	symbols []core.WorkspaceSymbol
}

func NewSimpleWorkspaceSymbolProvider() *SimpleWorkspaceSymbolProvider {
	return &SimpleWorkspaceSymbolProvider{
		symbols: []core.WorkspaceSymbol{},
	}
}

// AddSymbol adds a symbol to the index.
func (p *SimpleWorkspaceSymbolProvider) AddSymbol(symbol core.WorkspaceSymbol) {
	p.symbols = append(p.symbols, symbol)
}

// Clear removes all symbols from the index.
func (p *SimpleWorkspaceSymbolProvider) Clear() {
	p.symbols = []core.WorkspaceSymbol{}
}

// ProvideWorkspaceSymbols returns symbols matching the query.
func (p *SimpleWorkspaceSymbolProvider) ProvideWorkspaceSymbols(query string) []core.WorkspaceSymbol {
	if query == "" {
		// Return all symbols
		return p.symbols
	}

	var results []core.WorkspaceSymbol
	queryLower := strings.ToLower(query)

	for _, symbol := range p.symbols {
		// Case-insensitive substring match
		if strings.Contains(strings.ToLower(symbol.Name), queryLower) {
			results = append(results, symbol)
		}
	}

	return results
}

// FileSystemWorkspaceSymbolProvider scans the file system on demand.
// This is less efficient but doesn't require maintaining a cache.
type FileSystemWorkspaceSymbolProvider struct {
	WorkspaceRoot string
}

func (p *FileSystemWorkspaceSymbolProvider) ProvideWorkspaceSymbols(query string) []core.WorkspaceSymbol {
	var symbols []core.WorkspaceSymbol

	// Walk the workspace directory
	_ = filepath.Walk(p.WorkspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Only process .go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files and vendor directories
		if strings.HasSuffix(path, "_test.go") || strings.Contains(path, "/vendor/") {
			return nil
		}

		// Read and parse the file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, content, 0)
		if err != nil {
			return nil
		}

		// Extract symbols from this file
		uri := "file://" + path
		fileSymbols := p.extractSymbols(f, fset, uri, query)
		symbols = append(symbols, fileSymbols...)

		return nil
	})

	return symbols
}

func (p *FileSystemWorkspaceSymbolProvider) extractSymbols(f *ast.File, fset *token.FileSet, uri, query string) []core.WorkspaceSymbol {
	var symbols []core.WorkspaceSymbol
	queryLower := strings.ToLower(query)

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == "_" {
				continue
			}

			// Match query
			if query != "" && !strings.Contains(strings.ToLower(d.Name.Name), queryLower) {
				continue
			}

			kind := core.SymbolKindFunction
			containerName := f.Name.Name

			// Check if it's a method
			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = core.SymbolKindMethod
				// Extract receiver type as container
				recv := d.Recv.List[0]
				if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						containerName = ident.Name
					}
				} else if ident, ok := recv.Type.(*ast.Ident); ok {
					containerName = ident.Name
				}
			}

			pos := fset.Position(d.Name.Pos())
			endPos := fset.Position(d.Name.End())

			symbols = append(symbols, core.WorkspaceSymbol{
				Name:          d.Name.Name,
				Kind:          kind,
				ContainerName: containerName,
				Location: core.Location{
					URI: uri,
					Range: core.Range{
						Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
						End:   core.Position{Line: endPos.Line - 1, Character: endPos.Column - 1},
					},
				},
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					if ts.Name.Name == "_" {
						continue
					}

					// Match query
					if query != "" && !strings.Contains(strings.ToLower(ts.Name.Name), queryLower) {
						continue
					}

					kind := core.SymbolKindStruct
					switch ts.Type.(type) {
					case *ast.InterfaceType:
						kind = core.SymbolKindInterface
					case *ast.StructType:
						kind = core.SymbolKindStruct
					}

					pos := fset.Position(ts.Name.Pos())
					endPos := fset.Position(ts.Name.End())

					symbols = append(symbols, core.WorkspaceSymbol{
						Name:          ts.Name.Name,
						Kind:          kind,
						ContainerName: f.Name.Name,
						Location: core.Location{
							URI: uri,
							Range: core.Range{
								Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
								End:   core.Position{Line: endPos.Line - 1, Character: endPos.Column - 1},
							},
						},
					})
				}
			}
		}
	}

	return symbols
}

// Example usage in CLI tool
func CLIWorkspaceSymbolsExample() {
	provider := NewSimpleWorkspaceSymbolProvider()

	// Manually add some symbols (in real usage, these would be indexed from files)
	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "Server",
		Kind:          core.SymbolKindStruct,
		ContainerName: "main",
		Location: core.Location{
			URI: "file:///workspace/server.go",
			Range: core.Range{
				Start: core.Position{Line: 10, Character: 5},
				End:   core.Position{Line: 10, Character: 11},
			},
		},
	})

	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "HandleRequest",
		Kind:          core.SymbolKindMethod,
		ContainerName: "Server",
		Location: core.Location{
			URI: "file:///workspace/server.go",
			Range: core.Range{
				Start: core.Position{Line: 20, Character: 15},
				End:   core.Position{Line: 20, Character: 28},
			},
		},
	})

	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "StartServer",
		Kind:          core.SymbolKindFunction,
		ContainerName: "main",
		Location: core.Location{
			URI: "file:///workspace/main.go",
			Range: core.Range{
				Start: core.Position{Line: 5, Character: 5},
				End:   core.Position{Line: 5, Character: 16},
			},
		},
	})

	// Search for symbols
	results := provider.ProvideWorkspaceSymbols("server")

	println("Found", len(results), "symbols matching 'server':")
	for _, symbol := range results {
		container := ""
		if symbol.ContainerName != "" {
			container = " (" + symbol.ContainerName + ")"
		}
		println("  -", symbol.Name, container, "at", symbol.Location.URI)
	}
}

// Example usage in LSP server
// func (s *Server) WorkspaceSymbol(
// 	ctx *lsp.Context,
// 	params *protocol.WorkspaceSymbolParams,
// ) ([]protocol.WorkspaceSymbol, error) {
// 	// Use provider with core types
// 	coreSymbols := s.workspaceSymbolProvider.ProvideWorkspaceSymbols(params.Query)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolWorkspaceSymbols(coreSymbols), nil
// }
