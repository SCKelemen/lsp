package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol_3_16"
)

// CoreToProtocolFoldingRange converts a core.FoldingRange to a protocol FoldingRange.
func CoreToProtocolFoldingRange(fr core.FoldingRange, content string) protocol.FoldingRange {
	result := protocol.FoldingRange{
		StartLine: protocol.UInteger(fr.StartLine),
		EndLine:   protocol.UInteger(fr.EndLine),
	}

	// Convert start character if present
	if fr.StartCharacter != nil {
		// Get the line content to convert UTF-8 to UTF-16
		utf16Offset := core.UTF8ToUTF16Offset(content, fr.StartLine, *fr.StartCharacter)
		u := protocol.UInteger(utf16Offset)
		result.StartCharacter = &u
	}

	// Convert end character if present
	if fr.EndCharacter != nil {
		utf16Offset := core.UTF8ToUTF16Offset(content, fr.EndLine, *fr.EndCharacter)
		u := protocol.UInteger(utf16Offset)
		result.EndCharacter = &u
	}

	// Convert kind
	if fr.Kind != nil {
		kind := string(*fr.Kind)
		result.Kind = &kind
	}

	return result
}

// ProtocolToCoreFoldingRange converts a protocol FoldingRange to a core.FoldingRange.
func ProtocolToCoreFoldingRange(fr protocol.FoldingRange, content string) core.FoldingRange {
	result := core.FoldingRange{
		StartLine: int(fr.StartLine),
		EndLine:   int(fr.EndLine),
	}

	// Convert start character if present
	if fr.StartCharacter != nil {
		utf8Offset := core.UTF16ToUTF8Offset(content, result.StartLine, int(*fr.StartCharacter))
		result.StartCharacter = &utf8Offset
	}

	// Convert end character if present
	if fr.EndCharacter != nil {
		utf8Offset := core.UTF16ToUTF8Offset(content, result.EndLine, int(*fr.EndCharacter))
		result.EndCharacter = &utf8Offset
	}

	// Convert kind
	if fr.Kind != nil {
		kind := core.FoldingRangeKind(*fr.Kind)
		result.Kind = &kind
	}

	return result
}

// CoreToProtocolFoldingRanges converts a slice of core folding ranges to protocol folding ranges.
func CoreToProtocolFoldingRanges(ranges []core.FoldingRange, content string) []protocol.FoldingRange {
	result := make([]protocol.FoldingRange, len(ranges))
	for i, r := range ranges {
		result[i] = CoreToProtocolFoldingRange(r, content)
	}
	return result
}

// ProtocolToCoreFoldingRanges converts a slice of protocol folding ranges to core folding ranges.
func ProtocolToCoreFoldingRanges(ranges []protocol.FoldingRange, content string) []core.FoldingRange {
	result := make([]core.FoldingRange, len(ranges))
	for i, r := range ranges {
		result[i] = ProtocolToCoreFoldingRange(r, content)
	}
	return result
}

// CoreToProtocolTextEdit converts a core.TextEdit to a protocol TextEdit.
func CoreToProtocolTextEdit(edit core.TextEdit, content string) protocol.TextEdit {
	return protocol.TextEdit{
		Range:   CoreToProtocolRange(edit.Range, content),
		NewText: edit.NewText,
	}
}

// ProtocolToCoreTextEdit converts a protocol TextEdit to a core.TextEdit.
func ProtocolToCoreTextEdit(edit protocol.TextEdit, content string) core.TextEdit {
	return core.TextEdit{
		Range:   ProtocolToCoreRange(edit.Range, content),
		NewText: edit.NewText,
	}
}

// CoreToProtocolTextEdits converts a slice of core text edits to protocol text edits.
func CoreToProtocolTextEdits(edits []core.TextEdit, content string) []protocol.TextEdit {
	result := make([]protocol.TextEdit, len(edits))
	for i, edit := range edits {
		result[i] = CoreToProtocolTextEdit(edit, content)
	}
	return result
}

// ProtocolToCoreTextEdits converts a slice of protocol text edits to core text edits.
func ProtocolToCoreTextEdits(edits []protocol.TextEdit, content string) []core.TextEdit {
	result := make([]core.TextEdit, len(edits))
	for i, edit := range edits {
		result[i] = ProtocolToCoreTextEdit(edit, content)
	}
	return result
}

// CoreToProtocolSymbolKind converts a core symbol kind to protocol symbol kind.
func CoreToProtocolSymbolKind(kind core.SymbolKind) protocol.SymbolKind {
	return protocol.SymbolKind(kind)
}

// ProtocolToCoreSymbolKind converts a protocol symbol kind to core symbol kind.
func ProtocolToCoreSymbolKind(kind protocol.SymbolKind) core.SymbolKind {
	return core.SymbolKind(kind)
}

// CoreToProtocolDocumentSymbol converts a core.DocumentSymbol to a protocol DocumentSymbol.
func CoreToProtocolDocumentSymbol(sym core.DocumentSymbol, content string) protocol.DocumentSymbol {
	result := protocol.DocumentSymbol{
		Name:           sym.Name,
		Kind:           CoreToProtocolSymbolKind(sym.Kind),
		Range:          CoreToProtocolRange(sym.Range, content),
		SelectionRange: CoreToProtocolRange(sym.SelectionRange, content),
	}

	// Convert detail if present
	if sym.Detail != "" {
		result.Detail = &sym.Detail
	}

	// Convert tags
	if len(sym.Tags) > 0 {
		tags := make([]protocol.SymbolTag, len(sym.Tags))
		for i, tag := range sym.Tags {
			tags[i] = protocol.SymbolTag(tag)
		}
		result.Tags = tags
	}

	// Convert deprecated flag
	if sym.Deprecated {
		result.Deprecated = &sym.Deprecated
	}

	// Convert children recursively
	if len(sym.Children) > 0 {
		children := make([]protocol.DocumentSymbol, len(sym.Children))
		for i, child := range sym.Children {
			children[i] = CoreToProtocolDocumentSymbol(child, content)
		}
		result.Children = children
	}

	return result
}

// ProtocolToCoreDocumentSymbol converts a protocol DocumentSymbol to a core.DocumentSymbol.
func ProtocolToCoreDocumentSymbol(sym protocol.DocumentSymbol, content string) core.DocumentSymbol {
	result := core.DocumentSymbol{
		Name:           sym.Name,
		Kind:           ProtocolToCoreSymbolKind(sym.Kind),
		Range:          ProtocolToCoreRange(sym.Range, content),
		SelectionRange: ProtocolToCoreRange(sym.SelectionRange, content),
	}

	// Convert detail if present
	if sym.Detail != nil {
		result.Detail = *sym.Detail
	}

	// Convert tags
	if len(sym.Tags) > 0 {
		tags := make([]core.SymbolTag, len(sym.Tags))
		for i, tag := range sym.Tags {
			tags[i] = core.SymbolTag(tag)
		}
		result.Tags = tags
	}

	// Convert deprecated flag
	if sym.Deprecated != nil {
		result.Deprecated = *sym.Deprecated
	}

	// Convert children recursively
	if len(sym.Children) > 0 {
		children := make([]core.DocumentSymbol, len(sym.Children))
		for i, child := range sym.Children {
			children[i] = ProtocolToCoreDocumentSymbol(child, content)
		}
		result.Children = children
	}

	return result
}

// CoreToProtocolDocumentSymbols converts a slice of core document symbols to protocol document symbols.
func CoreToProtocolDocumentSymbols(symbols []core.DocumentSymbol, content string) []protocol.DocumentSymbol {
	result := make([]protocol.DocumentSymbol, len(symbols))
	for i, sym := range symbols {
		result[i] = CoreToProtocolDocumentSymbol(sym, content)
	}
	return result
}

// ProtocolToCoreDocumentSymbols converts a slice of protocol document symbols to core document symbols.
func ProtocolToCoreDocumentSymbols(symbols []protocol.DocumentSymbol, content string) []core.DocumentSymbol {
	result := make([]core.DocumentSymbol, len(symbols))
	for i, sym := range symbols {
		result[i] = ProtocolToCoreDocumentSymbol(sym, content)
	}
	return result
}
