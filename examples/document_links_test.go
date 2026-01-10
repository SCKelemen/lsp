package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestURLLinkProvider tests URL link detection.
func TestURLLinkProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantURLs  []string
	}{
		{
			name: "single HTTP URL",
			content: `// See https://example.com for details
func main() {}`,
			wantCount: 1,
			wantURLs:  []string{"https://example.com"},
		},
		{
			name: "multiple URLs",
			content: `// Links:
// - https://example.com
// - http://test.org/path
// - https://github.com/user/repo`,
			wantCount: 3,
			wantURLs:  []string{"https://example.com", "http://test.org/path", "https://github.com/user/repo"},
		},
		{
			name: "URL in string",
			content: `const doc = "Visit https://example.com/docs"
`,
			wantCount: 1,
			wantURLs:  []string{"https://example.com/docs"},
		},
		{
			name:      "no URLs",
			content:   `func main() { println("hello") }`,
			wantCount: 0,
			wantURLs:  nil,
		},
		{
			name: "URL with query parameters",
			content: `// https://example.com/path?foo=bar&baz=qux
`,
			wantCount: 1,
			wantURLs:  []string{"https://example.com/path?foo=bar&baz=qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &URLLinkProvider{}
			links := provider.ProvideDocumentLinks("file:///test.go", tt.content)

			if len(links) != tt.wantCount {
				t.Errorf("got %d links, want %d", len(links), tt.wantCount)
			}

			// Verify URLs
			for i, link := range links {
				if link.Target == nil {
					t.Errorf("link %d: target is nil", i)
					continue
				}

				if i < len(tt.wantURLs) && *link.Target != tt.wantURLs[i] {
					t.Errorf("link %d: got URL %q, want %q", i, *link.Target, tt.wantURLs[i])
				}

				// Verify the range extracts the correct URL
				startOffset := core.PositionToByteOffset(tt.content, link.Range.Start)
				endOffset := core.PositionToByteOffset(tt.content, link.Range.End)

				if startOffset < 0 || endOffset > len(tt.content) || startOffset >= endOffset {
					t.Errorf("link %d: invalid range offsets [%d:%d]", i, startOffset, endOffset)
					continue
				}

				extractedURL := tt.content[startOffset:endOffset]
				if extractedURL != *link.Target {
					t.Errorf("link %d: range extracts %q but target is %q", i, extractedURL, *link.Target)
				}
			}
		})
	}
}

// TestGoImportLinkProvider tests Go import link detection.
func TestGoImportLinkProvider(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		modulePath string
		wantCount  int
		checkLink  func(t *testing.T, link core.DocumentLink, content string)
	}{
		{
			name: "single import",
			content: `package main

import "github.com/user/repo/pkg"

func main() {}`,
			modulePath: "github.com/user/repo",
			wantCount:  1,
			checkLink: func(t *testing.T, link core.DocumentLink, content string) {
				if link.Target == nil {
					t.Error("target is nil")
					return
				}
				if !strings.Contains(*link.Target, "pkg") {
					t.Errorf("target %q should contain 'pkg'", *link.Target)
				}
			},
		},
		{
			name: "multiple imports",
			content: `package main

import (
	"github.com/user/repo/pkg1"
	"github.com/other/lib"
)`,
			modulePath: "github.com/user/repo",
			wantCount:  2,
		},
		{
			name: "skip standard library",
			content: `package main

import (
	"fmt"
	"os"
	"github.com/user/repo/pkg"
)`,
			modulePath: "github.com/user/repo",
			wantCount:  1, // Only the third-party import
		},
		{
			name:       "no imports",
			content:    `package main\n\nfunc main() {}`,
			modulePath: "github.com/user/repo",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoImportLinkProvider{
				ModulePath: tt.modulePath,
				SourceRoot: "/workspace/repo",
			}

			links := provider.ProvideDocumentLinks("file:///test.go", tt.content)

			if len(links) != tt.wantCount {
				t.Errorf("got %d links, want %d", len(links), tt.wantCount)
			}

			if tt.checkLink != nil && len(links) > 0 {
				tt.checkLink(t, links[0], tt.content)
			}

			// Verify all ranges are valid
			for i, link := range links {
				startOffset := core.PositionToByteOffset(tt.content, link.Range.Start)
				endOffset := core.PositionToByteOffset(tt.content, link.Range.End)

				if startOffset < 0 || endOffset > len(tt.content) || startOffset >= endOffset {
					t.Errorf("link %d: invalid range offsets [%d:%d]", i, startOffset, endOffset)
				}
			}
		})
	}
}

// TestGoImportLinkProvider_NonGoFile tests that provider skips non-Go files.
func TestGoImportLinkProvider_NonGoFile(t *testing.T) {
	provider := &GoImportLinkProvider{
		ModulePath: "github.com/user/repo",
		SourceRoot: "/workspace",
	}

	content := `import "github.com/user/repo/pkg"`

	links := provider.ProvideDocumentLinks("file:///test.txt", content)

	if links != nil {
		t.Errorf("expected nil for non-Go file, got %d links", len(links))
	}
}

// TestFilePathLinkProvider tests file path link detection.
func TestFilePathLinkProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		checkURLs func(t *testing.T, links []core.DocumentLink)
	}{
		{
			name: "relative path",
			content: `// See "./helper.go" for details
`,
			wantCount: 1,
			checkURLs: func(t *testing.T, links []core.DocumentLink) {
				if links[0].Target == nil {
					t.Error("target is nil")
					return
				}
				if !strings.Contains(*links[0].Target, "helper.go") {
					t.Errorf("target %q should contain 'helper.go'", *links[0].Target)
				}
			},
		},
		{
			name: "parent directory path",
			content: `// Import from "../lib/utils.go"
`,
			wantCount: 1,
		},
		{
			name: "absolute path",
			content: `// Config at "/etc/config.yml"
`,
			wantCount: 1,
			checkURLs: func(t *testing.T, links []core.DocumentLink) {
				if links[0].Target == nil {
					t.Error("target is nil")
					return
				}
				if !strings.HasPrefix(*links[0].Target, "file:///etc") {
					t.Errorf("absolute path should start with file:///etc, got %q", *links[0].Target)
				}
			},
		},
		{
			name: "simple filename",
			content: `// See "README.md"
`,
			wantCount: 1,
		},
		{
			name:      "no file paths",
			content:   `func main() { println(42) }`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &FilePathLinkProvider{
				WorkspaceRoot: "/workspace",
			}

			links := provider.ProvideDocumentLinks("file:///test.go", tt.content)

			if len(links) != tt.wantCount {
				t.Errorf("got %d links, want %d", len(links), tt.wantCount)
			}

			if tt.checkURLs != nil && len(links) > 0 {
				tt.checkURLs(t, links)
			}

			// Verify ranges
			for i, link := range links {
				if link.Range.Start.Line < 0 || link.Range.Start.Character < 0 {
					t.Errorf("link %d: negative position in range start", i)
				}
			}
		})
	}
}

// TestMarkdownLinkProvider tests markdown link detection.
func TestMarkdownLinkProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantURLs  []string
	}{
		{
			name:      "single markdown link",
			content:   `Check out [the docs](https://example.com/docs)`,
			wantCount: 1,
			wantURLs:  []string{"https://example.com/docs"},
		},
		{
			name: "multiple markdown links",
			content: `Resources:
- [Guide](https://example.com/guide)
- [API Reference](https://example.com/api)
- [Examples](https://github.com/user/repo)`,
			wantCount: 3,
			wantURLs:  []string{"https://example.com/guide", "https://example.com/api", "https://github.com/user/repo"},
		},
		{
			name:      "relative file link",
			content:   `See [helper](./helper.go) for details`,
			wantCount: 1,
			wantURLs:  []string{"./helper.go"},
		},
		{
			name:      "no markdown links",
			content:   `Just plain text with [brackets] and (parentheses) but not together`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MarkdownLinkProvider{}
			links := provider.ProvideDocumentLinks("file:///test.md", tt.content)

			if len(links) != tt.wantCount {
				t.Errorf("got %d links, want %d", len(links), tt.wantCount)
			}

			for i, link := range links {
				if link.Target == nil {
					t.Errorf("link %d: target is nil", i)
					continue
				}

				if i < len(tt.wantURLs) && *link.Target != tt.wantURLs[i] {
					t.Errorf("link %d: got URL %q, want %q", i, *link.Target, tt.wantURLs[i])
				}

				// Verify range extracts the URL part
				startOffset := core.PositionToByteOffset(tt.content, link.Range.Start)
				endOffset := core.PositionToByteOffset(tt.content, link.Range.End)

				if startOffset < 0 || endOffset > len(tt.content) || startOffset >= endOffset {
					t.Errorf("link %d: invalid range offsets [%d:%d]", i, startOffset, endOffset)
					continue
				}

				extractedURL := tt.content[startOffset:endOffset]
				if extractedURL != *link.Target {
					t.Errorf("link %d: range extracts %q but target is %q", i, extractedURL, *link.Target)
				}
			}
		})
	}
}

// TestCompositeDocumentLinkProvider tests combining multiple providers.
func TestCompositeDocumentLinkProvider(t *testing.T) {
	content := `// Documentation: https://example.com/docs
// See also [guide](https://example.com/guide)
// Related file: "./helper.go"

import "github.com/user/repo/pkg"
`

	provider := NewCompositeDocumentLinkProvider(
		&URLLinkProvider{},
		&MarkdownLinkProvider{},
		&FilePathLinkProvider{WorkspaceRoot: "/workspace"},
		&GoImportLinkProvider{
			ModulePath: "github.com/user/repo",
			SourceRoot: "/workspace",
		},
	)

	links := provider.ProvideDocumentLinks("file:///test.go", content)

	// Should find:
	// - 2 URLs (https://example.com/docs and https://example.com/guide)
	// - 1 markdown link URL (same as one of the URLs)
	// - 1 file path (./helper.go)
	// - 1 import path

	// Note: There's overlap (the markdown link URL is also detected by URLLinkProvider)
	// So we expect at least 4 unique logical links, but may have duplicates
	if len(links) < 4 {
		t.Errorf("got %d links, want at least 4", len(links))
	}

	// Verify all links have valid ranges
	for i, link := range links {
		if link.Range.Start.Line < 0 || link.Range.Start.Character < 0 {
			t.Errorf("link %d: negative position", i)
		}

		startOffset := core.PositionToByteOffset(content, link.Range.Start)
		endOffset := core.PositionToByteOffset(content, link.Range.End)

		if startOffset < 0 || endOffset > len(content) || startOffset >= endOffset {
			t.Errorf("link %d: invalid offsets [%d:%d]", i, startOffset, endOffset)
		}
	}

	// Count link types
	urlCount := 0
	fileCount := 0
	importCount := 0

	for _, link := range links {
		if link.Target == nil {
			continue
		}

		target := *link.Target
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			urlCount++
		} else if strings.Contains(target, "helper.go") {
			fileCount++
		} else if strings.Contains(target, "pkg") {
			importCount++
		}
	}

	t.Logf("Found %d URLs, %d file paths, %d imports", urlCount, fileCount, importCount)

	if urlCount == 0 {
		t.Error("expected to find URLs")
	}
	if fileCount == 0 {
		t.Error("expected to find file path")
	}
	if importCount == 0 {
		t.Error("expected to find import")
	}
}

// TestDocumentLinks_EdgeCases tests edge cases.
func TestDocumentLinks_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty file",
			content: "",
		},
		{
			name:    "only whitespace",
			content: "   \n\t\n   ",
		},
		{
			name:    "no links",
			content: "func main() { println(42) }",
		},
	}

	providers := []core.DocumentLinkProvider{
		&URLLinkProvider{},
		&GoImportLinkProvider{ModulePath: "test", SourceRoot: "/test"},
		&FilePathLinkProvider{WorkspaceRoot: "/test"},
		&MarkdownLinkProvider{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, provider := range providers {
				// Should not crash
				links := provider.ProvideDocumentLinks("file:///test.go", tt.content)

				if links == nil {
					links = []core.DocumentLink{}
				}

				t.Logf("provider %d returned %d links", i, len(links))

				// Verify all links have valid ranges
				for j, link := range links {
					if link.Range.Start.Line < 0 || link.Range.Start.Character < 0 {
						t.Errorf("provider %d, link %d: negative position", i, j)
					}
				}
			}
		})
	}
}

// TestDocumentLinks_RangeAccuracy tests that ranges are accurate.
func TestDocumentLinks_RangeAccuracy(t *testing.T) {
	content := `// https://example.com
// [link](https://test.org)
import "github.com/user/repo"
`

	provider := NewCompositeDocumentLinkProvider(
		&URLLinkProvider{},
		&MarkdownLinkProvider{},
		&GoImportLinkProvider{ModulePath: "github.com", SourceRoot: "/test"},
	)

	links := provider.ProvideDocumentLinks("file:///test.go", content)

	for i, link := range links {
		// Verify range is valid
		if link.Range.Start.Line < 0 || link.Range.Start.Character < 0 {
			t.Errorf("link %d: negative start position", i)
		}

		if link.Range.End.Line < link.Range.Start.Line {
			t.Errorf("link %d: end line before start line", i)
		}

		if link.Range.End.Line == link.Range.Start.Line && link.Range.End.Character <= link.Range.Start.Character {
			t.Errorf("link %d: end character not after start character", i)
		}

		// Convert to byte offsets and verify
		startOffset := core.PositionToByteOffset(content, link.Range.Start)
		endOffset := core.PositionToByteOffset(content, link.Range.End)

		if startOffset < 0 || startOffset >= len(content) {
			t.Errorf("link %d: start offset %d out of bounds [0,%d)", i, startOffset, len(content))
		}

		if endOffset < 0 || endOffset > len(content) {
			t.Errorf("link %d: end offset %d out of bounds [0,%d]", i, endOffset, len(content))
		}

		if startOffset >= endOffset {
			t.Errorf("link %d: start offset %d >= end offset %d", i, startOffset, endOffset)
		}
	}
}

// TestDocumentLinks_Unicode tests handling of multibyte characters.
func TestDocumentLinks_Unicode(t *testing.T) {
	content := `// 文档：https://example.com/文档
// 参考 [指南](https://example.com/guide)
`

	provider := NewCompositeDocumentLinkProvider(
		&URLLinkProvider{},
		&MarkdownLinkProvider{},
	)

	links := provider.ProvideDocumentLinks("file:///test.go", content)

	if len(links) < 2 {
		t.Errorf("got %d links, want at least 2", len(links))
	}

	for i, link := range links {
		startOffset := core.PositionToByteOffset(content, link.Range.Start)
		endOffset := core.PositionToByteOffset(content, link.Range.End)

		if startOffset < 0 || endOffset > len(content) || startOffset >= endOffset {
			t.Errorf("link %d: invalid byte offsets [%d:%d]", i, startOffset, endOffset)
			continue
		}

		// Verify positions are valid UTF-8 byte offsets
		lines := strings.Split(content, "\n")
		if link.Range.Start.Line < len(lines) {
			line := lines[link.Range.Start.Line]
			if link.Range.Start.Character > len(line) {
				t.Errorf("link %d: start character %d exceeds line length %d (UTF-8 bytes)",
					i, link.Range.Start.Character, len(line))
			}
		}
	}
}
