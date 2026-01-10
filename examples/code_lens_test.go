package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestTestRunnerCodeLensProvider tests test runner code lenses.
func TestTestRunnerCodeLensProvider(t *testing.T) {
	provider := &TestRunnerCodeLensProvider{}

	tests := []struct {
		name          string
		uri           string
		content       string
		wantLensCount int
		checkLenses   func(t *testing.T, lenses []core.CodeLens)
	}{
		{
			name: "single test function",
			uri:  "file:///example_test.go",
			content: `package example

import "testing"

func TestSum(t *testing.T) {
	// test code
}`,
			wantLensCount: 2, // Run + Debug
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				runFound := false
				debugFound := false
				for _, lens := range lenses {
					if lens.Command == nil {
						t.Error("expected command to be set")
						continue
					}
					if strings.Contains(lens.Command.Title, "Run TestSum") {
						runFound = true
						if lens.Command.Command != "go.test.run" {
							t.Errorf("expected command 'go.test.run', got %q", lens.Command.Command)
						}
					}
					if strings.Contains(lens.Command.Title, "Debug TestSum") {
						debugFound = true
						if lens.Command.Command != "go.test.debug" {
							t.Errorf("expected command 'go.test.debug', got %q", lens.Command.Command)
						}
					}
				}
				if !runFound {
					t.Error("expected to find 'Run TestSum' lens")
				}
				if !debugFound {
					t.Error("expected to find 'Debug TestSum' lens")
				}
			},
		},
		{
			name: "multiple test functions",
			uri:  "file:///example_test.go",
			content: `package example

import "testing"

func TestFoo(t *testing.T) {}
func TestBar(t *testing.T) {}
func TestBaz(t *testing.T) {}`,
			wantLensCount: 6, // 3 tests * 2 lenses each
		},
		{
			name: "non-test file returns nil",
			uri:  "file:///example.go",
			content: `package example

func TestLikeFunction() {}`,
			wantLensCount: 0,
		},
		{
			name: "test file without test functions",
			uri:  "file:///example_test.go",
			content: `package example

func Helper() {}`,
			wantLensCount: 0,
		},
		{
			name: "benchmark function not included",
			uri:  "file:///example_test.go",
			content: `package example

import "testing"

func BenchmarkSum(b *testing.B) {}`,
			wantLensCount: 0, // Only Test* functions, not Benchmark*
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CodeLensContext{
				URI:     tt.uri,
				Content: tt.content,
			}

			lenses := provider.ProvideCodeLenses(ctx)

			if len(lenses) != tt.wantLensCount {
				t.Errorf("got %d lenses, want %d", len(lenses), tt.wantLensCount)
			}

			if tt.checkLenses != nil {
				tt.checkLenses(t, lenses)
			}
		})
	}
}

// TestReferenceCountCodeLensProvider tests reference count code lenses.
func TestReferenceCountCodeLensProvider(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		counterFunc   func(uri, symbolName string) int
		wantLensCount int
		checkLenses   func(t *testing.T, lenses []core.CodeLens)
	}{
		{
			name: "exported function with references",
			content: `package example

func CalculateSum(a, b int) int {
	return a + b
}

func Example() {
	result := CalculateSum(1, 2)
	println(result)
}`,
			wantLensCount: 2, // CalculateSum + Example
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				for _, lens := range lenses {
					if lens.Command == nil {
						t.Error("expected command to be set")
						continue
					}
					if !strings.Contains(lens.Command.Title, "references") {
						t.Errorf("expected title to contain 'references', got %q", lens.Command.Title)
					}
					if lens.Command.Command != "editor.action.showReferences" {
						t.Errorf("expected command 'editor.action.showReferences', got %q", lens.Command.Command)
					}
				}
			},
		},
		{
			name: "unexported function excluded",
			content: `package example

func calculateSum(a, b int) int {
	return a + b
}`,
			wantLensCount: 0, // Unexported, so excluded
		},
		{
			name: "exported type with references",
			content: `package example

type Config struct {
	Value int
}

func NewConfig() *Config {
	return &Config{Value: 42}
}`,
			wantLensCount: 2, // Config type + NewConfig function
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				configFound := false
				for _, lens := range lenses {
					if lens.Command != nil && strings.Contains(lens.Command.Title, "references") {
						if len(lens.Command.Arguments) > 1 {
							if name, ok := lens.Command.Arguments[1].(string); ok && name == "Config" {
								configFound = true
							}
						}
					}
				}
				if !configFound {
					t.Error("expected to find code lens for Config type")
				}
			},
		},
		{
			name: "custom reference counter",
			content: `package example

func Calculate() int {
	return 42
}`,
			counterFunc: func(uri, symbolName string) int {
				if symbolName == "Calculate" {
					return 100
				}
				return 0
			},
			wantLensCount: 1,
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				if len(lenses) > 0 && lenses[0].Command != nil {
					if !strings.Contains(lenses[0].Command.Title, "100 references") {
						t.Errorf("expected '100 references', got %q", lenses[0].Command.Title)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &ReferenceCountCodeLensProvider{
				ReferenceCounter: tt.counterFunc,
			}

			ctx := core.CodeLensContext{
				URI:     "file:///example.go",
				Content: tt.content,
			}

			lenses := provider.ProvideCodeLenses(ctx)

			if len(lenses) != tt.wantLensCount {
				t.Errorf("got %d lenses, want %d", len(lenses), tt.wantLensCount)
			}

			if tt.checkLenses != nil {
				tt.checkLenses(t, lenses)
			}
		})
	}
}

// TestTODOCodeLensProvider tests TODO comment code lenses.
func TestTODOCodeLensProvider(t *testing.T) {
	provider := &TODOCodeLensProvider{}

	tests := []struct {
		name          string
		content       string
		wantLensCount int
		checkLenses   func(t *testing.T, lenses []core.CodeLens)
	}{
		{
			name: "simple TODO",
			content: `package example

// TODO: Implement this function
func DoSomething() {}`,
			wantLensCount: 1,
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				if lenses[0].Command == nil {
					t.Fatal("expected command to be set")
				}
				if !strings.Contains(lenses[0].Command.Title, "TODO") {
					t.Errorf("expected title to contain 'TODO', got %q", lenses[0].Command.Title)
				}
				if !strings.Contains(lenses[0].Command.Title, "Implement this function") {
					t.Errorf("expected message in title, got %q", lenses[0].Command.Title)
				}
			},
		},
		{
			name: "TODO with author",
			content: `package example

// TODO(alice): Add error handling here
func Process() {}`,
			wantLensCount: 1,
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				if lenses[0].Command == nil {
					t.Fatal("expected command to be set")
				}
				if !strings.Contains(lenses[0].Command.Title, "alice") {
					t.Errorf("expected author 'alice' in title, got %q", lenses[0].Command.Title)
				}
			},
		},
		{
			name: "multiple TODO types",
			content: `package example

// TODO: First task
func Foo() {}

// FIXME: Broken logic
func Bar() {}

// HACK: Temporary workaround
func Baz() {}

// XXX: Review this
func Qux() {}`,
			wantLensCount: 4,
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				types := make(map[string]bool)
				for _, lens := range lenses {
					if lens.Command != nil {
						title := lens.Command.Title
						if strings.Contains(title, "TODO") {
							types["TODO"] = true
						}
						if strings.Contains(title, "FIXME") {
							types["FIXME"] = true
						}
						if strings.Contains(title, "HACK") {
							types["HACK"] = true
						}
						if strings.Contains(title, "XXX") {
							types["XXX"] = true
						}
					}
				}
				expectedTypes := []string{"TODO", "FIXME", "HACK", "XXX"}
				for _, expected := range expectedTypes {
					if !types[expected] {
						t.Errorf("expected to find %s lens", expected)
					}
				}
			},
		},
		{
			name: "no TODOs",
			content: `package example

// Regular comment
func Normal() {}`,
			wantLensCount: 0,
		},
		{
			name: "long TODO message truncated",
			content: `package example

// TODO: This is a very long TODO message that should be truncated because it exceeds the maximum length
func LongTODO() {}`,
			wantLensCount: 1,
			checkLenses: func(t *testing.T, lenses []core.CodeLens) {
				if lenses[0].Command == nil {
					t.Fatal("expected command to be set")
				}
				// Check for truncation (...)
				if !strings.Contains(lenses[0].Command.Title, "...") {
					t.Log("Expected truncation in:", lenses[0].Command.Title)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CodeLensContext{
				URI:     "file:///example.go",
				Content: tt.content,
			}

			lenses := provider.ProvideCodeLenses(ctx)

			if len(lenses) != tt.wantLensCount {
				t.Errorf("got %d lenses, want %d", len(lenses), tt.wantLensCount)
			}

			if tt.checkLenses != nil {
				tt.checkLenses(t, lenses)
			}
		})
	}
}

// TestLazyCodeLensProvider tests lazy code lens resolution.
func TestLazyCodeLensProvider(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantLensCount int
		resolveFunc   func(lens core.CodeLens) core.CodeLens
		checkResolved func(t *testing.T, resolved core.CodeLens)
	}{
		{
			name: "unresolved lenses for exported functions",
			content: `package example

func PublicFunction() {}

func privateFunction() {}`,
			wantLensCount: 1, // Only exported function
			checkResolved: func(t *testing.T, resolved core.CodeLens) {
				if resolved.Command == nil {
					t.Error("expected command after resolution")
				}
			},
		},
		{
			name: "custom resolve function",
			content: `package example

func Calculate() int {
	return 42
}`,
			wantLensCount: 1,
			resolveFunc: func(lens core.CodeLens) core.CodeLens {
				lens.Command = &core.Command{
					Title:     "Custom Command",
					Command:   "custom.action",
					Arguments: []interface{}{"custom-arg"},
				}
				return lens
			},
			checkResolved: func(t *testing.T, resolved core.CodeLens) {
				if resolved.Command == nil {
					t.Fatal("expected command after resolution")
				}
				if resolved.Command.Title != "Custom Command" {
					t.Errorf("expected 'Custom Command', got %q", resolved.Command.Title)
				}
				if resolved.Command.Command != "custom.action" {
					t.Errorf("expected 'custom.action', got %q", resolved.Command.Command)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &LazyCodeLensProvider{
				ResolveFunc: tt.resolveFunc,
			}

			ctx := core.CodeLensContext{
				URI:     "file:///example.go",
				Content: tt.content,
			}

			lenses := provider.ProvideCodeLenses(ctx)

			if len(lenses) != tt.wantLensCount {
				t.Errorf("got %d lenses, want %d", len(lenses), tt.wantLensCount)
			}

			// Check that initial lenses have no command
			for i, lens := range lenses {
				if lens.Command != nil {
					t.Errorf("lens %d: expected nil command before resolution", i)
				}
				if lens.Data == nil {
					t.Errorf("lens %d: expected data to be set", i)
				}
			}

			// Resolve the first lens
			if len(lenses) > 0 {
				resolved := provider.ResolveCodeLens(lenses[0])
				if tt.checkResolved != nil {
					tt.checkResolved(t, resolved)
				}
			}
		})
	}
}

// TestCompositeCodeLensProvider tests combining multiple providers.
func TestCompositeCodeLensProvider(t *testing.T) {
	content := `package example

import "testing"

// TODO: Add more test cases
func TestExample(t *testing.T) {}

func PublicFunction() {}
`

	provider := NewCompositeCodeLensProvider(
		&TestRunnerCodeLensProvider{},
		&TODOCodeLensProvider{},
		&ReferenceCountCodeLensProvider{},
	)

	ctx := core.CodeLensContext{
		URI:     "file:///example_test.go",
		Content: content,
	}

	lenses := provider.ProvideCodeLenses(ctx)

	// Should have:
	// - 2 from TestRunnerCodeLensProvider (Run + Debug for TestExample)
	// - 1 from TODOCodeLensProvider (the TODO comment)
	// - 1 from ReferenceCountCodeLensProvider (PublicFunction)
	expectedMin := 4

	if len(lenses) < expectedMin {
		t.Errorf("got %d lenses, want at least %d", len(lenses), expectedMin)
	}

	// Check that we have lenses from different providers
	hasTestLens := false
	hasTodoLens := false
	hasRefLens := false

	for _, lens := range lenses {
		if lens.Command != nil {
			if strings.Contains(lens.Command.Title, "Run") || strings.Contains(lens.Command.Title, "Debug") {
				hasTestLens = true
			}
			if strings.Contains(lens.Command.Title, "TODO") {
				hasTodoLens = true
			}
			if strings.Contains(lens.Command.Title, "references") {
				hasRefLens = true
			}
		}
	}

	if !hasTestLens {
		t.Error("expected test runner lens")
	}
	if !hasTodoLens {
		t.Error("expected TODO lens")
	}
	if !hasRefLens {
		t.Error("expected reference count lens")
	}
}

// TestCodeLens_EdgeCases tests edge cases for code lenses.
func TestCodeLens_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		provider core.CodeLensProvider
		content  string
		test     func(t *testing.T, lenses []core.CodeLens)
	}{
		{
			name:     "empty file",
			provider: &TestRunnerCodeLensProvider{},
			content:  "",
			test: func(t *testing.T, lenses []core.CodeLens) {
				if len(lenses) != 0 {
					t.Error("expected no lenses for empty file")
				}
			},
		},
		{
			name:     "invalid Go syntax",
			provider: &TestRunnerCodeLensProvider{},
			content:  "this is not valid Go code",
			test: func(t *testing.T, lenses []core.CodeLens) {
				if len(lenses) != 0 {
					t.Error("expected no lenses for invalid syntax")
				}
			},
		},
		{
			name:     "range accuracy",
			provider: &TestRunnerCodeLensProvider{},
			content: `package example

import "testing"

func TestAccuracy(t *testing.T) {}`,
			test: func(t *testing.T, lenses []core.CodeLens) {
				if len(lenses) == 0 {
					t.Fatal("expected at least one lens")
				}

				// Check that range is valid
				lens := lenses[0]
				if lens.Range.Start.Line < 0 || lens.Range.Start.Character < 0 {
					t.Error("range has negative position")
				}
				if lens.Range.End.Line < lens.Range.Start.Line {
					t.Error("range end is before start")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CodeLensContext{
				URI:     "file:///test_test.go",
				Content: tt.content,
			}

			lenses := tt.provider.ProvideCodeLenses(ctx)
			tt.test(t, lenses)
		})
	}
}
