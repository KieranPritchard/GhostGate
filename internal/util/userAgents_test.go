package util

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// In-memory mock filesystems using testing/fstest.MapFS.
// This avoids reading real production files or requiring disk fixture files.

var validAgentsFS = fstest.MapFS{
	"resources/user_agents.txt": &fstest.MapFile{
		Data: []byte("AgentOne/1.0\nAgentTwo/2.0\n\nAgentThree/3.0\n"),
	},
}

var emptyAgentsFS = fstest.MapFS{
	"resources/user_agents.txt": &fstest.MapFile{
		Data: []byte(""),
	},
}

var missingAgentsFS = fstest.MapFS{}

// withAgentsFS temporarily swaps the package-level userAgentFiles variable,
// restoring the original (real) embedded filesystem once the test finishes.
func withAgentsFS(t *testing.T, fsys fs.FS) {
    t.Helper()
    original := userAgentFiles
    userAgentFiles = fsys
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

	want := []string{"AgentOne/1.0", "AgentTwo/2.0", "", "AgentThree/3.0"}
	if len(agents) != len(want) {
		t.Fatalf("loadUserAgents() returned %d agents, want %d\n got: %#v\nwant: %#v",
			len(agents), len(want), agents, want)
	}

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

	valid := map[string]bool{
		"AgentOne/1.0":   true,
		"AgentTwo/2.0":   true,
		"":               true,
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

func TestGetRandomHeader_EmptyAgentListPanics(t *testing.T) {
	withAgentsFS(t, emptyAgentsFS)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected GetRandomHeader() to panic with an empty agent list, but it did not")
		}
		if !strings.Contains(fmtRecover(r), "invalid argument") && !strings.Contains(fmtRecover(r), "slice bounds out of range") {
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
	return fmt.Sprintf("%v", r)
}