package core

import "testing"

func TestDiagnosticCodeString(t *testing.T) {
	if got := NewIntCode(123).String(); got != "123" {
		t.Fatalf("expected int diagnostic code to stringify as 123, got %q", got)
	}

	if got := NewStringCode("E_BAD").String(); got != "E_BAD" {
		t.Fatalf("expected string diagnostic code to stringify as E_BAD, got %q", got)
	}
}
