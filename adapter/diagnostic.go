package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol"
)

// ContentResolver returns document content for a URI when converting ranges.
// The boolean indicates whether content was found for the URI.
type ContentResolver func(uri string) (string, bool)

// CoreToProtocolDiagnosticSeverity converts a core diagnostic severity to protocol severity.
func CoreToProtocolDiagnosticSeverity(severity core.DiagnosticSeverity) protocol.DiagnosticSeverity {
	return protocol.DiagnosticSeverity(severity)
}

// ProtocolToCoreDiagnosticSeverity converts a protocol diagnostic severity to core severity.
func ProtocolToCoreDiagnosticSeverity(severity protocol.DiagnosticSeverity) core.DiagnosticSeverity {
	return core.DiagnosticSeverity(severity)
}

// CoreToProtocolDiagnosticTag converts a core diagnostic tag to protocol tag.
func CoreToProtocolDiagnosticTag(tag core.DiagnosticTag) protocol.DiagnosticTag {
	return protocol.DiagnosticTag(tag)
}

// ProtocolToCoreDiagnosticTag converts a protocol diagnostic tag to core tag.
func ProtocolToCoreDiagnosticTag(tag protocol.DiagnosticTag) core.DiagnosticTag {
	return core.DiagnosticTag(tag)
}

// CoreToProtocolDiagnostic converts a core.Diagnostic (UTF-8) to a protocol Diagnostic (UTF-16).
// The content parameter is the document content needed for UTF-8/UTF-16 conversion.
func CoreToProtocolDiagnostic(diag core.Diagnostic, content string) protocol.Diagnostic {
	return CoreToProtocolDiagnosticWithResolver(diag, content, nil)
}

// CoreToProtocolDiagnosticWithResolver converts a core.Diagnostic (UTF-8) to a protocol Diagnostic (UTF-16).
// The content parameter is used for the main diagnostic range.
// If resolver is provided, it is used to map related information ranges in their own files.
func CoreToProtocolDiagnosticWithResolver(diag core.Diagnostic, content string, resolver ContentResolver) protocol.Diagnostic {
	result := protocol.Diagnostic{
		Range:   CoreToProtocolRange(diag.Range, content),
		Message: diag.Message,
		Data:    diag.Data,
	}

	// Convert source
	if diag.Source != "" {
		result.Source = &diag.Source
	}

	// Convert severity
	if diag.Severity != nil {
		severity := CoreToProtocolDiagnosticSeverity(*diag.Severity)
		result.Severity = &severity
	}

	// Convert code
	if diag.Code != nil {
		code := &protocol.IntegerOrString{}
		if diag.Code.IsInt {
			code.Value = diag.Code.IntValue
		} else {
			code.Value = diag.Code.StringValue
		}
		result.Code = code
	}

	// Convert code description
	if diag.CodeDescription != nil {
		result.CodeDescription = &protocol.CodeDescription{
			HRef: protocol.URI(diag.CodeDescription.HRef),
		}
	}

	// Convert tags
	if len(diag.Tags) > 0 {
		tags := make([]protocol.DiagnosticTag, len(diag.Tags))
		for i, tag := range diag.Tags {
			tags[i] = CoreToProtocolDiagnosticTag(tag)
		}
		result.Tags = tags
	}

	// Convert related information
	if len(diag.RelatedInformation) > 0 {
		relatedInfo := make([]protocol.DiagnosticRelatedInformation, len(diag.RelatedInformation))
		for i, info := range diag.RelatedInformation {
			relatedContent := content
			if resolver != nil {
				if resolvedContent, ok := resolver(info.Location.URI); ok {
					relatedContent = resolvedContent
				}
			}

			relatedInfo[i] = protocol.DiagnosticRelatedInformation{
				Location: CoreToProtocolLocation(info.Location, relatedContent),
				Message:  info.Message,
			}
		}
		result.RelatedInformation = relatedInfo
	}

	return result
}

// ProtocolToCoreDiagnostic converts a protocol Diagnostic (UTF-16) to a core.Diagnostic (UTF-8).
// The content parameter is the document content needed for UTF-16/UTF-8 conversion.
func ProtocolToCoreDiagnostic(diag protocol.Diagnostic, content string) core.Diagnostic {
	return ProtocolToCoreDiagnosticWithResolver(diag, content, nil)
}

// ProtocolToCoreDiagnosticWithResolver converts a protocol Diagnostic (UTF-16) to a core.Diagnostic (UTF-8).
// The content parameter is used for the main diagnostic range.
// If resolver is provided, it is used to map related information ranges in their own files.
func ProtocolToCoreDiagnosticWithResolver(diag protocol.Diagnostic, content string, resolver ContentResolver) core.Diagnostic {
	result := core.Diagnostic{
		Range:   ProtocolToCoreRange(diag.Range, content),
		Message: diag.Message,
		Data:    diag.Data,
	}

	// Convert source
	if diag.Source != nil {
		result.Source = *diag.Source
	}

	// Convert severity
	if diag.Severity != nil {
		severity := ProtocolToCoreDiagnosticSeverity(*diag.Severity)
		result.Severity = &severity
	}

	// Convert code
	if diag.Code != nil && diag.Code.Value != nil {
		switch code := diag.Code.Value.(type) {
		case protocol.Integer:
			result.Code = &core.DiagnosticCode{IsInt: true, IntValue: int(code)}
		case int:
			result.Code = &core.DiagnosticCode{IsInt: true, IntValue: code}
		case string:
			result.Code = &core.DiagnosticCode{IsInt: false, StringValue: code}
		}
	}

	// Convert code description
	if diag.CodeDescription != nil {
		result.CodeDescription = &core.CodeDescription{
			HRef: string(diag.CodeDescription.HRef),
		}
	}

	// Convert tags
	if len(diag.Tags) > 0 {
		tags := make([]core.DiagnosticTag, len(diag.Tags))
		for i, tag := range diag.Tags {
			tags[i] = ProtocolToCoreDiagnosticTag(tag)
		}
		result.Tags = tags
	}

	// Convert related information
	if len(diag.RelatedInformation) > 0 {
		relatedInfo := make([]core.DiagnosticRelatedInformation, len(diag.RelatedInformation))
		for i, info := range diag.RelatedInformation {
			relatedContent := content
			if resolver != nil {
				if resolvedContent, ok := resolver(string(info.Location.URI)); ok {
					relatedContent = resolvedContent
				}
			}

			relatedInfo[i] = core.DiagnosticRelatedInformation{
				Location: ProtocolToCoreLocation(info.Location, relatedContent),
				Message:  info.Message,
			}
		}
		result.RelatedInformation = relatedInfo
	}

	return result
}

// CoreToProtocolDiagnostics converts a slice of core diagnostics to protocol diagnostics.
// This is a convenience function for the common case of converting diagnostic arrays.
func CoreToProtocolDiagnostics(diags []core.Diagnostic, content string) []protocol.Diagnostic {
	return CoreToProtocolDiagnosticsWithResolver(diags, content, nil)
}

// CoreToProtocolDiagnosticsWithResolver converts a slice of core diagnostics to protocol diagnostics.
// The resolver is used for related information locations in other files.
func CoreToProtocolDiagnosticsWithResolver(diags []core.Diagnostic, content string, resolver ContentResolver) []protocol.Diagnostic {
	result := make([]protocol.Diagnostic, len(diags))
	for i, diag := range diags {
		result[i] = CoreToProtocolDiagnosticWithResolver(diag, content, resolver)
	}
	return result
}

// ProtocolToCoreDiagnostics converts a slice of protocol diagnostics to core diagnostics.
func ProtocolToCoreDiagnostics(diags []protocol.Diagnostic, content string) []core.Diagnostic {
	return ProtocolToCoreDiagnosticsWithResolver(diags, content, nil)
}

// ProtocolToCoreDiagnosticsWithResolver converts a slice of protocol diagnostics to core diagnostics.
// The resolver is used for related information locations in other files.
func ProtocolToCoreDiagnosticsWithResolver(diags []protocol.Diagnostic, content string, resolver ContentResolver) []core.Diagnostic {
	result := make([]core.Diagnostic, len(diags))
	for i, diag := range diags {
		result[i] = ProtocolToCoreDiagnosticWithResolver(diag, content, resolver)
	}
	return result
}
