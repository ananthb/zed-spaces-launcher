package daemon

import (
	"fyne.io/fyne/v2"

	"github.com/ananth/cosmonaut/internal/codespace"
	"github.com/ananth/cosmonaut/internal/config"
	"github.com/ananth/cosmonaut/internal/history"
)

const (
	guiWidth  float32 = 500
	guiHeight float32 = 400
)

// historyLoad wraps history.Load for use in the daemon package.
func historyLoad() *history.History {
	return history.Load()
}

// countRecentRepos counts leading repos in sorted that appear in history.
func countRecentRepos(sorted []string, hist *history.History) int {
	n := 0
	for _, repo := range sorted {
		found := false
		for _, e := range hist.Entries {
			if e.Repository == repo {
				found = true
				break
			}
		}
		if !found {
			break
		}
		n++
	}
	return n
}

// mergeRepos adds extra repos to the list, skipping duplicates.
// Duplicated here from main.go since it's needed in the daemon package.
func mergeRepos(base, extra []string) []string {
	seen := make(map[string]bool, len(base))
	for _, r := range base {
		seen[r] = true
	}
	result := make([]string, len(base))
	copy(result, base)
	for _, r := range extra {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}
	return result
}

// configRepos returns unique repository names from config targets.
// Duplicated here from main.go since it's needed in the daemon package.
func configRepos(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	seen := make(map[string]bool)
	var repos []string
	for _, t := range cfg.Targets {
		if t.Repository != "" && !seen[t.Repository] {
			seen[t.Repository] = true
			repos = append(repos, t.Repository)
		}
	}
	return repos
}

// guiTargetForRepo finds a config target matching the repo, or builds a default.
// Duplicated from main.go targetForRepo.
func guiTargetForRepo(cfg *config.Config, repo string) (config.Target, string) {
	if cfg != nil {
		for name, t := range cfg.Targets {
			if t.Repository == repo {
				return t, name
			}
		}
	}

	parts := splitRepo(repo)
	repoName := parts[len(parts)-1]
	return config.Target{
		Repository:    repo,
		WorkspacePath: "/workspaces/" + repoName,
	}, repo
}

func splitRepo(repo string) []string {
	var parts []string
	start := 0
	for i, c := range repo {
		if c == '/' {
			parts = append(parts, repo[start:i])
			start = i + 1
		}
	}
	parts = append(parts, repo[start:])
	return parts
}

// createGUIWindow creates a standard-sized GUI window.
func (d *Daemon) createGUIWindow(title string) fyne.Window {
	win := d.app.NewWindow(title)
	win.Resize(fyne.NewSize(guiWidth, guiHeight))
	win.SetFixedSize(true)
	win.CenterOnScreen()
	return win
}

// showRepoListForCodespaces returns repos that have codespaces.
func reposWithCodespaces(codespaces []codespace.Codespace) []string {
	return codespace.UniqueRepos(codespaces)
}
