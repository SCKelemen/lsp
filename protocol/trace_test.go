package protocol

import "testing"

func TestSetTraceValueNormalizesInvalidValues(t *testing.T) {
	original := GetTraceValue()
	defer SetTraceValue(original)

	SetTraceValue("invalid")

	if got := GetTraceValue(); got != TraceValueOff {
		t.Fatalf("expected invalid trace value to normalize to off, got %q", got)
	}
	if HasTraceLevel(TraceValueMessage) {
		t.Fatal("expected message trace level to be disabled for normalized off value")
	}
}

func TestSetTraceValueSupportsMessagesAlias(t *testing.T) {
	original := GetTraceValue()
	defer SetTraceValue(original)

	SetTraceValue("messages")

	if got := GetTraceValue(); got != TraceValueMessage {
		t.Fatalf("expected messages alias to map to message, got %q", got)
	}
}

func TestHasTraceMessageTypeUnknownReturnsFalse(t *testing.T) {
	original := GetTraceValue()
	defer SetTraceValue(original)

	SetTraceValue(TraceValueVerbose)

	if HasTraceMessageType(MessageType(99)) {
		t.Fatal("expected unknown message type to return false")
	}
}
