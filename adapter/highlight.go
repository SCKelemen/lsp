package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
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

// CoreToProtocolReferenceContext converts a core reference context to protocol.
func CoreToProtocolReferenceContext(ctx core.ReferenceContext) protocol.ReferenceContext {
	return protocol.ReferenceContext{
		IncludeDeclaration: ctx.IncludeDeclaration,
	}
}

// ProtocolToCoreReferenceContext converts a protocol reference context to core.
func ProtocolToCoreReferenceContext(ctx protocol.ReferenceContext) core.ReferenceContext {
	return core.ReferenceContext{
		IncludeDeclaration: ctx.IncludeDeclaration,
	}
}

// CoreToProtocolInlayHintKind converts a core inlay hint kind to protocol.
func CoreToProtocolInlayHintKind(kind core.InlayHintKind) protocol.InlayHintKind {
	return protocol.InlayHintKind(kind)
}

// ProtocolToCoreInlayHintKind converts a protocol inlay hint kind to core.
func ProtocolToCoreInlayHintKind(kind protocol.InlayHintKind) core.InlayHintKind {
	return core.InlayHintKind(kind)
}

// CoreToProtocolInlayHint converts a core inlay hint to protocol.
func CoreToProtocolInlayHint(hint core.InlayHint, content string) protocol.InlayHint {
	result := protocol.InlayHint{
		Position: CoreToProtocolPosition(hint.Position, content),
		Label:    hint.Label,
	}

	if hint.Kind != nil {
		kind := CoreToProtocolInlayHintKind(*hint.Kind)
		result.Kind = &kind
	}

	if len(hint.TextEdits) > 0 {
		result.TextEdits = CoreToProtocolTextEdits(hint.TextEdits, content)
	}

	if hint.Tooltip != "" {
		result.Tooltip = protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: hint.Tooltip,
		}
	}

	result.PaddingLeft = &hint.PaddingLeft
	result.PaddingRight = &hint.PaddingRight

	if hint.Data != nil {
		result.Data = hint.Data
	}

	return result
}

// ProtocolToCoreInlayHint converts a protocol inlay hint to core.
func ProtocolToCoreInlayHint(hint protocol.InlayHint, content string) core.InlayHint {
	result := core.InlayHint{
		Position: ProtocolToCorePosition(hint.Position, content),
	}

	// Label handling - protocol can have structured labels, we simplify to string
	switch label := hint.Label.(type) {
	case string:
		result.Label = label
	case []protocol.InlayHintLabelPart:
		// Concatenate parts
		var str string
		for _, part := range label {
			str += part.Value
		}
		result.Label = str
	}

	if hint.Kind != nil {
		kind := ProtocolToCoreInlayHintKind(*hint.Kind)
		result.Kind = &kind
	}

	if hint.Tooltip != nil {
		// Handle both string and MarkupContent
		switch v := hint.Tooltip.(type) {
		case string:
			result.Tooltip = v
		case protocol.MarkupContent:
			result.Tooltip = v.Value
		}
	}

	if hint.TextEdits != nil {
		result.TextEdits = ProtocolToCoreTextEdits(hint.TextEdits, content)
	}

	result.PaddingLeft = hint.PaddingLeft != nil && *hint.PaddingLeft
	result.PaddingRight = hint.PaddingRight != nil && *hint.PaddingRight

	result.Data = hint.Data

	return result
}

// CoreToProtocolInlayHints converts a slice of core inlay hints to protocol.
func CoreToProtocolInlayHints(hints []core.InlayHint, content string) []protocol.InlayHint {
	if len(hints) == 0 {
		return nil
	}

	result := make([]protocol.InlayHint, len(hints))
	for i, h := range hints {
		result[i] = CoreToProtocolInlayHint(h, content)
	}
	return result
}

// ProtocolToCoreInlayHints converts a slice of protocol inlay hints to core.
func ProtocolToCoreInlayHints(hints []protocol.InlayHint, content string) []core.InlayHint {
	result := make([]core.InlayHint, len(hints))
	for i, h := range hints {
		result[i] = ProtocolToCoreInlayHint(h, content)
	}
	return result
}

