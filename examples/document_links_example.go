package examples

import (
	"regexp"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// URLLinkProvider finds HTTP/HTTPS URLs in documents.
// This is useful for making URLs in comments and strings clickable.
type URLLinkProvider struct{}

func (p *URLLinkProvider) ProvideDocumentLinks(uri, content string) []core.DocumentLink {
	var links []core.DocumentLink

	// Simple URL regex (matches http:// and https://)
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^\[\]` + "`" + `]+`)
	matches := urlRegex.FindAllStringIndex(content, -1)

	for _, match := range matches {
		start := match[0]
		end := match[1]

		url := content[start:end]
		startPos := core.ByteOffsetToPosition(content, start)
		endPos := core.ByteOffsetToPosition(content, end)

		links = append(links, core.DocumentLink{
			Range: core.Range{
				Start: startPos,
				End:   endPos,
			},
			Target: &url,
		})
	}

	return links
}

// GoImportLinkProvider finds Go import statements and creates file links.
// This makes import paths clickable in Go source files.
type GoImportLinkProvider struct {
	// ModulePath is the base module path (e.g., "github.com/user/repo")
	ModulePath string
	// SourceRoot is the file system path to the source root
	SourceRoot string
}

func (p *GoImportLinkProvider) ProvideDocumentLinks(uri, content string) []core.DocumentLink {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	var links []core.DocumentLink

	// Find import statements
	lines := strings.Split(content, "\n")
	lineOffset := 0

	for _, line := range lines {
		// Look for import statements: import "path" or "path" inside import ()
		if strings.Contains(line, "import") || (strings.Contains(line, "\"") && !strings.Contains(line, "//")) {
			// Find quoted strings (import paths)
			start := strings.Index(line, "\"")
			if start != -1 {
				end := strings.Index(line[start+1:], "\"")
				if end != -1 {
					end += start + 1

					importPath := line[start+1 : end]

					// Skip standard library imports (simple heuristic)
					if !strings.Contains(importPath, ".") {
						lineOffset += len(line) + 1 // +1 for newline
						continue
					}

					// Create link to the import path
					startOffset := lineOffset + start + 1 // +1 to skip opening quote
					endOffset := lineOffset + end

					startPos := core.ByteOffsetToPosition(content, startOffset)
					endPos := core.ByteOffsetToPosition(content, endOffset)

					// Convert import path to file URI
					target := p.importPathToURI(importPath)

					links = append(links, core.DocumentLink{
						Range: core.Range{
							Start: startPos,
							End:   endPos,
						},
						Target: &target,
					})
				}
			}
		}

		lineOffset += len(line) + 1 // +1 for newline
	}

	return links
}

func (p *GoImportLinkProvider) importPathToURI(importPath string) string {
	// Convert module-relative path to file URI
	// This is a simplified example - a real implementation would:
	// 1. Check if it's a relative import (./ or ../)
	// 2. Look in GOPATH/pkg/mod for external dependencies
	// 3. Handle standard library paths

	if strings.HasPrefix(importPath, p.ModulePath) {
		// Module-relative import
		relPath := strings.TrimPrefix(importPath, p.ModulePath)
		relPath = strings.TrimPrefix(relPath, "/")
		return "file://" + p.SourceRoot + "/" + relPath
	}

	// External module - would need to resolve from go.mod
	return "https://pkg.go.dev/" + importPath
}

// FilePathLinkProvider finds file paths in documents.
// This is useful for making file references in comments or strings clickable.
type FilePathLinkProvider struct {
	// WorkspaceRoot is the root directory of the workspace
	WorkspaceRoot string
}

func (p *FilePathLinkProvider) ProvideDocumentLinks(uri, content string) []core.DocumentLink {
	var links []core.DocumentLink

	// Find file path patterns (simplified - matches quoted strings that look like paths)
	// Matches: "./path", "../path", "/absolute/path", "filename.ext"
	pathRegex := regexp.MustCompile(`["'](\./[^\s"']+|\.\./[^\s"']+|/[^\s"']+|[a-zA-Z0-9_.-]+\.[a-zA-Z0-9]+)["']`)
	matches := pathRegex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		// match[2] and match[3] are the start/end of the first capture group (the path)
		if len(match) >= 4 {
			start := match[2]
			end := match[3]

			path := content[start:end]
			startPos := core.ByteOffsetToPosition(content, start)
			endPos := core.ByteOffsetToPosition(content, end)

			// Convert to file URI
			var target string
			if strings.HasPrefix(path, "/") {
				// Absolute path
				target = "file://" + path
			} else if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
				// Relative path - resolve against workspace root
				target = "file://" + p.WorkspaceRoot + "/" + path
			} else {
				// Just a filename - assume it's in the same directory
				target = "file://" + p.WorkspaceRoot + "/" + path
			}

			links = append(links, core.DocumentLink{
				Range: core.Range{
					Start: startPos,
					End:   endPos,
				},
				Target: &target,
			})
		}
	}

	return links
}

// MarkdownLinkProvider finds markdown-style links [text](url).
// This is useful for markdown and documentation files.
type MarkdownLinkProvider struct{}

func (p *MarkdownLinkProvider) ProvideDocumentLinks(uri, content string) []core.DocumentLink {
	var links []core.DocumentLink

	// Match markdown links: [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := linkRegex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		// match[4] and match[5] are the start/end of the URL (second capture group)
		if len(match) >= 6 {
			start := match[4]
			end := match[5]

			url := content[start:end]
			startPos := core.ByteOffsetToPosition(content, start)
			endPos := core.ByteOffsetToPosition(content, end)

			links = append(links, core.DocumentLink{
				Range: core.Range{
					Start: startPos,
					End:   endPos,
				},
				Target: &url,
			})
		}
	}

	return links
}

// CompositeDocumentLinkProvider combines multiple link providers.
// This allows detecting multiple types of links in a single document.
type CompositeDocumentLinkProvider struct {
	Providers []core.DocumentLinkProvider
}

func NewCompositeDocumentLinkProvider(providers ...core.DocumentLinkProvider) *CompositeDocumentLinkProvider {
	return &CompositeDocumentLinkProvider{
		Providers: providers,
	}
}

func (p *CompositeDocumentLinkProvider) ProvideDocumentLinks(uri, content string) []core.DocumentLink {
	var allLinks []core.DocumentLink

	for _, provider := range p.Providers {
		links := provider.ProvideDocumentLinks(uri, content)
		if links != nil {
			allLinks = append(allLinks, links...)
		}
	}

	return allLinks
}

// Example usage in CLI tool
func CLIDocumentLinksExample() {
	content := `package main

import (
	"fmt"
	"github.com/user/repo/pkg"
)

// See documentation at https://example.com/docs
// Related file: ./helper.go
func main() {
	// Check out [the guide](https://example.com/guide)
	fmt.Println("Hello")
}
`

	provider := NewCompositeDocumentLinkProvider(
		&URLLinkProvider{},
		&GoImportLinkProvider{
			ModulePath: "github.com/user/repo",
			SourceRoot: "/workspace/repo",
		},
		&FilePathLinkProvider{
			WorkspaceRoot: "/workspace/repo",
		},
		&MarkdownLinkProvider{},
	)

	links := provider.ProvideDocumentLinks("file:///main.go", content)

	println("Found", len(links), "links:")
	for _, link := range links {
		target := "<unresolved>"
		if link.Target != nil {
			target = *link.Target
		}
		println("  -", link.Range.String(), "->", target)
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentDocumentLink(
// 	ctx *lsp.Context,
// 	params *protocol.DocumentLinkParams,
// ) ([]protocol.DocumentLink, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Use provider with core types
// 	coreLinks := s.linkProvider.ProvideDocumentLinks(uri, content)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolDocumentLinks(coreLinks, content), nil
// }
