package zed

import (
	"testing"
)

func TestUpsertConnectionMergesByHostIdentity(t *testing.T) {
	settings := map[string]any{
		"theme": "One Dark",
		"ssh_connections": []any{
			map[string]any{
				"host":     "cs-demo",
				"nickname": "Old",
				"projects": []any{
					map[string]any{"paths": []any{"/old"}},
				},
			},
		},
	}

	conn := BuildConnection("cs-demo", "/workspaces/demo", "New", nil)
	updated := UpsertConnection(settings, conn)

	if updated["theme"] != "One Dark" {
		t.Error("theme should be preserved")
	}

	conns := updated["ssh_connections"].([]any)
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}

	c := conns[0].(map[string]any)
	if c["nickname"] != "New" {
		t.Errorf("nickname = %v, want New", c["nickname"])
	}

	projects := c["projects"].([]any)
	p := projects[0].(map[string]any)
	paths := p["paths"].([]any)
	if paths[0] != "/workspaces/demo" {
		t.Errorf("path = %v, want /workspaces/demo", paths[0])
	}
}

func TestUpsertConnectionAddsNewConnection(t *testing.T) {
	settings := map[string]any{}
	conn := BuildConnection("cs-new", "/workspaces", "test", nil)
	updated := UpsertConnection(settings, conn)

	conns := updated["ssh_connections"].([]any)
	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
}

func TestResolveNickname(t *testing.T) {
	if got := ResolveNickname("zed-nick", "display", "cs-display", "target"); got != "zed-nick" {
		t.Errorf("got %q, want zed-nick", got)
	}
	if got := ResolveNickname("", "display", "cs-display", "target"); got != "display" {
		t.Errorf("got %q, want display", got)
	}
	if got := ResolveNickname("", "", "cs-display", "target"); got != "cs-display" {
		t.Errorf("got %q, want cs-display", got)
	}
	if got := ResolveNickname("", "", "", "target"); got != "target" {
		t.Errorf("got %q, want target", got)
	}
}
