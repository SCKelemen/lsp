// Package adapter_3_17 provides conversion functions between core types and protocol_3_17 types.
// Since protocol_3_17 extends protocol_3_16 and reuses most base types (Position, Range, Diagnostic, etc.),
// this adapter primarily delegates to adapter_3_16 for base type conversions and adds any
// 3.17-specific conversions as needed.
package adapter_3_17

import (
	"github.com/SCKelemen/lsp/adapter_3_16"
	"github.com/SCKelemen/lsp/core"
	protocol316 "github.com/SCKelemen/lsp/protocol_3_16"
)

// Protocol 3.17 reuses protocol 3.16 base types, so we can delegate to adapter_3_16
// for all base type conversions.

// CoreToProtocolPosition converts a core.Position (UTF-8) to a protocol Position (UTF-16).
// Since protocol_3_17 uses protocol_3_16.Position, we delegate to adapter_3_16.
func CoreToProtocolPosition(pos core.Position, content string) protocol316.Position {
	return adapter_3_16.CoreToProtocolPosition(pos, content)
}

// ProtocolToCorePosition converts a protocol Position (UTF-16) to a core.Position (UTF-8).
func ProtocolToCorePosition(pos protocol316.Position, content string) core.Position {
	return adapter_3_16.ProtocolToCorePosition(pos, content)
}

// CoreToProtocolRange converts a core.Range (UTF-8) to a protocol Range (UTF-16).
func CoreToProtocolRange(r core.Range, content string) protocol316.Range {
	return adapter_3_16.CoreToProtocolRange(r, content)
}

// ProtocolToCoreRange converts a protocol Range (UTF-16) to a core.Range (UTF-8).
func ProtocolToCoreRange(r protocol316.Range, content string) core.Range {
	return adapter_3_16.ProtocolToCoreRange(r, content)
}

// CoreToProtocolLocation converts a core.Location (UTF-8) to a protocol Location (UTF-16).
func CoreToProtocolLocation(loc core.Location, content string) protocol316.Location {
	return adapter_3_16.CoreToProtocolLocation(loc, content)
}

// ProtocolToCoreLocation converts a protocol Location (UTF-16) to a core.Location (UTF-8).
func ProtocolToCoreLocation(loc protocol316.Location, content string) core.Location {
	return adapter_3_16.ProtocolToCoreLocation(loc, content)
}

// CoreToProtocolLocationLink converts a core.LocationLink (UTF-8) to a protocol LocationLink (UTF-16).
func CoreToProtocolLocationLink(link core.LocationLink, originContent, targetContent string) protocol316.LocationLink {
	return adapter_3_16.CoreToProtocolLocationLink(link, originContent, targetContent)
}

// ProtocolToCoreLocationLink converts a protocol LocationLink (UTF-16) to a core.LocationLink (UTF-8).
func ProtocolToCoreLocationLink(link protocol316.LocationLink, originContent, targetContent string) core.LocationLink {
	return adapter_3_16.ProtocolToCoreLocationLink(link, originContent, targetContent)
}

// CoreToProtocolDiagnosticSeverity converts a core diagnostic severity to protocol severity.
func CoreToProtocolDiagnosticSeverity(severity core.DiagnosticSeverity) protocol316.DiagnosticSeverity {
	return adapter_3_16.CoreToProtocolDiagnosticSeverity(severity)
}

// ProtocolToCoreDiagnosticSeverity converts a protocol diagnostic severity to core severity.
func ProtocolToCoreDiagnosticSeverity(severity protocol316.DiagnosticSeverity) core.DiagnosticSeverity {
	return adapter_3_16.ProtocolToCoreDiagnosticSeverity(severity)
}

// CoreToProtocolDiagnosticTag converts a core diagnostic tag to protocol tag.
func CoreToProtocolDiagnosticTag(tag core.DiagnosticTag) protocol316.DiagnosticTag {
	return adapter_3_16.CoreToProtocolDiagnosticTag(tag)
}

// ProtocolToCoreDiagnosticTag converts a protocol diagnostic tag to core tag.
func ProtocolToCoreDiagnosticTag(tag protocol316.DiagnosticTag) core.DiagnosticTag {
	return adapter_3_16.ProtocolToCoreDiagnosticTag(tag)
}

// CoreToProtocolDiagnostic converts a core.Diagnostic (UTF-8) to a protocol Diagnostic (UTF-16).
func CoreToProtocolDiagnostic(diag core.Diagnostic, content string) protocol316.Diagnostic {
	return adapter_3_16.CoreToProtocolDiagnostic(diag, content)
}

// ProtocolToCoreDiagnostic converts a protocol Diagnostic (UTF-16) to a core.Diagnostic (UTF-8).
func ProtocolToCoreDiagnostic(diag protocol316.Diagnostic, content string) core.Diagnostic {
	return adapter_3_16.ProtocolToCoreDiagnostic(diag, content)
}

// CoreToProtocolDiagnostics converts a slice of core diagnostics to protocol diagnostics.
// This is a convenience function for the common case of converting diagnostic arrays.
func CoreToProtocolDiagnostics(diags []core.Diagnostic, content string) []protocol316.Diagnostic {
	result := make([]protocol316.Diagnostic, len(diags))
	for i, diag := range diags {
		result[i] = CoreToProtocolDiagnostic(diag, content)
	}
	return result
}

// ProtocolToCoreDiagnostics converts a slice of protocol diagnostics to core diagnostics.
func ProtocolToCoreDiagnostics(diags []protocol316.Diagnostic, content string) []core.Diagnostic {
	result := make([]core.Diagnostic, len(diags))
	for i, diag := range diags {
		result[i] = ProtocolToCoreDiagnostic(diag, content)
	}
	return result
}
