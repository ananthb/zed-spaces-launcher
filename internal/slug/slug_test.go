package slug

import (
	"strings"
	"testing"
)

func TestSlugifyWorkLabelNormalizesText(t *testing.T) {
	got := SlugifyWorkLabel("Fix RPC / Health Checks")
	want := "fix-rpc-health-checks"
	if got != want {
		t.Errorf("SlugifyWorkLabel() = %q, want %q", got, want)
	}
}

func TestBuildDisplayNameTruncatesToCodespacesLimit(t *testing.T) {
	got := BuildDisplayName("rpcpool/rpcpool", "main", strings.Repeat("a", 80), "")
	if len(got) > 48 {
		t.Errorf("len = %d, want <= 48", len(got))
	}
	if !strings.HasPrefix(got, "rpcpool-main-") {
		t.Errorf("got %q, want prefix rpcpool-main-", got)
	}
}

func TestBuildDisplayNameUsesFallback(t *testing.T) {
	got := BuildDisplayName("rpcpool/rpcpool", "main", "", "rpcpool-main")
	if got != "rpcpool-main" {
		t.Errorf("got %q, want rpcpool-main", got)
	}
}
