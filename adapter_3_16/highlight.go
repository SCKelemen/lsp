package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol_3_16"
)

// CoreToProtocolDocumentHighlightKind converts a core highlight kind to protocol.
func CoreToProtocolDocumentHighlightKind(kind core.DocumentHighlightKind) protocol.DocumentHighlightKind {
	return protocol.DocumentHighlightKind(kind)
}

// ProtocolToCoreDocumentHighlightKind converts a protocol highlight kind to core.
func ProtocolToCoreDocumentHighlightKind(kind protocol.DocumentHighlightKind) core.DocumentHighlightKind {
	return core.DocumentHighlightKind(kind)
}

// CoreToProtocolDocumentHighlight converts a core document highlight to protocol.
func CoreToProtocolDocumentHighlight(highlight core.DocumentHighlight, content string) protocol.DocumentHighlight {
	result := protocol.DocumentHighlight{
		Range: CoreToProtocolRange(highlight.Range, content),
	}

	if highlight.Kind != nil {
		kind := CoreToProtocolDocumentHighlightKind(*highlight.Kind)
		result.Kind = &kind
	}

	return result
}

// ProtocolToCoreDocumentHighlight converts a protocol document highlight to core.
func ProtocolToCoreDocumentHighlight(highlight protocol.DocumentHighlight, content string) core.DocumentHighlight {
	result := core.DocumentHighlight{
		Range: ProtocolToCoreRange(highlight.Range, content),
	}

	if highlight.Kind != nil {
		kind := ProtocolToCoreDocumentHighlightKind(*highlight.Kind)
		result.Kind = &kind
	}

	return result
}

// CoreToProtocolDocumentHighlights converts a slice of core highlights to protocol.
func CoreToProtocolDocumentHighlights(highlights []core.DocumentHighlight, content string) []protocol.DocumentHighlight {
	result := make([]protocol.DocumentHighlight, len(highlights))
	for i, h := range highlights {
		result[i] = CoreToProtocolDocumentHighlight(h, content)
	}
	return result
}

// ProtocolToCoreDocumentHighlights converts a slice of protocol highlights to core.
func ProtocolToCoreDocumentHighlights(highlights []protocol.DocumentHighlight, content string) []core.DocumentHighlight {
	result := make([]core.DocumentHighlight, len(highlights))
	for i, h := range highlights {
		result[i] = ProtocolToCoreDocumentHighlight(h, content)
	}
	return result
}
