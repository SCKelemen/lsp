package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

// CoreToProtocolCompletionItemKind converts a core completion item kind to protocol.
func CoreToProtocolCompletionItemKind(kind core.CompletionItemKind) protocol.CompletionItemKind {
	return protocol.CompletionItemKind(kind)
}

// ProtocolToCoreCompletionItemKind converts a protocol completion item kind to core.
func ProtocolToCoreCompletionItemKind(kind protocol.CompletionItemKind) core.CompletionItemKind {
	return core.CompletionItemKind(kind)
}

// CoreToProtocolCompletionItem converts a core completion item to protocol.
func CoreToProtocolCompletionItem(item core.CompletionItem, content string) protocol.CompletionItem {
	result := protocol.CompletionItem{
		Label: item.Label,
	}

	// Convert kind
	if item.Kind != nil {
		kind := CoreToProtocolCompletionItemKind(*item.Kind)
		result.Kind = &kind
	}

	// Convert tags
	if len(item.Tags) > 0 {
		tags := make([]protocol.CompletionItemTag, len(item.Tags))
		for i, tag := range item.Tags {
			tags[i] = protocol.CompletionItemTag(tag)
		}
		result.Tags = tags
	}

	// Simple fields
	if item.Detail != "" {
		result.Detail = &item.Detail
	}

	if item.Documentation != "" {
		result.Documentation = &item.Documentation
	}

	if item.Deprecated {
		result.Deprecated = &item.Deprecated
	}

	if item.Preselect {
		result.Preselect = &item.Preselect
	}

	if item.SortText != "" {
		result.SortText = &item.SortText
	}

	if item.FilterText != "" {
		result.FilterText = &item.FilterText
	}

	if item.InsertText != "" {
		result.InsertText = &item.InsertText
	}

	// Convert insert text format
	if item.InsertTextFormat != nil {
		format := protocol.InsertTextFormat(*item.InsertTextFormat)
		result.InsertTextFormat = &format
	}

	// Convert text edit
	if item.TextEdit != nil {
		protocolEdit := CoreToProtocolTextEdit(*item.TextEdit, content)
		result.TextEdit = &protocolEdit
	}

	// Convert additional text edits
	if len(item.AdditionalTextEdits) > 0 {
		result.AdditionalTextEdits = CoreToProtocolTextEdits(item.AdditionalTextEdits, content)
	}

	// Commit characters
	if len(item.CommitCharacters) > 0 {
		result.CommitCharacters = item.CommitCharacters
	}

	// Command
	if item.Command != nil {
		result.Command = coreToProtocolCommand(*item.Command)
	}

	// Data
	result.Data = item.Data

	return result
}

// ProtocolToCoreCompletionItem converts a protocol completion item to core.
func ProtocolToCoreCompletionItem(item protocol.CompletionItem, content string) core.CompletionItem {
	result := core.CompletionItem{
		Label: item.Label,
	}

	// Convert kind
	if item.Kind != nil {
		kind := ProtocolToCoreCompletionItemKind(*item.Kind)
		result.Kind = &kind
	}

	// Convert tags
	if len(item.Tags) > 0 {
		tags := make([]core.CompletionItemTag, len(item.Tags))
		for i, tag := range item.Tags {
			tags[i] = core.CompletionItemTag(tag)
		}
		result.Tags = tags
	}

	// Simple fields
	if item.Detail != nil {
		result.Detail = *item.Detail
	}

	if item.Documentation != nil {
		switch doc := item.Documentation.(type) {
		case string:
			result.Documentation = doc
		}
	}

	if item.Deprecated != nil {
		result.Deprecated = *item.Deprecated
	}

	if item.Preselect != nil {
		result.Preselect = *item.Preselect
	}

	if item.SortText != nil {
		result.SortText = *item.SortText
	}

	if item.FilterText != nil {
		result.FilterText = *item.FilterText
	}

	if item.InsertText != nil {
		result.InsertText = *item.InsertText
	}

	// Convert insert text format
	if item.InsertTextFormat != nil {
		format := core.InsertTextFormat(*item.InsertTextFormat)
		result.InsertTextFormat = &format
	}

	// Convert text edit
	if item.TextEdit != nil {
		switch edit := item.TextEdit.(type) {
		case protocol.TextEdit:
			coreEdit := ProtocolToCoreTextEdit(edit, content)
			result.TextEdit = &coreEdit
		}
	}

	// Convert additional text edits
	if len(item.AdditionalTextEdits) > 0 {
		result.AdditionalTextEdits = ProtocolToCoreTextEdits(item.AdditionalTextEdits, content)
	}

	// Commit characters
	if len(item.CommitCharacters) > 0 {
		result.CommitCharacters = item.CommitCharacters
	}

	// Command
	if item.Command != nil {
		cmd := protocolToCoreCommand(*item.Command)
		result.Command = &cmd
	}

	// Data
	result.Data = item.Data

	return result
}

// CoreToProtocolCompletionList converts a core completion list to protocol.
func CoreToProtocolCompletionList(list *core.CompletionList, content string) *protocol.CompletionList {
	if list == nil {
		return nil
	}

	result := &protocol.CompletionList{
		IsIncomplete: list.IsIncomplete,
		Items:        make([]protocol.CompletionItem, len(list.Items)),
	}

	for i, item := range list.Items {
		result.Items[i] = CoreToProtocolCompletionItem(item, content)
	}

	return result
}

// ProtocolToCoreCompletionList converts a protocol completion list to core.
func ProtocolToCoreCompletionList(list *protocol.CompletionList, content string) *core.CompletionList {
	if list == nil {
		return nil
	}

	result := &core.CompletionList{
		IsIncomplete: list.IsIncomplete,
		Items:        make([]core.CompletionItem, len(list.Items)),
	}

	for i, item := range list.Items {
		result.Items[i] = ProtocolToCoreCompletionItem(item, content)
	}

	return result
}

// Helper functions for Command conversion
func coreToProtocolCommand(cmd core.Command) *protocol.Command {
	result := &protocol.Command{
		Title:     cmd.Title,
		Command:   cmd.Command,
		Arguments: cmd.Arguments,
	}

	// Tooltip (3.18+)
	if cmd.Tooltip != "" {
		result.Tooltip = &cmd.Tooltip
	}

	return result
}

func protocolToCoreCommand(cmd protocol.Command) core.Command {
	result := core.Command{
		Title:     cmd.Title,
		Command:   cmd.Command,
		Arguments: cmd.Arguments,
	}

	// Tooltip (3.18+)
	if cmd.Tooltip != nil {
		result.Tooltip = *cmd.Tooltip
	}

	return result
}

// CoreToProtocolInlineCompletionItem converts a core inline completion item to protocol.
func CoreToProtocolInlineCompletionItem(item core.InlineCompletionItem, content string) protocol.InlineCompletionItem {
	result := protocol.InlineCompletionItem{
		InsertText: item.InsertText,
	}

	// Filter text
	if item.FilterText != "" {
		result.FilterText = &item.FilterText
	}

	// Range
	if item.Range != nil {
		protocolRange := CoreToProtocolRange(*item.Range, content)
		result.Range = &protocolRange
	}

	// Command
	if item.Command != nil {
		result.Command = coreToProtocolCommand(*item.Command)
	}

	return result
}

// ProtocolToCoreInlineCompletionItem converts a protocol inline completion item to core.
func ProtocolToCoreInlineCompletionItem(item protocol.InlineCompletionItem, content string) core.InlineCompletionItem {
	result := core.InlineCompletionItem{
		InsertText: item.InsertText,
	}

	// Filter text
	if item.FilterText != nil {
		result.FilterText = *item.FilterText
	}

	// Range
	if item.Range != nil {
		switch r := item.Range.(type) {
		case protocol.Range:
			coreRange := ProtocolToCoreRange(r, content)
			result.Range = &coreRange
		}
	}

	// Command
	if item.Command != nil {
		cmd := protocolToCoreCommand(*item.Command)
		result.Command = &cmd
	}

	return result
}

// CoreToProtocolInlineCompletionList converts a core inline completion list to protocol.
func CoreToProtocolInlineCompletionList(list *core.InlineCompletionList, content string) *protocol.InlineCompletionList {
	if list == nil {
		return nil
	}

	result := &protocol.InlineCompletionList{
		Items: make([]protocol.InlineCompletionItem, len(list.Items)),
	}

	for i, item := range list.Items {
		result.Items[i] = CoreToProtocolInlineCompletionItem(item, content)
	}

	return result
}

// ProtocolToCoreInlineCompletionList converts a protocol inline completion list to core.
func ProtocolToCoreInlineCompletionList(list *protocol.InlineCompletionList, content string) *core.InlineCompletionList {
	if list == nil {
		return nil
	}

	result := &core.InlineCompletionList{
		Items: make([]core.InlineCompletionItem, len(list.Items)),
	}

	for i, item := range list.Items {
		result.Items[i] = ProtocolToCoreInlineCompletionItem(item, content)
	}

	return result
}
