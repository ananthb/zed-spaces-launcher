package codespace

import (
	"testing"

	"github.com/linuskendall/cosmonaut/internal/config"
)

func TestMatchesTargetComparesRepoBranchAndDisplayName(t *testing.T) {
	cs := Codespace{
		Name:        "demo-abc123",
		DisplayName: "demo-main",
		Repository:  "acme/demo",
		GitStatus:   &GitStatus{Ref: "main"},
	}

	target := config.Target{
		Repository:  "acme/demo",
		Branch:      "main",
		DisplayName: "demo-main",
	}
	if !MatchesTarget(&cs, &target) {
		t.Error("expected match")
	}

	noMatch := config.Target{
		Repository: "acme/demo",
		Branch:     "develop",
	}
	if MatchesTarget(&cs, &noMatch) {
		t.Error("expected no match for different branch")
	}
}

func TestMatchesTargetAcceptsStringRepositoryShape(t *testing.T) {
	cs := Codespace{
		Name:        "demo-abc123",
		DisplayName: "demo-main",
		Repository:  "acme/demo",
		GitStatus:   &GitStatus{Ref: "main"},
	}

	target := config.Target{
		Repository: "acme/demo",
		Branch:     "main",
	}
	if !MatchesTarget(&cs, &target) {
		t.Error("expected match with string repository")
	}
}

func TestChooseCodespaceErrorsOnAmbiguous(t *testing.T) {
	codespaces := []Codespace{
		{Name: "demo-a", Repository: "acme/demo"},
		{Name: "demo-b", Repository: "acme/demo"},
	}
	target := config.Target{Repository: "acme/demo"}

	_, err := ChooseCodespace(codespaces, &target)
	if err == nil {
		t.Fatal("expected error for ambiguous match")
	}
	if got := err.Error(); !contains(got, "Ambiguous") && !contains(got, "ambiguous") {
		t.Errorf("error = %q, want ambiguous", got)
	}
}

func TestBuildCreateArgsMapsOptionalFields(t *testing.T) {
	target := config.Target{
		Repository:       "acme/demo",
		Branch:           "main",
		DisplayName:      "demo-main",
		Machine:          "standardLinux32gb",
		Location:         "EastUs",
		DevcontainerPath: ".devcontainer/devcontainer.json",
		IdleTimeout:      "30m",
		RetentionPeriod:  "72h",
	}

	got := BuildCreateArgs(target)
	want := []string{
		"gh", "codespace", "create", "--repo", "acme/demo",
		"--branch", "main",
		"--display-name", "demo-main",
		"--machine", "standardLinux32gb",
		"--location", "EastUs",
		"--devcontainer-path", ".devcontainer/devcontainer.json",
		"--idle-timeout", "30m",
		"--retention-period", "72h",
	}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("arg[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDescribeCodespaceMarksMatchingEntry(t *testing.T) {
	cs := Codespace{
		Name:        "demo-a",
		DisplayName: "rpcpool-main",
		State:       "Available",
		GitStatus:   &GitStatus{Ref: "main"},
	}

	desc := DescribeCodespace(&cs, true)
	if !contains(desc, "[matches config]") {
		t.Errorf("missing [matches config] in %q", desc)
	}
	if !contains(desc, "branch=main") {
		t.Errorf("missing branch=main in %q", desc)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
