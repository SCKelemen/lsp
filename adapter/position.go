// Package adapter_3_16 provides conversion functions between core types and protocol_3_16 types.
// Core types use UTF-8 byte offsets for natural Go string handling, while protocol types
// use UTF-16 code unit offsets as specified by the LSP specification.
//
// LSP 3.16 Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/
//
// These adapters handle the critical conversion at API boundaries:
// - Convert core types (UTF-8 offsets) to protocol types (UTF-16 code units) when sending to clients
// - Convert protocol types (UTF-16 code units) to core types (UTF-8 offsets) when receiving from clients
package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

// CoreToProtocolPosition converts a core.Position (UTF-8) to a protocol Position (UTF-16).
// The content parameter is needed to perform the UTF-8 to UTF-16 offset conversion.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position
// Protocol requires UTF-16 code units for the character offset.
func CoreToProtocolPosition(pos core.Position, content string) protocol.Position {
	utf16Offset := core.UTF8ToUTF16Offset(content, pos.Line, pos.Character)
	return protocol.Position{
		Line:      protocol.UInteger(pos.Line),
		Character: protocol.UInteger(utf16Offset),
	}
}

// ProtocolToCorePosition converts a protocol Position (UTF-16) to a core.Position (UTF-8).
// The content parameter is needed to perform the UTF-16 to UTF-8 offset conversion.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position
// Protocol provides UTF-16 code units which must be converted to UTF-8 byte offsets.
func ProtocolToCorePosition(pos protocol.Position, content string) core.Position {
	utf8Offset := core.UTF16ToUTF8Offset(content, int(pos.Line), int(pos.Character))
	return core.Position{
		Line:      int(pos.Line),
		Character: utf8Offset,
	}
}

// CoreToProtocolRange converts a core.Range (UTF-8) to a protocol Range (UTF-16).
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#range
func CoreToProtocolRange(r core.Range, content string) protocol.Range {
	return protocol.Range{
		Start: CoreToProtocolPosition(r.Start, content),
		End:   CoreToProtocolPosition(r.End, content),
	}
}

// ProtocolToCoreRange converts a protocol Range (UTF-16) to a core.Range (UTF-8).
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#range
func ProtocolToCoreRange(r protocol.Range, content string) core.Range {
	return core.Range{
		Start: ProtocolToCorePosition(r.Start, content),
		End:   ProtocolToCorePosition(r.End, content),
	}
}

// CoreToProtocolLocation converts a core.Location (UTF-8) to a protocol Location (UTF-16).
// The content parameter should be the content of the document at the location's URI.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#location
func CoreToProtocolLocation(loc core.Location, content string) protocol.Location {
	return protocol.Location{
		URI:   protocol.DocumentUri(loc.URI),
		Range: CoreToProtocolRange(loc.Range, content),
	}
}

// ProtocolToCoreLocation converts a protocol Location (UTF-16) to a core.Location (UTF-8).
// The content parameter should be the content of the document at the location's URI.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#location
func ProtocolToCoreLocation(loc protocol.Location, content string) core.Location {
	return core.Location{
		URI:   string(loc.URI),
		Range: ProtocolToCoreRange(loc.Range, content),
	}
}

// CoreToProtocolLocationLink converts a core.LocationLink (UTF-8) to a protocol LocationLink (UTF-16).
// The originContent parameter is the content of the origin document (if OriginSelectionRange is set).
// The targetContent parameter is the content of the target document.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#locationLink
func CoreToProtocolLocationLink(link core.LocationLink, originContent, targetContent string) protocol.LocationLink {
	result := protocol.LocationLink{
		TargetURI:            protocol.DocumentUri(link.TargetURI),
		TargetRange:          CoreToProtocolRange(link.TargetRange, targetContent),
		TargetSelectionRange: CoreToProtocolRange(link.TargetSelectionRange, targetContent),
	}

	if link.OriginSelectionRange != nil {
		originRange := CoreToProtocolRange(*link.OriginSelectionRange, originContent)
		result.OriginSelectionRange = &originRange
	}

	return result
}

// ProtocolToCoreLocationLink converts a protocol LocationLink (UTF-16) to a core.LocationLink (UTF-8).
// The originContent parameter is the content of the origin document (if OriginSelectionRange is set).
// The targetContent parameter is the content of the target document.
//
// LSP Spec: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#locationLink
func ProtocolToCoreLocationLink(link protocol.LocationLink, originContent, targetContent string) core.LocationLink {
	result := core.LocationLink{
		TargetURI:            string(link.TargetURI),
		TargetRange:          ProtocolToCoreRange(link.TargetRange, targetContent),
		TargetSelectionRange: ProtocolToCoreRange(link.TargetSelectionRange, targetContent),
	}

	if link.OriginSelectionRange != nil {
		originRange := ProtocolToCoreRange(*link.OriginSelectionRange, originContent)
		result.OriginSelectionRange = &originRange
	}

	return result
}
