package adapter_3_16

import (
	"github.com/SCKelemen/lsp/core"
	protocol "github.com/SCKelemen/lsp/protocol_3_16"
)

// CoreToProtocolFormattingOptions converts core formatting options to protocol formatting options.
// Protocol uses a map[string]any while core uses a struct for type safety.
func CoreToProtocolFormattingOptions(opts core.FormattingOptions) protocol.FormattingOptions {
	result := make(protocol.FormattingOptions)

	result["tabSize"] = opts.TabSize
	result["insertSpaces"] = opts.InsertSpaces

	if opts.TrimTrailingWhitespace {
		result["trimTrailingWhitespace"] = opts.TrimTrailingWhitespace
	}

	if opts.InsertFinalNewline {
		result["insertFinalNewline"] = opts.InsertFinalNewline
	}

	if opts.TrimFinalNewlines {
		result["trimFinalNewlines"] = opts.TrimFinalNewlines
	}

	return result
}

// ProtocolToCoreFormattingOptions converts protocol formatting options to core formatting options.
// Provides sensible defaults for missing values.
func ProtocolToCoreFormattingOptions(opts protocol.FormattingOptions) core.FormattingOptions {
	result := core.FormattingOptions{
		TabSize:      4,    // Default tab size
		InsertSpaces: true, // Default to spaces
	}

	// Extract tabSize
	if tabSize, ok := opts["tabSize"]; ok {
		switch v := tabSize.(type) {
		case int:
			result.TabSize = v
		case float64:
			result.TabSize = int(v)
		}
	}

	// Extract insertSpaces
	if insertSpaces, ok := opts["insertSpaces"]; ok {
		if v, ok := insertSpaces.(bool); ok {
			result.InsertSpaces = v
		}
	}

	// Extract trimTrailingWhitespace
	if trim, ok := opts["trimTrailingWhitespace"]; ok {
		if v, ok := trim.(bool); ok {
			result.TrimTrailingWhitespace = v
		}
	}

	// Extract insertFinalNewline
	if insert, ok := opts["insertFinalNewline"]; ok {
		if v, ok := insert.(bool); ok {
			result.InsertFinalNewline = v
		}
	}

	// Extract trimFinalNewlines
	if trim, ok := opts["trimFinalNewlines"]; ok {
		if v, ok := trim.(bool); ok {
			result.TrimFinalNewlines = v
		}
	}

	return result
}
