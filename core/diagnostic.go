package core

// DiagnosticSeverity indicates the severity level of a diagnostic.
type DiagnosticSeverity int

const (
	// SeverityError indicates an error
	SeverityError DiagnosticSeverity = 1
	// SeverityWarning indicates a warning
	SeverityWarning DiagnosticSeverity = 2
	// SeverityInformation indicates informational message
	SeverityInformation DiagnosticSeverity = 3
	// SeverityHint indicates a hint
	SeverityHint DiagnosticSeverity = 4
)

// String returns a human-readable name for the severity
func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInformation:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

// DiagnosticTag provides additional metadata about a diagnostic.
type DiagnosticTag int

const (
	// TagUnnecessary indicates unused or unnecessary code that can be grayed out
	TagUnnecessary DiagnosticTag = 1
	// TagDeprecated indicates deprecated code that should be struck through
	TagDeprecated DiagnosticTag = 2
)

// String returns a human-readable name for the tag
func (t DiagnosticTag) String() string {
	switch t {
	case TagUnnecessary:
		return "unnecessary"
	case TagDeprecated:
		return "deprecated"
	default:
		return "unknown"
	}
}

// CodeDescription provides additional information about a diagnostic code,
// typically a URL to documentation.
type CodeDescription struct {
	// HRef is a URI pointing to documentation or more information
	HRef string
}

// DiagnosticRelatedInformation represents a related location and message for a diagnostic.
// This is used to show related code locations that contribute to or are affected by the diagnostic.
type DiagnosticRelatedInformation struct {
	// Location is where the related information is
	Location Location
	// Message describes the relationship
	Message string
}

// DiagnosticCode represents a diagnostic code which can be either an integer or string.
type DiagnosticCode struct {
	// IsInt indicates whether this is an integer code
	IsInt bool
	// IntValue holds the integer code if IsInt is true
	IntValue int
	// StringValue holds the string code if IsInt is false
	StringValue string
}

// NewIntCode creates a diagnostic code from an integer
func NewIntCode(code int) DiagnosticCode {
	return DiagnosticCode{IsInt: true, IntValue: code}
}

// NewStringCode creates a diagnostic code from a string
func NewStringCode(code string) DiagnosticCode {
	return DiagnosticCode{IsInt: false, StringValue: code}
}

// String returns the string representation of the code
func (c DiagnosticCode) String() string {
	if c.IsInt {
		return string(rune(c.IntValue))
	}
	return c.StringValue
}

// Diagnostic represents a diagnostic message such as a compiler error or warning.
// Diagnostics are the primary way to report problems found during analysis.
type Diagnostic struct {
	// Range is where the diagnostic applies
	Range Range

	// Severity indicates the diagnostic level (error, warning, info, hint)
	// If nil, the client decides how to interpret it
	Severity *DiagnosticSeverity

	// Code is an optional diagnostic code (can be int or string)
	Code *DiagnosticCode

	// CodeDescription provides a link to more information about the code
	CodeDescription *CodeDescription

	// Source is the name of the source of the diagnostic (e.g., "typescript", "eslint")
	Source string

	// Message is the human-readable diagnostic message
	Message string

	// Tags provide additional metadata (unnecessary, deprecated)
	Tags []DiagnosticTag

	// RelatedInformation provides additional locations related to this diagnostic
	RelatedInformation []DiagnosticRelatedInformation

	// Data is arbitrary data preserved between diagnostics and code actions
	// This can be used to carry information needed for quick fixes
	Data any
}

// IsError returns true if this diagnostic is an error
func (d Diagnostic) IsError() bool {
	return d.Severity != nil && *d.Severity == SeverityError
}

// IsWarning returns true if this diagnostic is a warning
func (d Diagnostic) IsWarning() bool {
	return d.Severity != nil && *d.Severity == SeverityWarning
}

// IsInformation returns true if this diagnostic is informational
func (d Diagnostic) IsInformation() bool {
	return d.Severity != nil && *d.Severity == SeverityInformation
}

// IsHint returns true if this diagnostic is a hint
func (d Diagnostic) IsHint() bool {
	return d.Severity != nil && *d.Severity == SeverityHint
}

// HasTag returns true if this diagnostic has the specified tag
func (d Diagnostic) HasTag(tag DiagnosticTag) bool {
	for _, t := range d.Tags {
		if t == tag {
			return true
		}
	}
	return false
}
