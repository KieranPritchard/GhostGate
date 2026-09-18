package util

import (
	"embed"
	"strings"
	"testing"
)

// Fixture filesystems used to exercise loadUserAgents/GetRandomHeader without
// touching the real embedded resources/user_agents.txt file. Each mirrors the
// "resources/user_agents.txt" path that loadUserAgents opens.
//
//go:embed resources/*
var validAgentsFS embed.FS

//go:embed testdata/*
var emptyAgentsFS embed.FS

//go:embed testdata/missing
var missingAgentsFS embed.FS

// withAgentsFS temporarily swaps the package-level userAgentFiles variable,
// restoring the original (real) embedded filesystem once the test finishes.
func withAgentsFS(t *testing.T, fs embed.FS) {
	t.Helper()
	original := userAgentFiles
	userAgentFiles = fs
	t.Cleanup(func() {
		userAgentFiles = original
	})
}

func TestLoadUserAgents_Valid(t *testing.T) {
	withAgentsFS(t, validAgentsFS)

	agents, err := loadUserAgents()
	if err != nil {
		t.Fatalf("loadUserAgents() returned unexpected error: %v", err)
	}

	// The fixture has 4 lines, including one blank line, all of which the
	// current implementation appends verbatim (no trimming/filtering).
	wantLen := 4
	if len(agents) != wantLen {
		t.Fatalf("loadUserAgents() returned %d agents, want %d (got %#v)", len(agents), wantLen, agents)
	}

	want := []string{"AgentOne/1.0", "AgentTwo/2.0", "", "AgentThree/3.0"}
	for i, w := range want {
		if agents[i] != w {
			t.Errorf("agents[%d] = %q, want %q", i, agents[i], w)
		}
	}
}

func TestLoadUserAgents_EmptyFile(t *testing.T) {
	withAgentsFS(t, emptyAgentsFS)

	agents, err := loadUserAgents()
	if err != nil {
		t.Fatalf("loadUserAgents() returned unexpected error: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("loadUserAgents() with empty file = %#v, want empty slice", agents)
	}
}

func TestLoadUserAgents_FileMissing(t *testing.T) {
	// This fixture filesystem does not contain resources/user_agents.txt,
	// so Open should fail and the error should be propagated.
	withAgentsFS(t, missingAgentsFS)

	agents, err := loadUserAgents()
	if err == nil {
		t.Fatal("loadUserAgents() expected an error for a missing file, got nil")
	}
	if agents != nil {
		t.Errorf("loadUserAgents() = %#v, want nil on error", agents)
	}
}

func TestGetRandomHeader_ReturnsKnownAgent(t *testing.T) {
	withAgentsFS(t, validAgentsFS)

	// Run multiple times since the result is randomized; every result must
	// still come from the loaded agent list.
	valid := map[string]bool{
		"AgentOne/1.0": true,
		"AgentTwo/2.0": true,
		"":             true,
		"AgentThree/3.0": true,
	}

	for i := 0; i < 20; i++ {
		got, err := GetRandomHeader()
		if err != nil {
			t.Fatalf("GetRandomHeader() returned unexpected error: %v", err)
		}
		if !valid[got] {
			t.Errorf("GetRandomHeader() = %q, want one of the loaded agents", got)
		}
	}
}

func TestGetRandomHeader_Randomness(t *testing.T) {
	withAgentsFS(t, validAgentsFS)

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		got, err := GetRandomHeader()
		if err != nil {
			t.Fatalf("GetRandomHeader() returned unexpected error: %v", err)
		}
		seen[got] = true
	}

	// With 4 distinct entries and 50 draws, we should see more than a single
	// value returned. This isn't a strict randomness proof, but it catches an
	// index calculation that always returns the same element.
	if len(seen) <= 1 {
		t.Errorf("GetRandomHeader() returned only %d distinct value(s) across 50 calls: %#v", len(seen), seen)
	}
}

func TestGetRandomHeader_PropagatesLoadError(t *testing.T) {
	withAgentsFS(t, missingAgentsFS)

	got, err := GetRandomHeader()
	if err == nil {
		t.Fatal("GetRandomHeader() expected an error when the agents file is missing, got nil")
	}
	if got != "" {
		t.Errorf("GetRandomHeader() = %q, want empty string on error", got)
	}
}

// TestGetRandomHeader_EmptyAgentListPanics documents a real bug: when the
// loaded agent list is empty, rand.Intn(0) panics rather than returning an
// error. This test asserts the current (undesirable) panic behavior so that
// fixing it later is a deliberate, visible change rather than a silent one.
func TestGetRandomHeader_EmptyAgentListPanics(t *testing.T) {
	withAgentsFS(t, emptyAgentsFS)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected GetRandomHeader() to panic with an empty agent list, but it did not")
		}
		if !strings.Contains(fmtRecover(r), "invalid argument") {
			t.Logf("panic value: %v", r)
		}
	}()

	_, _ = GetRandomHeader()
	t.Fatal("expected panic before reaching this point")
}

func fmtRecover(r interface{}) string {
	if err, ok := r.(error); ok {
		return err.Error()
	}
	return ""
}