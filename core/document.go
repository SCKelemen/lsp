package core

import (
	"sync"
)

// Document represents a text document with its content and metadata.
// This is a convenience type for managing documents in memory and working with core types.
type Document struct {
	URI     string
	Content string
	Version int
	mu      sync.RWMutex
}

// NewDocument creates a new document with the given URI and content.
func NewDocument(uri, content string, version int) *Document {
	return &Document{
		URI:     uri,
		Content: content,
		Version: version,
	}
}

// GetContent returns the current content of the document.
func (d *Document) GetContent() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Content
}

// SetContent updates the document content and increments the version.
func (d *Document) SetContent(content string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Content = content
	d.Version++
}

// ApplyEdit applies a text edit to the document.
// The range and replacement text use UTF-8 byte offsets.
func (d *Document) ApplyEdit(r Range, newText string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	content := d.Content
	start := PositionToByteOffset(content, r.Start)
	end := PositionToByteOffset(content, r.End)

	d.Content = content[:start] + newText + content[end:]
	d.Version++
}

// DocumentManager manages a collection of documents.
// This is useful for LSP server implementations that need to track open documents.
type DocumentManager struct {
	documents map[string]*Document
	mu        sync.RWMutex
}

// NewDocumentManager creates a new document manager.
func NewDocumentManager() *DocumentManager {
	return &DocumentManager{
		documents: make(map[string]*Document),
	}
}

// Open adds or updates a document in the manager.
func (dm *DocumentManager) Open(uri, content string, version int) *Document {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc := NewDocument(uri, content, version)
	dm.documents[uri] = doc
	return doc
}

// Get retrieves a document by URI.
func (dm *DocumentManager) Get(uri string) (*Document, bool) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	doc, ok := dm.documents[uri]
	return doc, ok
}

// Close removes a document from the manager.
func (dm *DocumentManager) Close(uri string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.documents, uri)
}

// Update updates a document's content.
func (dm *DocumentManager) Update(uri, content string) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	doc, ok := dm.documents[uri]
	if !ok {
		return false
	}

	doc.Content = content
	doc.Version++
	return true
}

// GetContent is a convenience method to get document content by URI.
// Returns empty string if document not found.
func (dm *DocumentManager) GetContent(uri string) string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if doc, ok := dm.documents[uri]; ok {
		return doc.Content
	}
	return ""
}

// ApplyEdit applies a text edit to a document.
func (dm *DocumentManager) ApplyEdit(uri string, r Range, newText string) bool {
	doc, ok := dm.Get(uri)
	if !ok {
		return false
	}

	doc.ApplyEdit(r, newText)
	return true
}
